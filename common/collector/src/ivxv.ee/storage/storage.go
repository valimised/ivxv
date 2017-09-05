/*
Package storage provides common code for communicating with storage services.
*/
package storage // import "ivxv.ee/storage"

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"strings"
	"time"

	"ivxv.ee/container"
	"ivxv.ee/errors"
	"ivxv.ee/log"
	"ivxv.ee/q11n"
	"ivxv.ee/yaml"
)

// Conf is the storage service client protocol configuration.
type Conf struct {
	Protocol Protocol  // The protocol which the client must implement.
	Conf     yaml.Node // Protocol-specific configuration.
}

// Services contains necessary information about the storage client and server
// service instances not part of the client protocol configuration.
type Services struct {
	// Sensitive is the path to the client service directory which can
	// contain sensitive storage client information that can not be passed
	// through the configuration, e.g., authentication credentials.
	Sensitive string

	// Servers are the addresses that the storage client protocol will
	// connect to if using a networked storage service.
	//
	// This is not part of the client protocol configuration to avoid
	// having to duplicate these addresses from the network configuration.
	// Instead it is the callers responsibility to retrieve these addresses
	// from there.
	Servers []string
}

// Client is used to access the storage service.
type Client struct {
	prot PutGetter // The underlying protocol.
}

// New initializes a new storage service client with the provided configuration
// and service instance information.
func New(c *Conf, services *Services) (client *Client, err error) {
	reglock.RLock()
	defer reglock.RUnlock()
	n, ok := registry[c.Protocol]
	if !ok {
		return nil, UnlinkedProtocolError{Protocol: c.Protocol}
	}
	client = new(Client)
	if client.prot, err = n(c.Conf, services); err != nil {
		return nil, ConfigureProtocolError{Protocol: c.Protocol, Err: err}
	}
	return
}

// NewWithProtocol initializes a new storage service client with the provided
// underlying protocol implementation.
//
// Using New with a proper configuration is preferred to NewWithProtocol, but
// the latter can be useful for testing.
func NewWithProtocol(protocol PutGetter) *Client {
	return &Client{prot: protocol}
}

// ensure attempts to put value into the storage service with the given key,
// but if the key already exists, then checks if it already has the specified
// value.
//
// Note that ensure does this in the exact order documented: attempts to put
// first and then checks. This is because it expects the put to generally
// succeed and avoids unnecessary gets.
func (c *Client) ensure(ctx context.Context, key string, value []byte) error {
	switch err := c.prot.Put(ctx, key, value); {
	case err == nil:
		return nil
	case errors.CausedBy(err, new(ExistError)) == nil:
		return EnsurePutError{Err: err}
	}

	existing, err := c.prot.Get(ctx, key)
	if err != nil {
		return EnsureGetExistingError{Err: err}
	}

	if !bytes.Equal(existing, value) {
		return EnsureExistingMismatchError{
			Existing: existing,
			New:      value,
		}
	}
	return nil
}

const (
	choicesPrefix = "/choices/"
	versionKey    = "version"
)

// PutChoices stores the map of choices lists.
func (c *Client) PutChoices(ctx context.Context, version string, choices map[string][]byte) (err error) {
	// Check if a choices list is already stored.
	oldver, err := c.GetChoicesVersion(ctx)
	switch {
	case err == nil:
		return PutChoicesExistsError{Version: oldver}
	case errors.CausedBy(err, new(NotExistError)) == nil:
		return PutChoicesCheckExistingError{Err: err}
	}

	progress := ImportingChoice{ID: "", Index: 0, Count: len(choices)}
	for id, list := range choices {
		// If debugging is enabled, then report progress.
		progress.ID = id
		progress.Index = progress.Index.(int) + 1
		log.Debug(ctx, progress)

		// Ensure that no IDs clash with our version key.
		if id == versionKey {
			return PutChoicesIDNotAllowedError{Choices: id}
		}

		// It is possible that a previous call to PutChoices failed
		// half-way through, resulting in some choices already
		// existing. Use ensure instead of put.
		if err = c.ensure(ctx, choicesPrefix+id, list); err != nil {
			return PutChoicesError{Choices: id, Err: err}
		}
	}

	// Set the marker that choices lists were successfully stored.
	if err = c.prot.Put(ctx, choicesPrefix+versionKey, []byte(version)); err != nil {
		return PutChoicesSetDoneError{Err: err}
	}
	return
}

