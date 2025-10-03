/*
The smartid service performs Smart-ID authentication and intermediates requests
for Smart-ID signing using the Smart-ID REST API.
*/
package main

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"os"
	"time"

	"ivxv.ee/common/collector/auth"
	"ivxv.ee/common/collector/auth/ticket"
	"ivxv.ee/common/collector/command"
	"ivxv.ee/common/collector/command/exit"
	"ivxv.ee/common/collector/conf"
	"ivxv.ee/common/collector/errors"
	"ivxv.ee/common/collector/identity"
	"ivxv.ee/common/collector/log"
	"ivxv.ee/common/collector/server"
	"ivxv.ee/common/collector/smartid"
	"ivxv.ee/common/collector/status/client"
	status "ivxv.ee/common/collector/status/client/rpc"
	internal "ivxv.ee/smartid/internal/sessionstatus/rpc"
	//ivxv:modules common/collector/container
)

const (
	// StatusPoll is returned as Status from AuthenticateStatus and
	// SignStatus if a Smart-ID session has not yet finished and the
	// client needs to poll again.
	StatusPoll = "POLL"

	// StatusOK is returned as Status from AuthenticateStatus and
	// SignStatus if a Smart-ID session has finished successfully.
	StatusOK = "OK"
)

// smartidToServerError maps an ivxv.ee/common/collector/smartid package error to an ivxv.ee/common/collector/server
// error to return to the client. It returns nil if err is not a recognized smartid
// error, i.e., it was caused by an internal server error.
func smartidToServerError(err error) error {
	switch {
	case errors.CausedBy(err, new(smartid.InputError)) != nil:
		return server.ErrBadRequest
	case errors.CausedBy(err, new(smartid.VerificationError)) != nil:
		return server.ErrSmartIDVerification
	case errors.CausedBy(err, new(smartid.AccountError)) != nil:
		return server.ErrSmartIDAccount
	case errors.CausedBy(err, new(smartid.CanceledError)) != nil:
		return server.ErrSmartIDCanceled
	case errors.CausedBy(err, new(smartid.ExpiredError)) != nil:
		return server.ErrSmartIDExpired
	case errors.CausedBy(err, new(smartid.CertificateError)) != nil:
		return server.ErrSmartIDCertificate
	case errors.CausedBy(err, new(smartid.StatusError)) != nil:
		return server.ErrSmartIDGeneral
	}
	return nil
}

// RPC is the handler for Smart-ID service calls.
type RPC struct {
	status   client.Verifier
	authEnd  time.Time
	smartid  *smartid.Client
	ticket   *ticket.T
	identify identity.Identifier

	sessionTimeout time.Duration
}

type ChallengeArgs struct {
	server.Header
}

type ChallengeResponse struct {
	server.Header
	Challenge    []byte
	XSmartIDAuth []byte
}

type challengeCookie struct {
	Challenge []byte
	SessionID string
	ExpiresAt time.Time
}

