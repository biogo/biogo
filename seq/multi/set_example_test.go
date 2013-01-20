// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package multi

import (
	"code.google.com/p/biogo/alphabet"
	"code.google.com/p/biogo/seq/linear"
	"fmt"
)

var set Set

func ExampleSet_AppendEach() {
	ss := [][]alphabet.Letter{
		[]alphabet.Letter("ACGCTGACTTGGTGCACGT"),
		[]alphabet.Letter("ACGACTGGGACGT"),
		[]alphabet.Letter("ACGCTGACTGGCCGT"),
		[]alphabet.Letter("GCCTTTGCACGT"),
	}
	set = make(Set, 4)
	for i := range set {
		set[i] = linear.NewSeq(fmt.Sprintf("example DNA %d", i), ss[i], alphabet.DNA)
	}
	as := [][]alphabet.QLetter{
		alphabet.QLetter{L: 'A'}.Repeat(2),
		alphabet.QLetter{L: 'C'}.Repeat(2),
		alphabet.QLetter{L: 'G'}.Repeat(2),
		alphabet.QLetter{L: 'T'}.Repeat(2),
	}

	set.AppendEach(as)

	for _, s := range set {
		fmt.Printf("%-s\n", s)
	}
	// Output:
	// ACGCTGACTTGGTGCACGTAA
	// ACGACTGGGACGTCC
	// ACGCTGACTGGCCGTGG
	// GCCTTTGCACGTTT
}

func ExampleSet_Rows() {
	fmt.Println(set.Rows())
	// Output:
	// 4
}

func ExampleSet_Get() {
	fmt.Printf("%-s\n", set.Row(2))
	// Output:
	// ACGCTGACTGGCCGTGG
}

func ExampleSet_Len() {
	fmt.Println(set.Len())
	// Output:
	// 21
}

func ExampleSet_RevComp() {
	set.RevComp()
	for _, s := range set {
		fmt.Println(s)
	}
}

func ExampleSet_Reverse() {
	set.RevComp()
	for _, s := range set {
		fmt.Println(s)
	}
}
