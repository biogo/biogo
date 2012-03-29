package seq

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
	"github.com/kortschak/biogo/bio"
	"github.com/kortschak/biogo/feat"
	"strings"
)

func ExampleSeq_New() {
	d := New("example sequence", []byte("ACGCTGACTTGGTGCACGT"), nil) // Default to bio.DNA
	fmt.Println(d, d.Moltype)
	// Alternative using struct literal
	r := &Seq{ID: "example RNA", Seq: d.Seq, Moltype: bio.RNA}
	fmt.Println(r, r.Moltype)
	if ok, pos := bio.ValidR.Check(r.Seq); ok {
		fmt.Println("valid RNA")
	} else {
		fmt.Println(strings.Repeat(" ", pos-1), "^ first invalid RNA position")
	}
	// Output:
	// ACGCTGACTTGGTGCACGT DNA
	// ACGCTGACTTGGTGCACGT RNA
	//     ^ first invalid RNA position
}

func ExampleSeq_Trunc_1() {
	s := &Seq{Seq: []byte("ACGCTGACTTGGTGCACGT")}
	fmt.Println(s)
	if t, err := s.Trunc(5, 12); err == nil {
		fmt.Println(t)
	}
	// Output:
	// ACGCTGACTTGGTGCACGT
	// GACTTGG
}

func ExampleSeq_Trunc_2() {
	s := &Seq{Seq: []byte("ACGCTGACTTGGTGCACGT"), Circular: true}
	fmt.Println(s, "Circular =", s.Circular)
	if t, err := s.Trunc(12, 5); err == nil {
		fmt.Println(t, "Circular =", t.Circular)
	} else {
		fmt.Println("Error:", err)
	}
	s.Circular = false
	fmt.Println(s, "Circular =", s.Circular)
	if t, err := s.Trunc(12, 5); err == nil {
		fmt.Println(t, "Circular =", t.Circular)
	} else {
		fmt.Println("Error:", err)
	}
	// Output:
	// ACGCTGACTTGGTGCACGT Circular = true
	// TGCACGTACGCT Circular = false
	// ACGCTGACTTGGTGCACGT Circular = false
	// Error: Start position greater than end position for non-circular molecule.
}

func ExampleSeq_RevComp_1() {
	s := &Seq{Seq: []byte("ATGCtGACTTGGTGCACGT")}
	fmt.Println(s)
	if t, err := s.RevComp(); err == nil {
		fmt.Println(t)
	}
	// Output:
	// ATGCtGACTTGGTGCACGT
	// ACGTGCACCAAGTCaGCAT
}

func ExampleSeq_RevComp_2() {
	q := &Quality{Qual: []Qsanger{
		2, 13, 19, 22, 19, 18, 20, 23, 23, 20, 16, 21, 24, 22, 22, 18, 17, 18, 22, 23, 22, 24, 22, 24, 20, 15,
		18, 18, 19, 19, 20, 12, 18, 17, 20, 20, 20, 18, 15, 18, 24, 21, 13, 8, 15, 20, 20, 19, 20, 20, 20, 18,
		16, 16, 16, 10, 15, 18, 18, 18, 11, 2, 11, 20, 19, 18, 18, 16, 10, 12, 22, 0, 0, 0, 0}}
	s := &Seq{Seq: []byte("NTTTCTTCTATATCCTTTTCATCTTTTAATCCATTCACCATTTTTTTCCCTCCACCTACCTNTCCTTCTCTTTCT"), Quality: q}
	fmt.Println("Forward:")
	fmt.Println(s)
	fmt.Println(s.Quality)
	if t, err := s.RevComp(); err == nil {
		fmt.Println("Reverse:")
		fmt.Println(t)
		fmt.Println(t.Quality)
	}
	// Output:
	// Forward:
	// NTTTCTTCTATATCCTTTTCATCTTTTAATCCATTCACCATTTTTTTCCCTCCACCTACCTNTCCTTCTCTTTCT
	// #.47435885169773237879795033445-3255530396.)05545553111+0333,#,54331+-7!!!!
	// Reverse:
	// AGAAAGAGAAGGANAGGTAGGTGGAGGGAAAAAAATGGTGAATGGATTAAAAGATGAAAAGGATATAGAAGAAAN
	// !!!!7-+13345,#,3330+11135554550).6930355523-54433059797873237796158853474.#
}

func ExampleQuality_Reverse() {
	q := &Quality{Qual: []Qsanger{40, 40, 40, 39, 40, 36, 38, 32, 21, 13, 9, 0, 0, 0}}
	fmt.Println(q)
	t := q.Reverse()
	fmt.Println(t)
	// Output:
	// IIIHIEGA6.*!!!
	// !!!*.6AGEIHIII
}

func ExampleJoin() {
	var s1, s2 *Seq
	s1 = &Seq{Seq: []byte("agctgtgctga")}
	s2 = &Seq{Seq: []byte("CGTGCAGTCATGAGTGA")}
	fmt.Println(s1, s2)
	if t, err := s1.Join(s2, Append); err == nil {
		fmt.Println(t)
	}
	s1 = &Seq{Seq: []byte("agctgtgctga")}
	s2 = &Seq{Seq: []byte("CGTGCAGTCATGAGTGA")}
	if t, err := s1.Join(s2, Prepend); err == nil {
		fmt.Println(t)
	}
	// Output:
	// agctgtgctga CGTGCAGTCATGAGTGA
	// agctgtgctgaCGTGCAGTCATGAGTGA
	// CGTGCAGTCATGAGTGAagctgtgctga
}

func ExampleStitch() {
	s := &Seq{Seq: []byte("aAGTATAAgtcagtgcagtgtctggcagTGCTCGTGCgtagtgaagtagGGTTAGTTTa")}
	f := feat.FeatureSet{
		&feat.Feature{Start: 1, End: 8},
		&feat.Feature{Start: 28, End: 37},
		&feat.Feature{Start: 49, End: s.Len() - 1},
	}
	fmt.Println(s)
	if t, err := s.Stitch(f); err == nil {
		fmt.Println(t)
	}
	// Output:
	// aAGTATAAgtcagtgcagtgtctggcagTGCTCGTGCgtagtgaagtagGGTTAGTTTa
	// AGTATAATGCTCGTGCGGTTAGTTT
}
