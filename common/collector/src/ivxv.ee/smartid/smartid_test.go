package smartid

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"ivxv.ee/log"
	"ivxv.ee/ocsp"
)

const (
	testIdentifier = "30303039914"
	testDocumentNo = "PNOEE-30303039914-MOCK-Q"
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
		URL:              "https://sid.demo.sk.ee/smart-id-rp/v2/",
		RelyingPartyUUID: "00000000-0000-0000-0000-000000000000",
		RelyingPartyName: "DEMO",
		CertificateLevel: "QUALIFIED",
		AuthInteractionsOrder: []allowedInteractionsOrder{{
			Type:          "displayTextAndPIN",
			DisplayText60: "authenticating",
		}},
		SignInteractionsOrder: []allowedInteractionsOrder{{
			Type:          "displayTextAndPIN",
			DisplayText60: "signing",
		}},
		AuthChallengeSize: 64,
		Roots:             []string{read("testdata/TEST_of_EE_Certification_Centre_Root_CA.pem")},
		Intermediates:     []string{read("testdata/TEST_of_EID-SK_2016.pem")},
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
		t.Skip("Short mode on, skipping test against test-SmartID-REST service")
	}
	t.Parallel()

	ctx := log.TestContext(context.Background())
	code, challengeRnd, _, err := testClient.Authenticate(ctx, testIdentifier)
	if err != nil {
		t.Fatal("failed to start authentication session:", err)
	}

	var cert *x509.Certificate
	var algo string
	var signature []byte
	for len(signature) == 0 {
		time.Sleep(1 * time.Second)
		fmt.Println("polling authentication") // Use fmt instead of t.Log to get running output.
		if cert, algo, signature, err = testClient.GetAuthenticateStatus(ctx, code); err != nil {
			t.Fatal("failed to check authentication status:", err)
		}
	}
	if !strings.Contains(cert.Subject.SerialNumber, testIdentifier) {
		t.Error("unexpected subject serial number:", cert.Subject.SerialNumber)
	}

	if err = VerifyAuthenticationSignature(cert, algo, challengeRnd, signature); err != nil {
		t.Error("failed to verify authentication challenge signature:", err)
	}
}

func TestCertificate(t *testing.T) {
	if testing.Short() {
		t.Skip("Short mode on, skipping test against test-SmartID-REST service")
	}
	t.Parallel()

	ctx := log.TestContext(context.Background())
	code, err := testClient.GetCertificateChoice(ctx, testIdentifier)
	if err != nil {
		t.Fatal("failed to get certificate choice:", err)
	}

	var cert *x509.Certificate
	var documentno string
	for cert == nil {
		time.Sleep(1 * time.Second)
		fmt.Println("polling certificate choice") // Use fmt instead of t.Log to get running output.
		if documentno, cert, err = testClient.GetCertificateChoiceStatus(ctx, code); err != nil {
			t.Fatal("failed to check certificate choice status:", err)
		}
	}
	if documentno != testDocumentNo {
		t.Fatal("different document numbers:", documentno, testDocumentNo)
	}
}

func TestSigning(t *testing.T) {
	if testing.Short() {
		t.Skip("Short mode on, skipping test against test-MID-REST service")
	}
	t.Parallel()

	ctx := log.TestContext(context.Background())

	code, err := testClient.GetCertificateChoice(ctx, testIdentifier)
	if err != nil {
		t.Fatal("failed to get certificate choice:", err)
	}

	var cert *x509.Certificate
	var documentno string
	for cert == nil {
		time.Sleep(1 * time.Second)
		fmt.Println("polling certificate choice") // Use fmt instead of t.Log to get running output.
		if documentno, cert, err = testClient.GetCertificateChoiceStatus(ctx, code); err != nil {
			t.Fatal("failed to check certificate choice status:", err)
		}
	}
	if documentno != testDocumentNo {
		t.Fatal("different document numbers:", documentno, testDocumentNo)
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

	code, err = testClient.SignHash(ctx, testDocumentNo, hash, hashFunctionNames[hashFun])
	if err != nil {
		t.Fatal("failed to start signing session:", err)
	}

	var algo string
	var signature []byte
	for len(signature) == 0 {
		time.Sleep(1 * time.Second)
		fmt.Println("polling signing") // Use fmt instead of t.Log to get running output.
		if algo, signature, err = testClient.GetSignHashStatus(ctx, code); err != nil {
			t.Fatal("failed to check signing status:", err)
		}
	}

	// Although VerifyAuthenticationSignature is meant for authentication
	// (as the name suggests), we can reuse it in this test case.
	if err = VerifyAuthenticationSignature(cert, algo, rnd, signature); err != nil {
		t.Error("failed to verify authentication challenge signature:", err)
	}
}
