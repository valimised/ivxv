package cms

import (
	"bytes"
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"tivi.io/core/crypto/util"
)

func TestVerifySignedData(t *testing.T) {
	type test struct {
		description      string
		cert             string // Filepath of the Certificate
		signedData       string // Filepath of the SignedData to be verified
		detachedData     string // Filepath of the detached data
		originalData     string // Filepath of the original data
		otherSignedAttrs map[string]AttrToVerify
		unsignedAttrs    map[string]AttrToVerify
		invalid          bool
	}

	tests := []test{
		{
			description:  "valid signed data with no callback",
			cert:         "./testdata/trusted.pem",
			signedData:   "./testdata/test1.txt.pk7",
			originalData: "./testdata/test1.txt",
		},
		{
			description:  "valid signed data with example 'past' callback",
			cert:         "./testdata/trusted.pem",
			signedData:   "./testdata/test1.txt.pk7",
			originalData: "./testdata/test1.txt",
			otherSignedAttrs: map[string]AttrToVerify{
				idSigningTime: {
					AttributeID:    idSigningTime,
					VerifyCallback: func(b []byte) error { return verifySigningTimePast(b) }}},
		},
		{
			description:  "valid signed data with example 'before' callback",
			cert:         "./testdata/trusted.pem",
			signedData:   "./testdata/test1.txt.pk7",
			originalData: "./testdata/test1.txt",
			otherSignedAttrs: map[string]AttrToVerify{
				idSigningTime: {
					AttributeID: idSigningTime,
					VerifyCallback: func(b []byte) error {
						return verifySigningTimeBefore(b, time.Date(2022, 5, 1, 11, 11, 11, 11, time.UTC))
					}}},
		},
		{
			description:  "valid signed data with example 'between' callback",
			cert:         "./testdata/trusted.pem",
			signedData:   "./testdata/test1.txt.pk7",
			originalData: "./testdata/test1.txt",
			otherSignedAttrs: map[string]AttrToVerify{
				idSigningTime: {
					AttributeID: idSigningTime,
					VerifyCallback: func(b []byte) error {
						return verifySigningTimeBetween(b, time.Date(2002, 5, 1, 11, 11, 11, 11, time.UTC), time.Date(2022, 5, 1, 11, 11, 11, 11, time.UTC))
					}}},
		},
		{
			description: "not valid signed data",
			cert:        "./testdata/trusted.pem",
			signedData:  "./testdata/test2.txt.pk7",
			invalid:     true,
		},
		{
			description: "not valid certificate",
			cert:        "./testdata/nottrusted.pem",
			signedData:  "./testdata/test1.txt.pk7",
			invalid:     true,
		},
		{
			description:  "valid signed data but missing attribute",
			cert:         "./testdata/trusted.pem",
			signedData:   "./testdata/test1.txt.pk7",
			originalData: "./testdata/test1.txt",
			unsignedAttrs: map[string]AttrToVerify{
				idSigningTime: {
					AttributeID: idSigningTime,
					VerifyCallback: func(b []byte) error {
						return verifySigningTimeBetween(b, time.Date(2002, 5, 1, 11, 11, 11, 11, time.UTC), time.Date(2022, 5, 1, 11, 11, 11, 11, time.UTC))
					}}},
			invalid: true,
		},
		{
			description:  "valid signed data but mismatch attached and detached data",
			cert:         "./testdata/trusted.pem",
			signedData:   "./testdata/test1.txt.pk7",
			detachedData: "./testdata/test2.txt.pk7",
			invalid:      true,
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
			var detachedData []byte
			if tc.detachedData != "" {
				detachedData, err = ioutil.ReadFile(tc.detachedData)
				if err != nil {
					t.Fatalf("read detached data file: %s", err)
				}
			}
			err = VerifySignedData(signedData, certs, detachedData, tc.otherSignedAttrs, tc.unsignedAttrs)
			if err != nil {
				if !tc.invalid {
					t.Fatal("unexpected failure:", err)
				}
				return // expected failure
			}
			if tc.invalid {
				t.Fatal("unexpected success")
			}
			data, err := GetDataEContent(signedData)
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

func TestGetDataEContent(t *testing.T) {
	type test struct {
		description  string
		signedData   string // Filepath of the SignedData to be verified
		originalData string // Filepath of the original data
		invalid      bool
	}

	tests := []test{
		{
			description:  "valid signed data",
			signedData:   "./testdata/test1.txt.pk7",
			originalData: "./testdata/test1.txt",
		},
		{
			description: "valid signed data with no attached data",
			signedData:  "../cades/testdata/plain-signed-attr.cms",
			invalid:     true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			signedData, err := ioutil.ReadFile(tc.signedData)
			if err != nil {
				t.Fatalf("read signed data: %s", err)
			}
			data, err := GetDataEContent(signedData)
			if err != nil {
				if !tc.invalid {
					t.Fatal("unexpected failure:", err)
				}
				return // expected failure
			}
			if tc.invalid {
				t.Fatal("unexpected success")
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

// verifySigningTimePast is an example of a VerifyCallback
// for an optional AttrToVerify. This method simply verifies
// if the SigningTime attribute is in the past, related to the moment
// of the verification of the signed data.
func verifySigningTimePast(value []byte) error {
	var signingTime time.Time
	rest, err := asn1.Unmarshal(value, &signingTime)
	if err != nil {
		return fmt.Errorf("unmarshal signing time: %w", err)
	}
	if len(rest) > 0 {
		return fmt.Errorf("unmarshal signing time excess bytes: %d", len(rest))
	}
	diff := time.Since(signingTime)
	if diff <= 0 {
		return fmt.Errorf("signing time '%s' in the future", signingTime)
	}
	return nil
}

// verifySigningTimeBefore is an example of a VerifyCallback
// for an optional AttrToVerify. This method simply verifies
// if the SigningTime attribute is before some specific time.
func verifySigningTimeBefore(value []byte, t time.Time) error {
	var signingTime time.Time
	rest, err := asn1.Unmarshal(value, &signingTime)
	if err != nil {
		return fmt.Errorf("unmarshal signing time: %w", err)
	}
	if len(rest) > 0 {
		return fmt.Errorf("unmarshal signing time excess bytes: %d", len(rest))
	}
	if !signingTime.Before(t) {
		return fmt.Errorf("signing time '%s' not before time '%s'", signingTime, t)
	}
	return nil
}

// verifySigningTimeBefore is an example of a VerifyCallback
// for an optional AttrToVerify. This method simply verifies
// if the SigningTime attribute is between two specific times: aT < signingTime < bT.
func verifySigningTimeBetween(value []byte, aT time.Time, bT time.Time) error {
	var signingTime time.Time
	rest, err := asn1.Unmarshal(value, &signingTime)
	if err != nil {
		return fmt.Errorf("unmarshal signing time: %w", err)
	}
	if len(rest) > 0 {
		return fmt.Errorf("unmarshal signing time excess bytes: %d", len(rest))
	}
	if !(aT.Before(signingTime) && signingTime.Before(bT)) {
		return fmt.Errorf("signing time '%s' not between '%s' and '%s'", signingTime, aT, bT)
	}
	return nil
}
