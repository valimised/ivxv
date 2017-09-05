package server

import (
	"context"
	"time"

	"ivxv.ee/conf/version"
	"ivxv.ee/log"
)

// Controller is a controller which controls the operation of an external service.
type Controller struct {
	status  *status
	startfn func(context.Context) error
	checkfn func(context.Context) error
	stopfn  func(context.Context) error
	failed  int
}

// NewController creates a new external service controller with the provided
// control functions. startfn and stopfn are used to start and stop the
// controlled service and checkfn is called regularly in-between to check if
// the service is still up, calling startfn again if not.
//
// The configuration version is necessary for reporting status.
func NewController(v *version.V, startfn, checkfn, stopfn func(context.Context) error) (
	c *Controller, err error) {

	c = &Controller{
		startfn: startfn,
		checkfn: checkfn,
		stopfn:  stopfn,
	}

	// Create a new status reporter for this controller.
	if c.status, err = newStatus(v); err != nil {
		return nil, NewControllerStatusError{Err: err}
	}
	return
}

// Control calls startfn and calls stopfn when ctx is cancelled.
func (c *Controller) Control(ctx context.Context) (err error) {
	if err = c.status.ready(); err != nil {
		return ControlStatusReadyError{Err: err}
	}

	// Start the service.
	log.Log(ctx, StartingControlledService{})
	if err = c.startfn(ctx); err != nil {
		return ControllerStartError{Err: err}
	}
	log.Log(ctx, ControlledServiceStarted{})

	if err = c.status.serving(); err != nil {
		return ControlStatusServingError{Err: err}
	}

	// Check the service every second until ctx is cancelled. If the poll
	// fails, then recurse and attempt to start again.
	const d = time.Second
	sleep := time.NewTimer(d)
poll:
	for {
		select {
		case <-ctx.Done():
			break poll
		case <-sleep.C:
			if err = c.checkfn(ctx); err != nil {
				log.Error(ctx, ControllerCheckError{Err: err})
				c.failed++

				// If we have restarted three times without a
				// check succeeding, then stop trying.
				if c.failed >= 3 {
					return ControllerAbortError{}
				}

				return c.Control(ctx)
			}
			c.failed = 0
			sleep.Reset(d)
		}
	}

	// If updating the status returns an error, then only log it. Returning
	// a non-nil error will cause the service manager to restart the
	// service, which we do not want if we did not have any prior errors.
	if serr := c.status.stopping(); serr != nil {
		log.Error(ctx, ControlStatusStoppingError{Err: err})
	}

	// Stop the service.
	log.Log(ctx, StoppingControlledService{})
	if err = c.stopfn(ctx); err != nil {
		return ControllerStopError{Err: err}
	}
	log.Log(ctx, ControlledServiceStopped{})
	return nil
}

// ControlAt waits until start and then calls Control.
func (c *Controller) ControlAt(ctx context.Context, start time.Time) error { // nolint: dupl
	// Ignore duplicate code from S.ServeAt: we want to keep them separate
	// for a clearer API.
	if err := c.status.ready(); err != nil {
		return ControlAtStatusReadyError{Err: err}
	}
	return waitStart(ctx, start, c.Control)
}
