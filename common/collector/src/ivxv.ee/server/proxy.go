package server

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"os"
	"syscall"
	"time"
)

var proxySig = []byte{0x0D, 0x0A, 0x0D, 0x0A, 0x00, 0x0D, 0x0A, 0x51, 0x55, 0x49, 0x54, 0x0A}

// readPROXY attempts to read a PROXY protocol prefix and get the client's
// address from it. It returns a possibly wrapped connection to use for further
// operations, the read address, and if this was a health check.
func readPROXY(c net.Conn) (wc net.Conn, addr net.Addr, health bool, err error) {
	// We are willing to read less, so accept an unexpected EOF.
	header := make([]byte, 16)
	if err = readTimeout(c, header); err != nil && err != io.ErrUnexpectedEOF {
		return c, nil, false, ReadPROXYHeaderError{Err: err}
	}

	if !bytes.Equal(header[:12], proxySig) {
		// The connection is not prefixed with the PROXY protocol:
		// create a wrapper connection which puts back the read bytes.
		return &prefixConn{Conn: c, prefix: header}, nil, false, nil
	}

	// Determine if this is a local or proxied connection.
	var local bool
	switch vercmd := header[12]; vercmd {
	case 0x20:
		local = true
	case 0x21:
	default:
		return c, nil, false, InvalidPROXYVerCmdError{VerCmd: vercmd}
	}

	extralen := int(binary.BigEndian.Uint16(header[14:16]))
	if !local {
		// For proxied connections, determine the transport
		// protocol and source address and port.
		switch transport := header[13]; transport {
		case 0x11: // TCP over IPv4.
			if addr, err = readTCPAddr(c, 4); err != nil {
				return c, nil, false, ReadPROXYTCPIP4AddressError{Err: err}
			}
			extralen -= 12
		case 0x21: // TCP over IPv6.
			if addr, err = readTCPAddr(c, 16); err != nil {
				return c, nil, false, ReadPROXYTCPIP6AddressError{Err: err}
			}
			extralen -= 36
		default:
			return c, nil, false, InvalidPROXYTransportError{Transport: transport}
		}
	}

	if extralen > 0 {
		// Discard any extra bytes.
		null := make([]byte, extralen)
		if _, err = io.ReadFull(c, null); err != nil {
			return c, nil, false, DiscardPROXYExtraError{Err: err}
		}
	}

	// Although the PROXY protocol says that the local command should be
	// used for health checks, HAProxy uses proxy commands. In order to
	// distinguish health checks from actual connections we try to read
	// from the connection: in case of a health check the connection will
	// be reset by HAProxy. For actual connections we should be able to
	// read at minimum the TLS ClientHello that HAProxy used to dispatch
	// the connection.
	one := []byte{0}
	if err = readTimeout(c, one); err != nil {
		if ne, ok := err.(*net.OpError); ok {
			if se, ok := ne.Err.(*os.SyscallError); ok {
				if se.Err == syscall.ECONNRESET {
					return c, addr, true, nil
				}
			}
		}
		return c, nil, false, ReadPostPROXYError{Err: err}
	}
	return &prefixConn{Conn: c, prefix: one}, addr, false, nil
}

func readTCPAddr(c net.Conn, size int) (net.Addr, error) {
	addrblock := make([]byte, 2*size+4) // 2 size addr, 2 uint16 port
	if err := readTimeout(c, addrblock); err != nil {
		return nil, err
	}
	tcp := new(net.TCPAddr)
	tcp.IP = addrblock[:size]
	tcp.Port = int(binary.BigEndian.Uint16(addrblock[2*size : 2*size+2]))
	return tcp, nil
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
		return n, nil
	}
	return c.Conn.Read(b)
}
