// Copyright ©2011-2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
// 
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// 
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// 
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

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
