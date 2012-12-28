// Copyright ©2012 The bíogo.llrb Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package matrix

import (
	check "launchpad.net/gocheck"
	"math/rand"
	"reflect"
	"testing"
)

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

func panics(fn Panicker) (panicked bool) {
	defer func() {
		r := recover()
		panicked = r != nil
	}()
	Maybe(func() Matrix {
		mi, _ := reflect.ValueOf(fn).Call(nil)[0].Interface().(Matrix)
		return mi
	})
	return
}

func (s *S) TestMaybe(c *check.C) {
	for i, test := range []struct {
		fn     Panicker
		panics bool
	}{
		{
			func() Matrix { return nil },
			false,
		},
		{
			func() Matrix { panic("panic") },
			true,
		},
		{
			func() Matrix { panic(Error("panic")) },
			false,
		},
	} {
		c.Check(panics(test.fn), check.Equals, test.panics, check.Commentf("Test %d", i))
	}
}

func (s *S) TestNewDense(c *check.C) {
	for i, test := range []struct {
		a          [][]float64
		rows, cols int
		min, max   float64
		fro        float64
		mat        *Dense
	}{
		{
			[][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
			3, 3,
			0, 0,
			0,
			&Dense{rows: 3, cols: 3, matrix: denseRow{0, 0, 0, 0, 0, 0, 0, 0, 0}},
		},
		{
			[][]float64{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}},
			3, 3,
			1, 1,
			3,
			&Dense{rows: 3, cols: 3, matrix: denseRow{1, 1, 1, 1, 1, 1, 1, 1, 1}},
		},
		{
			[][]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}},
			3, 3,
			0, 1,
			1.7320508075688772,
			&Dense{rows: 3, cols: 3, matrix: denseRow{1, 0, 0, 0, 1, 0, 0, 0, 1}},
		},
		{
			[][]float64{{-1, 0, 0}, {0, -1, 0}, {0, 0, -1}},
			3, 3,
			-1, 0,
			1.7320508075688772,
			&Dense{rows: 3, cols: 3, matrix: denseRow{-1, 0, 0, 0, -1, 0, 0, 0, -1}},
		},
		{
			[][]float64{{1, 2, 3}, {4, 5, 6}},
			2, 3,
			1, 6,
			9.539392014169456,
			&Dense{rows: 2, cols: 3, matrix: denseRow{1, 2, 3, 4, 5, 6}},
		},
		{
			[][]float64{{1, 2}, {3, 4}, {5, 6}},
			3, 2,
			1, 6,
			9.539392014169456,
			&Dense{rows: 3, cols: 2, matrix: denseRow{1, 2, 3, 4, 5, 6}},
		},
	} {
		m, err := NewDense(test.a)
		c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
		rows, cols := m.Dims()
		c.Check(rows, check.Equals, test.rows, check.Commentf("Test %d", i))
		c.Check(cols, check.Equals, test.cols, check.Commentf("Test %d", i))
		c.Check(m.Min(), check.Equals, test.min, check.Commentf("Test %d", i))
		c.Check(m.Max(), check.Equals, test.max, check.Commentf("Test %d", i))
		c.Check(m.Norm(Fro), check.Equals, test.fro, check.Commentf("Test %d", i))
		c.Check(m, check.DeepEquals, test.mat, check.Commentf("Test %d", i))
		c.Check(m.Equals(test.mat), check.Equals, true, check.Commentf("Test %d", i))
	}
}

