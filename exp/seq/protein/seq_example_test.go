package protein

// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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

import (
	"fmt"
	"github.com/kortschak/biogo/exp/alphabet"
	"github.com/kortschak/biogo/exp/seq"
	"github.com/kortschak/biogo/feat"
	"strings"
)

func ExampleNewSeq() {
	d := NewSeq("example protein", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.Protein)
	fmt.Println(d, d.Moltype())
	// Output:
	// ACGCTGACTTGGTGCACGT Protein
}

func ExampleSeq_Validate() {
	r := NewSeq("example protein", []alphabet.Letter("ACGCUGACTTGGTGCACGT"), alphabet.Protein)
	fmt.Println(r, r.Moltype())
	if ok, pos := r.Validate(); ok {
		fmt.Println("valid protein")
	} else {
		fmt.Println(strings.Repeat(" ", pos-1), "^ first invalid protein position")
	}
	// Output:
	// ACGCUGACTTGGTGCACGT Protein
	//     ^ first invalid protein position
}

func ExampleSeq_Truncate_1() {
	s := NewSeq("example protein", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.Protein)
	fmt.Println(s)
	if err := s.Truncate(5, 12); err == nil {
		fmt.Println(s)
	}
	// Output:
	// ACGCTGACTTGGTGCACGT
	// GACTTGG
}

func ExampleSeq_Truncate_2() {
	var s *Seq

	s = NewSeq("example Protein", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.Protein)
	s.Circular(true)
	fmt.Printf("%s Circular = %v\n", s, s.IsCircular())
	if err := s.Truncate(12, 5); err == nil {
		fmt.Printf("%s Circular = %v\n", s, s.IsCircular())
	} else {
		fmt.Println("Error:", err)
	}

	s = NewSeq("example Protein", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.Protein)
	fmt.Printf("%s Circular = %v\n", s, s.IsCircular())
	if err := s.Truncate(12, 5); err == nil {
		fmt.Printf("%s Circular = %v\n", s, s.IsCircular())
	} else {
		fmt.Println("Error:", err)
	}
	// Output:
	// ACGCTGACTTGGTGCACGT Circular = true
	// TGCACGTACGCT Circular = false
	// ACGCTGACTTGGTGCACGT Circular = false
	// Error: Start position greater than end position for non-circular sequence.
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
	if err := s1.Join(s2, seq.Start); err == nil {
		fmt.Println(s1)
	}

	s1 = NewSeq("a", []alphabet.Letter("agctgtgctga"), alphabet.Protein)
	s2 = NewSeq("b", []alphabet.Letter("CGTGCAGTCATGAGTGA"), alphabet.Protein)
	if err := s1.Join(s2, seq.End); err == nil {
		fmt.Println(s1)
	}
	// Output:
	// agctgtgctga CGTGCAGTCATGAGTGA
	// CGTGCAGTCATGAGTGAagctgtgctga
	// agctgtgctgaCGTGCAGTCATGAGTGA
}

func ExampleSeq_Stitch() {
	s := NewSeq("example Protein", []alphabet.Letter("aAGTATAAgtcagtgcagtgtctggcagTGCTCGTGCgtagtgaagtagGGTTAGTTTa"), alphabet.Protein)
	f := feat.FeatureSet{
		&feat.Feature{Start: 1, End: 8},
		&feat.Feature{Start: 28, End: 37},
		&feat.Feature{Start: 49, End: s.Len() - 1},
	}
	fmt.Println(s)
	if err := s.Stitch(f); err == nil {
		fmt.Println(s)
	}
	// Output:
	// aAGTATAAgtcagtgcagtgtctggcagTGCTCGTGCgtagtgaagtagGGTTAGTTTa
	// AGTATAATGCTCGTGCGGTTAGTTT
}

func ExampleSeq_Compose() {
	s := NewSeq("example Protein", []alphabet.Letter("aAGTATAAgtcagtgcagtgtctggcag<TS>gtagtgaagtagggttagttta"), alphabet.Protein)
	f := feat.FeatureSet{
		&feat.Feature{Start: 0, End: 32},
		&feat.Feature{Start: 1, End: 8},
		&feat.Feature{Start: 28, End: s.Len() - 1},
	}
	fmt.Println(s)
	if err := s.Compose(f); err == nil {
		fmt.Println(s)
	}
	// Output:
	// aAGTATAAgtcagtgcagtgtctggcag<TS>gtagtgaagtagggttagttta
	// aAGTATAAgtcagtgcagtgtctggcag<TS>AGTATAA<TS>gtagtgaagtagggttagttt
}
