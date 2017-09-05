package main

import (
	"bufio"
	"context"
	"errors"
	"io"
	"regexp"

	"ivxv.ee/log"
)

// Prefix used by github.com/coreos/pkg/capnslog.PrettyFormatter.
var re = regexp.MustCompile(`^\d\d\d\d-\d\d-\d\d \d\d:\d\d:\d\d\.\d{6} ([CEWNIDT]) \| (.+)$`)

func logger(ctx context.Context, r io.ReadCloser) {
	defer r.Close() // nolint: errcheck, PipeReader.Close always returns nil.

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			log.Error(ctx, EtcdUnexpectedLog{Line: line})
			continue
		}

		level := matches[1]
		msg := matches[2]

		switch level {
		case "C", "E", "W":
			// All etcd errors trigger an alert.
			log.Error(ctx, EtcdError{Err: log.Alert(errors.New(msg))})
		case "N", "I":
			log.Log(ctx, EtcdLog{Message: msg})
		case "D", "T":
			log.Debug(ctx, EtcdDebug{Message: msg})
		default:
			panic("regexp submatch unexpected value: " + level)
		}
	}
	// We are reading from a PipeReader and are not using CloseWithError,
	// so only EOF is returned: ignore scanner.Err().
}
