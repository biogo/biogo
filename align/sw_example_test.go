// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package align

import (
	"code.google.com/p/biogo/alphabet"
	"code.google.com/p/biogo/seq/linear"
	"fmt"
)

func ExampleSW_Align_1() {
	swsa := &linear.Seq{Seq: alphabet.BytesToLetters([]byte("ACACACTA"))}
	swsa.Alpha = alphabet.DNA
	swsb := &linear.Seq{Seq: alphabet.BytesToLetters([]byte("AGCACACA"))}
	swsb.Alpha = alphabet.DNA

	// w(gap) = -1
	// w(match) = +2
	// w(mismatch) = -1
	smith := SW{
		{2, -1, -1, -1, -1},
		{-1, 2, -1, -1, -1},
		{-1, -1, 2, -1, -1},
		{-1, -1, -1, 2, -1},
		{-1, -1, -1, -1, 0},
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

func ExampleSW_Align_2() {
	swsa := &linear.Seq{Seq: alphabet.BytesToLetters([]byte("AAAATTTAAAA"))}
	swsa.Alpha = alphabet.DNA
	swsb := &linear.Seq{Seq: alphabet.BytesToLetters([]byte("AAAAGGGAAAA"))}
	swsb.Alpha = alphabet.DNA

	// w(gap) = 0
	// w(match) = +2
	// w(mismatch) = -1
	smith := SW{
		{2, -1, -1, -1, 0},
		{-1, 2, -1, -1, 0},
		{-1, -1, 2, -1, 0},
		{-1, -1, -1, 2, 0},
		{0, 0, 0, 0, 0},
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
