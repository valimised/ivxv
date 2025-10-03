/*
The choiceimp application is used for loading choice lists into the storage
service.
*/
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"ivxv.ee/common/collector/command"
	"ivxv.ee/common/collector/command/exit"
	"ivxv.ee/common/collector/command/status"
	"ivxv.ee/common/collector/conf"
	"ivxv.ee/common/collector/conf/version"
	"ivxv.ee/common/collector/errors"
	"ivxv.ee/common/collector/log"
	"ivxv.ee/common/collector/storage"
	//ivxv:modules common/collector/container
	//ivxv:modules common/collector/storage
)

const usage = `choiceimp loads choice lists into the collector's storage service for use by
other services.

The choice list container must have an extension corresponding to the container
type it is, e.g., choicelist.bdoc.`

var (
	qp = flag.Bool("q", false, "quiet, do not show progress")

	progress *status.Line
)

func main() {
	// Call choiceimpmain in a separate function so that it can set up
	// defers and have them trigger before returning with a non-zero exit
	// code.
	os.Exit(choiceimpmain())
}

func choiceimpmain() (code int) {
	c := command.New("ivxv-choiceimp", usage, "choice list container")
	defer func() {
		code = c.Cleanup(code)
	}()

	if !*qp {
		progress = status.New()
	}

	// Only check the version if it was the specific check requested.
	if c.Until == command.CheckVersion {
		version1, err := c.Storage.GetChoicesVersion(c.Ctx)
		if err != nil {
			if errors.CausedBy(err, new(storage.NotExistError)) == nil {
				return c.Error(exit.Unavailable, CheckVersionError{Err: err,
					Description: _CHOICES_VERSION_FROM_DB},
					"failed to check imported list version:", err)
			}
		} else {
			fmt.Println(version1)
		}
	}

	if c.Until < command.CheckInput {
		return exit.OK
	}
	path := c.Args[0]

	// Open the choice list container file.
	cnt, err := c.Conf.Container.OpenFile(path)
	if err != nil {
		code = exit.DataErr
		if perr := errors.CausedBy(err, new(os.PathError)); perr != nil {
			if os.IsNotExist(perr) {
				code = exit.NoInput
			}
		}
		return c.Error(code, OpenContainerError{Container: path, Err: err,
			Description: _CHOICES_LIST_BDOC_OPEN},
			"failed to open choice list container:", err)
	}
	defer cnt.Close()

	// Ensure that the container is signed and log the signatures.
	signatures := cnt.Signatures()
	if len(signatures) == 0 {
		return c.Error(exit.DataErr, UnsignedContainerError{Container: path,
			Description: _CHOICES_LIST_BDOC_NO_SIG},
			"unsigned choice list container")
	}
	for _, s := range signatures {
		log.Log(c.Ctx, ContainerSignature{Signer: s.Signer, SigningTime: s.SigningTime,
			Description: _CHOICES_LIST_BDOC_SIG_INFO})
	}

	// Get the version string of the container.
	version1, err := version.Container(cnt)
	if err != nil {
		return c.Error(exit.DataErr, ContainerVersionError{Container: path, Err: err,
			Description: _CHOICES_LIST_BDOC_SIG_TO_JSON},
			"failed to format container version string:", err)
	}

	// Check that the container only has a single file.
	data := cnt.Data()
	if len(data) != 1 {
		return c.Error(exit.DataErr, DataCountError{Count: len(data),
			Description: _CHOICES_LIST_BDOC_HAS_MANY_FILES},
			"choice list container has", len(data), "files, expected 1")
	}

	// Process the choice list. We do not know the key, so do a single cycle loop.
	for key, list := range data {
		log.Log(c.Ctx, ProcessingList{List: key,
			Description: _CHOICES_LIST_READ})
		if err := choiceimp(c.Ctx, c.Until, c.Conf, c.Storage, version1, list); err != nil {
			return c.Error(exit.Unavailable, ImportChoicesError{Err: err,
				Description: _CHOICES_LIST_TO_DB_UPLOAD},
				"failed to import choice list", key+":", err)
		}
	}
	return exit.OK
}

type choicelist struct {
	Election string
	Choices  map[string]json.RawMessage
}

// choiceimp parses the list of choices and uploads them to the storage
// service.
func choiceimp(ctx context.Context, until int, c *conf.C, s *storage.Client,
	version string, list []byte) error {

	var l choicelist
	if err := json.Unmarshal(list, &l); err != nil {
		return JSONUnmarshalError{Err: err, Description: _CHOICES_LIST_INVALID_JSON}
	}

	// Ensure that the election identifier matches the configured one.
	if l.Election != c.Election.Identifier {
		return ElectionIDMismatchError{Conf: c.Election.Identifier, List: l.Election,
			Description: _CHOICES_LIST_ID_MISMATCH}
	}

	if until >= command.Execute {
		// Convert map[string]json.RawMessage to map[string][]byte.
		choices := make(map[string][]byte)
		for id, list := range l.Choices {
			choices[id] = list
		}

		log.Log(ctx, ImportingChoices{Count: len(choices), Description: _CHOICES_LIST_PREPARE_UPLOAD_TO_DB})
		progress.Static(fmt.Sprintf("Importing %d choices:", len(choices)))
		addprogress := progress.Percent(uint64(len(choices)), true)
		progress.Redraw()
		defer progress.Keep()

		if err := s.PutChoices(ctx, version, choices, addprogress); err != nil {
			return PutChoicesError{Err: err, Description: _CHOICES_LIST_UPLOAD_TO_DB_FAIL}
		}
	}
	return nil
}
