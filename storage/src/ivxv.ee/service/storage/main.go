/*
The storage service controls a locally running etcd instance.
*/
package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/coreos/etcd/clientv3"

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
	// Configuration for the etcd cluster client.
	bstrap  bool
	cli     clientv3.Config
	optime  time.Duration
	service *conf.Service
	members []*conf.Service

	// Configuration for starting the server.
	wd     string
	capath string
	capem  []byte
	nsock  string
	env    []string
	args   []string

	// Command for the started server and channel waiting for it to exit.
	cmd   *exec.Cmd
	waitc chan error

	// Cleanup function to call after etcd has stopped.
	cleanup func()
}

func newCtrl(cfg *conf.Technical, service *conf.Service, election string) (
	c *ctrl, code int, err error) {

	c = &ctrl{service: service, wd: conf.Sensitive(service.ID)}
	c.capath = filepath.Join(c.wd, "ca.pem")
	c.nsock = filepath.Join(c.wd, "notify.sock") // See ctrl.start for explanation.

	// Parse the storage configuration block as etcd configuration.
	var etcdCfg etcd.Conf
	if err = yaml.Apply(cfg.Storage.Conf, &etcdCfg); err != nil {
		return nil, exit.Config, EtcdConfigurationError{Err: err}
	}
	c.bstrap = bootstrap(service.ID, &etcdCfg)
	c.cli.DialTimeout = time.Duration(etcdCfg.ConnTimeout) * time.Second
	c.optime = time.Duration(etcdCfg.OpTimeout) * time.Second
	c.capem = []byte(etcdCfg.CA)

	// Parse the CA certificate and member TLS certificate and private key.
	c.cli.TLS = &tls.Config{MinVersion: tls.VersionTLS12}
	if c.cli.TLS.RootCAs, err = cryptoutil.PEMCertificatePool(etcdCfg.CA); err != nil {
		return nil, exit.Config, EtcdCAParseError{Err: err}
	}
	tlsPEMPath, tlsKeyPath := conf.TLS(c.wd)
	c.cli.TLS.Certificates = make([]tls.Certificate, 1)
	if c.cli.TLS.Certificates[0], err = tls.LoadX509KeyPair(tlsPEMPath, tlsKeyPath); err != nil {
		return nil, exit.DataErr, LoadTLSCertificateError{Err: err}
	}

	// Check that the TLS certificate is issued by the CA and has key usage
	// for both server and client auth.
	if err = checkTLS(c.cli.TLS.RootCAs, &c.cli.TLS.Certificates[0]); err != nil {
		return nil, exit.DataErr, err
	}

	// Collect all storage services in the entire network (not only this
	// segment) as client endpoints and cluster members.
	var cluster bytes.Buffer
	for _, segment := range cfg.Network {
		for _, member := range segment.Services.Storage {
			c.members = append(c.members, member)
			c.cli.Endpoints = append(c.cli.Endpoints, protocol(member.Address))
			fmt.Fprint(&cluster, ",", member.ID, "=", protocol(member.PeerAddress))
		}
	}
	cluster.Next(1) // Skip leading comma.

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

	// Assemble environment variables. Pass any ETCD_* environment
	// variables that we might want to tune in runtime. Be careful to not
	// add any environment variables related to cluster security: these
	// must all come from the signed configuration.
	c.env = []string{"NOTIFY_SOCKET=" + c.nsock}
	for _, key := range []string{
		"ETCD_SNAPSHOT_COUNT",     // For snapshot tuning.
		"ETCD_HEARTBEAT_INTERVAL", // For election tuning.
		"ETCD_ELECTION_TIMEOUT",   // For election tuning.
	} {
		if value, ok := os.LookupEnv(key); ok {
			c.env = append(c.env, fmt.Sprint(key, "=", value))
		}
	}

	// Assemble the command-line arguments.
	state := "new"
	if !c.bstrap {
		state = "existing"
	}
	c.args = []string{
		// Member flags.
		"--name", service.ID,
		"--data-dir", filepath.Join(c.wd, "etcd"),
		"--wal-dir", walDir(c.wd),
		"--listen-client-urls", protocol(cresv),
		"--listen-peer-urls", protocol(presv),

		// Clustering flags.
		"--advertise-client-urls", protocol(service.Address),
		"--initial-advertise-peer-urls", protocol(service.PeerAddress),
		"--initial-cluster", cluster.String(),
		"--initial-cluster-state", state,
		"--initial-cluster-token", election,

		// Security flags.
		"--cert-file", tlsPEMPath,
		"--key-file", tlsKeyPath,
		"--client-cert-auth",
		"--trusted-ca-file", c.capath,
		"--peer-client-cert-auth",
		"--peer-trusted-ca-file", c.capath,
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

func checkTLS(roots *x509.CertPool, cert *tls.Certificate) error {
	var err error
	if cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0]); err != nil {
		return ParseTLSCertificateError{Err: err}
	}
	if _, err = cert.Leaf.Verify(x509.VerifyOptions{
		Roots:     roots,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny}, // Checked manually.
	}); err != nil {
		return VerifyTLSCertificateError{Err: err}
	}

	// *x509.Certificate.Verify only checks if any of the specified
	// extended key usages is allowed. We want them all, so we need to
	// check manually.