// Challenge is the remote procedure call performed by clients to generate a Smart-ID verification code.
func (r *RPC) Challenge(args ChallengeArgs, resp *ChallengeResponse) (err error) {
	log.Log(args.Ctx, ChallengeReq{Description: _SMARTID_CHALREQ})

	if !time.Now().Before(r.authEnd) {
		log.Log(args.Ctx, ChallengeVotingEnded{Description: _SMARTID_EXPIRED})
		return server.ErrVotingEnd
	}

	// Build up VerifyReq for session status service
	verifyReq := status.NewVerifyReqBuilder().
		WithServiceMethod(internal.Challenge).
		WithRequest(args.Header).
		Build()

	// SessionID security check
	ok, err := r.status.Verify(&verifyReq)
	if err != nil {
		// Error during SessionID check - database unreachable, service stalled, etc.
		log.Error(args.Ctx, ChallengeVerifySessionIDError{Err: err, Description: _SMARTID_SESSION_ID})
		return server.ErrBadRequest
	}
	if !ok {
		// SessionID is unknown / has expired, we shall not further process the request
		log.Error(args.Ctx, ChallengeUpdateSessionIDError{Description: _SMARTID_SESSION_ID_EXPIRED})
		return server.ErrBadRequest
	}

	challenge, challengeDigest, err := r.smartid.Challenge()
	if err != nil {
		log.Error(args.Ctx, ChallengeError{Err: err, Description: _SMARTID_CHAL})
		return server.ErrInternal
	}

	b, err := asn1.Marshal(challengeCookie{
		Challenge: challenge,
		SessionID: args.SessionID,
		ExpiresAt: time.Now().Add(r.sessionTimeout).UTC(),
	})
	if err != nil {
		log.Error(args.Ctx, ChallengeCookieMarshalError{Err: err, Description: _SMARTID_CHAL_COOKIE_MARSHAL})
		return server.ErrInternal
	}

	cookie, err := r.ticket.CreateData(b)
	if err != nil {
		log.Error(args.Ctx, ChallengeCookieError{Err: err, Description: _SMARTID_CHAL_COOKIE})
		return server.ErrInternal
	}

	resp.Challenge = challengeDigest
	resp.XSmartIDAuth = cookie

	log.Log(args.Ctx, ChallengeResp{
		Challenge:   resp.Challenge,
		Description: _SMARTID_CHALRESP,
	})
	return nil
}

// AuthArgs are the arguments provided to a call of RPC.Authenticate.
type AuthArgs struct {
	server.Header
	Identifier string `size:"11"`
}

// AuthResponse is the response returned by RPC.Authenticate.
type AuthResponse struct {
	server.Header
	SessionCode string
}

// Authenticate is the remote procedure call performed by clients to start a
// Smart-ID authentication session.
func (r *RPC) Authenticate(args AuthArgs, resp *AuthResponse) (err error) {
	log.Log(args.Ctx, AuthenticateReq{Identifier: args.Identifier, Description: _SMARTID_AUTHREQ})

	// The server filter for voting end will only enable once we stop
	// serving signing requests, so we must manually check if we should
	// still serve authentication requests and refuse those requests when
	// not served anymore
	if !time.Now().Before(r.authEnd) { // not before == equal or after
		log.Log(args.Ctx, AuthenticateVotingEnded{Description: _SMARTID_EXPIRED})
		return server.ErrVotingEnd
	}

	// Build up VerifyReq for session status service
	verifyReq := status.NewVerifyReqBuilder().
		WithServiceMethod(internal.Authenticate).
		WithRequest(args.Header).
		Build()

	// SessionID security check
	ok, err := r.status.Verify(&verifyReq)
	if err != nil {
		// Error during SessionID check - database unreachable, service stalled, etc.
		log.Error(args.Ctx, AuthenticateVerifySessionIDError{Err: err, Description: _SMARTID_SESSION_ID})
		return server.ErrBadRequest
	}
	if !ok {
		// SessionID is unknown / has expired, we shall not further process the request
		log.Error(args.Ctx, AuthenticateUpdateSessionIDError{Description: _SMARTID_SESSION_ID_EXPIRED})
		return server.ErrBadRequest
	}

	data, err := r.ticket.TokenData(args.XSmartIDAuth)
	if err != nil {
		log.Error(args.Ctx, AuthenticateExtractCookieError{Err: err, Description: _SMARTID_COOKIE_EXTRACT})
		return server.ErrInternal
	}
	var cookie challengeCookie
	rest, err := asn1.Unmarshal(data, &cookie)
	if err != nil || len(rest) != 0 {
		log.Error(args.Ctx, AuthenticateUnmarshalCookieError{Err: err, Description: _SMARTID_COOKIE_UNMARSHAL})
		return server.ErrBadRequest
	}
	if cookie.SessionID != args.SessionID {
		log.Error(args.Ctx, AuthenticateCookieSessionIDError{
			CookieSessionID: cookie.SessionID,
			HeaderSessionID: args.SessionID,
			Description:     _SMARTID_COOKIE_SESSIONID})
		return server.ErrBadRequest
	}
	if time.Now().UTC().After(cookie.ExpiresAt) {
		log.Debug(args.Ctx, SessionTimeout{SessionID: args.SessionID,
			Description: _SMARTID_EXPIRED_AUTH_SESSION})
		return server.ErrBadRequest
	}

	resp.SessionCode, err = r.smartid.Authenticate(args.Ctx, args.Identifier, cookie.Challenge)
	if err != nil {
		if clierr := smartidToServerError(err); clierr != nil {
			// Log known smartid service error about failed authentication
			log.Error(args.Ctx, AuthenticateSmartIDError{Err: err, Description: _SMARTID_AUTH_RESP})
			return clierr
		}
		// Log unknown smartid service error about failed authentication
		log.Error(args.Ctx, AuthenticateError{Err: log.Alert(err), Description: _SMARTID_NO_RESP})
		return server.ErrInternal
	}

	// Authentication has been successfully initiated
	log.Log(args.Ctx, AuthenticateResp{
		SessionCode: resp.SessionCode,
		Description: _SMARTID_AUTHRESP,
	})
	return nil
}

