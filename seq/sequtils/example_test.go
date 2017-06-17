// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sequtils

import (
	"fmt"

	"github.com/biogo/biogo/feat"
	"github.com/biogo/biogo/seq"
)

func ExampleTruncate_a() {
	s := stringToConformRangeOffSlice("ACGCTGACTTGGTGCACGT")
	s.conf = feat.Linear
	fmt.Printf("%s\n", s)
	if err := Truncate(s, s, 5, 12); err == nil {
		fmt.Printf("%s\n", s)
	}
	// Output:
	// ACGCTGACTTGGTGCACGT
	// GACTTGG
}

func ExampleTruncate_b() {
	var (
		src = stringToConformRangeOffSlice("ACGCTGACTTGGTGCACGT")
		dst = &conformRangeOffSlice{}
	)
	src.conf = feat.Circular
	fmt.Printf("%s Conformation = %v\n", src, src.Conformation())
	if err := Truncate(dst, src, 12, 5); err == nil {
		fmt.Printf("%s\n", dst)
	} else {
		fmt.Println("Error:", err)
	}

	src.conf = feat.Linear
	fmt.Printf("%s Conformation = %v\n", src, src.Conformation())
	if err := Truncate(dst, src, 12, 5); err == nil {
		fmt.Printf("%s\n", dst)
	} else {
		fmt.Println("Error:", err)
	}
	// Output:
	// ACGCTGACTTGGTGCACGT Conformation = circular
	// TGCACGTACGCT
	// ACGCTGACTTGGTGCACGT Conformation = linear
	// Error: sequtils: start position greater than end position for linear sequence
}

func ExampleJoin() {
	var s1, s2 *offSlice

	s1 = stringToOffSlice("agctgtgctga")
	s2 = stringToOffSlice("CGTGCAGTCATGAGTGA")
	fmt.Printf("%s %s\n", s1, s2)
	Join(s1, s2, seq.Start)
	fmt.Printf("%s\n", s1)

	s1 = stringToOffSlice("agctgtgctga")
	s2 = stringToOffSlice("CGTGCAGTCATGAGTGA")
	Join(s1, s2, seq.End)
	fmt.Printf("%s\n", s1)
	// Output:
	// agctgtgctga 0 CGTGCAGTCATGAGTGA 0
	// CGTGCAGTCATGAGTGAagctgtgctga -17
	// agctgtgctgaCGTGCAGTCATGAGTGA 0
}

func ExampleStitch() {
	s := stringToConformRangeOffSlice("aAGTATAAgtcagtgcagtgtctggcagTGCTCGTGCgtagtgaagtagGGTTAGTTTa")
	f := fs{
		fe{s: 1, e: 8},
		fe{s: 28, e: 37},
		fe{s: 49, e: len(s.slice) - 1},
	}
	fmt.Printf("%s\n", s)
	if err := Stitch(s, s, f); err == nil {
		fmt.Printf("%s\n", s)
	}
	// Output:
	// aAGTATAAgtcagtgcagtgtctggcagTGCTCGTGCgtagtgaagtagGGTTAGTTTa
	// AGTATAATGCTCGTGCGGTTAGTTT
}