required:
	for _, r := range []x509.ExtKeyUsage{
		x509.ExtKeyUsageClientAuth, // For connecting to other peers.
		x509.ExtKeyUsageServerAuth, // For serving connections.
	} {
		for _, u := range cert.Leaf.ExtKeyUsage {
			if r == u {
				continue required
			}
		}
		return TLSCertificateMissingUsageError{Usage: r}
	}
	return nil
}

// Since start contains a lot of conditional behavior based on the state of the
// storage instance, here is a small table which hopefully gives an overview
// on what happens when:
//
//   first boot of      | restart of         | first boot of  | restart of
//   bootstrap instance | bootstrap instance | added instance | added instance
//   -------------------|--------------------|----------------|---------------
//                      |                    | prune members  |
//                      |                    | add new member |
//   start etcd         | start etcd         | start etcd     | start etcd
//                      | wait for ready     | wait for ready | wait for ready
//                      | prune members      |                | prune members
//
// The first boot of bootstrap instances includes no cluster membership
// updates, since there is no cluster to update. It also contains no waiting
// for etcd to become ready, because bootstrap instances are started serially
// and it would cause a deadlock (the instance being started is waiting for
// other cluster members, but other cluster members are not started until the
// first one is ready).
//
// The first boot of an added (non-bootstrap) instance updates cluster
// membership before starting etcd, because the cluster needs to know in
// advance about new members. Pruning also has to happen before adding (instead
// of after starting etcd), because members cannot be added to an unhealthy
// cluster.
//
// Restarts of both kinds of instances are identical: the etcd instance is
// started first to ensure that the cluster is healthy and members are pruned
// after etcd has notified it is ready.
func (c *ctrl) start(ctx context.Context) error {
	first, err := firstBoot(c.wd)
	if err != nil {
		return err
	}
	if first && !c.bstrap { // First boot of added instance: add member.
		if err := updateMembers(ctx, c.cli, c.optime, c.members, c.service); err != nil {
			return err
		}
	}

	// Write the CA to the working directory location added to arguments in
	// newCtrl. Do this when starting and not before to avoid overwriting
	// the CA file when only checking the configuration.
	if err := ioutil.WriteFile(c.capath, c.capem, 0600); err != nil {
		return EtcdCAWriteError{Err: err}
	}

	// Create notification socket waiting for etcd to become ready. Do this
	// even if this is a bootstrap instance and the channel is never read,
	// so etcd does not complain about the missing socket.
	readyctx, cancelReady := context.WithCancel(context.Background())
	readyc, err := listenReady(readyctx, c.nsock)
	if err != nil {
		cancelReady()
		return err
	}

	// Do not use CommandContext: we do not want it to be killed, but will
	// perform a graceful stop ourselves.
	c.cmd = exec.Command("/usr/bin/etcd", c.args...) //nolint:gosec // Args are trusted.
	c.cmd.Env = c.env

	// Forward all etcd output to logger.
	r, w := io.Pipe()
	c.cmd.Stdout = w
	c.cmd.Stderr = w
	logger, err := newLogger()
	if err != nil {
		cancelReady()
		return err
	}
	go logger.log(ctx, r)

	// Create cleanup function for after etcd has stopped.
	c.cleanup = func() {
		w.Close()
		cancelReady() // Start notification socket cleanup.
		// Wait until readyc is closed, i.e., cleanup is done,
		// discarding any notification messages along the way.
		for range readyc {
		}
	}

	// Start etcd.
	log.Log(ctx, StartingEtcd{Args: c.cmd.Args, Env: c.cmd.Env})
	if err := c.cmd.Start(); err != nil {
		c.cleanup()
		return StartEtcdError{Err: err}
	}

	// Start a goroutine which waits on c.cmd and sends the result on waitc.
	go func() { c.waitc <- c.cmd.Wait() }()

	// Block until etcd is ready, unless this is the first boot of a
	// boostrap instance (see the comment for start).
	//
	// If there is a ready notification error or the context is cancelled
	// then manually stop etcd, but only log stopping errors: return the
	// original error which caused the stop.
	if !(first && c.bstrap) {
		log.Log(ctx, WaitingForEtcdReady{})
		select {
		case err := <-c.waitc:
			c.cleanup()
			return EtcdStartupError{Err: err}
		case err := <-readyc:
			if err != nil {
				if serr := c.stop(ctx); serr != nil {
					log.Error(ctx, EtcdNotifyStopError{Err: serr})
				}
				return err
			}
			if !first { // Restarting an instance: prune members.
				if err := updateMembers(ctx, c.cli, c.optime,
					c.members, nil); err != nil {

					// Do not stop an already started
					// healthy instance if pruning fails,
					// but do log it as an alert to signal
					// that the configuration was not
					// applied as expected.
					log.Error(ctx, EtcdPruneError{Err: log.Alert(err)})
				}
			}
		case <-ctx.Done():
			if err := c.stop(ctx); err != nil {
				log.Error(ctx, EtcdCanceledStopError{Err: err})
			}
			return ctx.Err()
		}
	}
	return nil
}

