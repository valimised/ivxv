package ocsp

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ivxv.ee/common/collector/cryptoutil"
	"ivxv.ee/common/collector/log"
)

const (
	demoSkEeOCSPResponderIssuer = "TEST_of_KLASS3-SK_2016.pem"
	demoSkEeOCSPResponder       = "DEMO_of_KLASS3-SK_2016_SSL_OCSP_RESPONDER_2018.pem"
	demoSkEeOcspURL             = "http://demo.sk.ee/ocsp"
	// > 2022 year ID-cards' issuer
	esteid2018 = "esteid2018.pem"
	// <= 2022 year ID-cards'/Mobile-IDs' issuer
	testOfEsteidsk2015 = "TEST_of_ESTEID-SK_2015.pem"
	esteidsk2015       = "esteidsk2015.pem"
	// Smart-IDs'/Mobile-IDs' issuer
	eidsk2016 = "eidsk2016.pem"
)

func TestCheckOCSPOfIDCardWithParsingOCSPURLFromConf(t *testing.T) {
	responderPath := demoSkEeOCSPResponder
	responderBytes, err := os.ReadFile(filepath.Join("testdata", responderPath))
	if err != nil {
		t.Fatal(err)
	}

	client, err := New(&Conf{
		URL:        demoSkEeOcspURL,
		Responders: []string{string(responderBytes)},
		Retry:      2,
		MaxSkew:    300,
		MaxAge:     1,
	})
	if err != nil {
		t.Errorf("Create new OCSP config error: %v\n", err)
	}

	ctx := log.TestContext(context.Background())
	certPath := "good.pem"
	cert, err := testCert(certPath)
	if err != nil {
		t.Errorf("Obtain %v\n", err)
	}

	// Since issuer is nil, then we assume that responders will
	// verify the OCSP response
	check, err := client.Check(ctx, cert, nil, nil)
	if err != nil {
		t.Errorf("Error %v\n", err)
	}

	if !check.Good {
		t.Errorf("Certificate status is not good: %v\n", check.CertStatus)
	}
}

func TestCheckOCSPWithParsingOCSPURLFromCert(t *testing.T) {
	testName := fmt.Sprintf("Issuer-%s-cert-%s", esteid2018, "ID-PNOEE-39605244244.pem")
	t.Run(testName, func(t *testing.T) {
		responderIssuerBytes, err := os.ReadFile(filepath.Join("testdata", esteid2018))
		if err != nil {
			t.Fatal(err)
		}
		responderIssuer, err := cryptoutil.PEMCertificate(string(responderIssuerBytes))
		if err != nil {
			t.Fatal(err)
		}

		client, err := New(&Conf{
			// No URL defined forces to parse it from client cert
			Retry:   2,
			MaxSkew: 300,
			MaxAge:  1,
		})
		if err != nil {
			t.Errorf("Create new OCSP config error: %v\n", err)
		}

		ctx := log.TestContext(context.Background())
		cert2, err := testCert("ID-PNOEE-39605244244.pem")
		if err != nil {
			t.Errorf("Obtain %v\n", err)
		}

		check, err := client.Check(ctx, cert2, responderIssuer, nil)
		if err != nil {
			t.Errorf("Error %v\n", err)
		}

		if !check.Good {
			t.Errorf("Certificate status is not good: %v\n", check.CertStatus)
		}
	})
}

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
		{"http://demo.sk.ee/ocsp", []string{demoSkEeOCSPResponder}, []cert{
			{"good.pem", "", good},
			{"revoked.pem", "", revoked},
			{"unknown.pem", "", unknown},
		}},
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
			pem, err := os.ReadFile(filepath.Join("testdata", r))
			if err != nil {
				t.Fatal("failed to read responder certificate:", err)
			}
			pems = append(pems, string(pem))
		}

		client, err := New(&Conf{
			URL:        url,
			Responders: pems,
			Retry:      2,
			MaxSkew:    300,
			MaxAge:     1,
		})
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
	responder, err := os.ReadFile(
		filepath.Join("testdata", demoSkEeOCSPResponder))
	if err != nil {
		t.Fatal("failed to read responder certificate:", err)
	}

	client, err := New(&Conf{
		Responders: []string{string(responder)}})
	if err != nil {
		t.Fatal("failed to create client:", err)
	}

	cert, err := testCert("good.pem")
	if err != nil {
		t.Fatal("failed to parse certificate:", err)
	}

	resp, err := os.ReadFile("testdata/test_response")
	if err != nil {
		t.Fatal("failed to read stored response:", err)
	}

	nonce, err := os.ReadFile("testdata/test_nonce")
	if err != nil {
		t.Fatal("failed to read stored nonce:", err)
	}

	status, err := client.CheckResponse(resp, cert, nil, nonce, time.Time{})
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
	data, err := os.ReadFile(filepath.Join("testdata", cert))
	if err != nil {
		return nil, err
	}

	return cryptoutil.PEMCertificate(string(data))
}
