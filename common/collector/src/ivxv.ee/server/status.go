package server

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
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
	socket    string
	msg       message
	readySent bool
	lock      sync.Mutex
}

// newStatus creates a new status reporter which reports the version v
// alongside the server status.
func newStatus(v *version.V) (s *status, err error) {
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

// ready sets the systemd status string to "waiting" and notifies systemd that
// the server is ready.
//
// Only the first call to ready will actually do anything: this way it can be
// called from both Serve/Control and ServeAt/ControlAt.
func (s *status) ready() error {
	// We cannot use s.set due to the pre-check and post-operation.
	if s == nil {
		return nil
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.readySent {
		return nil
	}
	if err := s.notify(fmt.Sprintf("READY=1\nSTATUS=%s", s.format("waiting"))); err != nil {
		return err
	}
	s.readySent = true
	return nil
}

// serving sets the systemd status string to "serving".
func (s *status) serving() error { return s.set("serving") }

// ended sets the systemd status string to "ended".
func (s *status) ended() error { return s.set("ended") }

// stopping sets the systemd status string to "stopping" and notifies systemd that
// the server is stopping.
func (s *status) stopping() error { return s.set("stopping", "STOPPING=1") }

func (s *status) set(status string, extra ...string) error {
	if s == nil {
		return nil
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.notify(strings.Join(append(extra, "STATUS="+s.format(status)), "\n"))
}

// notify sends state to systemd over the notification socket. state can be a
// format string using arguments a.
func (s *status) notify(message string) (err error) {
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
	if _, err = conn.Write([]byte(message)); err != nil {
		return WriteNotifyConnectionError{Err: err}
	}
	return
}

// format formats the server's status including the version as a string.
// Unsynchronized and should not be called directly.
func (s *status) format(status string) string {
	// We expect all JSON encodings succeed, since the possible values come
	// from a very small set controlled by us. Panic otherwise.
	s.msg.Status = status
	b, err := json.Marshal(s.msg)
	if err != nil {
		panic(fmt.Sprint("JSON encoding status failed: ", err))
	}
	return string(b)
}
