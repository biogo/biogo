package nucleic

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
	d := NewSeq("example DNA", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.DNA)
	fmt.Println(d, d.Moltype())
	// Output:
	// ACGCTGACTTGGTGCACGT DNA
}

func ExampleSeq_Validate() {
	r := NewSeq("example RNA", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.RNA)
	fmt.Println(r, r.Moltype())
	if ok, pos := r.Validate(); ok {
		fmt.Println("valid RNA")
	} else {
		fmt.Println(strings.Repeat(" ", pos-1), "^ first invalid RNA position")
	}
	// Output:
	// ACGCTGACTTGGTGCACGT RNA
	//     ^ first invalid RNA position
}

func ExampleSeq_Truncate_1() {
	s := NewSeq("example DNA", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.DNA)
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

	s = NewSeq("example DNA", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.DNA)
	s.Circular(true)
	fmt.Printf("%s Circular = %v\n", s, s.IsCircular())
	if err := s.Truncate(12, 5); err == nil {
		fmt.Printf("%s Circular = %v\n", s, s.IsCircular())
	} else {
		fmt.Println("Error:", err)
	}

	s = NewSeq("example DNA", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.DNA)
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

func ExampleSeq_RevComp() {
	s := NewSeq("example DNA", []alphabet.Letter("ATGCtGACTTGGTGCACGT"), alphabet.DNA)
	fmt.Println(s)
	s.RevComp()
	fmt.Println(s)
	// Output:
	// ATGCtGACTTGGTGCACGT
	// ACGTGCACCAAGTCaGCAT
}

func ExampleSeq_Join() {
	var s1, s2 *Seq

	s1 = NewSeq("a", []alphabet.Letter("agctgtgctga"), alphabet.DNA)
	s2 = NewSeq("b", []alphabet.Letter("CGTGCAGTCATGAGTGA"), alphabet.DNA)
	fmt.Println(s1, s2)
	if err := s1.Join(s2, seq.Start); err == nil {
		fmt.Println(s1)
	}

	s1 = NewSeq("a", []alphabet.Letter("agctgtgctga"), alphabet.DNA)
	s2 = NewSeq("b", []alphabet.Letter("CGTGCAGTCATGAGTGA"), alphabet.DNA)
	if err := s1.Join(s2, seq.End); err == nil {
		fmt.Println(s1)
	}
	// Output:
	// agctgtgctga CGTGCAGTCATGAGTGA
	// CGTGCAGTCATGAGTGAagctgtgctga
	// agctgtgctgaCGTGCAGTCATGAGTGA
}

func ExampleSeq_Stitch() {
	s := NewSeq("example DNA", []alphabet.Letter("aAGTATAAgtcagtgcagtgtctggcagTGCTCGTGCgtagtgaagtagGGTTAGTTTa"), alphabet.DNA)
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
	s := NewSeq("example DNA", []alphabet.Letter("aAGTATAAgtcagtgcagtgtctggcag<TS>gtagtgaagtagggttagttta"), alphabet.DNA)
	f := feat.FeatureSet{
		&feat.Feature{Start: 0, End: 32},
		&feat.Feature{Start: 1, End: 8, Strand: -1},
		&feat.Feature{Start: 28, End: s.Len() - 1},
	}
	fmt.Println(s)
	if err := s.Compose(f); err == nil {
		fmt.Println(s)
	}
	// Output:
	// aAGTATAAgtcagtgcagtgtctggcag<TS>gtagtgaagtagggttagttta
	// aAGTATAAgtcagtgcagtgtctggcag<TS>TTATACT<TS>gtagtgaagtagggttagttt
}
