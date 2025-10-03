/*
Package tsp contains qualification protocols which perform PKIX timestamping
requests.

tsp registers the following qualifiers:

  - tsp, which requests a timestamp on the signature (there must be exactly
    one signature on the signed container and containers must implement
    TimestampDataer) and returns the timestamp token as the qualifying
    property, and

  - tspreg, which does the same as tsp, but uses a signature on the message
    imprint of the request as the nonce. This is an ad-hoc solution
    for signing timestamp protocol requests so that they can be used
    as registration requests. The signing key is a PEM-encoded RSA
    private key in the service directory with the name "tspreg.key".
*/
package tsp

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"os"
	"path/filepath"

	"ivxv.ee/common/collector/container"
	"ivxv.ee/common/collector/cryptoutil"
	"ivxv.ee/common/collector/q11n"
	"ivxv.ee/common/collector/tsp"
	"ivxv.ee/common/collector/yaml"
)

func init() {
	q11n.Register(q11n.TSP, newreg(false), tsp.ParseTime)
	q11n.Register(q11n.TSPREG, newreg(true), tsp.ParseTime)
}

type client struct {
	tsp *tsp.Client
	key *rsa.PrivateKey
}

func newreg(reg bool) func(yaml.Node, string) (q11n.Qualifier, error) {
	return func(n yaml.Node, sensitive string) (q q11n.Qualifier, err error) {
		var conf tsp.Conf
		if err = yaml.Apply(n, &conf); err != nil {
			return nil, YamlApplyError{Err: err,
				Description: _TSA_CFG}
		}

		c := new(client)
		if c.tsp, err = tsp.New(&conf); err != nil {
			return nil, ClientError{Err: err,
				Description: _TSA_NEW}
		}

		if reg {
			pem, err := os.ReadFile(filepath.Join(sensitive, "tspreg.key"))
			if err != nil {
				return nil, ReadPrivateKeyError{Err: err,
					Description: _TSA_KEY}
			}
			der, err := cryptoutil.PEMDecode(string(pem), "RSA PRIVATE KEY")
			if err != nil {
				return nil, DecodePrivateKeyError{Err: err,
					Description: _TSA_KEY_PARSE}
			}
			if c.key, err = x509.ParsePKCS1PrivateKey(der); err != nil {
				return nil, ParsePrivateKeyError{Err: err,
					Description: _TSA_KEY_PARSE_RSA}
			}
		}

		return c, nil
	}
}

func (c *client) Qualify(ctx context.Context, container container.Container) ([]byte, error) {
	sigs := container.Signatures()
	if len(sigs) != 1 {
		return nil, NoSingleSignatureError{Count: len(sigs),
			Description: _TSA_NO_SIG}
	}
	id := sigs[0].ID

	dataer, ok := container.(TimestampDataer)
	if !ok {
		return nil, ContainerNotTimestampDataerError{
			Description: _TSA_DATAER,
		}
	}
	data, err := dataer.TimestampData(id)
	if err != nil {
		return nil, TimestampDataError{ID: id, Err: err,
			Description: _TSA_SIG_VERIFY}
	}

	var nonce []byte
	if c.key != nil {
		nonce, err = c.sign(data)
		if err != nil {
			return nil, SignTimestampDataError{Err: err,
				Description: _TSA_SIG_TS}
		}
	}

	resp, err := c.tsp.Create(ctx, data, nonce)
	if err != nil {
		return nil, CreateTimestampError{Err: err,
			Description: _TSA_REQ}
	}
	return resp, nil
}

func (c *client) sign(data []byte) (signature []byte, err error) {
	// The hash function used here must match the one used in ivxv.ee/common/collector/tsp
	// to create the message imprint.
	hash := sha256.Sum256(data)
	signature, err = rsa.SignPKCS1v15(nil, c.key, crypto.SHA256, hash[:])
	if err != nil {
		return nil, SignHashError{Err: err,
			Description: _TSA_RSA_SIG}
	}

	return asn1.Marshal(struct {
		SignatureAlgorithm pkix.AlgorithmIdentifier
		Signature          []byte
	}{
		pkix.AlgorithmIdentifier{
			// sha256WithRSAEncryption
			Algorithm: asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 11},
		},
		signature,
	})
}

// TimestampDataer returns the data to be timestamped for the signature with
// the given ID.
type TimestampDataer interface {
	TimestampData(id string) ([]byte, error)
}