// AuthStatusArgs are the arguments provided to a call of RPC.AuthenticateStatus.
type AuthStatusArgs struct {
	server.Header
	SessionCode string `size:"36"`
}

// AuthStatusResponse is the response returned by RPC.AuthenticateStatus.
type AuthStatusResponse struct {
	server.Header
	Status       string
	GivenName    string
	Surname      string
	PersonalCode string
	AuthToken    []byte
	DataToken    []byte
}

// AuthenticateStatus is the remote procedure call performed by clients to
// check the status of a Smart-ID authentication session.
func (r *RPC) AuthenticateStatus(args AuthStatusArgs, resp *AuthStatusResponse) error {
	log.Log(args.Ctx, AuthenticateStatusReq{SessionCode: args.SessionCode, Description: _SMARTID_AUTHSTATUSREQ})

	// The server filter for voting end will only enable once we stop
	// serving signing requests, so we must manually check if we should
	// still serve authentication requests and refuse those requests when
	// not served anymore
	if !time.Now().Before(r.authEnd) { // not before == equal or after
		log.Log(args.Ctx, AuthenticateStatusVotingEnded{Description: _SMARTID_EXPIRED})
		return server.ErrVotingEnd
	}

	// Build up VerifyReq for session status service
	verifyReq := status.NewVerifyReqBuilder().
		WithServiceMethod(internal.AuthenticateStatus).
		WithRequest(args.Header).
		Build()

	// SessionID security check
	ok, err := r.status.Verify(&verifyReq)
	if err != nil {
		// Error during SessionID check - database unreachable, service stalled, etc.
		log.Error(args.Ctx, AuthenticateStatusVerifySessionIDError{Err: err, Description: _SMARTID_SESSION_ID})
		return server.ErrBadRequest
	}
	if !ok {
		// SessionID is unknown / has expired, we shall not further process the request
		log.Error(args.Ctx, AuthenticateStatusUpdateSessionIDError{Description: _SMARTID_SESSION_ID_EXPIRED})
		return server.ErrBadRequest
	}

	data, err := r.ticket.TokenData(args.XSmartIDAuth)
	if err != nil {
		log.Error(args.Ctx, AuthenticateStatusExtractCookieError{
			Err: err, Description: _SMARTID_COOKIE_EXTRACT})
		return server.ErrInternal
	}
	var cookie challengeCookie
	rest, err := asn1.Unmarshal(data, &cookie)
	if err != nil || len(rest) != 0 {
		log.Error(args.Ctx, AuthenticateStatusUnmarshalCookieError{
			Err: err, Description: _SMARTID_COOKIE_UNMARSHAL})
		return server.ErrBadRequest
	}
	if cookie.SessionID != args.SessionID {
		log.Error(args.Ctx, AuthenticateStatusCookieSessionIDError{
			CookieSessionID: cookie.SessionID,
			HeaderSessionID: args.SessionID,
			Description:     _SMARTID_COOKIE_SESSIONID})
		return server.ErrBadRequest
	}
	if time.Now().UTC().After(cookie.ExpiresAt) {
		var sessTimeout SessionTimeout
		sessTimeout.SessionID = args.SessionID
		log.Debug(args.Ctx, sessTimeout)
		return server.ErrBadRequest
	}

	documentno, cert, algorithm, signature, err := r.smartid.GetAuthenticateStatus(args.Ctx, args.SessionCode)
	if err != nil {
		if clierr := smartidToServerError(err); clierr != nil {
			// Log known smartid service error about failed authentication
			log.Error(args.Ctx, AuthenticateStatusSmartIDError{Err: err, Description: _SMARTID_AUTH_RESP})
			return clierr
		}
		// Log unknown smartid service error about failed authentication
		log.Error(args.Ctx, AuthenticateStatusError{Err: log.Alert(err), Description: _SMARTID_NO_RESP})
		return server.ErrInternal
	}

	resp.Status = StatusPoll
	if len(signature) > 0 {
		// Signed authentication response received
		log.Log(args.Ctx, AuthenticationSignature{Signature: signature, Description: _SMARTID_SIG_RESP})
		if cert == nil {
			// Certificate missing from the signed response
			log.Error(args.Ctx, AuthenticationCertificateMissingError{Err: err,
				Description: _SMARTID_NO_CERT_RESP})
			return server.ErrSmartIDGeneral
		}
		// Log detected certificate
		log.Log(args.Ctx, AuthenticationCertificate{Certificate: cert, Description: _SMARTID_CERT_RESP})

		if err = smartid.VerifyAuthenticationSignature(
			cert, algorithm, cookie.Challenge, signature); err != nil {
			// Error in verifying the signed response
			log.Error(args.Ctx, AuthenticationSignatureError{Err: err, Description: _SMARTID_VERIFY_SIG})
			return server.ErrSmartIDGeneral
		}

		resp.Status = StatusOK
		resp.GivenName = findName(&cert.Subject, asn1.ObjectIdentifier{2, 5, 4, 42})
		resp.Surname = findName(&cert.Subject, asn1.ObjectIdentifier{2, 5, 4, 4})
		if resp.PersonalCode, err = r.identify(&cert.Subject); err != nil {
			// Backend could not extract voter's personal code from certificate's Subject field
			log.Error(args.Ctx, AuthenticationSubjectIdentityError{Err: err,
				Description: _SMARTID_VOTER_ID})
			return server.ErrInternal
		}

		if resp.AuthToken, err = r.ticket.Create(cert.Subject); err != nil {
			// Error in authentication ticket creation
			log.Error(args.Ctx, AuthenticationTicketError{Err: err,
				Description: _SMARTID_AUTH_TOKEN_CREATE})
			return server.ErrInternal
		}

		if resp.DataToken, err = r.ticket.CreateData([]byte(documentno)); err != nil {
			// Error in creating the ticket
			log.Error(args.Ctx, DataTicketError{Err: err, Description: _SMARTID_DATA_TOKEN_CREATE})
			return server.ErrInternal
		}
	}

	// Successful AuthenticateStatusResp, AuthToken is treated as sensitive
	log.Log(args.Ctx, AuthenticateStatusResp{
		Status:       resp.Status,
		GivenName:    resp.GivenName,
		Surname:      resp.Surname,
		PersonalCode: resp.PersonalCode,
		AuthToken:    log.Sensitive(resp.AuthToken),
		DataToken:    log.Sensitive(resp.DataToken),
		Description:  _SMARTID_AUTHSTATUSRESP,
	})
	return nil
}