func (s *S) TestNewSparse(c *check.C) {
	for i, test := range []struct {
		a            [][]float64
		rows, cols   int
		min, max     float64
		minnz, maxnz float64
		fro          float64
		mat          *Sparse
	}{
		{
			[][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
			3, 3,
			0, 0,
			0, 0,
			0,
			&Sparse{rows: 3, cols: 3, matrix: []sparseRow{[]sparseElem(nil), []sparseElem(nil), []sparseElem(nil)}},
		},
		{
			[][]float64{{0, 0, 0}, {0, 0, 1}, {0, 0, 0}},
			3, 3,
			0, 1,
			1, 1,
			1,
			&Sparse{rows: 3, cols: 3, matrix: []sparseRow{[]sparseElem(nil), []sparseElem{sparseElem{index: 2, value: 1}}, []sparseElem(nil)}},
		},
		{
			[][]float64{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}},
			3, 3,
			1, 1,
			1, 1,
			3,
			&Sparse{rows: 3, cols: 3, matrix: []sparseRow{
				sparseRow{sparseElem{index: 0, value: 1}, sparseElem{index: 1, value: 1}, sparseElem{index: 2, value: 1}},
				sparseRow{sparseElem{index: 0, value: 1}, sparseElem{index: 1, value: 1}, sparseElem{index: 2, value: 1}},
				sparseRow{sparseElem{index: 0, value: 1}, sparseElem{index: 1, value: 1}, sparseElem{index: 2, value: 1}}},
			},
		},
		{
			[][]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}},
			3, 3,
			0, 1,
			1, 1,
			1.7320508075688772,
			&Sparse{rows: 3, cols: 3, matrix: []sparseRow{
				sparseRow{sparseElem{index: 0, value: 1}},
				sparseRow{sparseElem{index: 1, value: 1}},
				sparseRow{sparseElem{index: 2, value: 1}}},
			},
		},
		{
			[][]float64{{-1, 0, 0}, {0, -1, 0}, {0, 0, -1}},
			3, 3,
			-1, 0,
			-1, -1,
			1.7320508075688772,
			&Sparse{rows: 3, cols: 3, matrix: []sparseRow{
				sparseRow{sparseElem{index: 0, value: -1}},
				sparseRow{sparseElem{index: 1, value: -1}},
				sparseRow{sparseElem{index: 2, value: -1}}},
			},
		},
		{
			[][]float64{{1, 2, 3}, {4, 5, 6}},
			2, 3,
			1, 6,
			1, 6,
			9.539392014169456,
			&Sparse{rows: 2, cols: 3, matrix: []sparseRow{
				sparseRow{sparseElem{index: 0, value: 1}, sparseElem{index: 1, value: 2}, sparseElem{index: 2, value: 3}},
				sparseRow{sparseElem{index: 0, value: 4}, sparseElem{index: 1, value: 5}, sparseElem{index: 2, value: 6}}},
			},
		},
		{
			[][]float64{{1, 2}, {3, 4}, {5, 6}},
			3, 2,
			1, 6,
			1, 6,
			9.539392014169456,
			&Sparse{rows: 3, cols: 2, matrix: []sparseRow{
				sparseRow{sparseElem{index: 0, value: 1}, sparseElem{index: 1, value: 2}},
				sparseRow{sparseElem{index: 0, value: 3}, sparseElem{index: 1, value: 4}},
				sparseRow{sparseElem{index: 0, value: 5}, sparseElem{index: 1, value: 6}}},
			},
		},
	} {
		m, err := NewSparse(test.a)
		c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
		rows, cols := m.Dims()
		c.Check(rows, check.Equals, test.rows, check.Commentf("Test %d", i))
		c.Check(cols, check.Equals, test.cols, check.Commentf("Test %d", i))
		c.Check(m.Min(), check.Equals, test.min, check.Commentf("Test %d", i))
		c.Check(m.Max(), check.Equals, test.max, check.Commentf("Test %d", i))
		c.Check(m.MinNonZero(), check.Equals, test.minnz, check.Commentf("Test %d", i))
		c.Check(m.MaxNonZero(), check.Equals, test.maxnz, check.Commentf("Test %d", i))
		c.Check(m.Norm(Fro), check.Equals, test.fro, check.Commentf("Test %d", i))
		c.Check(m, check.DeepEquals, test.mat, check.Commentf("Test %d", i))
		c.Check(m.Equals(test.mat), check.Equals, true, check.Commentf("Test %d", i))
	}
}

