// Copyright ©2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package matrix_test

import (
	"code.google.com/p/biogo/matrix"
	"fmt"
)

func ExampleMaybe() {
	fmt.Println(matrix.Maybe(func() matrix.Matrix {
		return matrix.Must(matrix.IdentityDense(10)).Dot(matrix.Must(matrix.IdentityDense(2)), nil)
	}))

	// Output:
	// <nil> matrix: dimension mismatch
}
