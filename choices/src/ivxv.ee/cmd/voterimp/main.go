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
	"fmt"
	"io"
	"os"
	"strings"

	"ivxv.ee/command"
	"ivxv.ee/command/exit"
	"ivxv.ee/conf"
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
	path := c.Args[0]

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
	version, err := version.Container(cnt)
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
	if err = verifyRSA(c.Conf.Technical.VoterList.Key, list, sig); err != nil {
		return c.Error(exit.DataErr, VerifySignatureError{Err: err},
			"voter list signature failed to verify:", err)
	}

	// Import a new version of the voter list.
	if err := voterimp(c.Ctx, c.Until, c.Conf, c.Storage, version, list); err != nil {
		return c.Error(exit.Unavailable, ImportVotersError{Err: err},
			"failed to import voter list:", err)
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

// voterimp parses the list and uploads the changes to the storage service.
func voterimp(ctx context.Context, until int, c *conf.C, s *storage.Client,
	version string, list []byte) error {

	b := bytes.NewBuffer(list)

	// First line is the format version number, which must be 1.
	if fver, err := b.ReadString(delim); err != nil {
		return ReadFileVersionError{Err: err}
	} else if fver != "1"+string(delim) {
		return FileVersionError{Version: fver}
	}

	// Next is the election identifier. Must match the one in the election
	// configuration.
	if elid, err := b.ReadString(delim); err != nil {
		return ReadElectionIDError{Err: err}
	} else if elid != c.Election.Identifier+string(delim) {
		return ElectionIDMismatchError{Conf: c.Election.Identifier, List: elid}
	}

	// Next is the list type: either initial or changes.
	t, err := b.ReadString(delim)
	if err != nil {
		return ReadTypeError{Err: err}
	}
	t = t[:len(t)-1] // Trim the delimiter.

	switch t {
	case "algne":
		return initial(ctx, until, s, version, list, b)
	case "muudatused":
		return changes(ctx, until, s, version, list, b)
	default:
		return UnsupportedTypeError{Type: t}
	}
}

func initial(ctx context.Context, until int, s *storage.Client,
	cversion string, list []byte, b *bytes.Buffer) error {

	// Check if the context is canceled before contacting storage.
	select {
	case <-ctx.Done():
		return ImportInitialPreprocessingCanceled{Err: ctx.Err()}
	default:
	}

	// Fail fast: check if a list is already loaded.
	oldver, err := s.GetVotersListVersion(ctx)
	switch {
	case err == nil:
		return InitialVoterListExistsError{Version: oldver}
	case errors.CausedBy(err, new(storage.NotExistError)) != nil:
	default:
		return InitialCheckVoterListExistingError{Err: err}
	}

	// Preprocess the initial list.
	log.Log(ctx, PreprocessingInitialList{})
	progress := PreprocessingInitial{Voter: "", Choices: "", Index: 0}

	voters := make(map[string]string)
	line := 3 // Voter entries start after the third line.
	for {
		// Read the next line.
		line++
		voter, action, choices, err := next(ctx, b)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// If debugging is enabled, then report progress.
		progress.Voter = voter
		progress.Choices = choices
		progress.Index = progress.Index.(int) + 1
		log.Debug(ctx, progress)

		// Ensure that we are adding a new non-empty voter with
		// non-empty choices.
		if action != "lisamine" {
			return InitialNonAddActionError{Line: line, Action: action}
		}
		if len(voter) == 0 {
			return InitialEmptyVoterError{Line: line}
		}
		if len(choices) == 0 {
			return InitialEmptyChoicesError{Line: line, Voter: voter}
		}
		if _, ok := voters[voter]; ok {
			return InitialAddDuplicateVoterError{Line: line, Voter: voter}
		}
		voters[voter] = choices
	}

	// Calculate initial list version.
	v := listver("", list)

	// Check if the context is canceled before the actual import.
	select {
	case <-ctx.Done():
		return ImportInitialCanceled{Err: ctx.Err()}
	default:
	}

	// Store the initial voters.
	if until >= command.Execute {
		log.Log(ctx, ImportingInitialVoters{Version: v, Count: len(voters)})
		if err := s.PutVoters(ctx, cversion, voters, "", v); err != nil {
			return InitialPutVotersError{Err: err}
		}
	}
	return nil
}

func changes(ctx context.Context, until int, s *storage.Client,
	cversion string, list []byte, b *bytes.Buffer) error {

	// Check if the context is canceled before contacting storage.
	select {
	case <-ctx.Done():
		return ImportChangesPreprocessingCanceled{Err: ctx.Err()}
	default:
	}

	// Get the current voter list version.
	oldver, err := s.GetVotersListVersion(ctx)
	if err != nil {
		return ChangesGetVoterListVersionError{Err: err}
	}
	log.Log(ctx, CurrentVoterListVersion{Version: oldver})

	// Preprocess changes to the voter list.
	log.Log(ctx, PreprocessingChanges{})
	progress := PreprocessingChange{Action: "", Voter: "", Choices: "", Index: 0}

	voters := make(map[string]string)
	line := 3 // Voter entries start after the third line.
	for {
		// Read the next line.
		line++
		voter, action, choices, err := next(ctx, b)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// If debugging is enabled, then report progress.
		progress.Action = action
		progress.Voter = voter
		progress.Choices = choices
		progress.Index = progress.Index.(int) + 1
		log.Debug(ctx, progress)

		// Get the current choices for the voter. Check the in-memory
		// map first and only then storage.
		oldchoices, ok := voters[voter]
		if !ok {
			oldchoices, err = s.GetVoter(ctx, oldver, voter)
			switch {
			case err == nil:
			case errors.CausedBy(err, new(storage.NotExistError)) != nil:
				oldchoices = ""
			default:
				return GetOldVoterError{Line: line, Err: err}
			}
		}

		switch action {
		case "lisamine":
			// Ensure that we are adding a new non-empty voter with
			// non-empty choices or...
			if len(voter) == 0 {
				return ChangesEmptyVoterError{Line: line}
			}
			if len(choices) == 0 {
				return ChangesEmptyChoicesError{Line: line, Voter: voter}
			}
			if len(oldchoices) > 0 {
				return ChangesAddExistingVoterError{Line: line, Voter: voter}
			}
			voters[voter] = choices

		case "kustutamine":
			// ...removing an existing voter with the specified choices.
			if len(oldchoices) == 0 {
				return ChangesRemoveNotExistingVoterError{Line: line, Voter: voter}
			}
			if oldchoices != choices {
				return ChangesRemoveVoterChoicesMismatchError{
					Line:     line,
					Voter:    voter,
					Choices:  choices,
					Expected: oldchoices,
				}
			}
			voters[voter] = "" // Empty string to distinguish from unchanged voters.

		default:
			return ChangesUnsupportedActionError{Line: line, Action: action}
		}
	}

	// Calculate new list version.
	newver := listver(oldver, list)

	// Check if the context is canceled before the actual import.
	select {
	case <-ctx.Done():
		return ImportChangesCanceled{Err: ctx.Err()}
	default:
	}

	// Store the changed list.
	if until >= command.Execute {
		log.Log(ctx, ImportingChangedVoters{Version: newver, Count: len(voters)})
		if err := s.PutVoters(ctx, cversion, voters, oldver, newver); err != nil {
			return ChangesPutVotersError{Err: err}
		}
	}
	return nil
}

// next reads and parses the next voter line from b.
func next(ctx context.Context, b *bytes.Buffer) (voter, action, choices string, err error) {
	// Check if we should give the next line or cancel.
	select {
	case <-ctx.Done():
		err = PreprocessVoterListCanceled{Err: ctx.Err()}
		return
	default:
	}

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
