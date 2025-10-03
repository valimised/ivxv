package rpc

import (
	"reflect"

	"ivxv.ee/common/collector/server"
	statusRpc "ivxv.ee/common/collector/status/client/rpc"
)

const expectedCastForServerHeader = "server.Header"

// CastVerifyRequestToServerHeader tries to cast req to server.Header.
func CastVerifyRequestToServerHeader(req *statusRpc.VerifyReq) (*server.Header, error) {
	// request.Request should cast to server.Header
	header, ok := req.Request.(server.Header)
	if !ok {
		return nil, CastVerifyReqToServerHeaderError{
			Expected:    expectedCastForServerHeader,
			Got:         reflect.TypeOf(req.Request),
			Description: _SESSIONSTATUS_CAST_VERIFYREQ_TO_SERVERHEADER,
		}
	}

	return &header, nil
}
