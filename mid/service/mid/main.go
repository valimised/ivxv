/*
The mid service performs Mobile-ID authentication and intermediates requests
for Mobile-ID signing using the Mobile-ID REST API.
*/
package main

import (
	"context"
	"crypto/x509/pkix"
	"encoding/asn1"
	"os"
	"sync"
	"time"

	"ivxv.ee/common/collector/auth"
	"ivxv.ee/common/collector/auth/ticket"
	"ivxv.ee/common/collector/command"
	"ivxv.ee/common/collector/command/exit"
	"ivxv.ee/common/collector/conf"
	"ivxv.ee/common/collector/errors"
	"ivxv.ee/common/collector/identity"
	"ivxv.ee/common/collector/log"
	"ivxv.ee/common/collector/mid"
	"ivxv.ee/common/collector/server"
	"ivxv.ee/common/collector/status/client"
	status "ivxv.ee/common/collector/status/client/rpc"
	internal "ivxv.ee/mid/internal/sessionstatus/rpc"
	//ivxv:modules common/collector/container
)

const (
	// StatusPoll is returned as Status from AuthenticateStatus and
	// SignStatus if a Mobile-ID session has not yet finished and the
	// client needs to poll again.
	StatusPoll = "POLL"

	// StatusOK is returned as Status from AuthenticateStatus and
	// SignStatus if a Mobile-ID session has finished successfully.
	StatusOK = "OK"
)

// midToServerError maps an ivxv.ee/common/collector/mid package error to an ivxv.ee/common/collector/server
// error to return to the client. It returns nil if err is not a recognized mid
// error, i.e., it was caused by an internal server error.
func midToServerError(err error) error {
	switch {
	case errors.CausedBy(err, new(mid.InputError)) != nil:
		return server.ErrBadRequest
	case errors.CausedBy(err, new(mid.NotMIDUserError)) != nil:
		return server.ErrMIDNotUser
	case errors.CausedBy(err, new(mid.SIMError)) != nil:
		return server.ErrMIDOperator
	case errors.CausedBy(err, new(mid.AbsentError)) != nil:
		return server.ErrMIDAbsent
	case errors.CausedBy(err, new(mid.CanceledError)) != nil:
		return server.ErrMIDCanceled
	case errors.CausedBy(err, new(mid.ExpiredError)) != nil:
		return server.ErrMIDExpired
	case errors.CausedBy(err, new(mid.CertificateError)) != nil:
		return server.ErrMIDCertificate
	case errors.CausedBy(err, new(mid.StatusError)) != nil:
		return server.ErrMIDGeneral
	}
	return nil
}

// session is an outstanding Mobile-ID authentication session.
type session struct {
	created      time.Time
	challengeRnd []byte
}

// RPC is the handler for Mobile-ID service calls.
type RPC struct {
	status   client.Verifier
	authEnd  time.Time
	mid      *mid.Client
	ticket   *ticket.T
	identify identity.Identifier

	// sessions maps session codes to outstanding authentication sessions.
	// Not used for signing sessions because we have no state for those.
	sessions    map[string]session
	sessionLock sync.Mutex
}

// startCleaner starts a separate goroutine which removes sessions older than 5
// minutes from r.sessions every minute.
func (r *RPC) startCleaner(ctx context.Context) {
	go func() {
		for range time.Tick(1 * time.Minute) {
			r.sessionLock.Lock()
			earliest := time.Now().Add(-5 * time.Minute)
			for code, sess := range r.sessions {
				if sess.created.Before(earliest) {
					// Session removed due to expiry
					log.Debug(ctx, SessionTimeout{SessionCode: code,
						Description: _MID_EXPIRED_AUTH_SESSION})
					delete(r.sessions, code)
				}
			}
			r.sessionLock.Unlock()
		}
	}()
}

// AuthArgs are the arguments provided to a call of RPC.Authenticate.
type AuthArgs struct {
	server.Header
	IDCode  string `size:"11"`
	PhoneNo string `size:"20"`
}

// AuthResponse is the response returned by RPC.Authenticate.
type AuthResponse struct {
	server.Header
	SessionCode string
	Challenge   []byte
	DataToken   []byte
}

