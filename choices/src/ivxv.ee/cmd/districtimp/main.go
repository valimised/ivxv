/*
The districtimp application is used for loading district lists into the storage
service.
*/
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"ivxv.ee/command"
	"ivxv.ee/command/exit"
	"ivxv.ee/command/status"
	"ivxv.ee/conf"
	"ivxv.ee/conf/version"
	"ivxv.ee/errors"
	"ivxv.ee/log"
	"ivxv.ee/storage"
	//ivxv:modules container
	//ivxv:modules storage
)

const usage = `districtimp loads district lists into the collector's storage service for use
by other services.

The district list container must have an extension corresponding to the
container type it is, e.g., choicelist.bdoc.`

var (
	qp = flag.Bool("q", false, "quiet, do not show progress")

	progress *status.Line
)

func main() {
	// Call districtimpmain in a separate function so that it can set up
	// defers and have them trigger before returning with a non-zero exit
	// code.
	os.Exit(districtimpmain())
}

func districtimpmain() (code int) {
	c := command.New("ivxv-districtimp", usage, "district list container")
	defer func() {
		code = c.Cleanup(code)
	}()

	if !*qp {
		progress = status.New()
	}

	// Only check the version if it was the specific check requested.
	if c.Until == command.CheckVersion {
		version, err := c.Storage.GetDistrictsVersion(c.Ctx)
		if err != nil {
			if errors.CausedBy(err, new(storage.NotExistError)) == nil {
				return c.Error(exit.Unavailable, CheckVersionError{Err: err},
					"failed to check imported list version:", err)
			}
		} else {
			fmt.Println(version)
		}
	}

	if c.Until < command.CheckInput {
		return exit.OK
	}
	path := c.Args[0]

	// Open the district list container file.
	cnt, err := c.Conf.Container.OpenFile(path)
	if err != nil {
		code = exit.DataErr
		if perr := errors.CausedBy(err, new(os.PathError)); perr != nil {
			if os.IsNotExist(perr) {
				code = exit.NoInput
			}
		}
		return c.Error(code, OpenContainerError{Container: path, Err: err},
			"failed to open district list container:", err)
	}
	defer cnt.Close()

	// Ensure that the container is signed and log the signatures.
	signatures := cnt.Signatures()
	if len(signatures) == 0 {
		return c.Error(exit.DataErr, UnsignedContainerError{Container: path},
			"unsigned district list container")
	}
	for _, s := range signatures {
		log.Log(c.Ctx, ContainerSignature{Signer: s.Signer, SigningTime: s.SigningTime})
	}

	// Get the version string of the container.
	version, err := version.Container(cnt)
	if err != nil {
		return c.Error(exit.DataErr, ContainerVersionError{Container: path, Err: err},
			"failed to format container version string:", err)
	}

	// Check that the container only has a single file.
	data := cnt.Data()
	if len(data) != 1 {
		return c.Error(exit.DataErr, DataCountError{Count: len(data)},
			"district list container has", len(data), "files, expected 1")
	}

	// Process the district list. We do not know the key, so do a single cycle loop.
	for key, list := range data {
		log.Log(c.Ctx, ProcessingList{List: key})
		if err := districtimp(c.Ctx, c.Until, c.Conf, c.Storage, version, list); err != nil {
			return c.Error(exit.Unavailable, ImportDistrictsError{Err: err},
				"failed to import district list", key+":", err)
		}
	}
	return exit.OK
}

type districtlist struct {
	Election  string
	Districts map[string]district
	Counties  json.RawMessage
}

type district struct {
	Name   string
	Parish []string
}

// districtimp parses the districts list and uploads data from it to the
// storage service.
func districtimp(ctx context.Context, until int, c *conf.C, s *storage.Client,
	version string, list []byte) error {

	var l districtlist
	if err := json.Unmarshal(list, &l); err != nil {
		return JSONUnmarshalError{Err: err}
	}

	// Ensure that the election identifier matches the configured one.
	if l.Election != c.Election.Identifier {
		return ElectionIDMismatchError{Conf: c.Election.Identifier, List: l.Election}
	}

	// Convert districts to lookup table where a administrative unit code
	// and district number map to the district identifier.
	districts, err := districtsLookup(l)
	if err != nil {
		return err
	}

	// Although Counties is stored as-is, ensure that it is of expected
	// type, i.e., a JSON object with string array values.
	var counties map[string][]string
	if err := json.Unmarshal(l.Counties, &counties); err != nil {
		return JSONUnmarshalCountiesError{Err: err}
	}

	if until >= command.Execute {
		log.Log(ctx, ImportingDistricts{Count: len(l.Districts)})
		progress.Static(fmt.Sprintf("Importing %d districts:", len(l.Districts)))
		addprogress := progress.Percent(uint64(len(districts)), true)
		progress.Redraw()
		defer progress.Keep()

		err := s.PutDistricts(ctx, version, districts, l.Counties, addprogress)
		if err != nil {
			return PutDistrictsError{Err: err}
		}
	}
	return nil
}

// districtsLookup converts the districts from a district list to a lookup
// table from an administrative unit code and district number to a district
// identifier.
func districtsLookup(list districtlist) (map[string][]byte, error) {
	districts := make(map[string][]byte)
	parishCodes := make(map[string]string)
	for id, district := range list.Districts {
		splits := strings.Split(id, ".")
		if len(splits) != 2 || len(splits[0]) == 0 || len(splits[1]) == 0 {
			// Must contain a single period, but not first or last.
			return nil, DistrictIDFormatError{District: id}
		}
		code, nr := splits[0], splits[1]

		for _, parish := range district.Parish {
			if parishCode, ok := parishCodes[parish]; ok && code != parishCode {
				return nil, ParishWithMultipleDistrictCodesError{
					First:  parishCode,
					Second: code,
				}
			}

			districts[string(storage.EncodeAdminDistrict(parish, nr))] = []byte(id)
			parishCodes[parish] = code
		}
	}
	return districts, nil
}