func (s *S) TestAdd(c *check.C) {
	for i, test := range []struct {
		a, b, r [][]float64
	}{
		{
			[][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
			[][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
			[][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
		},
		{
			[][]float64{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}},
			[][]float64{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}},
			[][]float64{{2, 2, 2}, {2, 2, 2}, {2, 2, 2}},
		},
		{
			[][]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}},
			[][]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}},
			[][]float64{{2, 0, 0}, {0, 2, 0}, {0, 0, 2}},
		},
		{
			[][]float64{{-1, 0, 0}, {0, -1, 0}, {0, 0, -1}},
			[][]float64{{-1, 0, 0}, {0, -1, 0}, {0, 0, -1}},
			[][]float64{{-2, 0, 0}, {0, -2, 0}, {0, 0, -2}},
		},
		{
			[][]float64{{1, 2, 3}, {4, 5, 6}},
			[][]float64{{1, 2, 3}, {4, 5, 6}},
			[][]float64{{2, 4, 6}, {8, 10, 12}},
		},
	} {
		{
			a, err := NewSparse(test.a)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			b, err := NewSparse(test.b)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			r, err := NewSparse(test.r)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			c.Check(a.Add(b).Equals(r), check.Equals, true, check.Commentf("Test %d: %v add %v", i, test.a, test.b))
		}

		{
			a, err := NewDense(test.a)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			b, err := NewDense(test.b)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			r, err := NewDense(test.r)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			c.Check(a.Add(b).Equals(r), check.Equals, true, check.Commentf("Test %d: %v add %v", i, test.a, test.b))
		}
	}
}

func (s *S) TestSub(c *check.C) {
	for i, test := range []struct {
		a, b, r [][]float64
	}{
		{
			[][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
			[][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
			[][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
		},
		{
			[][]float64{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}},
			[][]float64{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}},
			[][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
		},
		{
			[][]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}},
			[][]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}},
			[][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
		},
		{
			[][]float64{{-1, 0, 0}, {0, -1, 0}, {0, 0, -1}},
			[][]float64{{-1, 0, 0}, {0, -1, 0}, {0, 0, -1}},
			[][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
		},
		{
			[][]float64{{1, 2, 3}, {4, 5, 6}},
			[][]float64{{1, 2, 3}, {4, 5, 6}},
			[][]float64{{0, 0, 0}, {0, 0, 0}},
		},
	} {
		{
			a, err := NewSparse(test.a)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			b, err := NewSparse(test.b)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			r, err := NewSparse(test.r)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			c.Check(a.Sub(b).Equals(r), check.Equals, true, check.Commentf("Test %d: %v sub %v", i, test.a, test.b))
		}

		{
			a, err := NewDense(test.a)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			b, err := NewDense(test.b)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			r, err := NewDense(test.r)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			c.Check(a.Sub(b).Equals(r), check.Equals, true, check.Commentf("Test %d: %v sub %v", i, test.a, test.b))
		}
	}
}

func (s *S) TestMulElem(c *check.C) {
	for i, test := range []struct {
		a, b, r [][]float64
	}{
		{
			[][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
			[][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
			[][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
		},
		{
			[][]float64{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}},
			[][]float64{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}},
			[][]float64{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}},
		},
		{
			[][]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}},
			[][]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}},
			[][]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}},
		},
		{
			[][]float64{{-1, 0, 0}, {0, -1, 0}, {0, 0, -1}},
			[][]float64{{-1, 0, 0}, {0, -1, 0}, {0, 0, -1}},
			[][]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}},
		},
		{
			[][]float64{{1, 2, 3}, {4, 5, 6}},
			[][]float64{{1, 2, 3}, {4, 5, 6}},
			[][]float64{{1, 4, 9}, {16, 25, 36}},
		},
	} {
		{
			a, err := NewSparse(test.a)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			b, err := NewSparse(test.b)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			r, err := NewSparse(test.r)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			c.Check(a.MulElem(b).Equals(r), check.Equals, true, check.Commentf("Test %d: %v mulelem %v", i, test.a, test.b))
		}

		{
			a, err := NewDense(test.a)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			b, err := NewDense(test.b)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			r, err := NewDense(test.r)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			c.Check(a.MulElem(b).Equals(r), check.Equals, true, check.Commentf("Test %d: %v dot %v", i, test.a, test.b))
		}
	}
}

