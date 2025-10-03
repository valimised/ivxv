/*
The proxy service controls a locally running HAProxy instance.
*/
package main

import (
	"context"
	"flag"
	"os"
	"time"

	"ivxv.ee/common/collector/command"
	"ivxv.ee/common/collector/command/exit"
	"ivxv.ee/common/collector/conf"
	"ivxv.ee/common/collector/log"
	"ivxv.ee/common/collector/server"
	//ivxv:modules common/collector/container
)

var (
	cfgp = flag.String("haproxy-config", "/etc/haproxy/haproxy.cfg",
		"`path` where to write the HAProxy configuration file\n")

	pidp = flag.String("haproxy-pid", "/run/haproxy.pid", "`path` to the HAProxy pid file")

	procp = flag.String("proc", "/proc", "`path` where proc filesystem is mounted")

	tmplp = flag.String("haproxy-template", "/usr/share/ivxv/haproxy.cfg.tmpl",
		"`path` to the HAProxy configuration template\n")
)

// reload updates the HAProxy configuration and restarts HAProxy to make it use
// the new configuration.
func reload(ctx context.Context, c *conf.Technical,
	network string, service *conf.Service, until int) (code int, err error) {

	cfg, code, err := generate(c, network, service, *tmplp)
	if err != nil {
		return code, GenerateHAProxyConfigurationError{Err: err, Description: _PROXY_PARSE_TEMPLATE_EXEC}
	}

	if code, err = check(ctx, cfg); err != nil {
		return code, CheckHAProxyConfigurationError{Err: err, Description: _PROXY_PARSE_TEMPLATE_VERIFY}
	}

	if until >= command.Execute {
		//nolint:gosec // Keep the configuration file permissions.
		if err = os.WriteFile(*cfgp, cfg, 0644); err != nil {
			return exit.CantCreate, WriteConfigurationError{Err: err, Description: _PROXY_HAPROXY_WRITE}
		}
		log.Log(ctx, UpdatedHAProxyConfiguration{Path: *cfgp, Description: _PROXY_HAPROXY_UPDATE})

		// It is possible here that the just created configuration file
		// gets replaced before HAProxy is restarted. This should not
		// happen during normal operation and there is little point in
		// worrying about someone doing this maliciously, because then
		// they can just replace it and restart HAProxy whenever they
		// want.

		pid, code, err := readPIDFile(*pidp)
		if err != nil {
			return code, ReadHAProxyPIDError{Err: err, Description: _PROXY_HAPROXY_PIDFILE}
		}
		if err = restart(ctx, *procp, pid); err != nil {
			return exit.Unavailable, RestartHAProxyError{Err: err, Description: _PROXY_HAPROXY_RESTART}
		}
		log.Log(ctx, HAProxyRestarted{Description: _PROXY_HAPROXY_RESTARTED})
	}
	return
}

// hapid is the PID of the HAProxy master process.
type hapid int

// start sets p to the PID of the HAProxy master process.
func (p *hapid) start(ctx context.Context) error {
	// We need to wait until the service manager has restarted
	// HAProxy. Replace this with something that actually ensures new
	// instances of HAProxy are up.
	time.Sleep(500 * time.Millisecond)

	// Get the PID of the HAProxy master processes.
	pid, _, err := readPIDFile(*pidp)
	if err != nil {
		return EnableReadPIDError{Err: err, Description: _PROXY_HAPROXY_PIDFILE}
	}
	*p = hapid(pid)
	log.Log(ctx, HAProxyPID{PID: *p, Description: _PROXY_HAPROXY_PID})
	return nil
}

// check checks if the HAProxy master process still exists.
func (p *hapid) check(_ context.Context) error {
	if err := checkPID(*procp, int(*p)); err != nil {
		return CheckHAProxyError{PID: *p, Err: err, Description: _PROXY_NO_PID}
	}
	return nil
}

// stop does nothing, because we do not have permissions to stop HAProxy
// instances.
func (p *hapid) stop(_ context.Context) error {
	return nil
}

func main() {
	// Call proxymain in a separate function so that it can set up defers
	// and have them trigger before returning with a non-zero exit code.
	os.Exit(proxymain())
}

func proxymain() (code int) {
	c := command.NewWithoutStorage("ivxv-proxy", "")
	defer func() {
		code = c.Cleanup(code)
	}()

	var s *server.Controller
	var err error
	if c.Conf.Technical != nil {
		// Create new HAProxy configuration and reload the service.
		if code, err = reload(c.Ctx, c.Conf.Technical,
			c.Network, c.Service, c.Until); err != nil {

			return c.Error(code, ReloadConfigurationError{Err: err, Description: _PROXY_HAPROXY_RELOAD},
				"reloading new configuration failed:", err)
		}

		// Configure a new proxy service controller. This actually does
		// not control anything and just monitors if the HAProxy
		// processes are still running or not.
		var p hapid
		if s, err = server.NewController(&c.Conf.Version,
			p.start, p.check, p.stop); err != nil {

			return c.Error(exit.Config, ControllerConfError{Err: err, Description: _PROXY_HAPROXY_CONTROL},
				"failed to configure controller:", err)
		}
	}
	if c.Until >= command.Execute {
		if err = s.Control(c.Ctx); err != nil {
			return c.Error(exit.Unavailable, ControlError{Err: err, Description: _PROXY_HAPROXY_SERVE},
				"failed to control proxy service:", err)
		}
	}
	return exit.OK
}
