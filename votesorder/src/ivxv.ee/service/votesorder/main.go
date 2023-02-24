/*
The votesorder service serves votes lists.
*/
package main

import (
	"os"
	"strconv"
	"time"

	"ivxv.ee/command"
	"ivxv.ee/command/exit"
	"ivxv.ee/conf"
	"ivxv.ee/log"
	"ivxv.ee/server"
	"ivxv.ee/storage"
	//ivxv:modules auth
	//ivxv:modules container
	//ivxv:modules storage
)

// RPC is the handler for votesorder service calls.
type RPC struct {
	storage *storage.Client
}

// VotesNoArgs are the arguments provided to a call of RPC.VoterChoices. There
// are none, because the identity of the voter will be extracted from the
// authentication info.
type VotesNoArgs struct {
	server.Header
}

// VotesNoResponse is the response returned by RPC.VotesSeqNo.
type VotesNoResponse struct {
	server.Header
	SeqNo uint64
}

// VotesSeqNo is the remote procedure call performed by votesorder service provider to
// retrieve current votes sequence number.
func (r *RPC) VotesSeqNo(args VotesNoArgs, resp *VotesNoResponse) (err error) {
	log.Log(args.Ctx, VotesNoReq{})

	if resp.SeqNo, err = r.storage.GetVotesCount(args.Ctx); err != nil {
		log.Error(args.Ctx, GetVotesNoError{Err: log.Alert(err)})
		return server.ErrInternal
	}

	log.Log(args.Ctx, VotesNoResp{SeqNo: resp.SeqNo})
	return
}

// VotesArgs are the arguments provided to a call of RPC.Votes.
type VotesArgs struct {
	server.Header
	VotesFrom    int
	BatchMaxSize int
}

// VotesResponse is the response returned by RPC.Votes.
type VotesResponse struct {
	server.Header
	BatchRecords []Vote `json:"batchRecords"`
}

type Vote struct {
	SeqNo               uint64 `json:"seqNo"`
	IDCode              string `json:"idCode"`
	VoterName           string `json:"voterName"`
	KovCode             string `json:"kovCode"`
	ElectoralDistrictNo uint64 `json:"electoralDistrictNo"`
}

// Votes is the remote procedure call to send votes to votesorder.
func (r *RPC) Votes(args VotesArgs, resp *VotesResponse) (err error) {
	log.Log(args.Ctx, VotesReq{VotesFrom: args.VotesFrom})
	var votes []Vote
	var seqNo uint64
	if seqNo, err = r.storage.GetVotesCount(args.Ctx); err != nil {
		log.Error(args.Ctx, GetVotesCountErr{Err: log.Alert(err)})
		return server.ErrInternal
	}
	if seqNo < uint64(args.VotesFrom) {
		log.Error(args.Ctx, VotesFromErr{SeqNo: seqNo, VotesFrom: args.VotesFrom})
		return server.ErrBadRequest
	}
	votesOrder, err := r.storage.GetVotesOrder(args.Ctx, args.VotesFrom, args.BatchMaxSize)
	if err != nil {
		log.Error(args.Ctx, GetVotesOrderError{Err: log.Alert(err)})
		return server.ErrInternal
	}
	for _, vote := range votesOrder {
		no, err := strconv.ParseUint(vote.SeqNo, 0, 64)
		if err != nil {
			log.Error(args.Ctx, VotesSeqNoParse{SeqNo: vote.SeqNo})
		}
		d, err := strconv.ParseUint(vote.DistrictNo, 0, 64)
		if err != nil {
			log.Error(args.Ctx, VotesDistrictNoParse{DistrictNo: vote.DistrictNo})
		}

		votes = append(votes, Vote{
			SeqNo:               no,
			VoterName:           vote.VoterName,
			IDCode:              vote.IDCode,
			KovCode:             vote.KovCode,
			ElectoralDistrictNo: d,
		})
	}
	log.Log(args.Ctx, VotesResp{})
	resp.BatchRecords = votes
	return
}

func main() {
	// Call votesorder in a separate function so that it can set up defers
	// and have them trigger before returning with a non-zero exit code.
	os.Exit(votesordermain())
}

func votesordermain() (code int) {
	c := command.New("ivxv-votesorder", "")
	defer func() {
		code = c.Cleanup(code)
	}()

	// Create new RPC instance with the storage client.
	rpc := &RPC{storage: c.Storage}

	var stop time.Time
	var clientCA string
	var err error
	if elec := c.Conf.Election; elec != nil {
		if stop, err = elec.ElectionStopTime(); err != nil {
			return c.Error(exit.Config, StopTimeError{Err: err},
				"bad election stop time:", err)
		}
		clientCA = elec.XRoad.CA
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
			ClientCA: clientCA,
		}, rpc); err != nil {
			return c.Error(exit.Config, ServerConfError{Err: err},
				"failed to configure server:", err)
		}
	}

	// Start listening for incoming connections during the voting period.
	if c.Until >= command.Execute {
		if err = s.Serve(c.Ctx); err != nil {
			return c.Error(exit.Unavailable, ServeError{Err: err},
				"failed to serve votesorder service:", err)
		}
	}
	return exit.OK
}