// findName searches name for oid and returns the value for that oid or an
// empty string. Panics if the value for the oid is not a string.
func findName(name *pkix.Name, oid asn1.ObjectIdentifier) string {
	for _, n := range name.Names {
		if n.Type.Equal(oid) {
			return n.Value.(string)
		}
	}
	return ""
}

// CertificateChoiceArgs are the arguments provided to a call of RPC.GetCertificate.
type CertificateChoiceArgs struct {
	server.Header
}

// CertificateChoiceResponse is the response returned by RPC.GetCertificateChoice.
type CertificateChoiceResponse struct {
	server.Header
	SessionCode string
}

// GetCertificateChoice is the remote procedure call performed by clients to get the
// Smart-ID signing certificate choice session that will be used to sign the vote.
func (r *RPC) GetCertificateChoice(args CertificateChoiceArgs, resp *CertificateChoiceResponse) error {
	log.Log(args.Ctx, GetCertificateChoiceReq{Description: _SMARTID_GETCERTREQ})

	// Get the voter serial number. If empty, then the request is not
	// authenticated.
	personalCode := server.VoterIdentity(args.Ctx)
	if len(personalCode) == 0 {
		log.Error(args.Ctx, UnauthenticatedGetCertificateError{Description: _SMARTID_VOTER_NO_AUTH})
		return server.ErrUnauthenticated
	}

	// Build up VerifyReq for session status service
	verifyReq := status.NewVerifyReqBuilder().
		WithServiceMethod(internal.GetCertificate).
		WithRequest(args.Header).
		Build()

	// SessionID security check
	ok, err := r.status.Verify(&verifyReq)
	if err != nil {
		// Error during SessionID check - database unreachable, service stalled, etc.
		log.Error(args.Ctx, GetCertificateVerifySessionIDError{Err: err, Description: _SMARTID_SESSION_ID})
		return server.ErrBadRequest
	}
	if !ok {
		// SessionID is unknown / has expired, we shall not further process the request
		log.Error(args.Ctx, GetCertificateUpdateSessionIDError{Description: _SMARTID_SESSION_ID_EXPIRED})
		return server.ErrBadRequest
	}

	documentno := server.VoterNumber(args.Ctx)
	if len(documentno) == 0 {
		var missingDocumentNo MissingDataSignError
		missingDocumentNo.Description = _SMARTID_VOTER_NO_PHONENR
		// Could not detect documentno from the token
		log.Error(args.Ctx, missingDocumentNo)
		return server.ErrUnauthenticated
	}

	s, err := r.smartid.GetCertificateChoice(args.Ctx, documentno)
	if err != nil {
		if clierr := smartidToServerError(err); clierr != nil {
			// Log known smartid service error about failed authentication
			log.Error(args.Ctx, GetCertificateChoiceSmartIDError{Err: err, Description: _SMARTID_AUTH_RESP})
			return clierr
		}
		// Log unknown smartid service error about failed authentication
		log.Error(args.Ctx, GetCertificateChoiceError{Err: log.Alert(err), Description: _SMARTID_NO_RESP})
		return server.ErrInternal
	}

	// GetCertificateChoice successfully initiated
	log.Log(args.Ctx, GetCertificateChoiceResp{
		Session:     resp.SessionCode,
		Description: _SMARTID_GETCERTRESP,
	})
	resp.SessionCode = s
	return nil
}

