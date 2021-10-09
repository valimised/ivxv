package mid

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/x509"
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
	testID    = "60001019906"
	testPhone = "+37200000766"
)

var (
	hashFun    = crypto.SHA384
	testClient *Client
)

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
		URL:               "https://tsp.demo.sk.ee/mid-api/",
		RelyingPartyUUID:  "00000000-0000-0000-0000-000000000000",
		RelyingPartyName:  "DEMO",
		Language:          "EST",
		AuthMessage:       "Test authentication message.",
		SignMessage:       "Test signing message.",
		MessageFormat:     "GSM-7",
		AuthChallengeSize: 64,
		Roots:             []string{read("testdata/TEST_of_EE_Certification_Centre_Root_CA.pem")},
		Intermediates:     []string{read("testdata/TEST_of_ESTEID-SK_2015.pem")},
		OCSP: ocsp.Conf{
			URL:        "http://demo.sk.ee/ocsp",
			Responders: []string{read("testdata/TEST_of_SK_OCSP_RESPONDER_2020.pem")},
		},
	}); err != nil {
		fmt.Fprintln(os.Stderr, "failed to create test client:", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestAuthentication(t *testing.T) {
	if testing.Short() {
		t.Skip("Short mode on, skipping test against test-MID-REST service")
	}
	t.Parallel()

	ctx := log.TestContext(context.Background())
	code, challengeRnd, _, err := testClient.MobileAuthenticate(ctx, testID, testPhone)
	if err != nil {
		t.Fatal("failed to start authentication session:", err)
	}

	var cert *x509.Certificate
	var algo string
	var signature []byte
	for len(signature) == 0 {
		time.Sleep(1 * time.Second)
		fmt.Println("polling authentication") // Use fmt instead of t.Log to get running output.
		if cert, algo, signature, err = testClient.GetMobileAuthenticateStatus(ctx, code); err != nil {
			t.Fatal("failed to check authentication status:", err)
		}
	}

	if cert.Subject.SerialNumber != testID {
		t.Error("unexpected subject serial number:", cert.Subject.SerialNumber)
	}

	if err = VerifyAuthenticationSignature(cert, algo, challengeRnd, signature); err != nil {
		t.Error("failed to verify authentication challenge signature:", err)
	}
}

func TestCertificate(t *testing.T) {
	if testing.Short() {
		t.Skip("Short mode on, skipping test against test-MID-REST service")
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
		t.Skip("Short mode on, skipping test against test-MID-REST service")
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

	rnd := make([]byte, 100)
	if _, err = rand.Read(rnd); err != nil {
		t.Fatal("failed to generate data to sign:", err)
	}
	h := hashFun.New()
	if _, err = h.Write(rnd); err != nil {
		t.Fatal("failed to hash data to sign:", err)
	}
	hash := h.Sum(nil)

	ctx := log.TestContext(context.Background())
	code, err := testClient.MobileSignHash(ctx, testID, testPhone, hash, hashFunctionNames[hashFun])
	if err != nil {
		t.Fatal("failed to start signing session:", err)
	}

	var algo string
	var signature []byte
	for len(signature) == 0 {
		time.Sleep(1 * time.Second)
		fmt.Println("polling signing") // Use fmt instead of t.Log to get running output.
		if algo, signature, err = testClient.GetMobileSignHashStatus(ctx, code); err != nil {
			t.Fatal("failed to check signing status:", err)
		}
	}

	// Although VerifyAuthenticationSignature is meant for authentication
	// (as the name suggests), we can reuse it in this test case.
	if err = VerifyAuthenticationSignature(cert, algo, rnd, signature); err != nil {
		t.Error("failed to verify authentication challenge signature:", err)
	}
}
