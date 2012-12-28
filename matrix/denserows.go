// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package matrix

import (
	"math"
)

type denseRow []float64

func (r denseRow) sum() float64 {
	var s float64
	for _, e := range r {
		s += e
	}

	return s
}

func (r denseRow) min() float64 {
	n := math.MaxFloat64
	for _, e := range r {
		n = math.Min(n, e)
	}

	return n
}

func (r denseRow) max() float64 {
	n := -math.MaxFloat64
	for _, e := range r {
		n = math.Max(n, e)
	}

	return n
}

func (r denseRow) scale(beta float64) denseRow {
	b := make(denseRow, 0, len(r))
	for _, e := range r {
		b = append(b, e*e)
	}

	return b
}

func (r denseRow) foldAdd(a denseRow) denseRow {
	b := make(denseRow, 0, len(r))
	for i, e := range r {
		b = append(b, e+a[i])
	}

	return b
}

func (r denseRow) foldSub(a denseRow) denseRow {
	b := make(denseRow, 0, len(r))
	for i, e := range r {
		b = append(b, e-a[i])
	}

	return b
}

func (r denseRow) foldMul(a denseRow) denseRow {
	b := make(denseRow, 0, len(r))
	for i, e := range r {
		b = append(b, e*a[i])
	}

	return b
}

func (r denseRow) foldMulSum(a denseRow) float64 {
	var s float64
	for i, e := range r {
		s += e * a[i]
	}

	return s
}

func (r denseRow) foldEqual(a denseRow) bool {
	for i, e := range r {
		if e != a[i] {
			return false
		}
	}

	return true
}

func (r denseRow) foldApprox(a denseRow, error float64) bool {
	for i, e := range r {
		if math.Abs(e-a[i]) > error {
			return false
		}
	}

	return true
}
