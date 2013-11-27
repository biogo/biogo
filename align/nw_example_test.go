// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package align

import (
	"code.google.com/p/biogo/alphabet"
	"code.google.com/p/biogo/seq/linear"

	"fmt"
)

func ExampleNW_Align() {
	nwsa := &linear.Seq{Seq: alphabet.BytesToLetters([]byte("AGACTAGTTA"))}
	nwsa.Alpha = alphabet.DNA
	nwsb := &linear.Seq{Seq: alphabet.BytesToLetters([]byte("GACAGACG"))}
	nwsb.Alpha = alphabet.DNA

	//		   Query letter
	//  	 A	 C	 G	 T	 -
	// A	10	-3	-1	-4	-5
	// C	-3	 9	-5	 0	-5
	// G	-1	-5	 7	-3	-5
	// T	-4	 0	-3	 8	-5
	// -	-5	-5	-5	-5	 0
	needle := NW{
		{10, -3, -1, -4, -5},
		{-3, 9, -5, 0, -5},
		{-1, -5, 7, -3, -5},
		{-4, 0, -3, 8, -5},
		{-5, -5, -5, -5, 0},
	}

	aln, err := needle.Align(nwsa, nwsb)
	if err == nil {
		fmt.Printf("%s\n", aln)
		fa := Format(nwsa, nwsb, aln, '-')
		fmt.Printf("%s\n%s\n", fa[0], fa[1])
	}
	// Output:
	//[[0,1)/-=-5 [1,4)/[0,3)=26 [4,5)/-=-5 [5,10)/[3,8)=12]
	// AGACTAGTTA
	// -GAC-AGACG
}

func ExampleNWAffine_Align() {
	nwsa := &linear.Seq{Seq: alphabet.BytesToLetters([]byte("ATAGGAAG"))}
	nwsa.Alpha = alphabet.DNA
	nwsb := &linear.Seq{Seq: alphabet.BytesToLetters([]byte("ATTGGCAATG"))}
	nwsb.Alpha = alphabet.DNA

	//		   Query letter
	//  	 A	 C	 G	 T	 -
	// A	 1	-1	-1	-1	-1
	// C	-1	 1	-1	 1	-1
	// G	-1	-1	 1	-1	-1
	// T	-1	 1	-1	 1	-1
	// -	-1	-1	-1	-1	 0
	//
	// Gap open: -5
	needle := NWAffine{
		Matrix: Linear{
			{1, -1, -1, -1, -1},
			{-1, 1, -1, -1, -1},
			{-1, -1, 1, -1, -1},
			{-1, -1, -1, 1, -1},
			{-1, -1, -1, -1, 0},
		},
		GapOpen: -5,
	}

	aln, err := needle.Align(nwsa, nwsb)
	if err == nil {
		fmt.Printf("%s\n", aln)
		fa := Format(nwsa, nwsb, aln, '-')
		fmt.Printf("%s\n%s\n", fa[0], fa[1])
	}
	// Output:
	// [[0,7)/[0,7)=3 -/[7,9)=-7 [7,8)/[9,10)=1]
	// ATAGGAA--G
	// ATTGGCAATG
}
