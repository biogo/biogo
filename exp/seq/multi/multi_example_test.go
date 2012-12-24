// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package multi

import (
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/feat"
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/exp/seq/linear"
	"fmt"
)

var m, n *Multi

func init() {
	var err error
	m, err = NewMulti("example multi",
		[]seq.Sequence{
			linear.NewSeq("example DNA 1", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.DNA),
			linear.NewSeq("example DNA 2", []alphabet.Letter("ACGGTGACCTGGCGCGCAT"), alphabet.DNA),
			linear.NewSeq("example DNA 3", []alphabet.Letter("ACGATGACGTGGCGCTCAT"), alphabet.DNA),
		},
		seq.DefaultConsensus)

	if err != nil {
		panic(err)
	}
}

func ExampleNewMulti() {
	m, err := NewMulti("example multi",
		[]seq.Sequence{
			linear.NewSeq("example DNA 1", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.DNA),
			linear.NewSeq("example DNA 2", []alphabet.Letter("ACGGTGACCTGGCGCGCAT"), alphabet.DNA),
			linear.NewSeq("example DNA 3", []alphabet.Letter("ACGATGACGTGGCGCTCAT"), alphabet.DNA),
		},
		seq.DefaultConsensus)

	if err != nil {
		return
	}

	fmt.Printf("%- s\n\n%-s\n", m, m.Consensus(false))
	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGATGACGTGGCGCTCAT
	// 
	// acgntgacntggcgcncat
}

func ExampleMulti_Add() {
	var err error
	fmt.Printf("%v %-s\n", m.Rows(), m.Consensus(false))
	err = m.Add(linear.NewQSeq("example DNA",
		[]alphabet.QLetter{{'a', 40}, {'c', 39}, {'g', 40}, {'C', 38}, {'t', 35}, {'g', 20}},
		alphabet.DNA, alphabet.Sanger))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%v %-s\n", m.Rows(), m.Consensus(false))
	err = m.Add(linear.NewQSeq("example RNA",
		[]alphabet.QLetter{{'a', 40}, {'c', 39}, {'g', 40}, {'C', 38}, {'t', 35}, {'g', 20}},
		alphabet.RNA, alphabet.Sanger))
	if err != nil {
		fmt.Println(err)
		return
	}
	// Output:
	// 3 acgntgacntggcgcncat
	// 4 acgctgacntggcgcncat
	// multi: inconsistent alphabets
}

func ExampleMulti_Copy() {
	n = m.Clone().(*Multi)
	n.Row(2).Set(3, alphabet.QLetter{L: 't'})
	fmt.Printf("%- s\n\n%-s\n\n%- s\n\n%-s\n",
		m, m.Consensus(false),
		n, n.Consensus(false),
	)
	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGATGACGTGGCGCTCAT
	// acgCtg
	// 
	// acgctgacntggcgcncat
	// 
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGtTGACGTGGCGCTCAT
	// acgCtg
	// 
	// acgctgacntggcgcncat
}

func ExampleMulti_Count() {
	fmt.Println(m.Rows())
	// Output:
	// 4
}

func ExampleMulti_IsFlush() {
	m.Row(3).SetOffset(13)
	fmt.Printf("%- s\n\n%-s\n", m, m.Consensus(false))
	fmt.Printf("\nFlush at left: %v\nFlush at right: %v\n", m.IsFlush(seq.Start), m.IsFlush(seq.End))
	m.Flush(seq.Start, '-')
	fmt.Printf("\n%- s\n\n%-s\n", m, m.Consensus(false))
	fmt.Printf("\nFlush at left: %v\nFlush at right: %v\n", m.IsFlush(seq.Start), m.IsFlush(seq.End))
	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGATGACGTGGCGCTCAT
	//              acgCtg
	// 
	// acgntgacntggcgcgcat
	// 
	// Flush at left: false
	// Flush at right: true
	// 
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGATGACGTGGCGCTCAT
	// -------------acgCtg
	// 
	// acgntgacntggcgcgcat
	//
	// Flush at left: true
	// Flush at right: true
}

func ExampleMulti_Join() {
	fmt.Printf("%- s\n\n%-s\n", n, n.Consensus(false))
	n.Join(m, seq.End)
	fmt.Printf("\n%- s\n\n%-s\n", n, n.Consensus(false))
	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGtTGACGTGGCGCTCAT
	// acgCtg
	// 
	// acgctgacntggcgcncat
	// 
	// ACGCTGACTTGGTGCACGTACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCATACGGTGACCTGGCGCGCAT
	// ACGtTGACGTGGCGCTCATACGATGACGTGGCGCTCAT
	// acgCtg--------------------------acgCtg
	// 
	// acgctgacntggcgcncatacgntgacntggcgcgcat
}

func ExampleMulti_Len() {
	fmt.Println(m.Len())
	// Output:
	// 19
}

func ExampleMulti_RevComp() {
	fmt.Printf("%- s\n\n%-s\n\n", m, m.Consensus(false))
	m.RevComp()
	fmt.Printf("%- s\n\n%-s\n", m, m.Consensus(false))
	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGATGACGTGGCGCTCAT
	// -------------acgCtg
	// 
	// acgntgacntggcgcgcat
	// 
	// ACGTGCACCAAGTCAGCGT
	// ATGCGCGCCAGGTCACCGT
	// ATGAGCGCCACGTCATCGT
	// caGcgt-------------
	// 
	// atgcgcgccangtcancgt
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

func ExampleMulti_Stitch() {
	f := fs{
		&fe{s: -1, e: 4},
		&fe{s: 30, e: 38},
	}
	fmt.Printf("%- s\n\n%-s\n\n", n, n.Consensus(false))
	if err := n.Stitch(f); err == nil {
		fmt.Printf("%- s\n\n%-s\n", n, n.Consensus(false))
	} else {
		fmt.Println(err)
	}
	// Output:
	// ACGCTGACTTGGTGCACGTACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCATACGGTGACCTGGCGCGCAT
	// ACGtTGACGTGGCGCTCATACGATGACGTGGCGCTCAT
	// acgCtg--------------------------acgCtg
	// 
	// acgctgacntggcgcncatacgntgacntggcgcgcat
	//
	// ACGCGTGCACGT
	// ACGGGCGCGCAT
	// ACGtGCGCTCAT
	// acgC--acgCtg
	// 
	// acgcgcgcgcat
}

func ExampleMulti_Truncate() {
	fmt.Printf("%- s\n\n%-s\n\n", m, m.Consensus(false))
	m.Truncate(4, 12)
	fmt.Printf("%- s\n\n%-s\n", m, m.Consensus(false))
	// Output:
	// ACGTGCACCAAGTCAGCGT
	// ATGCGCGCCAGGTCACCGT
	// ATGAGCGCCACGTCATCGT
	// caGcgt-------------
	// 
	// atgcgcgccangtcancgt
	// 
	// GCACCAAG
	// GCGCCAGG
	// GCGCCACG
	// gt------
	//
	// gcgccang
}
