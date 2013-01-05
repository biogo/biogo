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

type FloatFunc func() float64
type FilterFunc func(r, c int, v float64) bool
type ApplyFunc func(r, c int, v float64) float64

type Matrix interface {
	Clone(c Matrix) Matrix
	Dims() (int, int)
	At(r, c int) float64
	Norm(int) float64
	T(c Matrix) Matrix
	Det() float64
	Add(b, c Matrix) Matrix
	Sub(b, c Matrix) Matrix
	MulElem(b, c Matrix) Matrix
	Equals(b Matrix) bool
	EqualsApprox(b Matrix, epsilon float64) bool
	Scalar(f float64, c Matrix) Matrix
	Sum() (s float64)
	Dot(b, c Matrix) Matrix
	Inner(b Matrix) float64
	Stack(b, c Matrix) Matrix
	Augment(b, c Matrix) Matrix
	Apply(f ApplyFunc, c Matrix) Matrix
	ApplyAll(f ApplyFunc, c Matrix) Matrix
	Filter(f FilterFunc, c Matrix) Matrix
	Trace() float64
	U(Matrix) Matrix
	L(Matrix) Matrix
	Sparse(*Sparse) *Sparse
	Dense(*Dense) *Dense
}

type Mutable interface {
	Matrix
	New(r, c int) (Matrix, error)
	Set(r, c int, v float64)
}

// A Panicker is a function that returns a matrix and may panic.
type Panicker func() Matrix

// Maybe will recover a panic with a type matrix.Error from fn, and return this error.
// Any other error is re-panicked.
func Maybe(fn Panicker) (m Matrix, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			if err, ok = r.(Error); ok {
				return
			}
			panic(r)
		}
	}()
	return fn(), nil
}

// A FloatPanicker is a function that returns a float64 and may panic.
type FloatPanicker func() float64

// MaybeFloat will recover a panic with a type matrix.Error from fn, and return this error.
// Any other error is re-panicked.
func MaybeFloat(fn FloatPanicker) (f float64, err error) {
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

// Must can be used to wrap a function returning a matrix and an error.
// If the returned error is not nil, Must will panic.
func Must(m Matrix, err error) Matrix {
	if err != nil {
		panic(err)
	}
	return m
}

// Type Error represents matrix package errors. These errors can be recovered by Maybe wrappers.
type Error string

func (err Error) Error() string { return string(err) }

const (
	ErrIndexOutOfRange = Error("matrix: index out of range")
	ErrZeroLength      = Error("matrix: zero length in matrix definition")
	ErrRowLength       = Error("matrix: row length mismatch")
	ErrColLength       = Error("matrix: col length mismatch")
	ErrSquare          = Error("matrix: expect square matrix")
	ErrNormOrder       = Error("matrix: invalid norm order for matrix")
	ErrShape           = Error("matrix: dimension mismatch")
	ErrPivot           = Error("matrix: malformed pivot list")
)

// ElementsVector returns the matrix's elements concatenated, row-wise, into a float slice.
func ElementsVector(mats ...Matrix) []float64 {
	var length int
	for _, m := range mats {
		switch m := m.(type) {
		case *Dense:
			length += len(m.matrix)
		case *Sparse:
			for _, row := range m.matrix {
				length += len(row)
			}
		}
	}

	v := make([]float64, 0, length)
	for _, m := range mats {
		switch m := m.(type) {
		case *Dense:
			v = append(v, m.matrix...)
		case *Sparse:
			for _, row := range m.matrix {
				for _, e := range row {
					if e.value != 0 {
						v = append(v, e.value)
					}
				}
			}
		case Matrix:
			rows, cols := m.Dims()
			for r := 0; r < rows; r++ {
				for c := 0; c < cols; c++ {
					v = append(v, m.At(r, c))
				}
			}
		}
	}

	return v
}

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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
