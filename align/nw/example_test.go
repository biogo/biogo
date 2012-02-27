package nw

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
	"github.com/kortschak/BioGo/seq"
)

func ExampleAligner_Align() {
	nwsa := &seq.Seq{Seq: []byte("AGACTAGTTA")}
	nwsb := &seq.Seq{Seq: []byte("GACAGACG")}

	//  	 A	 C	 G	 T	 -
	// A	10	-3	-1	-4	-5
	// C	-3	 9	-5	 0	-5
	// G	-1	-5	 7	-3	-5
	// T	-4	 0	-3	 8	-5
	// -	-5	-5	-5	-5	 0
	nwm := [][]int{
		{10, -3, -1, -4, -5},
		{-3, 9, -5, 0, -5},
		{-1, -5, 7, -3, -5},
		{-4, 0, -3, 8, -5},
		{-4, -4, -4, -4, 0},
	}

	needle := &Aligner{Matrix: nwm, LookUp: LookUpN, GapChar: '-'}
	if nwa, err := needle.Align(nwsa, nwsb); err == nil {
		fmt.Printf("%s\n%s\n", nwa[0].Seq, nwa[1].Seq)
	}
	// Output:
	// AGACTAGTTA
	// -GAC-AGACG
}
