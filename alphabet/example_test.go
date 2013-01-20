// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package alphabet

import (
	"fmt"
)

func ExampleGeneric_AllValid() {
	fmt.Println(DNA.AllValid([]Letter("acgatcgatatagctatnagcatgc")))
	// Output:
	// false 17
}

func ExamplePairing_ComplementOf() {
	var (
		c  Letter
		ok bool
	)

	c, ok = DNA.(Complementor).Complement('a')
	fmt.Printf("%c %v\n", c, ok)
	c, ok = DNA.(Complementor).Complement('n')
	fmt.Printf("%c %v\n", c, ok)
	c, ok = RNA.(Complementor).Complement('a')
	fmt.Printf("%c %v\n", c, ok)
	_, ok = RNA.(Complementor).Complement('t')
	fmt.Printf("%v\n", ok)
	// Output:
	// t true
	// n true
	// u true
	// false
}
