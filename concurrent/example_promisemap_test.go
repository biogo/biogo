// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package concurrent_test

import (
	"code.google.com/p/biogo/concurrent"
	"fmt"
	"time"
)

type SlowCounter []int

func (c SlowCounter) Slice(i, j int) concurrent.Mapper { return c[i:j] }
func (c SlowCounter) Len() int                         { return len(c) }

func (c SlowCounter) Operation() (r interface{}, err error) {
	var sum int
	for _, v := range c {
		sum += v
		time.Sleep(1e8)
	}
	return sum, nil
}

func ExamplePromiseMap() {
	c := SlowCounter{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	p := concurrent.PromiseMap(c, 1, 2)
	fmt.Println("Waiting...")
	request1 := <-p.Wait()
	if request1.Err != nil {
		fmt.Println(request1.Err)
	} else {
		fmt.Println(request1.Value)
	}
	request2 := <-p.Wait()
	if request2.Err != nil {
		fmt.Println(request2.Err)
	} else {
		fmt.Println(request2.Value)
	}

	// Output:
	// Waiting...
	// [3 7 11 15 19]
	// [3 7 11 15 19]
}
