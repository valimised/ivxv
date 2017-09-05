/*
Package version contains types for denoting the version of a configuration
file.

This is a separate package to avoid import cycles between ivxv.ee/conf and
ivxv.ee/server.
*/
package version

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"ivxv.ee/container"
)

// V contains metadata about all signatures on the loaded trust, election, and
// technical configuration.
type V struct {
	Trust     Signatures
	Election  Signatures
	Technical Signatures
}

// Signatures is a slice of container signatures with custom JSON formatting.
//
// We define the type on the entire slice, although custom formatting is only
// applied to the single signatures, but this way when converting a slice of
// container signatures, we do not need to create a new slice and can just cast
// the entire slice.
type Signatures []container.Signature

// MarshalJSON marshals the signatures as a JSON array of "<CN> <timestamp>".
func (s Signatures) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	b.WriteByte('[')
	for i, c := range s {
		if i > 0 {
			b.WriteByte(',')
		}
		if err := e.Encode(fmt.Sprint(
			c.Signer.Subject.CommonName,
			" ",
			c.SigningTime.Format(time.RFC3339),
		)); err != nil {
			return nil, MarshalSignatureError{Err: err}
		}
	}
	b.WriteByte(']')
	return b.Bytes(), nil
}

// Container returns a container's version string.
func Container(c container.Container) (version string, err error) {
	b, err := json.Marshal(Signatures(c.Signatures()))
	version = string(b)
	if err != nil {
		err = ContainerError{Err: err}
	}
	return
}
