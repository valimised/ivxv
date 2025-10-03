package math

import (
	"math/big"
	"testing"
)

func TestSophieGermainPrime(t *testing.T) {
	p := new(big.Int).SetInt64(23)
	q := SophieGermainPrime(p) // (23 - 1) / 2
	expected := new(big.Int).SetInt64(11)
	if q.Cmp(expected) != 0 {
		t.Fatalf("expected Sophie-Germain prime to be=%d, got=%d\n", expected, q)
	}
}
