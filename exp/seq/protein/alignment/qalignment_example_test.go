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
	"code.google.com/p/biogo/exp/seq/sequtils"
	"fmt"
)

var (
	qm, qn   *QSeq
	qaligned = func(a *QSeq) {
		for i := 0; i < a.Rows(); i++ {
			s := a.Get(i)
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
		seq.DefaultQConsensus)

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
		seq.DefaultQConsensus)
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
	fmt.Printf("%v %v\n", qm.Rows(), qm)
	qm.Add(protein.NewQSeq("example Protein",
		[]alphabet.QLetter{{'a', 40}, {'c', 39}, {'g', 40}, {'C', 38}, {'t', 35}, {'g', 20}},
		alphabet.Protein, alphabet.Sanger))
	fmt.Printf("%v %v\n", qm.Rows(), qm)
	// Output:
	// 3 acgxtgacxtggcgcxcat
	// 4 acgctgacxtggcgcxcat
}

func ExampleQSeq_Copy() {
	qn = qm.Copy().(*QSeq)
	qn.Set(seq.Position{Col: 3, Row: 2}, alphabet.QLetter{L: 't', Q: 40})
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
	fmt.Println(qm.Rows())
	// Output:
	// 4
}

func ExampleQSeq_Join() {
	qaligned(qn)
	err := sequtils.Join(qn, qm, seq.End)
	if err == nil {
		fmt.Println()
		qaligned(qn)
	}
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
	f := fs{
		&fe{s: -1, e: 4},
		&fe{s: 30, e: 38},
	}
	qaligned(qn)
	fmt.Println()
	if err := sequtils.Stitch(qn, qn, f); err == nil {
		qaligned(qn)
	} else {
		fmt.Println(err)
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
	err := sequtils.Truncate(qm, qm, 4, 12)
	if err == nil {
		fmt.Println()
		qaligned(qm)
	}
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
