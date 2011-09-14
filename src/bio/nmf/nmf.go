// Implementation of NMF by alternative non-negative least squares using projected gradients
//
// Chih-Jen Lin (2007) Projected Gradient Methods for Nonnegative Matrix Factorization. Neural Computation 19:2756. 
//
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
package nmf

import (
	"time"
	"math"
	"bio/matrix"
	"bio/matrix/sparse"
)

// Default inner loop size for subproblem
var InnerLoop = 20

func Factors(X, Wo, Ho *sparse.Sparse, tolerance float64, iterations int, limit int64) (W, H *sparse.Sparse, ok bool) {
	W = Wo
	H = Ho
	to := time.Nanoseconds()

	hT := H.T()
	wT := W.T()

	gW := W.Dot(H.Dot(hT)).Sub(X.Dot(hT))
	gH := wT.Dot(W).Dot(H).Sub(wT.Dot(X))
	
	gradient := gW.Stack(gH.T()).Norm(matrix.Fro)
	toleranceW := math.Fmax(0.001, tolerance) * gradient
	toleranceH := toleranceW

	wFilter := func(r, c int, v float64) bool {
		w := W.At(r, c)
		return v < 0 || w > 0 // wFilter called with gW as receiver, so v, _ = gW.At(r, c)
	}
	hFilter := func(r, c int, v float64) bool {
		h := H.At(r, c)
		return v < 0 || h > 0 // hFilter called with gH as receiver, so v, _ = gH.At(r, c)
	}

	var (
		subOk  bool
		iW, iH int
	)
	ok = true

	for i := 1; i < iterations; i++ {
		projection := sparse.Elements(gW.Filter(wFilter), (gH.Filter(hFilter))).Norm(matrix.Fro)
		if projection < tolerance*gradient || time.Nanoseconds()-to > limit {
			break
		}

		if W, gW, iW, subOk = subproblem(X.T(), H.T(), W.T(), toleranceW, 1000); iW == 1 {
			toleranceW *= 0.1
		}
		ok = ok && subOk

		W = W.T()
		gW = gW.T()

		if H, gH, iH, subOk = subproblem(X, W, H, toleranceH, 1000); iH == 1 {
			toleranceH *= 0.1
		}
		ok = ok && subOk
	}

	return
}

func subproblem(X, W, Ho *sparse.Sparse, tolerance float64, iterations int) (H, G *sparse.Sparse, i int, ok bool) {
	H = Ho.Copy()
	WtV := W.T().Dot(X)
	WtW := W.T().Dot(W)

	var alpha, beta float64 = 1, 0.1

	for i := 0; i < iterations; i++ {
		G = WtW.Dot(H).Sub(WtV)

		filter := func(r, c int, v float64) bool {
			h := H.At(r, c)
			return v < 0 || h > 0 // filter called with G as receiver, so v, _ = G.At(r, c)
		}
		if projection := G.Filter(filter).Norm(matrix.Fro); projection < tolerance {
			break
		}
	}

	var (
		decrease bool
		Hp       *sparse.Sparse
	)

	for j := 0; j < InnerLoop; j++ {
		Hn := H.Sub(G.Scalar(alpha))
		filter := func(r, c int, v float64) bool {
			return v > 0 // filter called with Hn as receiver, so v, _ = Hn.At(r, c)
		}
		Hn = Hn.Filter(filter)

		d := Hn.Sub(H)
		gd := G.MulElem(d).Sum()
		dQd := WtW.Dot(d).MulElem(d).Sum()
		sufficient := 0.99*gd+0.5*dQd < 0
		if j == 0 {
			decrease = !sufficient
			Hp = H.Copy()
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
				ok = true
				break
			} else {
				alpha /= beta
				Hp = Hn.Copy()
			}
		}
	}

	return
}
