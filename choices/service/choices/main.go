/*
The choices service serves choice lists for voting and verification.
*/
package main

import (
	"os"
	"strings"
	"time"

	internal "ivxv.ee/choices/internal/client/sessionstatus/rpc"
	"ivxv.ee/common/collector/command"
	"ivxv.ee/common/collector/command/exit"
	"ivxv.ee/common/collector/conf"
	"ivxv.ee/common/collector/errors"
	"ivxv.ee/common/collector/log"
	"ivxv.ee/common/collector/server"
	"ivxv.ee/common/collector/status/client"
	status "ivxv.ee/common/collector/status/client/rpc"
	"ivxv.ee/common/collector/storage"
	//ivxv:modules common/collector/auth
	//ivxv:modules common/collector/container
	//ivxv:modules common/collector/storage
)

// RPC is the handler for choices service calls.
type RPC struct {
	status      client.Verifier
	storage     *storage.Client
	forceList   string // If set, VoterChoices always returns this list.
	foreignCode string // Administrative unit code for foreign voters.
}

// ChoicesArgs are the arguments provided to a call of RPC.Choices.
type ChoicesArgs struct {
	server.Header
	Choices string `size:"10"` // Identifier of the requested choice list.
}

// VoterArgs are the arguments provided to a call of RPC.VoterChoices. There
// are none, because the identity of the voter will be extracted from the
// authentication info.
type VoterArgs struct {
	server.Header
}

// Response is the response returned by RPC.Choices and RPC.VoterChoices.
type Response struct {
	server.Header
	Choices string // Identifier of the requested choices.
	List    []byte // The requested choices.
	Voted   bool   `json:",omitempty"` // Has the voter voted already?
}

// Deprecated: Choices RPC endpoint is not used by IVXV backend anymore.
// Choices is the remote procedure call performed by verification clients to
// retrieve a specific choices list.
func (r *RPC) Choices(args ChoicesArgs, resp *Response) (err error) {
	log.Log(args.Ctx, ChoicesReq{Choices: args.Choices,
		Description: _CHOICES_CHOICESREQ})
	resp.Choices = args.Choices

	if resp.List, err = r.storage.GetChoices(args.Ctx, args.Choices); err != nil {
		if errors.CausedBy(err, new(storage.NotExistError)) != nil {
			log.Error(args.Ctx, BadChoicesError{Err: err,
				Description: _CHOICES_NO_FOR_VOTER})
			return server.ErrBadRequest
		}
		log.Error(args.Ctx, GetChoicesError{Err: log.Alert(err),
			Description: _CHOICES_FROM_DB})
		return server.ErrInternal
	}

	// The choices are not actually sensitive, but just really large.
	log.Log(args.Ctx, ChoicesResp{List: log.Sensitive(resp.List),
		Description: _CHOICES_CHOICESRESP})
	return
}

