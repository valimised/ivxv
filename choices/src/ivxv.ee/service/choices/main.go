/*
The choices service serves choice lists for voting and verification.
*/
package main

import (
	"os"
	"strings"
	"time"

	"ivxv.ee/command"
	"ivxv.ee/command/exit"
	"ivxv.ee/conf"
	"ivxv.ee/errors"
	"ivxv.ee/log"
	"ivxv.ee/server"
	"ivxv.ee/storage"
	//ivxv:modules auth
	//ivxv:modules container
	//ivxv:modules storage
)

// RPC is the handler for choices service calls.
type RPC struct {
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

// Choices is the remote procedure call performed by verification clients to
// retrieve a specific choices list.
func (r *RPC) Choices(args ChoicesArgs, resp *Response) (err error) {
	log.Log(args.Ctx, ChoicesReq{Choices: args.Choices})
	resp.Choices = args.Choices

	if resp.List, err = r.storage.GetChoices(args.Ctx, args.Choices); err != nil {
		if errors.CausedBy(err, new(storage.NotExistError)) != nil {
			log.Error(args.Ctx, BadChoicesError{Err: err})
			return server.ErrBadRequest
		}
		log.Error(args.Ctx, GetChoicesError{Err: log.Alert(err)})
		return server.ErrInternal
	}

	// The choices are not actually sensitive, but just really large.
	log.Log(args.Ctx, ChoicesResp{List: log.Sensitive(resp.List)})
	return
}

// VoterChoices is the remote procedure call performed by voting clients to
// retrieve the choices list for a voter.
func (r *RPC) VoterChoices(args VoterArgs, resp *Response) (err error) {
	log.Log(args.Ctx, VoterChoicesReq{})

	// Get the voter identifier. If empty, then the request is not
	// authenticated.
	voter := server.VoterIdentity(args.Ctx)
	if len(voter) == 0 {
		log.Error(args.Ctx, UnauthenticatedVoterChoicesError{})
		return server.ErrUnauthenticated
	}

	if len(r.forceList) > 0 {
		resp.Choices = r.forceList
	} else {
		_, resp.Choices, err = r.storage.VoterChoices(args.Ctx, voter, r.foreignCode)
		if err != nil {
			if errors.CausedBy(err, new(storage.NotExistError)) != nil {
				log.Error(args.Ctx, IneligibleVoterError{Err: err})
				return server.ErrIneligible
			}
			log.Error(args.Ctx, VoterChoicesError{Err: log.Alert(err)})
			return server.ErrInternal
		}
	}
	log.Log(args.Ctx, VoterChoices{Choices: resp.Choices})

	if resp.List, err = r.storage.GetChoices(args.Ctx, resp.Choices); err != nil {
		log.Error(args.Ctx, GetVoterChoicesError{Err: log.Alert(err)})
		return server.ErrInternal
	}

	if resp.Voted, err = r.storage.CheckVoted(args.Ctx, voter); err != nil {
		log.Error(args.Ctx, CheckVotedError{Err: log.Alert(err)})
		return server.ErrInternal
	}

	// The choices are not actually sensitive, but just really large.
	log.Log(args.Ctx, VoterChoicesResp{List: log.Sensitive(resp.List)})
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

	// Create new RPC instance with the storage client.
	rpc := &RPC{storage: c.Storage}

	var start, stop time.Time
	var authConf server.AuthConf
	var err error
	if elec := c.Conf.Election; elec != nil {
		// Check election configuration time values.
		if start, err = elec.ServiceStartTime(); err != nil {
			return c.Error(exit.Config, StartTimeError{Err: err},
				"bad service start time:", err)
		}

		if stop, err = elec.ElectionStopTime(); err != nil {
			return c.Error(exit.Config, StopTimeError{Err: err},
				"bad election stop time:", err)
		}

		// Parse client-authentication configuration.
		if authConf, err = server.NewAuthConf(
			elec.Auth, elec.Identity, &elec.Age); err != nil {

			return c.Error(exit.Config, ServerAuthConfError{Err: err},
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
			return c.Error(exit.Config, ServerConfError{Err: err},
				"failed to configure server:", err)
		}
	}

	// Start listening for incoming connections during the voting period.
	if c.Until >= command.Execute {
		if err = s.WithAuth(authConf).ServeAt(c.Ctx, start); err != nil {
			return c.Error(exit.Unavailable, ServeError{Err: err},
				"failed to serve choices service:", err)
		}
	}
	return exit.OK
}
