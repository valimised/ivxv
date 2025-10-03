/*
The votesorder service serves votes lists.
*/
package main

import (
	"os"
	"strconv"
	"time"

	"ivxv.ee/common/collector/command"
	"ivxv.ee/common/collector/command/exit"
	"ivxv.ee/common/collector/conf"
	"ivxv.ee/common/collector/log"
	"ivxv.ee/common/collector/server"
	"ivxv.ee/common/collector/storage"
	//ivxv:modules common/collector/auth
	//ivxv:modules common/collector/container
	//ivxv:modules common/collector/storage
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
	log.Log(args.Ctx, VotesNoReq{Description: _VOTESORDER_VOTESSEQREQ})

	if resp.SeqNo, err = r.storage.GetVotesCount(args.Ctx); err != nil {
		log.Error(args.Ctx, GetVotesNoError{Err: log.Alert(err), Description: _VOTESORDER_VOTESCOUNT})
		return server.ErrInternal
	}

	log.Log(args.Ctx, VotesNoResp{SeqNo: resp.SeqNo, Description: _VOTESORDER_VOTESSEQRESP})
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
	log.Log(args.Ctx, VotesReq{VotesFrom: args.VotesFrom, Description: _VOTESORDER_VOTESREQ})
	var votes []Vote
	var seqNo uint64
	if seqNo, err = r.storage.GetVotesCount(args.Ctx); err != nil {
		log.Error(args.Ctx, GetVotesCountErr{Err: log.Alert(err),
			Description: _VOTESORDER_VOTESCOUNT})
		return server.ErrInternal
	}
	if seqNo < uint64(args.VotesFrom) { //nolint:gosec
		log.Error(args.Ctx, VotesFromErr{SeqNo: seqNo, VotesFrom: args.VotesFrom,
			Description: _VOTESORDER_NO_SEQ})
		return server.ErrBadRequest
	}
	votesOrder, err := r.storage.GetVotesOrder(args.Ctx, args.VotesFrom, args.BatchMaxSize)
	if err != nil {
		log.Error(args.Ctx, GetVotesOrderError{Err: log.Alert(err), Description: _VOTESORDER_ALL_VOTES})
		return server.ErrInternal
	}
	for _, vote := range votesOrder {
		no, err := strconv.ParseUint(vote.SeqNo, 0, 64)
		if err != nil {
			log.Error(args.Ctx, VotesSeqNoParse{SeqNo: vote.SeqNo, Description: _VOTESORDER_SEQ})
		}
		d, err := strconv.ParseUint(vote.DistrictNo, 0, 64)
		if err != nil {
			log.Error(args.Ctx, VotesDistrictNoParse{DistrictNo: vote.DistrictNo,
				Description: _VOTESORDER_DISTRICT})
		}

		votes = append(votes, Vote{
			SeqNo:               no,
			VoterName:           vote.VoterName,
			IDCode:              vote.IDCode,
			KovCode:             vote.KovCode,
			ElectoralDistrictNo: d,
		})
	}
	log.Log(args.Ctx, VotesResp{Description: _VOTESORDER_VOTESRESP})
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
			return c.Error(exit.Config, StopTimeError{Err: err, Description: _VOTESORDER_STOP},
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
			return c.Error(exit.Config, ServerConfError{Err: err, Description: _VOTESORDER_SERVER},
				"failed to configure server:", err)
		}
	}

	// Start listening for incoming connections during the voting period.
	if c.Until >= command.Execute {
		if err = s.Serve(c.Ctx); err != nil {
			return c.Error(exit.Unavailable, ServeError{Err: err, Description: _VOTESORDER_SERVER_SERVE},
				"failed to serve votesorder service:", err)
		}
	}
	return exit.OK
}