// CertificateChoiceStatusArgs is the response returned by RPC.GetCertificateChoiceStatus.
type CertificateChoiceStatusArgs struct {
	server.Header
	SessionCode string
}

// CertificateChoiceStatusResponse is the response returned by RPC.GetCertificateChoiceStatus.
type CertificateChoiceStatusResponse struct {
	server.Header
	Certificate []byte
	Status      string
}

// GetCertificateChoiceStatus is the remote procedure call performed by clients to get the
// Smart-ID signing certificate that will be used to sign the vote.
func (r *RPC) GetCertificateChoiceStatus(args CertificateChoiceStatusArgs,
	resp *CertificateChoiceStatusResponse) error {
	log.Log(args.Ctx, GetCertificateChoiceStatusReq{Session: args.SessionCode,
		Description: _SMARTID_GETCERTSTATUSREQ})

	// Build up VerifyReq for session status service
	verifyReq := status.NewVerifyReqBuilder().
		WithServiceMethod(internal.GetCertificateStatus).
		WithRequest(args.Header).
		Build()

	// SessionID security check
	ok, err := r.status.Verify(&verifyReq)
	if err != nil {
		// Error during SessionID check - database unreachable, service stalled, etc.
		log.Error(args.Ctx, GetCertificateStatusVerifySessionIDError{Err: err,
			Description: _SMARTID_SESSION_ID})
		return server.ErrBadRequest
	}
	if !ok {
		// SessionID is unknown / has expired, we shall not further process the request
		log.Error(args.Ctx, GetCertificateStatusUpdateSessionIDError{Description: _SMARTID_SESSION_ID_EXPIRED})
		return server.ErrBadRequest
	}

	_, c, err := r.smartid.GetCertificateChoiceStatus(args.Ctx, args.SessionCode)
	if err != nil {
		if clierr := smartidToServerError(err); clierr != nil {
			// Log known smartid service error about failed authentication
			log.Error(args.Ctx, GetCertificateChoiceStatusSmartIDError{Err: err,
				Description: _SMARTID_AUTH_RESP})
			return clierr
		}
		// Log unknown smartid service error about failed authentication
		log.Error(args.Ctx, GetCertificateChoiceStatusError{Err: log.Alert(err),
			Description: _SMARTID_NO_RESP})
		return server.ErrInternal
	}
	// Signing certificate successfully retrieved
	log.Log(args.Ctx, SigningCertificate{Certificate: c, Description: _SMARTID_CERT})

	resp.Status = StatusPoll
	if c != nil {
		resp.Status = StatusOK
		resp.Certificate = c.Raw
	}

	// Log successful GetCertificateResp
	log.Log(args.Ctx, GetCertificateResp{
		Certificate: resp.Certificate,
		Status:      resp.Status,
		Description: _SMARTID_GETCERTSTATUSRESP,
	})
	return nil
}

