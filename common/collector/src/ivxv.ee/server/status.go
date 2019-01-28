package server

import (
	"bytes"
	"encoding/json"
	"net"
	"os"
	"sync"
	"time"

	"ivxv.ee/conf/version"
)

type message struct {
	Status  string
	Version *json.RawMessage
}

// status is used for notifying systemd about the server's status.
type status struct {
	socket string
	msg    message
	lock   sync.Mutex
}

// newStatus creates a new status reporter which reports the version v
// alongside the server status.
func newStatus(v *version.V) (*status, error) {
	socket := os.Getenv("NOTIFY_SOCKET")
	if len(socket) == 0 {
		// If the notify socket is unset or empty, then return a nil
		// status where state changes do nothing.
		return nil, nil
	}

	// Precompute the version JSON.
	jsonver, err := json.Marshal(v)
	if err != nil {
		return nil, PrecomputeVersionJSONError{Err: err}
	}

	return &status{
		socket: socket,
		msg:    message{Version: (*json.RawMessage)(&jsonver)},
	}, nil
}

// waiting sets the systemd status string to "waiting" and notifies systemd
// that the server is ready.
func (s *status) waiting() error { return s.set("waiting", "READY=1") }

// serving sets the systemd status string to "serving" and notifies systemd
// that the server is ready. It is okay to notify systemd multiple times.
func (s *status) serving() error { return s.set("serving", "READY=1") }

// ended sets the systemd status string to "ended".
func (s *status) ended() error { return s.set("ended") }

// stopping sets the systemd status string to "stopping" and notifies systemd
// that the server is stopping.
func (s *status) stopping() error { return s.set("stopping", "STOPPING=1") }

// set updates the status and sends it and any extra values to systemd over the
// notification socket.
func (s *status) set(status string, extra ...string) (err error) {
	if s == nil {
		return nil
	}
	s.lock.Lock()
	defer s.lock.Unlock()

	message := bytes.NewBufferString("STATUS=")
	s.msg.Status = status
	if err = json.NewEncoder(message).Encode(s.msg); err != nil {
		return JSONEncodeStatusError{Err: err}
	}
	for _, line := range extra {
		message.WriteByte('\n')
		message.WriteString(line)
	}

	conn, err := net.DialTimeout("unixgram", s.socket, time.Second)
	if err != nil {
		return DialNotifySocketError{Socket: s.socket, Err: err}
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil && err == nil {
			err = CloseNotifyConnectionError{Err: err}
		}
	}()

	if err = conn.SetDeadline(time.Now().Add(time.Second)); err != nil {
		return SetDeadlineNotifyConnectionError{Err: err}
	}
	if _, err = message.WriteTo(conn); err != nil {
		return WriteNotifyConnectionError{Err: err}
	}
	return nil
}
