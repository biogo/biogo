// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linear

import (
	"fmt"
	"strings"

	"github.com/biogo/biogo/alphabet"
)

func ExampleNewQSeq() {
	d := NewQSeq("example DNA", []alphabet.QLetter{{'A', 40}, {'C', 39}, {'G', 40}, {'C', 38}, {'T', 35}, {'G', 20}}, alphabet.DNA, alphabet.Sanger)
	fmt.Printf("%-s %v\n", d, d.Moltype())
	// Output:
	// ACGCTG DNA
}

func ExampleQSeq_Validate() {
	r := NewQSeq("example RNA", []alphabet.QLetter{{'A', 40}, {'C', 39}, {'G', 40}, {'C', 38}, {'T', 35}, {'G', 20}}, alphabet.RNA, alphabet.Sanger)
	fmt.Printf("%-s %v\n", r, r.Moltype())
	if ok, pos := r.Validate(); ok {
		fmt.Println("valid RNA")
	} else {
		fmt.Println(strings.Repeat(" ", pos-1), "^ first invalid RNA position")
	}
	// Output:
	// ACGCTG RNA
	//     ^ first invalid RNA position
}

func ExampleQSeq_Append() {
	q := []alphabet.Qphred{
		1, 13, 19, 22, 19, 18, 20, 23, 23, 20, 16, 21, 24, 22, 22, 18, 17, 18, 22, 23, 22, 24, 22, 24, 20, 15,
		18, 18, 19, 19, 20, 12, 18, 17, 20, 20, 20, 18, 15, 18, 24, 21, 13, 8, 15, 20, 20, 19, 20, 20, 20, 18,
		16, 16, 16, 10, 15, 18, 18, 18, 11, 1, 11, 20, 19, 18, 18, 16, 10, 12, 22, 0, 0, 0, 0}
	l := []alphabet.Letter("NTTTCTTCTATATCCTTTTCATCTTTTAATCCATTCACCATTTTTTTCCCTCCACCTACCTNTCCTTCTCTTTCT")
	s := NewQSeq("example DNA", nil, alphabet.DNA, alphabet.Sanger)

	for i := range l {
		s.AppendQLetters(alphabet.QLetter{L: l[i], Q: q[i]})
	}
	fmt.Println("Forward:")
	fmt.Printf("%-s\n", s)
	s.RevComp()
	fmt.Println("Reverse:")
	fmt.Printf("%-s\n", s)
	// Output:
	// Forward:
	// nTTTCTTCTATATCCTTTTCATCTTTTAATCCATTCACCATTTTTTTCCCTCCACCTACCTnTCCTTCTCTnnnn
	// Reverse:
	// nnnnAGAGAAGGAnAGGTAGGTGGAGGGAAAAAAATGGTGAATGGATTAAAAGATGAAAAGGATATAGAAGAAAn
}
