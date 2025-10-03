/*
Package server provides common code for setting up collector services.
*/
package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/rpc"
	"runtime/debug"
	"sync"
	"time"

	"ivxv.ee/common/collector/age"
	"ivxv.ee/common/collector/auth"
	"ivxv.ee/common/collector/conf/version"
	"ivxv.ee/common/collector/cryptoutil"
	"ivxv.ee/common/collector/identity"
	"ivxv.ee/common/collector/log"
)

// S is a server which listens for incoming connections, filters them, and
// passes them to the configured handler.
type S struct {
	end     time.Time
	filters connFilters
	addr    *net.TCPAddr
	status  *status
}

// Conf is the configuration for a server instance.
type Conf struct {
	// CertPath and KeyPath are paths to the TLS-certificate and private
	// key to use for serving connections.
	CertPath string
	KeyPath  string

	// Address is the tcp host:port to listen on for requests.
	Address string

	// End is the time at which this server will start returning
	// ErrVotingEnd to all connections.
	End time.Time

	Filter  *FilterConf
	Version *version.V // Necessary for reporting server status.

	ClientCA string
}

// New creates a new server with the provided configuration and handler.
func New(c *Conf, handler interface{}) (*S, error) {
	s := &S{end: c.End}

	// Setup the RPC server with handler.
	r := rpc.NewServer()
	if err := r.Register(handler); err != nil {
		// err can only be non-nil if the handler is not a suitable
		// type, so this is a programmer error: panic.
		panic(err)
	}

	// Parse the TLS certificate-key pair.
	tlsCert, err := tls.LoadX509KeyPair(c.CertPath, c.KeyPath)
	if err != nil {
		return nil, TLSKeyPairError{Err: err, Description: _SERVER_CERT_K}
	}
	var certPool *x509.CertPool
	if c.ClientCA != "" {
		certPool, err = cryptoutil.PEMCertificatePool(c.ClientCA)
		if err != nil {
			return nil, ClientCAParsingError{Err: err, Description: _SERVER_CA}
		}
	}
	// Setup the chain of filters that serves a connection using r.
	if s.filters, err = newFilters(c.Filter, r, tlsCert, c.End, certPool); err != nil {
		return nil, FilterConfError{Err: err, Description: _SERVER_FILTERS}
	}

	// Resolve the requested listen address.
	if s.addr, err = net.ResolveTCPAddr("tcp", c.Address); err != nil {
		return nil, ResolveAddressError{Address: c.Address, Err: err,
			Description: _SERVER_ADDR}
	}

	// Create a new status reporter for this server.
	if s.status, err = newStatus(c.Version); err != nil {
		return nil, NewStatusError{Err: err, Description: _SERVER_STATUS}
	}
	return s, nil
}

// AuthConf is the configuration for S.WithAuth.
type AuthConf struct {
	Auth     auth.Auther
	Identity identity.Identifier
	Age      *age.Checker
}

// NewAuthConf initializes an AuthConf with the given configurations.
func NewAuthConf(a auth.Conf, i identity.Type, g *age.Conf) (AuthConf, error) {
	var conf AuthConf
	var err error
	if len(a) > 0 {
		if conf.Auth, err = auth.Configure(a); err != nil {
			return conf, AuthConfError{Err: err, Description: _SERVER_AUTH}
		}
	}

	if len(i) > 0 {
		// Determining the voter's unique identifier requires for them
		// to be authenticated.
		if conf.Auth == nil {
			return conf, IdentityWithoutAuthError{
				Description: _SERVER_ID_W_N_AUTH,
			}
		}
		if conf.Identity, err = identity.Get(i); err != nil {
			return conf, IdentityTypeError{Err: err, Description: _SERVER_IDENTITY}
		}
	}

	if g != nil && g.Limit > 0 {
		// Checking the voter's age requires a unique identifier.
		if conf.Identity == nil {
			return conf, AgeWithoutIdentityError{
				Description: _SERVER_AGE_W_N_IDENTITY,
			}
		}
		if conf.Age, err = age.New(g); err != nil {
			return conf, AgeConfError{Err: err,
				Description: _SERVER_AGE}
		}
	}
	return conf, nil
}

// WithAuth configures the server to authenticate clients and enables any
// features that depend on it, i.e., voter identifiers and age checking.
func (s *S) WithAuth(conf AuthConf) *S {
	s.filters.optional(conf.Auth, conf.Identity, conf.Age)
	return s
}

