/*
The voterimp application is used for loading voter lists into the storage
service.

The voter list is generated and signed by the Estonian Election Information
System and contains Estonian strings, so the format is very specific to
Estonia. voterimp checks the raw ECDSA signature with SHA-256 digest, parses
the contents of the voter list, and adds a new version of voter information
into the storage service.
*/
package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"ivxv.ee/command"
	"ivxv.ee/command/exit"
	"ivxv.ee/command/status"
	"ivxv.ee/cryptoutil"
	"ivxv.ee/errors"
	"ivxv.ee/log"
	"ivxv.ee/storage"
	//ivxv:modules container
	//ivxv:modules storage
)

const usage = `voterimp loads voter lists into the collector's storage service for use by
other services.

The voter list container must contain exactly two files: the voter list and the
voter list signature. The key for the voter list can be anything, but the key
for the signature must be the list key with a ".signature" suffix. E.g.,
"voter.list" and "voter.list.signature".

The voter list container must have an extension corresponding to the container
type it is, e.g., voterlist.bdoc. voterimp additionally supports unsigned ZIP
containers with metadata stored in the archive comment.`

var (
	qp = flag.Bool("q", false, "quiet, do not show progress")

	progress *status.Line
)

func main() {
	// Call voterimpmain in a separate function so that it can set up
	// defers and have them trigger before returning with a non-zero exit
	// code.
	os.Exit(voterimpmain())
}

func voterimpmain() (code int) {
	c := command.New("ivxv-voterimp", usage, "voter list container")
	defer func() {
		code = c.Cleanup(code)
	}()

	if !*qp {
		progress = status.New()
	}

	// Only check the version if it was the specific check requested.
	if c.Until == command.CheckVersion {
		version, err := c.Storage.GetVotersContainerVersions(c.Ctx)
		if err != nil {
			if errors.CausedBy(err, new(storage.NotExistError)) == nil {
				return c.Error(exit.Unavailable, CheckVersionError{Err: err},
					"failed to check imported container versions:", err)
			}
		} else {
			// Convert from new-line separation to JSON array.
			var vers []json.RawMessage
			for _, ver := range strings.Split(strings.TrimSpace(version), "\n") {
				vers = append(vers, json.RawMessage(ver))
			}
			if err := json.NewEncoder(os.Stdout).Encode(vers); err != nil {
				return c.Error(exit.Unavailable, ConvertVersionError{Err: err},
					"failed to encode container versions:", err)
			}
		}
	}

	if c.Until < command.CheckInput {
		return exit.OK
	}
	path := c.Args[0]

	// Open the voter list container file.
	cnt, trusted, err := openContainer(c.Conf.Container, path)
	if err != nil {
		code = exit.DataErr
		if perr := errors.CausedBy(err, new(os.PathError)); perr != nil {
			if os.IsNotExist(perr) {
				code = exit.NoInput
			}
		}
		return c.Error(code, OpenContainerError{Container: path, Err: err},
			"failed to open voter list container:", err)
	}
	defer cnt.Close()

	// Ensure that a trusted container is signed and log the signatures.
	signatures := cnt.Signatures()
	if trusted && len(signatures) == 0 {
		return c.Error(exit.DataErr, UnsignedContainerError{Container: path},
			"unsigned voter list container")
	}
	for _, s := range signatures {
		log.Log(c.Ctx, ContainerSignature{Signer: s.Signer, SigningTime: s.SigningTime})
	}

	// Get the version string of the container.
	cversion, err := containerVersion(cnt)
	if err != nil {
		return c.Error(exit.DataErr, ContainerVersionError{Container: path, Err: err},
			"failed to format container version string:", err)
	}

	// Get the contents of the container.
	list, sig, err := containerData(cnt.Data())
	if err != nil {
		return c.Error(exit.DataErr, ContainerDataError{Err: err},
			"failed to find expected data from container:", err)
	}

	// Check the signature.
	if err = verifyECDSA(c.Conf.Election.VoterList.Key, list, sig); err != nil {
		return c.Error(exit.DataErr, VerifySignatureError{Err: err},
			"failed to verify voter list signature:", err)
	}

	// Get the current voter list version.
	oldver, err := c.Storage.GetVotersListVersion(c.Ctx)
	switch {
	case err == nil:
		log.Log(c.Ctx, CurrentVoterListVersion{Version: oldver})
	case errors.CausedBy(err, new(storage.NotExistError)) != nil:
		oldver = "" // Do not rely on storage returning zero value on error.
	default:
		return c.Error(exit.Unavailable, CheckListVersionError{Err: err},
			"failed to check current list version:", err)
	}

	// Parse the voter list and preprocess into a map of changes.
	voters, newver, err := preprocess(c.Ctx, list, oldver, c.Conf.Election.Identifier, c.Storage)
	if err != nil {
		return c.Error(exit.DataErr, PreprocessVotersError{Err: err},
			"failed to preprocess voter list:", err)
	}

	// Store the new list.
	if c.Until >= command.Execute {
		log.Log(c.Ctx, ImportingVoters{Version: newver, Count: len(voters)})
		progress.Static(fmt.Sprintf("Importing %d voters:", len(voters)))
		addprogress := progress.Percent(uint64(len(voters)), true)
		progress.Redraw()
		defer progress.Keep()

		if err := c.Storage.PutVoters(c.Ctx, cversion, voters,
			oldver, newver, addprogress); err != nil {

			return c.Error(exit.Unavailable, PutVotersError{Err: err},
				"failed to import voter list:", err)
		}
	}
	return exit.OK
}

