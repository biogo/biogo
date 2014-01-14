// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package align

import (
	"code.google.com/p/biogo/alphabet"
	"code.google.com/p/biogo/seq/linear"

	"fmt"
)

func ExampleFitted_Align() {
	fsa := &linear.Seq{Seq: alphabet.BytesToLetters([]byte("GTTGACAGACTAGATTCACG"))}
	fsa.Alpha = alphabet.DNAgapped
	fsb := &linear.Seq{Seq: alphabet.BytesToLetters([]byte("GACAGACGA"))}
	fsb.Alpha = alphabet.DNAgapped

	//		   Query letter
	//  	 -	 A	 C	 G	 T
	// -	 0	-5	-5	-5	-5
	// A	-5	10	-3	-1	-4
	// C	-5	-3	 9	-5	 0
	// G	-5	-1	-5	 7	-3
	// T	-5	-4	 0	-3	 8
	fitted := Fitted{
		{0, -5, -5, -5, -5},
		{-5, 10, -3, -1, -4},
		{-5, -3, 9, -5, 0},
		{-5, -1, -5, 7, -3},
		{-5, -4, 0, -3, 8},
	}

	aln, err := fitted.Align(fsa, fsb)
	if err == nil {
		fmt.Printf("%s\n", aln)
		fa := Format(fsa, fsb, aln, '-')
		fmt.Printf("%s\n%s\n", fa[0], fa[1])
	}
	// Output:
	// [[3,10)/[0,7)=62 [10,12)/-=-10 [12,14)/[7,9)=17]
	// GACAGACTAGA
	// GACAGAC--GA
}

func ExampleFittedAffine_Align() {
	fsa := &linear.Seq{Seq: alphabet.BytesToLetters([]byte("ATTGGCAATGA"))}
	fsa.Alpha = alphabet.DNAgapped
	fsb := &linear.Seq{Seq: alphabet.BytesToLetters([]byte("ATAGGAA"))}
	fsb.Alpha = alphabet.DNAgapped

	//		   Query letter
	//  	 -	 A	 C	 G	 T
	// -	 0	-1	-1	-1	-1
	// A	-1	 1	-1	-1	-1
	// C	-1	-1	 1	-1	-1
	// G	-1	-1	-1	 1	-1
	// T	-1	-1	-1	-1	 1
	//
	// Gap open: -5
	fitted := FittedAffine{
		Matrix: Linear{
			{0, -1, -1, -1, -1},
			{-1, 1, -1, -1, -1},
			{-1, -1, 1, -1, -1},
			{-1, -1, -1, 1, -1},
			{-1, -1, -1, -1, 1},
		},
		GapOpen: -5,
	}

	aln, err := fitted.Align(fsa, fsb)
	if err == nil {
		fmt.Printf("%s\n", aln)
		fa := Format(fsa, fsb, aln, '-')
		fmt.Printf("%s\n%s\n", fa[0], fa[1])
	}
	// Output:
	// [[0,7)/[0,7)=3]
	// ATTGGCA
	// ATAGGAA
}
