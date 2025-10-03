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

	"ivxv.ee/common/collector/command"
	"ivxv.ee/common/collector/command/exit"
	"ivxv.ee/common/collector/conf"
	"ivxv.ee/common/collector/container"
	"ivxv.ee/common/collector/crypto/elgamal"
	"ivxv.ee/common/collector/errors"
	"ivxv.ee/common/collector/identity"
	"ivxv.ee/common/collector/log"
	"ivxv.ee/common/collector/q11n"
	"ivxv.ee/common/collector/server"
	"ivxv.ee/common/collector/status/client"
	status "ivxv.ee/common/collector/status/client/rpc"
	"ivxv.ee/common/collector/storage"
	"ivxv.ee/voting/internal"
	sessstatus "ivxv.ee/voting/internal/sessionstatus/rpc"

	"tivi.io/core/crypto"
	x_509 "tivi.io/core/crypto/x509"
	"tivi.io/core/math/group"
	_ "tivi.io/core/math/group/all"
	//ivxv:modules common/collector/auth
	//ivxv:modules common/collector/container
	//ivxv:modules common/collector/q11n
	//ivxv:modules common/collector/storage
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
	status             client.Verifier
	election           *conf.Election
	container          container.Opener
	identify           identity.Identifier
	q11n               q11n.Qualifiers
	storage            *storage.Client
	ciphertextVerifier internal.ASN1CiphertextVerifier

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
		Choices:     args.Choices,
		Type:        args.Type,
		Vote:        log.Sensitive(args.Vote),
		Description: _VOTING_VOTEREQ,
	})

	// Get the voter identifier associated with RPC call.
	// If empty, then the request is not authenticated.
	auther := server.VoterIdentity(args.Ctx)
	if len(auther) == 0 {
		log.Error(args.Ctx, UnauthenticatedVoteError{Description: _VOTING_VOTER_NO_AUTH})
		return server.ErrUnauthenticated
	}

	// Build up VerifyReq for session status service
	verifyReq := status.NewVerifyReqBuilder().
		WithServiceMethod(sessstatus.Vote).
		WithRequest(args.Header).
		Build()

	// SessionID security check
	ok, err := r.status.Verify(&verifyReq)
	if err != nil {
		// Error during SessionID check - database unreachable, service stalled, etc.
		log.Error(args.Ctx, VoteVerifySessionIDError{Err: err, Description: _VOTING_SESSION_ID})
		return server.ErrBadRequest
	}
	if !ok {
		// SessionID is unknown / has expired, we shall not further process the request
		log.Error(args.Ctx, VoteUpdateSessionIDError{Description: _VOTING_SESSION_ID_EXPIRED})
		return server.ErrBadRequest
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
	//
	// Return all voter ballots in order to perform ciphertext validation
	votec, ballots, signer, voterName, version, err := r.verify(args.Ctx, args.Choices, args.Type, args.Vote)
	if votec != nil {
		defer votec.Close()
	}
	if err != nil {
		// Errors have already been logged by verify.
		return err
	}

	// For each ballot, perform ciphertext correctness checks
	for _, ballot := range ballots {
		err := r.ciphertextVerifier.ASN1CiphertextVerify(ballot)
		if err != nil {
			log.Error(args.Ctx, VoteInvalidCiphertextError{
				Err:         err,
				Description: _VOTING_CIP,
			})
			return server.ErrBadRequest
		}
	}

	// Check that the submitter matches the signer.
	if auther != signer {
		log.Error(args.Ctx, AuthenticatedSignerMismatchError{
			Authenticated: auther,
			Signer:        signer,
			Description:   _VOTING_ID_MISMATCH,
		})
		return server.ErrIdentityMismatch
	}

	// Check if an authentication token has already specified the vote
	// identifier. If not, generate one. Alert, if it cannot be generated.
	resp.VoteID = server.VoteIdentifier(args.Ctx)
	if len(resp.VoteID) == 0 {
		if resp.VoteID, err = genIdent(); err != nil {
			log.Error(args.Ctx, GenerateVoteIDError{Err: log.Alert(err),
				Description: _VOTING_VOTEID})
			return server.ErrInternal
		}
	}

	// VoteID successfully generated
	log.Log(args.Ctx, VoteID{VoteID: resp.VoteID, Description: _VOTING_VOTEID_OK})

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
		// Error in storing the vote, database may be unreachable
		log.Error(args.Ctx, StoreVoteError{Err: log.Alert(err), Description: _VOTING_STORE})
		return server.ErrInternal
	}

	// Request qualifying properties for the vote and store them.
	resp.Qualification = make(q11n.Properties)
	logq11n := make(map[q11n.Protocol]log.Sensitive)

	// Init Transaction
	transaction := r.storage.Txn()
	log.Debug(args.Ctx, VoteTxnInit{VoteID: resp.VoteID, Description: _VOTING_TXN_INIT})
	// BEGIN
	txnOp, err := transaction.Begin(args.Ctx)
	if err != nil {
		// Backend cannot begin a transaction for vote storage. Database unreachable?
		log.Error(args.Ctx, VoteTxnSetVotedBeginError{Err: err, Description: _VOTING_TXN_START})
		return server.ErrInternal
	}
	// Transaction successfully initiated
	log.Debug(args.Ctx, VoteTxnBegin{VoteID: resp.VoteID, Description: _VOTING_TXN_STARTED})
	// Transaction: SET AUTOCOMMIT ON
	transaction.AutoCommit(args.Ctx, txnOp)
	log.Debug(args.Ctx, VoteTxnSetAutoCommitToOn{VoteID: resp.VoteID, Description: _VOTING_TXN_SET_AUTOCOMMIT})

	for _, q := range r.q11n {
		// Initiate query to request qualifying property
		log.Log(args.Ctx, RequestingQualifyingProperty{Protocol: q.Protocol, Description: _VOTING_Q_PROP})
		prop, err := q.Qualifier.Qualify(args.Ctx, votec)
		if err != nil {
			if errors.CausedBy(err, new(q11n.BadCertificateStatusError)) != nil {
				// Qualification service - OCSP - has reported bad certificate status
				log.Error(args.Ctx, BadSignerCertificateError{Err: err,
					Description: _VOTING_Q_PROP_BAD_CERT})
				return server.ErrCertificate
			}

			// Qualification service has returned an error
			log.Error(args.Ctx, QualifierError{
				Protocol:    q.Protocol,
				Err:         log.Alert(err),
				Description: _VOTING_Q_PROP_FAIL,
			})
			return server.ErrInternal
		}

		// Qualifying property has been successfully retreived, we treat it's value as
		// sensitive
		log.Log(args.Ctx, QualifyingProperty{
			Protocol:    q.Protocol,
			Property:    log.Sensitive(prop),
			Description: _VOTING_Q_PROP_OK,
		})

		switch q.Protocol {
		// Don't store TSPREG response in transaction, instead
		// store it immediately, this is a requirement!
		// TODO: review this code, since qualifying services are dynamically configured
		// What we need to achieve, that each response is stored as soon as it is received
		// and only final statistics is updated in transaction
		case q11n.TSPREG, q11n.TSP:
			err = r.storage.StoreQualifyingProperty(
				args.Ctx, resp.VoteID, q.Protocol, prop)
			if err != nil {
				// Error in storing the qualifying property, database may be unreachable.
				log.Error(args.Ctx, StoreQualifyingPropertyError{
					Err:         log.Alert(err),
					Description: _VOTING_Q_PROP_STORE,
				})
				return server.ErrInternal
			}
		case q11n.OCSP:
			r.storage.TxnStoreQualifyingProperty(
				args.Ctx, resp.VoteID, q.Protocol, prop, txnOp)
		}

		// Transaction properly finished
		log.Log(args.Ctx, VoteTxnStoreQualifyingProperty{Protocol: q.Protocol,
			Description: _VOTING_Q_PROP_STORE_INFO})
		resp.Qualification[q.Protocol] = prop
		logq11n[q.Protocol] = prop
	}

	ctime := submitted
	if q11nTime, err := q11n.CanonicalTime(resp.Qualification); err != nil {
		// Error in retrieving qulification time
		// Do not return an error here: as fas as the voter is
		// concerned, they voted successfully.
		// TODO: this logic needs validation
		log.Error(args.Ctx, QualificationTimeError{Err: log.Alert(err),
			Description: _VOTING_Q_PROP_TIMESTAMP})
	} else if !q11nTime.IsZero() {
		// Qualification time
		log.Log(args.Ctx, QualificationTime{Time: q11nTime, Description: _VOTING_TIMESTAMP})
		ctime = q11nTime
	}

	// Check if the vote canonical time was before election start: if so,
	// report to the voter that this is a test vote.
	if ctime.Before(r.start) {
		log.Log(args.Ctx, TestVote{VoteID: resp.VoteID, Description: _VOTING_TEST_VOTE})
		resp.TestVote = true
	}

	// Check qualifiers' times
	err = q11n.CompareQualificationTimes(r.q11n, resp.Qualification)
	if errors.CausedBy(err, new(q11n.NoPreconfiguredQualifiersError)) != nil {
		// Don't return err to client, since NoPreconfiguredQualifiersError means
		// that administrator didn't set "qualification:" in election.yml
		// and is aware of that
		log.Log(args.Ctx, VoteNoPreconfiguredQualifiersError{Err: err, Description: _VOTING_NO_TIMESTAMP_Q})
	} else if err != nil {
		// TSA timestamp is > OCSP producedAt value, this should never happen in production
		// and will only mean that OCSP authority and TSA authority are out of the sync
		log.Error(args.Ctx, VoteCompareQualificationTimesError{Err: log.Alert(err),
			Description: _VOTING_INVALID_TIMESTAMP})
		return server.ErrInternal
	}

	err = r.storage.TxnSetVoted(args.Ctx, txnOp, resp.VoteID, voterName, ctime, resp.TestVote)
	// Vote has been successfully stored. TODO - is this logging step in the correct place?
	log.Log(args.Ctx, VoteTxnSetVoted{VoteID: resp.VoteID, Description: _VOTING_TXN_END})

	// if err is equals or contains nested storage.UnexpectedValueError
	if errors.CausedBy(err, new(storage.UnexpectedValueError)) != nil {
		// Transaction with statistics could not be finished. Is database reachable?
		log.Error(args.Ctx, VoteTxnSetVotedAutoCommitError{Err: log.Alert(err),
			Description: _VOTING_TXN_ROLLBACK})
		return server.ErrInternal
	} else if err != nil {
		// Do not return an error here: as fas as the voter is
		// concerned, they voted successfully. TODO - this needs review.
		log.Error(args.Ctx, VoteTxnSetVotedError{Err: log.Alert(err),
			Description: _VOTING_TXN_COMMITTED})
	}

	// Vote has been successfully stored and reply provided to the voteapp
	log.Log(args.Ctx, VoteResp{
		VoteID:        resp.VoteID,
		Qualification: logq11n,
		Description:   _VOTING_VOTERESP,
	})
	return nil
}

