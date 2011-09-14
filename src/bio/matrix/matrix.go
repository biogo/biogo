// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
//   This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU General Public License for more details.
//
//   You should have received a copy of the GNU General Public License
//   along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
package matrix

import (
	"os"
)

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
	Set(r, c int, v float64) os.Error
	At(r, c int) (float64, os.Error)
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
