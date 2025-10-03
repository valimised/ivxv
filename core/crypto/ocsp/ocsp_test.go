package ocsp

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"tivi.io/core/crypto/util"
)

const testURL = "http://demo.sk.ee/ocsp"

var testCertGood, testCertRevoked, testCertUnknown, testIssuer, testResponder *x509.Certificate

const (
	good = iota
	revoked
	unknown
)

func init() {
	pem, err := ioutil.ReadFile("./testdata/TESTCERT_GOOD_12345678901-sign.pem")
	if err != nil {
		fmt.Fprintln(os.Stderr, "read PEM file:", err)
		os.Exit(1)
	}
	testCertGood, err = util.GetPEMCertificate(pem)
	if err != nil {
		fmt.Fprintln(os.Stderr, "get PEM certificate: %w", err)
		os.Exit(1)
	}
	pem, err = ioutil.ReadFile("./testdata/TESTCERT_REVOKED_12345678902-sign.pem")
	if err != nil {
		fmt.Fprintln(os.Stderr, "read PEM file:", err)
		os.Exit(1)
	}
	testCertRevoked, err = util.GetPEMCertificate(pem)
	if err != nil {
		fmt.Fprintln(os.Stderr, "get PEM certificate: %w", err)
		os.Exit(1)
	}
	pem, err = ioutil.ReadFile("./testdata/TESTCERT_UNKNOWN_12345678903-sign.pem")
	if err != nil {
		fmt.Fprintln(os.Stderr, "read PEM file:", err)
		os.Exit(1)
	}
	testCertUnknown, err = util.GetPEMCertificate(pem)
	if err != nil {
		fmt.Fprintln(os.Stderr, "get PEM certificate: %w", err)
		os.Exit(1)
	}
	pem, err = ioutil.ReadFile("./testdata/intermediate.pem")
	if err != nil {
		fmt.Fprintln(os.Stderr, "read PEM file:", err)
		os.Exit(1)
	}
	testIssuer, err = util.GetPEMCertificate(pem)
	if err != nil {
		fmt.Fprintln(os.Stderr, "get PEM certificate: %w", err)
		os.Exit(1)
	}
	pem, err = ioutil.ReadFile("./testdata/TEST_of_SK_OCSP_RESPONDER_2020.pem")
	if err != nil {
		fmt.Fprintln(os.Stderr, "read PEM file:", err)
		os.Exit(1)
	}
	testResponder, err = util.GetPEMCertificate(pem)
	if err != nil {
		fmt.Fprintln(os.Stderr, "get PEM certificate: %w", err)
		os.Exit(1)
	}
}

func TestVerifyOCSPCertificate(t *testing.T) {
	type test struct {
		description    string
		url            string
		testCert       *x509.Certificate
		testIssuer     *x509.Certificate
		testResponders []*x509.Certificate
		nonce          []byte
		status         int
		invalid        bool
	}
	tests := []test{
		{
			description:    "good certificate",
			url:            testURL,
			testCert:       testCertGood,
			testIssuer:     testIssuer,
			status:         good,
			testResponders: []*x509.Certificate{testResponder},
		},
		{
			description:    "revoked certificate",
			url:            testURL,
			testCert:       testCertRevoked,
			testIssuer:     testIssuer,
			status:         revoked,
			testResponders: []*x509.Certificate{testResponder},
		},
		{
			description:    "unknown certificate",
			url:            testURL,
			testCert:       testCertUnknown,
			testIssuer:     testIssuer,
			status:         unknown,
			testResponders: []*x509.Certificate{testResponder},
		},
		{
			description:    "wrong certificate",
			url:            testURL,
			testCert:       testIssuer,
			testIssuer:     testIssuer,
			status:         good,
			testResponders: []*x509.Certificate{testIssuer},
			invalid:        true,
		},
		{
			description:    "wrong issuer",
			url:            testURL,
			testCert:       testCertUnknown,
			testIssuer:     testResponder,
			status:         good,
			testResponders: []*x509.Certificate{testResponder},
			invalid:        true,
		},
		{
			description:    "wrong responder",
			url:            testURL,
			testCert:       testCertUnknown,
			testIssuer:     testIssuer,
			status:         good,
			testResponders: []*x509.Certificate{testIssuer},
			invalid:        true,
		},
		{
			description:    "wrong status",
			url:            testURL,
			testCert:       testCertUnknown,
			testIssuer:     testIssuer,
			status:         good,
			testResponders: []*x509.Certificate{testResponder},
			invalid:        true,
		},
	}
	ctx := context.Background()
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			if tc.nonce == nil {
				tc.nonce = make([]byte, 20)
				if _, err := rand.Read(tc.nonce); err != nil {
					t.Fatal(err)
				}
			}
			status, err := VerifyOCSPCertificate(ctx, tc.testCert, tc.testIssuer, tc.testResponders, tc.nonce, Conf{URL: tc.url})
			if err != nil {
				if !tc.invalid {
					t.Fatal("unexpected failure:", err)
				}
				return // expected failure
			}
			st := getStatusInt(status)
			if tc.invalid && st == tc.status {
				t.Fatal("unexpected success")
			}
			if !tc.invalid && st != tc.status {
				t.Fatalf("bad certificate status: wanted '%d' got '%d'", tc.status, st)
			}
		})
	}
}

