package group

import (
	"bytes"
	"math/big"
	"testing"
)

func TestAddScalar(t *testing.T) {
	mod1 := new(big.Int).SetInt64(3)
	mod2 := new(big.Int).SetInt64(3)

	s1 := OneScalar(mod1)
	s2 := OneScalar(mod2)

	// 1 + 1 = 2
	s3, err := s1.Add(s2)
	if err != nil {
		t.Fatal(err)
	}

	expected := new(big.Int).SetInt64(2)

	if s3.value.Cmp(expected) != 0 {
		t.Fatalf("expected s3=%d, got=%d\n", expected, s3.value)
	}
}

func TestSubScalar(t *testing.T) {
	mod1 := new(big.Int).SetInt64(3)
	mod2 := new(big.Int).SetInt64(3)

	s1 := NewScalar(new(big.Int).SetInt64(3), mod1)
	s2 := NewScalar(new(big.Int).SetInt64(2), mod2)

	// 3 - 2 = 1
	s3, err := s1.Sub(s2)
	if err != nil {
		t.Fatal(err)
	}

	expected := new(big.Int).SetInt64(1)

	if s3.value.Cmp(expected) != 0 {
		t.Fatalf("expected s3=%d, got=%d\n", expected, s3.value)
	}
}

func TestMulScalar(t *testing.T) {
	mod1 := new(big.Int).SetInt64(7)
	mod2 := new(big.Int).SetInt64(7)

	s1 := NewScalar(new(big.Int).SetInt64(2), mod1)
	s2 := NewScalar(new(big.Int).SetInt64(3), mod2)

	// 3 * 2 = 6
	s3, err := s1.Mul(s2)
	if err != nil {
		t.Fatal(err)
	}

	expected := new(big.Int).SetInt64(6)

	if s3.value.Cmp(expected) != 0 {
		t.Fatalf("expected s3=%d, got=%d\n", expected, s3.value)
	}
}

func TestExp(t *testing.T) {
	mod1 := new(big.Int).SetInt64(7)

	s1 := NewScalar(new(big.Int).SetInt64(2), mod1)

	exp := new(big.Int).SetInt64(3)

	// 2^3 = 8
	// 8 mod 7 = 1
	s3 := s1.Exp(exp)

	expected := new(big.Int).SetInt64(1)

	if s3.value.Cmp(expected) != 0 {
		t.Fatalf("expected s3=%d, got=%d\n", expected, s3.value)
	}
}

func TestNegate(t *testing.T) {
	mod1 := new(big.Int).SetInt64(7)

	s1 := NewScalar(new(big.Int).SetInt64(2), mod1)

	// -2 modulo 7 = 5, since 7 - 2 = 5
	s2 := s1.Negate()

	expected := new(big.Int).SetInt64(5)

	if s2.value.Cmp(expected) != 0 {
		t.Fatalf("expected s2=%d, got=%d\n", expected, s2.value)
	}
}

func TestInverse(t *testing.T) {
	mod1 := new(big.Int).SetInt64(7)

	s1 := NewScalar(new(big.Int).SetInt64(2), mod1)

	s2 := s1.Inverse()

	// 2 modinv 7 = 4, since 2*4 modulo 7 = 1
	expected := new(big.Int).SetInt64(4)

	if s2.value.Cmp(expected) != 0 {
		t.Fatalf("expected s2=%d, got=%d\n", expected, s2.value)
	}
}

func TestEqual(t *testing.T) {
	mod1 := new(big.Int).SetInt64(7)
	mod2 := new(big.Int).SetInt64(7)

	s1 := NewScalar(new(big.Int).SetInt64(2), mod1)
	s2 := NewScalar(new(big.Int).SetInt64(2), mod2)

	err := s1.Equal(s2)
	if err != nil {
		t.Fatal(err)
	}
}

func TestEqualMod(t *testing.T) {
	mod1 := new(big.Int).SetInt64(1)
	mod2 := new(big.Int).SetInt64(1)

	s1 := ZeroScalar(mod1)
	s2 := ZeroScalar(mod2)

	if err := s1.equalMod(s2); err != nil {
		t.Fatal(err)
	}
}

func TestEqualModDifferentMods(t *testing.T) {
	mod1 := new(big.Int).SetInt64(1)
	mod2 := new(big.Int).SetInt64(2)

	s1 := ZeroScalar(mod1)
	s2 := ZeroScalar(mod2)

	if err := s1.equalMod(s2); err == nil {
		t.Fatal("expected error, but got nil")
	}
}

func TestScalarValueOfReader(t *testing.T) {
	mod1 := new(big.Int).SetInt64(7)

	s1 := NewScalar(new(big.Int).SetInt64(2), mod1)

	// Always returns 2
	reader := bytes.NewReader(s1.value.Bytes())

	s3, err := ScalarValueOfReader(reader, s1.modulo)
	if err != nil {
		t.Fatal(err)
	}

	if s3.value.Cmp(s1.value) != 0 {
		t.Fatalf("expected s1=%d to be equal to s3=%d\n", s1.value, s3.value)
	}
}

func TestScalarMarshalAndUnmarshalASN1(t *testing.T) {
	for _, g := range All() {
		t.Run(g.Name(), func(t *testing.T) {
			s1, err := RandomScalar(g.Order())
			if err != nil {
				t.Fatal(err)
			}

			der, err := s1.Marshal()
			if err != nil {
				t.Fatal(err)
			}

			s2, err := UnmarshalScalar(der, g.Order())
			if err != nil {
				t.Fatal(err)
			}

			err = s1.Equal(s2)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
