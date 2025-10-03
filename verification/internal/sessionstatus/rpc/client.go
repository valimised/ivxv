package rpc

import (
	"strconv"

	"ivxv.ee/common/collector/command"
	"ivxv.ee/common/collector/server"
	"ivxv.ee/common/collector/status/client"
	status "ivxv.ee/common/collector/status/client/rpc"
	api "ivxv.ee/sessionstatus/api/rpc"
)

const (
	// This should be a StatusReadResp.Caller value when calling RPC.Verify
	// for a first time.
	Vote = "RPC.Vote"

	// This should be a StatusReadResp.Caller value when calling RPC.Verify
	// for a next time.
	Verify = "RPC.Verify"
)

const exitCodeOK = 0

type RPC struct {
	verifyTTL int64
	client    client.TLSDialer
}

// NewClient initializes session status server client.
func NewClient(c *command.C) (client.Verifier, int) {
	// Initialize RPC TLS session status client
	tlsDialer, errCode := api.NewClient(c)
	if errCode != exitCodeOK {
		return nil, errCode
	}

	return &RPC{
		client:    tlsDialer,
		verifyTTL: c.Conf.Technical.Status.Session.VerifyTTL,
	}, exitCodeOK
}

func (r *RPC) Verify(dto interface{}) (bool, error) {
	// dto should cast to *status.VerifyReq
	verifyReq, err := status.CastAnyToVerifyReq(dto)
	if err != nil {
		return false, CastAnyToVerifyReqError{Err: err, Description: _SESSIONSTATUS_CAST_ANY_TO_VERIFYREQ}
	}

	// verifyReq.Request should cast to server.Header
	header, err := api.CastVerifyRequestToServerHeader(verifyReq)
	if err != nil {
		return false, CastVerifyRequestToServerHeaderError{Err: err,
			Description: _SESSIONSTATUS_CAST_VERIFYREQ_TO_SERVERHEADER}
	}

	// Send request to session status server and verify response
	ok, err := r.verifyAndUpdateSessionStatus(verifyReq.ServiceMethod, *header)
	if err != nil {
		return false, VerifyAndUpdateSessionStatusError{Err: err,
			Description: _SESSIONSTATUS_SEND_UPDATE_REQ_AND_VERIFY_IT}
	}

	return ok, nil
}

// verifyAndUpdateSessionStatus will first check h.Header.SessionID
// record against the underlying storage, and if everything is correct,
// then will update h.Header.SessionID record in the underlying storage
// by marking session status Caller as serviceMethod.
//
// Note, that here serviceMethod is the RPC method that calls this function.
func (r *RPC) verifyAndUpdateSessionStatus(serviceMethod string, h server.Header) (bool, error) {
	// Create new session read status request
	reqRead := api.NewSessionStatusReadReqBuilder().
		WithHeader(h).
		Build()

	// Create new RPC request to status server, embeds session status request
	reqReadRPC := status.NewStatusReqBuilder().
		WithServiceMethod(api.Endpoint.SessionStatusRead).
		WithRequest(reqRead).
		Build()

	// RPC call to .WithServiceMethod(...)
	respReadRaw, err := r.client.TLSDial(&reqReadRPC)
	if err != nil {
		return false, SessionReadReqTLSDialError{Err: err, Description: _SESSIONSTATUS_TLS_DIAL}
	}

	// Process raw RPC response, doesn't care about the embedded status type
	respReadRPC := status.NewStatusRespBuilder().
		WithResponse(respReadRaw).
		Build()

	// Process session read status response
	respRead := api.NewSessionStatusReadRespBuilder().
		WithResponse(respReadRPC.Response).
		Build()

	// NB! Most important part, that prevents any attack on SessionID
	var ok bool
	var ttl string
	ok, err = verifyStatusReadResp(&respRead, verifyHandler)
	ttl = strconv.FormatInt(r.verifyTTL, 10)
	if !ok || err != nil {
		return ok, VerifyStatusReadRespError{Err: err, Description: _SESSIONSTATUS_RESP_VERIFY}
	}

	// Create new session update status request
	reqUpdate := api.NewSessionStatusUpdateReqBuilder().
		WithHeader(h).
		WithCaller(serviceMethod).
		WithAuth(respRead.Auth).
		WithLease(respRead.Lease).
		WithTTL(ttl).
		Build()

	// Create new RPC request to status server, embeds session status request
	reqUpdateRPC := status.NewStatusReqBuilder().
		WithServiceMethod(api.Endpoint.SessionStatusUpdate).
		WithRequest(reqUpdate).
		Build()

	// RPC call to .WithServiceMethod(...)
	respUpdateRaw, err := r.client.TLSDial(&reqUpdateRPC)
	if err != nil {
		return false, SessionUpdateReqTLSDialError{Err: err, Description: _SESSIONSTATUS_TLS_DIAL}
	}

	// Process raw RPC response, doesn't care about the embedded status type
	respUpdateRPC := status.NewStatusRespBuilder().
		WithResponse(respUpdateRaw).
		Build()

	// Process session update status response
	respUpdate := api.NewSessionStatusUpdateRespBuilder().
		WithResponse(respUpdateRPC.Response).
		Build()

	// If true, then status has been successfully updated
	ok = respUpdate.Ok
	if !ok {
		return false, SessionStatusUpdateError{
			Caller:      reqUpdate.Caller,
			Auth:        respRead.Auth,
			Description: _SESSIONSTATUS_UPDATE_FAIL,
		}
	}

	return true, nil
}

// verifyStatusReadResp r by applying an appropriate handler h.
func verifyStatusReadResp(r *api.StatusReadResp, h func(*api.StatusReadResp) (bool, error)) (bool, error) {
	return h(r)
}

// authenticateHandler performs filter operation on StatusReadResp r to
// detect invalid SessionID in a client RPC.Authenticate request.
func verifyHandler(r *api.StatusReadResp) (bool, error) {
	// When voted with ID card
	idCardAuth := r.Caller == Vote && r.Auth == client.IDcardAuth ||
		r.Caller == Verify && r.Auth == client.IDcardAuth

	// When voted with Mobile-ID
	midAuth := r.Caller == Vote && r.Auth == client.MobileIDAuth ||
		r.Caller == Verify && r.Auth == client.MobileIDAuth

	// When voted with Smart-ID
	sidAuth := r.Caller == Vote && r.Auth == client.SmartIDAuth ||
		r.Caller == Verify && r.Auth == client.SmartIDAuth

	// When voted with Web eID
	widAuth := r.Caller == Vote && r.Auth == client.WebeIDAuth ||
		r.Caller == Verify && r.Auth == client.WebeIDAuth

	if !(idCardAuth) && !(midAuth) && !(sidAuth) && !(widAuth) {
		return false, VerifyInvalidCallerOrAuthForSessionID{
			Method:      Verify,
			Caller:      r.Caller,
			Auth:        r.Auth,
			Description: _SESSIONSTATUS_MALFORMED_SESSION_ID,
		}
	}

	return true, nil
}
