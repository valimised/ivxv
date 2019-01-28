package server

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
)

// Header is a common protocol message header, which should be embedded in the
// beginning of all requests and responses.
type Header struct {
	// Ctx gets injected into requests, but is not part of the transmitted
	// message.
	//
	// Although the context package specifically advises against storing
	// contexts inside struct types, the net/rpc package expects all
	// handler methods to only take two arguments, which means we cannot
	// pass context as the first argument. Instead of doing some complex
	// reflection workaround, we will ignore the context package advice and
	// put the context here. This may change if the net/rpc package starts
	// natively supporting contexts.
	Ctx context.Context `json:"-"`

	// SessionID is a unique session identifier which is generated when the
	// client first connects and should be included in all following
	// requests to help tie connections of a single session together.
	SessionID string `size:"32"`

	// OS is the client-provided information about the operating system
	// that they are using.
	OS string `json:",omitempty" size:"100"`

	// AuthMethod is the client authentication method used. Not included in
	// the response.
	AuthMethod string `json:",omitempty" size:"10"`

	// AuthToken is an authentication token used by authentication
	// verifiers to authenticate the client. It may be omitted, depending
	// on the authentication method. Not included in the response.
	AuthToken []byte `json:",omitempty" size:"16000"`
}

// header is an unexported interface to check if a message contains a Header.
type header interface {
	header() *Header
}

func (h *Header) header() *Header {
	return h
}

// key is the key type of values stored in context by this package.
type key int

const (
	tlsClientKey  key = iota // Context key for TLS client certificates.
	authClientKey            // Context key for authenticated client's distinguished name.
	voteIDKey                // Context key for vote identifier from authentication token.
	voterIDKey               // Context key for authenticated client's unique identifier.

	// Keys only used internally.
	addrKey // Context key for connection's remote address.
)

// TLSClient returns the list of certificates presented by the TLS client.
//
// NB! The server package only requests the client's certificate and, if one is
// provided, checks if the client is in possession of the private key: it does
// NOT verify the certificate. If client authentication is required, then it is
// up to the authentication module to verify the certificate.
func TLSClient(ctx context.Context) []*x509.Certificate {
	if val := ctx.Value(tlsClientKey); val != nil {
		return val.([]*x509.Certificate)
	}
	return nil
}

// AuthenticatedClient returns the name of the authenticated client or nil if
// no authentication was done in this context.
func AuthenticatedClient(ctx context.Context) *pkix.Name {
	if val := ctx.Value(authClientKey); val != nil {
		return val.(*pkix.Name)
	}
	return nil
}

// VoteIdentifier returns the vote identifier specified by the authentication
// token or nil if no authentication was done in this context or the used
// authentication method does not specify a vote identifier.
func VoteIdentifier(ctx context.Context) []byte {
	if val := ctx.Value(voteIDKey); val != nil {
		return val.([]byte)
	}
	return nil
}

// VoterIdentity returns the unique identifier of the authenticated client or
// empty string if no authentication was done in this context.
func VoterIdentity(ctx context.Context) string {
	if val := ctx.Value(voterIDKey); val != nil {
		return val.(string)
	}
	return ""
}
