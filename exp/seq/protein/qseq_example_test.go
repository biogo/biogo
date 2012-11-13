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
	"fmt"
	"strings"
)

func ExampleNewQSeq() {
	d := NewQSeq("example Protein", []alphabet.QLetter{{'A', 40}, {'C', 39}, {'G', 40}, {'C', 38}, {'T', 35}, {'G', 20}}, alphabet.Protein, alphabet.Sanger)
	fmt.Println(d, d.Moltype())
	// Output:
	// ACGCTG Protein
}

func ExampleQSeq_Validate() {
	r := NewQSeq("example Protein", []alphabet.QLetter{{'A', 40}, {'C', 39}, {'G', 40}, {'C', 38}, {'O', 35}, {'G', 20}}, alphabet.Protein, alphabet.Sanger)
	fmt.Println(r, r.Moltype())
	if ok, pos := r.Validate(); ok {
		fmt.Println("valid Protein")
	} else {
		fmt.Println(strings.Repeat(" ", pos-1), "^ first invalid Protein position")
	}
	// Output:
	// ACGCOG Protein
	//     ^ first invalid Protein position
}

func ExampleQSeq_Append() {
	q := []alphabet.Qphred{
		2, 13, 19, 22, 19, 18, 20, 23, 23, 20, 16, 21, 24, 22, 22, 18, 17, 18, 22, 23, 22, 24, 22, 24, 20, 15,
		18, 18, 19, 19, 20, 12, 18, 17, 20, 20, 20, 18, 15, 18, 24, 21, 13, 8, 15, 20, 20, 19, 20, 20, 20, 18,
		16, 16, 16, 10, 15, 18, 18, 18, 11, 2, 11, 20, 19, 18, 18, 16, 10, 12, 22, 0, 0, 0, 0}
	l := []alphabet.Letter("NTTTCTTCTATATCCTTTTCATCTTTTAATCCATTCACCATTTTTTTCCCTCCACCTACCTNTCCTTCTCTTTCT")
	s := NewQSeq("example Protein", nil, alphabet.Protein, alphabet.Sanger)

	for i := range l {
		s.AppendQLetters(alphabet.QLetter{L: l[i], Q: q[i]})
	}
	fmt.Println("Forward:")
	fmt.Println(s)
	s.Reverse()
	fmt.Println("Reverse:")
	fmt.Println(s)
	// Output:
	// Forward:
	// xTTTCTTCTATATCCTTTTCATCTTTTAATCCATTCACCATTTTTTTCCCTCCACCTACCTxTCCTTCTCTxxxx
	// Reverse:
	// xxxxTCTCTTCCTxTCCATCCACCTCCCTTTTTTTACCACTTACCTAATTTTCTACTTTTCCTATATCTTCTTTx
}
