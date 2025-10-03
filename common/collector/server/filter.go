package server

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"io"
	"net"
	"net/rpc"
	"regexp"
	"time"

	"ivxv.ee/common/collector/age"
	"ivxv.ee/common/collector/auth"
	"ivxv.ee/common/collector/errors"
	"ivxv.ee/common/collector/identity"
	"ivxv.ee/common/collector/log"
)

// connFilter is a filter which works on the network connection level.
type connFilter interface {
	// filter wraps c to filter data read and written, and calls the next
	// filter in the chain. It returns a possibly modified context.
	//
	// In case of errors, filter logs them, notifies the client via c if
	// applicable, closes c, and returns without calling next.
	filter(ctx context.Context, c net.Conn, chain connFilters) context.Context
}

// connFilters is a chain of connFilters.
type connFilters []connFilter

// next calls the next filter in the chain, passing the rest of the chain as
// the continuation. next panics if we reached the end of the chain: the last
// filter should always block the chain and not call next.
func (cfs connFilters) next(ctx context.Context, c net.Conn) context.Context {
	if len(cfs) == 0 {
		panic("end of chain")
	}
	return cfs[0].filter(ctx, c, cfs[1:])
}

// connFilterFunc is a helper type which creates connFilters out of functions.
type connFilterFunc func(ctx context.Context, c net.Conn, chain connFilters) context.Context

func (f connFilterFunc) filter(ctx context.Context, c net.Conn, chain connFilters) context.Context {
	return f(ctx, c, chain)
}

// headerFilter is a filter which works on the server header of messages.
type headerFilter interface {
	// filter filters data in the header and calls the next filter in the
	// chain. If filter returns a non-nil error, then the request is not
	// handled. The non-nil error is sent to the client, so the filter must
	// be as generic as possible and log any relevant error information
	// itself.
	//
	// header.Ctx is used as the context instead of passing it as the first
	// argument: see server.Header for the reasoning behind this.
	filter(header *Header, chain headerFilters) error
}

// headerFilters is a chain of headerFilters.
type headerFilters []headerFilter

// next calls the next filter in the chain, passing the rest of the chain as
// the continuation. next does nothing if the chain is empty.
func (hfs headerFilters) next(header *Header) error {
	if len(hfs) == 0 {
		return nil
	}
	return hfs[0].filter(header, hfs[1:])
}

// headerFilterFunc is a helper type which creates headerFilters out of functions.
type headerFilterFunc func(header *Header, chain headerFilters) error

func (f headerFilterFunc) filter(header *Header, chain headerFilters) error {
	return f(header, chain)
}

// FilterConf is the configuration for filters used by servers.
type FilterConf struct {
	TLS   TLSConf
	Codec CodecConf
}

