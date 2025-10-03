package rpc

import (
	"context"
	"encoding/base64"
	"reflect"
	"testing"

	"ivxv.ee/common/collector/server"
	api "ivxv.ee/sessionstatus/api/rpc"
)

const (
	msgExpectNoErrors        = "Expected no errors, got %v\n"
	msgExpectOneGotAnother   = "Excepted %v, got %v\n"
	msgExpectNilAndError     = "Expected value nil and error, got value: %v, error: %v\n"
	msgExpectValueAndNoError = "Expected value %v and no error, got value: %v, error: %v\n"
	someServiceMethodCaller  = "RPC.SomeServiceMethodCaller"
	somAuth                  = "SomeAuth"
)

var goodSessionStatuses = [][]byte{
	// b64("Hello" + separator + "World")
	[]byte("SGVsbG8fV29ybGQ="),
	// b64("Hello" + separator + "")
	[]byte("SGVsbG8f"),
	// b64("" + separator + "World")
	[]byte("H1dvcmxk"),
	// b64("" + separator + "")
	[]byte("Hw=="),
}

var goodSessionStatusResults = [][]string{
	{"Hello", "World"},
	{"Hello", ""},
	{"", "World"},
	{"", ""},
}

var badFormatVals = [][]byte{
	// b64("Hello" + "\xF1" + "World")
	[]byte("SGVsbG/xV29ybGQ="),
	// b64("Hello" + "\xFF" + "World")
	[]byte("SGVsbG//V29ybGQ="),
	// b64("Hello" + "\\s" + "World")
	[]byte("SGVsbG9cc1dvcmxk"),
	// b64("Hello" + "\\r" + "World")
	[]byte("SGVsbG9ccldvcmxk"),
	// b64("Hello" + " " + "World")
	[]byte("SGVsbG8gV29ybGQ="),
	// b64("Hello" + "World")
	[]byte("SGVsbG9Xb3JsZA=="),
	// b64("")
	[]byte(""),
	// b64(nil)
	[]byte(nil),
}

var badBase64Vals = [][]byte{
	[]byte("13rf34545545434343452224"),
	[]byte("X"),
	[]byte("x"),
	[]byte("Äratus!"),
}

var header = server.Header{
	// Example of any RPC request Header against IVXV backend. That kind of
	// Header will see any RPC method
	Ctx:        context.Background(),
	SessionID:  "0101e9342abab1577b8b2844d6a1d317",
	OS:         "Ubuntu Jammy 22.04 LTS",
	AuthMethod: "",
	AuthToken:  nil,
	DataToken:  nil,
}

var goodStatusReadReqs = []*api.StatusReadReq{
	{Header: header},
	{Header: server.Header{SessionID: "AAABBBCCCDDD"}},
	{Header: server.Header{SessionID: "A", AuthMethod: ""}},
	{Header: server.Header{}},
}

var badStatusReadReqs = []any{
	// not a *StatusReadReq
	&api.StatusUpdateReq{Header: server.Header{Ctx: context.Background()}},
	// not a *StatusReadReq
	api.StatusUpdateResp{Header: server.Header{SessionID: "AAABBBCCCDDD"}},
	// not a pointer!
	api.StatusReadReq{Header: server.Header{SessionID: "AAABBBCCCDDD"}},
	// emptiness
	nil,
}

var goodStatusReadResps = []*api.StatusReadResp{
	{
		Header: server.Header{Ctx: context.Background()},
		Caller: someServiceMethodCaller,
		Auth:   somAuth,
	},
	{
		Header: server.Header{Ctx: context.Background()},
		Caller: "",
		Auth:   "",
	},
	{
		Header: server.Header{Ctx: context.Background()},
	},
	{
		Header: server.Header{Ctx: context.Background()},
		Auth:   somAuth,
	},
}

var badStatusReadResps = []any{
	&api.StatusReadReq{Header: server.Header{Ctx: context.Background()}},
	api.StatusReadReq{Header: server.Header{SessionID: "AAABBBCCCDDD"}},
	// anonymous struct
	struct{ hello string }{hello: "world"},
	// not a pointer!
	api.StatusReadResp{Header: server.Header{SessionID: "AAABBBCCCDDD"}},
	nil,
}

