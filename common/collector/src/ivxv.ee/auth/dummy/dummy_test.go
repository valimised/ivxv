package dummy

import (
	"context"
	"testing"

	"ivxv.ee/errors"
)

const (
	authenticatedDN = "\x30\x1f\x31\x1d\x30\x1b\x06\x03\x55\x04\x03\x13" +
		"\x14\x61\x75\x74\x68\x65\x6e\x74\x69\x63\x61\x74\x65\x64" +
		"\x20\x63\x6c\x69\x65\x6e\x74"

	unauthenticatedDN = "\x30\x21\x31\x1f\x30\x1d\x06\x03\x55\x04\x03" +
		"\x13\x16\x75\x6e\x61\x75\x74\x68\x65\x6e\x74\x69\x63\x61" +
		"\x74\x65\x64\x20\x63\x6c\x69\x65\x6e\x74"
)

func TestVerify(t *testing.T) {
	v := Conf{
		Authenticated: []string{"authenticated client"},
	}

	t.Run("authenticated", func(t *testing.T) {
		if _, err := v.Verify(context.Background(), []byte(authenticatedDN)); err != nil {
			t.Error("unexpected authentication failure:", err)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		want := new(NotAllowedError)
		if _, err := v.Verify(context.Background(), []byte(unauthenticatedDN)); err == nil {
			t.Error("unexpected authentication success")
		} else if errors.CausedBy(err, want) == nil {
			t.Errorf("unexpected authentication error: %v; want cause %T", err, want)
		}
	})
}
