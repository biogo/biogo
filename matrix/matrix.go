// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Packages providing basic linear algebra operations
package matrix

const (
	Cols = true
	Rows = !Cols
	Inf  = int(^uint(0) >> 1)
	Fro  = -Inf - 1
)

var (
	Precision int          = 8
	Format    byte         = 'e'
	Pad       map[byte]int = map[byte]int{'e': 6, 'f': 2, 'g': 6, 'E': 6, 'F': 2, 'G': 6}
)

type FilterFunc func(r, c int, v float64) bool
type ApplyFunc func(r, c int, v float64) float64

type Matrix interface {
	Copy() Matrix
	Dims() (int, int)
	Set(r, c int, v float64) error
	At(r, c int) (float64, error)
	Norm() float64
	T() Matrix
	Det() float64
	Add(b Matrix) Matrix
	Sub(b Matrix) Matrix
	MulElem(b Matrix) Matrix
	Equals(b Matrix) bool
	EqualsApprox(b Matrix, error float64) bool
	Scalar(f float64) Matrix
	Sum() (s float64)
	Dot(b Matrix) Matrix
	Inner(b Matrix) Matrix
	Stack(b Matrix) Matrix
	Augment(b Matrix) Matrix
	Filter(f FilterFunc) Matrix
}
