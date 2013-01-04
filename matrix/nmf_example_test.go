// Copyright ©2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package matrix_test

import (
	"code.google.com/p/biogo/matrix"
	"fmt"
	"math"
	"math/rand"
	"time"
)

func ExampleFactors() {
	V := matrix.Must(matrix.NewDense([][]float64{{20, 0, 30, 0}, {0, 16, 1, 9}, {0, 10, 6, 11}}))
	fmt.Printf("V =\n%v\n\n", V)

	categories := 5

	rows, cols := V.Dims()
	rand.Seed(1)
	posNorm := func() float64 { return math.Abs(rand.NormFloat64()) }
	Wo := matrix.Must(matrix.FuncDense(rows, categories, 1, posNorm))
	Ho := matrix.Must(matrix.FuncDense(categories, cols, 1, posNorm))

	var (
		tolerance  = 1e-5
		iterations = 1000
		limit      = time.Second
	)

	W, H, ok := matrix.Factors(V, Wo, Ho, tolerance, iterations, limit)

	P := W.Dot(H, nil)

	fmt.Printf("Successfully factorised: %v\n\n", ok)
	fmt.Printf("W =\n%.3f\n\nH =\n%.3f\n\n", W, H)
	fmt.Printf("P =\n%.3f\n\n", P)
	fmt.Printf("delta = %.3f\n", V.Sub(P, nil).Norm(matrix.Fro))

	// Output:
	// V =
	// ⎡20   0  30   0⎤
	// ⎢ 0  16   1   9⎥
	// ⎣ 0  10   6  11⎦
	//
	// Successfully factorised: true
	//
	// W =
	// ⎡ 0.000   0.000  15.891   0.000   0.000⎤
	// ⎢16.693   1.603   0.000   0.000   0.000⎥
	// ⎣ 0.000   4.017   0.000   4.155   0.000⎦
	//
	// H =
	// ⎡0.000  0.868  0.030  0.393⎤
	// ⎢0.000  0.938  0.309  1.523⎥
	// ⎢1.259  0.000  1.888  0.000⎥
	// ⎢0.000  1.500  1.145  1.175⎥
	// ⎣0.776  1.420  0.330  0.247⎦
	//
	// P =
	// ⎡20.000   0.000  30.000   0.000⎤
	// ⎢ 0.000  16.000   1.000   9.000⎥
	// ⎣ 0.000  10.000   6.000  11.000⎦
	//
	// delta = 0.000
}
