package rpc

import (
	"reflect"
	"testing"

	"ivxv.ee/common/collector/server"
	status "ivxv.ee/common/collector/status/client"
)

var goodRPCStatusReqs = []*StatusReq{
	{
		ServiceMethod: someServiceMethod,
		Request: struct{ Header server.Header }{
			Header: header,
		},
	},
	// Example of StatusReq with embedded StatusUpdateReq
	// of Smart-ID RPC.GetCertificate endpoint
	{
		ServiceMethod: someServiceMethod,
		Request: struct {
			Header server.Header
			Caller string
			Auth   string
		}{
			Header: header,
			Caller: someCallerServiceMethod,
			Auth:   status.SmartIDAuth,
		},
	},
}

var badRPCStatusReqs = []any{
	// not a pointer!
	StatusReq{
		ServiceMethod: someServiceMethod,
		Request: struct{ Header server.Header }{
			Header: header,
		},
	},
	// not a StatusReq
	StatusResp{
		Response: nil,
	},
	// anonymous struct
	struct{ Header server.Header }{
		Header: header,
	},
	// emptiness
	nil,
}

func TestCastAnyToRPCStatusReq(t *testing.T) {
	for _, goodRPCStatusReq := range goodRPCStatusReqs {
		req, err := castAnyToStatusReq(goodRPCStatusReq)
		if req == nil || err != nil {
			msg := "Expected value %v and no error, got value: %v, error: %v\n"
			t.Errorf(msg, goodRPCStatusReq, req, err)
		}
	}

	for _, badRPCStatusReq := range badRPCStatusReqs {
		req, err := castAnyToStatusReq(badRPCStatusReq)
		if req != nil || err == nil {
			msg := "Expected value nil and error, got value: %v, error: %v\n"
			t.Errorf(msg, req, err)
		}

		// Check that error is exactly the same as expected
		expected := new(CastToStatusReqError)
		expected.Expected = expectedCastForStatusReq
		expected.Got = reflect.TypeOf(badRPCStatusReq)
		expected.Description = _RPC_ANY_TO_STATUSREQ

		if !reflect.DeepEqual(err, *expected) {
			msg := "Expected %v, got: %v¸\n"
			t.Errorf(msg, expected, err)
		}
	}
}
