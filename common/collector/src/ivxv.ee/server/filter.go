package server

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"io"
	"net"
	"net/rpc"
	"time"

	"ivxv.ee/age"
	"ivxv.ee/auth"
	"ivxv.ee/errors"
	"ivxv.ee/identity"
	"ivxv.ee/log"
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
func newFilters(conf *FilterConf, r *rpc.Server, cert tls.Certificate, end time.Time) (
	connFilters, error) {

	tlsFilter, err := newTLSFilter(&conf.TLS, cert)
	if err != nil {
		return nil, TLSConfError{Err: err}
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
func close(ctx context.Context, c io.Closer, err log.ErrorEntry) {
	if err != nil {
		log.Error(ctx, err)
	}
	if err := c.Close(); err != nil {
		log.Error(ctx, CloseError{Err: err})
	}
}

// logFilter logs when a new connection is accepted and closed.
func logFilter(ctx context.Context, c net.Conn, chain connFilters) context.Context {
	log.Log(ctx, AcceptedConnection{Remote: c.RemoteAddr()})
	ctx = chain.next(ctx, c)
	log.Log(ctx, ClosedConnection{Remote: c.RemoteAddr()})
	return ctx
}

// connIDFilter assigns an identifier to the connection and stores it as a
// context value.
func connIDFilter(ctx context.Context, c net.Conn, chain connFilters) context.Context {
	cid := make([]byte, 16)
	if _, err := rand.Read(cid); err != nil {
		close(ctx, c, GenerateConnectionIDError{Err: log.Alert(err)})
		return ctx
	}
	ctx = log.WithConnectionID(ctx, hex.EncodeToString(cid))
	log.Log(ctx, AssignedConnectionID{Remote: c.RemoteAddr()})
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
		close(ctx, c, PROXYProtocolError{Err: err})
		return ctx
	}
	if health {
		log.Log(ctx, HealthCheck{})
		close(ctx, c, nil)
		return ctx
	}
	if addr != nil {
		log.Log(ctx, PROXYProtocol{Address: addr})
	}

	// XXX: Put remote address into context for addrFilter. Remove once
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
			Certificates:             []tls.Certificate{cert},
			ClientAuth:               tls.RequestClientCert,
			PreferServerCipherSuites: true,
			SessionTicketsDisabled:   true,
			MinVersion:               tls.VersionTLS12,
		},
	}
	for _, name := range conf.CipherSuites {
		value, ok := tlsCipherSuites[name] // Generated above.
		if !ok {
			return nil, UnsupportedTLSCipherSuiteError{CipherSuite: name}
		}
		f.tlsConf.CipherSuites = append(f.tlsConf.CipherSuites, value)
	}
	return f, nil
}

func (f *tlsFilter) filter(ctx context.Context, c net.Conn, chain connFilters) context.Context {
	// Create the TLS connection.
	tlsc := tls.Server(c, f.tlsConf)

	// Before interacting with the client, check if we are done.
	select {
	case <-ctx.Done():
		close(ctx, tlsc, CancelingBeforeHandshake{Err: ctx.Err()})
		return ctx
	default:
	}

	// Explicitly perform the handshake to catch any errors early.
	deadline := time.Now().Add(time.Duration(f.serverConf.HandshakeTimeout) * time.Second)
	if err := tlsc.SetDeadline(deadline); err != nil {
		close(ctx, tlsc, SetHandshakeTimeoutError{Err: err})
		return ctx
	}
	if err := tlsc.Handshake(); err != nil {
		close(ctx, tlsc, HandshakeError{Err: err})
		return ctx
	}

	// Log the TLS connection details.
	state := tlsc.ConnectionState()
	log.Log(ctx, HandshakeComplete{
		Version:            state.Version,
		CipherSuite:        state.CipherSuite,
		ServerName:         state.ServerName,
		ClientCertificates: state.PeerCertificates,
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
		log.Log(header.Ctx, VotingEnded{})
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
			log.Error(header.Ctx, GenerateSessionIDError{Err: log.Alert(err)})
			return ErrInternal
		}
		header.SessionID = hex.EncodeToString(sid)
		entry = AssignedSessionID{}
	} else {
		entry = ReadSessionID{}
	}

	// Set session ID in logging context and log.
	header.Ctx = log.WithSessionID(header.Ctx, header.SessionID)
	log.Log(header.Ctx, entry)
	return chain.next(header)
}

// addrFilter re-logs the remote address of the connection after we have a
// SessionID.
//
// XXX: This is a temporary filter until the log monitor is capable of
// extracting the address based on ConnectionID.
func addrFilter(header *Header, chain headerFilters) error {
	log.Log(header.Ctx, RemoteAddress{Address: header.Ctx.Value(addrKey)})
	return chain.next(header)
}

// infoFilter logs any client-provided information about the platform used to
// perform the request, e.g., the operating system, and clears the fields from
// the header so that they are not included in the response.
func infoFilter(header *Header, chain headerFilters) error {
	log.Log(header.Ctx, OperatingSystem{OS: header.OS})
	header.OS = ""
	return chain.next(header)
}

// authFilter verifies the authentication token in the request, clears the
// fields from the header so that they are not included in the response, and
// stores the authenticated client's name in the context as a value.
type authFilter auth.Auther

func (a authFilter) filter(header *Header, chain headerFilters) error {
	if len(header.AuthMethod) > 0 {
		log.Log(header.Ctx, Authenticating{
			Method: header.AuthMethod,
			Token:  log.Sensitive(header.AuthToken),
		})
		name, voteid, err := auth.Auther(a).Verify(
			header.Ctx, auth.Type(header.AuthMethod), header.AuthToken)
		if err != nil {
			log.Error(header.Ctx, AuthenticationError{Err: err})
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
		header.Ctx = context.WithValue(header.Ctx, authClientKey, name)
		log.Log(header.Ctx, Authenticated{ClientName: name})

		if len(voteid) > 0 {
			header.Ctx = context.WithValue(header.Ctx, voterIDKey, voteid)
			log.Log(header.Ctx, AuthenticationVoteID{VoteID: voteid})
		}
	}

	header.AuthMethod = ""
	header.AuthToken = nil
	return chain.next(header)
}

// identityFilter extracts a unique identifier from an authenticated client's
// name and stores it as a context value.
type identityFilter identity.Identifier

func (i identityFilter) filter(header *Header, chain headerFilters) error {
	if name := AuthenticatedClient(header.Ctx); name != nil {
		id, err := identity.Identifier(i)(name)
		if err != nil {
			log.Error(header.Ctx, IdentityError{Err: err})
			return ErrIneligible
		}
		header.Ctx = context.WithValue(header.Ctx, voterIDKey, id)
		log.Log(header.Ctx, Identity{Identity: id})
	}
	return chain.next(header)
}

// ageFilter determines the voter's age and checks if they are over a voting
// age limit.
type ageFilter age.Checker

func (a *ageFilter) filter(header *Header, chain headerFilters) error {
	if id := VoterIdentity(header.Ctx); len(id) > 0 {
		if err := (*age.Checker)(a).Check(id); err != nil {
			log.Error(header.Ctx, AgeError{Err: err})
			if errors.CausedBy(err, new(age.TooYoungError)) != nil {
				return ErrTooYoung
			}
			return ErrIneligible
		}
	}
	return chain.next(header)
}
