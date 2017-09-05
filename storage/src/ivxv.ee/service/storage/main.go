/*
The storage service controls a locally running etcd instance.
*/
package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"ivxv.ee/command"
	"ivxv.ee/command/exit"
	"ivxv.ee/conf"
	"ivxv.ee/cryptoutil"
	"ivxv.ee/log"
	"ivxv.ee/server"
	"ivxv.ee/storage"
	"ivxv.ee/storage/etcd"
	"ivxv.ee/yaml"
	//ivxv:modules container
)

type ctrl struct {
	wd    string
	args  []string
	waitc chan error
	cmd   *exec.Cmd
}

func newCtrl(cfg *conf.Technical, service *conf.Service, election string) (
	c *ctrl, code int, err error) {

	// Listening addresses cannot be hostnames and must be resolved.
	caddr, err := net.ResolveTCPAddr("tcp", service.Address)
	if err != nil {
		return nil, exit.Config, ResolveClientAddressError{
			Address: service.Address,
			Err:     err,
		}
	}
	cresv := caddr.String()

	paddr, err := net.ResolveTCPAddr("tcp", service.PeerAddress)
	if err != nil {
		return nil, exit.Config, ResolvePeerAddressError{
			Address: service.PeerAddress,
			Err:     err,
		}
	}
	presv := paddr.String()

	// Create the static cluster description using storage services from
	// the entire network, not only this segment.
	const protocol = "https://"
	var cluster []string
	for _, segment := range cfg.Network {
		for _, member := range segment.Services.Storage {
			cluster = append(cluster, fmt.Sprint(member.ID, "=", protocol, member.PeerAddress))
		}
	}

	c = new(ctrl)
	c.wd = conf.Sensitive(service.ID)

	// Check if the TLS certificate and key are loaded before we start etcd
	// which will fail if they are not.
	tlsPEMPath := filepath.Join(c.wd, "tls.pem")
	if _, err := os.Stat(tlsPEMPath); err != nil {
		return nil, exit.NoInput, TLSCertificateError{Err: err}
	}
	tlsKeyPath := filepath.Join(c.wd, "tls.key")
	if _, err := os.Stat(tlsKeyPath); err != nil {
		return nil, exit.NoInput, TLSKeyError{Err: err}
	}

	// Parse the CA certificate from the configuration and write to working
	// directory for etcd to access.
	var etcdCfg etcd.Conf
	if err := yaml.Apply(cfg.Storage.Conf, &etcdCfg); err != nil {
		return nil, exit.Config, EtcdConfigurationError{Err: err}
	}
	if _, err := cryptoutil.PEMCertificate(etcdCfg.CA); err != nil {
		return nil, exit.Config, EtcdCAParseError{Err: err}
	}
	tlsCAPath := filepath.Join(c.wd, "ca.pem")
	if err := ioutil.WriteFile(tlsCAPath, []byte(etcdCfg.CA), 0600); err != nil {
		return nil, exit.CantCreate, EtcdCAWriteError{Err: err}
	}

	// Assemble the command-line arguments.
	c.args = []string{
		// Member flags.
		"--name", service.ID,
		"--data-dir", filepath.Join(c.wd, "etcd"),
		"--listen-client-urls", protocol + cresv,
		"--listen-peer-urls", protocol + presv,

		// Clustering flags.
		"--advertise-client-urls", protocol + service.Address,
		"--initial-advertise-peer-urls", protocol + service.PeerAddress,
		"--initial-cluster", strings.Join(cluster, ","),
		"--initial-cluster-token", election,

		// Security flags.
		"--cert-file", tlsPEMPath,
		"--key-file", tlsKeyPath,
		"--client-cert-auth",
		"--trusted-ca-file", tlsCAPath,
		"--peer-client-cert-auth",
		"--peer-trusted-ca-file", tlsCAPath,
		"--peer-cert-file", tlsPEMPath,
		"--peer-key-file", tlsKeyPath,

		// Increase backend quota to 8 GB, the maximum size supported
		// before degraded performance. Default is 2 GB.
		"--quota-backend-bytes", "8589934592",
	}
	if cfg.Debug {
		c.args = append(c.args, "--debug")
	}

	c.waitc = make(chan error, 1)
	return
}

