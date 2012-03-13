package concurrent

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

import (
	"fmt"
	"testing"
)

func TestLazy(t *testing.T) {
}

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

	mutator := Lazily(
		func(state ...interface{}) (interface{}, State) {
			b, c := []byte(state[0].(string)), state[1].(int)
			b[c] = ROT13(b[c])
			ms := string(b)
			return state[0], State{ms, (c + 1) % len(ms)}
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
