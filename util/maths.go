// Copyright ©2011-2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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

package util

import (
	"math"
)

const (
	MaxUint = ^uint(0)                                                     // The maximum uint value.
	MinUint = uint(0)                                                      // The minimum uint value.
	MaxInt  = int(^uint(0) >> 1)                                           // The maximum int value.
	MinInt  = -MaxInt - 1                                                  // The minimum int value.
	Ln4     = 1.3862943611198906188344642429163531361510002687205105082413 // The natural log of 4.
)

// Returns the minimum int of a...
func Min(a ...int) (min int) {
	min = MaxInt
	for _, i := range a {
		if i < min {
			min = i
		}
	}
	return
}

// Returns the minimum uint of a...
func UMin(a ...uint) (min uint) {
	min = MaxUint
	for _, i := range a {
		if i < min {
			min = i
		}
	}
	return
}

// Returns the maximum int of a...
func Max(a ...int) (max int) {
	max = MinInt
	for _, i := range a {
		if i > max {
			max = i
		}
	}
	return
}

// Returns the maximum uint of a...
func UMax(a ...uint) (max uint) {
	max = MinUint
	for _, i := range a {
		if i > max {
			max = i
		}
	}
	return
}

// Return the exp'th power of base.
func Pow(base int, exp byte) (r int) {
	r = 1
	for exp > 0 {
		if exp&1 != 0 {
			r *= base
		}
		exp >>= 1
		base *= base
	}

	return
}

// Returns the nth power of 4.
func Pow4(n int) uint {
	return uint(1) << (2 * uint(n))
}

// Returns the log base 4 of x.
func Log4(x float64) float64 {
	return math.Log(x) / Ln4
}
