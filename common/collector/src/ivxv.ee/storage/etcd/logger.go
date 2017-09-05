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

func (l *logger) print(msg string) {
	log.Log(l.ctx, ClientLog{Message: msg})
}

func (l *logger) fatal(msg string) {
	log.Error(l.ctx, ClientError{Err: log.Alert(errors.New(msg))})
}

func (l *logger) Print(a ...interface{})                 { l.print(fmt.Sprint(a...)) }
func (l *logger) Printf(format string, a ...interface{}) { l.print(fmt.Sprintf(format, a...)) }
func (l *logger) Println(a ...interface{})               { l.print(fmt.Sprintln(a...)) }
func (l *logger) Fatal(a ...interface{})                 { l.fatal(fmt.Sprint(a...)) }
func (l *logger) Fatalf(format string, a ...interface{}) { l.fatal(fmt.Sprintf(format, a...)) }
func (l *logger) Fatalln(a ...interface{})               { l.fatal(fmt.Sprintln(a...)) }
