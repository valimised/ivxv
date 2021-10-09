package main

import (
	"bufio"
	"context"
	"io"
	"log/syslog"
	"regexp"

	"ivxv.ee/log"
)

// Prefix used by github.com/coreos/pkg/capnslog.PrettyFormatter.
var re = regexp.MustCompile(`^\d\d\d\d-\d\d-\d\d \d\d:\d\d:\d\d\.\d{6} ([CEWNIDT]) \| (.+)$`)

type logger struct {
	*syslog.Writer
}

func newLogger() (logger, error) {
	// Same facility as used in ivxv.ee/log.
	w, err := syslog.New(syslog.LOG_LOCAL0, "etcd")
	if err != nil {
		return logger{nil}, EtcdSyslogError{Err: err}
	}
	return logger{w}, nil
}

func (l logger) log(ctx context.Context, r io.ReadCloser) {
	defer r.Close()
	defer func() {
		if err := l.Close(); err != nil {
			log.Error(ctx, EtcdSyslogCloseError{Err: err})
		}
	}()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			log.Error(ctx, EtcdUnexpectedLogError{Line: line})
			continue
		}

		level := matches[1]
		msg := matches[2]

		var err error
		switch level {
		case "C":
			err = l.Crit(msg)
		case "E":
			err = l.Err(msg)
		case "W":
			err = l.Warning(msg)
		case "N":
			err = l.Notice(msg)
		case "I":
			err = l.Info(msg)
		case "D", "T":
			err = l.Debug(msg)
		default:
			panic("regexp submatch unexpected value: " + level)
		}
		if err != nil {
			log.Error(ctx, EtcdLogError{
				Level:   level,
				Message: msg,
				Err:     log.Alert(err),
			})
		}
	}
	// We are reading from a PipeReader and are not using CloseWithError,
	// so only EOF is returned: ignore scanner.Err().
}
