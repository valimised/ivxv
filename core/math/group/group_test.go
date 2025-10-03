package group_test

import (
	"bytes"
	"golang.org/x/crypto/cryptobyte"
	"golang.org/x/crypto/cryptobyte/asn1"
	"testing"

	"tivi.io/core/math/group"
	_ "tivi.io/core/math/group/all"
)

func TestGroupName(t *testing.T) {
	t.Log("Ensure that group.Name works as expected for backwards compatibility")

	supported := []string{
		"RFC3526ModPGroup2048",
		"RFC3526ModPGroup3072",
		"RFC3526ModPGroup4096",
		"RFC3526ModPGroup8192",
		"NIST-P256",
		"NIST-P384",
		"NIST-P521",
		"Edwards25519",
	}

	if len(supported) != len(group.All()) {
		t.Fatal("Supported list of groups in test is incomplete")
	}

	for _, name := range supported {
		t.Run(name, func(t *testing.T) {
			g, err := group.Get(name)
			if err != nil {
				t.Fatal(err)
			}

			if g.Name() != name {
				t.Fatalf("expected %s name, got %s", name, g.Name())
			}
		})
	}
}

func TestGroupOrderInstance(t *testing.T) {
	t.Log("Ensure that group.Order always returns new instance (check immutability property)")

	for _, g := range group.All() {
		t.Run(g.Name(), func(t *testing.T) {
			g, err := group.Get(g.Name())
			if err != nil {
				t.Fatal(err)
			}

			if g.Order() == g.Order() {
				t.Fatal("expected order to be unique pointer")
			}
		})
	}
}

func TestGroupFieldOrderInstance(t *testing.T) {
	t.Log("Ensure that group.FieldOrder always returns new instance (check immutability property)")

	for _, g := range group.All() {
		t.Run(g.Name(), func(t *testing.T) {
			g, err := group.Get(g.Name())
			if err != nil {
				t.Fatal(err)
			}

			if g.FieldOrder() == g.FieldOrder() {
				t.Fatal("expected field order to be unique pointer")
			}
		})
	}
}

func TestGroupIdentityInstance(t *testing.T) {
	t.Log("Ensure that group.Identity always returns new instance (check immutability property)")

	for _, g := range group.All() {
		t.Run(g.Name(), func(t *testing.T) {
			g, err := group.Get(g.Name())
			if err != nil {
				t.Fatal(err)
			}

			if g.Identity() == g.Identity() {
				t.Fatal("expected identity to be unique pointer")
			}
		})
	}
}