func (c *ctrl) start(ctx context.Context) error {
	// Do not use CommandContext: we do not want it to be killed, but will
	// perform a graceful stop ourselves.
	c.cmd = exec.Command("/usr/bin/etcd", c.args...) // nolint: gas, args are trusted.

	// Forward all etcd output to logger.
	r, w := io.Pipe()
	c.cmd.Stdout = w
	c.cmd.Stderr = w
	go logger(ctx, r)

	// etcd expects to be run under systemd with Type=notify. This means
	// that we can open our own notification socket and listen on it for
	// etcd to be ready: this way we can block start until the storage
	// service is actually ready.
	sockAddr := filepath.Join(c.wd, "notify.sock")
	conn, err := net.ListenPacket("unixgram", sockAddr)
	if err != nil {
		return ListenNotifyError{Address: sockAddr, Err: err}
	}
	defer os.Remove(sockAddr) // nolint: errcheck, ignore removal errors of temporary socket.
	defer conn.Close()        // nolint: errcheck, ignore close failure of read-only socket.
	c.cmd.Env = []string{"NOTIFY_SOCKET=" + sockAddr}

	// Start a goroutine which listens for the notification. It will put
	// the message in sockMsg and send the error (if any) on sockc.
	//
	// Note that this means sockMsg is NOT safe for use before something is
	// received on sockc!
	const ready = "READY=1"
	sockMsg := make([]byte, len(ready)+1) // Allocate extra byte to detect trailing garbage.
	sockc := make(chan error, 1)
	go func() {
		n, _, cerr := conn.ReadFrom(sockMsg)
		sockMsg = sockMsg[:n]
		sockc <- cerr
	}()

	// Start etcd.
	log.Log(ctx, StartingEtcd{Args: c.cmd.Args, Env: c.cmd.Env})
	if err = c.cmd.Start(); err != nil {
		w.Close() // nolint: errcheck, PipeWriter.Close always returns nil.
		return StartEtcdError{Err: err}
	}

	// Start a goroutine which waits on c.cmd and sends the result on waitc.
	go func() {
		c.waitc <- c.cmd.Wait()
		w.Close() // nolint: errcheck, PipeWriter.Close always returns nil.
	}()

	// Wait until etcd notifies that it is ready or its process exits.
	log.Log(ctx, WaitingForEtcdReady{})
	select {
	case <-ctx.Done(): // Stop waiting when cancelled.
		return ctx.Err()
	case err = <-c.waitc:
		return EtcdStartupError{Err: err}
	case err = <-sockc:
		// If there was an error or unexpected message in the
		// notification socket, then signal for etcd to stop, but do
		// not handle any stopping errors because the original one
		// supersedes them.
		if err != nil {
			c.stop(ctx) // nolint: errcheck, see above.
			return ReadNotifyError{Err: err}
		}
		if msg := string(sockMsg); msg != ready {
			c.stop(ctx) // nolint: errcheck, see above.
			return UnexpectedNotifyMessageError{Message: msg}
		}
	}
	return nil
}

func (c *ctrl) check(_ context.Context) error {
	select {
	case err := <-c.waitc:
		return EtcdTerminatedError{Err: err}
	default:
		return nil
	}
}

func (c *ctrl) stop(_ context.Context) error {
	if err := c.cmd.Process.Signal(os.Interrupt); err != nil {
		return InterruptEtcdError{PID: c.cmd.Process.Pid, Err: err}
	}

	err := <-c.waitc
	if exit, ok := err.(*exec.ExitError); ok {
		// Exiting because of interrupt signal is expected.
		if exit.Sys().(syscall.WaitStatus).Signal() == os.Interrupt {
			err = nil
		}
	}
	return err
}

func main() {
	// Call storagemain in a separate function so that it can set up defers
	// and have them trigger before returning with a non-zero exit code.
	os.Exit(storagemain())
}

func storagemain() (code int) {
	c := command.NewWithoutStorage("ivxv-storage", "")
	defer func() {
		code = c.Cleanup(code)
	}()

	var s *server.Controller
	var err error
	if c.Conf.Technical != nil {
		prot := c.Conf.Technical.Storage.Protocol
		if prot != storage.Etcd {
			return c.Error(exit.Config, StorageProtocolError{Protocol: prot},
				"etcd storage protocol must be used for the",
				"storage service, but protocol is", prot)
		}

		var cmd *ctrl
		if cmd, code, err = newCtrl(c.Conf.Technical, c.Service,
			c.Conf.Election.Identifier); err != nil {

			return c.Error(code, EtcdControllerError{Err: err},
				"failed to configure etcd options:", err)
		}

		if s, err = server.NewController(&c.Conf.Version,
			cmd.start, cmd.check, cmd.stop); err != nil {

			return c.Error(exit.Config, ControllerConfError{Err: err},
				"failed to configure controller:", err)
		}
	}

	// Control etcd during the voting period.
	if c.Until >= command.Execute {
		if err = s.Control(c.Ctx); err != nil {
			return c.Error(exit.Unavailable, ControlError{Err: err},
				"failed to control storage service:", err)
		}
	}
	return exit.OK
}
