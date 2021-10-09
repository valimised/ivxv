/*
The storageidx application rebuilds indexes in the storage service.
*/
package main

import (
	"context"
	"flag"
	"os"
	"time"

	"ivxv.ee/command"
	"ivxv.ee/command/exit"
	"ivxv.ee/command/status"
	"ivxv.ee/errors"
	"ivxv.ee/log"
	"ivxv.ee/q11n"
	"ivxv.ee/storage"
	//ivxv:modules container
	//ivxv:modules q11n
	//ivxv:modules storage
)

const usage = `storageidx rebuilds indexes in the storage service.

An index consists of a set of keys in the storage service which do not contain
primary data, but build upon existing values, e.g., by constructing secondary
lookup tables or aggregating data. Indices help optimize operations which
would be too slow when operating on primary data only.

It is usually not necessary to manually rebuild an index in the storage
service, but use of this application might help in cases where index corruption
is detected, e.g., when new data was stored but a following index update
failed.

Since the storage service does not allow deletion of any keys, the application
does not clear the indexes before starting the rebuild, but simply overwrites
all values.`

var (
	qp = flag.Bool("q", false, "quiet, do not show progress")

	progress *status.Line
)

func main() {
	// Call storageidxmain in a separate function so that it can set up
	// defers and have them trigger before returning with a non-zero exit
	// code.
	os.Exit(storageidxmain())
}

func storageidxmain() (code int) {
	c := command.New("ivxv-storageidx", usage)
	defer func() {
		code = c.Cleanup(code)
	}()

	if !*qp {
		progress = status.New()
	}

	start, err := c.Conf.Election.ElectionStartTime()
	if err != nil {
		return c.Error(exit.Config, ElectionStartTimeError{Err: err},
			"failed to parse election start time:", err)
	}

	var qps []q11n.Protocol
	for _, q := range c.Conf.Election.Qualification {
		qps = append(qps, q.Protocol)
	}

	if c.Until < command.Execute {
		return exit.OK
	}

	if err := rebuildVotedStats(c.Ctx, start, qps, c.Storage); err != nil {
		return c.Error(exit.Unavailable, RebuildVotedStatsError{Err: err},
			"failed to rebuild successful voter index:", err)
	}

	return
}

func rebuildVotedStats(ctx context.Context, start time.Time, qps []q11n.Protocol,
	s *storage.Client) error {

	log.Log(ctx, RebuildingVotedStats{})
	progress.Static("Rebuilding successful voter statistics index:")
	addvoter := progress.Count(0, false)
	progress.Static("voters with")
	addprogress := progress.Count(0, true)
	progress.Static("votes")
	progress.Redraw()
	defer progress.Keep()

	// To rebuild the successful voter statistics index retrieve all
	// complete votes from storage and re-mark them as voted. The SetVoted
	// method will compare the vote with the existing index value for that
	// voter and update if necessary: the order of calls does not matter.
	c, errc := s.GetVotes(ctx, qps, nil)
	voters := make(map[string]struct{})
	for {
		select {
		case vote, ok := <-c:
			if !ok {
				// Replace closed channel with nil so that it
				// is not selected anymore. Keep reading from
				// error channel.
				c = nil
				continue
			}

			ctime := vote.Time
			if q11nTime, err := q11n.CanonicalTime(vote.Qualification); err != nil {
				return err
			} else if !q11nTime.IsZero() {
				ctime = q11nTime
			}

			testVote := ctime.Before(start)
			if err := s.SetVoted(ctx, vote.VoteID, ctime, testVote); err != nil {
				return err
			}
			if _, ok := voters[vote.Voter]; !ok {
				voters[vote.Voter] = struct{}{}
				addvoter(1)
			}
			addprogress(1)
		case err, ok := <-errc:
			if !ok {
				// Error channel was closed: we are done.
				return nil
			}
			if errors.CausedBy(err, (new(storage.GetVotesFatalError))) == nil {
				// Ignore any non-fatal errors. They are not
				// the problem of this application.
				continue
			}
			return err
		}
	}
}
