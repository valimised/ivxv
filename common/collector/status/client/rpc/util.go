package rpc

import "reflect"

const (
	expectedCastForStatusReq = "*rpc.StatusReq"
	expectedCastForVerifyReq = "*rpc.VerifyReq"
)

// castAnyToStatusReq tries to cast request req to *StatusReq.
func castAnyToStatusReq(req interface{}) (*StatusReq, error) {
	// Cast req to *StatusReq
	statusReq, ok := req.(*StatusReq)
	if !ok {
		return nil, CastToStatusReqError{
			Expected:    expectedCastForStatusReq,
			Got:         reflect.TypeOf(req),
			Description: _RPC_ANY_TO_STATUSREQ,
		}
	}

	return statusReq, nil
}

// CastAnyToVerifyReq tries to cast request req to *VerifyReq.
func CastAnyToVerifyReq(req interface{}) (*VerifyReq, error) {
	// Cast req to *VerifyReq
	verifyReq, ok := req.(*VerifyReq)
	if !ok {
		return nil, CastToVerifyReqError{
			Expected:    expectedCastForVerifyReq,
			Got:         reflect.TypeOf(req),
			Description: _RPC_ANY_TO_VERIFYREQ,
		}
	}

	return verifyReq, nil
}
