package ocsp

import (
	"context"
	"crypto/x509"
	"io/ioutil"
	"path/filepath"
	"testing"

	"ivxv.ee/cryptoutil"
	"ivxv.ee/log"
)

func TestCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Short mode on, skipping OCSP test against live responder")
	}

	// Enumeration of certificate statuses to test for.
	const (
		good = iota
		revoked
		unknown
	)

	// Structure for single certificate test case.
	type cert struct {
		name   string
		issuer string
		status int
	}

	tests := []struct {
		url        string
		responders []string
		certs      []cert
	}{
		{"http://demo.sk.ee/ocsp", []string{"TEST_of_SK_OCSP_RESPONDER_2011.pem"}, []cert{
			{"good.pem", "", good},
			{"revoked.pem", "", revoked},
			{"unknown.pem", "", unknown},
		}},

		// The following tests require actual certificates, which we do
		// not want to include the repository. Provide the certificates
		// and uncomment to run these tests.
		//
		//{"http://aia.sk.ee/esteid2015", nil, []cert{
		//	{"auth2011.pem", "ESTEID-SK_2011.pem", revoked},
		//	{"sign2011.pem", "ESTEID-SK_2011.pem", revoked},
		//	{"auth2015.pem", "ESTEID-SK_2015.pem", good},
		//	{"sign2015.pem", "ESTEID-SK_2015.pem", good},
		//}},
	}

	// Define test functions for a single certificate and for a client
	// beforehand, so that our test loop in the end does not nest too deep.

	ctx := log.TestContext(context.Background())
	testSingle := func(t *testing.T, client *Client, cert, issuer string, status int) {
		t.Parallel()

		c, err := testCert(cert)
		if err != nil {
			t.Fatal("failed to parse certificate:", err)
		}

		var i *x509.Certificate
		if len(issuer) > 0 {
			if i, err = testCert(issuer); err != nil {
				t.Fatal("failed to parse certificate issuer:", err)
			}
		}

		resp, err := client.Check(ctx, c, i, nil)
		if err != nil {
			t.Fatal("failed to check certificate status:", err)
		}

		switch {
		case status == good && !resp.Good:
			fallthrough
		case status == revoked && (resp.Good || resp.Unknown):
			fallthrough
		case status == unknown && !resp.Unknown:
			t.Errorf("unexpected status, good: %t, reason: %d, unknown: %t",
				resp.Good, resp.RevocationReason, resp.Unknown)
		}
	}

	testClient := func(t *testing.T, url string, responders []string, certs []cert) {
		t.Parallel()

		var pems []string
		for _, r := range responders {
			pem, err := ioutil.ReadFile(filepath.Join("testdata", r))
			if err != nil {
				t.Fatal("failed to read responder certificate:", err)
			}
			pems = append(pems, string(pem))
		}

		client, err := New(&Conf{URL: url, Responders: pems})
		if err != nil {
			t.Fatal("failed to create client:", err)
		}

		for _, cert := range certs {
			t.Run(cert.name, func(t *testing.T) {
				testSingle(t, client, cert.name, cert.issuer, cert.status)
			})
		}
	}

	for _, test := range tests {
		t.Run(test.url, func(t *testing.T) {
			testClient(t, test.url, test.responders, test.certs)
		})
	}
}

func TestCheckResponse(t *testing.T) {
	responder, err := ioutil.ReadFile(
		filepath.Join("testdata", "TEST_of_SK_OCSP_RESPONDER_2011.pem"))
	if err != nil {
		t.Fatal("failed to read responder certificate:", err)
	}

	client, err := New(&Conf{Responders: []string{string(responder)}})
	if err != nil {
		t.Fatal("failed to create client:", err)
	}

	cert, err := testCert("good.pem")
	if err != nil {
		t.Fatal("failed to parse certificate:", err)
	}

	resp, err := ioutil.ReadFile("testdata/test_response")
	if err != nil {
		t.Fatal("failed to read stored response:", err)
	}

	nonce, err := ioutil.ReadFile("testdata/test_nonce")
	if err != nil {
		t.Fatal("failed to read stored nonce:", err)
	}

	status, err := client.CheckResponse(resp, cert, nil, nonce)
	if err != nil {
		t.Fatal("failed to check response:", err)
	}

	switch {
	case status.Unknown:
		t.Error("certificate status unknown")
	case !status.Good:
		t.Error("certificate revoked, reason:", status.RevocationReason)
	}
}

func testCert(cert string) (*x509.Certificate, error) {
	data, err := ioutil.ReadFile(filepath.Join("testdata", cert))
	if err != nil {
		return nil, err
	}

	return cryptoutil.PEMCertificate(string(data))
}