// newFilters returns a new chain of mandatory filters.
func newFilters(conf *FilterConf, r *rpc.Server, cert tls.Certificate, end time.Time, certPool *x509.CertPool) (
	connFilters, error) {

	tlsFilter, err := newTLSFilter(&conf.TLS, cert)
	if err != nil {
		return nil, TLSConfError{Err: err, Description: _SERVER_FILTER_TLS}
	}
	if certPool != nil {
		tlsFilter.tlsConf.ClientCAs = certPool
		tlsFilter.tlsConf.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return connFilters{
		connFilterFunc(logFilter),
		connFilterFunc(connIDFilter),
		connFilterFunc(proxyFilter),
		tlsFilter,
		&codecFilter{&conf.Codec, r, headerFilters{
			endFilter(end),
			headerFilterFunc(sessIDFilter),
			headerFilterFunc(addrFilter),
			headerFilterFunc(infoFilter),
		}},
	}, nil
}

// optional adds optional filters to the chain.
func (cfs connFilters) optional(auth auth.Auther, id identity.Identifier, age *age.Checker) {
	// All three need to be added to the codec filter. Feel free to panic
	// if f is empty or the last filter is not codecFilter, because that
	// means there is a programmer error.
	cf := cfs[len(cfs)-1].(*codecFilter)

	// The auth filter is required for the identity filter and the identity
	// filter is required for the age filter, so they need to be added in
	// that order and only if the preceding one is added.
	if len(auth) > 0 {
		cf.filters = append(cf.filters, authFilter(auth))

		if id != nil {
			cf.filters = append(cf.filters, identityFilter(id))

			if age != nil {
				cf.filters = append(cf.filters, (*ageFilter)(age))
			}
		}
	}
}

// close logs entry if not nil, closes c, and logs any closing errors.
func close(ctx context.Context, c io.Closer, err log.ErrorEntry) { //nolint: revive
	if err != nil {
		// Error to be logged while closing the connection
		log.Error(ctx, err)
	}
	if err := c.Close(); err != nil {
		// Error while closing connection
		log.Error(ctx, CloseError{Err: err, Description: _SERVER_CLOSE})
	}
}

// logFilter logs when a new connection is accepted and closed.
func logFilter(ctx context.Context, c net.Conn, chain connFilters) context.Context {
	log.Log(ctx, AcceptedConnection{Remote: c.RemoteAddr(),
		Description: _SERVER_FILTER_LOG})
	ctx = chain.next(ctx, c)
	log.Log(ctx, ClosedConnection{Remote: c.RemoteAddr(),
		Description: _SERVER_FILTER_LOG_CLOSE})
	return ctx
}

// connIDFilter assigns an identifier to the connection and stores it as a
// context value.
func connIDFilter(ctx context.Context, c net.Conn, chain connFilters) context.Context {
	cid := make([]byte, 16)
	if _, err := rand.Read(cid); err != nil {
		// Error generating a random connection identifier
		close(ctx, c, GenerateConnectionIDError{Err: log.Alert(err),
			Description: _SERVER_FILTER_CONNID})
		return ctx
	}
	ctx = log.WithConnectionID(ctx, hex.EncodeToString(cid))
	// Log the identifier that we have assigned to the connection
	log.Log(ctx, AssignedConnectionID{Remote: c.RemoteAddr(),
		Description: _SERVER_FILTER_CONNID_OK})
	return chain.next(ctx, c)
}

// proxyFilter checks if the connection is prefixed with the PROXY protocol
// version 2 (http://www.haproxy.org/download/1.8/doc/proxy-protocol.txt). If
// so, it logs the proxied address before passing the rest through unchanged.
//
// Actually the PROXY protocol mandates that you either use PROXY or not and
// should not accept both cases. However, we made it optional to simplify
// development environments where HAProxy is not used. In a production
// environment, we always use PROXY and all connections to services come
// through HAProxy, so there is no danger of spoofed addresses.
func proxyFilter(ctx context.Context, c net.Conn, chain connFilters) context.Context {
	c, addr, health, err := readPROXY(c)
	if err != nil {
		close(ctx, c, PROXYProtocolError{Err: err, Description: _SERVER_FILTER_PROXY_HEADER})
		return ctx
	}
	if health {
		// Log health check
		log.Log(ctx, HealthCheck{Description: _SERVER_FILTER_PROXY_HC})
		close(ctx, c, nil)
		return ctx
	}
	if addr != nil {
		// Log actual remote received via proxy
		log.Log(ctx, PROXYProtocol{Address: addr, Description: _SERVER_FILTER_PROXY_CLIENT_ADDR})
	}

	// Put remote address into context for addrFilter. Remove once
	// addrFilter is no longer necessary.
	ctx = context.WithValue(ctx, addrKey, c.RemoteAddr())
	if addr != nil {
		ctx = context.WithValue(ctx, addrKey, addr)
	}

	return chain.next(ctx, c)
}

// TLSConf is the TLS filter configuration.
type TLSConf struct {
	HandshakeTimeout int64    // TLS handshake timeout in seconds.
	CipherSuites     []string // Supported cipher suites, TLS 1.2 only.
}

// tlsFilter creates a new TLS connection that uses c as the underlying
// transport. It performs the handshake and puts any provided client
// certificates into the context.
type tlsFilter struct {
	serverConf *TLSConf
	tlsConf    *tls.Config
}

//go:generate ./tlsciphersuites tlsCipherSuites
func newTLSFilter(conf *TLSConf, cert tls.Certificate) (*tlsFilter, error) {
	f := &tlsFilter{
		serverConf: conf,
		tlsConf: &tls.Config{
			Certificates:           []tls.Certificate{cert},
			ClientAuth:             tls.RequestClientCert,
			SessionTicketsDisabled: true,
			MinVersion:             tls.VersionTLS12,
		},
	}
	for _, name := range conf.CipherSuites {
		value, ok := tlsCipherSuites[name] // Generated above.
		if !ok {
			// Unsupported ciphersuite detected
			return nil, UnsupportedTLSCipherSuiteError{CipherSuite: name,
				Description: _SERVER_FILTER_TLS_CIPHERS}
		}
		f.tlsConf.CipherSuites = append(f.tlsConf.CipherSuites, value)
	}
	return f, nil
}

func (f *tlsFilter) filter(ctx context.Context, c net.Conn, chain connFilters) context.Context {
	// Create the TLS connection.
	tlsc := tls.Server(c, f.tlsConf)

	// Before interacting with the client, check if we are done
	// If it indeed is the case - close with an error
	select {
	case <-ctx.Done():
		close(ctx, tlsc, CancelingBeforeHandshake{Err: ctx.Err(),
			Description: _SERVER_FILTER_TLS_HSHAKE})
		return ctx
	default:
	}

	// Explicitly perform the handshake to catch any errors early.
	deadline := time.Now().Add(time.Duration(f.serverConf.HandshakeTimeout) * time.Second)
	if err := tlsc.SetDeadline(deadline); err != nil {
		// Error in setting TLS handshake timeout
		close(ctx, tlsc, SetHandshakeTimeoutError{Err: err,
			Description: _SERVER_FILTER_TLS_HSHAKE_TIME})
		return ctx
	}
	if err := tlsc.Handshake(); err != nil {
		// Error in the TLS handshake
		close(ctx, tlsc, HandshakeError{Err: err, Description: _SERVER_FILTER_TLS_HSHAKE_ERR})
		return ctx
	}

	// Log the TLS connection details for successful connection
	state := tlsc.ConnectionState()
	log.Log(ctx, HandshakeComplete{
		Version:            state.Version,
		CipherSuite:        state.CipherSuite,
		ServerName:         state.ServerName,
		ClientCertificates: state.PeerCertificates,
		Description:        _SERVER_FILTER_TLS_HSHAKE_OK,
	})

	// Add PeerCertificates to the context and pass tlsc to next.
	return chain.next(context.WithValue(ctx, tlsClientKey, state.PeerCertificates), tlsc)
}

// CodecConf if the codec filter configuration.
type CodecConf struct {
	RWTimeout   int64 // Timeout for reading reading requests and writing responses in seconds.
	RequestSize int64 // Maximum accepted request size in bytes. 0 disables size limiting.
	LogRequests bool  // Should requests be logged?
}

// codecFilter terminates the connfilter chain: it passes c to the RPC server
// codec and closes it when the RPC call has finished. It does not call the
// next filter in the chain.
type codecFilter struct {
	conf    *CodecConf
	server  *rpc.Server
	filters headerFilters
}

func (f *codecFilter) filter(ctx context.Context, c net.Conn, _ connFilters) context.Context {
	codec := newCodec(ctx, f.conf, c, f.filters)

	// ServeRequest can return three types of errors:
	//
	// 1. An error reading the header of the RPC request: in this case,
	//    nothing is sent to the client (because it is assumed that they
	//    are not speaking the correct protocol) and the connection is
	//    simply closed. The error returned from the codec is converted to
	//    a string using its Error method, a prefix is added, and the
	//    string is converted to an error using errors.New. This destroys
	//    any hierarchical information about the errors and removes any
	//    extra implemented interfaces.
	//
	// 2. The header was read successfully, but the requested method does
	//    not exist: in this case, the rpc package generates an error
	//    itself, sends it to the client, and returns it here.
	//
	// 3. An error reading the body of the request or in one of the header
	//    filters (because the filters are executed as part of reading the
	//    body in serverCodec): in this case the error returned by the
	//    codec is sent to the client(!) and returned here (unmodified,
	//    i.e., without a prefix).
	//
	// Note that if the called method or writing the response(!) returns an
	// error, it is not returned here.
	//
	// This behavior means that this filter cannot simply rely on logging
	// the error returned by ServeRequest, because the errors will be
	// malformed (the first case) or not returned at all (if writing the
	// response failed). To solve this, ignore any errors returned here and
	// make the codec and filters responsible for logging them. This means
	// that the codec must also log errors generated by the rpc package
	// itself (the second case): it does this by intercepting any rpc
	// package errors sent to the client and logging them.
	//
	// Additionally, the rpc package sends too much internal error
	// information to the client (the third case): the codec and filters
	// must be sure to make any returned errors generic enough, that they
	// can be sent to the client.
	f.server.ServeRequest(codec) //nolint:errcheck // Read above.
	close(ctx, codec, nil)
	return ctx
}

// endFilter returns ErrVotingEnd to all requests starting from end time.
type endFilter time.Time

func (e endFilter) filter(header *Header, chain headerFilters) error {
	if !time.Now().Before(time.Time(e)) { // not before == equal or after
		// Time that is set in election.yml for period:servicestop is over
		log.Log(header.Ctx, VotingEnded{Description: _SERVER_FILTER_END})
		return ErrVotingEnd
	}
	return chain.next(header)
}

// sessIDFilter checks if a session ID is provided by the client or generates a
// new one if not.
func sessIDFilter(header *Header, chain headerFilters) error {
	var entry log.Entry
	if len(header.SessionID) == 0 {
		sid := make([]byte, 16)
		if _, err := rand.Read(sid); err != nil {
			// Error in generating new random session identifier
			log.Error(header.Ctx, GenerateSessionIDError{Err: log.Alert(err),
				Description: _SERVER_FILTER_SESSID})
			return ErrInternal
		}
		header.SessionID = hex.EncodeToString(sid)
		entry = AssignedSessionID{Description: _SERVER_FILTER_SESSID_OK}
	} else {
		entry = ReadSessionID{Description: _SERVER_FILTER_SESSID_REUSE}
	}

	// Validate that the header.SessionID is valid HEX
	invalidSessionID, _ := regexp.MatchString("[^0-9A-Fa-f]", header.SessionID)
	if invalidSessionID {
		log.Error(header.Ctx, InvalidSessionID{Value: header.SessionID,
			Description: _SERVER_FILTER_SESSID_REGEX})
		return ErrBadRequest
	}

	// Set session ID in logging context and log.
	header.Ctx = log.WithSessionID(header.Ctx, header.SessionID)
	log.Log(header.Ctx, entry)
	return chain.next(header)
}

// addrFilter re-logs the remote address of the connection after we have a
// SessionID.
//
// This is a temporary filter until the log monitor is capable of
// extracting the address based on ConnectionID.
func addrFilter(header *Header, chain headerFilters) error {
	log.Log(header.Ctx, RemoteAddress{Address: header.Ctx.Value(addrKey),
		Description: _SERVER_FILTER_ADDR})
	return chain.next(header)
}

// infoFilter logs any client-provided information about the platform used to
// perform the request, e.g., the operating system, and clears the fields from
// the header so that they are not included in the response.
func infoFilter(header *Header, chain headerFilters) error {
	log.Log(header.Ctx, OperatingSystem{OS: header.OS,
		Description: _SERVER_FILTER_OS})
	header.OS = ""
	return chain.next(header)
}

// authFilter verifies the authentication token in the request, clears the
// fields from the header so that they are not included in the response, and
// stores the authenticated client's name in the context as a value.
type authFilter auth.Auther

func (a authFilter) filter(header *Header, chain headerFilters) error {
	if len(header.AuthMethod) > 0 {
		// Initiate authentication with specific method. Token is sensitive.
		log.Log(header.Ctx, Authenticating{
			Method:      header.AuthMethod,
			Token:       log.Sensitive(header.AuthToken),
			Description: _SERVER_FILTER_AUTH,
		})
		name, voteid, err := auth.Auther(a).Verify(
			header.Ctx, auth.Type(header.AuthMethod), header.AuthToken)

		// Record authentication method for SessionID tamper check
		header.Ctx = WithAuthMethod(header.Ctx, header.AuthMethod)

		if err != nil {
			// Based on client RPC "Header.AuthMethod" type, backend has chosen an
			// authentication service and the result of that authentication has failed
			log.Error(header.Ctx, AuthenticationError{Err: err, Description: _SERVER_FILTER_AUTH_FAIL})
			switch {
			case errors.CausedBy(err, new(auth.UnconfiguredTypeError)) != nil:
				fallthrough
			case errors.CausedBy(err, new(auth.MalformedTokenError)) != nil:
				return ErrBadRequest
			case errors.CausedBy(err, new(auth.CertificateError)) != nil:
				return ErrCertificate
			case errors.CausedBy(err, new(auth.UnauthorizedError)) != nil:
				return ErrIneligible
			}
			return ErrInternal
		}
		// Log successful authentication
		header.Ctx = context.WithValue(header.Ctx, authClientKey, name)
		log.Log(header.Ctx, Authenticated{ClientName: name, Description: _SERVER_FILTER_AUTH_OK})

		if len(voteid) > 0 {
			// Log VoteID provided with the authentication
			header.Ctx = context.WithValue(header.Ctx, voterIDKey, voteid)
			log.Log(header.Ctx, AuthenticationVoteID{VoteID: voteid,
				Description: _SERVER_FILTER_AUTH_VID})
		}
		if header.DataToken != nil {
			// In case of DataToken, log the value, treating it as sensitive
			log.Log(header.Ctx, AuthData{
				Token:       log.Sensitive(header.DataToken),
				Description: _SERVER_FILTER_AUTH_DTOKEN,
			})
			data, err := auth.Auther(a).Data(auth.Type(header.AuthMethod), header.DataToken)
			if err != nil {
				// If client has passed RPC "Header.DataToken" (Smart-ID/mobile-ID) and that
				// data token verification has failed
				log.Error(header.Ctx, AuthenticationDataError{Err: err,
					Description: _SERVER_FILTER_AUTH_DTOKEN_FAIL})
				return ErrInternal
			}
			// Log data that was encoded in the DataToken
			header.Ctx = context.WithValue(header.Ctx, voterIDNumber, string(data))
			log.Log(header.Ctx, AuthenticationData{Number: string(data),
				Description: _SERVER_FILTER_AUTH_DTOKEN_OK})
		}

	}

	header.AuthMethod = ""
	header.AuthToken = nil
	header.DataToken = nil
	return chain.next(header)
}

// identityFilter extracts a unique identifier from an authenticated client's
// name and stores it as a context value.
type identityFilter identity.Identifier

func (i identityFilter) filter(header *Header, chain headerFilters) error {
	if name := AuthenticatedClient(header.Ctx); name != nil {
		id, err := identity.Identifier(i)(name)
		if err != nil {
			// Voter's personal code was not extracted correctly from authenticated data
			log.Error(header.Ctx, IdentityError{Err: err,
				Description: _SERVER_FILTER_IDENTITY})
			return ErrIneligible
		}
		// Log detected personal code
		header.Ctx = context.WithValue(header.Ctx, voterIDKey, id)
		log.Log(header.Ctx, Identity{Identity: id, Description: _SERVER_FILTER_IDENTITY_OK})
	}
	return chain.next(header)
}

// ageFilter determines the voter's age and checks if they are over a voting
// age limit.
type ageFilter age.Checker

func (a *ageFilter) filter(header *Header, chain headerFilters) error {
	if id := VoterIdentity(header.Ctx); len(id) > 0 {
		if err := (*age.Checker)(a).Check(id); err != nil {
			// Voter does not pass the age verification, most likely too young (nested error)
			log.Error(header.Ctx, AgeError{Err: err, Description: _SERVER_FILTER_AGE})
			if errors.CausedBy(err, new(age.TooYoungError)) != nil {
				return ErrTooYoung
			}
			return ErrIneligible
		}
	}
	return chain.next(header)
}