func listenReady(ctx context.Context, nsock string) (<-chan error, error) {
	// etcd expects to be run under systemd with Type=notify. This means
	// that we can open our own notification socket at nsock and listen on
	// it for etcd to be ready.
	conn, err := net.ListenPacket("unixgram", nsock)
	if err != nil {
		return nil, ListenNotifyError{Address: nsock, Err: err}
	}

	// Start a goroutine which listens for the notification and sends the
	// result on sockc.
	sockc := make(chan error, 1)
	go func() {
		const ready = "READY=1"
		sockMsg := make([]byte, len(ready)+1) // Allocate extra byte to detect trailing garbage.
		n, _, err := conn.ReadFrom(sockMsg)
		if err != nil {
			sockc <- ReadFromNotifyError{Err: err}
			return
		}
		if msg := string(sockMsg[:n]); msg != ready {
			sockc <- UnexpectedNotifyMessageError{Message: msg}
			return
		}
		sockc <- nil
	}()

	// Start a goroutine which forwards sockc to readyc and cleans up the
	// notification socket after something is received from sockc or the
	// context is cancelled.
	readyc := make(chan error, 1)
	go func() {
		defer close(readyc)
		defer os.Remove(nsock)
		defer conn.Close()
		select {
		case <-ctx.Done():
		case err := <-sockc:
			readyc <- err
		}
	}()
	return readyc, nil
}

func (c *ctrl) check(_ context.Context) error {
	select {
	case err := <-c.waitc:
		c.cleanup()
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
	c.cleanup()
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
