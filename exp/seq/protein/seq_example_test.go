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

package protein

import (
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/feat"
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/exp/seq/sequtils"
	"fmt"
	"strings"
)

func ExampleNewSeq() {
	d := NewSeq("example Protein", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.Protein)
	fmt.Println(d, d.Moltype())
	// Output:
	// ACGCTGACTTGGTGCACGT Protein
}

func ExampleSeq_Validate() {
	r := NewSeq("example Protein", []alphabet.Letter("ACGCOGACTTGGTGCACGT"), alphabet.Protein)
	fmt.Println(r, r.Moltype())
	if ok, pos := r.Validate(); ok {
		fmt.Println("valid Protein")
	} else {
		fmt.Println(strings.Repeat(" ", pos-1), "^ first invalid Protein position")
	}
	// Output:
	// ACGCOGACTTGGTGCACGT Protein
	//     ^ first invalid Protein position
}

func ExampleSeq_Truncate_1() {
	s := NewSeq("example Protein", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.Protein)
	fmt.Println(s)
	if err := sequtils.Truncate(s, s, 5, 12); err == nil {
		fmt.Println(s)
	}
	// Output:
	// ACGCTGACTTGGTGCACGT
	// GACTTGG
}

func ExampleSeq_Reverse() {
	s := NewSeq("example Protein", []alphabet.Letter("ATGCtGACTTGGTGCACGT"), alphabet.Protein)
	fmt.Println(s)
	s.Reverse()
	fmt.Println(s)
	// Output:
	// ATGCtGACTTGGTGCACGT
	// TGCACGTGGTTCAGtCGTA
}

func ExampleSeq_Join() {
	var s1, s2 *Seq

	s1 = NewSeq("a", []alphabet.Letter("agctgtgctga"), alphabet.Protein)
	s2 = NewSeq("b", []alphabet.Letter("CGTGCAGTCATGAGTGA"), alphabet.Protein)
	fmt.Println(s1, s2)
	if err := sequtils.Join(s1, s2, seq.Start); err == nil {
		fmt.Println(s1)
	}

	s1 = NewSeq("a", []alphabet.Letter("agctgtgctga"), alphabet.Protein)
	s2 = NewSeq("b", []alphabet.Letter("CGTGCAGTCATGAGTGA"), alphabet.Protein)
	if err := sequtils.Join(s1, s2, seq.End); err == nil {
		fmt.Println(s1)
	}
	// Output:
	// agctgtgctga CGTGCAGTCATGAGTGA
	// CGTGCAGTCATGAGTGAagctgtgctga
	// agctgtgctgaCGTGCAGTCATGAGTGA
}

type fe struct {
	s, e int
	or   feat.Orientation
	feat.Feature
}

func (f fe) Start() int                    { return f.s }
func (f fe) End() int                      { return f.e }
func (f fe) Len() int                      { return f.e - f.s }
func (f fe) Orientation() feat.Orientation { return f.or }

type fs []feat.Feature

func (f fs) Features() []feat.Feature { return []feat.Feature(f) }

func ExampleSeq_Stitch() {
	s := NewSeq("example Protein", []alphabet.Letter("aAGTATAAgtcagtgcagtgtctggcagTGCTCGTGCgtagtgaagtagGGTTAGTTTa"), alphabet.Protein)
	f := fs{
		fe{s: 1, e: 8},
		fe{s: 28, e: 37},
		fe{s: 49, e: s.Len() - 1},
	}
	fmt.Println(s)
	if err := sequtils.Stitch(s, s, f); err == nil {
		fmt.Println(s)
	}
	// Output:
	// aAGTATAAgtcagtgcagtgtctggcagTGCTCGTGCgtagtgaagtagGGTTAGTTTa
	// AGTATAATGCTCGTGCGGTTAGTTT
}

func ExampleSeq_Compose() {
	s := NewSeq("example Protein", []alphabet.Letter("aAGTATAAgtcagtgcagtgtctggcag<TS>gtagtgaagtagggttagttta"), alphabet.Protein)
	f := fs{
		fe{s: 0, e: 32},
		fe{s: 1, e: 8, or: -1},
		fe{s: 28, e: s.Len() - 1},
	}
	fmt.Println(s)
	if err := sequtils.Compose(s, s, f); err == nil {
		fmt.Println(s)
	}
	// Output:
	// aAGTATAAgtcagtgcagtgtctggcag<TS>gtagtgaagtagggttagttta
	// aAGTATAAgtcagtgcagtgtctggcag<TS>AATATGA<TS>gtagtgaagtagggttagttt
}
