package ecqp

import (
	"math/big"
	"reflect"
	"testing"
	"tivi.io/core/math/group"
	_ "tivi.io/core/math/group/edwards25519"
	"tivi.io/core/math/group/internal"
)

func TestElementScaleModulo(t *testing.T) {
	t.Log("TODO:")

	s := group.NewScalar(new(big.Int).SetInt64(4), ecqPGroupP384.FieldOrder())
	E := ecqPGroupP384.Identity()
	_, err := E.Scale(s)
	if err == nil {
		t.Fatal(err)
	}
	if reflect.TypeOf(err) != reflect.TypeOf(internal.ScalarModCmpOrder{}) {
		t.Fatalf("expected error %v, but got %v", internal.ScalarModCmpOrder{}, reflect.TypeOf(err))
	}
}

func TestElementOpCast(t *testing.T) {
	t.Log("TODO:")

	g, err := group.Get("Edwards25519")
	if err != nil {
		t.Fatal(err)
	}

	E := ecqPGroupP384.Identity()
	_, err = E.Op(g.Generator())
	if err == nil {
		t.Fatal(err)
	}
}

func TestElementOpDifferentGroup(t *testing.T) {
	t.Log("TODO:")

	g, err := group.Get("NIST-P256")
	if err != nil {
		t.Fatal(err)
	}

	E := ecqPGroupP384.Identity()
	_, err = E.Op(g.Generator())
	if err == nil {
		t.Fatal(err)
	}
}

func TestElementEqualCast(t *testing.T) {
	t.Log("TODO:")

	g, err := group.Get("Edwards25519")
	if err != nil {
		t.Fatal(err)
	}

	E := ecqPGroupP384.Identity()
	err = E.Equal(g.Generator())
	if err == nil {
		t.Fatal(err)
	}
}

func TestElementEqualDifferentGroup(t *testing.T) {
	t.Log("TODO:")

	g, err := group.Get("NIST-P521")
	if err != nil {
		t.Fatal(err)
	}

	E := ecqPGroupP384.Identity()
	err = E.Equal(g.Generator())
	if err == nil {
		t.Fatal(err)
	}
}

func TestElementEqualDifferentX(t *testing.T) {
	t.Log("TODO:")

	g, err := group.Get("NIST-P384")
	if err != nil {
		t.Fatal(err)
	}

	E := ecqPGroupP384.Identity()
	err = E.Equal(g.Generator())
	if err == nil {
		t.Fatal(err)
	}
}

func TestElementEqualDifferentY(t *testing.T) {
	t.Log("TODO:")

	g, err := group.Get("NIST-P384")
	if err != nil {
		t.Fatal(err)
	}

	E1 := g.Generator()
	E2 := g.Generator()
	E2.(*ecqPElement).y.SetInt64(11112018)

	err = E1.Equal(E2)
	if err == nil {
		t.Fatal(err)
	}
}

func TestElementDecode(t *testing.T) {
	t.Log("TODO:")

	E, err := ecqPGroupP384.Encode([]byte("Hello!")) // Don't modify me!
	if err != nil {
		t.Fatal(err)
	}

	// Add one bit to make size not equal to field order - 1
	E.(*ecqPElement).x.Add(E.(*ecqPElement).x, new(big.Int).SetInt64(1))

	// Luckily we have y2 perfect root here, but it is a pure luck because of combination of
	// curve and "Hello!" message
	y2 := shortWeierstrassFunc(E.(*ecqPElement).x, E.(*ecqPElement).g.Curve)
	y := new(big.Int).ModSqrt(y2, ecqPGroupP384.FieldOrder())
	E.(*ecqPElement).y.Set(y)

	_, err = E.Decode()
	if err == nil {
		t.Fatal("expected error got nil")
	}
}

func TestElementMarshal(t *testing.T) {
	t.Log("TODO:")

	E, err := ecqPGroupP384.Encode([]byte("=]"))
	if err != nil {
		t.Fatal(err)
	}

	// Add one bit to make size not equal to field order - 1
	E.(*ecqPElement).x.Sub(E.(*ecqPElement).x, new(big.Int).SetInt64(11112018))

	_, err = E.Marshal()
	if err == nil {
		t.Fatal("expected error got nil")
	}
	if reflect.TypeOf(err) != reflect.TypeOf(internal.NotOnCurve{}) {
		t.Fatalf("expected error %v, but got %v", internal.NotOnCurve{}, reflect.TypeOf(err))
	}
}