// GetChoices retrieves the choices list with the given identifier.
func (c *Client) GetChoices(ctx context.Context, choices string) (list []byte, err error) {
	if list, err = c.prot.Get(ctx, choicesPrefix+choices); err != nil {
		err = GetChoicesError{Choices: choices, Err: err}
	}
	return
}

// GetChoicesVersion retrieves the choices list version string.
func (c *Client) GetChoicesVersion(ctx context.Context) (version string, err error) {
	vb, err := c.prot.Get(ctx, choicesPrefix+versionKey)
	version = string(vb)
	if err != nil {
		err = GetChoicesVersionError{Err: err}
	}
	return
}

const (
	votersPrefix = "/voters/"
	previousKey  = "previous"
)

// PutVoters stores a new version of the voters list.
//
// A non-initial voter list acts as an overlay to the previous version of the
// list: if a voter is not specified in the version, then their choices from
// the previous version will persist, recursively. A voter with empty choices
// is considered a deletion of that voter from the list and will be reported as
// not existing.
//
// Voters lists have two different version strings attached to them: the
// container versions and the list version. The former is a concatenation of
// the versions of signed containers used to import voter lists. The latter is
// a version number calculated based on the actual contents of the voter lists.
// Therefore if the same lists are imported but in different signed containers,
// then the container versions will differ but the list version will not.
//
// The container versions are used to provide feedback about which signed
// containers were used to build the voters list. The list version is used to
// compare actual contents of two different lists.
//
// PutVoters concatenates cversion to the current container version and
// compare-and-swaps the list version from oldver to newver, unless oldver is
// empty, in which case it creates a new version file.
func (c *Client) PutVoters(ctx context.Context, cversion string, voters map[string]string,
	oldver, newver string) (err error) {

	const vkey = votersPrefix + versionKey
	prefix := versionPrefix(newver)

	// Do an early check of the old version.
	initial := len(oldver) == 0
	version, err := c.GetVotersListVersion(ctx)
	notexist := errors.CausedBy(err, new(NotExistError)) != nil

	// During an initial load, we expect the version to not exist.
	if initial && !notexist {
		if err != nil {
			// We encountered some other error.
			return PutVotersCheckVersionExistsError{Err: err}
		}
		return PutVotersVersionExistsError{Version: version}
	}

	var oldcver string
	if !initial {
		// When putting a new list version, the old one must exist and
		// match expected.
		if err != nil {
			return PutVotersGetVersionError{Err: err}
		}
		if version != oldver {
			// We cannot use a struct literal, because gen would
			// report it as a duplicate error type.
			var err UnexpectedValueError
			err.Key = vkey
			err.Err = PutVotersVersionMismatchError{
				Version:  version,
				Expected: oldver,
			}
			return err
		}

		// Store a reference to the previous list version.
		if err = c.ensure(ctx, prefix+previousKey, []byte(oldver)); err != nil {
			return PutVotersPreviousVersionError{Version: newver, Err: err}
		}

		// Get the current container versions to concatenate to.
		oldcver, err = c.getCVersion(ctx, version)
		if err != nil {
			return PutVotersGetContainerVersionsError{
				ListVersion: version,
				Err:         err,
			}
		}
	}

	// Create a new voters list version.
	progress := ImportingVoter{Version: newver, Voter: "", Choices: "", Index: 0, Count: len(voters)}
	for voter, choices := range voters {
		// If debugging is enabled, then report progress.
		progress.Voter = voter
		progress.Choices = choices
		progress.Index = progress.Index.(int) + 1
		log.Debug(ctx, progress)

		// Ensure that no voters clash with our version or previous key.
		switch voter {
		case versionKey, previousKey:
			return PutVotersVoterNotAllowedError{Voter: voter}
		}

		// It is possible that a previous call to PutVoters failed
		// half-way through, resulting in some voters already existing.
		// Use ensure instead of put.
		if err = c.ensure(ctx, prefix+voter, []byte(choices)); err != nil {
			return PutVotersError{Version: newver, Voter: voter, Err: err}
		}
	}
	if err = c.ensure(ctx, prefix+versionKey, []byte(oldcver+cversion+"\n")); err != nil {
		return PutVotersContainerVersionsError{Version: newver, Err: err}
	}

	// Either create or CAS the new version number.
	if initial {
		if err = c.prot.Put(ctx, vkey, []byte(newver)); err != nil {
			err = PutVotersNewVersionError{Version: newver, Err: err}
		}
		return
	}

	if _, err = c.prot.CASAndGet(ctx, vkey, []byte(oldver), []byte(newver)); err != nil {
		err = PutVotersCASVersionError{Old: oldver, New: newver, Err: err}
	}
	return
}