var goodStatusUpdateReqs = []*api.StatusUpdateReq{
	{
		Header: header,
		Caller: someServiceMethodCaller,
		Auth:   somAuth,
		Lease:  "0",
	},
	{
		Header: header,
		Caller: someServiceMethodCaller,
		Auth:   somAuth,
		// int64 = 768699984047453265
		Lease: "aaaf8900fff3451",
	},
}

var badStatusUpdateReqs = []any{
	// not a *StatusUpdateReq
	&api.StatusReadReq{Header: header},
	// not a *StatusUpdateReq
	api.StatusReadReq{Header: header},
	// anonymous struct
	struct{ hello string }{hello: "world"},
	// not a pointer!
	api.StatusUpdateReq{Header: server.Header{SessionID: ""}},
	nil,
}

var goodStatusDeleteReqs = []*api.StatusDeleteReq{
	{Header: header},
	{Header: server.Header{SessionID: "AAABBBCCCDDD"}},
	{Header: server.Header{SessionID: "A", AuthMethod: ""}},
	{Header: server.Header{}},
}

var badStatusDeleteReqs = []any{
	// not a *StatusDeleteReq
	&api.StatusUpdateReq{Header: server.Header{Ctx: context.Background()}},
	// not a *StatusDeleteReq
	api.StatusUpdateResp{Header: server.Header{SessionID: "AAABBBCCCDDD"}},
	// not a pointer!
	api.StatusDeleteReq{Header: server.Header{SessionID: "AAABBBCCCDDD"}},
	// emptiness
	nil,
}

func TestParseSessionStatus(t *testing.T) {
	// Only good values from a database
	for i, goodSessionStatus := range goodSessionStatuses {
		parsed, err := parseSessionStatus(goodSessionStatus)
		if err != nil {
			t.Errorf(msgExpectNoErrors, err)
		}
		if parsed[0] != goodSessionStatusResults[i][0] {
			t.Errorf(msgExpectOneGotAnother, goodSessionStatusResults[0], parsed[0])
		}
		if parsed[1] != goodSessionStatusResults[i][1] {
			t.Errorf(msgExpectOneGotAnother, goodSessionStatusResults[1], parsed[1])
		}
	}

	// Correctly base64 encoded but incorrectly formatted values
	for _, badFormatVal := range badFormatVals {
		parsed, err := parseSessionStatus(badFormatVal)
		if err == nil || parsed != nil {
			msg := "Excepted error, got %v\n"
			t.Errorf(msg, err)
		}
		b64, err := base64.StdEncoding.DecodeString(string(badFormatVal))
		if err != nil {
			msg := "Excepted value %v to be base64 decodable, got %v\n"
			t.Errorf(msg, string(badFormatVal), err)
		}
		expected := new(InvalidReadStatusDatabaseRecordCountError)
		expected.Expected = statusReadRespDBRecordCount
		expected.Got = 1
		expected.Record = b64
		expected.Description = _SESSIONSTATUS_PARSE
		if reflect.DeepEqual(err, *expected) {
			t.Errorf(msgExpectOneGotAnother, expected, err)
		}
	}

	// Incorrectly base64 encoded values
	for _, badBase64Val := range badBase64Vals {
		r, err := parseSessionStatus(badBase64Val)
		if err == nil || r != nil {
			t.Errorf(msgExpectNilAndError, r, err)
		}
	}
}

func TestCastAnyToSessionStatusReadReq(t *testing.T) {
	// Only good values from a database Read are allowed
	for _, goodStatusReadReq := range goodStatusReadReqs {
		parsed, err := castAnyToSessionStatusReadReq(goodStatusReadReq)
		if parsed == nil || err != nil {
			t.Errorf(msgExpectValueAndNoError, goodStatusReadReq, parsed, err)
		}
	}

	// Bad values from a database Read
	for _, badStatusReadReq := range badStatusReadReqs {
		parsed, err := castAnyToSessionStatusReadReq(badStatusReadReq)
		if parsed != nil || err == nil {
			t.Errorf(msgExpectNilAndError, parsed, err)
		}
		expected := new(CastToStatusReadReqError)
		expected.Expected = expectedCastForStatusReadReq
		expected.Got = reflect.TypeOf(badStatusReadReq)
		expected.Description = _SESSIONSTATUS_CAST_ANY_TO_SESSSTATUSREADREQ
		if !reflect.DeepEqual(err, *expected) {
			t.Errorf(msgExpectOneGotAnother, expected, err)
		}
	}
}

