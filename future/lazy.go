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

// Evaluation function
type Eval func(...interface{}) (interface{}, []interface{})

// Simple implementation of a lazy function generator with control of spawned goroutine.
// 
// It is probably worth while wrapping this in a function that returns the result as the type you are actually using.
// e.g.
//
//   func LazilyString(f Eval, rc chan string, reaper <-chan struct{}, init ...string) func() (string) {
//       return func() string {
//           return Lazily(f, rc, reaper, init...)().(string)
//       }
//   }
//
// Lazy functions are terminated by closing the reaper channel. nil should be passed as
// a reaper for perpetual lazy functions.
//
// Function to generate the lazy evaluator
func Lazily(f Eval, rc chan interface{}, reaper <-chan struct{}, init ...interface{}) func() interface{} {
	go func() {
		defer close(rc)
		var state []interface{} = init
		var result interface{}

		for {
			result, state = f(state...)
			select {
			case rc <- result:
			case <-reaper:
				return
			}
		}
	}()

	return func() interface{} {
		return <-rc
	}
}
