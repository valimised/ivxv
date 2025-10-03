// The sessionstatus service (Session status) is used to provide a secure
// way of transferring SessionID during client-server communication.
// Session status service is an internal-network microservice.
package main

import (
	"os"
	"time"

	"ivxv.ee/common/collector/command"
	"ivxv.ee/common/collector/command/exit"
	"ivxv.ee/common/collector/conf"
	"ivxv.ee/common/collector/server"
	"ivxv.ee/common/collector/storage/etcd"
	"ivxv.ee/common/collector/yaml"
	internal "ivxv.ee/sessionstatus/internal/rpc"
	//ivxv:modules common/collector/auth
	//ivxv:modules common/collector/container
	//ivxv:modules common/collector/storage
)

func main() {
	// Call sessionstatus in a separate function so that it can set up defers
	// and have them trigger before returning with a non-zero exit code.
	os.Exit(sessionstatus())
}

func sessionstatus() (code int) {
	c := command.New("ivxv-sessionstatus", "")
	defer func() {
		code = c.Cleanup(code)
	}()

	var start, stop time.Time
	var err error

	if elec := c.Conf.Election; elec != nil {
		// Status server starts in a test-voting period
		if start, err = elec.ServiceStartTime(); err != nil {
			return c.Error(exit.Config, StartTimeError{Err: err,
				Description: _SESSIONSTATUS_START},
				"bad service start time:", err)
		}

		// Status server shuts down at the same time as all services do
		if stop, err = elec.ServiceStopTime(); err != nil {
			return c.Error(exit.Config, StopTimeError{Err: err,
				Description: _SESSIONSTATUS_STOP},
				"bad election stop time:", err)
		}
	}

	// Extract x509 CA certificate of a storage service
	// from election.yml `storage:conf:ca`.
	// All IVXV services have common x509 CA certificate,
	// which means we can use same CA for a status server as
	// ClientCAs and as RootCAs for any IVXV services which
	// will establish TLS connection with sessionstatus service
	var storageConf etcd.Conf
	err = yaml.Apply(c.Conf.Technical.Storage.Conf, &storageConf)
	if err != nil {
		return c.Error(exit.Config, GetStorageCAError{Err: err,
			Description: _SESSIONSTATUS_STORAGE_CONF},
			"failed to get CA cert from a storage config:", err)
	}

	// Register Session status server as an RPC server
	var rpc *internal.RPC

	// Create desired repository for the server
	r := c.Storage.SessionStatusRepository()
	repository := internal.NewStatusRepository(r)
	// Create desired handler for the server
	rpc = internal.NewHandler(repository)

	var s *server.S

	if c.Conf.Technical != nil {
		cert, key := conf.TLS(conf.Sensitive(c.Service.ID))
		if s, err = server.New(&server.Conf{
			// Ensure that cert's `X509v3 Subject Alternative Name:DNS:`
			// value is THE SAME AS `sessionStatusConf.ServerName`,
			// e.g. if your cert has
			// X509v3 Subject Alternative Name:DNS:session.status.inttest.ivxv.ee
			// and
			// sessionStatusConf.ServerName == session.status.inttest.ivxv.ee
			// then OK
			CertPath: cert,
			KeyPath:  key,
			Address:  c.Service.Address,
			End:      stop,
			Filter:   &c.Conf.Technical.Filter,
			Version:  &c.Conf.Version,
			// will set `tls.RequireAndVerifyClientCert` to the server TLS,
			// which means that any client should include RootCAs in their
			// TLS configuration
			ClientCA: storageConf.CA,
		}, rpc); err != nil {
			return c.Error(exit.Config, ServerConfError{Err: err,
				Description: _SESSIONSTATUS_SERVER},
				"failed to configure server:", err)
		}
	}

	// Start listening for incoming connections during the voting period.
	if c.Until >= command.Execute {
		if err = s.ServeAt(c.Ctx, start); err != nil {
			return c.Error(exit.Unavailable, ServeError{Err: err,
				Description: _SESSIONSTATUS_SERVER_SERVE},
				"failed to serve choices service:", err)
		}
	}
	return exit.OK
}