// containerData returns voter list and signature contents from container data.
func containerData(data map[string][]byte) (list, signature []byte, err error) {
	// The container data should have exactly two keys: *.utf and *.sig.
	// Although the files are usually named more specifically
	//
	//   <election_identifier>-voters-<changeset>.{utf,sig},
	//
	// do not enforce this.
	const utf = ".utf"
	const sig = ".sig"

	if len(data) != 2 {
		return nil, nil, KeyCountError{Count: len(data)}
	}
	var utfKey, sigKey string
	for key, content := range data {
		if strings.HasSuffix(key, utf) {
			utfKey = key
			list = content
			sigKey = key[:len(key)-len(utf)] + sig
			break
		}
	}
	if len(utfKey) == 0 {
		return nil, nil, MissingUTFKeyError{}
	}
	signature, ok := data[sigKey]
	if !ok {
		return nil, nil, MissingSigKeyError{Expected: sigKey}
	}

	return list, signature, nil
}

// verifyECDSA parses an ECDSA public key from a PEM-encoded X.509 structure
// and verifies the signature sig on data.
func verifyECDSA(pub string, data, sig []byte) error {
	der, err := cryptoutil.PEMDecode(pub, "PUBLIC KEY")
	if err != nil {
		return ECDSAPEMDecodeError{Err: err}
	}
	parsed, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return ECDSAParsePKIXError{Err: err}
	}
	key, ok := parsed.(*ecdsa.PublicKey)
	if !ok {
		return ECDSAPublicKeyNotECDSAError{Type: fmt.Sprintf("%T", parsed)}
	}

	hashed := sha256.Sum256(data) // Hardcoded regardless of key parameters.
	r, s, err := cryptoutil.ParseECDSAASN1Signature(sig)
	if err != nil {
		return ECDSAParseSignatureError{Err: err}
	}
	if !ecdsa.Verify(key, hashed[:], r, s) {
		return ECDSASignatureVerificationError{}
	}
	return nil
}

const (
	delim = '\n'
	sep   = '\t'
)

// linefunc is the type of functions used to process voter lines. Given a
// voter, action, administrative unit code and district number, it reports any
// problems or if there were none, adds an entry for voter into voters.
// previousErrors indicates if there were previous lines with errors for this
// voter. version is the currently applied voter list version.
type linefunc func(ctx context.Context, voter, action, adminCode, district string,
	voters map[string][]byte, previousErrors bool, version string, s *storage.Client) (
	errs []error)

