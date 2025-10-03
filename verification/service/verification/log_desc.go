package main

const (
	_VERIFICATION_VERIFYREQ          = "RPC.VerifyReq"
	_VERIFICATION_SESSION_ID         = "Malformed SessionID"
	_VERIFICATION_SESSION_ID_EXPIRED = "SessionID has been expired"
	_VERIFICATION_NO_VOTES           = " Voter hasn't voted yet"
	_VERIFICATION_NO_DB              = "Cannot fetch verification data for the voter from database"
	_VERIFICATION_LIMIT              = "Verification limit is exceeded"
	_VERIFICATION_TIMEOUT            = "Verification attempts timeout"
	_VERIFICATION_OLD_VOTE           = "Voter has tampered a request and tries to verify old vote"
	_VERIFICATION_RACE               = "Race condition, where voter tries to verify a vote in one session and gives a vote in another session"
	_VERIFICATION_VOTE_NO_DB         = "Cannot fetch vote of the voter from database"
	_VERIFICATION_VERIFYRESP         = "RPC.VerifyResp"
	_VERIFICATION_START              = "Failed to transform service start time from election configuration file to RFC3339 format"
	_VERIFICATION_STOP               = "Failed to transform verification stop time from election configuration file to RFC3339 format"
	_VERIFICATION_SERVER             = "Failed to parse server configuration for verification service"
	_VERIFICATION_SERVER_SERVE       = "Failed to serve verification service to clients"
)
