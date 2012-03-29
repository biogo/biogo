package sequtils

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
)

func ExampleTruncate_1() {
	s := []byte("ACGCTGACTTGGTGCACGT")
	fmt.Printf("%s\n", s)
	if t, err := Truncate(s, 5, 12, false); err == nil {
		fmt.Printf("%s\n", t)
	}
	// Output:
	// ACGCTGACTTGGTGCACGT
	// GACTTGG
}

func ExampleTruncate_2() {
	s := []byte("ACGCTGACTTGGTGCACGT")

	circular := true
	fmt.Printf("%s Circular = %v\n", s, circular)
	if t, err := Truncate(s, 12, 5, circular); err == nil {
		fmt.Printf("%s\n", t)
	} else {
		fmt.Println("Error:", err)
	}

	circular = false
	fmt.Printf("%s Circular = %v\n", s, circular)
	if t, err := Truncate(s, 12, 5, circular); err == nil {
		fmt.Printf("%s\n", t)
	} else {
		fmt.Println("Error:", err)
	}
	// Output:
	// ACGCTGACTTGGTGCACGT Circular = true
	// TGCACGTACGCT
	// ACGCTGACTTGGTGCACGT Circular = false
	// Error: Start position greater than end position for non-circular sequence.
}

func ExampleReverse() {
	String := func(Q []alphabet.Qphred) string {
		b := make([]byte, 0, len(Q))
		for _, q := range Q {
			b = append(b, q.Encode(alphabet.Sanger))
		}
		return string(b)
	}
	q := []alphabet.Qphred{40, 40, 40, 39, 40, 36, 38, 32, 21, 13, 9, 0, 0, 0}

	fmt.Println(String(q))
	t := Reverse(q).([]alphabet.Qphred)
	fmt.Println(String(t))
	// Output:
	// IIIHIEGA6.*!!!
	// !!!*.6AGEIHIII
}

func ExampleJoin() {
	var (
		s1, s2 []byte
		t      interface{}
		offset int
	)

	s1 = []byte("agctgtgctga")
	s2 = []byte("CGTGCAGTCATGAGTGA")
	fmt.Printf("%s %s\n", s1, s2)
	t, offset = Join(s1, s2, seq.Start)
	fmt.Printf("%s %d\n", t, offset)

	s1 = []byte("agctgtgctga")
	s2 = []byte("CGTGCAGTCATGAGTGA")
	t, offset = Join(s1, s2, seq.End)
	fmt.Printf("%s %d\n", t, offset)
	// Output:
	// agctgtgctga CGTGCAGTCATGAGTGA
	// CGTGCAGTCATGAGTGAagctgtgctga -17
	// agctgtgctgaCGTGCAGTCATGAGTGA 0
}

func ExampleStitch() {
	s := []byte("aAGTATAAgtcagtgcagtgtctggcagTGCTCGTGCgtagtgaagtagGGTTAGTTTa")
	f := feat.FeatureSet{
		&feat.Feature{Start: 1, End: 8},
		&feat.Feature{Start: 28, End: 37},
		&feat.Feature{Start: 49, End: len(s) - 1},
	}
	fmt.Printf("%s\n", s)
	if t, err := Stitch(s, 0, f); err == nil {
		fmt.Printf("%s\n", t)
	}
	// Output:
	// aAGTATAAgtcagtgcagtgtctggcagTGCTCGTGCgtagtgaagtagGGTTAGTTTa
	// AGTATAATGCTCGTGCGGTTAGTTT
}
