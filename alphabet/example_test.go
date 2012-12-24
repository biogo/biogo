// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package alphabet

import (
	"fmt"
)

func ExampleGeneric_AllValid() {
	fmt.Println(DNA.AllValid([]byte("acgatcgatatagctatnagcatgc")))
	// Output:
	// false 17
}

func ExamplePairing_ComplementOf() {
	var (
		c  byte
		ok bool
	)

	c, ok = DNA.ComplementOf('a')
	fmt.Printf("%c %v\n", c, ok)
	c, ok = DNA.ComplementOf('n')
	fmt.Printf("%c %v\n", c, ok)
	c, ok = RNA.ComplementOf('a')
	fmt.Printf("%c %v\n", c, ok)
	_, ok = RNA.ComplementOf('t')
	fmt.Printf("%v\n", ok)
	// Output:
	// t true
	// n true
	// u true
	// false
}
