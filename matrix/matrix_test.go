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
	var (
		tempSparse *Sparse
		tempDense  *Dense
	)
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
			c.Check(a.Add(b, nil).Equals(r), check.Equals, true, check.Commentf("Test %d: %v add %v", i, test.a, test.b))
			c.Check(a.Add(b, tempSparse).Equals(r), check.Equals, true, check.Commentf("Test %d: %v add %v", i, test.a, test.b))
			t := a.CloneSparse(nil)
			c.Check(a.Add(b, a).Equals(r), check.Equals, true, check.Commentf("Test %d: %v sub %v", i, test.a, test.b))
			a = t
			c.Check(a.Add(b, b).Equals(r), check.Equals, true, check.Commentf("Test %d: %v sub %v", i, test.a, test.b))
		}

		{
			a, err := NewDense(test.a)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			b, err := NewDense(test.b)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			r, err := NewDense(test.r)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			c.Check(a.Add(b, nil).Equals(r), check.Equals, true, check.Commentf("Test %d: %v add %v", i, test.a, test.b))
			c.Check(a.Add(b, tempDense).Equals(r), check.Equals, true, check.Commentf("Test %d: %v add %v", i, test.a, test.b))
			t := a.CloneDense(nil)
			c.Check(a.Add(b, a).Equals(r), check.Equals, true, check.Commentf("Test %d: %v sub %v", i, test.a, test.b))
			a = t
			c.Check(a.Add(b, b).Equals(r), check.Equals, true, check.Commentf("Test %d: %v sub %v", i, test.a, test.b))
		}
	}
}

func (s *S) TestSub(c *check.C) {
	var (
		tempSparse *Sparse
		tempDense  *Dense
	)
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
			c.Check(a.Sub(b, nil).Equals(r), check.Equals, true, check.Commentf("Test %d: %v sub %v", i, test.a, test.b))
			c.Check(a.Sub(b, tempSparse).Equals(r), check.Equals, true, check.Commentf("Test %d: %v sub %v", i, test.a, test.b))
			t := a.CloneSparse(nil)
			c.Check(a.Sub(b, a).Equals(r), check.Equals, true, check.Commentf("Test %d: %v sub %v", i, test.a, test.b))
			a = t
			c.Check(a.Sub(b, b).Equals(r), check.Equals, true, check.Commentf("Test %d: %v sub %v", i, test.a, test.b))
		}

		{
			a, err := NewDense(test.a)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			b, err := NewDense(test.b)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			r, err := NewDense(test.r)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			c.Check(a.Sub(b, nil).Equals(r), check.Equals, true, check.Commentf("Test %d: %v sub %v", i, test.a, test.b))
			c.Check(a.Sub(b, tempDense).Equals(r), check.Equals, true, check.Commentf("Test %d: %v sub %v", i, test.a, test.b))
			t := a.CloneDense(nil)
			c.Check(a.Sub(b, a).Equals(r), check.Equals, true, check.Commentf("Test %d: %v sub %v", i, test.a, test.b))
			a = t
			c.Check(a.Sub(b, b).Equals(r), check.Equals, true, check.Commentf("Test %d: %v sub %v", i, test.a, test.b))
		}
	}
}

func (s *S) TestMulElem(c *check.C) {
	var (
		tempSparse *Sparse
		tempDense  *Dense
	)
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
			c.Check(a.MulElem(b, nil).Equals(r), check.Equals, true, check.Commentf("Test %d: %v mulelem %v", i, test.a, test.b))
			c.Check(a.MulElem(b, tempSparse).Equals(r), check.Equals, true, check.Commentf("Test %d: %v mulelem %v", i, test.a, test.b))
			t := a.CloneSparse(nil)
			c.Check(a.MulElem(b, a).Equals(r), check.Equals, true, check.Commentf("Test %d: %v sub %v", i, test.a, test.b))
			a = t
			c.Check(a.MulElem(b, b).Equals(r), check.Equals, true, check.Commentf("Test %d: %v sub %v", i, test.a, test.b))
		}

		{
			a, err := NewDense(test.a)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			b, err := NewDense(test.b)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			r, err := NewDense(test.r)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			c.Check(a.MulElem(b, nil).Equals(r), check.Equals, true, check.Commentf("Test %d: %v dot %v", i, test.a, test.b))
			c.Check(a.MulElem(b, tempDense).Equals(r), check.Equals, true, check.Commentf("Test %d: %v dot %v", i, test.a, test.b))
			t := a.CloneDense(nil)
			c.Check(a.MulElem(b, a).Equals(r), check.Equals, true, check.Commentf("Test %d: %v sub %v", i, test.a, test.b))
			a = t
			c.Check(a.MulElem(b, b).Equals(r), check.Equals, true, check.Commentf("Test %d: %v sub %v", i, test.a, test.b))
		}
	}
}

