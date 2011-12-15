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

// AGACTAGTTA
// -GAC-AGACG
func ExampleAligner_Align() {
	nwsa := &seq.Seq{Seq: []byte("AGACTAGTTA")}
	nwsb := &seq.Seq{Seq: []byte("GACAGACG")}

	// w(gap) = -5
	nwgap := -5
	//  	 A	 C	 G	 T
	// A	10	-3	-1	-4
	// C	-3	 9	-5	 0
	// G	-1	-5	 7	-3
	// T	-4	 0	-3	 8
	nwm := [][]int{
		{10, -3, -1, -4},
		{-3, 9, -5, 0},
		{-1, -5, 7, -3},
		{-4, 0, -3, 8},
	}

	needle := &Aligner{Gap: nwgap, Matrix: nwm, GapChar: '-'}
	nwa := needle.Align(nwsa, nwsb)
	fmt.Printf("%s\n%s\n", nwa[0].Seq, nwa[1].Seq)
}
