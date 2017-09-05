//ivxv:development

/*
Package dummy implements a dummy container used for testing.

The dummy container is just a YAML-encoding of the values that should be
returned by the interface methods:

	signatures:
	  - signer:      <PEM-encoding of signer certificate>
	    issuer:      <PEM-encoding of issuer certificate>
	    signingtime: <RFC3339-formatted signing time>
	data:
	  <key>: <string>

If issuer is omitted, then it is assumed that signer is self-signed and will
also be used as the issuer.

If signingtime is omitted, then 1970-01-01T00:00:00Z will be used.

The string value of a data key can be wrapped with "base64(" and ")" in which
case the contents will be Base64 decoded during parsing.
*/
package dummy // import "ivxv.ee/container/dummy"

import (
	"crypto/x509"
	"encoding/base64"
	"io"
	"strconv"
	"strings"
	"time"

	"ivxv.ee/container"
	"ivxv.ee/cryptoutil"
	"ivxv.ee/yaml"
)

func init() {
	container.Register(container.Dummy,
		func(n yaml.Node) (container.OpenFunc, error) {
			var c Conf
			if err := yaml.Apply(n, &c); err != nil {
				return nil, ConfigurationError{Err: err}
			}
			return c.Open, nil
		},
		func(encoded io.Reader) (container.Container, error) {
			return new(Conf).Open(encoded)
		})
}

// unverifiedOpen opens a dummy signature container and returns data from it
// without verifying any signatures (which dummy does anyway).
//
// Unexported function, because nobody should use this except during
// bootstrapping via container.UnverifiedOpen.

// Conf is the dummy container opener configuration.
type Conf struct {
	// Trusted is a slice of Common Names. If not empty, then all signers
	// of dummy containers opened with this configuration must be included
	// in Trusted.
	Trusted []string
}

// Open opens a dummy signature container and returns a container.Container for
// accessing it.
func (c *Conf) Open(encoded io.Reader) (cnt container.Container, err error) {
	// Unmarshal the YAML.
	var y struct {
		Signatures []struct {
			Signer      string
			Issuer      string
			SigningTime string
		}
		Data map[string]string
	}
	if err = yaml.Unmarshal(encoded, nil, &y); err != nil {
		return nil, UnmarshalContainerError{Err: err}
	}

	// Parse the signatures into container.Signatures.
	d := &dummy{data: make(map[string][]byte)}
	for i, s := range y.Signatures {
		id := strconv.Itoa(i)

		var cert *x509.Certificate
		if cert, err = cryptoutil.PEMCertificate(s.Signer); err != nil {
			return nil, ParseSignerError{ID: id, Certificate: s.Signer, Err: err}
		}

		// Check that the signer is trusted.
		if !c.trusted(cert.Subject.CommonName) {
			return nil, UntrustedSignerError{ID: id, Signer: cert.Subject.CommonName}
		}

		issuer := cert
		if len(s.Issuer) > 0 {
			if issuer, err = cryptoutil.PEMCertificate(s.Issuer); err != nil {
				return nil, ParseIssuerError{ID: id, Certificate: s.Issuer, Err: err}
			}
		}

		t := time.Date(1970, time.January, 01, 00, 00, 00, 00, time.UTC)
		if len(s.SigningTime) > 0 {
			t, err = time.Parse(time.RFC3339, s.SigningTime)
			if err != nil {
				return nil, ParseSigningTimeError{ID: id, Time: s.SigningTime, Err: err}
			}
		}

		d.signatures = append(d.signatures, container.Signature{
			ID:          id,
			Signer:      cert,
			Issuer:      issuer,
			SigningTime: t,
		})
	}

	// Decode the data.
	for k, s := range y.Data {
		if d.data[k], err = decode(s); err != nil {
			return nil, DecodeValueError{Key: k, Err: err}
		}
	}

	return d, nil
}

func (c *Conf) trusted(cn string) bool {
	if len(c.Trusted) == 0 {
		return true
	}
	for _, t := range c.Trusted {
		if t == cn {
			return true
		}
	}
	return false
}

// decode checks if s has a "base64(" prefix and a ")" suffix and Base64
// decodes the inner content. If the prefix or suffix is not found, then s is
// returned unmodified.
func decode(s string) ([]byte, error) {
	const prefix, suffix = "base64(", ")"
	t := strings.TrimSpace(s)
	if strings.HasPrefix(t, prefix) && strings.HasSuffix(t, suffix) {
		return base64.StdEncoding.DecodeString(t[len(prefix) : len(t)-len(suffix)])
	}
	return []byte(s), nil
}

// dummy is a parsed dummy Container.
type dummy struct {
	signatures []container.Signature
	data       map[string][]byte
}

func (d *dummy) Signatures() []container.Signature {
	return d.signatures
}

func (d *dummy) Data() map[string][]byte {
	return d.data
}

func (d *dummy) Close() error {
	return nil
}

// SignatureValue implements the ivxv.ee/q11n/ocsp.SignatureValuer interface.
// It checks if the signature ID is valid, but then returns a constant since
// dummy containers do not have actual signatures.
func (d *dummy) SignatureValue(id string) ([]byte, error) {
	return []byte("signature value"), d.checkID(id)
}

// TimestampData implements the ivxv.ee/q11n/tsp.TimestampDataer interface. It
// checks if the signature ID is valid, but then returns a constant since dummy
// containers do not have actual signatures to timestamp.
func (d *dummy) TimestampData(id string) ([]byte, error) {
	return []byte("timestamp data"), d.checkID(id)
}

func (d *dummy) checkID(id string) error {
	// id must be a valid index.
	i, err := strconv.Atoi(id)
	if err != nil {
		return SignatureIDParseError{ID: id, Err: err}
	}
	if i < 0 || i >= len(d.signatures) {
		return NoSuchIDError{ID: id}
	}
	return nil
}
