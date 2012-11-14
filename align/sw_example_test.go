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

package align

import (
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/seq/linear"
	"fmt"
)

func ExampleSW_Align_1() {
	swsa := &linear.Seq{Seq: alphabet.BytesToLetters([]byte("ACACACTA"))}
	swsa.Alpha = alphabet.DNA
	swsb := &linear.Seq{Seq: alphabet.BytesToLetters([]byte("AGCACACA"))}
	swsb.Alpha = alphabet.DNA

	// w(gap) = -1
	// w(match) = +2
	// w(mismatch) = -1
	smith := SW{
		{2, -1, -1, -1, -1},
		{-1, 2, -1, -1, -1},
		{-1, -1, 2, -1, -1},
		{-1, -1, -1, 2, -1},
		{-1, -1, -1, -1, 0},
	}

	aln, err := smith.Align(swsa, swsb)
	if err == nil {
		fmt.Printf("%v\n", aln)
		fa := Format(swsa, swsb, aln, '-')
		fmt.Printf("%s\n%s\n", fa[0], fa[1])
	}
	// Output:
	// [[0,1)/[0,1)=2 -/[1,2)=-1 [1,6)/[2,7)=10 [6,7)/-=-1 [7,8)/[7,8)=2]
	// A-CACACTA
	// AGCACAC-A
}

func ExampleSW_Align_2() {
	swsa := &linear.Seq{Seq: alphabet.BytesToLetters([]byte("AAAATTTAAAA"))}
	swsa.Alpha = alphabet.DNA
	swsb := &linear.Seq{Seq: alphabet.BytesToLetters([]byte("AAAAGGGAAAA"))}
	swsb.Alpha = alphabet.DNA

	// w(gap) = 0
	// w(match) = +2
	// w(mismatch) = -1
	smith := SW{
		{2, -1, -1, -1, 0},
		{-1, 2, -1, -1, 0},
		{-1, -1, 2, -1, 0},
		{-1, -1, -1, 2, 0},
		{0, 0, 0, 0, 0},
	}

	aln, err := smith.Align(swsa, swsb)
	if err == nil {
		fmt.Printf("%v\n", aln)
		fa := Format(swsa, swsb, aln, '-')
		fmt.Printf("%s\n%s\n", fa[0], fa[1])
	}
	// Output:
	// [[0,4)/[0,4)=8 -/[4,7)=0 [4,7)/-=0 [7,11)/[7,11)=8]
	// AAAA---TTTAAAA
	// AAAAGGG---AAAA
}
