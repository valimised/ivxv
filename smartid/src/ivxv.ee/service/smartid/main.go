/*
The smartid service performs Smart-ID authentication and intermediates requests
for Smart-ID signing using the Smart-ID REST API.
*/
package main

import (
	"context"
	"crypto/x509/pkix"
	"encoding/asn1"
	"os"
	"sync"
	"time"

	"ivxv.ee/auth"
	"ivxv.ee/auth/ticket"
	"ivxv.ee/command"
	"ivxv.ee/command/exit"
	"ivxv.ee/conf"
	"ivxv.ee/errors"
	"ivxv.ee/identity"
	"ivxv.ee/log"
	"ivxv.ee/server"
	"ivxv.ee/smartid"
	//ivxv:modules container
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

// smartidToServerError maps an ivxv.ee/smartid package error to an ivxv.ee/server
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

// session is an outstanding Smart-ID authentication session.
type session struct {
	created      time.Time
	challengeRnd []byte
}

// RPC is the handler for Smart-ID service calls.
type RPC struct {
	authEnd  time.Time
	smartid  *smartid.Client
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
					log.Debug(ctx, SessionTimeout{SessionCode: code})
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
	Identifier string `size:"11"`
}

// AuthResponse is the response returned by RPC.Authenticate.
type AuthResponse struct {
	server.Header
	SessionCode string
	Challenge   []byte
}

// Authenticate is the remote procedure call performed by clients to start a
// Smart-ID authentication session.
func (r *RPC) Authenticate(args AuthArgs, resp *AuthResponse) (err error) {
	log.Log(args.Ctx, AuthenticateReq{Identifier: args.Identifier})

	// The server filter for voting end will only enable once we stop
	// serving signing requests, so we must manually check if we should
	// still serve authentication requests.
	if !time.Now().Before(r.authEnd) { // not before == equal or after
		log.Log(args.Ctx, AuthenticateVotingEnded{})
		return server.ErrVotingEnd
	}

	sess := session{created: time.Now()}
	resp.SessionCode, sess.challengeRnd, resp.Challenge, err =
		r.smartid.Authenticate(args.Ctx, args.Identifier)
	if err != nil {
		if clierr := smartidToServerError(err); clierr != nil {
			log.Error(args.Ctx, AuthenticateSmartIDError{Err: err})
			return clierr
		}
		log.Error(args.Ctx, AuthenticateError{Err: log.Alert(err)})
		return server.ErrInternal
	}

	r.sessionLock.Lock()
	defer r.sessionLock.Unlock()
	if _, ok := r.sessions[resp.SessionCode]; ok {
		log.Error(args.Ctx, DuplicateSessionCodeError{Code: resp.SessionCode})
		return server.ErrInternal
	}
	r.sessions[resp.SessionCode] = sess

	log.Log(args.Ctx, AuthenticateResp{
		SessionCode: resp.SessionCode,
		Challenge:   resp.Challenge,
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
// check the status of a Smart-ID authentication session.
func (r *RPC) AuthenticateStatus(args AuthStatusArgs, resp *AuthStatusResponse) error {
	log.Log(args.Ctx, AuthenticateStatusReq{SessionCode: args.SessionCode})

	// The server filter for voting end will only enable once we stop
	// serving signing requests, so we must manually check if we should
	// still serve authentication requests.
	if !time.Now().Before(r.authEnd) { // not before == equal or after
		log.Log(args.Ctx, AuthenticateStatusVotingEnded{})
		return server.ErrVotingEnd
	}

	r.sessionLock.Lock()
	sess, ok := r.sessions[args.SessionCode]
	r.sessionLock.Unlock()
	if !ok {
		log.Error(args.Ctx, UnknownSessionCodeError{})
		return server.ErrBadRequest
	}

	cert, algorithm, signature, err := r.smartid.GetAuthenticateStatus(args.Ctx, args.SessionCode)
	if err != nil {
		r.sessionLock.Lock()
		delete(r.sessions, args.SessionCode)
		r.sessionLock.Unlock()

		if clierr := smartidToServerError(err); clierr != nil {
			log.Error(args.Ctx, AuthenticateStatusSmartIDError{Err: err})
			return clierr
		}
		log.Error(args.Ctx, AuthenticateStatusError{Err: log.Alert(err)})
		return server.ErrInternal
	}

	resp.Status = StatusPoll
	if len(signature) > 0 {
		log.Log(args.Ctx, AuthenticationSignature{Signature: signature})
		if cert == nil {
			log.Error(args.Ctx, AuthenticationCertificateMissingError{Err: err})
			return server.ErrSmartIDGeneral
		}
		log.Log(args.Ctx, AuthenticationCertificate{Certificate: cert})

		r.sessionLock.Lock()
		delete(r.sessions, args.SessionCode)
		r.sessionLock.Unlock()

		if err = smartid.VerifyAuthenticationSignature(
			cert, algorithm, sess.challengeRnd, signature); err != nil {

			log.Error(args.Ctx, AuthenticationSignatureError{Err: err})
			return server.ErrSmartIDGeneral
		}

		resp.Status = StatusOK
		resp.GivenName = findName(&cert.Subject, asn1.ObjectIdentifier{2, 5, 4, 42})
		resp.Surname = findName(&cert.Subject, asn1.ObjectIdentifier{2, 5, 4, 4})
		if resp.PersonalCode, err = r.identify(&cert.Subject); err != nil {
			log.Error(args.Ctx, AuthenticationSubjectIdentityError{Err: err})
			return server.ErrInternal
		}

		if resp.AuthToken, err = r.ticket.Create(cert.Subject); err != nil {
			log.Error(args.Ctx, AuthenticationTicketError{Err: err})
			return server.ErrInternal
		}
	}

	log.Log(args.Ctx, AuthenticateStatusResp{
		Status:       resp.Status,
		GivenName:    resp.GivenName,
		Surname:      resp.Surname,
		PersonalCode: resp.PersonalCode,
		AuthToken:    log.Sensitive(resp.AuthToken),
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
	log.Log(args.Ctx, GetCertificateChoiceReq{})

	// Get the voter serial number. If empty, then the request is not
	// authenticated.
	identity := server.VoterIdentity(args.Ctx)
	if len(identity) == 0 {
		log.Error(args.Ctx, UnauthenticatedGetCertificateError{})
		return server.ErrUnauthenticated
	}

	s, err := r.smartid.GetCertificateChoice(args.Ctx, identity)
	if err != nil {
		if clierr := smartidToServerError(err); clierr != nil {
			log.Error(args.Ctx, GetCertificateChoiceSmartIDError{Err: err})
			return clierr
		}
		log.Error(args.Ctx, GetCertificateChoiceError{Err: log.Alert(err)})
		return server.ErrInternal
	}
	log.Log(args.Ctx, GetCertificateChoiceResp{
		Session: resp.SessionCode,
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
	DataToken   []byte
}

// GetCertificateChoiceStatus is the remote procedure call performed by clients to get the
// Smart-ID signing certificate that will be used to sign the vote.
func (r *RPC) GetCertificateChoiceStatus(args CertificateChoiceStatusArgs,
	resp *CertificateChoiceStatusResponse) error {
	log.Log(args.Ctx, GetCertificateChoiceStatusReq{Session: args.SessionCode})

	documentno, c, err := r.smartid.GetCertificateChoiceStatus(args.Ctx, args.SessionCode)
	if err != nil {
		if clierr := smartidToServerError(err); clierr != nil {
			log.Error(args.Ctx, GetCertificateChoiceStatusSmartIDError{Err: err})
			return clierr
		}
		log.Error(args.Ctx, GetCertificateChoiceStatusError{Err: log.Alert(err)})
		return server.ErrInternal
	}
	log.Log(args.Ctx, SigningCertificate{Certificate: c})
	if resp.DataToken, err = r.ticket.CreateData([]byte(documentno)); err != nil {
		log.Error(args.Ctx, DataTicketError{Err: err})
		return server.ErrInternal
	}
	resp.Status = StatusPoll
	if c != nil {
		resp.Status = StatusOK
		resp.Certificate = c.Raw
	}

	log.Log(args.Ctx, GetCertificateResp{
		Certificate: resp.Certificate,
		DataToken:   log.Sensitive(resp.DataToken),
		Status:      resp.Status,
	})
	return nil
}

// SignArgs are the arguments provided to a call of RPC.Sign.
type SignArgs struct {
	server.Header
	Hash     []byte `size:"64"` // Hash of the data to sign.
	HashType string `size:"10"` // Allowed values described in 'ivxv.ee/smartid' package.
}

// SignResponse is the response returned by RPC.Sign.
type SignResponse struct {
	server.Header
	SessionCode string
}

// Sign is the remote procedure call performed by clients to start a Smart-ID
// signing session.
func (r *RPC) Sign(args SignArgs, resp *SignResponse) (err error) {
	log.Log(args.Ctx, SignReq{HashType: args.HashType, Hash: args.Hash})

	// Get the voter serial number. If empty, then the request is not
	// authenticated.
	identity := server.VoterIdentity(args.Ctx)
	if len(identity) == 0 {
		log.Error(args.Ctx, UnauthenticatedSignError{})
		return server.ErrUnauthenticated
	}
	documentno := server.VoterNumber(args.Ctx)
	if len(documentno) == 0 {
		log.Error(args.Ctx, MissingDataSignError{})
		return server.ErrUnauthenticated
	}
	resp.SessionCode, err = r.smartid.SignHash(
		args.Ctx, documentno, args.Hash, args.HashType)
	if err != nil {
		if clierr := smartidToServerError(err); clierr != nil {
			log.Error(args.Ctx, SignSmartIDError{Err: err})
			return clierr
		}
		log.Error(args.Ctx, SignError{Err: log.Alert(err)})
		return server.ErrInternal
	}

	log.Log(args.Ctx, SignResp{
		SessionCode: resp.SessionCode,
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
	Algorithm string // The signature algorithm. Allowed values described in 'ivxv.ee/smartid' package.
}

// SignStatus is the remote procedure call performed by clients to check the
// status of a Smart-ID signing session.
func (r *RPC) SignStatus(args SignStatusArgs, resp *SignStatusResponse) (err error) {
	log.Log(args.Ctx, SignStatusReq{SessionCode: args.SessionCode})

	resp.Algorithm, resp.Signature, err = r.smartid.GetSignHashStatus(args.Ctx, args.SessionCode)
	if err != nil {
		if clierr := smartidToServerError(err); clierr != nil {
			log.Error(args.Ctx, SignStatusSmartIDError{Err: err})
			return clierr
		}
		log.Error(args.Ctx, SignStatusError{Err: log.Alert(err)})
		return server.ErrInternal
	}

	resp.Status = StatusPoll
	if len(resp.Signature) > 0 {
		resp.Status = StatusOK
	}

	log.Log(args.Ctx, SignStatusResp{
		Status:    resp.Status,
		Signature: resp.Signature,
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
	// Create new RPC instance and start the session cleaner.
	rpc := &RPC{sessions: make(map[string]session)}
	rpc.startCleaner(c.Ctx)

	var start, stop time.Time
	var authConf server.AuthConf
	var err error
	if c.Conf.Election != nil {
		// Check election configuration time values.
		if start, err = c.Conf.Election.ServiceStartTime(); err != nil {
			return c.Error(exit.Config, StartTimeError{Err: err},
				"bad service start time:", err)
		}

		if rpc.authEnd, err = c.Conf.Election.ElectionStopTime(); err != nil {
			return c.Error(exit.Config, ElectionStopTimeError{Err: err},
				"bad election stop time:", err)
		}

		if stop, err = c.Conf.Election.ServiceStopTime(); err != nil {
			return c.Error(exit.Config, ServiceStopTimeError{Err: err},
				"bad service stop time:", err)
		}

		// Configure the Smart-ID REST API client.
		if rpc.smartid, err = smartid.New(&c.Conf.Election.SmartID); err != nil {
			return c.Error(exit.Config, SmartIDConfError{Err: err},
				"failed to configure SmartID-REST API client:", err)
		}

		// Configure the ticket manager for issuing authentication
		// tickets.
		ticketConf, ok := c.Conf.Election.Auth[auth.Ticket]
		if !ok {
			return c.Error(exit.Config, TicketAuthError{},
				"ticket authentication is mandatory for smartid")
		}
		if rpc.ticket, err = ticket.NewFromSystem(); err != nil {
			return c.Error(exit.Config, TicketConfError{Err: err},
				"failed to configure ticket manager:", err)
		}

		// Parse configuration for authenticating with tickets issued
		// by this server.
		if authConf, err = server.NewAuthConf(auth.Conf{auth.Ticket: ticketConf},
			c.Conf.Election.Identity, nil); err != nil {

			return c.Error(exit.Config, ServerAuthConfError{Err: err},
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
			return c.Error(exit.Config, ServerConfError{Err: err},
				"failed to configure server:", err)
		}
	}

	// Start listening for incoming connections during the voting period.
	if c.Until >= command.Execute {
		if err = s.WithAuth(authConf).ServeAt(c.Ctx, start); err != nil {
			return c.Error(exit.Unavailable, ServeError{Err: err},
				"failed to serve smartid service:", err)
		}
	}
	return exit.OK
}
