package rpc

import (
	"encoding/base64"
	"reflect"
	"strings"

	api "ivxv.ee/sessionstatus/api/rpc"
)

const (
	expectedCastForStatusReadReq   = "*api.StatusReadReq"
	expectedCastForStatusReadResp  = "*api.StatusReadResp"
	expectedCastForStatusUpdateReq = "*api.StatusUpdateReq"
	expectedCastForStatusDeleteReq = "*api.StatusDeleteReq"
)

// parseSessionStatus parses val. This function expects val to be
// base64("string" + "\x1F" + "string"), otherwise returns nil and error.
func parseSessionStatus(val []byte) ([]string, error) {
	// val is always in a form of:
	// base64("RPC.Method" + "\x1F" + "Auth")
	// Auth is "id" or "mid" or "sid" or "wid"
	sessionStatus, err := base64.StdEncoding.DecodeString(string(val))
	if err != nil {
		return nil, Base64DecodeSessionStatusError{Err: err,
			Description: _SESSIONSTATUS_PARSE_B64}
	}

	sessionStatusStr := string(sessionStatus)

	array := strings.Split(sessionStatusStr, separator)
	if len(array) != statusReadRespDBRecordCount {
		return nil, InvalidReadStatusDatabaseRecordCountError{
			Expected:    statusReadRespDBRecordCount,
			Got:         len(array),
			Record:      sessionStatusStr,
			Description: _SESSIONSTATUS_PARSE,
		}
	}

	return array, nil
}

// castAnyToSessionStatusReadReq tries to cast req to *StatusReadReq.
func castAnyToSessionStatusReadReq(req interface{}) (*api.StatusReadReq, error) {
	// Cast to *StatusReadReq
	sessionStatusReadReq, ok := req.(*api.StatusReadReq)
	if !ok {
		return nil, CastToStatusReadReqError{
			Expected:    expectedCastForStatusReadReq,
			Got:         reflect.TypeOf(req),
			Description: _SESSIONSTATUS_CAST_ANY_TO_SESSSTATUSREADREQ,
		}
	}

	return sessionStatusReadReq, nil
}

// castAnyToSessionStatusReadResp tries to cast req to *StatusReadResp.
func castAnyToSessionStatusReadResp(req interface{}) (*api.StatusReadResp, error) {
	// Cast to *StatusReadResp
	sessionStatusReadResp, ok := req.(*api.StatusReadResp)
	if !ok {
		return nil, CastToStatusReadRespError{
			Expected:    expectedCastForStatusReadResp,
			Got:         reflect.TypeOf(req),
			Description: _SESSIONSTATUS_CAST_ANY_TO_SESSSTATREADRESP,
		}
	}

	return sessionStatusReadResp, nil
}

// castAnyToSessionStatusUpdateReq tries to cast req to *StatusUpdateReq.
func castAnyToSessionStatusUpdateReq(req interface{}) (*api.StatusUpdateReq, error) {
	// Cast to *StatusUpdateReq
	sessionStatusUpdateReq, ok := req.(*api.StatusUpdateReq)
	if !ok {
		return nil, CastToStatusUpdateReqError{
			Expected:    expectedCastForStatusUpdateReq,
			Got:         reflect.TypeOf(req),
			Description: _SESSIONSTATUS_CAST_ANY_TO_SESSSTATUSUPDATEDREQ,
		}
	}

	return sessionStatusUpdateReq, nil
}

// castAnyToSessionStatusDeleteReq tries to cast req to *StatusDeleteReq.
func castAnyToSessionStatusDeleteReq(req interface{}) (*api.StatusDeleteReq, error) {
	// Cast to *StatusDeleteReq
	sessionStatusDeleteReq, ok := req.(*api.StatusDeleteReq)
	if !ok {
		return nil, CastToStatusDeleteReqError{
			Expected:    expectedCastForStatusDeleteReq,
			Got:         reflect.TypeOf(req),
			Description: _SESSIONSTATUS_CAST_ANY_TO_SESSSTATUSDELETEDREQ,
		}
	}

	return sessionStatusDeleteReq, nil
}

// toSessionStorageKey returns NoSQL repository key for a given sessionID.
func toSessionStorageKey(sessionID string) string {
	return sessionIDPrefix + "/" + sessionID
}
