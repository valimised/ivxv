package ecqp

import (
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"
	"tivi.io/core/math/group"
	_ "tivi.io/core/math/group/edwards25519"
	"tivi.io/core/math/group/internal"
)

func init() {
	name := "NIST-P384"
	var err error
	ecqPGroupP384, err = group.Get(name)
	if err != nil {
		panic(err)
	}
	ecqPGroupP384.(*ecqPGroup).fieldEncodingBits = 7
}

var ecqPGroupP384 group.Group

var ecqPGroupP224 = &ecqPGroup{name: "NIST-P224", Curve: elliptic.P224()}

func TestGroupEqualDifferentOrder(t *testing.T) {
	t.Log("Equality checks should fail for different groups, but succeed for the same ones")

	for _, g := range group.All() {
		t.Run(fmt.Sprintf("Negative cases for %s group", g.Name()), func(t *testing.T) {
			g, err := group.Get(g.Name())
			if err != nil {
				t.Fatal(err)
			}

			if err := g.Equal(ecqPGroupP224); err == nil {
				t.Fatal(err)
			}
		})
	}

	// Positive cases
	for _, g := range group.All() {
		t.Run(fmt.Sprintf("Positive cases for %s group", g.Name()), func(t *testing.T) {
			g, err := group.Get(g.Name())
			if err != nil {
				t.Fatal(err)
			}

			if err := g.Equal(g); err != nil {
				t.Fatalf("expected nil error, but got %v", err)
			}
		})
	}
}

