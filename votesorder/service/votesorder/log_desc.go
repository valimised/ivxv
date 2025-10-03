package main

const (
	_VOTESORDER_VOTESSEQREQ  = "RPC.VotesSeqReq"
	_VOTESORDER_VOTESCOUNT   = "Cannot fetch from a database an amount of total votes given (duplicate votes included)"
	_VOTESORDER_VOTESSEQRESP = "RPC.VotesSeqResp"
	_VOTESORDER_VOTESREQ     = "RPC.VotesReq"
	_VOTESORDER_NO_SEQ       = "Trying to fetch a vote sequence nr. that doesn't exist"
	_VOTESORDER_ALL_VOTES    = "Failed to fetch from database (by a batch) all successful votes with range of [from, to)"
	_VOTESORDER_SEQ          = "Cannot parse sequence nr. to base-10 integer"
	_VOTESORDER_DISTRICT     = "Cannot parse district nr. to base-10 integer"
	_VOTESORDER_VOTESRESP    = "RPC.VotesResp"
	_VOTESORDER_STOP         = "Failed to transform service start stop from election configuration file to RFC3339 format"
	_VOTESORDER_SERVER       = "Failed to parse server configuration for votesorder service"
	_VOTESORDER_SERVER_SERVE = "Failed to serve votesorder service to clients"
)
