/*
The voteexp application is used for exporting votes from the storage service.
*/
package main

import (
	"archive/zip"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"ivxv.ee/command"
	"ivxv.ee/command/exit"
	"ivxv.ee/command/status"
	"ivxv.ee/errors"
	"ivxv.ee/log"
	"ivxv.ee/q11n"
	"ivxv.ee/storage"
	//ivxv:modules container
	//ivxv:modules storage
)

const usage = `voteexp exports votes from the collector's storage service for use by the
processing application.

voteexp only exports complete votes and votes which are only missing fields
listed with the optional flag.

If there were non-fatal errors, e.g. there were some partial votes in storage,
then voteexp exits with code 2.`

var (
	optionalp = flag.String("optional", "tspreg",
		// End with newline for printing default value.
		"comma-separated `list` of vote fields which are optional\n")

	qp = flag.Bool("q", false, "quiet, do not show progress")

	progress *status.Line
)

func main() {
	// Call voteexpmain in a separate function so that it can set up defers
	// and have them trigger before returning with a non-zero exit code.
	os.Exit(voteexpmain())
}

func voteexpmain() (code int) {
	c := command.New("ivxv-voteexp", usage, "output archive")
	defer func() {
		code = c.Cleanup(code)
	}()

	var qps []q11n.Protocol
	for _, q := range c.Conf.Election.Qualification {
		qps = append(qps, q.Protocol)
	}
	var opt []string
	if len(*optionalp) > 0 {
		opt = strings.Split(*optionalp, ",")
	}

	if !*qp {
		progress = status.New()
	}

	if c.Until < command.Execute {
		return exit.OK
	}
	output := c.Args[0]

	// Create the output container file.
	fp, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return c.Error(exit.CantCreate, CreateOutputError{Output: output, Err: err},
			"failed to create output archive:", err)
	}

	// Export the votes into the opened file descriptor.
	err = export(c.Ctx, c.Storage, qps, opt, fp)
	cerr := fp.Close()

	if errors.CausedBy(err, new(NonFatalError)) != nil {
		// Ad-hoc code to signal that everything got done, but
		// there were non-fatal errors.
		code = 2
		err = nil
	}
	if err != nil {
		return c.Error(exit.Unavailable, ExportVotesError{Err: err},
			"failed to export votes:", err)
	}
	if cerr != nil {
		return c.Error(exit.IOErr, CloseOutputError{Err: err},
			"failed to close output file:", err)
	}

	return
}

// export reads the votes and related metadata from the storage service and
// puts them in a ZIP archive at fp.
func export(ctx context.Context, s *storage.Client, qps []q11n.Protocol,
	optional []string, fp io.Writer) (err error) {

	// Check if the context is canceled before the expensive GetVotes
	// operation.
	select {
	case <-ctx.Done():
		return ExportCanceled{Err: ctx.Err()}
	default:
	}

	// Wrap the file writer into a zip archive writer.
	w := zip.NewWriter(fp)
	defer func() {
		if cerr := w.Close(); cerr != nil && err == nil {
			err = CloseZIPArchiveError{Err: err}
		}
	}()

	// Start exporting votes from the storage service.
	log.Log(ctx, ExportingVotes{})
	progress.Static("Exporting votes:")
	addprogress := progress.Count(0, true)
	progress.Redraw()
	defer progress.Keep()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // Stop GetVotes if we terminate early.
	c, errc := s.GetVotes(ctx, qps, optional)

	// Start a goroutine which reads and logs errors parallel to this one.
	// It either sends nil on lerrc if everything went fine, NonFatalError,
	// if non-fatal errors occurred, or a fatal error from GetVotes.
	gerrc := make(chan error, 1)
	go func() {
		var gerr log.ErrorEntry
		for err := range errc {
			gerr = GetVotesError{Err: err}
			if errors.CausedBy(err, new(storage.GetVotesFatalError)) != nil {
				break
			}

			gerr = NonFatalError{Err: gerr}
			log.Error(ctx, gerr)
			progress.Hide()
			fmt.Fprintln(os.Stderr, "error: non-fatal error:", gerr)
			progress.Show()
		}
		gerrc <- gerr
	}()
	defer func() {
		if gerr := <-gerrc; err == nil {
			err = gerr
		}
	}()

	// Add the entries to the ZIP archive in the following hierarchy:
	//
	//     votes/
	//     └ <voter id>/
	//       ├ <timestamp>.version
	//       ├ <timestamp>.<vote type>
	//       └ <timestamp>.<q11n protocol>*
	//
	var count uint64
	countlog := VoteExportProgress{Current: 0}
	const logstep = 10000 // Log progress after each logstep.
	for vote := range c {
		prefix := fmt.Sprintf("votes/%s/%s.", vote.Voter,
			strings.Replace(vote.Time.Format("20060102150405.000-0700"), ".", "", -1))

		if err = addFile(w, vote.Time, prefix+"version", []byte(vote.Version)); err != nil {
			return AddVersionError{VoteID: vote.VoteID, Prefix: prefix, Err: err}
		}
		if err = addFile(w, vote.Time, prefix+string(vote.VoteType), vote.Vote); err != nil {
			return AddVoteError{VoteID: vote.VoteID, Prefix: prefix, Err: err}
		}
		for p, b := range vote.Qualification {
			if err = addFile(w, vote.Time, prefix+string(p), b); err != nil {
				return AddQualifyingPropertyError{
					VoteID:   vote.VoteID,
					Prefix:   prefix,
					Protocol: p,
					Err:      err,
				}
			}
		}

		if count = addprogress(1); count%logstep == 0 {
			countlog.Current = count
			log.Log(ctx, countlog)
		}
	}
	log.Log(ctx, VoteCount{Count: count})
	return
}

func addFile(w *zip.Writer, mod time.Time, name string, value []byte) error {
	h := &zip.FileHeader{Name: name, Method: zip.Deflate}
	h.SetModTime(mod) // nolint: megacheck, Use deprecated API to support Go 1.9.
	f, err := w.CreateHeader(h)
	if err != nil {
		return AddFileCreateError{Name: name, Err: err}
	}
	if _, err := f.Write(value); err != nil {
		return AddFileWriteError{Name: name, Err: err}
	}
	return nil
}
