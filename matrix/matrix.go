// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package matrix provides basic linear algebra operations.
package matrix

import (
	"math"
)

const (
	Cols = true
	Rows = !Cols
)

const (
	Inf = int(^uint(0) >> 1)
	Fro = -Inf - 1
)

type FilterFunc func(r, c int, v float64) bool
type ApplyFunc func(r, c int, v float64) float64

type Matrix interface {
	Clone() Matrix
	Dims() (int, int)
	At(r, c int) float64
	Norm(int) float64
	T() Matrix
	Det() float64
	Add(b Matrix) Matrix
	Sub(b Matrix) Matrix
	MulElem(b Matrix) Matrix
	Equals(b Matrix) bool
	EqualsApprox(b Matrix, epsilon float64) bool
	Scalar(f float64) Matrix
	Sum() (s float64)
	Dot(b Matrix) Matrix
	Inner(b Matrix) float64
	Stack(b Matrix) Matrix
	Augment(b Matrix) Matrix
	Filter(f FilterFunc) Matrix
	Trace() float64
	U() Matrix
	L() Matrix
}

type Mutable interface {
	Matrix
	Set(r, c int, v float64)
}

type Panicker func() Matrix

func Maybe(fn Panicker) (m Matrix, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(Error); ok {
				err = e
				return
			}
			panic(r)
		}
	}()
	return fn(), nil
}

func Must(m Matrix, err error) Matrix {
	if err != nil {
		panic(err)
	}
	return m
}

type Error string

func (err Error) Error() string { return string(err) }

const (
	ErrIndexOutOfBounds = Error("matrix: index out of bounds")
	ErrZeroLength       = Error("matrix: zero length in matrix definition")
	ErrRowLength        = Error("matrix: row length mismatch")
	ErrColLength        = Error("matrix: col length mismatch")
	ErrSquare           = Error("matrix: expect square matrix")
	ErrNormOrder        = Error("matrix: invalid norm order for matrix")
	ErrShape            = Error("matrix: dimension mismatch")
)

// Determine a variety of norms on a vector.
func Norm(v []float64, ord int) float64 {
	var n float64
	switch ord {
	case 0:
		for _, e := range v {
			if e != 0 {
				n += e
			}
		}
	case Inf:
		for _, e := range v {
			n = math.Max(math.Abs(e), n)
		}
	case -Inf:
		n = math.MaxFloat64
		for _, e := range v {
			n = math.Min(math.Abs(e), n)
		}
	case Fro, 2:
		for _, e := range v {
			n += e * e
		}
		return math.Sqrt(n)
	default:
		ord := float64(ord)
		for _, e := range v {
			n += math.Pow(math.Abs(e), ord)
		}
		return math.Pow(n, 1/ord)
	}
	return n
}
