// Future handling package
package future
// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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

import "runtime"

// Evaluation function
type Eval func(...interface{}) (interface{}, []interface{})

// Simple implementation of a lazy function generator with control of spawned goroutine.
// 
// It is probably worth while wrapping this in a function that returns the result as the type you are actually using.
// e.g.
//
//   func LazilyString(f Eval, rc chan string, init ...string) func() (string) {
//       return func() string {
//           return Lazily(f, rc, init...)().(string)
//       }
//   }
//
// Lazy functions will result in goroutine leaks. TODO
//
// Function to generate the lazy evaluator
func Lazily(f Eval, rc chan interface{}, init ...interface{}) func() interface{} {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if e, ok := r.(runtime.Error); ok {
					if e.Error() == "runtime error: send on closed channel" {
						return
					}
				}
				panic(r)
			}
		}()

		var state []interface{} = init
		var result interface{}

		for {
			result, state = f(state...)
			rc <- result
		}
	}()

	return func() interface{} {
		return <-rc
	}
}