// preprocess parses the list and preprocesses the changes for storage.
func preprocess(ctx context.Context, list []byte, version, election string, s *storage.Client) (
	voters map[string][]byte, newver string, err error) {

	b := bytes.NewBuffer(list)

	// Parse the header to determine list version and type.
	newver, lf, err := header(b, election, version)
	if err != nil {
		return nil, "", err
	}

	// Loop over all list entries, calling lf for each. Report progress of
	// this process.
	log.Log(ctx, PreprocessingVoterList{Version: newver})
	progress.Static("Preprocessing voter list:")
	addcount := progress.Count(0, false) // Do not redraw every time.
	const countstep = 10000              // Redraw after each countstep.
	progress.Redraw()
	defer progress.Keep()

	voters = make(map[string][]byte)
	withErrors := make(map[string]struct{}) // Voters with previous errors.
	var errcount int

	line := 3 // Voter entries start after the third line.
loop:
	for ; ; stepadd(ctx, addcount, 1, countstep) {
		// Check if preprocessing was cancelled.
		select {
		case <-ctx.Done():
			return nil, "", PreprocessVoterListCanceled{Err: ctx.Err()}
		default:
		}

		// Read the next line.
		line++
		action, voter, adminCode, district, err := next(b)
		switch {
		case err == nil:
		case err == io.EOF:
			break loop
		case errors.CausedBy(err, new(FieldCountError)) != nil:
			report(ctx, VoterEntryFormatError{Line: line, Err: err})
			errcount++
			continue loop
		default:
			report(ctx, ReadVoterEntryError{Line: line, Err: err})
			errcount++
			break loop
		}

		// Call lf for the line.
		_, previous := withErrors[voter]
		errs := lf(ctx, voter, action, adminCode, district, voters, previous, version, s)
		if len(errs) > 0 {
			withErrors[voter] = struct{}{}
		}
		for _, err = range errs {
			report(ctx, PreprocessVoterEntryError{Line: line, Err: err})
			errcount++
		}
	}
	// Report final count if not already done by stepadd.
	if addcount(0)%countstep > 0 {
		stepadd(ctx, addcount, 0, 0)
	}
	if errcount > 0 {
		return nil, "", PreprocessVoterListError{ErrorCount: errcount}
	}
	return voters, newver, nil
}

func header(b *bytes.Buffer, election, oldver string) (newver string, lf linefunc, err error) {
	// First line is the format version number, which must be 2.
	fver, err := readString(b, delim)
	if err != nil {
		return "", nil, ReadFileVersionError{Err: err}
	}
	if fver != "2" {
		return "", nil, FileVersionError{Version: fver}
	}

	// Second is the election identifier. Must match the one in the election
	// configuration.
	elid, err := readString(b, delim)
	if err != nil {
		return "", nil, ReadElectionIDError{Err: err}
	}
	if elid != election {
		return "", nil, ElectionIDMismatchError{Conf: election, List: elid}
	}

	// Third is the list version number.
	newver, err = readString(b, delim)
	if err != nil {
		return "", nil, ReadListVersionError{Err: err}
	}
	newver64, err := strconv.ParseUint(newver, 10, 64)
	if err != nil {
		return "", nil, ParseListVersionError{Err: err}
	}
	if newver64 == 0 {
		if len(oldver) > 0 {
			return "", nil, VoterListExistsError{Version: oldver}
		}
		lf = initial
	} else {
		var oldver64 uint64
		if len(oldver) == 0 {
			return "", nil, NoExistingVoterListError{}
		} else if oldver64, err = strconv.ParseUint(oldver, 10, 64); err != nil {
			return "", nil, ParseCurrentListVersionError{Err: err}
		} else if oldver64 >= newver64 {
			return "", nil, UnexpectedListVersionError{
				Current: oldver64,
				List:    newver64,
			}
		}
		lf = changes
	}

	// Fourth are two timestamps representing the period. Not used by the
	// collector, but make sure that the format is valid.
	fromstr, err := readString(b, sep)
	if err != nil {
		return "", nil, ReadPeriodFromError{Err: err}
	}
	if _, err = time.Parse(time.RFC3339, fromstr); err != nil {
		return "", nil, ParsePeriodFromError{Err: err}
	}
	tostr, err := readString(b, delim)
	if err != nil {
		return "", nil, ReadPeriodToError{Err: err}
	}
	if _, err = time.Parse(time.RFC3339, tostr); err != nil {
		return "", nil, ParsePeriodToError{Err: err}
	}

	return newver, lf, nil
}

func initial(ctx context.Context, voter, action, adminCode, district string,
	voters map[string][]byte, previousErrors bool, version string, s *storage.Client) (
	errs []error) {

	// Skip duplicate checking if there are relevant errors. We want to
	// check duplicates even if the voter had previous errors, so use a
	// flag separate from previousErrors.
	var skip bool

	if len(voter) == 0 {
		errs = append(errs, InitialEmptyVoterError{})
		skip = true
	}
	if action != "lisamine" {
		errs = append(errs, InitialNonAddActionError{Action: action})
		skip = true
	}
	if len(adminCode) == 0 {
		errs = append(errs, InitialEmptyAdminUnitCodeError{})
	}
	if len(district) == 0 {
		errs = append(errs, InitialEmptyDistrictNumberError{})
	}
	if !skip {
		if _, ok := voters[voter]; ok {
			errs = append(errs, InitialAddDuplicateVoterError{Voter: voter})
		}
		// It is OK to add invalid entries (skip was not set for
		// adminCode or district errors) since then we will error
		// anyway and the map values are not used during preprocessing.
		voters[voter] = storage.EncodeAdminDistrict(adminCode, district)
	}
	return
}

