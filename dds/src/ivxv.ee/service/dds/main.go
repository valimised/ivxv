/*
The dds service performs Mobile-ID authentication and intermediates requests
for Mobile-ID signing.
*/
package main

import (
	"context"
	"crypto/x509"
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
	"ivxv.ee/dds"
	"ivxv.ee/errors"
	"ivxv.ee/identity"
	"ivxv.ee/log"
	"ivxv.ee/server"
	//ivxv:modules container
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

// ddsToServerError maps an ivxv.ee/dds package error to an ivxv.ee/server
// error to return to the client. It returns nil if err is not a recognized dds
// error, i.e., it was caused by an internal server error.
func ddsToServerError(err error) error {
	switch {
	case errors.CausedBy(err, new(dds.InputError)) != nil:
		return server.ErrBadRequest
	case errors.CausedBy(err, new(dds.NotMIDUserError)) != nil:
		return server.ErrMIDNotUser
	case errors.CausedBy(err, new(dds.AbsentError)) != nil:
		return server.ErrMIDAbsent
	case errors.CausedBy(err, new(dds.CanceledError)) != nil:
		return server.ErrMIDCanceled
	case errors.CausedBy(err, new(dds.ExpiredError)) != nil:
		return server.ErrMIDExpired
	case errors.CausedBy(err, new(dds.CertificateError)) != nil:
		return server.ErrMIDCertificate
	case errors.CausedBy(err, new(dds.StatusError)) != nil:
		return server.ErrMIDGeneral
	}
	return nil
}

// session is an outstanding Mobile-ID authentication session.
type session struct {
	created   time.Time
	challenge []byte
	cert      *x509.Certificate
}

// RPC is the handler for Mobile-ID service calls.
type RPC struct {
	authEnd time.Time
	dds     *dds.Client
	ticket  *ticket.T

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
	IDCode  string `size:"11"`
	PhoneNo string `size:"20"`
}

// AuthResponse is the response returned by RPC.Authenticate.
type AuthResponse struct {
	server.Header
	SessionCode string
	ChallengeID string
}

// Authenticate is the remote procedure call performed by clients to start a
// Mobile-ID authentication session.
func (r *RPC) Authenticate(args AuthArgs, resp *AuthResponse) (err error) {
	log.Log(args.Ctx, AuthenticateReq{PhoneNo: args.PhoneNo})

	// The server filter for voting end will only enable once we stop
	// serving signing requests, so we must manually check if we should
	// still serve authentication requests.
	if !time.Now().Before(r.authEnd) { // not before == equal or after
		log.Log(args.Ctx, AuthenticateVotingEnded{})
		return server.ErrVotingEnd
	}

	sess := session{created: time.Now()}
	resp.SessionCode, resp.ChallengeID, sess.challenge, sess.cert, err =
		r.dds.MobileAuthenticate(args.Ctx, args.IDCode, args.PhoneNo)
	// nolint: dupl, ignore duplicate error handling code, this is small
	// enough to not warrant deduplication.
	if err != nil {
		if clierr := ddsToServerError(err); clierr != nil {
			log.Error(args.Ctx, AuthenticateDDSError{Err: err})
			return clierr
		}
		log.Error(args.Ctx, AuthenticateError{Err: log.Alert(err)})
		return server.ErrInternal
	}
	log.Log(args.Ctx, AuthenticationCertificate{Certificate: sess.cert})

	r.sessionLock.Lock()
	defer r.sessionLock.Unlock()
	if _, ok := r.sessions[resp.SessionCode]; ok {
		log.Error(args.Ctx, DuplicateSessionCodeError{Code: resp.SessionCode})
		return server.ErrInternal
	}
	r.sessions[resp.SessionCode] = sess

	log.Log(args.Ctx, AuthenticateResp{
		SessionCode: resp.SessionCode,
		ChallengeID: resp.ChallengeID,
	})
	return nil
}

// AuthStatusArgs are the arguments provided to a call of RPC.AuthenticateStatus.
type AuthStatusArgs struct {
	server.Header
	SessionCode string `size:"32"`
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

	signature, err := r.dds.GetMobileAuthenticateStatus(args.Ctx, args.SessionCode)
	if err != nil {
		r.sessionLock.Lock()
		delete(r.sessions, args.SessionCode)
		r.sessionLock.Unlock()

		if clierr := ddsToServerError(err); clierr != nil {
			log.Error(args.Ctx, AuthenticateStatusDDSError{Err: err})
			return clierr
		}
		log.Error(args.Ctx, AuthenticateStatusError{Err: log.Alert(err)})
		return server.ErrInternal
	}

	resp.Status = StatusPoll
	if len(signature) > 0 {
		log.Log(args.Ctx, AuthenticationSignature{Signature: signature})

		r.sessionLock.Lock()
		delete(r.sessions, args.SessionCode)
		r.sessionLock.Unlock()

		if err = dds.VerifyAuthenticationSignature(
			sess.cert, sess.challenge, signature); err != nil {

			log.Error(args.Ctx, AuthenticationSignatureError{Err: err})
			return server.ErrMIDGeneral
		}

		resp.Status = StatusOK
		resp.GivenName = findName(&sess.cert.Subject, asn1.ObjectIdentifier{2, 5, 4, 42})
		resp.Surname = findName(&sess.cert.Subject, asn1.ObjectIdentifier{2, 5, 4, 4})
		resp.PersonalCode = sess.cert.Subject.SerialNumber

		if resp.AuthToken, err = r.ticket.Create(sess.cert.Subject); err != nil {
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

// CertificateArgs are the arguments provided to a call of RPC.GetCertificate.
type CertificateArgs struct {
	server.Header
	PhoneNo string `size:"20"`
}

// CertificateResponse is the response returned by RPC.GetCertificate.
type CertificateResponse struct {
	server.Header
	Certificate []byte
}

// GetCertificate is the remote procedure call performed by clients to get the
// Mobile-ID signing certificate that will be used to sign the vote.
func (r *RPC) GetCertificate(args CertificateArgs, resp *CertificateResponse) error {
	log.Log(args.Ctx, GetCertificateReq{PhoneNo: args.PhoneNo})

	// Get the voter serial number. If empty, then the request is not
	// authenticated.
	identity := server.VoterIdentity(args.Ctx)
	if len(identity) == 0 {
		log.Error(args.Ctx, UnauthenticatedGetCertificateError{})
		return server.ErrUnauthenticated
	}

	c, err := r.dds.GetMobileCertificate(args.Ctx, identity, args.PhoneNo)
	// nolint: dupl, ignore duplicate error handling code, this is small
	// enough to not warrant deduplication.
	if err != nil {
		if clierr := ddsToServerError(err); clierr != nil {
			log.Error(args.Ctx, GetCertificateDDSError{Err: err})
			return clierr
		}
		log.Error(args.Ctx, GetCertificateError{Err: log.Alert(err)})
		return server.ErrInternal
	}
	log.Log(args.Ctx, SigningCertificate{Certificate: c})
	resp.Certificate = c.Raw

	log.Log(args.Ctx, GetCertificateResp{
		Certificate: resp.Certificate,
	})
	return nil
}

// SignArgs are the arguments provided to a call of RPC.Sign.
type SignArgs struct {
	server.Header
	PhoneNo string `size:"20"`
	Hash    []byte `size:"32"` // SHA-256 hash of the data to sign.
}

// SignResponse is the response returned by RPC.Sign.
type SignResponse struct {
	server.Header
	SessionCode string
	ChallengeID string
}

// Sign is the remote procedure call performed by clients to start a Mobile-ID
// signing session.
func (r *RPC) Sign(args SignArgs, resp *SignResponse) (err error) {
	log.Log(args.Ctx, SignReq{PhoneNo: args.PhoneNo, Hash: args.Hash})

	// Get the voter serial number. If empty, then the request is not
	// authenticated.
	identity := server.VoterIdentity(args.Ctx)
	if len(identity) == 0 {
		log.Error(args.Ctx, UnauthenticatedSignError{})
		return server.ErrUnauthenticated
	}

	resp.SessionCode, resp.ChallengeID, err = r.dds.MobileSignHash(
		args.Ctx, identity, args.PhoneNo, args.Hash)
	// nolint: dupl, ignore duplicate error handling code, this is small
	// enough to not warrant deduplication.
	if err != nil {
		if clierr := ddsToServerError(err); clierr != nil {
			log.Error(args.Ctx, SignDDSError{Err: err})
			return clierr
		}
		log.Error(args.Ctx, SignError{Err: log.Alert(err)})
		return server.ErrInternal
	}

	log.Log(args.Ctx, SignResp{
		SessionCode: resp.SessionCode,
		ChallengeID: resp.ChallengeID,
	})
	return
}

// SignStatusArgs are the arguments provided to a call of RPC.SignStatus.
type SignStatusArgs struct {
	server.Header
	SessionCode string `size:"32"`
}

// SignStatusResponse is the response returned by RPC.SignStatus.
type SignStatusResponse struct {
	server.Header
	Status    string
	Signature []byte
}

// SignStatus is the remote procedure call performed by clients to check the
// status of a Mobile-ID signing session.
func (r *RPC) SignStatus(args SignStatusArgs, resp *SignStatusResponse) (err error) {
	log.Log(args.Ctx, SignStatusReq{SessionCode: args.SessionCode})

	resp.Signature, err = r.dds.GetMobileSignHashStatus(args.Ctx, args.SessionCode)
	// nolint: dupl, ignore duplicate error handling code, this is small
	// enough to not warrant deduplication.
	if err != nil {
		if clierr := ddsToServerError(err); clierr != nil {
			log.Error(args.Ctx, SignStatusDDSError{Err: err})
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
	// Call ddsmain in a separate function so that it can set up defers
	// and have them trigger before returning with a non-zero exit code.
	os.Exit(ddsmain())
}

func ddsmain() (code int) {
	c := command.NewWithoutStorage("ivxv-dds", "")
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

		// Configure the DigiDocService client.
		if rpc.dds, err = dds.New(&c.Conf.Election.DDS); err != nil {
			return c.Error(exit.Config, DDSConfError{Err: err},
				"failed to configure DigiDocService client:", err)
		}

		// Configure the ticket manager for issuing authentication
		// tickets.
		ticketConf, ok := c.Conf.Election.Auth[auth.Ticket]
		if !ok {
			return c.Error(exit.Config, TicketAuthError{},
				"ticket authentication is mandatory for dds")
		}
		if rpc.ticket, err = ticket.NewFromSystem(); err != nil {
			return c.Error(exit.Config, TicketConfError{Err: err},
				"failed to configure ticket manager:", err)
		}

		// Parse configuration for authenticating with tickets issued
		// by this server. Also, require serial number identities
		// regardless of configuration, since those are the ones used
		// by Mobile-ID.
		if authConf, err = server.NewAuthConf(auth.Conf{auth.Ticket: ticketConf},
			identity.SerialNumber, nil); err != nil {

			return c.Error(exit.Config, ServerAuthConfError{Err: err},
				"failed to configure client authentication:", err)
		}
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
				"failed to serve dds service:", err)
		}
	}
	return exit.OK
}
