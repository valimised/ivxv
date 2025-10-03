package group

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"

	asn_1 "tivi.io/core/encoding/asn1"

	"golang.org/x/crypto/cryptobyte"
)

// Scalar is an immutable scalar of an underlying group, all scalar operations
// are performed modulo group generator order.
type Scalar struct {
	value *big.Int
	// modulo is a group generator order.
	modulo *big.Int
}

// NewScalar returns new Scalar of val.
// NB! If mod is <= 0, then behaviour of Scalar is undefined.
func NewScalar(val, mod *big.Int) *Scalar {
	return &Scalar{
		value:  new(big.Int).Set(val),
		modulo: new(big.Int).Set(mod),
	}
}

// ZeroScalar returns new group scalar with value=0.
func ZeroScalar(mod *big.Int) *Scalar {
	return NewScalar(new(big.Int), mod)
}

// OneScalar returns new group scalar with value=1.
func OneScalar(mod *big.Int) *Scalar {
	return NewScalar(big.NewInt(1), mod)
}

// RandomScalar returns new random group scalar with value=[1,mod).
// Randomness is obtained via rand.Reader.
// See ScalarValueOfReader if you prefer custom randomness source.
func RandomScalar(mod *big.Int) (*Scalar, error) {
	return ScalarValueOfReader(rand.Reader, mod)
}

// Value returns value of a group scalar.
func (s *Scalar) Value() *big.Int {
	return new(big.Int).Set(s.value)
}

// Modulo returns modulo of a group scalar.
func (s *Scalar) Modulo() *big.Int {
	return new(big.Int).Set(s.modulo)
}

// Add returns
//
//	this + s2
func (s *Scalar) Add(s2 *Scalar) (*Scalar, error) {
	err := s.equalMod(s2)
	if err != nil {
		return nil, err
	}

	s3 := NewScalar(s.value, s.modulo)
	s3.value.Add(s3.value, s2.value)
	s3.value.Mod(s3.value, s3.modulo)

	return s3, nil
}

// Sub returns
//
//	this - s2
func (s *Scalar) Sub(s2 *Scalar) (*Scalar, error) {
	err := s.equalMod(s2)
	if err != nil {
		return nil, err
	}

	s3 := NewScalar(s.value, s.modulo)
	s3.value.Sub(s3.value, s2.value)
	s3.value.Mod(s3.value, s3.modulo)

	return s3, nil
}

// Mul returns
//
//	this * s2
func (s *Scalar) Mul(s2 *Scalar) (*Scalar, error) {
	err := s.equalMod(s2)
	if err != nil {
		return nil, err
	}

	s3 := NewScalar(s.value, s.modulo)
	s3.value.Mul(s3.value, s2.value)
	s3.value.Mod(s3.value, s3.modulo)

	return s3, nil
}

// Exp returns
//
//	this ^ e
func (s *Scalar) Exp(e *big.Int) *Scalar { //nolint:revive
	s2 := NewScalar(s.value, s.modulo)
	s2.value.Exp(s2.value, e, s2.modulo)

	return s2
}

// Negate returns
//
//	-this
func (s *Scalar) Negate() *Scalar { //nolint:revive
	s2 := NewScalar(s.value, s.modulo)
	s2.value.Sub(s2.modulo, s2.value)
	s2.value.Mod(s2.value, s2.modulo)

	return s2
}

// Inverse returns s2 such that condition is met
//
//	this * s2 = 1
func (s *Scalar) Inverse() *Scalar { //nolint:revive
	s2 := NewScalar(s.value, s.modulo)
	s2.value.ModInverse(s2.value, s2.modulo)

	return s2
}

// Equal return nil if this == s2.
func (s *Scalar) Equal(s2 *Scalar) error {
	err := s.equalMod(s2)
	if err != nil {
		return err
	}

	if s.value.Cmp(s2.value) != 0 {
		return fmt.Errorf("math/group: different scalar values")
	}

	return nil
}

// Returns nil if this.modulo == s2.modulo.
func (s *Scalar) equalMod(s2 *Scalar) error {
	if s.modulo.Cmp(s2.modulo) != 0 {
		return fmt.Errorf("math/group: different scalar modulos")
	}

	return nil
}

// ScalarValueOfReader reads a value from r, then returns a new group scalar with
// value=[1, mod). The value is deterministic unless r is randomness source.
func ScalarValueOfReader(r io.Reader, mod *big.Int) (*Scalar, error) {
	if mod.Cmp(new(big.Int).SetInt64(0)) != 1 {
		return nil, fmt.Errorf("math/group: scalar mod <= 0")
	}

	// If mod <= 0, then rand.Int will panic.
	// val=[0,mod)
	val, err := rand.Int(r, mod)
	if err != nil {
		return nil, err
	}

	// Don't allow val=0
	if val.Cmp(new(big.Int).SetInt64(0)) == 0 {
		val = new(big.Int).SetInt64(1)
	}

	return NewScalar(val, mod), nil
}

// Marshal ASN.1 marshals group scalar as
//
//	s ::= INTEGER
//
// where INTEGER is a group scalar value.
func (s *Scalar) Marshal() (der asn_1.DER, err error) { //nolint:revive
	var builder cryptobyte.Builder
	builder.AddASN1BigInt(s.value)

	return builder.Bytes()
}

// UnmarshalScalar ASN.1 unmarshalls group scalar.
func UnmarshalScalar(der asn_1.DER, mod *big.Int) (*Scalar, error) {
	integer := cryptobyte.String(der)
	val := new(big.Int)

	if !integer.ReadASN1Integer(val) {
		return nil, fmt.Errorf("math/group: scalar is not an ASN.1 INTEGER")
	}

	if !integer.Empty() {
		return nil, fmt.Errorf("math/group: incomplete ASN.1 scalar")
	}

	if val.Cmp(mod) >= 0 {
		return nil, fmt.Errorf("math/group: scalar >= mod")
	}

	return NewScalar(val, mod), nil
}
