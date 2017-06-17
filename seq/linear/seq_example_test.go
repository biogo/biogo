// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linear

import (
	"fmt"
	"strings"

	"github.com/biogo/biogo/alphabet"
	"github.com/biogo/biogo/feat"
	"github.com/biogo/biogo/seq"
	"github.com/biogo/biogo/seq/sequtils"
)

func ExampleNewSeq() {
	d := NewSeq("example DNA", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.DNA)
	fmt.Printf("%-s %v\n", d, d.Moltype())
	// Output:
	// ACGCTGACTTGGTGCACGT DNA
}

func ExampleSeq_Validate() {
	r := NewSeq("example RNA", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.RNA)
	fmt.Printf("%-s %v\n", r, r.Moltype())
	if ok, pos := r.Validate(); ok {
		fmt.Println("valid RNA")
	} else {
		fmt.Println(strings.Repeat(" ", pos-1), "^ first invalid RNA position")
	}
	// Output:
	// ACGCTGACTTGGTGCACGT RNA
	//     ^ first invalid RNA position
}

func ExampleSeq_truncate_a() {
	s := NewSeq("example DNA", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.DNA)
	fmt.Printf("%-s\n", s)
	if err := sequtils.Truncate(s, s, 5, 12); err == nil {
		fmt.Printf("%-s\n", s)
	}
	// Output:
	// ACGCTGACTTGGTGCACGT
	// GACTTGG
}

func ExampleSeq_truncate_b() {
	var s *Seq

	s = NewSeq("example DNA", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.DNA)
	s.Conform = feat.Circular
	fmt.Printf("%-s Conformation = %v\n", s, s.Conformation())
	if err := sequtils.Truncate(s, s, 12, 5); err == nil {
		fmt.Printf("%-s Conformation = %v\n", s, s.Conformation())
	} else {
		fmt.Println("Error:", err)
	}

	s = NewSeq("example DNA", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.DNA)
	fmt.Printf("%-s Conformation = %v\n", s, s.Conformation())
	if err := sequtils.Truncate(s, s, 12, 5); err == nil {
		fmt.Printf("%-s Conformation = %v\n", s, s.Conformation())
	} else {
		fmt.Println("Error:", err)
	}
	// Output:
	// ACGCTGACTTGGTGCACGT Conformation = circular
	// TGCACGTACGCT Conformation = linear
	// ACGCTGACTTGGTGCACGT Conformation = linear
	// Error: sequtils: start position greater than end position for linear sequence
}

func ExampleSeq_RevComp() {
	s := NewSeq("example DNA", []alphabet.Letter("ATGCtGACTTGGTGCACGT"), alphabet.DNA)
	fmt.Printf("%-s\n", s)
	s.RevComp()
	fmt.Printf("%-s\n", s)
	// Output:
	// ATGCtGACTTGGTGCACGT
	// ACGTGCACCAAGTCaGCAT
}

func ExampleSeq_join() {
	var s1, s2 *Seq

	s1 = NewSeq("a", []alphabet.Letter("agctgtgctga"), alphabet.DNA)
	s2 = NewSeq("b", []alphabet.Letter("CGTGCAGTCATGAGTGA"), alphabet.DNA)
	fmt.Printf("%-s %-s\n", s1, s2)
	if err := sequtils.Join(s1, s2, seq.Start); err == nil {
		fmt.Printf("%-s\n", s1)
	}

	s1 = NewSeq("a", []alphabet.Letter("agctgtgctga"), alphabet.DNA)
	s2 = NewSeq("b", []alphabet.Letter("CGTGCAGTCATGAGTGA"), alphabet.DNA)
	if err := sequtils.Join(s1, s2, seq.End); err == nil {
		fmt.Printf("%-s\n", s1)
	}
	// Output:
	// agctgtgctga CGTGCAGTCATGAGTGA
	// CGTGCAGTCATGAGTGAagctgtgctga
	// agctgtgctgaCGTGCAGTCATGAGTGA
}

type fe struct {
	s, e int
	st   seq.Strand
	feat.Feature
}

func (f fe) Start() int                    { return f.s }
func (f fe) End() int                      { return f.e }
func (f fe) Len() int                      { return f.e - f.s }
func (f fe) Orientation() feat.Orientation { return feat.Orientation(f.st) }

type fs []feat.Feature

func (f fs) Features() []feat.Feature { return []feat.Feature(f) }

func ExampleSeq_stitch() {
	s := NewSeq("example DNA", []alphabet.Letter("aAGTATAAgtcagtgcagtgtctggcagTGCTCGTGCgtagtgaagtagGGTTAGTTTa"), alphabet.DNA)
	f := fs{
		fe{s: 1, e: 8},
		fe{s: 28, e: 37},
		fe{s: 49, e: s.Len() - 1},
	}
	fmt.Printf("%-s\n", s)
	if err := sequtils.Stitch(s, s, f); err == nil {
		fmt.Printf("%-s\n", s)
	}
	// Output:
	// aAGTATAAgtcagtgcagtgtctggcagTGCTCGTGCgtagtgaagtagGGTTAGTTTa
	// AGTATAATGCTCGTGCGGTTAGTTT
}

func ExampleSeq_compose() {
	s := NewSeq("example DNA", []alphabet.Letter("aAGTATAAgtcagtgcagtgtctggcag<TS>gtagtgaagtagggttagttta"), alphabet.DNA)
	f := fs{
		fe{s: 0, e: 32},
		fe{s: 1, e: 8, st: -1},
		fe{s: 28, e: s.Len() - 1},
	}
	fmt.Printf("%-s\n", s)
	if err := sequtils.Compose(s, s, f); err == nil {
		fmt.Printf("%-s\n", s)
	}
	// Output:
	// aAGTATAAgtcagtgcagtgtctggcag<TS>gtagtgaagtagggttagttta
	// aAGTATAAgtcagtgcagtgtctggcag<TS>TTATACT<TS>gtagtgaagtagggttagttt
}
