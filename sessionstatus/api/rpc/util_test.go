package rpc

import (
	"context"
	"reflect"
	"testing"

	"ivxv.ee/common/collector/server"
	client "ivxv.ee/common/collector/status/client/rpc"
)

const (
	someServiceMethod = "RPC.SomeServiceMethod"
	msgExpectNoErrors = "Expected no errors, got %v\n"
)

var goodVerifyReqs = []*client.VerifyReq{
	{
		ServiceMethod: someServiceMethod,
		Request: server.Header{
			SessionID:  "b64",
			OS:         "My perfect OS",
			AuthMethod: "tls",
		},
	},
	{
		ServiceMethod: someServiceMethod,
		Request: server.Header{
			SessionID: "b64",
		},
	},
	{
		ServiceMethod: someServiceMethod,
		Request:       server.Header{},
	},
}

var badVerifyReqs = []*client.VerifyReq{
	{
		ServiceMethod: someServiceMethod,
		Request: struct {
			Ctx        context.Context
			SessionID  string
			OS         string
			AuthMethod string
			AuthToken  []byte
			DataToken  []byte
		}{
			SessionID:  "b64",
			OS:         "My perfect OS",
			AuthMethod: "tls",
		},
	},
	{
		ServiceMethod: someServiceMethod,
		Request: struct {
			SessionID  string
			OS         string
			AuthMethod string
		}{
			SessionID:  "b64",
			OS:         "My perfect OS",
			AuthMethod: "tls",
		},
	},
	{
		ServiceMethod: someServiceMethod,
		Request:       struct{}{},
	},
	{
		ServiceMethod: someServiceMethod,
		Request:       nil,
	},
}

func TestCastVerifyRequestToServerHeader(t *testing.T) {
	for _, goodverifyReq := range goodVerifyReqs {
		header, err := CastVerifyRequestToServerHeader(goodverifyReq)
		if err != nil {
			t.Errorf(msgExpectNoErrors, err)
		}

		expectedHeader := goodverifyReq.Request.(server.Header)

		if !reflect.DeepEqual(*header, expectedHeader) {
			msg := "Expected v1 == v2, got v1: %v, v2: %v\n"
			t.Errorf(msg, *header, expectedHeader)
		}
	}

	for _, badverifyReq := range badVerifyReqs {
		header, err := CastVerifyRequestToServerHeader(badverifyReq)

		msg := "Expected CastVerifyReqToServerHeaderError, got %v\n"
		if err == nil || header != nil {
			t.Errorf(msg, err)
		}

		expected := new(CastVerifyReqToServerHeaderError)
		expected.Expected = expectedCastForServerHeader
		expected.Got = reflect.TypeOf(badverifyReq.Request)
		expected.Description = _SESSIONSTATUS_CAST_VERIFYREQ_TO_SERVERHEADER

		if !reflect.DeepEqual(*expected, err) {
			msg = "Expected %v, got %v\n"
			t.Errorf(msg, expected, err)
		}
	}
}