// GetVoter returns the identifier of the requested voter's choices list for
// the specified voter list version.
func (c *Client) GetVoter(ctx context.Context, version, voter string) (choices string, err error) {
	// Recursively walk back through version until we find the voter or hit
	// the beginning.
	currentver := version
	for {
		prefix := versionPrefix(currentver)
		choicesb, err := c.prot.Get(ctx, prefix+voter)
		switch {
		case err == nil: // We found an entry for the voter.
			if len(choicesb) == 0 { // The voter has been deleted.
				var ne NotExistError
				ne.Key = prefix + voter
				ne.Err = GetVoterDeletedError{
					Version:   version,
					Voter:     voter,
					DeletedAt: currentver,
				}
				return "", ne
			}
			return string(choicesb), err // Existing voter.

		case errors.CausedBy(err, new(NotExistError)) != nil:
			prevb, verr := c.prot.Get(ctx, prefix+previousKey)
			switch {
			case verr == nil: // Try the previous version.
				currentver = string(prevb)
				continue
			case errors.CausedBy(verr, new(NotExistError)) != nil:
				// We are at the beginning: the voter was not found.
				var ne NotExistError
				ne.Key = versionPrefix(version) + voter
				ne.Err = GetVoterNotFoundError{
					Version: version,
					Voter:   voter,
				}
				return "", ne
			}
			err = verr
		}
		return "", GetVoterError{
			Version: version,
			Voter:   voter,
			Step:    currentver,
			Err:     err,
		}
	}
}

// VoterChoices returns the the current voter list version and the identifier
// of voter's choices list, e.g., a district number.
func (c *Client) VoterChoices(ctx context.Context, voter string) (version, choices string, err error) {
	if version, err = c.GetVotersListVersion(ctx); err != nil {
		return "", "", VoterChoicesVersionError{Err: err}
	}
	if choices, err = c.GetVoter(ctx, version, voter); err != nil {
		return "", "", VoterChoicesError{Err: err}
	}
	return
}

// GetVotersContainerVersions returns the container versions of the current
// voters list.
func (c *Client) GetVotersContainerVersions(ctx context.Context) (cversion string, err error) {
	version, err := c.GetVotersListVersion(ctx)
	if err != nil {
		return "", GetVotersContainerVersionsVersionError{Err: err}
	}
	cversion, err = c.getCVersion(ctx, version)
	if err != nil {
		err = GetVotersContainerVersionsError{
			ListVersion: version,
			Err:         err,
		}
	}
	return
}

// GetVotersListVersion returns the list version of the current voters list.
func (c *Client) GetVotersListVersion(ctx context.Context) (version string, err error) {
	versionb, err := c.prot.Get(ctx, votersPrefix+versionKey)
	if err != nil {
		err = GetVotersListVersionError{Err: err}
	}
	return string(versionb), err
}

// getCVersion is a helper method used to get the voters list container
// versions of a list version.
func (c *Client) getCVersion(ctx context.Context, version string) (cversion string, err error) {
	cvb, err := c.prot.Get(ctx, versionPrefix(version)+versionKey)
	return string(cvb), err
}

func versionPrefix(version string) string {
	return votersPrefix + version + "/"
}

