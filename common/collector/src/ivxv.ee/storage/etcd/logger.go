package etcd

import (
	"context"
	"errors"
	"fmt"

	"ivxv.ee/log"
)

// logger implements the clientv3.Logger interface.
type logger struct {
	ctx context.Context
}

func newLogger(ctx context.Context) *logger {
	return &logger{ctx}
}

func (l *logger) info(msg string) {
	log.Log(l.ctx, ClientInfo{Message: msg})
}

func (l *logger) warning(msg string) {
	log.Log(l.ctx, ClientWarning{Message: msg})
}

func (l *logger) error(msg string) {
	log.Error(l.ctx, ClientError{Err: errors.New(msg)})
}

func (l *logger) fatal(msg string) {
	log.Error(l.ctx, ClientFatal{Err: log.Alert(errors.New(msg))})
}

func (l *logger) Info(a ...interface{})                    { l.info(fmt.Sprint(a...)) }
func (l *logger) Infoln(a ...interface{})                  { l.info(fmt.Sprintln(a...)) }
func (l *logger) Infof(format string, a ...interface{})    { l.info(fmt.Sprintf(format, a...)) }
func (l *logger) Warning(a ...interface{})                 { l.warning(fmt.Sprint(a...)) }
func (l *logger) Warningln(a ...interface{})               { l.warning(fmt.Sprintln(a...)) }
func (l *logger) Warningf(format string, a ...interface{}) { l.warning(fmt.Sprintf(format, a...)) }
func (l *logger) Error(a ...interface{})                   { l.error(fmt.Sprint(a...)) }
func (l *logger) Errorln(a ...interface{})                 { l.error(fmt.Sprintln(a...)) }
func (l *logger) Errorf(format string, a ...interface{})   { l.error(fmt.Sprintf(format, a...)) }
func (l *logger) Fatal(a ...interface{})                   { l.fatal(fmt.Sprint(a...)) }
func (l *logger) Fatalf(format string, a ...interface{})   { l.fatal(fmt.Sprintf(format, a...)) }
func (l *logger) Fatalln(a ...interface{})                 { l.fatal(fmt.Sprintln(a...)) }
func (l *logger) V(v int) bool                             { return true }
