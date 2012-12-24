// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package alignment

import (
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/feat"
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/exp/seq/linear"
	"code.google.com/p/biogo/exp/seq/sequtils"
	"fmt"
)

var m, n *Seq

func init() {
	var err error
	m, err = NewSeq("example alignment",
		[]string{"seq 1", "seq 2", "seq 3"},
		[][]alphabet.Letter{
			[]alphabet.Letter("AAA"),
			[]alphabet.Letter("CCC"),
			[]alphabet.Letter("GGG"),
			[]alphabet.Letter("CGA"),
			[]alphabet.Letter("TTT"),
			[]alphabet.Letter("GGG"),
			[]alphabet.Letter("AAA"),
			[]alphabet.Letter("CCC"),
			[]alphabet.Letter("TCG"),
			[]alphabet.Letter("TTT"),
			[]alphabet.Letter("GGG"),
			[]alphabet.Letter("GGG"),
			[]alphabet.Letter("TCC"),
			[]alphabet.Letter("GGG"),
			[]alphabet.Letter("CCC"),
			[]alphabet.Letter("AGT"),
			[]alphabet.Letter("CCC"),
			[]alphabet.Letter("GAA"),
			[]alphabet.Letter("TTT"),
		},
		alphabet.DNA,
		seq.DefaultConsensus)

	if err != nil {
		panic(err)
	}
}

func ExampleNewSeq() {
	m, err := NewSeq("example alignment",
		[]string{"seq 1", "seq 2", "seq 3"},
		[][]alphabet.Letter{
			[]alphabet.Letter("AAA"),
			[]alphabet.Letter("CCC"),
			[]alphabet.Letter("GGG"),
			[]alphabet.Letter("CGA"),
			[]alphabet.Letter("TTT"),
			[]alphabet.Letter("GGG"),
			[]alphabet.Letter("AAA"),
			[]alphabet.Letter("CCC"),
			[]alphabet.Letter("TCG"),
			[]alphabet.Letter("TTT"),
			[]alphabet.Letter("GGG"),
			[]alphabet.Letter("GGG"),
			[]alphabet.Letter("TCC"),
			[]alphabet.Letter("GGG"),
			[]alphabet.Letter("CCC"),
			[]alphabet.Letter("AGT"),
			[]alphabet.Letter("CCC"),
			[]alphabet.Letter("GAA"),
			[]alphabet.Letter("TTT"),
		},
		alphabet.DNA,
		seq.DefaultConsensus)
	if err == nil {
		fmt.Printf("%-s\n\n%-s\n", m, m.Consensus(false))
	}

	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGATGACGTGGCGCTCAT
	// 
	// acgntgacntggcgcncat
}

func ExampleSeq_Add() {
	fmt.Printf("%v %-s\n", m.Rows(), m.Consensus(false))
	m.Add(linear.NewQSeq("example DNA",
		[]alphabet.QLetter{{'a', 40}, {'c', 39}, {'g', 40}, {'C', 38}, {'t', 35}, {'g', 20}},
		alphabet.DNA, alphabet.Sanger))
	fmt.Printf("%v %-s\n", m.Rows(), m.Consensus(false))
	// Output:
	// 3 acgntgacntggcgcncat
	// 4 acgctgacntggcgcncat
}

func ExampleSeq_Copy() {
	n = m.Clone().(*Seq)
	n.Row(2).Set(3, alphabet.QLetter{L: 't'})
	fmt.Printf("%-s\n\n%-s\n\n", m, m.Consensus(false))
	fmt.Printf("%-s\n\n%-s\n", n, n.Consensus(false))
	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGATGACGTGGCGCTCAT
	// acgCtg-------------
	// 
	// acgctgacntggcgcncat
	// 
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGtTGACGTGGCGCTCAT
	// acgCtg-------------
	// 
	// acgctgacntggcgcncat
}

func ExampleSeq_Count() {
	fmt.Println(m.Rows())
	// Output:
	// 4
}

func ExampleSeq_Join() {
	fmt.Printf("%-s\n\n%-s\n", n, n.Consensus(false))
	err := sequtils.Join(n, m, seq.End)
	if err == nil {
		fmt.Printf("\n%-s\n\n%-s\n", n, n.Consensus(false))
	}
	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGtTGACGTGGCGCTCAT
	// acgCtg-------------
	// 
	// acgctgacntggcgcncat
	// 
	// ACGCTGACTTGGTGCACGTACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCATACGGTGACCTGGCGCGCAT
	// ACGtTGACGTGGCGCTCATACGATGACGTGGCGCTCAT
	// acgCtg-------------acgCtg-------------
	// 
	// acgctgacntggcgcncatacgctgacntggcgcncat
}

func ExampleAlignment_Len() {
	fmt.Println(m.Len())
	// Output:
	// 19
}

func ExampleSeq_RevComp() {
	fmt.Printf("%-s\n\n%-s\n", m, m.Consensus(false))
	fmt.Println()
	m.RevComp()
	fmt.Printf("%-s\n\n%-s\n", m, m.Consensus(false))
	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGATGACGTGGCGCTCAT
	// acgCtg-------------
	// 
	// acgctgacntggcgcncat
	// 
	// ACGTGCACCAAGTCAGCGT
	// ATGCGCGCCAGGTCACCGT
	// ATGAGCGCCACGTCATCGT
	// -------------caGcgt
	// 
	// atgngcgccangtcagcgt
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

func ExampleSeq_Stitch() {
	f := fs{
		&fe{s: -1, e: 4},
		&fe{s: 30, e: 38},
	}
	fmt.Printf("%-s\n\n%-s\n", n, n.Consensus(false))
	if err := sequtils.Stitch(n, n, f); err == nil {
		fmt.Printf("\n%-s\n\n%-s\n", n, n.Consensus(false))
	} else {
		fmt.Println(err)
	}
	// Output:
	// ACGCTGACTTGGTGCACGTACGTGCACCAAGTCAGCGT
	// ACGGTGACCTGGCGCGCATATGCGCGCCAGGTCACCGT
	// ACGtTGACGTGGCGCTCATATGAGCGCCACGTCATCGT
	// acgCtg--------------------------caGcgt
	// 
	// acgctgacntggcgcncatatgngcgccangtcagcgt
	// 
	// ACGCGTCAGCGT
	// ACGGGTCACCGT
	// ACGtGTCATCGT
	// acgC--caGcgt
	// 
	// acgcgtcagcgt
}

func ExampleSeq_Truncate() {
	fmt.Printf("%-s\n\n%-s\n", m, m.Consensus(false))
	err := sequtils.Truncate(m, m, 4, 12)
	if err == nil {
		fmt.Printf("\n%-s\n\n%-s\n", m, m.Consensus(false))
	}
	// Output:
	// ACGTGCACCAAGTCAGCGT
	// ATGCGCGCCAGGTCACCGT
	// ATGAGCGCCACGTCATCGT
	// -------------caGcgt
	// 
	// atgngcgccangtcagcgt
	// 
	// GCACCAAG
	// GCGCCAGG
	// GCGCCACG
	// --------
	// 
	// gcgccang
}
