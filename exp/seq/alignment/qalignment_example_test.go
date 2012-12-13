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
	"code.google.com/p/biogo/exp/seq/linear"
	"code.google.com/p/biogo/exp/seq/sequtils"
	"fmt"
)

var qm, qn *QSeq

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
		alphabet.DNA,
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
		alphabet.DNA,
		alphabet.Sanger,
		seq.DefaultQConsensus)
	if err == nil {
		fmt.Printf("%-s\n\n%-s\n", qm, qm.Consensus(false))
	}

	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGATGACGTGGCGCTCAT
	// 
	// acgntgacntggcgcncat
}

func ExampleQSeq_Add() {
	fmt.Printf("%v %-s\n", qm.Rows(), qm.Consensus(false))
	qm.Add(linear.NewQSeq("example DNA",
		[]alphabet.QLetter{{'a', 40}, {'c', 39}, {'g', 40}, {'C', 38}, {'t', 35}, {'g', 20}},
		alphabet.DNA, alphabet.Sanger))
	fmt.Printf("%v %-s\n", qm.Rows(), qm.Consensus(false))
	// Output:
	// 3 acgntgacntggcgcncat
	// 4 acgctgacntggcgcncat
}

func ExampleQSeq_Copy() {
	qn = qm.Clone().(*QSeq)
	qn.Row(2).Set(3, alphabet.QLetter{L: 't', Q: 40})
	fmt.Printf("%-s\n\n%-s\n\n", qm, qm.Consensus(false))
	fmt.Printf("%-s\n\n%-s\n", qn, qn.Consensus(false))
	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGATGACGTGGCGCTCAT
	// acgCtg-------------
	// 
	// acgctgacntggcgcncat
	// 
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGtTGACGTGGCGCTCAT
	// acgCtg-------------
	// 
	// acgctgacntggcgcncat
}

func ExampleQSeq_Count() {
	fmt.Println(qm.Rows())
	// Output:
	// 4
}

func ExampleQSeq_Join() {
	fmt.Printf("%-s\n\n%-s\n", qn, qn.Consensus(false))
	err := sequtils.Join(qn, qm, seq.End)
	if err == nil {
		fmt.Printf("\n%-s\n\n%-s\n", qn, qn.Consensus(false))
	}
	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGtTGACGTGGCGCTCAT
	// acgCtg-------------
	// 
	// acgctgacntggcgcncat
	// 
	// ACGCTGACTTGGTGCACGTACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCATACGGTGACCTGGCGCGCAT
	// ACGtTGACGTGGCGCTCATACGATGACGTGGCGCTCAT
	// acgCtg-------------acgCtg-------------
	// 
	// acgctgacntggcgcncatacgctgacntggcgcncat
}

func ExampleQAlignment_Len() {
	fmt.Println(qm.Len())
	// Output:
	// 19
}

func ExampleQSeq_RevComp() {
	fmt.Printf("%-s\n\n%-s\n\n", qm, qm.Consensus(false))
	qm.RevComp()
	fmt.Printf("%-s\n\n%-s\n", qm, qm.Consensus(false))
	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGATGACGTGGCGCTCAT
	// acgCtg-------------
	// 
	// acgctgacntggcgcncat
	// 
	// ACGTGCACCAAGTCAGCGT
	// ATGCGCGCCAGGTCACCGT
	// ATGAGCGCCACGTCATCGT
	// -------------caGcgt
	// 
	// atgngcgccangtcagcgt
}

func ExampleQSeq_Stitch() {
	f := fs{
		&fe{s: -1, e: 4},
		&fe{s: 30, e: 38},
	}
	fmt.Printf("%-s\n\n%-s\n", qn, qn.Consensus(false))
	if err := sequtils.Stitch(qn, qn, f); err == nil {
		fmt.Printf("\n%-s\n\n%-s\n", qn, qn.Consensus(false))
	} else {
		fmt.Println(err)
	}
	// Output:
	// ACGCTGACTTGGTGCACGTACGTGCACCAAGTCAGCGT
	// ACGGTGACCTGGCGCGCATATGCGCGCCAGGTCACCGT
	// ACGtTGACGTGGCGCTCATATGAGCGCCACGTCATCGT
	// acgCtg--------------------------caGcgt
	// 
	// acgctgacntggcgcncatatgngcgccangtcagcgt
	// 
	// ACGCGTCAGCGT
	// ACGGGTCACCGT
	// ACGtGTCATCGT
	// acgC--caGcgt
	// 
	// acgcgtcagcgt
}

func ExampleQSeq_Truncate() {
	fmt.Printf("%-s\n\n%-s\n", qm, qm.Consensus(false))
	err := sequtils.Truncate(qm, qm, 4, 12)
	if err == nil {
		fmt.Printf("\n%-s\n\n%-s\n", qm, qm.Consensus(false))
	}
	// Output:
	// ACGTGCACCAAGTCAGCGT
	// ATGCGCGCCAGGTCACCGT
	// ATGAGCGCCACGTCATCGT
	// -------------caGcgt
	// 
	// atgngcgccangtcagcgt
	// 
	// GCACCAAG
	// GCGCCAGG
	// GCGCCACG
	// --------
	// 
	// gcgccang
}