func TestGroupIdentity(t *testing.T) {
	t.Log("Ensure that any element operation on group.Identity always returns that element")

	for _, g := range group.All() {
		t.Run(g.Name(), func(t *testing.T) {
			g, err := group.Get(g.Name())
			if err != nil {
				t.Fatal(err)
			}

			E1 := g.Generator()
			E2, err := E1.Op(g.Identity())
			if err != nil {
				t.Fatal(err)
			}

			if err := E1.Equal(E2); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestGroupGeneratorInstance(t *testing.T) {
	t.Log("Ensure that group.Generator always returns new instance (check immutability property)")

	for _, g := range group.All() {
		t.Run(g.Name(), func(t *testing.T) {
			g, err := group.Get(g.Name())
			if err != nil {
				t.Fatal(err)
			}

			if g.Generator() == g.Generator() {
				t.Fatal("expected generator to be unique pointer")
			}
		})
	}
}

func TestGroupGetUnknownGroup(t *testing.T) {
	unknown := "UNKNOWN"

	_, err := group.Get(unknown)
	if err == nil {
		t.Fatal("expected err, got nil")
	}
}

func TestASN1UnmarshalReadASN1Error(t *testing.T) {
	notAGroup := []byte("Not an ASN.1 group")

	_, err := group.Unmarshal(notAGroup)
	if err == nil {
		t.Fatal("expected err, got nil")
	}
}

func TestASN1UnmarshalNameMissingError(t *testing.T) {
	var builder cryptobyte.Builder

	builder.AddASN1(asn1.SEQUENCE, func(_ *cryptobyte.Builder) {})

	der, err := builder.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	_, err = group.Unmarshal(der)
	if err == nil {
		t.Fatal("expected err, got nil")
	}
}

func TestASN1UnmarshalTrailingBytesError(t *testing.T) {
	var builder cryptobyte.Builder
	builder.AddASN1(asn1.SEQUENCE, func(builder *cryptobyte.Builder) {
		builder.AddASN1(asn1.PrintableString, func(_ *cryptobyte.Builder) {})
	})

	// Add just any bytes to the already structured DER, and they will be
	// considered as trailing bytes
	trailingBytes := []byte{0x00}
	builder.AddBytes(trailingBytes)

	der, err := builder.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	_, err = group.Unmarshal(der)
	if err == nil {
		t.Fatal("expected err, got nil")
	}
}

func TestEncodeAndDecodeMessage(t *testing.T) {
	msg := []byte("foobar")

	for _, g := range group.All() {
		t.Run(g.Name(), func(t *testing.T) {
			// So how do we encode a plaintext:

			// 1. We pad a plaintext, returns padded plaintext
			msg2, err := g.PadBytes(msg)
			if err != nil {
				t.Fatal(err)
			}

			// 2. We encode padded plaintext into a group element
			E, err := g.Encode(msg2)
			if err != nil {
				t.Fatal(err)
			}

			// 3. Once we obtained a group element from plaintext,
			// we can do any cryptographic operations on it.
			// At some time, there may be a need to check whether a plaintext
			// was encoded into a group element or not? That's what we do here
			if _, err = E.Decode(); err != nil {
				t.Fatal(err)
			}

			// 4. Once we ensured that group element can be decoded to a plaintext,
			// we can decode it, returns padded plaintext.
			// NB! This isDecodable() check is optional, but allows to fail fast, if
			// group element doesn't contain a plaintext
			msg3, err := E.Decode()
			if err != nil {
				t.Fatal(err)
			}

			// 5. Now we unpad a padded plaintext, returns initial plaintext
			msg4, err := g.UnpadBytes(msg3)
			if !bytes.Equal(msg, msg4) {
				t.Fatalf("Unequal messages m1=%v and m2=%v\n", msg, msg2)
			}
		})
	}
}

func TestGroupMarshalAndUnmarshal(t *testing.T) {
	for _, g := range group.All() {
		t.Run(g.Name(), func(t *testing.T) {
			der, err := group.Marshal(g)
			if err != nil {
				t.Fatal(err)
			}

			g2, err := group.Unmarshal(der)
			if err != nil {
				t.Fatal(err)
			}

			if g.Name() != g2.Name() {
				t.Fatalf("Unequal groups g1=%v and g2=%v\n", g, g2)
			}
		})
	}
}

func TestElementMarshalAndUnmarshal(t *testing.T) {
	for _, g := range group.All() {
		t.Run(g.Name(), func(t *testing.T) {
			E, err := group.RandomElement(g)
			if err != nil {
				t.Fatal(err)
			}

			der, err := E.Marshal()
			if err != nil {
				t.Fatal(err)
			}

			E2, err := g.ElementOf(der)
			if err != nil {
				t.Fatal(err)
			}

			err = E2.Equal(E)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestMath(t *testing.T) {
	for _, g := range group.All() {
		t.Run(g.Name(), func(t *testing.T) {
			E1, err := group.RandomElement(g)
			if err != nil {
				t.Error(err)
			}

			E2, err := group.RandomElement(g)
			if err != nil {
				t.Error(err)
			}

			E3, err := E1.Op(E2)
			if err != nil {
				t.Error(err)
			}

			E4 := E2.Inverse()

			E5, err := E3.Op(E4)
			if err != nil {
				t.Error(err)
			}

			err = E5.Equal(E1)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestScalarMath(t *testing.T) {
	for _, g := range group.All() {
		t.Run(g.Name(), func(t *testing.T) {
			s, err := group.RandomScalar(g.Order())
			if err != nil {
				t.Fatal(err)
			}

			s2, err := s.Add(s)
			if err != nil {
				t.Fatal(err)
			}

			E, err := group.RandomElement(g)
			if err != nil {
				t.Fatal(err)
			}

			E2, err := E.Scale(s2)
			if err != nil {
				t.Fatal(err)
			}

			E, err = E.Scale(s)
			if err != nil {
				t.Fatal(err)
			}

			E, err = E.Op(E)
			if err != nil {
				t.Fatal(err)
			}

			err = E.Equal(E2)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
