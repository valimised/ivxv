/*
The voting service verifies votes, checks voter eligibility, and requests
qualifying properties for the signatures which are all stored in the storage
service.
*/
package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"os"
	"strings"
	"time"

	"ivxv.ee/command"
	"ivxv.ee/command/exit"
	"ivxv.ee/conf"
	"ivxv.ee/container"
	"ivxv.ee/errors"
	"ivxv.ee/identity"
	"ivxv.ee/log"
	"ivxv.ee/q11n"
	"ivxv.ee/server"
	"ivxv.ee/storage"
	//ivxv:modules auth
	//ivxv:modules container
	//ivxv:modules q11n
	//ivxv:modules storage
)

// identLen is the number of random bytes generated for a unique vote
// identifier.
const identLen = 16

func genIdent() (ident []byte, err error) {
	ident = make([]byte, identLen)
	_, err = rand.Read(ident)
	return
}

// RPC is the handler for voting service calls.
type RPC struct {
	election  *conf.Election
	container container.Opener
	identify  identity.Identifier
	q11n      q11n.Qualifiers
	storage   *storage.Client

	// Election start time: all votes registered before this are test
	// votes and are not counted.
	start time.Time

	skipEligible bool   // Should we skip checking voter eligibility?
	foreignCode  string // Administrative unit code for foreign voters.
}

// Args are the arguments provided to a call of RPC.Vote.
type Args struct {
	server.Header
	Choices string         `size:"10"` // Identifier of the choice list used.
	Type    container.Type `size:"10"` // The type of container that the ballot is encapsulated in.
	Vote    []byte         // The signed container of the ballot. Size is limited by codec filter.
}

// Response is the response returned by RPC.Vote.
type Response struct {
	server.Header
	VoteID        []byte          // Generated vote identifier.
	Qualification q11n.Properties // Qualifying properties for the vote.
	TestVote      bool            `json:",omitempty"` // Is this a test vote?
}

