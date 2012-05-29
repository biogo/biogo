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

package concurrent

import (
	"fmt"
	"time"
)

type CountConsumer []int

func (c CountConsumer) Slice(i, j int) Mapper { return c[i:j] }
func (c CountConsumer) Len() int              { return len(c) }

func (c CountConsumer) Operation() (r interface{}, err error) {
	var sum int
	for i, v := range c {
		sum += v
		c[i] = 0
	}
	return sum, nil
}

func ExampleMap() {
	// Given:
	//
	//	type CountConsumer []int
	//
	//	func (c CountConsumer) Slice(i, j int) Mapper { return c[i:j] }
	//	func (c CountConsumer) Len() int              { return len(c) }
	//
	//	func (c CountConsumer) Operator() (r interface{}, err error) {
	//		var sum int
	//		for _, v := range c {
	//			sum += v
	//			c[i] = 0
	//		}
	//		return sum, nil
	//	}

	c := CountConsumer{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	fmt.Println(c)

	result, err := Map(c, 1, 2)
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

type SlowCounter []int

func (c SlowCounter) Slice(i, j int) Mapper { return c[i:j] }
func (c SlowCounter) Len() int              { return len(c) }

func (c SlowCounter) Operation() (r interface{}, err error) {
	var sum int
	for _, v := range c {
		sum += v
		time.Sleep(1e8)
	}
	return sum, nil
}

func ExamplePromiseMap() {
	// Given:
	//
	//	type SlowCounter []int
	//	
	//	func (c SlowCounter) Slice(i, j int) Mapper { return c[i:j] }
	//	func (c SlowCounter) Len() int              { return len(c) }
	//	
	//	func (c SlowCounter) Operation() (r interface{}, err error) {
	//		var sum int
	//		for _, v := range c {
	//			sum += v
	//			time.Sleep(1e8)
	//		}
	//		return sum, nil
	//	}

	c := SlowCounter{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	p := PromiseMap(c, 1, 2)
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
