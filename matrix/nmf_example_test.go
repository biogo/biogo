// Copyright ©2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package matrix_test

import (
	"code.google.com/p/biogo/matrix"
	"fmt"
	"math/rand"
	"time"
)

func ExampleFactors() {
	V := matrix.Must(matrix.NewDense([][]float64{{20, 0, 30, 0}, {0, 16, 1, 9}, {0, 10, 6, 11}}))
	fmt.Printf("V =\n%v\n\n", V)

	categories := 5

	rows, cols := V.Dims()
	rand.Seed(1)
	Wo := matrix.Must(matrix.FuncDense(rows, categories, 1, rand.NormFloat64))
	Ho := matrix.Must(matrix.FuncDense(categories, cols, 1, rand.NormFloat64))

	var (
		tolerance  = 1e-5
		iterations = 1000
		limit      = time.Second
	)

	W, H, ok := matrix.Factors(V, Wo, Ho, tolerance, iterations, limit)

	P := W.Dot(H)

	fmt.Printf("Successfully factorised: %v\n\n", ok)
	fmt.Printf("W =\n%.3f\n\nH =\n%.3f\n\n", W, H)
	fmt.Printf("P =\n%.3f\n\n", P)
	fmt.Printf("delta = %.3f\n", V.Sub(P).Norm(matrix.Fro))

	// Output:
	// V =
	// ⎡20   0  30   0⎤
	// ⎢ 0  16   1   9⎥
	// ⎣ 0  10   6  11⎦
	//
	// Successfully factorised: true
	//
	// W =
	// ⎡ 0.000  23.860   0.000  18.962   0.000⎤
	// ⎢19.343   0.000   0.000   0.000  11.348⎥
	// ⎣32.072   0.000   0.000   2.743   7.030⎦
	//
	// H =
	// ⎡0.000  0.004  0.052  0.270⎤
	// ⎢0.838  0.000  0.000  0.000⎥
	// ⎢0.927  0.972  0.510  0.251⎥
	// ⎢0.000  0.000  1.582  0.000⎥
	// ⎣0.000  1.402  0.000  0.333⎦
	//
	// P =
	// ⎡20.000   0.000  30.000   0.000⎤
	// ⎢ 0.000  16.000   1.001   9.000⎥
	// ⎣ 0.000  10.000   5.999  11.000⎦
	//
	// delta = 0.001
}
