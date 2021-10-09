package main

import (
	"archive/zip"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"regexp"

	"ivxv.ee/conf/version"
	"ivxv.ee/container"
	"ivxv.ee/errors"
)

// openContainer attemps to open path as a trusted signed container using
// opener, but if this fails because it is an unknown type, then it opens it as
// an untrusted ZIP-archive mimicing the container interface.
func openContainer(opener container.Opener, path string) (
	cnt container.Container, trusted bool, err error) {

	cnt, err = opener.OpenFile(path)
	if errors.CausedBy(err, new(container.UnconfiguredTypeError)) == nil {
		return cnt, true, err
	}

	// OpenFile failed because it was an unknown type: check if it is ZIP.
	if filepath.Ext(path) != ".zip" {
		return nil, false, err // Return the UnconfiguredTypeError.
	}

	archive, err := zip.OpenReader(path)
	if err != nil {
		return nil, false, OpenZIPContainerError{Err: err}
	}
	defer archive.Close()

	data := make(map[string][]byte, len(archive.File))
	for _, file := range archive.File {
		rc, err := file.Open()
		if err != nil {
			return nil, false, OpenZIPFileError{File: file.Name, Err: err}
		}
		defer rc.Close()

		bytes, err := ioutil.ReadAll(rc)
		if err != nil {
			return nil, false, ReadZIPFileError{File: file.Name, Err: err}
		}
		data[file.Name] = bytes
	}

	return zipContainer{
		comment: archive.Comment,
		data:    data,
	}, false, nil
}

type zipContainer struct {
	comment string
	data    map[string][]byte
}

// Close does nothing, because there are no resources to manually free.
func (z zipContainer) Close() error {
	return nil
}

// Signatures returns an empty slice since the container is not signed.
func (z zipContainer) Signatures() []container.Signature {
	return nil
}

// Data returns the container data from the ZIP-archive.
func (z zipContainer) Data() map[string][]byte {
	return z.data
}

var containerVersionRE = regexp.MustCompile("(?m)^Version: +(.+)$")

// containerVersion returns a container's version string. If the container is a
// trusted signed container, then it calls version.Container, otherwise it
// parses version metadata from the ZIP-archive comments.
func containerVersion(cnt container.Container) (string, error) {
	z, ok := cnt.(zipContainer)
	if !ok {
		return version.Container(cnt)
	}

	matches := containerVersionRE.FindAllStringSubmatch(z.comment, -1)
	if len(matches) == 0 {
		return "", MissingZIPVersionError{}
	}
	versions := make([]string, len(matches))
	for i, match := range matches {
		versions[i] = match[1] // The first submatch is the version string.
	}
	version, err := json.Marshal(versions)
	return string(version), err
}
