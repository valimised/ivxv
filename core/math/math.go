package math

import (
	"math/big"
)

// SophieGermainPrime computes a value of
//
//	q = (p - 1) / 2
func SophieGermainPrime(p *big.Int) *big.Int {
	q := new(big.Int)
	q.Sub(p, new(big.Int).SetInt64(1))
	q.Div(q, new(big.Int).SetInt64(2))
	return q
}
