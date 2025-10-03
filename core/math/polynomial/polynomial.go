/*
Package polynomial provides an implementation of univariate algebraic
polynomials. An univariate polynomial is a function in the form

	p(x) = a_0 + a_1 x + a_2 x^2 + ... + a_n x^n

In this case, we say that the degree of the polynomial p is n.

By evaluating the polynomial at a point x, we obtain the value

	y <- p(x)

Using Lagrange interpolation, it is possible to recover the polynomial if n+1
values are known.
*/
package polynomial

import (
	"math/big"
	asn_1 "tivi.io/core/encoding/asn1"

	"golang.org/x/crypto/cryptobyte"
	"golang.org/x/crypto/cryptobyte/asn1"

	"tivi.io/core/math/group"
)

// Point is the polynomial point (x,y) == (x, p(x)).
type Point struct {
	// x
	X *group.Scalar
	// p(x)
	Y *group.Scalar
}

// polynomialPoints is a helper type which you can use to transform from []Point
// to []group.Scalar, and therefore being able to use xCoords() method to
// obtain only Point.X coordinates of a polynomial.
type polynomialPoints []*Point

func (points polynomialPoints) xCoords() []*group.Scalar {
	xCoords := make([]*group.Scalar, len(points))
	for i := range points {
		xCoords[i] = points[i].X
	}
	return xCoords
}

// ASN1Marshal marshals a polynomial point
//
//	e ::= SEQUENCE {
//		INTEGER
//		INTEGER
//	}
//
// where first INTEGER is an x-coordinate and second is y-coordinate.
func (e *Point) MarshalASN1() ([]byte, error) {
	xcoord, err := e.X.Marshal()
	if err != nil {
		return nil, ASN1MarshalPointASN1MarshalXError{Err: err}
	}

	ycoord, err := e.Y.Marshal()
	if err != nil {
		return nil, ASN1MarshalPointASN1MarshalYError{Err: err}
	}

	var c cryptobyte.Builder
	c.AddASN1(asn1.SEQUENCE, func(c *cryptobyte.Builder) {
		c.AddBytes(xcoord)
		c.AddBytes(ycoord)
	})

	b, err := c.Bytes()
	if err != nil {
		return nil, ASN1MarshalPointError{Err: err}
	}

	return b, nil
}

// UnmarshalPoint unmarshalls polynomial point.
func UnmarshalPoint(g group.Group, data []byte) (*Point, error) {
	c := cryptobyte.String(data)
	var inner cryptobyte.String
	var xB cryptobyte.String
	var yB cryptobyte.String

	if !c.ReadASN1(&inner, asn1.SEQUENCE) {
		return nil, UnmarshalPointReadASN1Error{}
	}

	if !inner.ReadAnyASN1Element(&xB, nil) {
		return nil, UnmarshalPointReadAnyASN1ElementXError{}
	}

	if !inner.ReadAnyASN1Element(&yB, nil) {
		return nil, UnmarshalPointReadAnyASN1ElementYError{}
	}

	x, err := group.UnmarshalScalar(asn_1.DER(xB), g.Order())
	if err != nil {
		return nil, UnmarshalPointASN1UnmarshalScalarXError{Err: err}
	}

	y, err := group.UnmarshalScalar(asn_1.DER(yB), g.Order())
	if err != nil {
		return nil, UnmarshalPointASN1UnmarshalScalarYError{Err: err}
	}

	if !c.Empty() || !inner.Empty() {
		return nil, UnmarshalPointTrailingBytesError{Err: err}
	}

	return &Point{
		X: x,
		Y: y,
	}, nil
}

type Polynomial struct {
	Group group.Group
	//	5x^3 + 6x^2 + 8x + 1
	//
	// where Coefficients = {5, 6, 8}
	Coefficients []*group.Scalar
}

// NewRandom generates a polynomial, that has exactly degree+1 points.
// All polynomial coefficients are chosen randomly from a set of [0, p),
// where p is the group order.
func NewRandom(g group.Group, degree uint64) (*Polynomial, error) {
	coeffs := make([]*group.Scalar, degree+1)

	for i := range coeffs {
		coeff, err := group.RandomScalar(g.Order())
		if err != nil {
			return nil, NewRandomRandomScalarError{Err: err}
		}
		coeffs[i] = coeff
	}

	return &Polynomial{
		Group:        g,
		Coefficients: coeffs,
	}, nil
}