// Serve starts listening for incoming connections and serves them with the
// server handler on new goroutines. It blocks until ctx is cancelled or a
// non-temporary error occurs, after which it waits until all open connections
// are served.
func (s *S) Serve(ctx context.Context) error {
	l, err := net.ListenTCP("tcp", s.addr)
	if err != nil {
		return ServeListenError{Address: s.addr, Err: err,
			Description: _SERVER_LISTEN_TCP}
	}

	// Set the server state to serving and set to ended at s.end.
	if err := s.status.serving(); err != nil {
		return ServeStatusServingError{Err: err, Description: _SERVER_SYSD_SERVE}
	}

	//nolint:errcheck // Only returns nil or context canceled errors.
	go wait(ctx, s.end, func(ctx context.Context) error { // Required by wait.
		if err := s.status.ended(); err != nil {
			log.Error(ctx, ServeStatusEndedError{Err: err,
				Description: _SERVER_CLOSED})
		}
		return nil
	})

	var wg sync.WaitGroup
	defer func() {
		// Wait until all connections are handled.
		wg.Wait()
		log.Log(ctx, AllConnectionsClosed{Description: _SERVER_CLOSE_ALL})
	}()

	errc := make(chan error, 1)
	go func() {
		log.Log(ctx, AcceptingConnections{Address: l.Addr(),
			Description: _SERVER_ACC})
		for {
			conn, err := l.Accept()
			if err != nil {

				// Send a non-temporary error on the channel
				// and stop accepting.
				//
				// Note that this path will also be taken in
				// the non-error case where the listener was
				// closed. This is okay, since then Serve is no
				// longer receiving on errc and the error in
				// the buffered channel will just get garbage
				// collected and never logged.
				errc <- log.Alert(AcceptError{Err: err,
					Description: _SERVER_ACC_FAIL})
				return
			}

			// Handle the accepted connection.
			wg.Add(1)
			go func(c net.Conn) {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						log.Error(ctx, ConnectionPanicError{
							Err:         log.Alert(fmt.Errorf("%v", r)),
							Stack:       string(debug.Stack()),
							Description: _SERVER_CONN_PANIC,
						})
						// Ensure c is closed. This is
						// a last resort close without
						// any TLS cleanup.
						c.Close()
					}
				}()
				s.filters.next(ctx, c)
			}(conn)
		}
	}()

	// Wait until context is cancelled or an error occurs.
	select {
	case <-ctx.Done():
	case err := <-errc:
		log.Error(ctx, AcceptingConnectionFailed{Err: err, Description: _SERVER_CONN_FAIL})
	}

	// If updating the status returns an error, then only log it: do not
	// skip closing the listener.
	if err := s.status.stopping(); err != nil {
		log.Error(ctx, ServeStatusStoppingError{Err: err,
			Description: _SERVER_CONN_STOP})
	}

	if err := l.Close(); err != nil {
		return CloseListenerError{Err: err, Description: _SERVER_CLOSED_GRACE}
	}
	log.Log(ctx, ListeningSocketClosed{Description: _SERVER_SYSD_CLOSED})
	return nil
}

// ServeAt waits until start and then calls Serve.
func (s *S) ServeAt(ctx context.Context, start time.Time) error {
	// Ensure that the server can bind to s.addr. Preferrably we would
	// create and bind a socket here and only start listening on it in
	// Serve. However the the net package does not support this very well
	// unless we create a socket ourselves using the syscall package.
	//
	// Another option would be to create the listener here and not accept
	// any connections until Serve is called, but that would cause the
	// kernel to start establishing connections prematurely. Take the
	// naive approach of attempting to listen on the socket here and
	// immediately closing it. This does not guarantee that Serve will be
	// able to listen on the address, but at least performs some elementary
	// checks.
	l, err := net.ListenTCP("tcp", s.addr)
	if err != nil {
		return ServeAtListenError{Address: s.addr, Err: err,
			Description: _SERVER_LISTEN_TCP}
	}
	if err := l.Close(); err != nil {
		return ServeAtCloseListenerError{Err: err, Description: _SERVER_CLOSED_GRACE}
	}

	if err := s.status.waiting(); err != nil {
		return ServeAtStatusWaitingError{Err: err,
			Description: _SERVER_SYSD_WAIT}
	}
	return waitStart(ctx, start, s.Serve)
}
