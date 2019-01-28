/*
The voterimp application is used for loading voter lists into the storage
service.

The voter list is generated and signed by the Estonian Population Register and
contains Estonian strings, so the format is very specific to Estonia. voterimp
checks the raw RSA signature with SHA-256 and PKCS #1 v1.5 encoding, parses the
contents of the voter list, and adds a new version of voter information into
the storage service.
*/
package main

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"ivxv.ee/command"
	"ivxv.ee/command/exit"
	"ivxv.ee/command/status"
	"ivxv.ee/conf/version"
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
type it is, e.g., voterlist.bdoc.`

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
	cnt, err := c.Conf.Container.OpenFile(path)
	// nolint: dupl, ignore duplicate from choiceimp, extracting this
	// snippet of code to a common function is not worth the effort.
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
	defer cnt.Close() // nolint: errcheck, ignore close failure of read-only container.

	// Ensure that the container is signed and log the signatures.
	signatures := cnt.Signatures()
	if len(signatures) == 0 {
		return c.Error(exit.DataErr, UnsignedContainerError{Container: path},
			"unsigned voter list container")
	}
	for _, s := range signatures {
		log.Log(c.Ctx, ContainerSignature{Signer: s.Signer, SigningTime: s.SigningTime})
	}

	// Get the version string of the container.
	cversion, err := version.Container(cnt)
	if err != nil {
		return c.Error(exit.DataErr, ContainerVersionError{Container: path, Err: err},
			"failed to format container version string:", err)
	}

	// Check that the container has only two files: * and *.signature.
	data := cnt.Data()
	if len(data) != 2 {
		return c.Error(exit.DataErr, DataCountError{Count: len(data)},
			"voter list container has", len(data), "files, expected 2")
	}
	keys := make([]string, 0, 2)
	for key := range data {
		keys = append(keys, key)
	}
	if len(keys[0]) > len(keys[1]) {
		keys[0], keys[1] = keys[1], keys[0] // Put shorter one first.
	}
	expected := keys[0] + ".signature"
	if keys[1] != expected {
		return c.Error(exit.DataErr, KeyError{Keys: keys},
			fmt.Sprintf("voter list container has keys %q and %q, but want %[1]q and %[3]q",
				keys[0], keys[1], expected))
	}
	list, sig := data[keys[0]], data[keys[1]]

	// Check the signature.
	if err = verifyRSA(c.Conf.Election.VoterList.Key, list, sig); err != nil {
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
	voters, err := preprocess(c.Ctx, list, oldver, c.Conf.Election.Identifier, c.Storage)
	if err != nil {
		return c.Error(exit.DataErr, PreprocessVotersError{Err: err},
			"failed to preprocess voter list:", err)
	}

	// Store the new list.
	if c.Until >= command.Execute {
		newver := listver(oldver, list)
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

// verifyRSA parses an RSA public key from a PEM-encoded X.509 structure and
// verifies the signature sig on data using SHA-256 and PKCS #1 v1.5 encoding.
func verifyRSA(pub string, data, sig []byte) error {
	der, err := cryptoutil.PEMDecode(pub, "PUBLIC KEY")
	if err != nil {
		return RSAPEMDecodeError{Err: err}
	}

	// subjectPublicKeyInfo is the X.509 structure used to encode public
	// keys. See https://tools.ietf.org/html/rfc5280#section-4.1.2.7
	type subjectPublicKeyInfo struct {
		Algorithm        pkix.AlgorithmIdentifier
		SubjectPublicKey asn1.BitString
	}
	oidRSA := asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 1}

	var x509 subjectPublicKeyInfo
	if _, err := asn1.Unmarshal(der, &x509); err != nil {
		return RSAUnmarshalX509Error{Err: err}
	}
	if !oidRSA.Equal(x509.Algorithm.Algorithm) {
		return RSANotAnRSAKeyError{OID: x509.Algorithm.Algorithm}
	}

	key := new(rsa.PublicKey)
	if _, err := asn1.Unmarshal(x509.SubjectPublicKey.RightAlign(), key); err != nil {
		return RSAUnmarshalKeyError{Err: err}
	}

	hashed := sha256.Sum256(data)
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, hashed[:], sig); err != nil {
		return RSAVerifyError{Err: err}
	}
	return nil
}

const (
	delim = '\n'
	sep   = '\t'
)

// linefunc is the type of functions used to process voter lines. Given a
// voter, action and choices, it reports any problems or if there were none,
// adds an entry for voter into voters. previousErrors indicates if there were
// previous lines with errors for this voter. version is the currently applied
// voter list version.
type linefunc func(ctx context.Context, voter, action, choices string, voters map[string][]byte,
	previousErrors bool, version string, s *storage.Client) []error

// preprocess parses the list and preprocesses the changes for storage.
func preprocess(ctx context.Context, list []byte, version, election string, s *storage.Client) (
	voters map[string][]byte, err error) {

	b := bytes.NewBuffer(list)

	// First line is the format version number, which must be 1.
	fver, err := b.ReadString(delim)
	if err != nil {
		return nil, ReadFileVersionError{Err: err}
	}
	if fver != "1"+string(delim) {
		return nil, FileVersionError{Version: fver}
	}

	// Next is the election identifier. Must match the one in the election
	// configuration.
	elid, err := b.ReadString(delim)
	if err != nil {
		return nil, ReadElectionIDError{Err: err}
	}
	if elid != election+string(delim) {
		return nil, ElectionIDMismatchError{Conf: election, List: elid}
	}

	// Next is the list type: either initial or changes.
	t, err := b.ReadString(delim)
	if err != nil {
		return nil, ReadTypeError{Err: err}
	}
	t = t[:len(t)-1] // Trim the delimiter.

	var lf linefunc
	switch t {
	case "algne":
		if len(version) > 0 {
			return nil, VoterListExistsError{Version: version}
		}
		lf = initial
	case "muudatused":
		if len(version) == 0 {
			return nil, NoExistingVoterListError{}
		}
		lf = changes
	default:
		return nil, UnsupportedListTypeError{Type: t}
	}

	// Finally, loop over all list entries, calling lf for each. Report
	// progress of this process.
	log.Log(ctx, PreprocessingVoterList{Type: t})
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
			return nil, PreprocessVoterListCanceled{Err: ctx.Err()}
		default:
		}

		// Read the next line.
		line++
		voter, action, choices, err := next(b)
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
		errs := lf(ctx, voter, action, choices, voters, previous, version, s)
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
		return nil, PreprocessVoterListError{ErrorCount: errcount}
	}
	return voters, nil
}

func initial(ctx context.Context, voter, action, choices string, voters map[string][]byte,
	previousErrors bool, version string, s *storage.Client) (errs []error) {

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
	if len(choices) == 0 {
		errs = append(errs, InitialEmptyChoicesError{})
		// Do not set skip: detect duplicate voter regardless of faulty
		// choices. It is OK to add empty choices since we will error
		// anyway and the map values are not used.
	}
	if !skip {
		if _, ok := voters[voter]; ok {
			errs = append(errs, InitialAddDuplicateVoterError{Voter: voter})
		}
		voters[voter] = []byte(choices)
	}
	return
}

func changes(ctx context.Context, voter, action, choices string, voters map[string][]byte,
	previousErrors bool, version string, s *storage.Client) (errs []error) {

	if len(voter) == 0 {
		errs = append(errs, ChangesEmptyVoterError{})
		previousErrors = true
	} else if previousErrors {
		errs = append(errs, ChangesVoterWithPreviousErrorsError{Voter: voter})
	}

	if len(choices) == 0 {
		errs = append(errs, ChangesEmptyChoicesError{})
		previousErrors = true
	}

	// Only perform consistency checks if there are no previous errors.
	var oldchoices []byte
	if !previousErrors {
		// Get the current choices for the voter. Check the
		// in-memory map first and only then storage.
		var ok bool
		if oldchoices, ok = voters[voter]; !ok {
			oldchoicesstr, err := s.GetVoter(ctx, version, voter)
			switch {
			case err == nil:
				oldchoices = []byte(oldchoicesstr)
			case errors.CausedBy(err, new(storage.NotExistError)) != nil:
			default:
				errs = append(errs, GetOldVoterError{Voter: voter, Err: err})
				previousErrors = true
			}
		}
	}

	switch action {
	case "lisamine":
		if !previousErrors {
			if len(oldchoices) > 0 {
				errs = append(errs, ChangesAddDuplicateVoterError{Voter: voter})
			}
			voters[voter] = []byte(choices)
		}

	case "kustutamine":
		if !previousErrors {
			if len(oldchoices) == 0 {
				errs = append(errs, ChangesRemoveNotExistingVoterError{Voter: voter})
			} else if string(oldchoices) != choices {
				errs = append(errs, ChangesRemoveVoterChoicesMismatchError{
					Voter:    voter,
					Choices:  choices,
					Expected: string(oldchoices),
				})
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
func next(b *bytes.Buffer) (voter, action, choices string, err error) {
	line, err := b.ReadString(delim)
	if err != nil {
		if err != io.EOF || len(line) > 0 {
			err = ReadNextError{Line: line, Err: err}
		}
		return
	}
	line = line[:len(line)-1] // Trim the delimiter.

	// Split on tabs and expect nine fields:
	//
	//	0. voter ID,
	//	1. voter name (ignored),
	//	2. action,
	//	3. voting station EHAK (ignored),
	//	4. voting station number (ignored),
	//	5. district EHAK,
	//	6. district number,
	//	7. row number (ignored), and
	//	8. change reason (ignored).
	//
	// The choices identifier is formed from the district EHAK and number.
	fields := strings.Split(line, string(sep))
	if len(fields) != 9 {
		err = FieldCountError{Fields: len(fields), Expected: 9}
		return
	}

	return fields[0], fields[2], fmt.Sprint(fields[5], ".", fields[6]), nil
}

// report reports an error to both the log and standard output.
func report(ctx context.Context, entry log.ErrorEntry) {
	log.Error(ctx, entry)
	progress.Hide()
	defer progress.Show()
	fmt.Fprintln(os.Stderr, "error:", entry)
}

// listever calculates a list version number based on the previous one and data.
func listver(old string, data []byte) string {
	b64 := base64.StdEncoding
	n := b64.EncodedLen(sha256.Size)
	v := make([]byte, len(old)+n)

	// v = old | b64(sha256(data))
	copy(v, old)
	h := sha256.Sum256(data)
	b64.Encode(v[len(old):], h[:])

	// v = b64(sha256(v))
	h = sha256.Sum256(v)
	b64.Encode(v, h[:])
	return string(v[:n])
}
