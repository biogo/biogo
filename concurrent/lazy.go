// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package concurrent

// Evaluator is a function for lazy evaluation.
type Evaluator func(...interface{}) (interface{}, State)

type State []interface{}

// Lazily is function to generate a lazy evaluator.
//
// Lazy functions are terminated by closing the reaper channel. nil should be passed as
// a reaper for perpetual lazy functions.
func Lazily(f Evaluator, lookahead int, reaper <-chan struct{}, init ...interface{}) func() interface{} {
	rc := make(chan interface{}, lookahead)
	go func(rc chan interface{}) {
		defer close(rc)
		var state State = init
		var result interface{}

		for {
			result, state = f(state...)
			select {
			case rc <- result:
			case <-reaper:
				return
			}
		}
	}(rc)

	return func() interface{} {
		return <-rc
	}
}