// SignArgs are the arguments provided to a call of RPC.Sign.
type SignArgs struct {
	server.Header
	Hash     []byte `size:"64"` // Hash of the data to sign.
	HashType string `size:"10"` // Allowed values described in 'ivxv.ee/common/collector/smartid' package.
}

// SignResponse is the response returned by RPC.Sign.
type SignResponse struct {
	server.Header
	SessionCode string
}

// Sign is the remote procedure call performed by clients to start a Smart-ID
// signing session.
func (r *RPC) Sign(args SignArgs, resp *SignResponse) (err error) {
	log.Log(args.Ctx, SignReq{HashType: args.HashType, Hash: args.Hash,
		Description: _SMARTID_SIGNREQ})

	// Get the voter serial number. If empty, then the request is not
	// authenticated.
	identity := server.VoterIdentity(args.Ctx)
	if len(identity) == 0 {
		log.Error(args.Ctx, UnauthenticatedSignError{Description: _SMARTID_VOTER_NO_AUTH})
		return server.ErrUnauthenticated
	}

	// Build up VerifyReq for session status service
	verifyReq := status.NewVerifyReqBuilder().
		WithServiceMethod(internal.Sign).
		WithRequest(args.Header).
		Build()

	// SessionID security check
	ok, err := r.status.Verify(&verifyReq)
	if err != nil {
		// Error during SessionID check - database unreachable, service stalled, etc.
		log.Error(args.Ctx, SignVerifySessionIDError{Err: err, Description: _SMARTID_SESSION_ID})
		return server.ErrBadRequest
	}
	if !ok {
		// SessionID is unknown / has expired, we shall not further process the request
		log.Error(args.Ctx, SignUpdateSessionIDError{Description: _SMARTID_SESSION_ID_EXPIRED})
		return server.ErrBadRequest
	}

	documentno := server.VoterNumber(args.Ctx)
	if len(documentno) == 0 {
		// Could not detect documentno from the token
		log.Error(args.Ctx, MissingDataSignError{Description: _SMARTID_VOTER_NO_PHONENR})
		return server.ErrUnauthenticated
	}
	resp.SessionCode, err = r.smartid.SignHash(
		args.Ctx, documentno, args.Hash, args.HashType)
	if err != nil {
		if clierr := smartidToServerError(err); clierr != nil {
			// Log known smartid service error about failed authentication
			log.Error(args.Ctx, SignSmartIDError{Err: err, Description: _SMARTID_AUTH_RESP})
			return clierr
		}
		// Log unknown smartid service error about failed authentication
		log.Error(args.Ctx, SignError{Err: log.Alert(err), Description: _SMARTID_NO_RESP})
		return server.ErrInternal
	}

	// Signing successfully initiated

	log.Log(args.Ctx, SignResp{
		SessionCode: resp.SessionCode,
		Description: _SMARTID_SIGNRESP,
	})
	return
}