// EligibleVoter checks if the voter is included in the current voter list as an
// eligible voter. Additionally EligibleVoter checks if choices matches the
// current choices list identifier for the voter: otherwise the choices
// available to the voter have been updated and they are asked to restart their
// session. If err is non-nil, then version is the current voter list version.
func (c *Client) EligibleVoter(ctx context.Context, voter, choices string) (version string, err error) {
	version, current, err := c.VoterChoices(ctx, voter)
	if err != nil {
		return "", EligibleVoterChoicesError{Err: err}
	}
	if choices != current {
		return "", OutdatedVoterChoicesError{Choices: choices, Current: current}
	}
	return
}

const votedPrefix = "/voted/"

// SetVoted marks the voter as voted.
func (c *Client) SetVoted(ctx context.Context, voter string) (err error) {
	if err = c.ensure(ctx, votedPrefix+voter, nil); err != nil {
		err = SetVotedError{Voter: voter, Err: err}
	}
	return
}

// CheckVoted checks if the voter has already voted.
func (c *Client) CheckVoted(ctx context.Context, voter string) (voted bool, err error) {
	switch _, err = c.prot.Get(ctx, votedPrefix+voter); {
	case err == nil:
		return true, nil
	case errors.CausedBy(err, new(NotExistError)) != nil:
		return false, nil
	default:
		return false, CheckVotedError{Voter: voter, Err: err}
	}
}

// StoredVote contains a vote and accompanying data. The type is used for
// storing and exporting votes.
type StoredVote struct {
	VoteID        []byte          // Unique identifier of this vote.
	Time          time.Time       // Time the vote was received.
	VoteType      container.Type  // Type of the signed container.
	Vote          []byte          // Signed container that is the actual vote.
	Voter         string          // Identity of the voter.
	Version       string          // Active voter list version when the vote was stored.
	Qualification q11n.Properties // Qualifying properties for the signed container.
}

// missing returns the names of fields of s which are empty. qps lists the
// qualifying properties which must be present.
func (s StoredVote) missing(qps ...q11n.Protocol) (m []string) {
	if len(s.VoteID) == 0 {
		m = append(m, "VoteID")
	}
	if s.Time.IsZero() {
		m = append(m, "Time")
	}
	if len(s.VoteType) == 0 {
		m = append(m, "VoteType")
	}
	if len(s.Vote) == 0 {
		m = append(m, "Vote")
	}
	if len(s.Voter) == 0 {
		m = append(m, "Voter")
	}
	if len(s.Version) == 0 {
		m = append(m, "Version")
	}
	for _, qp := range qps {
		if prop, ok := s.Qualification[qp]; !ok || len(prop) == 0 {
			m = append(m, string(qp))
		}
	}
	return
}

const (
	votePrefix = "/vote/"

	timeKey  = "time"
	typeKey  = "type"
	voteKey  = "vote"
	voterKey = "voter"
	// versionKey is already defined.
	countKey = "count" // The number of times a vote has been verified.

	timefmt = time.RFC3339Nano
)

// StoreVote stores the provided vote and accompanying data in the storage
// service.
func (c *Client) StoreVote(ctx context.Context, vote StoredVote) error {
	if m := vote.missing(); len(m) > 0 {
		return StoreIncompleteVoteError{VoteID: vote.VoteID, Missing: m}
	}

	prefix := voteIDPrefix(vote.VoteID)

	// If putting votes fails, then raise the errors as alerts.

	if err := c.prot.Put(ctx, prefix+timeKey, []byte(vote.Time.Format(timefmt))); err != nil {
		return log.Alert(StoreTimeError{VoteID: vote.VoteID, Err: err})
	}

	if err := c.prot.Put(ctx, prefix+typeKey, []byte(vote.VoteType)); err != nil {
		return log.Alert(StoreVoteTypeError{VoteID: vote.VoteID, Err: err})
	}

	if err := c.prot.Put(ctx, prefix+voteKey, vote.Vote); err != nil {
		return log.Alert(StoreVoteError{VoteID: vote.VoteID, Err: err})
	}

	if err := c.prot.Put(ctx, prefix+voterKey, []byte(vote.Voter)); err != nil {
		return log.Alert(StoreVoterError{VoteID: vote.VoteID, Err: err})
	}

	if err := c.prot.Put(ctx, prefix+versionKey, []byte(vote.Version)); err != nil {
		return log.Alert(StoreVersionError{VoteID: vote.VoteID, Err: err})
	}

	if err := c.prot.Put(ctx, prefix+countKey, make([]byte, 8)); err != nil {
		return log.Alert(StoreVerificationCountError{VoteID: vote.VoteID, Err: err})
	}

	return nil
}

