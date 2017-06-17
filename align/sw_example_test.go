// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package align

import (
	"github.com/biogo/biogo/alphabet"
	"github.com/biogo/biogo/seq/linear"

	"fmt"
)

func ExampleSW_Align_a() {
	swsa := &linear.Seq{Seq: alphabet.BytesToLetters([]byte("ACACACTA"))}
	swsa.Alpha = alphabet.DNAgapped
	swsb := &linear.Seq{Seq: alphabet.BytesToLetters([]byte("AGCACACA"))}
	swsb.Alpha = alphabet.DNAgapped

	// w(gap) = -1
	// w(match) = +2
	// w(mismatch) = -1
	smith := SW{
		{0, -1, -1, -1, -1},
		{-1, 2, -1, -1, -1},
		{-1, -1, 2, -1, -1},
		{-1, -1, -1, 2, -1},
		{-1, -1, -1, -1, 2},
	}

	aln, err := smith.Align(swsa, swsb)
	if err == nil {
		fmt.Printf("%v\n", aln)
		fa := Format(swsa, swsb, aln, '-')
		fmt.Printf("%s\n%s\n", fa[0], fa[1])
	}
	// Output:
	// [[0,1)/[0,1)=2 -/[1,2)=-1 [1,6)/[2,7)=10 [6,7)/-=-1 [7,8)/[7,8)=2]
	// A-CACACTA
	// AGCACAC-A
}

func ExampleSW_Align_b() {
	swsa := &linear.Seq{Seq: alphabet.BytesToLetters([]byte("AAAATTTAAAA"))}
	swsa.Alpha = alphabet.DNAgapped
	swsb := &linear.Seq{Seq: alphabet.BytesToLetters([]byte("AAAAGGGAAAA"))}
	swsb.Alpha = alphabet.DNAgapped

	// w(gap) = 0
	// w(match) = +2
	// w(mismatch) = -1
	smith := SW{
		{0, 0, 0, 0, 0},
		{0, 2, -1, -1, -1},
		{0, -1, 2, -1, -1},
		{0, -1, -1, 2, -1},
		{0, -1, -1, -1, 2},
	}

	aln, err := smith.Align(swsa, swsb)
	if err == nil {
		fmt.Printf("%v\n", aln)
		fa := Format(swsa, swsb, aln, '-')
		fmt.Printf("%s\n%s\n", fa[0], fa[1])
	}
	// Output:
	// [[0,4)/[0,4)=8 -/[4,7)=0 [4,7)/-=0 [7,11)/[7,11)=8]
	// AAAA---TTTAAAA
	// AAAAGGG---AAAA
}

func ExampleSWAffine_Align() {
	swsa := &linear.Seq{Seq: alphabet.BytesToLetters([]byte("ATAGGAAG"))}
	swsa.Alpha = alphabet.DNAgapped
	swsb := &linear.Seq{Seq: alphabet.BytesToLetters([]byte("ATTGGCAATG"))}
	swsb.Alpha = alphabet.DNAgapped

	//		   Query letter
	//  	 -	 A	 C	 G	 T
	// -	 0	-1	-1	-1	-1
	// A	-1	 1	-1	-1	-1
	// C	-1	-1	 1	-1	-1
	// G	-1	-1	-1	 1	-1
	// T	-1	-1	-1	-1	 1
	//
	// Gap open: -5
	smith := SWAffine{
		Matrix: Linear{
			{0, -1, -1, -1, -1},
			{-1, 1, -1, -1, -1},
			{-1, -1, 1, -1, -1},
			{-1, -1, -1, 1, -1},
			{-1, -1, -1, -1, 1},
		},
		GapOpen: -5,
	}

	aln, err := smith.Align(swsa, swsb)
	if err == nil {
		fmt.Printf("%s\n", aln)
		fa := Format(swsa, swsb, aln, '-')
		fmt.Printf("%s\n%s\n", fa[0], fa[1])
	}
	// Output:
	// [[0,7)/[0,7)=3]
	// ATAGGAA
	// ATTGGCA
}
