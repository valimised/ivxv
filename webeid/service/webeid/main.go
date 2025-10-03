// Package main provides a Web eID authentication method against IVXV backend.
package main

import (
	"os"
	"time"

	"ivxv.ee/common/collector/auth"
	"ivxv.ee/common/collector/auth/ticket"
	_ "ivxv.ee/common/collector/auth/tls"
	"ivxv.ee/common/collector/command"
	"ivxv.ee/common/collector/command/exit"
	"ivxv.ee/common/collector/conf"
	"ivxv.ee/common/collector/cookie"
	"ivxv.ee/common/collector/server"
	"ivxv.ee/common/collector/status/client"
	internal "ivxv.ee/webeid/internal/sessionstatus/rpc"
	//ivxv:modules common/collector/container
)

// RPC is a handler for Web eID service calls.
type RPC struct {
	status  client.Verifier
	cookie  *cookie.C
	auther  *server.AuthConf
	origin  []byte
	authEnd time.Time
	ticket  *ticket.T
}

func main() {
	os.Exit(webeidmain())
}

func webeidmain() (code int) {
	c := command.NewWithoutStorage("ivxv-webeid", "")
	defer func() {
		code = c.Cleanup(code)
	}()

	// Configure session status client
	statusClient, errCode := internal.NewClient(c)
	if statusClient == nil || errCode != 0 {
		return errCode
	}

	// Check that origin is correct HTTPS
	if !server.VerifyHTTPSOrigin(c.Service.Origin) {
		return c.Error(exit.Config, BadUrlSchema{Description: _WEBEID_BAD_URL},
			"bad service origin URL:", c.Service.Origin)
	}

	// Create new RPC instance and start the session cleaner.
	rpc := &RPC{
		origin: []byte(c.Service.Origin),
		status: statusClient,
	}

	var start, stop time.Time
	var authConf server.AuthConf
	var err error

	if elec := c.Conf.Election; elec != nil {
		// Check election configuration time values - service start
		if start, err = c.Conf.Election.ServiceStartTime(); err != nil {
			return c.Error(exit.Config, StartTimeError{Err: err,
				Description: _WEBEID_START},
				"bad service start time:", err)
		}

		// Check election configuration time values - election stop
		if rpc.authEnd, err = c.Conf.Election.ElectionStopTime(); err != nil {
			return c.Error(exit.Config, ElectionStopTimeError{Err: err,
				Description: _WEBEID_AUTH_STOP},
				"bad election stop time:", err)
		}

		// Check election configuration time values - service stop
		if stop, err = c.Conf.Election.ServiceStopTime(); err != nil {
			return c.Error(exit.Config, ServiceStopTimeError{Err: err,
				Description: _WEBEID_STOP},
				"bad service stop time:", err)
		}

		if rpc.ticket, err = ticket.NewFromSystem(); err != nil {
			return c.Error(exit.Config, TicketConfError{Err: err,
				Description: _WEBEID_TICKET},
				"failed to configure ticket manager:", err)
		}

		// Auther for Web eID client authentication in RPC authFilter,
		// after server.Header.AuthToken is issued
		ticketConf, ok := c.Conf.Election.Auth[auth.Ticket]
		if !ok {
			return c.Error(exit.Config, TicketAuthError{Description: _WEBEID_TICKET_AUTH},
				"ticket authentication is mandatory for webeid")
		}
		if authConf, err = server.NewAuthConf(auth.Conf{auth.Ticket: ticketConf},
			elec.Identity, &elec.Age); err != nil {
			return c.Error(exit.Config, ServerTicketAuthConfError{Err: err,
				Description: _WEBEID_AUTH},
				"failed to configure client ticket authentication:", err)
		}

		// Auther for Web eID auth token TLS validation in RPC.Token method
		tlsConf, ok := c.Conf.Election.Auth[auth.TLS]
		if !ok {
			return c.Error(exit.Config, TLSAuthError{Description: _WEBEID_TLS_AUTH},
				"TLS authentication is mandatory for webeid")
		}
		var auther server.AuthConf
		if auther, err = server.NewAuthConf(auth.Conf{auth.TLS: tlsConf},
			"", nil); err != nil {
			return c.Error(exit.Config, ServerTLSAuthConfError{Err: err,
				Description: _WEBEID_TLS_AUTH_CFG},
				"failed to configure client TLS authentication:", err)
		}
		// This allows to perform TLS validation inside RPC.Token or any other
		// RPC method
		rpc.auther = &auther
	}

	var s *server.S
	if c.Conf.Technical != nil {
		// Configure a new server with the service instance
		// configuration and the RPC handler instance.
		cert, key := conf.TLS(conf.Sensitive(c.Service.ID))
		if s, err = server.New(&server.Conf{
			CertPath: cert,
			KeyPath:  key,
			Address:  c.Service.Address,
			End:      stop,
			Filter:   &c.Conf.Technical.Filter,
			Version:  &c.Conf.Version,
		}, rpc); err != nil {
			return c.Error(exit.Config, ServerConfError{Err: err,
				Description: _WEBEID_SERVER},
				"failed to configure server:", err)
		}

		// cookie is used as a shared secret to encrypt Bearer tokens.
		// Bearer tokens are used to provide stateless nonce verification
		// for Web eID authentication
		rpc.cookie, err = ticket.NewFromSystemAsCookie()
		if err != nil {
			return c.Error(exit.Config, ReadSharedSecretRPCConfError{Err: err,
				Description: _WEBEID_BEARER_COOKIE},
				"failed to read RPC shared secret (cookie):", err)
		}
	}

	// Start listening for incoming connections during the voting period.
	if c.Until >= command.Execute {
		if err = s.WithAuth(authConf).ServeAt(c.Ctx, start); err != nil {
			return c.Error(exit.Unavailable, ServeError{Err: err,
				Description: _WEBEID_SERVER_SERVE},
				"failed to serve webeid service:", err)
		}
	}
	return exit.OK
}
