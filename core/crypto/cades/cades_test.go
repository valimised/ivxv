package cades

import (
	"bytes"
	"crypto/x509"
	"io/ioutil"
	"testing"

	"tivi.io/core/crypto/cms"
	"tivi.io/core/crypto/util"
)

func TestCAdESBVerifySignedData(t *testing.T) {
	type test struct {
		description  string
		cert         string // Filepath of the Certificate
		signedData   string // Filepath of the SignedData to be verified
		originalData string // Filepath of the original data
		invalid      bool
	}

	tests := []test{
		{
			description:  "valid signed data",
			cert:         "../cms/testdata/trusted.pem",
			signedData:   "../cms/testdata/test1.txt.pk7",
			originalData: "../cms/testdata/test1.txt",
		},
		{
			description: "not valid signed data",
			cert:        "../cms/testdata/trusted.pem",
			signedData:  "../cms/testdata/test2.txt.pk7",
			invalid:     true,
		},
		{
			description: "not valid certificate",
			cert:        "../cms/testdata/nottrusted.pem",
			signedData:  "../cms/testdata/test1.txt.pk7",
			invalid:     true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			signedData, err := ioutil.ReadFile(tc.signedData)
			if err != nil {
				t.Fatalf("read signed data: %s", err)
			}
			pem, err := ioutil.ReadFile(tc.cert)
			if err != nil {
				t.Fatalf("read PEM file: %s", err)
			}
			cert, err := util.GetPEMCertificate(pem)
			if err != nil {
				t.Fatalf("get PEM certificate: %s", err)
			}
			certs := []*x509.Certificate{cert}
			err = CAdESBVerifySignedData(signedData, certs, nil)
			if err != nil {
				if !tc.invalid {
					t.Fatal("unexpected failure:", err)
				}
				return // expected failure
			}
			if tc.invalid {
				t.Fatal("unexpected success")
			}
			data, err := cms.GetDataEContent(signedData)
			if err != nil {
				if !tc.invalid {
					t.Fatal("unexpected failure:", err)
				}
				return // expected failure
			}
			originalData, err := ioutil.ReadFile(tc.originalData)
			if err != nil {
				t.Fatalf("read original data: %s", err)
			}
			if !bytes.Equal(originalData, data) {
				t.Fatalf("original data not equal to data content")
			}
		})
	}
}

func TestCAdESTVerifySignedData(t *testing.T) {
	type test struct {
		description  string
		certs        []string // Filepath of the Certificate
		signedData   string   // Filepath of the SignedData to be verified
		originalData string   // Filepath of the original data
		nonce        []byte
		invalid      bool
	}

	tests := []test{
		{
			description: "valid signed data",
			certs: []string{
				"../timestamp/testdata/DEMO_SK_TIMESTAMPING_AUTHORITY_2020.pem",
				"./testdata/demo-ca-cert.pem",
				"./testdata/demo-user-cert.pem",
			},
			nonce:        []byte{1, 128, 179, 189, 0, 230},
			signedData:   "./testdata/plain-signed-attr.cms",
			originalData: "./testdata/plain-unsigned.txt",
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			signedData, err := ioutil.ReadFile(tc.signedData)
			if err != nil {
				t.Fatalf("read signed data: %s", err)
			}
			var certs []*x509.Certificate
			for _, c := range tc.certs {
				pem, err := ioutil.ReadFile(c)
				if err != nil {
					t.Fatalf("read PEM file: %s", err)
				}
				cert, err := util.GetPEMCertificate(pem)
				if err != nil {
					t.Fatalf("get PEM certificate: %s", err)
				}
				certs = append(certs, cert)
			}
			originalData, err := ioutil.ReadFile(tc.originalData)
			if err != nil {
				t.Fatalf("read signed data: %s", err)
			}
			err = CAdESTVerifySignedData(signedData, certs, originalData, tc.nonce)
			if err != nil {
				if !tc.invalid {
					t.Fatal("unexpected failure:", err)
				}
				return // expected failure
			}
			if tc.invalid {
				t.Fatal("unexpected success")
			}
		})
	}
}