// SignStatusArgs are the arguments provided to a call of RPC.SignStatus.
type SignStatusArgs struct {
	server.Header
	SessionCode string `size:"36"`
}

// SignStatusResponse is the response returned by RPC.SignStatus.
type SignStatusResponse struct {
	server.Header
	Status    string
	Signature []byte
	//nolint: lll
	Algorithm string // The signature algorithm. Allowed values described in 'ivxv.ee/common/collector/smartid' package.
}

// SignStatus is the remote procedure call performed by clients to check the
// status of a Smart-ID signing session.
func (r *RPC) SignStatus(args SignStatusArgs, resp *SignStatusResponse) (err error) {
	log.Log(args.Ctx, SignStatusReq{SessionCode: args.SessionCode,
		Description: _SMARTID_SIGNSTATUSREQ})

	// Build up VerifyReq for session status service
	verifyReq := status.NewVerifyReqBuilder().
		WithServiceMethod(internal.SignStatus).
		WithRequest(args.Header).
		Build()

	// SessionID security check
	ok, err := r.status.Verify(&verifyReq)
	if err != nil {
		// Error during SessionID check - database unreachable, service stalled, etc.
		log.Error(args.Ctx, SignStatusVerifySessionIDError{Err: err,
			Description: _SMARTID_SESSION_ID})
		return server.ErrBadRequest
	}
	if !ok {
		// SessionID is unknown / has expired, we shall not further process the request
		log.Error(args.Ctx, SignStatusUpdateSessionIDError{Description: _SMARTID_SESSION_ID_EXPIRED})
		return server.ErrBadRequest
	}

	resp.Algorithm, resp.Signature, err = r.smartid.GetSignHashStatus(args.Ctx, args.SessionCode)
	if err != nil {
		if clierr := smartidToServerError(err); clierr != nil {
			// Log known smartid service error about failed authentication
			log.Error(args.Ctx, SignStatusSmartIDError{Err: err, Description: _SMARTID_AUTH_RESP})
			return clierr
		}
		// Log unknown smartid service error about failed authentication
		log.Error(args.Ctx, SignStatusError{Err: log.Alert(err), Description: _SMARTID_NO_RESP})
		return server.ErrInternal
	}

	resp.Status = StatusPoll
	if len(resp.Signature) > 0 {
		resp.Status = StatusOK
	}

	// Signed response successfully retrieved

	log.Log(args.Ctx, SignStatusResp{
		Status:      resp.Status,
		Signature:   resp.Signature,
		Description: _SMARTID_SIGNSTATUSRESP,
	})
	return
}