func (s *S) TestDot(c *check.C) {
	for i, test := range []struct {
		a, b, r [][]float64
	}{
		{
			[][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
			[][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
			[][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
		},
		{
			[][]float64{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}},
			[][]float64{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}},
			[][]float64{{3, 3, 3}, {3, 3, 3}, {3, 3, 3}},
		},
		{
			[][]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}},
			[][]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}},
			[][]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}},
		},
		{
			[][]float64{{-1, 0, 0}, {0, -1, 0}, {0, 0, -1}},
			[][]float64{{-1, 0, 0}, {0, -1, 0}, {0, 0, -1}},
			[][]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}},
		},
		{
			[][]float64{{1, 2, 3}, {4, 5, 6}},
			[][]float64{{1, 2}, {3, 4}, {5, 6}},
			[][]float64{{22, 28}, {49, 64}},
		},
	} {
		{
			a, err := NewSparse(test.a)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			b, err := NewSparse(test.b)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			r, err := NewSparse(test.r)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			c.Check(a.Dot(b).Equals(r), check.Equals, true, check.Commentf("Test %d: %v dot %v", i, test.a, test.b))
		}

		{
			a, err := NewDense(test.a)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			b, err := NewDense(test.b)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			r, err := NewDense(test.r)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			c.Check(a.Dot(b).Equals(r), check.Equals, true, check.Commentf("Test %d: %v dot %v", i, test.a, test.b))
		}
	}
}

func (s *S) TestLU(c *check.C) {
	for _, fns := range []struct {
		I func(int) (Matrix, error)
		R func(sized int, d float64, rnd func() float64) (Matrix, error)
	}{
		{
			I: func(size int) (Matrix, error) {
				return IdentitySparse(size)
			},
			R: func(size int, d float64, rnd func() float64) (Matrix, error) {
				return FuncSparse(size, size, d, rnd)
			},
		},
		{
			I: func(size int) (Matrix, error) {
				return IdentityDense(size)
			},
			R: func(size int, d float64, rnd func() float64) (Matrix, error) {
				return FuncDense(size, size, d, rnd)
			},
		},
	} {
		for i := 0; i < 100; i++ {
			size := rand.Intn(100)
			I, err := fns.I(size)
			if size == 0 {
				c.Check(err, check.Equals, ErrZeroLength)
			} else {
				c.Check(err, check.Equals, nil)
			}
			r, err := fns.R(size, rand.Float64(), rand.NormFloat64)
			if size == 0 {
				c.Check(err, check.Equals, ErrZeroLength)
				continue
			}
			c.Check(err, check.Equals, nil)
			u := r.U()
			l := r.L()
			d := r.MulElem(I)
			c.Check(r.Equals(u.Add(l.Sub(d))), check.Equals, true, check.Commentf("Test %d", i))
			c.Check(d.Equals(l.MulElem(I)), check.Equals, true, check.Commentf("Test %d", i))
			c.Check(d.Equals(u.MulElem(I)), check.Equals, true, check.Commentf("Test %d", i))
		}
	}
}