// ratelimit applies rate limiting to vote submissions by the same voter.
func (r *RPC) ratelimit(ctx context.Context, voter string, submitted time.Time) error {
	start := r.election.Voting.RateLimitStart
	minutes := time.Duration(r.election.Voting.RateLimitMinutes) * time.Minute //nolint:gosec

	// If the vote submission statistics changed between retrieving and
	// attempting to update, then that means that there was another
	// concurrent voting session for this voter that got there first. If
	// this happens, then try from the beginning, but still use the same
	// submission time.
	for {
		submissions, last, err := r.storage.GetVoterRateStats(ctx, voter)
		if err != nil {
			// Backend could not fetch voter's voting frequency data from the storage.
			// Database may be unreachable.
			log.Error(ctx, GetVoterRateStatsError{Err: log.Alert(err),
				Description: _VOTING_RATE})
			return server.ErrInternal
		}

		// Check if rate limiting should be applied to this voter.
		if submissions >= start && submitted.Before(last.Add(minutes)) {
			// Voter has exceeded the maximum amount of voting attemps and is restricted to
			// vote exactly 1 time per ratelimitminutes set in election.yml. Current session
			// is rejected, since it does not respect the limit
			log.Error(ctx, RateLimitAppliedError{
				Submissions: submissions,
				Last:        last,
				Now:         submitted,
				Description: _VOTING_LIMIT_TIME,
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
		// not refresh the timeout if a new vote came too early.
		if err = r.storage.SetVoterRateStats(ctx, voter,
			submissions, last, timestamp); err != nil {

			// We attempted to update statistics for rate limit and received an
			// unexpected value, which may indicate a parallel voting process for the
			// particular voter
			if errors.CausedBy(err, new(storage.UnexpectedValueError)) != nil {
				log.Log(ctx, ConcurrentVotingWarning{Err: err, Description: _VOTING_RACE_VOTE})
				continue
			}

			// Error updating statistics in the storage, database may be unreachable
			log.Error(ctx, SetVoterRateStatsError{Err: log.Alert(err), Description: _VOTING_RATE_UPDATE})
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
	votec container.Container, ballots map[string][]byte, identity, voterName, version string, err error) {

	// ASiCE alias of BDOC is not currently allowed for voting.
	if t == container.ASiCE {
		log.Error(ctx, ASiCEVoteNotAllowedError{Description: _VOTING_ASICE})
		err = server.ErrBadRequest
		return
	}

	// Open the container.
	votec, err = r.container.Open(t, bytes.NewReader(containerb))
	if err != nil {
		// Malformed vote-container received from the voting application
		log.Error(ctx, OpenContainerError{Err: err, Description: _VOTING_BDOC_OPEN})
		votec = nil // Ensure we do not have a half-initialized container.
		err = server.ErrBadRequest
		return
	}

	signatures := votec.Signatures()

	// Container successfully opened
	log.Log(ctx, ContainerOpened{Signatures: signatures, Description: _VOTING_BDOC_SIG})

	// Only one signer is allowed to be associated with the vote container
	if len(signatures) != 1 {
		log.Error(ctx, NoSingleSignatureError{Description: _VOTING_BDOC_MANY_SIG})
		err = server.ErrBadRequest
		return
	}
	signer := signatures[0].Signer
	if identity, err = r.identify(&signer.Subject); err != nil {
		// Backend cannot parse personal code from a client's RPC "Vote" container
		log.Error(ctx, SignerIdentityError{Err: err, Description: _VOTING_ID_FROM_BDOC})
		err = server.ErrIneligible
		return
	}

	// Signer successfully identified
	log.Log(ctx, Signer{Identity: identity, Description: _VOTING_IDENTITY})
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
				// Voter successfully authenticated to the backend but
				// was not found in the current voterlist
				log.Error(ctx, IneligibleVoterError{Err: err, Description: _VOTING_NO_IN_VOTERSLIST})
				err = server.ErrIneligible
				return
			}
			// Backend cannot fetch choices list from storage, database may be unreachable
			log.Error(ctx, VoterChoicesError{Err: log.Alert(err), Description: _VOTING_VOTERSLIST})
			err = server.ErrInternal
			return
		}
		if choices != current {
			// Backend fetched current choices list and it doesn't match RPC "Choices"
			// argument that client has sent, more recent voter list is available
			log.Error(ctx, OutdatedChoicesError{Choices: choices, Current: current,
				Description: _VOTING_DISTRICTS_MISMATCH})
			err = server.ErrOutdatedChoices
			return
		}

		// Voter eligibility has been successfully verified
		log.Log(ctx, VoterEligible{Version: version, Description: _VOTING_ELIGIBLE})
	}

	// All ballots must have a key of "<election>.<question>.ballot".
	ballots = make(map[string][]byte)
ballots:
	for key, value := range votec.Data() {
		for _, q := range r.election.Questions {
			if key == fmt.Sprintf("%s.%s.ballot", r.election.Identifier, q) {
				ballots[q] = value
				continue ballots
			}
		}

		// More datafiles found from the container than expected.
		// Their content is treated as sensitive information
		log.Error(ctx, ExtraDataError{Key: key, Value: log.Sensitive(value),
			Description: _VOTING_BDOC_EXTRA})
		err = server.ErrBadRequest
		return
	}
	if got, want := len(ballots), len(r.election.Questions); got != want {
		var gotid []string
		for key := range ballots {
			gotid = append(gotid, key)
		}
		// If IVXV is configured to have multi-contest election, then we expect that
		// all contests are answered with single ballot.
		log.Error(ctx, MissingBallotsError{Ballots: gotid, Questions: r.election.Questions,
			Description: _VOTING_NO_BALLOT})
		err = server.ErrBadRequest
		return
	}

	// Ballot has been validated successfully
	log.Log(ctx, BallotsOK{Description: _VOTING_BALLOT_OK})

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

	// Configure session status client
	statusClient, errCode := sessstatus.NewClient(c)
	if statusClient == nil || errCode != 0 {
		return errCode
	}

	// Create new RPC instance with the election configuration and storage
	// client.
	rpc := &RPC{election: c.Conf.Election, storage: c.Storage, status: statusClient}

	var start, stop time.Time
	var authConf server.AuthConf
	var err error

	if elec := c.Conf.Election; elec != nil {
		// Check election configuration time values.
		if start, err = elec.ServiceStartTime(); err != nil {
			return c.Error(exit.Config, ServiceStartTimeError{Err: err,
				Description: _VOTING_START},
				"bad service start time:", err)
		}

		if rpc.start, err = elec.ElectionStartTime(); err != nil {
			return c.Error(exit.Config, ElectionStartTimeError{Err: err,
				Description: _VOTING_START_ACCEPT},
				"bad election start time:", err)
		}

		if stop, err = elec.ServiceStopTime(); err != nil {
			return c.Error(exit.Config, StopTimeError{Err: err,
				Description: _VOTING_STOP},
				"bad service stop time:", err)
		}

		// Skip voter eligibility if we are told to ignore it.
		rpc.skipEligible = len(elec.IgnoreVoterList) > 0

		// Set administrative unit code to use for foreign voters.
		rpc.foreignCode = strings.TrimSpace(elec.VoterForeignEHAKDefault())

		// Check voting rate limit values. Non-zero start indicates
		// that rate limiting is desired, but zero minutes disables it.
		if elec.Voting.RateLimitStart > 0 && elec.Voting.RateLimitMinutes == 0 {
			return c.Error(exit.Config, RateLimitError{Description: _VOTING_RATE_LIMITS},
				"voting rate limit start set, but minutes is 0")
		}

		// Create ASN.1 ciphertext correctness verifier
		var g group.Group
		// Prefer public key (optional) over group name (required) from election config
		if elec.Ballot.EncPkey != "" { //nolint:gocritic
			// Get group from a x509 public key
			unmarshaller := x_509.NewUnmarshaller[crypto.EncryptionKey](elgamal.NewX509Unmarshaller()) //nolint:lll
			params, _, err := unmarshaller.Unmarshal([]byte(elec.Ballot.EncPkey))
			if err != nil {
				return c.Error(exit.Config, X509KeyUnmarshalError{Err: err,
					Description: _VOTING_ENC_PUB},
					"failed to x509 unmarshal vote encryption public key:", err)
			}
			g = params.Group()
		} else if elec.Ballot.EncPkeyGroup != "" {
			// Get group straight from a config
			g, err = group.Get(elec.Ballot.EncPkeyGroup)
			if err != nil {
				return c.Error(exit.Config, GroupFetchError{Err: err,
					Description: _VOTING_ENC_G},
					"failed to fetch a group for ciphertext verifier:", err)
			}
		} else {
			return c.Error(exit.Config, NoCiphertextVerifierParamsError{
				Description: _VOTING_NO_CIP_PARAMS},
				"no ciphertext verifier parameters set")
		}

		rpc.ciphertextVerifier = internal.NewElGamalASN1CiphertextVerifier(g)

		// Parse client-authentication configuration.
		if authConf, err = server.NewAuthConf(
			elec.Auth, elec.Identity, &elec.Age); err != nil {

			return c.Error(exit.Config, ServerAuthConfError{Err: err,
				Description: _VOTING_AUTH},
				"failed to configure client authentication:", err)
		}

		// Configure supported ballot container parsers for this
		// election.
		if rpc.container, err = container.Configure(elec.Vote); err != nil {
			return c.Error(exit.Config, ContainerConfError{Err: err,
				Description: _VOTING_BDOC_VERIFIER},
				"failed to configure container parsers:", err)
		}

		// Store the voter identifier for signer identification.
		rpc.identify = authConf.Identity

		// Configure vote qualifiers.
		if rpc.q11n, err = q11n.Configure(elec.Qualification,
			conf.Sensitive(c.Service.ID)); err != nil {

			return c.Error(exit.Config, QualificationConfError{Err: err,
				Description: _VOTING_QUALIFIER},
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
			return c.Error(exit.Config, ServerConfError{Err: err,
				Description: _VOTING_SERVER},
				"failed to configure server:", err)
		}
	}

	// Start listening for incoming connections during the voting period.
	if c.Until >= command.Execute {
		if err = s.WithAuth(authConf).ServeAt(c.Ctx, start); err != nil {
			return c.Error(exit.Unavailable, ServeError{Err: err,
				Description: _VOTING_SERVER_SERVE},
				"failed to serve voting service:", err)
		}
	}
	return exit.OK
}