// NewRandomWithConstant generates a new polynomial where coefficient with degree=0
// is given by constant, e.g.
//
//	p(x) = Ax^3 + Bx^2 + Cx + constant
//
// which in turn means that if x=0, then p(x) = constant.
//
// Rest of coefficients are chosen randomly from a set of [0, p), where p is
// the group order.
func NewRandomWithConstant(g group.Group, degree uint64, constant *group.Scalar) (*Polynomial, error) {
	coeffs := make([]*group.Scalar, degree+1)

	coeff := group.NewScalar(constant.Value(), constant.Modulo())
	coeffs[0] = coeff

	for i := 1; i < len(coeffs); i++ {
		coeff, err := group.RandomScalar(g.Order())
		if err != nil {
			return nil, NewRandomWithFreeRandomScalarError{Err: err}
		}
		coeffs[i] = coeff
	}

	return &Polynomial{
		Group:        g,
		Coefficients: coeffs,
	}, nil
}

// Degree returns the degree of the polynomial.
func (p *Polynomial) Degree() uint64 {
	if len(p.Coefficients) == 0 {
		// if no Coefficients, then assume that constant coefficient is zero
		return 0
	}
	return uint64(len(p.Coefficients) - 1)
}

// Coefficient returns the value of the i-th coefficient.
func (p *Polynomial) Coefficient(i uint64) *group.Scalar {
	if i > p.Degree() {
		return group.ZeroScalar(p.Group.Order())
	}
	return p.Coefficients[i]
}

// Evaluate evaluates the polynomial
//
//	p(x) = a_0 + a_1 * x^1 + a_2 * x^2 + ... + a_i * x^i
//
// Note that if x=0, then p(x) = a_0
func (p *Polynomial) Evaluate(x *group.Scalar) (*Point, error) {
	if len(p.Coefficients) == 0 {
		// (x, a_0)
		return &Point{X: x, Y: group.ZeroScalar(p.Group.Order())}, nil
	}

	var err error

	// Initially it is x_i = x^0
	var x_i *group.Scalar //nolint:revive
	// a_i * x_i
	var mul_ai_xi *group.Scalar //nolint:revive
	// p(x) = a_0
	pX := group.NewScalar(p.Coefficients[0].Value(), p.Coefficients[0].Modulo())

	// p(x) = a_0 + a_1 * x^1 + a_2 * x^2 + ... + a_i * x^i
	for i := uint64(1); i <= p.Degree(); i++ {
		// x^i
		x_i = x.Exp(new(big.Int).SetUint64(i))

		// a_i * x^i
		mul_ai_xi, err = x_i.Mul(p.Coefficients[i])
		if err != nil {
			return nil, EvaluatePolynomialMulError{Err: err}
		}

		// p(x) = p(x) + (a_i * x^i)
		pX, err = pX.Add(mul_ai_xi)
		if err != nil {
			return nil, EvaluatePolynomialAddError{Err: err}
		}
	}

	return &Point{X: x, Y: pX}, nil
}

// ASN1Marshal marshals polynomial
//
//	p ::= SEQUENCE {
//		INTEGER
//		INTEGER
//		...
//	}
//
// where INTEGER is a polynomial coefficient and amount of INTEGERs depending on
// a polynomial degree.
func (p *Polynomial) MarshalASN1() ([]byte, error) {
	coeffs := make([][]byte, 0, len(p.Coefficients))

	for _, coeff := range p.Coefficients {
		b, err := coeff.Marshal()
		if err != nil {
			return nil, ASN1MarshalPolynomialASN1MarshalCoefficientError{Err: err}
		}
		coeffs = append(coeffs, b)
	}

	var c cryptobyte.Builder

	c.AddASN1(asn1.SEQUENCE, func(c *cryptobyte.Builder) {
		for _, coeff := range coeffs {
			c.AddBytes(coeff)
		}
	})

	b, err := c.Bytes()
	if err != nil {
		return nil, ASN1MarshalPolynomialError{Err: err}
	}

	return b, nil
}

// UnmarshalPolynomial unmarshalls a polynomial.
func UnmarshalPolynomial(g group.Group, data []byte) (*Polynomial, error) {
	var coeffs []*group.Scalar
	c := cryptobyte.String(data)
	var inner cryptobyte.String

	if !c.ReadASN1(&inner, asn1.SEQUENCE) {
		return nil, UnmarshalPolynomialReadASN1Error{}
	}
	for !inner.Empty() {
		var coefdata cryptobyte.String

		if !inner.ReadAnyASN1Element(&coefdata, nil) {
			return nil, UnmarshalPolynomialReadAnyASN1ElementError{}
		}

		coeff, err := group.UnmarshalScalar(asn_1.DER(coefdata), g.Order())
		if err != nil {
			return nil, UnmarshalPolynomialUnmarshalScalarError{Err: err}
		}

		coeffs = append(coeffs, coeff)
	}

	if !c.Empty() {
		return nil, UnmarshalPolynomialTrailingBytesError{}
	}

	return &Polynomial{
		Group:        g,
		Coefficients: coeffs,
	}, nil
}