// Authenticate is the remote procedure call performed by clients to start a
// Mobile-ID authentication session.
func (r *RPC) Authenticate(args AuthArgs, resp *AuthResponse) (err error) {
	log.Log(args.Ctx, AuthenticateReq{PhoneNo: args.PhoneNo, Description: _MID_AUTHREQ})

	// The server filter for voting end will only enable once we stop
	// serving signing requests, so we must manually check if we should
	// still serve authentication requests and refuse those requests when
	// not served anymore
	if !time.Now().Before(r.authEnd) { // not before == equal or after
		log.Log(args.Ctx, AuthenticateVotingEnded{Description: _MID_EXPIRED})
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
		log.Error(args.Ctx, AuthenticateVerifySessionIDError{Err: err, Description: _MID_SESSION_ID})
		return server.ErrBadRequest
	}
	if !ok {
		// SessionID is unknown / has expired, we shall not further process the request
		log.Error(args.Ctx, AuthenticateUpdateSessionIDError{Description: _MID_SESSION_ID_EXPIRED})
		return server.ErrBadRequest
	}

	sess := session{created: time.Now()}
	resp.SessionCode, sess.challengeRnd, resp.Challenge, err =
		r.mid.MobileAuthenticate(args.Ctx, args.IDCode, args.PhoneNo)
	if err != nil {
		if clierr := midToServerError(err); clierr != nil {
			// Log known mID service error about failed authentication
			log.Error(args.Ctx, AuthenticateMIDError{Err: err, Description: _MID_AUTH_RESP})
			return clierr
		}
		// Log unknown mID service error about failed authentication
		log.Error(args.Ctx, AuthenticateError{Err: log.Alert(err), Description: _MID_NO_RESP})
		return server.ErrInternal
	}
	if resp.DataToken, err = r.ticket.CreateData([]byte(args.PhoneNo)); err != nil {
		// Error creating data token
		log.Error(args.Ctx, DataTicketError{Err: err, Description: _MID_DATA_TOKEN_CREATE})
		return server.ErrInternal
	}
	r.sessionLock.Lock()
	defer r.sessionLock.Unlock()
	if _, ok := r.sessions[resp.SessionCode]; ok {
		// Mobile-ID service has issued a session code that has already been issued.
		// This is bug of Mobile-ID service if this ever happens
		log.Error(args.Ctx, DuplicateSessionCodeError{Code: resp.SessionCode,
			Description: _MID_SAME_AUTH_SESSION})
		return server.ErrInternal
	}
	r.sessions[resp.SessionCode] = sess

	// Authentication process successfully initiated
	log.Log(args.Ctx, AuthenticateResp{
		SessionCode: resp.SessionCode,
		Challenge:   resp.Challenge,
		Description: _MID_AUTHRESP,
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
}

// AuthenticateStatus is the remote procedure call performed by clients to
// check the status of a Mobile-ID authentication session.
func (r *RPC) AuthenticateStatus(args AuthStatusArgs, resp *AuthStatusResponse) error {
	log.Log(args.Ctx, AuthenticateStatusReq{SessionCode: args.SessionCode, Description: _MID_AUTHSTATUSREQ})

	// The server filter for voting end will only enable once we stop
	// serving signing requests, so we must manually check if we should
	// still serve authentication requests and refuse those requests when
	// not served anymore
	if !time.Now().Before(r.authEnd) { // not before == equal or after
		log.Log(args.Ctx, AuthenticateStatusVotingEnded{Description: _MID_EXPIRED})
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
		log.Error(args.Ctx, AuthenticateStatusVerifySessionIDError{Err: err, Description: _MID_SESSION_ID})
		return server.ErrBadRequest
	}
	if !ok {
		// SessionID is unknown / has expired, we shall not further process the request
		log.Error(args.Ctx, AuthenticateStatusUpdateSessionIDError{Description: _MID_SESSION_ID_EXPIRED})
		return server.ErrBadRequest
	}

	r.sessionLock.Lock()
	sess, ok := r.sessions[args.SessionCode]
	r.sessionLock.Unlock()
	if !ok {
		// Client has presented "RPC.AuthenticateStatus.SessionCode" argument that
		// hasn't been authenticated in RPC.Authenticate
		log.Error(args.Ctx, UnknownSessionCodeError{Description: _MID_UNKNOWN_AUTH_SESSION})
		return server.ErrBadRequest
	}

	cert, algorithm, signature, err := r.mid.GetMobileAuthenticateStatus(args.Ctx, args.SessionCode)
	if err != nil {
		r.sessionLock.Lock()
		delete(r.sessions, args.SessionCode)
		r.sessionLock.Unlock()

		if clierr := midToServerError(err); clierr != nil {
			// Log known mID service error about failed authentication
			log.Error(args.Ctx, AuthenticateStatusMIDError{Err: err, Description: _MID_AUTH_RESP})
			return clierr
		}
		// Log unknown mID service error about failed authentication
		log.Error(args.Ctx, AuthenticateStatusError{Err: log.Alert(err), Description: _MID_NO_RESP})
		return server.ErrInternal
	}

	resp.Status = StatusPoll
	if len(signature) > 0 {
		// AuthenticationStatus has provided a signed result
		log.Log(args.Ctx, AuthenticationSignature{Signature: signature, Description: _MID_SIG_RESP})
		if cert == nil {
			// Signature missing from the authentication reply with signature
			log.Error(args.Ctx, AuthenticationCertificateMissingError{Err: err,
				Description: _MID_NO_CERT_RESP})
			return server.ErrMIDGeneral
		}
		// Certificate received from the signed response
		log.Log(args.Ctx, AuthenticationCertificate{Certificate: cert, Description: _MID_CERT_RESP})

		r.sessionLock.Lock()
		delete(r.sessions, args.SessionCode)
		r.sessionLock.Unlock()

		// Verify signature of authentication reply, refuse to proceed, if fails
		if err = mid.VerifyAuthenticationSignature(
			cert, algorithm, sess.challengeRnd, signature); err != nil {
			log.Error(args.Ctx, AuthenticationSignatureError{Err: err, Description: _MID_VERIFY_SIG})
			return server.ErrMIDGeneral
		}

		resp.Status = StatusOK
		resp.GivenName = findName(&cert.Subject, asn1.ObjectIdentifier{2, 5, 4, 42})
		resp.Surname = findName(&cert.Subject, asn1.ObjectIdentifier{2, 5, 4, 4})
		if resp.PersonalCode, err = r.identify(&cert.Subject); err != nil {
			// Backend cannot extract voter's personal code from certificate's Subject field
			log.Error(args.Ctx, AuthenticationSubjectIdentityError{Err: err,
				Description: _MID_VOTER_ID})
			return server.ErrInternal
		}

		if resp.AuthToken, err = r.ticket.Create(cert.Subject); err != nil {
			// Backend internal error, cannot create AES shared secret that is used as
			// authenitcation cookie in ticket-based methods
			log.Error(args.Ctx, AuthenticationTicketError{Err: err, Description: _MID_AUTH_TOKEN_CREATE})
			return server.ErrInternal
		}
	}

	// Successful authentication, log information, handle AuthToken as sensitive
	log.Log(args.Ctx, AuthenticateStatusResp{
		Status:       resp.Status,
		GivenName:    resp.GivenName,
		Surname:      resp.Surname,
		PersonalCode: resp.PersonalCode,
		AuthToken:    log.Sensitive(resp.AuthToken),
		Description:  _MID_AUTHSTATUSRESP,
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

// CertificateArgs are the arguments provided to a call of RPC.GetCertificate.
type CertificateArgs struct {
	server.Header
}

// CertificateResponse is the response returned by RPC.GetCertificate.
type CertificateResponse struct {
	server.Header
	Certificate []byte
}

// GetCertificate is the remote procedure call performed by clients to get the
// Mobile-ID signing certificate that will be used to sign the vote.
func (r *RPC) GetCertificate(args CertificateArgs, resp *CertificateResponse) error {
	log.Log(args.Ctx, GetCertificateReq{Description: _MID_GETCERTREQ})

	// Get the voter serial number. If empty, then the request is not
	// authenticated.
	identity := server.VoterIdentity(args.Ctx)
	if len(identity) == 0 {
		log.Error(args.Ctx, UnauthenticatedGetCertificateError{Description: _MID_VOTER_NO_AUTH})
		return server.ErrUnauthenticated
	}

	// Get voter phone no from the ticket, not authenticated if does not exist
	phoneno := server.VoterNumber(args.Ctx)
	if len(phoneno) == 0 {
		log.Error(args.Ctx, MissingDataGetCertificateError{Description: _MID_VOTER_NO_PHONENR})
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
		log.Error(args.Ctx, GetCertificateVerifySessionIDError{Err: err, Description: _MID_SESSION_ID})
		return server.ErrBadRequest
	}
	if !ok {
		// SessionID is unknown / has expired, we shall not further process the request
		log.Error(args.Ctx, GetCertificateUpdateSessionIDError{Description: _MID_SESSION_ID_EXPIRED})
		return server.ErrBadRequest
	}

	c, err := r.mid.GetMobileCertificate(args.Ctx, identity, phoneno)
	if err != nil {
		if clierr := midToServerError(err); clierr != nil {
			// Log known mID service error about failed authentication
			log.Error(args.Ctx, GetCertificateMIDError{Err: err, Description: _MID_AUTH_RESP})
			return clierr
		}
		// Log unknown mID service error about failed authentication
		log.Error(args.Ctx, GetCertificateError{Err: log.Alert(err), Description: _MID_NO_RESP})
		return server.ErrInternal
	}

	// Log signing certificate.
	// TODO: identical to GetCertificateResp, why both?
	log.Log(args.Ctx, SigningCertificate{Certificate: c, Description: _MID_CERT})
	resp.Certificate = c.Raw

	// Log successful reply
	// TODO: identical to SigningCertificate, why both?
	log.Log(args.Ctx, GetCertificateResp{
		Certificate: resp.Certificate,
		Description: _MID_GETCERTRESP,
	})
	return nil
}

// SignArgs are the arguments provided to a call of RPC.Sign.
type SignArgs struct {
	server.Header
	Hash     []byte `size:"64"` // Hash of the data to sign.
	HashType string `size:"10"` // Allowed values described in 'ivxv.ee/common/collector/mid' package.
}

// SignResponse is the response returned by RPC.Sign.
type SignResponse struct {
	server.Header
	SessionCode string
}

// Sign is the remote procedure call performed by clients to start a Mobile-ID
// signing session.
func (r *RPC) Sign(args SignArgs, resp *SignResponse) (err error) {
	log.Log(args.Ctx, SignReq{Hash: args.Hash, Description: _MID_SIGNREQ})

	// Get the voter serial number. If empty, then the request is not
	// authenticated.
	identity := server.VoterIdentity(args.Ctx)
	if len(identity) == 0 {
		log.Error(args.Ctx, UnauthenticatedSignError{Description: _MID_VOTER_NO_AUTH})
		return server.ErrUnauthenticated
	}

	// Get voter phone no from the ticket, not authenticated if does not exist
	phoneno := server.VoterNumber(args.Ctx)
	if len(phoneno) == 0 {
		log.Error(args.Ctx, MissingDataSignError{Description: _MID_VOTER_NO_PHONENR})
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
		log.Error(args.Ctx, SignVerifySessionIDError{Err: err, Description: _MID_SESSION_ID})
		return server.ErrBadRequest
	}
	if !ok {
		// SessionID is unknown / has expired, we shall not further process the request
		log.Error(args.Ctx, SignUpdateSessionIDError{Description: _MID_SESSION_ID_EXPIRED})
		return server.ErrBadRequest
	}

	resp.SessionCode, err = r.mid.MobileSignHash(
		args.Ctx, identity, phoneno, args.Hash, args.HashType)
	if err != nil {
		if clierr := midToServerError(err); clierr != nil {
			// Log known mID service error about failed authentication
			log.Error(args.Ctx, SignMIDError{Err: err, Description: _MID_AUTH_RESP})
			return clierr
		}
		// Log unknown mID service error about failed authentication
		log.Error(args.Ctx, SignError{Err: log.Alert(err), Description: _MID_NO_RESP})
		return server.ErrInternal
	}

	// Successfully initiated signing
	log.Log(args.Ctx, SignResp{
		SessionCode: resp.SessionCode,
		Description: _MID_SIGNRESP,
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
	Algorithm string // The signature algorithm. Allowed values described in 'ivxv.ee/common/collector/mid' package.
}

// SignStatus is the remote procedure call performed by clients to check the
// status of a Mobile-ID signing session.
func (r *RPC) SignStatus(args SignStatusArgs, resp *SignStatusResponse) (err error) {
	log.Log(args.Ctx, SignStatusReq{SessionCode: args.SessionCode, Description: _MID_SIGNSTATUSREQ})

	// Build up VerifyReq for session status service
	verifyReq := status.NewVerifyReqBuilder().
		WithServiceMethod(internal.SignStatus).
		WithRequest(args.Header).
		Build()

	// SessionID security check
	ok, err := r.status.Verify(&verifyReq)
	if err != nil {
		// Error during SessionID check - database unreachable, service stalled, etc.
		log.Error(args.Ctx, SignStatusVerifySessionIDError{Err: err, Description: _MID_SESSION_ID})
		return server.ErrBadRequest
	}
	if !ok {
		// SessionID is unknown / has expired, we shall not further process the request
		log.Error(args.Ctx, SignStatusUpdateSessionIDError{Description: _MID_SESSION_ID_EXPIRED})
		return server.ErrBadRequest
	}

	resp.Algorithm, resp.Signature, err = r.mid.GetMobileSignHashStatus(args.Ctx, args.SessionCode)
	if err != nil {
		if clierr := midToServerError(err); clierr != nil {
			// Log known mID service error about failed authentication
			log.Error(args.Ctx, SignStatusMIDError{Err: err, Description: _MID_AUTH_RESP})
			return clierr
		}
		// Log unknown mID service error about failed authentication
		log.Error(args.Ctx, SignStatusError{Err: log.Alert(err), Description: _MID_NO_RESP})
		return server.ErrInternal
	}

	resp.Status = StatusPoll
	if len(resp.Signature) > 0 {
		resp.Status = StatusOK
	}

	// Log sucessfully received signature
	log.Log(args.Ctx, SignStatusResp{
		Status:      resp.Status,
		Signature:   resp.Signature,
		Description: _MID_SIGNSTATUSRESP,
	})
	return
}

func main() {
	// Call midmain in a separate function so that it can set up defers
	// and have them trigger before returning with a non-zero exit code.
	os.Exit(midmain())
}

func midmain() (code int) {
	c := command.NewWithoutStorage("ivxv-mid", "")
	defer func() {
		code = c.Cleanup(code)
	}()

	// Configure session status client
	statusClient, errCode := internal.NewClient(c)
	if statusClient == nil || errCode != 0 {
		return errCode
	}

	// Create new RPC instance and start the session cleaner.
	rpc := &RPC{sessions: make(map[string]session), status: statusClient}
	rpc.startCleaner(c.Ctx)

	var start, stop time.Time
	var authConf server.AuthConf
	var err error

	if c.Conf.Election != nil {
		// Check election configuration time values - service start
		if start, err = c.Conf.Election.ServiceStartTime(); err != nil {
			return c.Error(exit.Config, StartTimeError{Err: err, Description: _MID_START},
				"bad service start time:", err)
		}

		// Check election configuration time values - election stop
		if rpc.authEnd, err = c.Conf.Election.ElectionStopTime(); err != nil {
			return c.Error(exit.Config, ElectionStopTimeError{Err: err, Description: _MID_AUTH_STOP},
				"bad election stop time:", err)
		}

		// Check election configuration time values - service stop
		if stop, err = c.Conf.Election.ServiceStopTime(); err != nil {
			return c.Error(exit.Config, ServiceStopTimeError{Err: err, Description: _MID_STOP},
				"bad service stop time:", err)
		}

		// Configure the MID-REST API client.
		if rpc.mid, err = mid.New(&c.Conf.Election.MID); err != nil {
			return c.Error(exit.Config, MIDConfError{Err: err, Description: _MID_CLIENT_CONF},
				"failed to configure MID-REST API client:", err)
		}

		// Configure the ticket manager for issuing authentication
		// tickets.
		ticketConf, ok := c.Conf.Election.Auth[auth.Ticket]
		if !ok {
			return c.Error(exit.Config, TicketAuthError{Description: _MID_TICKET_AUTH},
				"ticket authentication is mandatory for mid")
		}
		if rpc.ticket, err = ticket.NewFromSystem(); err != nil {
			return c.Error(exit.Config, TicketConfError{Err: err, Description: _MID_TICKET},
				"failed to configure ticket manager:", err)
		}

		// Parse configuration for authenticating with tickets issued
		// by this server.
		if authConf, err = server.NewAuthConf(auth.Conf{auth.Ticket: ticketConf},
			c.Conf.Election.Identity, nil); err != nil {

			return c.Error(exit.Config, ServerAuthConfError{Err: err, Description: _MID_AUTH},
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
			return c.Error(exit.Config, ServerConfError{Err: err, Description: _MID_SERVER},
				"failed to configure server:", err)
		}
	}

	// Start listening for incoming connections during the voting period.
	if c.Until >= command.Execute {
		if err = s.WithAuth(authConf).ServeAt(c.Ctx, start); err != nil {
			return c.Error(exit.Unavailable, ServeError{Err: err, Description: _MID_SERVER_SERVE},
				"failed to serve mid service:", err)
		}
	}
	return exit.OK
}
