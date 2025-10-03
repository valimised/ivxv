package main

const (
	_CHOICES_CHOICESREQ              = "RPC.ChoicesReq"
	_CHOICES_ADMIN_CODE              = "Unable to fetch admin code from database for this voter ID"
	_CHOICES_NO_FOR_VOTER            = "No choices for this voter ID"
	_CHOICES_FROM_DB                 = "Unable to fetch choices for this voter ID from a database"
	_CHOICES_CHOICESRESP             = "RPC.ChoicesResp"
	_CHOICES_VOTERCHOICESREQ         = "RPC.VoterChoicesReq"
	_CHOICES_VOTER_NO_AUTH           = "Voter has not been authenticated"
	_CHOICES_SESSION_ID              = "Malformed SessionID"
	_CHOICES_SESSION_ID_EXPIRED      = "SessionID has been expired"
	_CHOICES_NO_VOTER_IN_VOTERS_LIST = "No such voter ID in a voters list"
	_CHOICES_DB                      = "Successfully got choices for this voter ID from a database"
	_CHOICES_CHECK_VOTED             = "Unable to get a response from database whether this voter ID has already voted or not"
	_CHOICES_VOTERCHOICESRESP        = "RPC.VoterChoicesResp"
	_CHOICES_START                   = "Failed to transform service start time from election configuration file to RFC3339 format"
	_CHOICES_STOP                    = "Failed to transform service stop time from election configuration file to RFC3339 format"
	_CHOICES_AUTH                    = "Failed to parse authentication configuration for choices service"
	_CHOICES_SERVER                  = "Failed to parse server configuration for choices service"
	_CHOICES_SERVER_SERVE            = "Failed to serve choices service to clients"
)
