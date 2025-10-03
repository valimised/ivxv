package rpc

import (
	"crypto/tls"
	"net/rpc/jsonrpc"

	status "ivxv.ee/common/collector/status/client"
)

const tcp = "tcp"

// tlsClient uses TLS configuration tls to establish TLS connection
// to a status server on a given addr.
type tlsClient struct {
	addr string
	tls  *tls.Config
}

// NewTLSClient returns a new RPC TLS Client to the caller.
//
// TLS connection should be authenticated on both ends,
// which means that caller should include `RootCAs:` and
// its own `Certificates:` into TLS configuration conf.
//
// Also, if addr (without port and https:// prefix, e.g.
// session.status.inttest.ivxv.ee) doesn't match server certificate
// `X509v3 Subject Alternative Name:DNS:`, then client should
// also provide a ServerName to a TLS configuration conf.
func NewTLSClient(addr string, conf *tls.Config) status.TLSDialer {
	return &tlsClient{
		addr: addr,
		tls:  conf,
	}
}

func (r *tlsClient) TLSDial(req interface{}) (interface{}, error) {
	// Establish TLS connection, server should respond
	tlsConn, err := tls.Dial(tcp, r.addr, r.tls)
	if err != nil {
		return nil, TLSDialError{Err: err, Addr: r.addr, Description: _RPC_TLS}
	}

	// Any data that is passed to TLSDial should be of a
	// *StatusReq type, otherwise error
	statusReq, err := castAnyToStatusReq(req)
	if err != nil {
		return nil, CastAnyToStatusReqError{Err: err, Description: _RPC_ANY_TO_STATUSREQ}
	}

	// Pass TLS connection to RPC connection and establish it lazily, i.e.
	// establish RPC connection only if rpcConn.Call() is invoked
	rpcConn := jsonrpc.NewClient(tlsConn)

	// rpcConn.Call(..., ..., reply), where reply is a map[string]any
	resp := new(StatusResp)

	// RPC call to status server, resp.Response is a map[string]any
	err = rpcConn.Call(statusReq.ServiceMethod, statusReq.Request, &resp.Response)
	if err != nil {
		return nil, RPCCallError{
			ServiceMethod: statusReq.ServiceMethod,
			Err:           err,
			Description:   _RPC_CALL,
		}
	}

	return resp, nil
}
