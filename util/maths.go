// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
