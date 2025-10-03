package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log/syslog"
	"net/url"

	"ivxv.ee/common/collector/log"
)

type etcdLog struct {
	Level     string `json:"level"`
	Timestamp string `json:"ts"`
	Message   string `json:"msg"`
	// other fields are not important
}

const (
	Error  = "error"
	Warn   = "warn"
	Notice = "notice"
	Info   = "info"
	Debug  = "debug"
)

type logger struct {
	*syslog.Writer
}

func newLogger() (logger, error) {
	// Same facility as used in ivxv.ee/common/collector/log.
	w, err := syslog.New(syslog.LOG_LOCAL0, "etcd")
	if err != nil {
		return logger{nil}, EtcdSyslogError{Err: err, Description: _STORAGE_ETCD_SYSLOG}
	}
	return logger{w}, nil
}

func (l logger) log(ctx context.Context, r io.ReadCloser) {
	defer r.Close()
	defer func() {
		if err := l.Close(); err != nil {
			log.Error(ctx, EtcdSyslogCloseError{Err: err,
				Description: _STORAGE_ETCD_SYSLOG_CLOSE})
		}
	}()

	scanner := bufio.NewScanner(r)

	// All etcd logs should be valid JSON, however there are exceptions
	// which were found out during stress testing.
	//
	// For example etcd could be killed by systemd oomd.service, in that
	// case `line = Killed`. But it doesn't mean that ivxv-storage should
	// panic due to invalid log.
	// ivxv-storage will catch etcd's SIGKILL or any other signal and then
	// will try to restart a node.
	//
	// Another example is etcd warning messages that aren't valid JSON,
	// as an example:
	// `line = Server.processUnaryRPC failed to write connection error: desc = "transport is closing"`
	// This message is not harmful for etcd operation and only tells that
	// etcd cannot write log messages to the client right now, source:
	// https://github.com/etcd-io/etcd/issues/12895
	for scanner.Scan() {
		// Before: %7B%22level%22%3A%22info%22%2C%22ts%22%3A%
		line := scanner.Text()

		// After: "{"level": "info", "ts":
		unescaped, err := url.QueryUnescape(line)
		if err != nil {
			log.Error(ctx, QueryUnescapeLogError{Line: line,
				Description: _STORAGE_ETCD_ESCAPE})
			continue
		}

		var elog etcdLog
		err = json.Unmarshal([]byte(unescaped), &elog)
		if err != nil {
			log.Error(ctx, EtcdUnexpectedLogError{Line: line,
				Description: _STORAGE_ETCD_FORMAT})
			continue
		}

		level := elog.Level
		msg := elog.Message

		switch level {
		case Error:
			err = l.Err(msg)
		case Warn:
			err = l.Warning(msg)
		case Notice:
			err = l.Notice(msg)
		case Info:
			err = l.Info(msg)
		case Debug:
			err = l.Debug(msg)
		}
		if err != nil {
			log.Error(ctx, EtcdLogError{
				Level:       level,
				Message:     msg,
				Err:         log.Alert(err),
				Description: _STORAGE_ETCD_UNKNOWN_LVL,
			})
		}
	}
	// We are reading from a PipeReader and are not using CloseWithError,
	// so only EOF is returned: ignore scanner.Err().
}
