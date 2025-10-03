package server

import (
	"context"
	"time"

	"ivxv.ee/common/collector/conf/version"
	"ivxv.ee/common/collector/log"
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
	*Controller, error) {

	c := &Controller{
		startfn: startfn,
		checkfn: checkfn,
		stopfn:  stopfn,
	}

	// Create a new status reporter for this controller.
	var err error
	if c.status, err = newStatus(v); err != nil {
		return nil, NewControllerStatusError{Err: err,
			Description: _SERVER_STAT}
	}
	return c, nil
}

// Control calls startfn and calls stopfn when ctx is cancelled.
func (c *Controller) Control(ctx context.Context) error {
	// Start the service.
	log.Log(ctx, StartingControlledService{
		Description: _SERVER_SYSD_START,
	})
	if err := c.startfn(ctx); err != nil {
		return ControllerStartError{Err: err,
			Description: _SERVER_SYSD_START_FAIL}
	}
	log.Log(ctx, ControlledServiceStarted{Description: _SERVER_SYSD_STARTED})

	if err := c.status.serving(); err != nil {
		return ControlStatusServingError{Err: err,
			Description: _SERVER_SYSD_SERVE}
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
			if err := c.checkfn(ctx); err != nil {
				log.Error(ctx, ControllerCheckError{Err: err,
					Description: _SERVER_SYSD_CHECK})
				c.failed++

				// If we have restarted three times without a
				// check succeeding, then stop trying.
				if c.failed >= 3 {
					return ControllerAbortError{Description: _SERVER_SYSD_ATTEMPT}
				}

				return c.Control(ctx)
			}
			c.failed = 0
			sleep.Reset(d)
		}
	}

	// If updating the status returns an error, then only log it: do not
	// skip the stop function.
	if err := c.status.stopping(); err != nil {
		log.Error(ctx, ControlStatusStoppingError{Err: err,
			Description: "Initiate systemd service stopping failed"})
	}

	// Stop the service.
	log.Log(ctx, StoppingControlledService{Description: _SERVER_SYSD_STOP})
	if err := c.stopfn(ctx); err != nil {
		return ControllerStopError{Err: err, Description: _SERVER_SYSD_STOP_FAIL}
	}
	log.Log(ctx, ControlledServiceStopped{Description: _SERVER_SYSD_STOPPED})
	return nil
}

// ControlAt waits until start and then calls Control.
func (c *Controller) ControlAt(ctx context.Context, start time.Time) error {
	if err := c.status.waiting(); err != nil {
		return ControlAtStatusWaitingError{Err: err, Description: _SERVER_SYSD_WAIT}
	}
	return waitStart(ctx, start, c.Control)
}
