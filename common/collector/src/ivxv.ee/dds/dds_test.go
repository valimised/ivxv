package dds

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"ivxv.ee/cryptoutil"
	"ivxv.ee/log"
	"ivxv.ee/ocsp"
)

const (
	testID    = "11412090004"
	testPhone = "+37200000766"
)

var testClient *Client

func TestMain(m *testing.M) {
	read := func(path string) string {
		b, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to read file:", err)
			os.Exit(1)
		}
		return string(b)
	}

	var err error
	if testClient, err = New(&Conf{
		URL:           "https://tsp.demo.sk.ee/v2/",
		Language:      "EST",
		ServiceName:   "Testimine",
		AuthMessage:   "Test authentication message.",
		SignMessage:   "Test signing message.",
		Roots:         []string{read("testdata/TEST_of_EE_Certification_Centre_Root_CA.pem")},
		Intermediates: []string{read("testdata/TEST_of_ESTEID-SK_2011.pem")},
		OCSP: ocsp.Conf{
			Responders: []string{read("testdata/TEST_of_SK_OCSP_RESPONDER_2011.pem")},
		},
	}); err != nil {
		fmt.Fprintln(os.Stderr, "failed to create test client:", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestAuthentication(t *testing.T) {
	if testing.Short() {
		t.Skip("Short mode on, skipping test against test-DigiDocService")
	}
	t.Parallel()

	ctx := log.TestContext(context.Background())
	code, _, challenge, cert, err := testClient.MobileAuthenticate(ctx, testPhone)
	if err != nil {
		t.Fatal("failed to start authentication session:", err)
	}

	if cert.Subject.SerialNumber != testID {
		t.Error("unexpected subject serial number:", cert.Subject.SerialNumber)
	}

	var signature []byte
	for len(signature) == 0 {
		time.Sleep(1 * time.Second)
		fmt.Println("polling authentication") // Use fmt instead of t.Log to get running output.
		if signature, err = testClient.GetMobileAuthenticateStatus(ctx, code); err != nil {
			t.Fatal("failed to check authentication status:", err)
		}
	}

	if err = VerifyAuthenticationSignature(cert, challenge, signature); err != nil {
		t.Error("failed to verify authentication challenge signature:", err)
	}
}

func TestCertificate(t *testing.T) {
	if testing.Short() {
		t.Skip("Short mode on, skipping test against test-DigiDocService")
	}
	t.Parallel()

	ctx := log.TestContext(context.Background())
	cert, err := testClient.GetMobileCertificate(ctx, testID, testPhone)
	if err != nil {
		t.Fatal("failed to get signing certificate:", err)
	}

	if cert.Subject.SerialNumber != testID {
		t.Error("unexpected subject serial number:", cert.Subject.SerialNumber)
	}
}

func TestSigning(t *testing.T) {
	if testing.Short() {
		t.Skip("Short mode on, skipping test against test-DigiDocService")
	}
	t.Parallel()

	pem, err := ioutil.ReadFile("testdata/signer.pem")
	if err != nil {
		t.Fatal("failed to read signer certificate:", err)
	}
	cert, err := cryptoutil.PEMCertificate(string(pem))
	if err != nil {
		t.Fatal("failed to parse signer certificate:", err)
	}

	hash := make([]byte, sha256.Size)
	if _, err = rand.Read(hash); err != nil {
		t.Fatal("failed to generate data to sign:", err)
	}

	ctx := log.TestContext(context.Background())
	code, _, err := testClient.MobileSignHash(ctx, testID, testPhone, hash)
	if err != nil {
		t.Fatal("failed to start signing session:", err)
	}

	var signature []byte
	for len(signature) == 0 {
		time.Sleep(1 * time.Second)
		fmt.Println("polling signing") // Use fmt instead of t.Log to get running output.
		if signature, err = testClient.GetMobileSignHashStatus(ctx, code); err != nil {
			t.Fatal("failed to check signing status:", err)
		}
	}

	// Although VerifyAuthenticationSignature is meant for authentication
	// (as the name suggests), we can reuse it in this test case.
	if err = VerifyAuthenticationSignature(cert, hash, signature); err != nil {
		t.Error("failed to verify authentication challenge signature:", err)
	}
}