func TestGroupIsGroupElementCastError(t *testing.T) {
	t.Log("TODO:")

	g, err := group.Get("Edwards25519")
	if err != nil {
		panic(err)
	}

	if err = ecqPGroupP384.IsGroupElement(g.Generator()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGroupIsGroupElementDifferentGroups(t *testing.T) {
	t.Log("TODO:")

	g, err := group.Get("NIST-P521")
	if err != nil {
		panic(err)
	}

	if err = ecqPGroupP384.IsGroupElement(g.Generator()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGroupIsGroupElementNotOnCurve(t *testing.T) {
	t.Log("TODO:")

	g, err := group.Get("NIST-P384")
	if err != nil {
		panic(err)
	}

	E := g.Generator()
	E.(*ecqPElement).x.SetInt64(11112018)

	if err = ecqPGroupP384.IsGroupElement(E); err == nil {
		t.Fatal("expected error, got nil")
	}
	if reflect.TypeOf(err) != reflect.TypeOf(internal.NotOnCurve{}) {
		t.Fatalf("expected error %v, but got %v", internal.NotOnCurve{}, reflect.TypeOf(err))
	}
}

func TestGroupElementOfASN1OctetString(t *testing.T) {
	t.Log("TODO:")

	// Generator element of P384 group as ASN.1 BitString hex encoded
	E := "03620004aa87ca22be8b05378eb1c71ef320ad746e1d3b628ba79b9859f741e082542a385502f25dbf55296c3a545e3872760ab73617de4a96262c6f5d9e98bf9292dc29f8f41dbd289a147ce9da3113b5f0b8c00a60b1ce1d7e819d7a431d7c90ea0e5f"

	der, err := hex.DecodeString(E)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = ecqPGroupP384.ElementOf(der); err == nil {
		t.Fatal("expected error, got nil")
	}
	if reflect.TypeOf(err) != reflect.TypeOf(internal.NotASN1OctetString{}) {
		t.Fatalf("expected error %v, but got %v", internal.NotASN1OctetString{}, reflect.TypeOf(err))
	}
}

func TestGroupElementOfASN1TrailingBytes(t *testing.T) {
	t.Log("TODO:")

	// Generator element of P384 group as ASN.1 OctetString hex encoded
	// with 0xAA, 0xBB, 0xCC trailing bytes (appended at the end)
	E := "046104aa87ca22be8b05378eb1c71ef320ad746e1d3b628ba79b9859f741e082542a385502f25dbf55296c3a545e3872760ab73617de4a96262c6f5d9e98bf9292dc29f8f41dbd289a147ce9da3113b5f0b8c00a60b1ce1d7e819d7a431d7c90ea0e5faabbcc"

	der, err := hex.DecodeString(E)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = ecqPGroupP384.ElementOf(der); err == nil {
		t.Fatal("expected error, got nil")
	}
	if reflect.TypeOf(err) != reflect.TypeOf(internal.ASN1TrailingBytes{}) {
		t.Fatalf("expected error %v, but got %v", internal.ASN1TrailingBytes{}, reflect.TypeOf(err))
	}
}

func TestGroupElementOfNotUncompressedPoint(t *testing.T) {
	t.Log("TODO:")

	// Generator element of P384 group as ASN.1 OctetString hex encoded
	// with 0xAA, 0xBB, 0xCC trailing bytes (appended at the end)
	E := "046100aa87ca22be8b05378eb1c71ef320ad746e1d3b628ba79b9859f741e082542a385502f25dbf55296c3a545e3872760ab73617de4a96262c6f5d9e98bf9292dc29f8f41dbd289a147ce9da3113b5f0b8c00a60b1ce1d7e819d7a431d7c90ea0e5f"

	der, err := hex.DecodeString(E)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = ecqPGroupP384.ElementOf(der); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGroupElementOfCoordinatesOverflow(t *testing.T) {
	t.Log("TODO:")

	// Generator element of P384 group as ASN.1 OctetString hex encoded
	// X and Y coordinates are of exactly elliptic curve P parameter value,
	// which makes it to error at unmarshalling
	E := "046104fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffeffffffff0000000000000000fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffeffffffff0000000000000000ffffffff"

	der, err := hex.DecodeString(E)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = ecqPGroupP384.ElementOf(der); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGroupElementOfNotUOnCurve(t *testing.T) {
	t.Log("TODO:")

	// Generator element of P384 group as ASN.1 OctetString hex encoded
	// X coordinate has 0xAA as fifth byte, which makes it not compatible with
	// Y coordinate and therefore outside the curve
	E := "046104aa87ca22aa8b05378eb1c71ef320ad746e1d3b628ba79b9859f741e082542a385502f25dbf55296c3a545e3872760ab73617de4a96262c6f5d9e98bf9292dc29f8f41dbd289a147ce9da3113b5f0b8c00a60b1ce1d7e819d7a431d7c90ea0e5f"

	der, err := hex.DecodeString(E)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = ecqPGroupP384.ElementOf(der); err == nil {
		t.Fatal("expected error, got nil")
	}
	if reflect.TypeOf(err) != reflect.TypeOf(internal.NotOnCurve{}) {
		t.Fatalf("expected error %v, but got %v", internal.NotOnCurve{}, reflect.TypeOf(err))
	}
}

func TestGroupElementOfPointInvalidFormat(t *testing.T) {
	t.Log("TODO:")

	// Generator element of P384 group as ASN.1 OctetString hex encoded
	// X coordinate has 0xAA as fifth byte, which makes it not compatible with
	// Y coordinate and therefore outside the curve
	E := "046204aa87ca22be8b05378eb1c71ef320ad746e1d3b628ba79b9859f741e082542a385502f25dbf55296c3a545e3872760ab73617de4a96262c6f5d9e98bf9292dc29f8f41dbd289a147ce9da3113b5f0b8c00a60b1ce1d7e819d7a431d7c90ea0e5fff"

	der, err := hex.DecodeString(E)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = ecqPGroupP384.ElementOf(der); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGroupElementOfPointAtInfinity(t *testing.T) {
	t.Log("TODO:")

	// Generator element of P384 group as ASN.1 OctetString hex encoded
	// X coordinate has 0xAA as fifth byte, which makes it not compatible with
	// Y coordinate and therefore outside the curve
	E := ecqPGroupP384.Identity()
	der, err := E.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	if _, err = ecqPGroupP384.ElementOf(der); err != nil {
		t.Fatal(err)
	}
}
