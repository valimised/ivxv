package polynomial

import (
	"fmt"
	"testing"

	"tivi.io/core/math/group"
	_ "tivi.io/core/math/group/all"
)

func TestNewRandom(t *testing.T) {
	for _, g := range group.All() {
		for i := uint64(1); i < 11; i++ {
			t.Run(fmt.Sprintf("%s-degree-%d", g.Name(), i), func(t *testing.T) {
				poly, err := NewRandom(g, i)
				if err != nil {
					t.Error(err)
				}

				if poly.Degree() != i {
					t.Errorf("Expected degree=%v, got=%v\n", poly.Degree(), i)
				}
			})
		}
	}
}

func TestEvaluate(t *testing.T) {
	for _, g := range group.All() {
		for i := uint64(1); i < 11; i++ {
			t.Run(fmt.Sprintf("%s-degree-%d", g.Name(), i), func(t *testing.T) {
				poly1, err := NewRandom(g, i)
				if err != nil {
					t.Error(err)
				}

				poly2, err := NewRandom(g, i)
				if err != nil {
					t.Error(err)
				}

				s, err := group.RandomScalar(g.Order())
				if err != nil {
					t.Error(err)
				}

				ev1, err := poly1.Evaluate(s)
				if err != nil {
					t.Error(err)
				}

				ev2, err := poly2.Evaluate(s)
				if err != nil {
					t.Error(err)
				}

				// Randomness of polynomial coefficient is broken
				// and there is a chance that same polynomial will be given
				if err = ev1.Y.Equal(ev2.Y); err == nil {
					t.Error("Same polynomial, despite randomness was used")
				}
			})
		}
	}
}

func TestRandomWithConstantAndEvaluate(t *testing.T) {
	for _, g := range group.All() {
		for i := uint64(1); i < 11; i++ {
			t.Run(fmt.Sprintf("%s-degree-%d", g.Name(), i), func(t *testing.T) {
				constant, err := group.RandomScalar(g.Order())
				if err != nil {
					t.Error(err)
				}

				// p(x) = Ax^n + Bx^n-1 + Cx^n-2 + ... + constant
				poly, err := NewRandomWithConstant(g, i, constant)
				if err != nil {
					t.Error(err)
				}

				s := group.ZeroScalar(g.Order())

				// p(x) = constant + 0 + 0 + ... = constant
				ev, err := poly.Evaluate(s)
				if err != nil {
					t.Error(err)
				}

				// p(x) == constant
				err = ev.Y.Equal(constant)
				if err != nil {
					t.Error(err)
				}
			})
		}
	}
}

func TestInterpolate(t *testing.T) {
	for _, g := range group.All() {
		for i := uint64(1); i < 11; i++ {
			t.Run(fmt.Sprintf("%s-degree-%d", g.Name(), i), func(t *testing.T) {
				constant, err := group.RandomScalar(g.Order())
				if err != nil {
					t.Error(err)
				}

				// Polynomial with coefficients
				poly, err := NewRandomWithConstant(g, i, constant)
				if err != nil {
					t.Error(err)
				}

				// NB! Note that len(points) == i, and polynomial degree
				// is also == i, but for polynomial interpolation
				// len(points) should be i+1, i.e. degree+1
				points := make([]*Point, i)

				// Construct a polynomial using coefficients and coordinates
				for j := 0; j < len(points); j++ {
					// x-coordinate
					point, err := group.RandomScalar(g.Order())
					if err != nil {
						t.Error(err)
					}

					// y-coordinate
					points[j], err = poly.Evaluate(point)
					if err != nil {
						t.Error(err)
					}
				}

				// x=0
				s := group.ZeroScalar(g.Order())

				// interpolation is not correctly done, since degree of poly
				// is the same as len(points), there should always be at least
				// len(points) == degree+1
				interpolated, err := LagrangeInterpolate(g, points, s)
				if err != nil {
					t.Error(err)
				}

				// Poly is interpolated incorrectly, since not enough points given
				if interpolated.Y.Equal(constant) == nil {
					t.Error("Interpolation hasn't failed, but should")
				}

				// Another x-coordinate
				point, err := group.RandomScalar(g.Order())
				if err != nil {
					t.Error(err)
				}

				// Another y-coordinate
				ev, err := poly.Evaluate(point)
				if err != nil {
					t.Error(err)
				}

				points = append(points, ev)

				// Now, it should be correctly interpolated, since
				// len(points) == poly degree+1
				interpolated, err = LagrangeInterpolate(g, points, s)
				if err != nil {
					t.Error(err)
				}

				err = interpolated.Y.Equal(constant)
				if err != nil {
					t.Error(err)
				}
			})
		}
	}
}

func TestMarshal(t *testing.T) {
	for _, g := range group.All() {
		t.Run(string(g.Name()), func(t *testing.T) {
			poly, err := NewRandom(g, 3)
			if err != nil {
				t.Error(err)
			}

			data, err := poly.MarshalASN1()
			if err != nil {
				t.Error(err)
			}

			poly2, err := UnmarshalPolynomial(g, data)
			if err != nil {
				t.Error(err)
			}

			if len(poly2.Coefficients) != len(poly.Coefficients) {
				t.Errorf("Unequal polynomial lengths p1=%v, p2=%v\n", len(poly2.Coefficients), len(poly.Coefficients))
			}

			for i := range poly.Coefficients {
				err = poly.Coefficients[i].Equal(poly2.Coefficients[i])
				if err != nil {
					t.Error(err)
				}
			}

			sc, err := group.RandomScalar(g.Order())
			if err != nil {
				t.Error(err)
			}

			ev, err := poly.Evaluate(sc)
			if err != nil {
				t.Error(err)
			}

			data, err = ev.MarshalASN1()
			if err != nil {
				t.Error(err)
			}

			ev2, err := UnmarshalPoint(g, data)
			if err != nil {
				t.Error(err)
			}

			err = ev.X.Equal(ev2.X)
			if err != nil {
				t.Error(err)
			}

			err = ev.Y.Equal(ev2.Y)
			if err != nil {
				t.Error(err)
			}
		})
	}
}
