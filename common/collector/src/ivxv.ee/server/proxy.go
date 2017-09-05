package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net"
	"os"
	"syscall"
	"time"

	"ivxv.ee/log"
)

var proxySig = []byte{0x0D, 0x0A, 0x0D, 0x0A, 0x00, 0x0D, 0x0A, 0x51, 0x55, 0x49, 0x54, 0x0A}

// readPROXY attempts to read a PROXY protocol prefix and log its contents. It
// returns a possibly wrapped connection to use for further operations and if
// this was a health check.
func readPROXY(ctx context.Context, c net.Conn) (wc net.Conn, health bool, err error) {
	// We are willing to read less, so accept an unexpected EOF.
	header := make([]byte, 16)
	if err = readTimeout(c, header); err != nil && err != io.ErrUnexpectedEOF {
		return c, false, ReadPROXYHeaderError{Err: err}
	}

	if !bytes.Equal(header[:12], proxySig) {
		// The connection is not prefixed with the PROXY protocol:
		// create a wrapper connection which puts back the read bytes.
		return &prefixConn{Conn: c, prefix: header}, false, nil
	}

	// Determine if this is a local or proxied connection.
	var local bool
	switch vercmd := header[12]; vercmd {
	case 0x20:
		local = true
	case 0x21:
	default:
		return c, false, InvalidPROXYVerCmdError{VerCmd: vercmd}
	}

	extralen := int(binary.BigEndian.Uint16(header[14:16]))
	if !local {
		// For proxied connections, determine the transport
		// protocol and source address and port.
		var addrblock []byte
		var ip net.IP
		var port uint16
		switch transport := header[13]; transport {
		case 0x11: // TCP over IPv4.
			addrblock = make([]byte, 12)
			if err = readTimeout(c, addrblock); err != nil {
				return c, false, ReadPROXYTCPIP4AddressError{Err: err}
			}
			ip = addrblock[:4]
			port = binary.BigEndian.Uint16(addrblock[8:10])
		case 0x21: // TCP over IPv6.
			addrblock = make([]byte, 36)
			if err = readTimeout(c, addrblock); err != nil {
				return c, false, ReadPROXYTCPIP6AddressError{Err: err}
			}
			ip = addrblock[:16]
			port = binary.BigEndian.Uint16(addrblock[32:34])
		default:
			return c, false, InvalidPROXYTransportError{Transport: transport}
		}
		log.Log(ctx, PROXYAddress{IP: ip, Port: port})
		extralen -= len(addrblock)
	}

	if extralen > 0 {
		// Discard any extra bytes.
		null := make([]byte, extralen)
		if _, err = io.ReadFull(c, null); err != nil {
			return c, false, DiscardPROXYExtraError{Err: err}
		}
	}

	// Although the PROXY protocol says that the local command should be
	// used for health checks, HAProxy uses proxy commands. In order to
	// distinguish health checks from actual connections we try to read
	// from the connection: in case of a health check the connection will
	// be reset by HAProxy. For actual connections we should be able to
	// read at minimum the TLS ClientHello that HAProxy used to dispatch
	// the connection.
	// XXX: Can we determine if this was a health check without having to
	//      use prefixConn to put back the read byte?
	one := []byte{0}
	if err = readTimeout(c, one); err != nil {
		if ne, ok := err.(*net.OpError); ok {
			if se, ok := ne.Err.(*os.SyscallError); ok {
				if se.Err == syscall.ECONNRESET {
					return c, true, nil
				}
			}
		}
		return c, false, ReadPostPROXYError{Err: err}
	}
	return &prefixConn{Conn: c, prefix: one}, false, nil
}

func readTimeout(c net.Conn, buf []byte) error {
	// This timeout does not need to be configurable, because the messages
	// are small and coming from the proxy server.
	if err := c.SetDeadline(time.Now().Add(time.Second)); err != nil {
		return SetPROXYTimeoutError{Err: err}
	}
	_, err := io.ReadFull(c, buf)
	return err
}

// prefixConn is a net.Conn wrapper which reads the prefix before it continues
// reading the connection.
type prefixConn struct {
	net.Conn
	prefix []byte
}

func (c *prefixConn) Read(b []byte) (n int, err error) {
	if len(c.prefix) > 0 {
		n = copy(b, c.prefix)
		c.prefix = c.prefix[n:]
		return
	}
	return c.Conn.Read(b)
}