func TestVerifyResponse(t *testing.T) {
	type test struct {
		description    string
		derResponse    string
		testCert       *x509.Certificate
		testIssuer     *x509.Certificate
		testResponders []*x509.Certificate
		nonce          []byte
		status         int
		invalid        bool
	}
	tests := []test{
		{
			description:    "good certificate",
			derResponse:    "./testdata/response-good.der",
			testCert:       testCertGood,
			testIssuer:     testIssuer,
			testResponders: []*x509.Certificate{testResponder},
			status:         good,
			nonce:          []byte{105, 170, 130, 62, 49, 32, 51, 136, 220, 146, 100, 163, 65, 252, 87, 24, 240, 170, 169, 47},
		},
		{
			description:    "revoked certificate",
			derResponse:    "./testdata/response-revoked.der",
			testCert:       testCertRevoked,
			testIssuer:     testIssuer,
			testResponders: []*x509.Certificate{testResponder},
			status:         revoked,
			nonce:          []byte{4, 61, 203, 172, 247, 222, 81, 169, 206, 108, 252, 160, 85, 175, 45, 255, 150, 95, 149, 160},
		},
		{
			description:    "unknown certificate",
			derResponse:    "./testdata/response-unknown.der",
			testCert:       testCertUnknown,
			testIssuer:     testIssuer,
			testResponders: []*x509.Certificate{testResponder},
			status:         unknown,
			nonce:          []byte{101, 131, 166, 243, 5, 130, 23, 208, 40, 148, 249, 244, 79, 10, 5, 248, 117, 182, 142, 87},
		},
		{
			description:    "wrong certificate",
			derResponse:    "./testdata/response-unknown.der",
			testCert:       testCertGood,
			testIssuer:     testResponder,
			testResponders: []*x509.Certificate{testResponder},
			status:         unknown,
			nonce:          []byte{100, 131, 166, 243, 5, 130, 23, 208, 40, 148, 249, 244, 79, 10, 5, 248, 117, 182, 142, 87},
			invalid:        true,
		},
		{
			description:    "wrong issuer",
			derResponse:    "./testdata/response-unknown.der",
			testCert:       testCertUnknown,
			testIssuer:     testResponder,
			testResponders: []*x509.Certificate{testResponder},
			status:         unknown,
			nonce:          []byte{100, 131, 166, 243, 5, 130, 23, 208, 40, 148, 249, 244, 79, 10, 5, 248, 117, 182, 142, 87},
			invalid:        true,
		},
		{
			description:    "wrong responder",
			derResponse:    "./testdata/response-unknown.der",
			testCert:       testCertUnknown,
			testIssuer:     testResponder,
			testResponders: []*x509.Certificate{testResponder},
			status:         unknown,
			nonce:          []byte{100, 131, 166, 243, 5, 130, 23, 208, 40, 148, 249, 244, 79, 10, 5, 248, 117, 182, 142, 87},
			invalid:        true,
		},
		{
			description:    "wrong status",
			derResponse:    "./testdata/response-unknown.der",
			testCert:       testCertUnknown,
			testIssuer:     testIssuer,
			testResponders: []*x509.Certificate{testResponder},
			status:         good,
			nonce:          []byte{101, 131, 166, 243, 5, 130, 23, 208, 40, 148, 249, 244, 79, 10, 5, 248, 117, 182, 142, 87},
			invalid:        true,
		},
		{
			description:    "wrong nonce",
			derResponse:    "./testdata/response-unknown.der",
			testCert:       testCertUnknown,
			testIssuer:     testIssuer,
			testResponders: []*x509.Certificate{testResponder},
			status:         unknown,
			nonce:          []byte{100, 131, 166, 243, 5, 130, 23, 208, 40, 148, 249, 244, 79, 10, 5, 248, 117, 182, 142, 87},
			invalid:        true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			bytes, err := ioutil.ReadFile(tc.derResponse)
			if err != nil {
				t.Fatal(err)
			}
			status, err := VerifyResponse(bytes, tc.testCert, tc.testIssuer, tc.testResponders, tc.nonce)
			if err != nil {
				if !tc.invalid {
					t.Fatal("unexpected failure:", err)
				}
				return // expected failure
			}
			st := getStatusInt(status)
			if tc.invalid && st == tc.status {
				t.Fatal("unexpected success")
			}
			if !tc.invalid && st != tc.status {
				t.Fatalf("bad certificate status: wanted '%d' got '%d'", tc.status, st)
			}
		})
	}
}

func getStatusInt(status LiveCertStatus) int {
	if status.Good {
		return good
	}
	if !status.Unknown {
		return revoked
	}
	return unknown
}