// VoterChoices is the remote procedure call performed by voting clients to
// retrieve the choices list for a voter.
func (r *RPC) VoterChoices(args VoterArgs, resp *Response) (err error) {
	log.Log(args.Ctx, VoterChoicesReq{Description: _CHOICES_VOTERCHOICESREQ})

	// Get the voter identifier associated with the RPC call. If empty, then the request
	// was not properly authenticated and error shall be logged
	voter := server.VoterIdentity(args.Ctx)
	if len(voter) == 0 {
		log.Error(args.Ctx, UnauthenticatedVoterChoicesError{Description: _CHOICES_VOTER_NO_AUTH})
		return server.ErrUnauthenticated
	}

	// Build up VerifyReq for session status service
	verifyReq := status.NewVerifyReqBuilder().
		WithServiceMethod(internal.VoterChoices).
		WithRequest(args.Header).
		Build()

	// SessionID security check
	ok, err := r.status.Verify(&verifyReq)
	if err != nil {
		// Error during SessionID check - database unreachable, service stalled, etc.
		log.Error(args.Ctx, VoterChoicesVerifySessionIDError{Err: err,
			Description: _CHOICES_SESSION_ID})
		return server.ErrBadRequest
	}
	if !ok {
		// SessionID is unknown / has expired, we shall not further process the request
		log.Error(args.Ctx, VoterChoicesUpdateSessionIDError{Description: _CHOICES_SESSION_ID_EXPIRED})
		return server.ErrBadRequest
	}

	if len(r.forceList) > 0 {
		resp.Choices = r.forceList
	} else {
		_, resp.Choices, err = r.storage.VoterChoices(args.Ctx, voter, r.foreignCode)
		if err != nil {
			if errors.CausedBy(err, new(storage.NotExistError)) != nil {
				// Voter successfully authenticated to the backend but
				// was not found in the current voterlist
				log.Error(args.Ctx, IneligibleVoterError{Err: err,
					Description: _CHOICES_NO_VOTER_IN_VOTERS_LIST})
				return server.ErrIneligible
			}
			// Backend cannot fetch choices list from storage, database may be unreachable
			log.Error(args.Ctx, VoterChoicesError{Err: log.Alert(err),
				Description: _CHOICES_ADMIN_CODE})
			return server.ErrInternal
		}
	}
	log.Log(args.Ctx, VoterChoices{Choices: resp.Choices, Description: _CHOICES_DB})

	if resp.List, err = r.storage.GetChoices(args.Ctx, resp.Choices); err != nil {
		// Backend cannot fetch choices list from storage, database may be unreachable
		log.Error(args.Ctx, GetVoterChoicesError{Err: log.Alert(err),
			Description: _CHOICES_FROM_DB})
		return server.ErrInternal
	}

	if resp.Voted, err = r.storage.CheckVoted(args.Ctx, voter); err != nil {
		// Backend cannot check whether voter has already voted or not,
		// database may be unreachable
		log.Error(args.Ctx, CheckVotedError{Err: log.Alert(err),
			Description: _CHOICES_CHECK_VOTED})
		return server.ErrInternal
	}

	// The choices are not actually sensitive, but just really large.
	log.Log(args.Ctx, VoterChoicesResp{List: log.Sensitive(resp.List),
		Description: _CHOICES_VOTERCHOICESRESP})
	return
}

func main() {
	// Call choicesmain in a separate function so that it can set up defers
	// and have them trigger before returning with a non-zero exit code.
	os.Exit(choicesmain())
}

func choicesmain() (code int) {
	c := command.New("ivxv-choices", "")
	defer func() {
		code = c.Cleanup(code)
	}()

	// Configure session status client
	statusClient, errCode := internal.NewClient(c)
	if statusClient == nil || errCode != 0 {
		return errCode
	}

	// Create new RPC instance with storage and status clients
	rpc := &RPC{storage: c.Storage, status: statusClient}

	var start, stop time.Time
	var authConf server.AuthConf
	var err error

	if elec := c.Conf.Election; elec != nil {
		// Check election configuration time values.
		if start, err = elec.ServiceStartTime(); err != nil {
			return c.Error(exit.Config, StartTimeError{Err: err,
				Description: _CHOICES_START},
				"bad service start time:", err)
		}

		if stop, err = elec.ElectionStopTime(); err != nil {
			return c.Error(exit.Config, StopTimeError{Err: err,
				Description: _CHOICES_STOP},
				"bad election stop time:", err)
		}

		// Parse client-authentication configuration.
		if authConf, err = server.NewAuthConf(
			elec.Auth, elec.Identity, &elec.Age); err != nil {

			return c.Error(exit.Config, ServerAuthConfError{Err: err,
				Description: _CHOICES_AUTH},
				"failed to configure client authentication:", err)
		}

		rpc.forceList = strings.TrimSpace(elec.IgnoreVoterList)
		rpc.foreignCode = strings.TrimSpace(elec.VoterForeignEHAKDefault())
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
				Description: _CHOICES_SERVER},
				"failed to configure server:", err)
		}
	}

	// Start listening for incoming connections during the voting period.
	if c.Until >= command.Execute {
		if err = s.WithAuth(authConf).ServeAt(c.Ctx, start); err != nil {
			return c.Error(exit.Unavailable, ServeError{Err: err,
				Description: _CHOICES_SERVER_SERVE},
				"failed to serve choices service:", err)
		}
	}
	return exit.OK
}
