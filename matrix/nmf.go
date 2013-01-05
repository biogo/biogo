// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package matrix

import (
	"math"
	"time"
)

// Default inner loop size for subproblem
var NmfInnerLoop = 20

// Factors is an implementation of non-negative matrix factorisation by alternative non-negative least
// squares using projected gradients according to:
//
// Chih-Jen Lin (2007) 'Projected Gradient Methods for Non-negative Matrix Factorization.'
// Neural Computation 19:2756.
//
// Factors returns matrices W and H that are non-negative factors of V within the specified tolerance,
// given initial non-negative solutions Wo and Ho.
func Factors(V, Wo, Ho Matrix, tolerance float64, iterations int, limit time.Duration) (W, H Matrix, ok bool) {
	W = Wo
	H = Ho
	to := time.Now()

	hT := H.T(nil)
	wT := W.T(nil)

	gW := W.Dot(H.Dot(hT, nil), nil).Sub(V.Dot(hT, nil), nil)
	gH := wT.Dot(W, nil).Dot(H, nil).Sub(wT.Dot(V, nil), nil)

	gradient := gW.Stack(gH.T(nil), nil).Norm(Fro)
	toleranceW := math.Max(0.001, tolerance) * gradient
	toleranceH := toleranceW

	var (
		subOk  bool
		iW, iH int
	)

	wFilter := func(r, c int, v float64) bool {
		return v < 0 || W.At(r, c) > 0 // wFilter called with gW as receiver, so v = gW.At(r, c)
	}
	hFilter := func(r, c int, v float64) bool {
		return v < 0 || H.At(r, c) > 0 // hFilter called with gH as receiver, so v = gH.At(r, c)
	}

	for i := 0; i < iterations; i++ {
		ok = true
		projection := Norm(ElementsVector(gW.Filter(wFilter, gW), gH.Filter(hFilter, gH)), Fro)
		if projection < tolerance*gradient || time.Now().Sub(to) > limit {
			break
		}
		W, gW, iW, subOk = subproblem(V.T(nil), H.T(nil), W.T(nil), toleranceW, iterations)
		if iW == 0 {
			toleranceW *= 0.1
		}
		ok = ok && subOk
		W = W.T(nil)
		gW = gW.T(nil)

		H, gH, iH, subOk = subproblem(V, W, H, toleranceH, iterations)
		if iH == 0 {
			toleranceH *= 0.1
		}
		ok = ok && subOk
	}

	return
}

func subproblem(V, W, Ho Matrix, tolerance float64, iterations int) (H, G Matrix, i int, ok bool) {
	H = Ho.Clone(nil)
	WtV := W.T(nil).Dot(V, nil)
	WtW := W.T(nil).Dot(W, nil)

	var alpha, beta float64 = 1, 0.1

	hFilter := func(r, c int, v float64) bool {
		return v > 0 // filter called with H* as receiver, so v = H*.At(r, c)
	}
	ghFilter := func(r, c int, v float64) bool {
		return v < 0 || H.At(r, c) > 0 // filter called with G as receiver, so v = G.At(r, c)
	}

	for i = 0; i < iterations; i++ {
		G = WtW.Dot(H, nil).Sub(WtV, G)
		if projection := G.Filter(ghFilter, G).Norm(Fro); projection < tolerance {
			break
		}

		var (
			decrease bool
			Hp       Matrix
		)

		for j := 0; j < NmfInnerLoop; j++ {
			Hn := G.Scalar(alpha, nil)
			Hn = H.Sub(Hn, Hn)
			Hn = Hn.Filter(hFilter, Hn)

			d := Hn.Sub(H, nil)
			gd := G.MulElem(d, nil).Sum()
			dQd := WtW.Dot(d, nil).MulElem(d, d).Sum()
			sufficient := 0.99*gd+0.5*dQd < 0
			if j == 0 {
				decrease = !sufficient
				Hp = H
			}
			if decrease {
				if sufficient {
					H = Hn
					ok = true
					break
				} else {
					alpha *= beta
				}
			} else {
				if !sufficient || Hp.Equals(Hn) {
					H = Hp
					break
				} else {
					alpha /= beta
					Hp = Hn
				}
			}
		}
	}

	return
}
