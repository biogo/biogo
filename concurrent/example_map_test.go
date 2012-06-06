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

package concurrent_test

import (
	"code.google.com/p/biogo/concurrent"
	"fmt"
)

type CountConsumer []int

func (c CountConsumer) Slice(i, j int) concurrent.Mapper { return c[i:j] }
func (c CountConsumer) Len() int                         { return len(c) }

func (c CountConsumer) Operation() (r interface{}, err error) {
	var sum int
	for i, v := range c {
		sum += v
		c[i] = 0
	}
	return sum, nil
}

func ExampleMap() {
	c := CountConsumer{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	fmt.Println(c)

	result, err := concurrent.Map(c, 1, 2)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(result)
	}

	fmt.Println(c)

	// Output:
	// [1 2 3 4 5 6 7 8 9 10]
	// [3 7 11 15 19]
	// [0 0 0 0 0 0 0 0 0 0]
}