// LagrangeBasisPolynomial checks whether polynomial passes x-coordinate by using
// following Lagrange basis polynomial equation
//
//	l_i(x) = (x - j_0) / (i - j_0) * (x - j_1) / (i - j_1) * ... * (x - j_n) / (i - j_n)
//
// where x is an x-coordinate of a point, for which we try to evaluate Lagrange
// basis polynomial, i.e. to understand, whether that polynomial passes this
// particular x-coordinate or not.
//
// i is an x-coordinate of a polynomial point that we try to evaluate Lagrange
// basis polynomial for. Recall, we evaluate Lagrange basis polynomial for each
// i, which means that we call this function exactly j_n times in order to
// evaluate all i.
//
// j_0, j_1, ..., j_n are all x-coordinates where polynomial passes points through
func LagrangeBasisPolynomial(g group.Group, x *group.Scalar, points []*group.Scalar, i *group.Scalar) (*group.Scalar, error) {
	var err error
	var l_i *group.Scalar //nolint:revive
	// At start, (x - j_0) == (x - j_n)
	nom := group.OneScalar(g.Order())
	// At start, (i - j_0) == (i - j_n)
	denom := group.OneScalar(g.Order())

	// Loop over all polynomial points, starting with n=0
	for _, j_n := range points { //nolint:revive

		// i == j_n will result in (x - j_n) / (i - j_n), where (i - j_n) == 0,
		// division by zero is prohibited!
		err = i.Equal(j_n)
		if err == nil {
			continue
		}

		// (x - j_n+1)
		l_i, err = x.Sub(j_n)
		if err != nil {
			return nil, LagrangeBasisPolynomialSubNomError{Err: err}
		}

		// (x - j_n) * (x - j_n+1)
		nom, err = nom.Mul(l_i)
		if err != nil {
			return nil, LagrangeBasisPolynomialMulNomError{Err: err}
		}

		// (i - j_n+1)
		l_i, err = i.Sub(j_n)
		if err != nil {
			return nil, LagrangeBasisPolynomialSubDenomError{Err: err}
		}

		// (i - j_n) * (i - j_n+1)
		denom, err = denom.Mul(l_i)
		if err != nil {
			return nil, LagrangeBasisPolynomialMulDenomError{Err: err}
		}
	}

	// denom^-1 == 1/denom
	denom = denom.Inverse()

	// nom * denom
	l_i, err = nom.Mul(denom)
	if err != nil {
		return nil, LagrangeBasisPolynomialMulNomAndDenomError{Err: err}
	}

	return l_i, nil
}

// LagrangeInterpolate computes the value of a polynomial at x-coordinate, using
// Lagrange interpolation
//
//	L(x) = f(i_0) * l_i_0(x) + f(i_1) * l_i_1(x) + ... + f(i_n) * l_i_n(x)
//
// where (i, f(i_n)) are (x,y) coordinates of a polynomial point
//
// This function will loop until all polynomial points are interpolated as shown
// in the equation above. If less than degree+1 evaluations are given, then the
// interpolated value is uniformly random.
func LagrangeInterpolate(g group.Group, points []*Point, x *group.Scalar) (*Point, error) {
	// x is found, it is already located at polynomial
	for _, i := range points {
		err := i.X.Equal(x)
		if err == nil {
			return i, nil
		}
	}

	// At start, L_x = 0
	L_x := group.ZeroScalar(g.Order()) //nolint:revive

	// Loop over all polynomial points
	for _, i := range points {
		l_i, err := LagrangeBasisPolynomial(g, x, polynomialPoints(points).xCoords(), i.X) //nolint:revive
		if err != nil {
			return nil, LagrangeInterpolateLagrangeBasisPolynomial{Err: err}
		}

		// (l_i * i)
		l_i, err = l_i.Mul(i.Y)
		if err != nil {
			return nil, LagrangeInterpolateMulError{Err: err}
		}

		// L_x += (l_i * i)
		L_x, err = L_x.Add(l_i)
		if err != nil {
			return nil, LagrangeInterpolateAddPolynomial{Err: err}
		}
	}

	return &Point{X: x, Y: L_x}, nil
}