// Vote is the remote procedure call performed by clients to submit votes to
// the collector.
func (r *RPC) Vote(args Args, resp *Response) error {
	log.Log(args.Ctx, VoteReq{
		Choices: args.Choices,
		Type:    args.Type,
		Vote:    log.Sensitive(args.Vote),
	})

	// Get the voter identifier. If empty, then the request is not
	// authenticated.
	auther := server.VoterIdentity(args.Ctx)
	if len(auther) == 0 {
		log.Error(args.Ctx, UnauthenticatedVoteError{})
		return server.ErrUnauthenticated
	}

	// Apply rate limiting to vote submissions if enabled.
	submitted := time.Now()
	if r.election.Voting.RateLimitMinutes > 0 {
		if err := r.ratelimit(args.Ctx, auther, submitted); err != nil {
			// Errors have already been logged by ratelimit.
			return err
		}
	}

	// Verify the vote container and get the signer.
	votec, signer, voterName, version, err := r.verify(args.Ctx, args.Choices, args.Type, args.Vote)
	if votec != nil {
		defer votec.Close()
	}
	if err != nil {
		// Errors have already been logged by verify.
		return err
	}

	// Check that the submitter matches the signer.
	if auther != signer {
		log.Error(args.Ctx, AuthenticatedSignerMismatchError{
			Authenticated: auther,
			Signer:        signer,
		})
		return server.ErrIdentityMismatch
	}

	// Check if an authentication token has already specified the vote
	// identifier. If not, generate one.
	resp.VoteID = server.VoteIdentifier(args.Ctx)
	if len(resp.VoteID) == 0 {
		if resp.VoteID, err = genIdent(); err != nil {
			log.Error(args.Ctx, GenerateVoteIDError{Err: log.Alert(err)})
			return server.ErrInternal
		}
	}
	log.Log(args.Ctx, VoteID{VoteID: resp.VoteID})

	// Store the vote identifier, submission time, vote container, voter
	// identity, and voter list version.
	if err = r.storage.StoreVote(args.Ctx, storage.StoredVote{
		VoteID:   resp.VoteID,
		Time:     submitted,
		VoteType: args.Type,
		Vote:     args.Vote,
		Voter:    signer,
		Version:  version,
	}); err != nil {
		log.Error(args.Ctx, StoreVoteError{Err: log.Alert(err)})
		return server.ErrInternal
	}

	// Request qualifying properties for the vote and store them.
	resp.Qualification = make(q11n.Properties)
	logq11n := make(map[q11n.Protocol]log.Sensitive)
	for _, q := range r.q11n {
		log.Log(args.Ctx, RequestingQualifyingProperty{Protocol: q.Protocol})
		prop, err := q.Qualifier.Qualify(args.Ctx, votec)
		if err != nil {
			if errors.CausedBy(err, new(q11n.BadCertificateStatusError)) != nil {
				log.Error(args.Ctx, BadSignerCertificateError{Err: err})
				return server.ErrCertificate
			}
			log.Error(args.Ctx, QualifierError{
				Protocol: q.Protocol,
				Err:      log.Alert(err),
			})
			return server.ErrInternal
		}

		log.Log(args.Ctx, QualifyingProperty{
			Protocol: q.Protocol,
			Property: log.Sensitive(prop),
		})

		if err = r.storage.StoreQualifyingProperty(args.Ctx,
			resp.VoteID, q.Protocol, prop); err != nil {

			log.Error(args.Ctx, StoreQualifyingPropertyError{Err: log.Alert(err)})
			return server.ErrInternal
		}
		log.Log(args.Ctx, StoredQualifyingProperty{Protocol: q.Protocol})

		resp.Qualification[q.Protocol] = prop
		logq11n[q.Protocol] = prop
	}

	ctime := submitted
	if q11nTime, err := q11n.CanonicalTime(resp.Qualification); err != nil {
		log.Error(args.Ctx, QualificationTimeError{Err: log.Alert(err)})
		// Do not return an error here: as fas as the voter is
		// concerned, they voted successfully.
	} else if !q11nTime.IsZero() {
		log.Log(args.Ctx, QualificationTime{Time: q11nTime})
		ctime = q11nTime
	}

	// Check if the vote canonical time was before election start: if so,
	// report to the voter that this is a test vote.
	if ctime.Before(r.start) {
		log.Log(args.Ctx, TestVote{VoteID: resp.VoteID})
		resp.TestVote = true
	}

	if err := r.storage.SetVoted(args.Ctx, resp.VoteID, voterName, ctime, resp.TestVote); err != nil {
		log.Error(args.Ctx, SetVotedError{Err: log.Alert(err)})
		// Do not return an error here: as fas as the voter is
		// concerned, they voted successfully.
	}

	log.Log(args.Ctx, VoteResp{
		VoteID:        resp.VoteID,
		Qualification: logq11n,
	})
	return nil
}

// ratelimit applies rate limiting to vote submissions by the same voter.
func (r *RPC) ratelimit(ctx context.Context, voter string, submitted time.Time) error {
	start := r.election.Voting.RateLimitStart
	minutes := time.Duration(r.election.Voting.RateLimitMinutes) * time.Minute

	// If the vote submission statistics changed between retrieving and
	// attempting to update, then that means that there was another
	// concurrent voting session for this voter that got there first. If
	// this happens, then try from the beginning, but still use the same
	// submission time.
	for {
		submissions, last, err := r.storage.GetVoterRateStats(ctx, voter)
		if err != nil {
			log.Error(ctx, GetVoterRateStatsError{Err: log.Alert(err)})
			return server.ErrInternal
		}

		// Check if rate limiting should be applied to this voter.
		if submissions >= start && submitted.Before(last.Add(minutes)) {
			log.Error(ctx, RateLimitAppliedError{
				Submissions: submissions,
				Last:        last,
				Now:         submitted,
			})
			return server.ErrVotingRateLimit
		}

		// Store the latter time of last and submitted as the
		// timestamp: this way we do not rewind the timestamp of the
		// last vote in case a vote submitted before gets processed
		// later.
		timestamp := submitted
		if last.After(timestamp) {
			timestamp = last
		}

		// Only update statistics if the rate limit was not applied: do
		// not to refresh the timeout if a new vote came too early.
		if err = r.storage.SetVoterRateStats(ctx, voter,
			submissions, last, timestamp); err != nil {

			if errors.CausedBy(err, new(storage.UnexpectedValueError)) != nil {
				log.Log(ctx, ConcurrentVotingWarning{Err: err})
				continue
			}
			log.Error(ctx, SetVoterRateStatsError{Err: log.Alert(err)})
			return server.ErrInternal
		}
		return nil
	}
}

