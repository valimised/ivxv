package main

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"io/ioutil"
	"net"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"

	"ivxv.ee/command/exit"
	"ivxv.ee/conf"
	"ivxv.ee/log"
)

type data struct {
	Debug    bool
	Socket   string
	Service  *conf.Service
	Backends []*backend
}

type backend struct {
	Name    string
	Servers []*conf.Service
}

// generate creates a HAProxy configuration based on the technical configuration
// and a template.
func generate(c *conf.Technical, network string, service *conf.Service, tmpl, socket string) (
	cfg []byte, code int, err error) {

	t, err := template.New(filepath.Base(tmpl)).Funcs(template.FuncMap{
		"replace": strings.Replace,
	}).ParseFiles(tmpl)
	if err != nil {
		return nil, exit.DataErr, ParseTemplateError{Path: tmpl, Err: err}
	}

	d := data{
		Debug:   c.Debug,
		Socket:  socket,
		Service: service,
	}
	services := c.Services(network)
	rt := reflect.TypeOf(services).Elem()
	rv := reflect.ValueOf(services).Elem()
	for i := 0; i < rt.NumField(); i++ {
		switch rt.Field(i).Name {
		case "Proxy", "Storage":
			continue // Do not proxy to other proxies or storage.
		}
		d.Backends = append(d.Backends, &backend{
			Name:    strings.ToLower(rt.Field(i).Name),
			Servers: rv.Field(i).Interface().([]*conf.Service),
		})
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, d); err != nil {
		return nil, exit.DataErr, ExecuteTemplateError{Err: err}
	}

	return buf.Bytes(), exit.OK, nil
}

// check calls the HAProxy binary to verify that cfg is valid.
func check(ctx context.Context, cfg []byte) (code int, err error) {
	// nolint: gas, this is a false-positive from gas, which has not
	// whitelisted the ctx argument for exec.CommandContext.
	cmd := exec.CommandContext(ctx, "/usr/sbin/haproxy", "-c", "--", "/dev/stdin")
	cmd.Stdin = bytes.NewReader(cfg)
	if out, err := cmd.CombinedOutput(); err != nil {
		return exit.DataErr, CheckConfigurationError{
			Configuration: string(cfg),
			Output:        string(out),
			Err:           err,
		}
	}
	return
}

// readPIDFile opens pidfile and reads a PID per line.
func readPIDFile(pidfile string) (pids []int, code int, err error) {
	pidsb, err := ioutil.ReadFile(pidfile)
	if err != nil {
		return nil, exit.NoInput, ReadPIDFileError{Path: pidfile, Err: err}
	}
	split := bytes.Split(pidsb, []byte{'\n'})
	for _, pidb := range split {
		if len(pidb) == 0 {
			continue // Skip empty lines.
		}
		pid, err := strconv.Atoi(string(pidb))
		if err != nil {
			return nil, exit.DataErr,
				ParsePIDFilePIDError{PID: string(pidb), Err: err}
		}
		pids = append(pids, pid)
	}
	return
}

// Since we do not have enough permissions to reload HAProxy we have to
// improvise a bit: we expect to be the same user that HAProxy child processes
// are running under and send USR1 to them. This will cause them to gracefully
// close and HAProxy to stop. The service manager will pick up on this and
// restart HAProxy, now with the new configuration.
func restart(ctx context.Context, pids []int) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start a goroutine per PID. Create a buffered channel to hold all the
	// results without blocking.
	c := make(chan error, len(pids))
	for _, pid := range pids {
		go func(pid int) { c <- stopPID(ctx, pid) }(pid)
	}

	// Read up to as many results as goroutines were started.
	for range pids {
		if err := <-c; err != nil {
			return StopPIDError{Err: err}
		}
	}
	return nil
}

func stopPID(ctx context.Context, pid int) error {
	if err := syscall.Kill(pid, syscall.SIGUSR1); err != nil {
		return SignalPIDError{PID: pid, Err: err}
	}

	// Call check on the PID until it no longer succeeds. There is a chance
	// that between two signals, the old process stops and a new process
	// starts under the same user with the same PID, but this is negligible
	// enough that we ignore it.
	log.Log(ctx, WaitHAProxyStop{PID: pid})
	sleep := time.NewTimer(0)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-sleep.C:
			if err := (*hapid)(&pid).check(context.Background()); err != nil {
				log.Log(ctx, HAProxyStopped{PID: pid})
				return nil
			}
			sleep.Reset(200 * time.Millisecond)
		}
	}
}

// socket writes command to the socket at path and returns the response.
func socket(path string, command string) (resp string, err error) {
	c, err := net.Dial("unix", path)
	if err != nil {
		return "", DialHAProxySocketError{Path: path, Err: err}
	}
	// nolint: errcheck, ignore close failures, because we either have a
	// higher-priority error or already got a successful response.
	defer c.Close()

	if _, err = io.WriteString(c, command+"\n"); err != nil {
		return "", WriteHAProxySocketError{Err: err}
	}

	rb, err := ioutil.ReadAll(c)
	if err != nil {
		return "", ReadHAProxySocketError{Err: err}
	}
	return string(rb), nil
}