func main() {
	// Call smartidmain in a separate function so that it can set up defers
	// and have them trigger before returning with a non-zero exit code.
	os.Exit(smartidmain())
}

func smartidmain() (code int) {
	c := command.NewWithoutStorage("ivxv-smartid", "")
	defer func() {
		code = c.Cleanup(code)
	}()

	// Configure session status client
	statusClient, errCode := internal.NewClient(c)
	if statusClient == nil || errCode != 0 {
		return errCode
	}

	// Create new RPC instance and start the session cleaner.
	rpc := &RPC{sessionTimeout: time.Minute * 5, status: statusClient}

	var start, stop time.Time
	var authConf server.AuthConf
	var err error

	if c.Conf.Election != nil {
		// Check election configuration time values - service start
		if start, err = c.Conf.Election.ServiceStartTime(); err != nil {
			return c.Error(exit.Config, StartTimeError{Err: err, Description: _SMARTID_START},
				"bad service start time:", err)
		}

		// Check election configuration time values - election stop
		if rpc.authEnd, err = c.Conf.Election.ElectionStopTime(); err != nil {
			return c.Error(exit.Config, ElectionStopTimeError{Err: err, Description: _SMARTID_AUTH_STOP},
				"bad election stop time:", err)
		}

		// Check election configuration time values - service stop
		if stop, err = c.Conf.Election.ServiceStopTime(); err != nil {
			return c.Error(exit.Config, ServiceStopTimeError{Err: err, Description: _SMARTID_STOP},
				"bad service stop time:", err)
		}

		// Configure the Smart-ID REST API client.
		if rpc.smartid, err = smartid.New(&c.Conf.Election.SmartID); err != nil {
			return c.Error(exit.Config, SmartIDConfError{Err: err, Description: _SMARTID_CLIENT_CONF},
				"failed to configure SmartID-REST API client:", err)
		}

		// Configure the ticket manager for issuing authentication
		// tickets.
		ticketConf, ok := c.Conf.Election.Auth[auth.Ticket]
		if !ok {
			return c.Error(exit.Config, TicketAuthError{Description: _SMARTID_TICKET_AUTH},
				"ticket authentication is mandatory for smartid")
		}
		if rpc.ticket, err = ticket.NewFromSystem(); err != nil {
			return c.Error(exit.Config, TicketConfError{Err: err, Description: _SMARTID_TICKET},
				"failed to configure ticket manager:", err)
		}

		// Parse configuration for authenticating with tickets issued
		// by this server.
		if authConf, err = server.NewAuthConf(auth.Conf{auth.Ticket: ticketConf},
			c.Conf.Election.Identity, nil); err != nil {

			return c.Error(exit.Config, ServerAuthConfError{Err: err, Description: _SMARTID_AUTH},
				"failed to configure client authentication:", err)
		}

		// Store the voter identifier for signer identification.
		rpc.identify = authConf.Identity
	}

	var s *server.S
	if c.Conf.Technical != nil {
		// Configure a new server with the service instance
		// configuration and the RPC handler instance.
		cert, key := conf.TLS(conf.Sensitive(c.Service.ID))
		if s, err = server.New(&server.Conf{
			CertPath: cert,
			KeyPath:  key,
			Address:  c.Service.Address,
			End:      stop,
			Filter:   &c.Conf.Technical.Filter,
			Version:  &c.Conf.Version,
		}, rpc); err != nil {
			return c.Error(exit.Config, ServerConfError{Err: err, Description: _SMARTID_SERVER},
				"failed to configure server:", err)
		}
	}

	// Start listening for incoming connections during the voting period.
	if c.Until >= command.Execute {
		if err = s.WithAuth(authConf).ServeAt(c.Ctx, start); err != nil {
			return c.Error(exit.Unavailable, ServeError{Err: err, Description: _SMARTID_SERVER_SERVE},
				"failed to serve smartid service:", err)
		}
	}
	return exit.OK
}