// StoreQualifyingProperty stores a qualifying property for a vote in the
// storage service.
func (c *Client) StoreQualifyingProperty(ctx context.Context, voteID []byte,
	protocol q11n.Protocol, property []byte) error {

	prefix := voteIDPrefix(voteID)
	if err := c.prot.Put(ctx, prefix+string(protocol), property); err != nil {
		return log.Alert(StoreQualifyingPropertyError{
			VoteID:   voteID,
			Protocol: protocol,
			Err:      err,
		})
	}
	return nil
}

// GetVotes retrieves the list of currently stored votes. qps lists the
// qualifying properties which are expected for each vote and optional lists
// the fields of votes that may be missing for the vote to still get exported.
//
// The channels returned by GetVotes behave a little bit different from usual
// pipelines. Instead of reading from the result channel until closed and then
// reading a single error from errc, results and errors should be read in
// parallel. GetVotes will not stop sending results if a non-fatal error is
// encountered: it will just send the error on the error channel and skip the
// errenous entry, continuing with following votes. This way bad entries will
// not cause the entire operation to fail.
//
// Sending on both channels happens in the same goroutine, so callers must read
// from both in parallel for GetVotes not to get blocked on a send. Sending
// results will only stop once all results are sent, the context is cancelled,
// or a fatal error occurs. Sending errors will continue until all errors have
// been read. This means that if the caller cancels the context, they must
// still drain the error channel to unblock GetVotes! Only after this will the
// result and error channels be closed.
//
// Fatal errors will be wrapped in GetVotesFatalError and will never be
// followed by more errors.
func (c *Client) GetVotes(ctx context.Context, qps []q11n.Protocol, optional []string) (
	<-chan *StoredVote, <-chan error) {

	ch := make(chan *StoredVote)
	errc := make(chan error) // Not buffered!
	go func() {
		defer close(errc)
		defer close(ch)

		// partial contains StoredVotes which are still missing fields.
		partial := make(map[string]*StoredVote)
		protc, proterrc := c.prot.GetWithPrefix(ctx, votePrefix)
		for r := range protc {
			// Split the storage key into vote ID and value key.
			voteid, key, err := splitVoteKey(r.Key)
			if err != nil {
				errc <- log.Alert(GetVotesSplitKeyError{Key: r.Key, Err: err})
				continue
			}

			if key == countKey {
				continue // Verification counts are not exported.
			}

			// Get or create the vote ID entry.
			vote, ok := partial[string(voteid)]
			if !ok {
				vote = &StoredVote{
					VoteID:        voteid,
					Qualification: make(q11n.Properties),
				}
				partial[string(voteid)] = vote
			}

			// Set the field according to the key.
		keySwitch:
			switch key {
			case timeKey:
				if vote.Time, err = time.Parse(timefmt, string(r.Value)); err != nil {
					errc <- log.Alert(GetVotesParseTimeError{
						VoteID: voteid,
						Err:    err,
					})
					continue
				}
			case typeKey:
				vote.VoteType = container.Type(r.Value)
			case voteKey:
				vote.Vote = r.Value
			case voterKey:
				vote.Voter = string(r.Value)
			case versionKey:
				vote.Version = string(r.Value)
			default:
				// Check for qualifying properties.
				for _, qp := range qps {
					if key == string(qp) {
						vote.Qualification[qp] = r.Value
						break keySwitch
					}
				}

				errc <- log.Alert(GetVotesInvalidKeyError{
					VoteID: voteid,
					Key:    key,
				})
				continue
			}

			// If vote is now complete, then send it.
			if len(vote.missing(qps...)) == 0 {
				delete(partial, string(voteid))
				select {
				case ch <- vote:
				case <-ctx.Done():
					errc <- ctx.Err()
					return
				}
			}
		}
		if err := <-proterrc; err != nil {
			errc <- log.Alert(GetVotesFatalError{Err: err})
			return
		}

		// Process any remaining partial votes.
		for _, vote := range partial {
			// Send votes that are missing only optional fields.
			if allIn(vote.missing(qps...), optional) {
				select {
				case ch <- vote:
				case <-ctx.Done():
					errc <- ctx.Err()
					return
				}
			}

			// Send an error about any incomplete entries, even if
			// we sent the partial vote.
			errc <- GetVotesIncompleteVoteError{
				VoteID:  vote.VoteID,
				Missing: vote.missing(qps...),
			}
		}
	}()
	return ch, errc
}

