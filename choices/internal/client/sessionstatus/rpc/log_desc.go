package rpc

const (
	_SESSIONSTATUS_CAST_ANY_TO_VERIFYREQ          = "Cannot cast any to VerifyReq"
	_SESSIONSTATUS_CAST_VERIFYREQ_TO_SERVERHEADER = "Cannot cast VerifyReq to server.Header"
	_SESSIONSTATUS_SEND_UPDATE_REQ_AND_VERIFY_IT  = "SessionID verification request that has been sent to sessionstatus service has been failed"
	_SESSIONSTATUS_AUTHMETHOD_FROM_CTX            = "Unable to parse server.AuthMethod from context"
	_SESSIONSTATUS_AUTHMETHOD_EMPTY               = "server.AuthMethod parsed from context is empty"
	_SESSIONSTATUS_TLS_DIAL                       = "TLS dial to sessionstatus service failed"
	_SESSIONSTATUS_RESP_VERIFY                    = "Sessionstatus service response verification failed"
	_SESSIONSTATUS_IDCARD                         = "Empty server.AuthMethod is only allowed for ID card users"
	_SESSIONSTATUS_UPDATE_FAIL                    = "Sessionstatus service hasn't updated the Session ID state"
	_SESSIONSTATUS_MALFORMED_SESSION_ID           = "SessionID has been attempted to tamper"
)
