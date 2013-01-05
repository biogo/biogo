// Copyright ©2012 The bíogo.llrb Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package matrix

import (
	"fmt"
	check "launchpad.net/gocheck"
	"math"
)

func (s *S) TestFormat(c *check.C) {
	type rp struct {
		format string
		output string
	}
	sqrt := func(_, _ int, v float64) float64 { return math.Sqrt(v) }
	for i, test := range []struct {
		m   Matrix
		rep []rp
	}{
		// Dense matrix representation
		{
			Must(NewDense([][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}})),
			[]rp{
				{"%v", "⎡0  0  0⎤\n⎢0  0  0⎥\n⎣0  0  0⎦"},
				{"%#f", "⎡.  .  .⎤\n⎢.  .  .⎥\n⎣.  .  .⎦"},
				{"%#v", "&matrix.Dense{Margin:0, rows:3, cols:3, matrix:matrix.denseRow{0, 0, 0, 0, 0, 0, 0, 0, 0}}"},
				{"%s", "%!s(*matrix.Dense=Dims(3, 3))"},
			},
		},
		{
			Must(NewDense([][]float64{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}})),
			[]rp{
				{"%v", "⎡1  1  1⎤\n⎢1  1  1⎥\n⎣1  1  1⎦"},
				{"%#f", "⎡1  1  1⎤\n⎢1  1  1⎥\n⎣1  1  1⎦"},
				{"%#v", "&matrix.Dense{Margin:0, rows:3, cols:3, matrix:matrix.denseRow{1, 1, 1, 1, 1, 1, 1, 1, 1}}"},
			},
		},
		{
			Must(NewDense([][]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}})),
			[]rp{
				{"%v", "⎡1  0  0⎤\n⎢0  1  0⎥\n⎣0  0  1⎦"},
				{"%#f", "⎡1  .  .⎤\n⎢.  1  .⎥\n⎣.  .  1⎦"},
				{"%#v", "&matrix.Dense{Margin:0, rows:3, cols:3, matrix:matrix.denseRow{1, 0, 0, 0, 1, 0, 0, 0, 1}}"},
			},
		},
		{
			Must(NewDense([][]float64{{1, 2, 3}, {4, 5, 6}})),
			[]rp{
				{"%v", "⎡1  2  3⎤\n⎣4  5  6⎦"},
				{"%#f", "⎡1  2  3⎤\n⎣4  5  6⎦"},
				{"%#v", "&matrix.Dense{Margin:0, rows:2, cols:3, matrix:matrix.denseRow{1, 2, 3, 4, 5, 6}}"},
			},
		},
		{
			Must(NewDense([][]float64{{1, 2}, {3, 4}, {5, 6}})),
			[]rp{
				{"%v", "⎡1  2⎤\n⎢3  4⎥\n⎣5  6⎦"},
				{"%#f", "⎡1  2⎤\n⎢3  4⎥\n⎣5  6⎦"},
				{"%#v", "&matrix.Dense{Margin:0, rows:3, cols:2, matrix:matrix.denseRow{1, 2, 3, 4, 5, 6}}"},
			},
		},
		{
			MustDense(NewDense([][]float64{{0, 1, 2}, {3, 4, 5}})).ApplyAll(sqrt, nil),
			[]rp{
				{"%v", "⎡                 0                   1  1.4142135623730951⎤\n⎣1.7320508075688772                   2    2.23606797749979⎦"},
				{"%#f", "⎡                 .                   1  1.4142135623730951⎤\n⎣1.7320508075688772                   2    2.23606797749979⎦"},
				{"%#v", "&matrix.Dense{Margin:0, rows:2, cols:3, matrix:matrix.denseRow{0, 1, 1.4142135623730951, 1.7320508075688772, 2, 2.23606797749979}}"},
			},
		},
		{
			MustDense(NewDense([][]float64{{0, 1}, {2, 3}, {4, 5}})).ApplyAll(sqrt, nil),
			[]rp{
				{"%v", "⎡                 0                   1⎤\n⎢1.4142135623730951  1.7320508075688772⎥\n⎣                 2    2.23606797749979⎦"},
				{"%#f", "⎡                 .                   1⎤\n⎢1.4142135623730951  1.7320508075688772⎥\n⎣                 2    2.23606797749979⎦"},
				{"%#v", "&matrix.Dense{Margin:0, rows:3, cols:2, matrix:matrix.denseRow{0, 1, 1.4142135623730951, 1.7320508075688772, 2, 2.23606797749979}}"},
			},
		},
		{
			func() Matrix {
				m := MustDense(NewDense([][]float64{{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}))
				m.Margin = 3
				return m
			}(),
			[]rp{
				{"%v", "Dims(1, 10)\n[ 1   2   3  ...  ...   8   9  10]"},
			},
		},
		{
			func() Matrix {
				m := MustDense(NewDense([][]float64{{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10}}))
				m.Margin = 3
				return m
			}(),
			[]rp{
				{"%v", "Dims(10, 1)\n⎡ 1⎤\n⎢ 2⎥\n⎢ 3⎥\n .\n .\n .\n⎢ 8⎥\n⎢ 9⎥\n⎣10⎦"},
			},
		},
		{
			func() Matrix { m := MustDense(IdentityDense(10)); m.Margin = 3; return m }(),
			[]rp{
				{"%v", "Dims(10, 10)\n⎡1  0  0  ...  ...  0  0  0⎤\n⎢0  1  0            0  0  0⎥\n⎢0  0  1            0  0  0⎥\n .\n .\n .\n⎢0  0  0            1  0  0⎥\n⎢0  0  0            0  1  0⎥\n⎣0  0  0  ...  ...  0  0  1⎦"},
			},
		},

		// Sparse matrix representation
		{
			Must(NewSparse([][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}})),
			[]rp{
				{"%v", "⎡0  0  0⎤\n⎢0  0  0⎥\n⎣0  0  0⎦"},
				{"%#f", "⎡.  .  .⎤\n⎢.  .  .⎥\n⎣.  .  .⎦"},
				{"%#v", "&matrix.Sparse{Margin:0, rows:3, cols:3, matrix:[]matrix.sparseRow{matrix.sparseRow(nil), matrix.sparseRow(nil), matrix.sparseRow(nil)}}"},
				{"%s", "%!s(*matrix.Sparse=Dims(3, 3))"},
			},
		},
		{
			Must(NewSparse([][]float64{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}})),
			[]rp{
				{"%v", "⎡1  1  1⎤\n⎢1  1  1⎥\n⎣1  1  1⎦"},
				{"%#f", "⎡1  1  1⎤\n⎢1  1  1⎥\n⎣1  1  1⎦"},
				{"%#v", "&matrix.Sparse{Margin:0, rows:3, cols:3, matrix:[]matrix.sparseRow{matrix.sparseRow{matrix.sparseElem{index:0, value:1}, matrix.sparseElem{index:1, value:1}, matrix.sparseElem{index:2, value:1}}, matrix.sparseRow{matrix.sparseElem{index:0, value:1}, matrix.sparseElem{index:1, value:1}, matrix.sparseElem{index:2, value:1}}, matrix.sparseRow{matrix.sparseElem{index:0, value:1}, matrix.sparseElem{index:1, value:1}, matrix.sparseElem{index:2, value:1}}}}"},
			},
		},
		{
			Must(NewSparse([][]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}})),
			[]rp{
				{"%v", "⎡1  0  0⎤\n⎢0  1  0⎥\n⎣0  0  1⎦"},
				{"%#f", "⎡1  .  .⎤\n⎢.  1  .⎥\n⎣.  .  1⎦"},
				{"%#v", "&matrix.Sparse{Margin:0, rows:3, cols:3, matrix:[]matrix.sparseRow{matrix.sparseRow{matrix.sparseElem{index:0, value:1}}, matrix.sparseRow{matrix.sparseElem{index:1, value:1}}, matrix.sparseRow{matrix.sparseElem{index:2, value:1}}}}"},
			},
		},
		{
			Must(NewSparse([][]float64{{1, 2, 3}, {4, 5, 6}})),
			[]rp{
				{"%v", "⎡1  2  3⎤\n⎣4  5  6⎦"},
				{"%#f", "⎡1  2  3⎤\n⎣4  5  6⎦"},
				{"%#v", "&matrix.Sparse{Margin:0, rows:2, cols:3, matrix:[]matrix.sparseRow{matrix.sparseRow{matrix.sparseElem{index:0, value:1}, matrix.sparseElem{index:1, value:2}, matrix.sparseElem{index:2, value:3}}, matrix.sparseRow{matrix.sparseElem{index:0, value:4}, matrix.sparseElem{index:1, value:5}, matrix.sparseElem{index:2, value:6}}}}"},
			},
		},
		{
			Must(NewSparse([][]float64{{1, 2}, {3, 4}, {5, 6}})),
			[]rp{
				{"%v", "⎡1  2⎤\n⎢3  4⎥\n⎣5  6⎦"},
				{"%#f", "⎡1  2⎤\n⎢3  4⎥\n⎣5  6⎦"},
				{"%#v", "&matrix.Sparse{Margin:0, rows:3, cols:2, matrix:[]matrix.sparseRow{matrix.sparseRow{matrix.sparseElem{index:0, value:1}, matrix.sparseElem{index:1, value:2}}, matrix.sparseRow{matrix.sparseElem{index:0, value:3}, matrix.sparseElem{index:1, value:4}}, matrix.sparseRow{matrix.sparseElem{index:0, value:5}, matrix.sparseElem{index:1, value:6}}}}"},
			},
		},
		{
			MustSparse(NewSparse([][]float64{{0, 1, 2}, {3, 4, 5}})).ApplyAll(sqrt, nil),
			[]rp{
				{"%v", "⎡                 0                   1  1.4142135623730951⎤\n⎣1.7320508075688772                   2    2.23606797749979⎦"},
				{"%#f", "⎡                 .                   1  1.4142135623730951⎤\n⎣1.7320508075688772                   2    2.23606797749979⎦"},
				{"%#v", "&matrix.Sparse{Margin:0, rows:2, cols:3, matrix:[]matrix.sparseRow{matrix.sparseRow{matrix.sparseElem{index:1, value:1}, matrix.sparseElem{index:2, value:1.4142135623730951}}, matrix.sparseRow{matrix.sparseElem{index:0, value:1.7320508075688772}, matrix.sparseElem{index:1, value:2}, matrix.sparseElem{index:2, value:2.23606797749979}}}}"},
			},
		},
		{
			MustSparse(NewSparse([][]float64{{0, 1}, {2, 3}, {4, 5}})).ApplyAll(sqrt, nil),
			[]rp{
				{"%v", "⎡                 0                   1⎤\n⎢1.4142135623730951  1.7320508075688772⎥\n⎣                 2    2.23606797749979⎦"},
				{"%#f", "⎡                 .                   1⎤\n⎢1.4142135623730951  1.7320508075688772⎥\n⎣                 2    2.23606797749979⎦"},
				{"%#v", "&matrix.Sparse{Margin:0, rows:3, cols:2, matrix:[]matrix.sparseRow{matrix.sparseRow{matrix.sparseElem{index:1, value:1}}, matrix.sparseRow{matrix.sparseElem{index:0, value:1.4142135623730951}, matrix.sparseElem{index:1, value:1.7320508075688772}}, matrix.sparseRow{matrix.sparseElem{index:0, value:2}, matrix.sparseElem{index:1, value:2.23606797749979}}}}"},
			},
		},
		{
			func() Matrix {
				m := MustSparse(NewSparse([][]float64{{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}))
				m.Margin = 3
				return m
			}(),
			[]rp{
				{"%v", "Dims(1, 10)\n[ 1   2   3  ...  ...   8   9  10]"},
			},
		},
		{
			func() Matrix {
				m := MustSparse(NewSparse([][]float64{{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10}}))
				m.Margin = 3
				return m
			}(),
			[]rp{
				{"%v", "Dims(10, 1)\n⎡ 1⎤\n⎢ 2⎥\n⎢ 3⎥\n .\n .\n .\n⎢ 8⎥\n⎢ 9⎥\n⎣10⎦"},
			},
		},
		{
			func() Matrix { m := MustSparse(IdentitySparse(10)); m.Margin = 3; return m }(),
			[]rp{
				{"%v", "Dims(10, 10)\n⎡1  0  0  ...  ...  0  0  0⎤\n⎢0  1  0            0  0  0⎥\n⎢0  0  1            0  0  0⎥\n .\n .\n .\n⎢0  0  0            1  0  0⎥\n⎢0  0  0            0  1  0⎥\n⎣0  0  0  ...  ...  0  0  1⎦"},
			},
		},
	} {
		for _, rp := range test.rep {
			c.Check(fmt.Sprintf(rp.format, test.m), check.Equals, rp.output, check.Commentf("Test %d", i))
		}
	}
}
