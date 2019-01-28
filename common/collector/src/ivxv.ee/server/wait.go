package server

import (
	"context"
	"time"

	"ivxv.ee/log"
)

// wait waits until t, then calls f. wait does not simply start a new timer,
// but periodically checks the time to see if t is close enough and then starts
// a timer. This approach makes it obey wall clock changes made after wait has
// been called, but still precise when close to t.
func wait(ctx context.Context, t time.Time, f func(context.Context) error) error {
	// Check wall clock every minute until we are at most one minute away
	// from t.
	ticker := time.NewTicker(time.Minute)
	for time.Until(t) > time.Minute {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return ctx.Err()
		case <-ticker.C: // Check time again.
		}
	}
	ticker.Stop()

	// Do not follow wall clock changes anymore and just wait until t.
	timer := time.NewTimer(time.Until(t))
	select {
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
		return ctx.Err()
	case <-timer.C:
		return f(ctx)
	}
}

// waitStart is like wait, but performs additional logging and converts
// cancellation errors to nil. Meant for delayed start functions.
func waitStart(ctx context.Context, start time.Time, f func(context.Context) error) error {
	log.Log(ctx, WaitingForStart{Start: start})
	switch err := wait(ctx, start, f); err {
	case context.Canceled, context.DeadlineExceeded:
		// We know that waiting was canceled because we only get plain
		// context errors from wait: all context cancellation errors
		// from f are expected to be wrapped.
		log.Log(ctx, WaitingCanceled{})
		return nil
	default:
		return err
	}
}
