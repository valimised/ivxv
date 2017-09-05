//ivxv:development

/*
Package dummy implement a dummy authentication verifier user for testing.

The dummy authentication verifier expects the authentication token to be a
DER-encoded Distinguished Name and it just returns that name. The dummy
implementation can also be optionally configured with a limited set of names
for which authentication succeeds and all other names will be reported as
unauthenticated.
*/
package dummy // import "ivxv.ee/auth/dummy"

import (
	"context"
	"crypto/x509/pkix"
	"encoding/asn1"

	"ivxv.ee/auth"
	"ivxv.ee/yaml"
)

func init() {
	auth.Register(auth.Dummy, func(n yaml.Node) (auth.Verifier, error) {
		c := new(Conf)
		if err := yaml.Apply(n, c); err != nil {
			return nil, ConfigurationError{Err: err}
		}
		return c, nil
	})
}

// Conf is the dummy authentication verifier configuration.
type Conf struct {
	// Authenticated is a slice of Common Names. If not empty, then only
	// names included in this list will be authenticated.
	Authenticated []string
}

// Verify verifies a dummy authentication token.
func (c *Conf) Verify(_ context.Context, token []byte) (name *pkix.Name, err error) {
	var dn pkix.RDNSequence
	if _, err := asn1.Unmarshal(token, &dn); err != nil {
		return nil, auth.MalformedTokenError{Err: UnmarshalTokenError{Err: err}}
	}
	name = new(pkix.Name)
	name.FillFromRDNSequence(&dn)
	if len(c.Authenticated) > 0 {
		for _, cn := range c.Authenticated {
			if cn == name.CommonName {
				return
			}
		}
		return nil, auth.UnauthorizedError{Err: NotAllowedError{Name: name}}
	}
	return
}
