/*
The proxy service controls a locally running HAProxy instance.
*/
package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
	"time"

	"ivxv.ee/command"
	"ivxv.ee/command/exit"
	"ivxv.ee/conf"
	"ivxv.ee/log"
	"ivxv.ee/server"
	//ivxv:modules container
)

var (
	cfgp = flag.String("haproxy-config", "/etc/haproxy/haproxy.cfg",
		"`path` where to write the HAProxy configuration file\n    \t")

	pidp = flag.String("haproxy-pid", "/run/haproxy.pid", "`path` to the HAProxy pid file")

	socketp = flag.String("haproxy-socket", "/run/haproxy/admin.sock",
		"`path` where to open the HAProxy admin socket\n    \t")

	tmplp = flag.String("haproxy-template", "/usr/share/ivxv/haproxy.cfg.tmpl",
		"`path` to the HAProxy configuration template\n    \t")
)

// reload updates the HAProxy configuration and restarts HAProxy to make it use
// the new configuration.
func reload(ctx context.Context, c *conf.Technical,
	network string, service *conf.Service, until int) (code int, err error) {

	pids, code, err := readPIDFile(*pidp)
	if err != nil {
		return code, ReadHAProxyPIDsError{Err: err}
	}

	cfg, code, err := generate(c, network, service, *tmplp, *socketp)
	if err != nil {
		return code, GenerateHAProxyConfigurationError{Err: err}
	}

	if code, err = check(ctx, cfg); err != nil {
		return code, CheckHAProxyConfigurationError{Err: err}
	}

	if until >= command.Execute {
		if err = ioutil.WriteFile(*cfgp, cfg, 0644); err != nil {
			return exit.CantCreate, WriteConfigurationError{Err: err}
		}
		log.Log(ctx, UpdatedHAProxyConfiguration{Path: *cfgp})

		// It is possible here that the just created configuration file
		// gets replaced before HAProxy is restarted. This should not
		// happen during normal operation and there is little point in
		// worrying about someone doing this maliciously, because then
		// they can just replace it and restart HAProxy whenever they
		// want.

		if err = restart(ctx, pids); err != nil {
			return exit.Unavailable, RestartHAProxyError{Err: err}
		}
		log.Log(ctx, HAProxyRestarted{})
	}
	return
}

// hapid is the PID of the HAProxy instance we are controlling.
type hapid int

// enable sets p to the PID of the HAProxy instance and ensures that the
// HAProxy frontend is enabled.
func (p *hapid) enable(ctx context.Context) error {
	// XXX: We need to wait until the service manager has restarted
	// HAProxy. Replace this with something that actually ensures a new
	// instance of HAProxy is up.
	time.Sleep(500 * time.Millisecond)

	// Get the PID of the HAProxy process.
	pids, _, err := readPIDFile(*pidp)
	if err != nil {
		return EnableReadPIDError{Err: err}
	}
	if len(pids) != 1 {
		return MultipleHAProxyPIDsError{Count: len(pids)}
	}
	*(*int)(p) = pids[0]
	log.Log(ctx, HAProxyPID{PID: *p})

	// Enable the frontend.
	resp, err := socket(*socketp, "enable server election/on")
	if err != nil {
		return EnableHAProxyFrontendError{Err: err}
	}
	if len(strings.TrimSpace(resp)) > 0 {
		return EnableHAProxyFrontendCommandError{Resp: resp}
	}
	return nil
}

// check checks if the HAProxy instance is running.
func (p *hapid) check(_ context.Context) error {
	if *p == 0 {
		panic("called check on uninitialized hapid")
	}
	if err := syscall.Kill(int(*p), syscall.Signal(0)); err != nil {
		return SignalHAProxyError{PID: p, Err: err}
	}
	return nil
}

// disable ensures that the HAProxy frontend is disabled.
func (p *hapid) disable(_ context.Context) error {
	resp, err := socket(*socketp, "disable server election/on")
	if err != nil {
		return DisableHAProxyFrontendError{Err: err}
	}
	if len(strings.TrimSpace(resp)) > 0 {
		return DisableHAProxyFrontendCommandError{Resp: resp}
	}
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

	var start time.Time
	var err error
	if c.Conf.Election != nil {
		// Check election configuration time values.
		if start, err = c.Conf.Election.ServiceStartTime(); err != nil {
			return c.Error(exit.Config, StartTimeError{Err: err},
				"bad service start time:", err)
		}
	}

	var s *server.Controller
	if c.Conf.Technical != nil {
		// Create new HAProxy configuration and reload the service.
		if code, err = reload(c.Ctx, c.Conf.Technical,
			c.Network, c.Service, c.Until); err != nil {

			return c.Error(code, ReloadConfigurationError{Err: err},
				"reloading new configuration failed:", err)
		}

		// Configure a new proxy service controller.
		var p hapid
		if s, err = server.NewController(&c.Conf.Version,
			p.enable, p.check, p.disable); err != nil {

			return c.Error(exit.Config, ControllerConfError{Err: err},
				"failed to configure controller:", err)
		}
	}

	// Control HAProxy during the voting period.
	if c.Until >= command.Execute {
		if err = s.ControlAt(c.Ctx, start); err != nil {
			return c.Error(exit.Unavailable, ControlError{Err: err},
				"failed to control proxy service:", err)
		}
	}
	return exit.OK
}
