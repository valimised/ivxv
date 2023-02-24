package main

import (
	"context"
	"io/ioutil"
	"strings"
	"testing"

	"ivxv.ee/conf"
	"ivxv.ee/container"
	"ivxv.ee/identity"
	"ivxv.ee/log"
	"ivxv.ee/server"
	"ivxv.ee/storage"
	"ivxv.ee/storage/memory"

	_ "ivxv.ee/container/dummy"
)

func TestBallot(t *testing.T) {
	var err error
	rpc := new(RPC)

	// Set election and question identifiers.
	rpc.election = &conf.Election{
		Identifier: "voting",
		Questions:  []string{"test"},
	}

	// Setup container parser for dummy containers.
	rpc.container, err = container.Configure(container.Conf{container.Dummy: nil})
	if err != nil {
		t.Fatal("failed to configure container parser:", err)
	}

	rpc.identify, err = identity.Get(identity.CommonName)
	if err != nil {
		t.Fatal("failed to get voter identifier:", err)
	}

	// Setup storage client with eligible voters.
	district := string(storage.EncodeAdminDistrict("100", "1"))
	rpc.storage = storage.NewWithProtocol(memory.New(map[string]string{
		"/voters/version":          "0",
		"/voters/0/eligible voter": district,
		"/districts/" + district:   "100.1",
	}))

	// Helper functions to simplify actual tests.
	ctx := log.TestContext(context.Background())
	parse := func(choices, encoded string) (string, error) {
		_, voter, _, _, err := rpc.verify(ctx, choices, container.Dummy, []byte(encoded))
		return voter, err
	}

	// Test certificates.
	eligible := signer(t, "testdata/eligible.pem")
	ineligible := signer(t, "testdata/ineligible.pem")

	// Subtests.
	for _, test := range []struct {
		name     string
		expected error
		choices  string
		encoded  string
	}{
		{"no signers", server.ErrBadRequest, "", ""},
		{"multiple signers", server.ErrBadRequest, "", `
signatures:
  - signer: ` + eligible + `
  - signer: ` + ineligible},

		{"ineligible signer", server.ErrIneligible, "", `
signatures:
  - signer: ` + ineligible},

		{"outdated choices", server.ErrOutdatedChoices, "200.1", `
signatures:
  - signer: ` + eligible},

		{"no ballot", server.ErrBadRequest, "100.1", `
signatures:
  - signer: ` + eligible + `
data:`},

		{"extra data", server.ErrBadRequest, "100.1", `
signatures:
  - signer: ` + eligible + `
data:
  voting.test.ballot: |
  extra.data: |`},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := parse(test.choices, test.encoded)
			if err.Error() != test.expected.Error() {
				t.Errorf("unexpected error: %q, want %q", err, test.expected)
			}
		})
	}

	t.Run("OK", func(t *testing.T) {
		id, err := parse("100.1", `
signatures:
  - signer: `+eligible+`
data:
  voting.test.ballot: choice`)
		if err != nil {
			t.Fatal("unexpected error:", err)
		}
		if id != "eligible voter" {
			t.Errorf("unexpected voter identity: %s", id)
		}
	})
}

// signer loads a test certificate from an external file as a literal value and
// indents all lines to match what is expected of signers.
func signer(t *testing.T, path string) string {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal("failed to load test certificate:", err)
	}
	return strings.ReplaceAll("|\n"+string(b), "\n", "\n      ")
}
