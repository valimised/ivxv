package tsp

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"ivxv.ee/log"

	_ "crypto/sha256"
)

var client *Client

func init() {
	pem, err := ioutil.ReadFile("testdata/DEMO-SK_TIMESTAMPING_AUTHORITY.pem")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	client, err = New(&Conf{
		URL:       "http://demo.sk.ee/tsa/",
		Signers:   []string{string(pem)},
		DelayTime: 1,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func TestCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("Short mode on, skipping TSP test against live responder")
	}

	ctx := log.TestContext(context.Background())

	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	if _, err := client.Create(ctx, data, nil); err != nil {
		t.Fatal(err)
	}
}

func TestCheck(t *testing.T) {
	bytes, err := ioutil.ReadFile("testdata/test_response")
	if err != nil {
		t.Fatal(err)
	}

	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	nonce := []byte{0x23, 0x0f, 0x52, 0x18, 0x68, 0x30, 0xe1, 0x1b, 0x6d,
		0x46, 0xb0, 0x28, 0x03, 0x21, 0x11, 0x86, 0xa2, 0xba, 0x11, 0x1a}
	if err = client.Check(bytes, data, nonce); err != nil {
		t.Fatal(err)
	}
}
