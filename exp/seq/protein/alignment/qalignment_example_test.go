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
	"fmt"
	"github.com/kortschak/biogo/exp/alphabet"
	"github.com/kortschak/biogo/exp/seq"
	"github.com/kortschak/biogo/exp/seq/protein"
	"github.com/kortschak/biogo/feat"
)

var (
	qm, qn   *QSeq
	qaligned = func(a *QSeq) {
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
	qm, err = NewQSeq("example alignment",
		[]string{"seq 1", "seq 2", "seq 3"},
		[][]alphabet.QLetter{
			{{'A', 40}, {'A', 40}, {'A', 40}},
			{{'C', 40}, {'C', 40}, {'C', 40}},
			{{'G', 40}, {'G', 40}, {'G', 40}},
			{{'C', 40}, {'G', 40}, {'A', 40}},
			{{'T', 40}, {'T', 40}, {'T', 40}},
			{{'G', 40}, {'G', 40}, {'G', 40}},
			{{'A', 40}, {'A', 40}, {'A', 40}},
			{{'C', 40}, {'C', 40}, {'C', 40}},
			{{'T', 40}, {'C', 40}, {'G', 40}},
			{{'T', 40}, {'T', 40}, {'T', 40}},
			{{'G', 40}, {'G', 40}, {'G', 40}},
			{{'G', 40}, {'G', 40}, {'G', 40}},
			{{'T', 40}, {'C', 40}, {'C', 40}},
			{{'G', 40}, {'G', 40}, {'G', 40}},
			{{'C', 40}, {'C', 40}, {'C', 40}},
			{{'A', 40}, {'G', 40}, {'T', 40}},
			{{'C', 40}, {'C', 40}, {'C', 40}},
			{{'G', 40}, {'A', 40}, {'A', 40}},
			{{'T', 40}, {'T', 40}, {'T', 40}},
		},
		alphabet.Protein,
		alphabet.Sanger,
		protein.QConsensify)

	if err != nil {
		panic(err)
	}
}

func ExampleNewQSeq() {
	qm, err := NewQSeq("example alignment",
		[]string{"seq 1", "seq 2", "seq 3"},
		[][]alphabet.QLetter{
			{{'A', 40}, {'A', 40}, {'A', 40}},
			{{'C', 40}, {'C', 40}, {'C', 40}},
			{{'G', 40}, {'G', 40}, {'G', 40}},
			{{'C', 40}, {'G', 40}, {'A', 40}},
			{{'T', 40}, {'T', 40}, {'T', 40}},
			{{'G', 40}, {'G', 40}, {'G', 40}},
			{{'A', 40}, {'A', 40}, {'A', 40}},
			{{'C', 40}, {'C', 40}, {'C', 40}},
			{{'T', 40}, {'C', 40}, {'G', 40}},
			{{'T', 40}, {'T', 40}, {'T', 40}},
			{{'G', 40}, {'G', 40}, {'G', 40}},
			{{'G', 40}, {'G', 40}, {'G', 40}},
			{{'T', 40}, {'C', 40}, {'C', 40}},
			{{'G', 40}, {'G', 40}, {'G', 40}},
			{{'C', 40}, {'C', 40}, {'C', 40}},
			{{'A', 40}, {'G', 40}, {'T', 40}},
			{{'C', 40}, {'C', 40}, {'C', 40}},
			{{'G', 40}, {'A', 40}, {'A', 40}},
			{{'T', 40}, {'T', 40}, {'T', 40}},
		},
		alphabet.Protein,
		alphabet.Sanger,
		protein.QConsensify)
	if err != nil {
		panic(err)
	}

	qaligned(qm)
	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGATGACGTGGCGCTCAT
	// 
	// acgxtgacxtggcgcxcat
}

func ExampleQSeq_Add() {
	fmt.Printf("%v %v\n", qm.Count(), qm)
	qm.Add(protein.NewQSeq("example DNA",
		[]alphabet.QLetter{{'a', 40}, {'c', 39}, {'g', 40}, {'C', 38}, {'t', 35}, {'g', 20}},
		alphabet.Protein, alphabet.Sanger))
	fmt.Printf("%v %v\n", qm.Count(), qm)
	// Output:
	// 3 acgxtgacxtggcgcxcat
	// 4 acgctgacxtggcgcxcat
}

func ExampleQSeq_Copy() {
	qn = qm.Copy().(*QSeq)
	qn.Set(seq.Position{Pos: 3, Ind: 2}, alphabet.QLetter{L: 't', Q: 40})
	qaligned(qm)
	fmt.Println()
	qaligned(qn)
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

func ExampleQSeq_Count() {
	fmt.Println(qm.Count())
	// Output:
	// 4
}

func ExampleQSeq_Join() {
	qaligned(qn)
	qn.Join(qm, seq.End)
	fmt.Println()
	qaligned(qn)
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

func ExampleQAlignment_Len() {
	fmt.Println(qm.Len())
	// Output:
	// 19
}

func ExampleQSeq_Reverse() {
	qaligned(qm)
	fmt.Println()
	qm.Reverse()
	qaligned(qm)
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

func ExampleQSeq_Stitch() {
	f := feat.FeatureSet{
		&feat.Feature{Start: -1, End: 4},
		&feat.Feature{Start: 30, End: 38},
	}
	qaligned(qn)
	fmt.Println()
	err := qn.Stitch(f)
	if err != nil {
		fmt.Println(err)
	} else {
		qaligned(qn)
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

func ExampleQSeq_Truncate() {
	qaligned(qm)
	qm.Truncate(4, 12)
	fmt.Println()
	qaligned(qm)
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
