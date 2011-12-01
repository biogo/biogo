package util
// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

import (
	"math"
)

const (
	MaxUint = ^uint(0)
	MinUint = 0
	MaxInt  = int(^uint(0) >> 1)
	MinInt  = -MaxInt - 1
	Ln4     = 1.3862943611198906188344642429163531361510002687205105082413
)

func Min(a ...int) (min int) {
	min = MaxInt
	for _, i := range a {
		if i < min {
			min = i
		}
	}
	return
}

func UMin(a ...uint) (min uint) {
	min = MaxUint
	for _, i := range a {
		if i < min {
			min = i
		}
	}
	return
}

func Max(a ...int) (max int) {
	max = MinInt
	for _, i := range a {
		if i > max {
			max = i
		}
	}
	return
}

func UMax(a ...uint) (max uint) {
	max = MinUint
	for _, i := range a {
		if i > max {
			max = i
		}
	}
	return
}

func Pow4(n int) uint {
	return uint(1) << (2 * uint(n))
}

func Log4(x float64) float64 {
	return math.Log(x) / Ln4
}
