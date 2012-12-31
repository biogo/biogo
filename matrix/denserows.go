// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package matrix

import (
	"github.com/ziutek/blas"
	"math"
)

type denseRow []float64

func (r denseRow) zero() {
	r[0] = 0
	for i := 1; i < len(r); {
		i += copy(r[i:], r[:i])
	}
}

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

func (r denseRow) scale(beta float64, b denseRow) denseRow {
	if b == nil || cap(b) < len(r) {
		b = make(denseRow, 0, len(r))
		for _, e := range r {
			b = append(b, e*beta)
		}
		return b
	}
	for i, e := range r {
		b[i] = e * beta
	}
	return b[:len(r)]
}

func (r denseRow) foldAdd(a, b denseRow) denseRow {
	for i, e := range r {
		b[i] = e + a[i]
	}

	return b
}

func (r denseRow) foldSub(a, b denseRow) denseRow {
	for i, e := range r {
		b[i] = e - a[i]
	}

	return b
}

func (r denseRow) foldMul(a, b denseRow) denseRow {
	for i, e := range r {
		b[i] = e * a[i]
	}

	return b
}

func (r denseRow) foldMulSum(a denseRow) float64 {
	// var s float64
	// for i, e := range r {
	// 	s += e * a[i]
	// }
	// return s
	return blas.Ddot(len(r), r, 1, a, 1)
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
