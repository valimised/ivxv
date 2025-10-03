package rpc

import (
	"strconv"

	"ivxv.ee/common/collector/auth"
	"ivxv.ee/common/collector/command"
	"ivxv.ee/common/collector/server"
	"ivxv.ee/common/collector/status/client"
	status "ivxv.ee/common/collector/status/client/rpc"
	api "ivxv.ee/sessionstatus/api/rpc"
)

const (
	// This should be a StatusReadResp.Caller value for ID card
	Empty = ""

	// This should be a StatusReadResp.Caller value for Mobile-ID/Smart-ID
	AuthenticateStatus = "RPC.AuthenticateStatus"

	// This should be a StatusReadResp.Caller value for Web eID
	Token = "RPC.Token"

	// This should be a StatusUpdateReq.Caller value for ID card/Mobile-ID/Smart-ID/Web eID
	VoterChoices = "RPC.VoterChoices"
)

const exitCodeOK = 0

type RPC struct {
	choiceTTL int64
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
		choiceTTL: c.Conf.Technical.Status.Session.ChoiceTTL,
	}, exitCodeOK
}

func (r *RPC) Verify(dto interface{}) (bool, error) {
	// dto should cast to *status.VerifyReq
	verifyReq, err := status.CastAnyToVerifyReq(dto)
	if err != nil {
		return false, CastAnyToVerifyReqError{Err: err,
			Description: _SESSIONSTATUS_CAST_ANY_TO_VERIFYREQ}
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
	// Extract authentication method from a header.Ctx
	authFilter, err := server.AuthMethod(h.Ctx)
	if err != nil {
		return false, AuthMethodFromCtxError{Err: err,
			Description: _SESSIONSTATUS_AUTHMETHOD_FROM_CTX}
	}
	if authFilter == "" {
		return false, AuthMethodIsEmptyError{AuthFilter: authFilter,
			Description: _SESSIONSTATUS_AUTHMETHOD_EMPTY}
	}

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
		return false, SessionReadReqTLSDialError{Err: err,
			Description: _SESSIONSTATUS_TLS_DIAL}
	}

	// Process raw RPC response, doesn't care about the embedded status type
	respReadRPC := status.NewStatusRespBuilder().
		WithResponse(respReadRaw).
		Build()

	// Process session read status response
	respRead := api.NewSessionStatusReadRespBuilder().
		WithResponse(respReadRPC.Response).
		Build()

	ttl := strconv.FormatInt(r.choiceTTL, 10)
	// Reset the LeaseID
	respRead.Lease = ""

	// NB! Most important part, that prevents any attack on SessionID
	ok, err := verifyStatusReadResp(&respRead, voterChoicesHandler)
	if err != nil || !ok {
		return false, VerifyStatusReadRespError{Err: err,
			Description: _SESSIONSTATUS_RESP_VERIFY}
	}

	// SessionID is valid, however there is one more possibility to tamper
	// a SessionID == to wait until SessionID TTL is expired and perform
	// RPC.VoterChoices query. This is the only edge case that could happen
	// since RPC.VoterChoices is the only endpoint which allows
	// empty Auth and Caller fields.
	//
	// This behaviour can be prevented if we know, how user exactly authenticated
	// in an authentication filter. For Mobile-ID/Smart-ID/Web eID it is auth.Ticket,
	// however for ID card it is auth.TLS
	if respRead.Auth == client.NoAuth && auth.Type(authFilter) != auth.TLS {
		// SessionID is attempted to tamper
		return false, EmptyAuthAndCallerForNonIDCardUserError{
			Method:      VoterChoices,
			Auth:        respRead.Auth,
			Caller:      respRead.Caller,
			AuthMethod:  authFilter,
			Description: _SESSIONSTATUS_IDCARD,
		}
	}

	// Here is 100% clear, that user is ID card
	if respRead.Auth == client.NoAuth {
		respRead.Auth = client.IDcardAuth
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
		return false, SessionUpdateReqTLSDialError{Err: err,
			Description: _SESSIONSTATUS_TLS_DIAL}
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
func verifyStatusReadResp(r *api.StatusReadResp,
	h func(*api.StatusReadResp) (bool, error)) (bool, error) {
	return h(r)
}

// voterChoicesHandler performs filter operation on StatusReadResp r to
// detect invalid SessionID in a client RPC.VoterChoices request.
func voterChoicesHandler(r *api.StatusReadResp) (bool, error) {
	// When authenticating with ID card, i.e. first interaction with IVXV
	idCardAuth := r.Caller == Empty && r.Auth == client.NoAuth

	// When authenticating with Mobile-ID, then IVXV requires that client
	// has previously interacted with IVXV using RPC.AuthenticateStatus method
	midAuth := r.Caller == AuthenticateStatus && r.Auth == client.MobileIDAuth

	// When authenticating with Smart-ID, then IVXV requires that client
	// has previously interacted with IVXV using RPC.AuthenticateStatus method
	sidAuth := r.Caller == AuthenticateStatus && r.Auth == client.SmartIDAuth

	// When authenticating with Web eID, then IVXV requires that client
	// has previously interacted with IVXV using RPC.Token method
	widAuth := r.Caller == Token && r.Auth == client.WebeIDAuth

	// All conditions must satisfy simultaneously!
	if !(idCardAuth) && !(midAuth) && !(sidAuth) && !(widAuth) {
		return false, VoterChoicesInvalidCallerOrAuthForSessionID{
			Method:      VoterChoices,
			Caller:      r.Caller,
			Auth:        r.Auth,
			Description: _SESSIONSTATUS_MALFORMED_SESSION_ID,
		}
	}
	return true, nil
}
