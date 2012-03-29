package packed

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
	_ "strings"
)

func ExampleNewQSeq_1() {
	if d, err := NewQSeq("example DNA", nil, alphabet.DNA, alphabet.Sanger); err != nil {
	} else if err = d.Append([]alphabet.QLetter{{'A', 40}, {'C', 39}, {'G', 40}, {'C', 38}, {'T', 35}, {'G', 20}}...); err == nil {
		fmt.Println(d, d.Moltype())
	}
	// Output:
	// acgctg DNA
}

func ExampleQSeq_RevComp() {
	q := []alphabet.Qphred{
		2, 13, 19, 22, 19, 18, 20, 23, 23, 20, 16, 21, 24, 22, 22, 18, 17, 18, 22, 23, 22, 24, 22, 24, 20, 15,
		18, 18, 19, 19, 20, 12, 18, 17, 20, 20, 20, 18, 15, 18, 24, 21, 13, 8, 15, 20, 20, 19, 20, 20, 20, 18,
		16, 16, 16, 10, 15, 18, 18, 18, 11, 2, 11, 20, 19, 18, 18, 16, 10, 12, 22, 0, 0, 0, 0}
	l := []alphabet.Letter("NTTTCTTCTATATCCTTTTCATCTTTTAATCCATTCACCATTTTTTTCCCTCCACCTACCTNTCCTTCTCTTTCT")
	if s, err := NewQSeq("example DNA", nil, alphabet.DNA, alphabet.Sanger); err == nil {
		s.Stringify = func(p seq.Polymer) string {
			s := p.(*QSeq)
			lb, qb, b := []alphabet.Letter{}, []byte{}, []byte{}
			for i, qp := range s.S {
				ql := qp.Unpack(s.Alphabet().(alphabet.Nucleic))
				if ql.Q > 2 {
					ql.L &^= 0x20
				} else {
					ql.L = 'n'
				}
				lb = append(lb, ql.L)
				qb = append(qb, s.QEncode(seq.Position{Pos: i}))
			}
			b = append(b, alphabet.LettersToBytes(lb)...)
			b = append(b, '\n')
			b = append(b, qb...)
			return string(b)
		}

		for i := range l {
			s.Append(alphabet.QLetter{L: l[i], Q: q[i]})
		}
		fmt.Println("Forward:")
		fmt.Println(s)
		s.RevComp()
		fmt.Println("Reverse:")
		fmt.Println(s)
	}
	// Output:
	// Forward:
	// nTTTCTTCTATATCCTTTTCATCTTTTAATCCATTCACCATTTTTTTCCCTCCACCTACCTnTCCTTCTCTnnnn
	// #.47435885169773237879795033445-3255530396.)05545553111+0333,#,54331+-7!!!!
	// Reverse:
	// nnnnAGAGAAGGAnAGGTAGGTGGAGGGAAAAAAATGGTGAATGGATTAAAAGATGAAAAGGATATAGAAGAAAn
	// !!!!7-+13345,#,3330+11135554550).6930355523-54433059797873237796158853474.#
}