func changes(ctx context.Context, voter, action, adminCode, district string,
	voters map[string][]byte, previousErrors bool, version string, s *storage.Client) (
	errs []error) {

	if len(voter) == 0 {
		errs = append(errs, ChangesEmptyVoterError{})
		previousErrors = true
	} else if previousErrors {
		errs = append(errs, ChangesVoterWithPreviousErrorsError{Voter: voter})
	}

	// Only perform consistency checks if there are no previous errors.
	var voterExists bool    // Is this voter on the current voter list?
	var voterProcessed bool // Has this voter already been processed?
	if !previousErrors {
		// Check the current entry for the voter. Look at the in-memory
		// map first and only then storage.
		if entry, ok := voters[voter]; ok {
			voterExists = entry != nil
			voterProcessed = true
		} else {
			var err error
			_, _, err = s.GetVoter(ctx, version, voter)
			switch {
			case err == nil:
				voterExists = true
			case errors.CausedBy(err, new(storage.NotExistError)) != nil:
			default:
				errs = append(errs, GetOldVoterError{Voter: voter, Err: err})
				previousErrors = true
			}
		}
	}

	switch action {
	case "lisamine":
		if len(adminCode) == 0 {
			errs = append(errs, ChangesEmptyAdminUnitCodeError{})
			previousErrors = true
		}
		if len(district) == 0 {
			errs = append(errs, ChangesEmptyDistrictNumberError{})
			previousErrors = true
		}
		if !previousErrors {
			if voterExists {
				errs = append(errs, ChangesAddDuplicateVoterError{Voter: voter})
			}
			voters[voter] = storage.EncodeAdminDistrict(adminCode, district)
		}

	case "kustutamine":
		if !previousErrors {
			switch {
			case !voterExists:
				errs = append(errs, ChangesRemoveNotExistingVoterError{Voter: voter})
			case voterProcessed:
				errs = append(errs, ChangesRemoveAddedVoterError{Voter: voter})
			}
			voters[voter] = nil // nil to distinguish from unchanged voters.
		}

	default:
		errs = append(errs, ChangesUnsupportedActionError{Action: action})
	}
	return
}

// stepadd calls add with count and, if step is zero or the new total modulo
// step is zero, logs and redraws progress.
func stepadd(ctx context.Context, add status.Add, count, step uint64) {
	if new := add(count); step == 0 || new%step == 0 {
		log.Log(ctx, PreprocessProgress{Count: new})
		progress.Redraw()
	}
}

// next reads and parses the next voter line from b.
func next(b *bytes.Buffer) (action, voter, adminCode, district string, err error) {
	line, err := readString(b, delim)
	if err != nil {
		if err != io.EOF || len(line) > 0 {
			err = ReadNextError{Line: line, Err: err}
		}
		return
	}

	// Split on tabs and expect five fields or two fields when 'kustutamine':
	//
	//	0. action,
	//	1. voter ID,
	//	2. voter name (ignored),
	//	3. administrative unit code, and
	//	4. district number.
	//
	//	0. action,
	//	1. voter ID,
	//
	// The choices identifier is formed from the administrative unit code
	// and district number.
	fields := strings.Split(line, string(sep))
	switch len(fields) {
	case 2:
		return fields[0], fields[1], "", "", nil
	case 5:
		return fields[0], fields[1], fields[3], fields[4], nil
	default:
		err = FieldCountError{Fields: len(fields)}
		return
	}
}

// report reports an error to both the log and standard output.
func report(ctx context.Context, entry log.ErrorEntry) {
	log.Error(ctx, entry)
	progress.Hide()
	defer progress.Show()
	fmt.Fprintln(os.Stderr, "error:", entry)
}

// readString returns the result of b.ReadString(delim) with delim trimmed.
func readString(b *bytes.Buffer, delim byte) (string, error) {
	line, err := b.ReadString(delim)
	if err == nil {
		line = line[:len(line)-1] // Trim the delimiter.
	}
	return line, err
}