func (s *S) TestDot(c *check.C) {
	var (
		tempSparse *Sparse
		tempDense  *Dense
	)
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
		{
			[][]float64{{0, 1, 1}, {0, 1, 1}, {0, 1, 1}},
			[][]float64{{0, 1, 1}, {0, 1, 1}, {0, 1, 1}},
			[][]float64{{0, 2, 2}, {0, 2, 2}, {0, 2, 2}},
		},
	} {
		{
			a, err := NewSparse(test.a)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			b, err := NewSparse(test.b)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			r, err := NewSparse(test.r)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			c.Check(a.Dot(b, nil).Equals(r), check.Equals, true, check.Commentf("Test %d: %v dot %v", i, test.a, test.b))
			c.Check(a.Dot(b, tempSparse).Equals(r), check.Equals, true, check.Commentf("Test %d: %v dot %v", i, test.a, test.b))
		}

		{
			a, err := NewDense(test.a)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			b, err := NewDense(test.b)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			r, err := NewDense(test.r)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			c.Check(a.Dot(b, nil).Equals(r), check.Equals, true, check.Commentf("Test %d: %v dot %v", i, test.a, test.b))
			c.Check(a.Dot(b, tempDense).Equals(r), check.Equals, true, check.Commentf("Test %d: %v dot %v", i, test.a, test.b))
		}
	}
}

func (s *S) TestLU(c *check.C) {
	for _, fns := range []struct {
		I func(int) (Matrix, error)
		R func(size int, d float64, rnd func() float64) (Matrix, error)
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
			u := r.U(nil)
			l := r.L(nil)
			d := r.MulElem(I, nil)
			c.Check(r.Equals(u.Add(l.Sub(d, nil), nil)), check.Equals, true, check.Commentf("Test %d: type: %T", i, r))
			c.Check(d.Equals(l.MulElem(I, nil)), check.Equals, true, check.Commentf("Test %d: type: %T", i, r))
			c.Check(d.Equals(u.MulElem(I, nil)), check.Equals, true, check.Commentf("Test %d: type: %T", i, r))
			t := r.Clone(nil)
			r.U(r)
			c.Check(r.Equals(u), check.Equals, true, check.Commentf("Test %d: type: %T\ninput =\n%#.3f\n\nobtained =\n%#.3f\n\nexpected =\n%#.3f", i, r, t, r, u))
			r, t = t, t.Clone(nil)
			r.L(r)
			c.Check(r.Equals(l), check.Equals, true, check.Commentf("Test %d: type: %T\ninput =\n%#.3f\n\nobtained =\n%#.3f\n\nexpected =\n%#.3f", i, r, t, r, l))
		}
	}
}

func (s *S) TestTranspose(c *check.C) {
	var (
		tempSparse *Sparse
		tempDense  *Dense
	)
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
			c.Check(a.T(nil).Equals(t), check.Equals, true, check.Commentf("Test %d: %v transpose = %v", i, test.a, test.t))
			c.Check(a.T(nil).T(nil).Equals(a), check.Equals, true, check.Commentf("Test %d: %v transpose = I", i, test.a))
			c.Check(a.T(tempSparse).Equals(t), check.Equals, true, check.Commentf("Test %d: %v transpose = %v", i, test.a, test.t))
			c.Check(a.T(tempSparse).T(tempSparse).Equals(a), check.Equals, true, check.Commentf("Test %d: %v transpose = I", i, test.a))
		}

		{
			a, err := NewDense(test.a)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			t, err := NewDense(test.t)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			c.Check(a.T(nil).Equals(t), check.Equals, true, check.Commentf("Test %d: %v transpose = %v", i, test.a, test.t))
			c.Check(a.T(nil).T(nil).Equals(a), check.Equals, true, check.Commentf("Test %d: %v transpose = I", i, test.a, test.t))
			c.Check(a.T(tempDense).Equals(t), check.Equals, true, check.Commentf("Test %d: %v transpose = %v", i, test.a, test.t))
			c.Check(a.T(tempDense).T(tempDense).Equals(a), check.Equals, true, check.Commentf("Test %d: %v transpose = I", i, test.a, test.t))
		}
	}
}

func (s *S) TestStackAugment(c *check.C) {
	var (
		tempSparse *Sparse
		tempDense  *Dense
	)
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
		for _, C := range []*Sparse{nil, tempSparse} {
			a, err := NewSparse(test.a)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			b, err := NewSparse(test.b)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))

			var aug *Sparse
			if test.aug != nil {
				aug, err = NewSparse(test.aug)
				c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			}
			augr, err := Maybe(func() Matrix { return a.Augment(b, C) })
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
			stackr, err := Maybe(func() Matrix { return a.Stack(b, C) })
			if err == nil {
				c.Check(stackr.Equals(stack), check.Equals, true, check.Commentf("Test %d: %v stack %v", i, test.a, test.b))
			} else {
				c.Check(err, check.Equals, test.stackErr)
			}
		}

		for _, C := range []*Dense{nil, tempDense} {
			a, err := NewDense(test.a)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			b, err := NewDense(test.b)
			c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))

			var aug *Dense
			if test.aug != nil {
				aug, err = NewDense(test.aug)
				c.Check(err, check.Equals, nil, check.Commentf("Test %d", i))
			}
			augr, err := Maybe(func() Matrix { return a.Augment(b, C) })
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
			stackr, err := Maybe(func() Matrix { return a.Stack(b, C) })
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
// a.Filter(f, c)
// a.Apply(f, c)
// a.ApplyAll(f, c)

