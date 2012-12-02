// Copyright ©2011-2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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

package multi

import (
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/feat"
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/exp/seq/linear"
	"fmt"
	"strings"
)

var (
	m, n    *Multi
	aligned = func(a *Multi) {
		start := a.Start()
		for i := 0; i < a.Rows(); i++ {
			s := a.Row(i)
			fmt.Printf("%s%-s\n", strings.Repeat(" ", s.Start()-start), s)
		}
		fmt.Println()
		fmt.Println(a)
	}
)

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

	aligned(m)
	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGATGACGTGGCGCTCAT
	// 
	// acgntgacntggcgcncat
}

func ExampleMulti_Add() {
	var err error
	fmt.Printf("%v %-s\n", m.Rows(), m)
	err = m.Add(linear.NewQSeq("example DNA",
		[]alphabet.QLetter{{'a', 40}, {'c', 39}, {'g', 40}, {'C', 38}, {'t', 35}, {'g', 20}},
		alphabet.DNA, alphabet.Sanger))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%v %-s\n", m.Rows(), m)
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
	n = m.Copy().(*Multi)
	n.Row(2).Set(3, alphabet.QLetter{L: 't'})
	aligned(m)
	fmt.Println()
	aligned(n)
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
	aligned(m)
	fmt.Printf("\nFlush at left: %v\nFlush at right: %v\n", m.IsFlush(seq.Start), m.IsFlush(seq.End))
	m.Flush(seq.Start, '-')
	fmt.Println()
	aligned(m)
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
	aligned(n)
	n.Join(m, seq.End)
	fmt.Println()
	aligned(n)
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
	aligned(m)
	fmt.Println()
	m.RevComp()
	aligned(m)
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
	aligned(n)
	fmt.Println()
	if err := n.Stitch(f); err == nil {
		aligned(n)
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
	aligned(m)
	m.Truncate(4, 12)
	fmt.Println()
	aligned(m)
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
