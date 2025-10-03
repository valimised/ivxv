package timestamp

import (
	"context"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"tivi.io/core/crypto/util"
)

var conf Conf

func init() {
	pem, err := ioutil.ReadFile("./testdata/DEMO_SK_TIMESTAMPING_AUTHORITY_2020.pem")
	if err != nil {
		fmt.Fprintln(os.Stderr, "read PEM file:", err)
		os.Exit(1)
	}
	cert, err := util.GetPEMCertificate(pem)
	if err != nil {
		fmt.Fprintln(os.Stderr, "get PEM certificate: %w", err)
		os.Exit(1)
	}
	conf.URL = "http://demo.sk.ee/tsa/"
	conf.Signers = []*x509.Certificate{cert}
}

func TestRequestTimestamp(t *testing.T) {
	ctx := context.Background()
	for len := 15; len < 20; len++ {
		t.Run(fmt.Sprint(len), func(t *testing.T) {
			t.Parallel()
			data := make([]byte, len)
			_, err := rand.Read(data)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = RequestTimestamp(ctx, data, nil, conf); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestVerifyTimestamp(t *testing.T) {
	type test struct {
		description string
		response    string
		data        string
		nonce       string
		invalid     bool
	}
	tests := []test{
		{
			description: "valid timestamp #1",
			response:    "./testdata/response_1_test",
			data:        "./testdata/test_file_1",
			nonce:       "123456789",
		},
		{
			description: "valid timestamp #2",
			response:    "./testdata/response_2_test",
			data:        "./testdata/test_file_2",
			nonce:       "11111111111",
		},
		{
			description: "valid timestamp #3",
			response:    "./testdata/response_3_test",
			data:        "./testdata/test_file_3",
			nonce:       "987654321",
		},
		{
			description: "invalid nonce",
			response:    "./testdata/response_3_test",
			data:        "./testdata/test_file_3",
			nonce:       "000000000000000",
			invalid:     true,
		},
		{
			description: "invalid timestamp response",
			response:    "./testdata/response_4_test",
			data:        "./testdata/test_file_4",
			nonce:       "987654321",
			invalid:     true,
		},
		{
			description: "invalid timestamp data",
			response:    "./testdata/response_5_test",
			data:        "./testdata/test_file_5",
			nonce:       "987654321",
			invalid:     true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			bytes, err := ioutil.ReadFile(tc.response)
			if err != nil {
				t.Fatal(err)
			}
			data, err := ioutil.ReadFile(tc.data)
			if err != nil {
				t.Fatal(err)
			}
			nonce := []byte(tc.nonce)
			_, err = VerifyTimestamp(bytes, data, nonce, conf)
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

func TestGetGenTime(t *testing.T) {
	type test struct {
		description string
		response    string
		genTime     time.Time
		invalid     bool
	}
	tests := []test{
		{
			description: "valid gen time #1",
			response:    "./testdata/response_1_test",
			genTime:     time.Date(2022, 04, 28, 10, 55, 45, 0, time.UTC),
		},
		{
			description: "valid gen time #2",
			response:    "./testdata/response_2_test",
			genTime:     time.Date(2022, 04, 28, 10, 55, 52, 0, time.UTC),
		},
		{
			description: "valid gen time #3",
			response:    "./testdata/response_3_test",
			genTime:     time.Date(2022, 04, 28, 10, 56, 01, 0, time.UTC),
		},
		{
			description: "invalid timestamp response",
			response:    "./testdata/response_4_test",
			genTime:     time.Date(2022, 04, 29, 10, 56, 01, 0, time.UTC),
			invalid:     true,
		},
		{
			description: "invalid gen time",
			response:    "./testdata/response_5_test",
			genTime:     time.Date(2020, 04, 29, 10, 56, 01, 0, time.UTC),
			invalid:     true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			bytes, err := ioutil.ReadFile(tc.response)
			if err != nil {
				t.Fatal(err)
			}
			genTime, err := GetGenTime(bytes)
			if err != nil {
				if !tc.invalid {
					t.Fatal("unexpected failure:", err)
				}
				return // expected failure
			}
			if tc.invalid {
				if genTime.Equal(tc.genTime) {
					t.Fatal("unexpected success")
				}
			} else {
				if !genTime.Equal(tc.genTime) {
					t.Fatalf("bad genTime: wanted '%s' got '%s'", tc.genTime, genTime)
				}
			}
		})
	}
}
