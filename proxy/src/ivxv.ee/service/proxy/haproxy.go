package main

import (
	"bytes"
	"context"
	"html/template"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
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
	Service  *conf.Service
	Backends []*backend
}

type backend struct {
	Name    string
	Servers []*conf.Service
}

// generate creates a HAProxy configuration based on the technical configuration
// and a template.
func generate(c *conf.Technical, network string, service *conf.Service, tmpl string) (
	cfg []byte, code int, err error) {

	t, err := template.New(filepath.Base(tmpl)).Funcs(template.FuncMap{
		"port": func(addr string) (port string, err error) {
			_, port, err = net.SplitHostPort(addr)
			return
		},
		"replace": strings.Replace,
	}).ParseFiles(tmpl)
	if err != nil {
		return nil, exit.DataErr, ParseTemplateError{Path: tmpl, Err: err}
	}

	d := data{
		Debug:   c.Debug,
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

// readPIDFile opens pidfile and reads the PID of the master process.
func readPIDFile(pidfile string) (pid int, code int, err error) {
	pidb, err := ioutil.ReadFile(pidfile)
	if err != nil {
		return 0, exit.NoInput, ReadPIDFileError{Path: pidfile, Err: err}
	}
	eol := bytes.IndexByte(pidb, '\n')
	if eol < 0 {
		eol = len(pidb)
	}
	pid, err = strconv.Atoi(string(pidb[:eol]))
	if err != nil {
		return 0, exit.DataErr, ParsePIDFilePIDError{PIDFile: string(pidb), Err: err}
	}
	return pid, exit.OK, nil
}

// Since we do not have enough permissions to reload HAProxy we have to
// improvise a bit: we expect to be the same user that HAProxy child processes
// are running under and send USR1 to them. This will cause them to gracefully
// close and HAProxy to stop. The service manager will pick up on this and
// restart HAProxy, now with the new configuration.
func restart(ctx context.Context, proc string, pid int) error {
	cpids, err := childPIDs(proc, pid)
	if err != nil {
		return ChildPIDsError{ParentPID: pid, Err: err}
	}
	if len(cpids) == 0 {
		return NoChildPIDsError{ParentPID: pid}
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start a goroutine per child PID. Create a buffered channel to hold
	// all the results without blocking.
	c := make(chan error, len(cpids))
	for _, pid := range cpids {
		go func(pid int) { c <- stopPID(ctx, pid) }(pid)
	}

	// Read up to as many results as goroutines were started.
	for range cpids {
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

	// Signal the PID with 0 until it no longer succeeds. There is a chance
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
			if err := syscall.Kill(pid, syscall.Signal(0)); err != nil {
				log.Log(ctx, HAProxyStopped{PID: pid})
				return nil
			}
			sleep.Reset(200 * time.Millisecond)
		}
	}
}

// childPIDs scans through proc looking for processes whose parent PID is pid.
func childPIDs(proc string, pid int) ([]int, error) {
	ppid := regexp.MustCompile("\nPPid:\t" + strconv.Itoa(pid) + "\n")

	dir, err := os.Open(proc)
	if err != nil {
		return nil, OpenProcError{Proc: proc, Err: err}
	}
	defer dir.Close()

	names, err := dir.Readdirnames(-1)
	if err != nil {
		return nil, ReadProcError{Proc: proc, Err: err}
	}
	var cpids []int
	for _, name := range names {
		cpid, err := strconv.Atoi(name)
		if err != nil {
			continue // Skip non-integer directory names.
		}
		status, err := ioutil.ReadFile(filepath.Join(proc, name, "status"))
		if err != nil {
			// Skip processes where we cannot read status (probably
			// exited after Readdirnames).
			continue
		}
		if ppid.Match(status) {
			cpids = append(cpids, cpid)
		}
	}
	return cpids, nil
}

// checkPID checks proc if the process PID still exists.
func checkPID(proc string, pid int) error {
	if _, err := os.Stat(filepath.Join(proc, strconv.Itoa(pid), "status")); err != nil {
		return StatProcStatusError{PID: pid, Err: err}
	}
	return nil
}