func TestCastAnyToSessionStatusReadResp(t *testing.T) {
	// Only good values from a database Read are allowed
	for _, goodStatusReadResp := range goodStatusReadResps {
		parsed, err := castAnyToSessionStatusReadResp(goodStatusReadResp)
		if parsed == nil || err != nil {
			t.Errorf(msgExpectValueAndNoError, goodStatusReadResp, parsed, err)
		}
	}

	// Bad values from a database Read
	for _, badStatusReadResp := range badStatusReadResps {
		parsed, err := castAnyToSessionStatusReadResp(badStatusReadResp)
		if parsed != nil || err == nil {
			t.Errorf(msgExpectNilAndError, parsed, err)
		}

		expected := new(CastToStatusReadRespError)
		expected.Expected = expectedCastForStatusReadResp
		expected.Got = reflect.TypeOf(badStatusReadResp)
		expected.Description = _SESSIONSTATUS_CAST_ANY_TO_SESSSTATREADRESP
		if !reflect.DeepEqual(err, *expected) {
			t.Errorf(msgExpectOneGotAnother, expected, err)
		}
	}
}

func TestCastAnyToSessionStatusUpdateReq(t *testing.T) {
	// Only good values from a database Update are allowed
	for _, goodStatusUpdateReq := range goodStatusUpdateReqs {
		parsed, err := castAnyToSessionStatusUpdateReq(goodStatusUpdateReq)
		if parsed == nil || err != nil {
			t.Errorf(msgExpectValueAndNoError, goodStatusUpdateReq, parsed, err)
		}
	}

	// Bad values from a database Update
	for _, badStatusUpdateReq := range badStatusUpdateReqs {
		parsed, err := castAnyToSessionStatusUpdateReq(badStatusUpdateReq)
		if parsed != nil || err == nil {
			t.Errorf(msgExpectNilAndError, parsed, err)
		}
		expected := new(CastToStatusUpdateReqError)
		expected.Expected = expectedCastForStatusUpdateReq
		expected.Got = reflect.TypeOf(badStatusUpdateReq)
		expected.Description = _SESSIONSTATUS_CAST_ANY_TO_SESSSTATUSUPDATEDREQ
		if !reflect.DeepEqual(err, *expected) {
			t.Errorf(msgExpectOneGotAnother, expected, err)
		}
	}
}

func TestCastAnyToSessionStatusDeleteReq(t *testing.T) {
	// Only good values from a database Delete are allowed
	for _, goodStatusDeleteReq := range goodStatusDeleteReqs {
		parsed, err := castAnyToSessionStatusDeleteReq(goodStatusDeleteReq)
		if parsed == nil || err != nil {
			t.Errorf(msgExpectValueAndNoError, goodStatusDeleteReq, parsed, err)
		}
	}

	// Bad values from a database Delete
	for _, badStatusDeleteReq := range badStatusDeleteReqs {
		parsed, err := castAnyToSessionStatusDeleteReq(badStatusDeleteReq)
		if parsed != nil || err == nil {
			t.Errorf(msgExpectNilAndError, parsed, err)
		}
		expected := new(CastToStatusDeleteReqError)
		expected.Expected = expectedCastForStatusDeleteReq
		expected.Got = reflect.TypeOf(badStatusDeleteReq)
		expected.Description = _SESSIONSTATUS_CAST_ANY_TO_SESSSTATUSDELETEDREQ
		if !reflect.DeepEqual(err, *expected) {
			t.Errorf(msgExpectOneGotAnother, expected, err)
		}
	}
}

func TestToSessionStorageKey(t *testing.T) {
	expected := "/session/123456789abc"
	got := toSessionStorageKey("123456789abc")
	if expected != got {
		msg := "Expected s1 == s2, got s1: %v, s2: %v\n"
		t.Errorf(msg, expected, got)
	}
}
