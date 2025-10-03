package rpc

import (
	"crypto/tls"

	"ivxv.ee/common/collector/command"
	"ivxv.ee/common/collector/command/exit"
	"ivxv.ee/common/collector/conf"
	"ivxv.ee/common/collector/cryptoutil"
	status "ivxv.ee/common/collector/status/client"
	client "ivxv.ee/common/collector/status/client/rpc"
	"ivxv.ee/common/collector/storage/etcd"
	"ivxv.ee/common/collector/yaml"
)

type Client struct {
	status.TLSDialer
}

// NewClient configures session status TLS client to communicate with a
// session status service.
func NewClient(c *command.C) (status.TLSDialer, int) {
	// Get session status client configuration from technical.yml
	observable := c.Conf.Technical.Status.Session
	if observable == nil {
		return nil, c.Error(exit.Config, SessionObservableNotConfiguredError{
			Description: _SESSIONSTATUS_CONF,
		},
			"failed to read session observable client from configuration")
	}

	// Get storage CA certificate from technical.yml (storage:conf:ca)
	var storageConf etcd.Conf
	err := yaml.Apply(c.Conf.Technical.Storage.Conf, &storageConf)
	if err != nil {
		return nil, c.Error(exit.Config, ReadStorageCAFromTechicalConfigError{Err: err,
			Description: _SESSIONSTATUS_STORAGE_CONF},
			"failed to read CA certificate from storage configuration:", err)
	}

	// Add CA certificate to in-memory certificate pool
	certPool, err := cryptoutil.PEMCertificatePool(storageConf.CA)
	if err != nil {
		return nil, c.Error(exit.Config, AddStorageCAToCAPoolError{Err: err,
			Description: _SESSIONSTATUS_STORAGE_CA},
			"failed to add storage CA to certificate pool:", err)
	}

	// Get filepath of a client TLS cert and key
	cert, key := conf.TLS(conf.Sensitive(c.Service.ID))

	// Parse client TLS certificate-key pair
	tlsCert, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, c.Error(exit.Config, ParseTLSKeyPairError{Err: err,
			Description: _SESSIONSTATUS_KEY_PAIR},
			"failed to parse TLS client certificate-key pair:", err)
	}

	// Get network segment of a client
	network, _ := c.Conf.Technical.Service(c.Service.ID)
	// List of services for a given network segment
	services := c.Conf.Technical.Services(network)
	// Read session status service address from services
	addr := services.SessionStatus[0].Address

	// Create session status RPC TLS client
	return &Client{
		TLSDialer: client.NewTLSClient(addr, &tls.Config{
			RootCAs:      certPool,
			Certificates: []tls.Certificate{tlsCert},
			MinVersion:   tls.VersionTLS12,
			ServerName:   observable.ServerName,
		}),
	}, 0
}
