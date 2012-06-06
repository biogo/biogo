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

package alignment

import (
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/exp/seq/protein"
	"code.google.com/p/biogo/feat"
	"fmt"
)

var (
	m, n    *Seq
	aligned = func(a *Seq) {
		for i := 0; i < a.Count(); i++ {
			s := a.Extract(i)
			fmt.Printf("%v\n", s)
		}
		fmt.Println()
		fmt.Println(a)
	}
)

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
		alphabet.Protein,
		protein.Consensify)

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
		alphabet.Protein,
		protein.Consensify)
	if err != nil {
		panic(err)
	}

	aligned(m)
	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGATGACGTGGCGCTCAT
	// 
	// acgxtgacxtggcgcxcat
}

func ExampleSeq_Add() {
	fmt.Printf("%v %v\n", m.Count(), m)
	m.Add(protein.NewQSeq("example Protein",
		[]alphabet.QLetter{{'a', 40}, {'c', 39}, {'g', 40}, {'C', 38}, {'t', 35}, {'g', 20}},
		alphabet.Protein, alphabet.Sanger))
	fmt.Printf("%v %v\n", m.Count(), m)
	// Output:
	// 3 acgxtgacxtggcgcxcat
	// 4 acgctgacxtggcgcxcat
}

func ExampleSeq_Copy() {
	n = m.Copy().(*Seq)
	n.Set(seq.Position{Pos: 3, Ind: 2}, alphabet.QLetter{L: 't'})
	aligned(m)
	fmt.Println()
	aligned(n)
	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGATGACGTGGCGCTCAT
	// acgCtg-------------
	// 
	// acgctgacxtggcgcxcat
	// 
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGtTGACGTGGCGCTCAT
	// acgCtg-------------
	// 
	// acgctgacxtggcgcxcat
}

func ExampleSeq_Count() {
	fmt.Println(m.Count())
	// Output:
	// 4
}

func ExampleSeq_Join() {
	aligned(n)
	n.Join(m, seq.End)
	fmt.Println()
	aligned(n)
	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGtTGACGTGGCGCTCAT
	// acgCtg-------------
	// 
	// acgctgacxtggcgcxcat
	// 
	// ACGCTGACTTGGTGCACGTACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCATACGGTGACCTGGCGCGCAT
	// ACGtTGACGTGGCGCTCATACGATGACGTGGCGCTCAT
	// acgCtg-------------acgCtg-------------
	// 
	// acgctgacxtggcgcxcatacgctgacxtggcgcxcat
}

func ExampleAlignment_Len() {
	fmt.Println(m.Len())
	// Output:
	// 19
}

func ExampleSeq_Reverse() {
	aligned(m)
	fmt.Println()
	m.Reverse()
	aligned(m)
	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGATGACGTGGCGCTCAT
	// acgCtg-------------
	// 
	// acgctgacxtggcgcxcat
	// 
	// TGCACGTGGTTCAGTCGCA
	// TACGCGCGGTCCAGTGGCA
	// TACTCGCGGTGCAGTAGCA
	// -------------gtCgca
	// 
	// tacxcgcggtxcagtcgca
}

func ExampleSeq_Stitch() {
	f := feat.FeatureSet{
		&feat.Feature{Start: -1, End: 4},
		&feat.Feature{Start: 30, End: 38},
	}
	aligned(n)
	fmt.Println()
	err := n.Stitch(f)
	if err != nil {
		fmt.Println(err)
	} else {
		aligned(n)
	}
	// Output:
	// ACGCTGACTTGGTGCACGTACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCATACGGTGACCTGGCGCGCAT
	// ACGtTGACGTGGCGCTCATACGATGACGTGGCGCTCAT
	// acgCtg-------------acgCtg-------------
	// 
	// acgctgacxtggcgcxcatacgctgacxtggcgcxcat
	// 
	// ACGCGTGCACGT
	// ACGGGCGCGCAT
	// ACGtGCGCTCAT
	// acgC--------
	// 
	// acgcgcgcxcat
}

func ExampleSeq_Truncate() {
	aligned(m)
	m.Truncate(4, 12)
	fmt.Println()
	aligned(m)
	// Output:
	// TGCACGTGGTTCAGTCGCA
	// TACGCGCGGTCCAGTGGCA
	// TACTCGCGGTGCAGTAGCA
	// -------------gtCgca
	// 
	// tacxcgcggtxcagtcgca
	// 
	// CGTGGTTC
	// CGCGGTCC
	// CGCGGTGC
	// --------
	// 
	// cgcggtxc
}