var (
	workDense  *Dense
	workSparse *Sparse
)

func BenchmarkDotDense100Half(b *testing.B)        { denseDotBench(b, 100, 0.5) }
func BenchmarkDotDense100Tenth(b *testing.B)       { denseDotBench(b, 100, 0.1) }
func BenchmarkDotDense1000Half(b *testing.B)       { denseDotBench(b, 1000, 0.5) }
func BenchmarkDotDense1000Tenth(b *testing.B)      { denseDotBench(b, 1000, 0.1) }
func BenchmarkDotDense1000Hundredth(b *testing.B)  { denseDotBench(b, 1000, 0.01) }
func BenchmarkDotDense1000Thousandth(b *testing.B) { denseDotBench(b, 1000, 0.001) }
func denseDotBench(b *testing.B, size int, rho float64) {
	b.StopTimer()
	a := MustDense(FuncDense(size, size, rho, rand.NormFloat64))
	d := MustDense(FuncDense(size, size, rho, rand.NormFloat64))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		workDense = a.DotDense(d, nil)
	}

}

func BenchmarkDotSparse100Half(b *testing.B)        { sparseDotBench(b, 100, 0.5) }
func BenchmarkDotSparse100Tenth(b *testing.B)       { sparseDotBench(b, 100, 0.1) }
func BenchmarkDotSparse1000Half(b *testing.B)       { sparseDotBench(b, 1000, 0.5) }
func BenchmarkDotSparse1000Tenth(b *testing.B)      { sparseDotBench(b, 1000, 0.1) }
func BenchmarkDotSparse1000Hundredth(b *testing.B)  { sparseDotBench(b, 1000, 0.01) }
func BenchmarkDotSparse1000Thousandth(b *testing.B) { sparseDotBench(b, 1000, 0.001) }
func sparseDotBench(b *testing.B, size int, rho float64) {
	b.StopTimer()
	a := MustSparse(FuncSparse(size, size, rho, rand.NormFloat64))
	d := MustSparse(FuncSparse(size, size, rho, rand.NormFloat64))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		workSparse = a.DotSparse(d, nil)
	}

}

func BenchmarkPreDotDense100Half(b *testing.B)        { denseDotBench(b, 100, 0.5) }
func BenchmarkPreDotDense100Tenth(b *testing.B)       { denseDotBench(b, 100, 0.1) }
func BenchmarkPreDotDense1000Half(b *testing.B)       { denseDotBench(b, 1000, 0.5) }
func BenchmarkPreDotDense1000Tenth(b *testing.B)      { denseDotBench(b, 1000, 0.1) }
func BenchmarkPreDotDense1000Hundredth(b *testing.B)  { denseDotBench(b, 1000, 0.01) }
func BenchmarkPreDotDense1000Thousandth(b *testing.B) { denseDotBench(b, 1000, 0.001) }
func densePreDotBench(b *testing.B, size int, rho float64) {
	b.StopTimer()
	a := MustDense(FuncDense(size, size, rho, rand.NormFloat64))
	d := MustDense(FuncDense(size, size, rho, rand.NormFloat64))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		workDense = a.DotDense(d, workDense)
	}

}

func BenchmarkPreDotSparse100Half(b *testing.B)        { sparseDotBench(b, 100, 0.5) }
func BenchmarkPreDotSparse100Tenth(b *testing.B)       { sparseDotBench(b, 100, 0.1) }
func BenchmarkPreDotSparse1000Half(b *testing.B)       { sparseDotBench(b, 1000, 0.5) }
func BenchmarkPreDotSparse1000Tenth(b *testing.B)      { sparseDotBench(b, 1000, 0.1) }
func BenchmarkPreDotSparse1000Hundredth(b *testing.B)  { sparseDotBench(b, 1000, 0.01) }
func BenchmarkPreDotSparse1000Thousandth(b *testing.B) { sparseDotBench(b, 1000, 0.001) }
func sparsePreDotBench(b *testing.B, size int, rho float64) {
	b.StopTimer()
	a := MustSparse(FuncSparse(size, size, rho, rand.NormFloat64))
	d := MustSparse(FuncSparse(size, size, rho, rand.NormFloat64))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		workSparse = a.DotSparse(d, workSparse)
	}

}