func (s *S) TestTranspose(c *check.C) {
	for i, test := range []struct {
		a, t [][]float64
	}{
		{
			[][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
			[][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
		},
		{
			[][]float64{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}},
			[][]float64{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}},
		},
		{
			[][]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}},
			[][]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}},
		},
		{
			[][]float64{{-1, 0, 0}, {0, -1, 0}, {0, 0, -1}},
			[][]float64{{-1, 0, 0}, {0, -1, 0}, {0, 0, -1}},
		},
		{
			[][]float64{{1, 2, 3}, {4, 5, 6}},
			[][]float64{{1, 4}, {2, 5}, {3, 6}},
		},
	} {
		{
			a, err := NewSparse(test.a)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			t, err := NewSparse(test.t)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			c.Check(a.T().Equals(t), check.Equals, true, check.Commentf("Test %d: %v transpose = %v", i, test.a, test.t))
			c.Check(a.T().T().Equals(a), check.Equals, true, check.Commentf("Test %d: %v transpose = I", i, test.a))
		}

		{
			a, err := NewDense(test.a)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			t, err := NewDense(test.t)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			c.Check(a.T().Equals(t), check.Equals, true, check.Commentf("Test %d: %v transpose = %v", i, test.a, test.t))
			c.Check(a.T().T().Equals(a), check.Equals, true, check.Commentf("Test %d: %v transpose = I", i, test.a, test.t))
		}
	}
}

func (s *S) TestStackAugment(c *check.C) {
	for i, test := range []struct {
		a, b     [][]float64
		aug      [][]float64
		augErr   error
		stack    [][]float64
		stackErr error
	}{
		{
			[][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
			[][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
			[][]float64{{0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}}, nil,
			[][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}, {0, 0, 0}, {0, 0, 0}, {0, 0, 0}}, nil,
		},
		{
			[][]float64{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}},
			[][]float64{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}},
			[][]float64{{1, 1, 1, 1, 1, 1}, {1, 1, 1, 1, 1, 1}, {1, 1, 1, 1, 1, 1}}, nil,
			[][]float64{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}, {1, 1, 1}, {1, 1, 1}, {1, 1, 1}}, nil,
		},
		{
			[][]float64{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}},
			[][]float64{{-1, -2, -3}, {-4, -5, -6}, {-7, -8, -9}},
			[][]float64{{1, 2, 3, -1, -2, -3}, {4, 5, 6, -4, -5, -6}, {7, 8, 9, -7, -8, -9}}, nil,
			[][]float64{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}, {-1, -2, -3}, {-4, -5, -6}, {-7, -8, -9}}, nil,
		},
		{
			[][]float64{{1, 2, 3}, {4, 5, 6}},
			[][]float64{{1, 2}, {3, 4}, {5, 6}},
			nil, ErrColLength,
			nil, ErrRowLength,
		},
	} {
		{
			a, err := NewSparse(test.a)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			b, err := NewSparse(test.b)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))

			var aug *Sparse
			if test.aug != nil {
				aug, err = NewSparse(test.aug)
				c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			}
			augr, err := Maybe(func() Matrix { return a.Augment(b) })
			if err == nil {
				c.Check(augr.Equals(aug), check.Equals, true, check.Commentf("Test %d: %v augment %v", i, test.a, test.b))
			} else {
				c.Check(err, check.Equals, test.augErr)
			}

			var stack *Sparse
			if test.stack != nil {
				stack, err = NewSparse(test.stack)
				c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			}
			stackr, err := Maybe(func() Matrix { return a.Stack(b) })
			if err == nil {
				c.Check(stackr.Equals(stack), check.Equals, true, check.Commentf("Test %d: %v stack %v", i, test.a, test.b))
			} else {
				c.Check(err, check.Equals, test.stackErr)
			}
		}

		{
			a, err := NewDense(test.a)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			b, err := NewDense(test.b)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))

			var aug *Dense
			if test.aug != nil {
				aug, err = NewDense(test.aug)
				c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			}
			augr, err := Maybe(func() Matrix { return a.Augment(b) })
			if err == nil {
				c.Check(augr.Equals(aug), check.Equals, true, check.Commentf("Test %d: %v augment %v", i, test.a, test.b))
			} else {
				c.Check(err, check.Equals, test.augErr)
			}

			var stack *Dense
			if test.stack != nil {
				stack, err = NewDense(test.stack)
				c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			}
			stackr, err := Maybe(func() Matrix { return a.Stack(b) })
			if err == nil {
				c.Check(stackr.Equals(stack), check.Equals, true, check.Commentf("Test %d: %v stack %v", i, test.a, test.b))
			} else {
				c.Check(err, check.Equals, test.stackErr)
			}
		}
	}
}

