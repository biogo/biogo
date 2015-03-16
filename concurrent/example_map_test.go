// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package concurrent_test

import (
	"github.com/biogo/biogo/concurrent"

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

	for c.Len() > 1 {
		result, err := concurrent.Map(c, 1, 2)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(result)
			c = c[:0]
			for _, r := range result {
				c = append(c, r.(int))
			}
		}
	}

	// Output:
	// [1 2 3 4 5 6 7 8 9 10]
	// [3 7 11 15 19]
	// [10 26 19]
	// [36 19]
	// [55]
}
