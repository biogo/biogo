// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package concurrent_test

import (
	"code.google.com/p/biogo/concurrent"

	"fmt"
)

func ExampleLazily() {
	sentence := "A sentence to be slowly ROT'ed."

	ROT13 := func(b byte) byte {
		c := b & ('a' - 'A')
		i := b&^c - 'A'
		if i < 26 {
			return (i+13)%26 + 'A' | c
		}
		return b
	}

	mutator := concurrent.Lazily(
		func(state ...interface{}) (interface{}, concurrent.State) {
			b, c := []byte(state[0].(string)), state[1].(int)
			b[c] = ROT13(b[c])
			ms := string(b)
			return state[0], concurrent.State{ms, (c + 1) % len(ms)}
		},
		0,           // No lookahead
		nil,         // Perpetual evaluator
		sentence, 0, // Initial state
	)

	var r string
	for i := 0; i < len(sentence)*2; i++ {
		r = mutator().(string)
		if i%10 == 0 {
			fmt.Println(r)
		}
	}
	fmt.Println(r)
	// Output:
	// A sentence to be slowly ROT'ed.
	// N fragrapr to be slowly ROT'ed.
	// N fragrapr gb or fybwly ROT'ed.
	// N fragrapr gb or fybjyl EBG'rq.
	// A sentencr gb or fybjyl EBG'rq.
	// A sentence to be slbjyl EBG'rq.
	// A sentence to be slowly ROT'eq.
	// A sentence to be slowly ROT'ed.
}
