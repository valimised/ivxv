package dummy

import (
	"strings"
	"testing"
	"time"

	"ivxv.ee/errors"
)

const (
	dataKey   = "signed data"
	dataValue = `this is some data that is "signed" in the container`
	signerCN  = "dummy container test certificate"
	issuerCN  = "dummy container test issuer"

	encoded = `
signatures:
  - signer: |
      -----BEGIN CERTIFICATE-----
      MIIBnjCCAUOgAwIBAgIJAKztlcHSEoWWMAoGCCqGSM49BAMCMCsxKTAnBgNVBAMM
      IGR1bW15IGNvbnRhaW5lciB0ZXN0IGNlcnRpZmljYXRlMB4XDTE2MTAwNzA4NTA1
      MloXDTI2MDgxNjA4NTA1MlowKzEpMCcGA1UEAwwgZHVtbXkgY29udGFpbmVyIHRl
      c3QgY2VydGlmaWNhdGUwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAToQ4OuVhmz
      69qimYxoyVNmgRzFd1LJJtsm6wWeWpaoFa5yl+GR0Lw53RK2I2WcZ8zwdKtusOmf
      eJB50hNXJaBfo1AwTjAdBgNVHQ4EFgQU6FDwDvLpx5CiJwiJRLA8zLlMPGswHwYD
      VR0jBBgwFoAU6FDwDvLpx5CiJwiJRLA8zLlMPGswDAYDVR0TBAUwAwEB/zAKBggq
      hkjOPQQDAgNJADBGAiEA/gbSkL0VquQWMvlogaI/GEY2XT/lpXKIdakCf/Qjg24C
      IQCsBZbv02YVVnVmGJ4wCw6sSzuQcBxjGlDULea3JrWMFA==
      -----END CERTIFICATE-----
    issuer: |
      -----BEGIN CERTIFICATE-----
      MIIBlzCCAT6gAwIBAgIJAKMVxT8JoMPDMAoGCCqGSM49BAMCMCYxJDAiBgNVBAMM
      G2R1bW15IGNvbnRhaW5lciB0ZXN0IGlzc3VlcjAgFw0xNzA4MTUxNTEyMThaGA8y
      MTE3MDcyMjE1MTIxOFowJjEkMCIGA1UEAwwbZHVtbXkgY29udGFpbmVyIHRlc3Qg
      aXNzdWVyMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAExDNTim5vmEt3/I1Ii8lC
      chSOYAQzQIY88U1cPcMEHqpcll5UZmoxBKuv0X8rRECUK3z6vXK65djcJF4IWUiV
      dKNTMFEwHQYDVR0OBBYEFAQVpWuMNqCeLXVLlinwXftO0VmGMB8GA1UdIwQYMBaA
      FAQVpWuMNqCeLXVLlinwXftO0VmGMA8GA1UdEwEB/wQFMAMBAf8wCgYIKoZIzj0E
      AwIDRwAwRAIgVQ4tzmXu8UX/0+9zgob+SvDroz7fvWPeh0iauBWMx5wCIGHIdyjL
      2X0l/NsePltPkvwTOfqVUcW9hbDQnuvhGVSO
      -----END CERTIFICATE-----
    signingtime: 2017-02-28T11:30:20Z
data:
  ` + dataKey + ": " + dataValue
)

var signingTime = time.Date(2017, time.February, 28, 11, 30, 20, 0, time.UTC)

func TestDummy(t *testing.T) {
	c, err := new(Conf).Open(strings.NewReader(encoded))
	if err != nil {
		t.Fatal("parsing dummy container failed:", err)
	}

	if s := c.Signatures(); len(s) != 1 {
		t.Error("unexpected signatures count:", len(s))
	} else if cn := s[0].Signer.Subject.CommonName; cn != signerCN {
		t.Error("unexpected signer CN value:", cn)
	} else if is := s[0].Issuer.Subject.CommonName; is != issuerCN {
		t.Error("unexpected issuer CN value:", cn)
	} else if st := s[0].SigningTime; !st.Equal(signingTime) {
		t.Error("unexpected signing time value:", st)
	}

	d := c.Data()
	if len(d) != 1 {
		t.Error("unexpected data key count:", len(d))
	}
	if v, ok := d[dataKey]; !ok {
		t.Errorf("missing data key %q", dataKey)
	} else if string(v) != dataValue {
		t.Errorf("unexpected data value of key %q: %s", dataKey, v)
	}
}

func TestTrusted(t *testing.T) {
	for _, test := range []struct {
		name  string
		conf  *Conf
		cause error
	}{
		{"unconfigured", new(Conf), nil},
		{"trusted", &Conf{Trusted: []string{signerCN}}, nil},
		{"untrusted", &Conf{Trusted: []string{"other signer"}}, new(UntrustedSignerError)},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.conf.Open(strings.NewReader(encoded))
			if err != test.cause && errors.CausedBy(err, test.cause) == nil {
				t.Errorf("unexpected error: %v, want cause %T", err, test.cause)
			}
		})
	}
}
