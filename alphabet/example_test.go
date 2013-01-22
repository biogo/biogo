// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package alphabet

import (
	"fmt"
)

func Example_1() {
	fmt.Println(DNA.AllValid([]Letter("acgatcgatatagctatnagcatgc")))
	// Output:
	// false 17
}

func Example_2() {
	var (
		c  Letter
		ok bool
	)

	c, ok = DNA.Complement('a')
	fmt.Printf("%c %v\n", c, ok)
	c, ok = DNA.Complement('n')
	fmt.Printf("%c %v\n", c, ok)
	c, ok = RNA.Complement('a')
	fmt.Printf("%c %v\n", c, ok)
	c, ok = RNA.Complement('t')
	fmt.Printf("%c %v\n", c, ok)
	// Output:
	// t true
	// n true
	// u true
	// t false
}
