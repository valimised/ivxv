/*
Package server provides common code for setting up collector services.
*/
package server // import "ivxv.ee/server"

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/rpc"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"

	"ivxv.ee/age"
	"ivxv.ee/auth"
	"ivxv.ee/conf/version"
	"ivxv.ee/identity"
	"ivxv.ee/log"
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
	// Sensitive is the path to the service directory which contains
	// sensitive information that can not be passed through the
	// configuration, e.g., TLS keys.
	Sensitive string

	// Address is the tcp host:port to listen on for requests.
	Address string

	// End is the time at which this server will start returning
	// ErrVotingEnd to all connections.
	End time.Time

	Filter  *FilterConf
	Version *version.V // Necessary for reporting server status.
}

// New creates a new server with the provided configuration and handler.
func New(c *Conf, handler interface{}) (s *S, err error) {
	s = &S{end: c.End}

	// Setup the RPC server with handler.
	r := rpc.NewServer()
	if err = r.Register(handler); err != nil {
		// err can only be non-nil if the handler is not a suitable
		// type, so this is a programmer error: panic.
		panic(err)
	}

	// Parse the TLS certificate-key pair.
	tlsCert, err := tls.LoadX509KeyPair(
		filepath.Join(c.Sensitive, "tls.pem"),
		filepath.Join(c.Sensitive, "tls.key"))
	if err != nil {
		return nil, TLSKeyPairError{Err: err}
	}

	// Setup the chain of filters that serves a connection using r.
	s.filters = newFilters(c.Filter, r, tlsCert, c.End)

	// Resolve the requested listen address.
	if s.addr, err = net.ResolveTCPAddr("tcp", c.Address); err != nil {
		return nil, ResolveAddressError{Address: c.Address, Err: err}
	}

	// Create a new status reporter for this server.
	if s.status, err = newStatus(c.Version); err != nil {
		return nil, NewStatusError{Err: err}
	}
	return
}

// WithAuth configures the server to authenticate clients and enables any
// features that depend on it, i.e., voter identifiers and age checking.
func (s *S) WithAuth(a auth.Conf, i identity.Type, g *age.Conf) (err error) {
	var auther auth.Auther
	if len(a) > 0 {
		auther, err = auth.Configure(a)
		if err != nil {
			return AuthConfError{Err: err}
		}
	}

	var identifier identity.Identifier
	if len(i) > 0 {
		// Determining the voter's unique identifier requires for them
		// to be authenticated.
		if auther == nil {
			return IdentityWithoutAuthError{}
		}
		identifier, err = identity.Get(i)
		if err != nil {
			return IdentityTypeError{Err: err}
		}
	}

	var agechecker *age.Checker
	if g != nil && g.Limit > 0 {
		// Checking the voter's age requires a unique identifier.
		if identifier == nil {
			return AgeWithoutIdentityError{}
		}
		agechecker, err = age.New(g)
		if err != nil {
			return AgeConfError{Err: err}
		}
	}

	s.filters.optional(auther, identifier, agechecker)
	return
}

// Serve starts listening for incoming connections and serves them with the
// server handler on new goroutines. It blocks until ctx is cancelled or a
// non-temporary error occurs, after which it waits until all open connections
// are served.
func (s *S) Serve(ctx context.Context) (err error) {
	if err = s.status.ready(); err != nil {
		return ServeStatusReadyError{Err: err}
	}

	l, err := net.ListenTCP("tcp", s.addr)
	if err != nil {
		return ServeListenError{Address: s.addr, Err: err}
	}

	// Set the server state to serving and set to ended at s.end.
	if err = s.status.serving(); err != nil {
		return ServeStatusServingError{Err: err}
	}
	// nolint: errcheck, only returns nil or context canceled errors.
	go wait(ctx, s.end, func(ctx context.Context) error {
		if serr := s.status.ended(); err != nil {
			log.Error(ctx, ServeStatusEndedError{Err: serr})
		}
		return nil
	})

	var wg sync.WaitGroup
	defer func() {
		// Wait until all connections are handled.
		wg.Wait()
		log.Log(ctx, AllConnectionsClosed{})
	}()

	errc := make(chan error, 1)
	go func() {
		log.Log(ctx, AcceptingConnections{Address: l.Addr()})
		for {
			conn, lerr := l.Accept()
			if lerr != nil {
				if ne, ok := lerr.(net.Error); ok {
					if ne.Temporary() {
						log.Error(ctx, TemporaryAcceptError{Err: lerr})
						// Sleep a bit before retrying.
						time.Sleep(10 * time.Millisecond)
						continue
					}
				}

				// Send a non-temporary error on the channel
				// and stop accepting.
				//
				// Note that this path will also be taken in
				// the non-error case where the listener was
				// closed. This is okay, since then Serve is no
				// longer receiving on errc and the error in
				// the buffered channel will just get garbage
				// collected and never logged.
				errc <- log.Alert(AcceptError{Err: lerr})
				return
			}

			// Handle the accepted connection.
			wg.Add(1)
			go func(c net.Conn) {
				defer func() {
					if r := recover(); r != nil {
						log.Error(ctx, ConnectionPanicError{
							Err:   log.Alert(fmt.Errorf("%v", r)),
							Stack: string(debug.Stack()),
						})
						// Ensure c is closed. This is
						// a last resort close without
						// any TLS cleanup.
						_ = c.Close()
					}
					wg.Done()
				}()
				s.filters.next(ctx, c)
			}(conn)
		}
	}()

	// Wait until context is cancelled or an error occurs.
	select {
	case <-ctx.Done():
	case err = <-errc:
		log.Error(ctx, AcceptingConnectionFailed{Err: err})
	}

	// If updating the status or closing the listener returns an error,
	// then only log it. Returning a non-nil error will cause the service
	// manager to restart the service, which we do not want if we did not
	// have any prior errors.
	if serr := s.status.stopping(); serr != nil {
		log.Error(ctx, ServeStatusStoppingError{Err: err})
	}

	if cerr := l.Close(); cerr != nil {
		log.Error(ctx, CloseListenerError{Err: cerr})
	}
	log.Log(ctx, StoppedAcceptingConnections{})
	return
}

// ServeAt waits until start and then calls Serve.
func (s *S) ServeAt(ctx context.Context, start time.Time) error { // nolint: dupl
	// Ignore duplicate code from Controller.ControlAt: we want to keep
	// them separate for a clearer API.
	if err := s.status.ready(); err != nil {
		return ServeAtStatusReadyError{Err: err}
	}
	return waitStart(ctx, start, s.Serve)
}