// TODO:
// a.MinAxis(matrix.Cols))
// a.MaxAxis(matrix.Cols))
// a.MinAxis(matrix.Rows))
// a.MaxAxis(matrix.Rows))
// a.SumAxis(matrix.Cols))
// a.SumAxis(matrix.Rows))

func BenchmarkDotDense100Half(b *testing.B) {
	b.StopTimer()
	a := Must(FuncDense(100, 100, 0.5, rand.NormFloat64))
	d := Must(FuncDense(100, 100, 0.5, rand.NormFloat64))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = a.Dot(d)
	}
}

func BenchmarkDotSparse100Half(b *testing.B) {
	b.StopTimer()
	a := Must(FuncSparse(100, 100, 0.5, rand.NormFloat64))
	d := Must(FuncSparse(100, 100, 0.5, rand.NormFloat64))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = a.Dot(d)
	}
}

func BenchmarkDotDense100Tenth(b *testing.B) {
	b.StopTimer()
	a := Must(FuncDense(100, 100, 0.1, rand.NormFloat64))
	d := Must(FuncDense(100, 100, 0.1, rand.NormFloat64))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = a.Dot(d)
	}
}

func BenchmarkDotSparse100Tenth(b *testing.B) {
	b.StopTimer()
	a := Must(FuncSparse(100, 100, 0.1, rand.NormFloat64))
	d := Must(FuncSparse(100, 100, 0.1, rand.NormFloat64))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = a.Dot(d)
	}
}

func BenchmarkDotDense1000Half(b *testing.B) {
	b.StopTimer()
	a := Must(FuncDense(1000, 1000, 0.5, rand.NormFloat64))
	d := Must(FuncDense(1000, 1000, 0.5, rand.NormFloat64))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = a.Dot(d)
	}
}

func BenchmarkDotSparse1000Half(b *testing.B) {
	b.StopTimer()
	a := Must(FuncSparse(1000, 1000, 0.5, rand.NormFloat64))
	d := Must(FuncSparse(1000, 1000, 0.5, rand.NormFloat64))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = a.Dot(d)
	}
}

func BenchmarkDotDense1000Tenth(b *testing.B) {
	b.StopTimer()
	a := Must(FuncDense(1000, 1000, 0.1, rand.NormFloat64))
	d := Must(FuncDense(1000, 1000, 0.1, rand.NormFloat64))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = a.Dot(d)
	}
}

func BenchmarkDotSparse1000Tenth(b *testing.B) {
	b.StopTimer()
	a := Must(FuncSparse(1000, 1000, 0.1, rand.NormFloat64))
	d := Must(FuncSparse(1000, 1000, 0.1, rand.NormFloat64))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = a.Dot(d)
	}
}

func BenchmarkDotDense1000Hundredth(b *testing.B) {
	b.StopTimer()
	a := Must(FuncDense(1000, 1000, 0.01, rand.NormFloat64))
	d := Must(FuncDense(1000, 1000, 0.01, rand.NormFloat64))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = a.Dot(d)
	}
}

func BenchmarkDotSparse1000Hundredth(b *testing.B) {
	b.StopTimer()
	a := Must(FuncSparse(1000, 1000, 0.01, rand.NormFloat64))
	d := Must(FuncSparse(1000, 1000, 0.01, rand.NormFloat64))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = a.Dot(d)
	}
}