// verify verifies the vote container, verifies voter eligibility and choice
// list used, and checks that the contents of the container are sane. It
// returns the vote container, voter identifier, and version of the voter list
// used for eligibility checks.
func (r *RPC) verify(ctx context.Context, choices string, t container.Type, containerb []byte) (
	votec container.Container, identity, voterName, version string, err error) {

	// As a special case, disallow the ASiCE alias of BDOC for voting.
	if t == container.ASiCE {
		log.Error(ctx, ASiCEVoteNotAllowedError{})
		err = server.ErrBadRequest
		return
	}

	// Open the container.
	votec, err = r.container.Open(t, bytes.NewReader(containerb))
	if err != nil {
		log.Error(ctx, OpenContainerError{Err: err})
		votec = nil // Ensure we do not have a half-initialized container.
		err = server.ErrBadRequest
		return
	}
	signatures := votec.Signatures()
	log.Log(ctx, ContainerOpened{Signatures: signatures})

	// Identify the single signer.
	if len(signatures) != 1 {
		log.Error(ctx, NoSingleSignatureError{})
		err = server.ErrBadRequest
		return
	}
	signer := signatures[0].Signer
	if identity, err = r.identify(&signer.Subject); err != nil {
		log.Error(ctx, SignerIdentityError{Err: err})
		err = server.ErrIneligible
		return
	}
	log.Log(ctx, Signer{Identity: identity})
	firstName := findName(&signer.Subject, asn1.ObjectIdentifier{2, 5, 4, 42})
	lastName := findName(&signer.Subject, asn1.ObjectIdentifier{2, 5, 4, 4})
	voterName = firstName + " " + lastName

	// Verify voter eligibility and choices, unless skipEligible is set.
	version = "N/A" // Mock voter list version used when the voter list is ignored.
	if !r.skipEligible {
		var current string
		version, current, err = r.storage.VoterChoices(ctx, identity, r.foreignCode)
		if err != nil {
			if errors.CausedBy(err, new(storage.NotExistError)) != nil {
				log.Error(ctx, IneligibleVoterError{Err: err})
				err = server.ErrIneligible
				return
			}
			log.Error(ctx, VoterChoicesError{Err: log.Alert(err)})
			err = server.ErrInternal
			return
		}
		if choices != current {
			log.Error(ctx, OutdatedChoicesError{Choices: choices, Current: current})
			err = server.ErrOutdatedChoices
			return
		}
		log.Log(ctx, VoterEligible{Version: version})
	}

	// All ballots must have a key of "<election>.<question>.ballot".
	ballots := make(map[string][]byte)
ballots:
	for key, value := range votec.Data() {
		for _, q := range r.election.Questions {
			if key == fmt.Sprintf("%s.%s.ballot", r.election.Identifier, q) {
				ballots[q] = value
				continue ballots
			}
		}
		log.Error(ctx, ExtraDataError{Key: key, Value: log.Sensitive(value)})
		err = server.ErrBadRequest
		return
	}
	if got, want := len(ballots), len(r.election.Questions); got != want {
		var gotid []string
		for key := range ballots {
			gotid = append(gotid, key)
		}
		log.Error(ctx, MissingBallotsError{Ballots: gotid, Questions: r.election.Questions})
		err = server.ErrBadRequest
		return
	}
	log.Log(ctx, BallotsOK{})

	return
}