func allIn(needles, haystack []string) bool {
next:
	for _, n := range needles {
		for _, p := range haystack {
			if n == p {
				continue next
			}
		}
		return false
	}
	return true
}

// GetVerificationStats returns the statistics about a vote required for
// allowing access to verification: the number of times the vote has been
// verified and when it was cast.
func (c *Client) GetVerificationStats(ctx context.Context, voteid []byte) (
	count uint64, at time.Time, err error) {

	prefix := voteIDPrefix(voteid)

	timeb, err := c.prot.Get(ctx, prefix+timeKey)
	if err != nil {
		err = GetVerificationTimeError{VoteID: voteid, Err: err}
		return
	}

	// If getting the verification time succeeded, then the vote ID is
	// valid and any following errors raise alerts.

	if at, err = time.Parse(timefmt, string(timeb)); err != nil {
		err = log.Alert(ParseVerificationTimeError{VoteID: voteid, Err: err})
		return
	}

	countb, err := c.prot.Get(ctx, prefix+countKey)
	if err != nil {
		err = log.Alert(GetVerificationCountError{VoteID: voteid, Err: err})
		return
	}

	if len(countb) != 8 {
		err = log.Alert(InvalidVerificationCountLengthError{
			VoteID: voteid, Length: len(countb)})
		return
	}
	count = binary.BigEndian.Uint64(countb)

	return
}

// GetVerification returns the subset of a stored vote and accompanying data
// necessary for verification - the vote type, the vote itself and the
// qualifying properties listed in qps - and increases the verification count.
// It does this only if the old verification count is unchanged.
func (c *Client) GetVerification(ctx context.Context,
	voteid []byte, count uint64, qps ...q11n.Protocol) (
	vote StoredVote, err error) {

	prefix := voteIDPrefix(voteid)

	// Serialize the CAS values.
	old := make([]byte, 8)
	new := make([]byte, 8)
	binary.BigEndian.PutUint64(old, count)
	binary.BigEndian.PutUint64(new, count+1)

	// Attempt to read the values.
	paths := make([]string, 2, len(qps)+2)
	paths[0] = prefix + typeKey
	paths[1] = prefix + voteKey
	for _, qp := range qps {
		paths = append(paths, prefix+string(qp))
	}

	m, err := c.prot.CASAndGet(ctx, prefix+countKey, old, new, paths...)
	if err != nil {
		err = GetVerificationError{VoteID: voteid, Err: err}
		return
	}

	// Map the values into the vote.
	vote.VoteType = container.Type(m[prefix+typeKey])
	vote.Vote = m[prefix+voteKey]
	vote.Qualification = make(q11n.Properties)
	for _, qp := range qps {
		vote.Qualification[qp] = m[prefix+string(qp)]
	}
	return
}

func voteIDPrefix(voteid []byte) string {
	return votePrefix + base64.URLEncoding.EncodeToString(voteid) + "/"
}

func splitVoteKey(fullkey string) (voteid []byte, key string, err error) {
	tokens := strings.Split(fullkey[len(votePrefix):], "/")
	if len(tokens) != 2 {
		return nil, "", SplitVoteKeyError{Count: len(tokens), Expected: 2}
	}

	voteid, err = base64.URLEncoding.DecodeString(tokens[0])
	if err != nil {
		return nil, "", SplitVoteKeyDecodeError{Err: err}
	}
	return voteid, tokens[1], nil
}
