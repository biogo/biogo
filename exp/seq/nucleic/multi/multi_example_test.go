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
	"fmt"
	"github.com/kortschak/biogo/exp/alphabet"
	"github.com/kortschak/biogo/exp/seq"
	"github.com/kortschak/biogo/exp/seq/nucleic"
	"github.com/kortschak/biogo/feat"
	"strings"
)

var (
	m, n    *Multi
	aligned = func(a *Multi) {
		start := a.Start()
		for i := 0; i < a.Count(); i++ {
			s := a.Get(i)
			fmt.Printf("%s%v\n", strings.Repeat(" ", s.Start()-start), s)
		}
		fmt.Println()
		fmt.Println(a)
	}
)

func init() {
	var err error
	m, err = NewMulti("example multi",
		[]nucleic.Sequence{
			nucleic.NewSeq("example DNA 1", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.DNA),
			nucleic.NewSeq("example DNA 2", []alphabet.Letter("ACGGTGACCTGGCGCGCAT"), alphabet.DNA),
			nucleic.NewSeq("example DNA 3", []alphabet.Letter("ACGATGACGTGGCGCTCAT"), alphabet.DNA),
		},
		nucleic.Consensify)

	if err != nil {
		panic(err)
	}
}

func ExampleNewMulti() {
	m, err := NewMulti("example multi",
		[]nucleic.Sequence{
			nucleic.NewSeq("example DNA 1", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.DNA),
			nucleic.NewSeq("example DNA 2", []alphabet.Letter("ACGGTGACCTGGCGCGCAT"), alphabet.DNA),
			nucleic.NewSeq("example DNA 3", []alphabet.Letter("ACGATGACGTGGCGCTCAT"), alphabet.DNA),
		},
		nucleic.Consensify)

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
	fmt.Printf("%v %v\n", m.Count(), m)
	err = m.Add(nucleic.NewQSeq("example DNA",
		[]alphabet.QLetter{{'a', 40}, {'c', 39}, {'g', 40}, {'C', 38}, {'t', 35}, {'g', 20}},
		alphabet.DNA, alphabet.Sanger))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%v %v\n", m.Count(), m)
	err = m.Add(nucleic.NewQSeq("example RNA",
		[]alphabet.QLetter{{'a', 40}, {'c', 39}, {'g', 40}, {'C', 38}, {'t', 35}, {'g', 20}},
		alphabet.RNA, alphabet.Sanger))
	if err != nil {
		fmt.Println(err)
		return
	}
	// Output:
	// 3 acgntgacntggcgcncat
	// 4 acgctgacntggcgcncat
	// Inconsistent alphabets
}

func ExampleMulti_Copy() {
	n = m.Copy().(*Multi)
	n.Set(seq.Position{Pos: 3, Ind: 2}, alphabet.QLetter{L: 't'})
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
	fmt.Println(m.Count())
	// Output:
	// 4
}

func ExampleMulti_IsFlush() {
	m.Get(3).Offset(13)
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

func ExampleMulti_Stitch() {
	f := feat.FeatureSet{
		&feat.Feature{Start: -1, End: 4},
		&feat.Feature{Start: 30, End: 38},
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