// findName searches name for oid and returns the value for that oid or an
// empty string. Panics if the value for the oid is not a string.
func findName(name *pkix.Name, oid asn1.ObjectIdentifier) string {
	for _, n := range name.Names {
		if n.Type.Equal(oid) {
			return n.Value.(string)
		}
	}
	return ""
}
func main() {
	// Call votemain in a separate function so that it can set up defers
	// and have them trigger before returning with a non-zero exit code.
	os.Exit(votemain())
}

func votemain() (code int) {
	c := command.New("ivxv-voting", "")
	defer func() {
		code = c.Cleanup(code)
	}()

	// Create new RPC instance with the election configuration and storage
	// client.
	rpc := &RPC{election: c.Conf.Election, storage: c.Storage}

	var start, stop time.Time
	var authConf server.AuthConf
	var err error
	if elec := c.Conf.Election; elec != nil {
		// Check election configuration time values.
		if start, err = elec.ServiceStartTime(); err != nil {
			return c.Error(exit.Config, ServiceStartTimeError{Err: err},
				"bad service start time:", err)
		}

		if rpc.start, err = elec.ElectionStartTime(); err != nil {
			return c.Error(exit.Config, ElectionStartTimeError{Err: err},
				"bad election start time:", err)
		}

		if stop, err = elec.ServiceStopTime(); err != nil {
			return c.Error(exit.Config, StopTimeError{Err: err},
				"bad service stop time:", err)
		}

		// Skip voter eligibility if we are told to ignore it.
		rpc.skipEligible = len(elec.IgnoreVoterList) > 0

		// Set administrative unit code to use for foreign voters.
		rpc.foreignCode = strings.TrimSpace(elec.VoterForeignEHAKDefault())

		// Check voting rate limit values. Non-zero start indicates
		// that rate limiting is desired, but zero minutes disables it.
		if elec.Voting.RateLimitStart > 0 && elec.Voting.RateLimitMinutes == 0 {
			return c.Error(exit.Config, RateLimitError{},
				"voting rate limit start set, but minutes is 0")
		}

		// Parse client-authentication configuration.
		if authConf, err = server.NewAuthConf(
			elec.Auth, elec.Identity, &elec.Age); err != nil {

			return c.Error(exit.Config, ServerAuthConfError{Err: err},
				"failed to configure client authentication:", err)
		}

		// Configure supported ballot container parsers for this
		// election.
		if rpc.container, err = container.Configure(elec.Vote); err != nil {
			return c.Error(exit.Config, ContainerConfError{Err: err},
				"failed to configure container parsers:", err)
		}

		// Store the voter identifier for signer identification.
		rpc.identify = authConf.Identity

		// Configure vote qualifiers.
		if rpc.q11n, err = q11n.Configure(elec.Qualification,
			conf.Sensitive(c.Service.ID)); err != nil {

			return c.Error(exit.Config, QualificationConfError{Err: err},
				"failed to configure vote qualifiers:", err)
		}
	}

	var s *server.S
	if c.Conf.Technical != nil {
		// Configure a new server with the service instance
		// configuration and the RPC handler instance.
		cert, key := conf.TLS(conf.Sensitive(c.Service.ID))
		if s, err = server.New(&server.Conf{
			CertPath: cert,
			KeyPath:  key,
			Address:  c.Service.Address,
			End:      stop,
			Filter:   &c.Conf.Technical.Filter,
			Version:  &c.Conf.Version,
		}, rpc); err != nil {
			return c.Error(exit.Config, ServerConfError{Err: err},
				"failed to configure server:", err)
		}
	}

	// Start listening for incoming connections during the voting period.
	if c.Until >= command.Execute {
		if err = s.WithAuth(authConf).ServeAt(c.Ctx, start); err != nil {
			return c.Error(exit.Unavailable, ServeError{Err: err},
				"failed to serve voting service:", err)
		}
	}
	return exit.OK
}
