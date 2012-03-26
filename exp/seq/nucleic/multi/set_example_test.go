package multi

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
	"github.com/kortschak/BioGo/exp/alphabet"
	"github.com/kortschak/BioGo/exp/seq/nucleic"
)

var set Set

func ExampleSet_AppendEach() {
	ss := [][]alphabet.Letter{
		[]alphabet.Letter("ACGCTGACTTGGTGCACGT"),
		[]alphabet.Letter("ACGACTGGGACGT"),
		[]alphabet.Letter("ACGCTGACTGGCCGT"),
		[]alphabet.Letter("GCCTTTGCACGT"),
	}
	set = make(Set, 4)
	for i := range set {
		set[i] = nucleic.NewSeq(fmt.Sprintf("example DNA %d", i), ss[i], alphabet.DNA)
	}
	as := [][]alphabet.QLetter{
		alphabet.QLetter{L: 'A'}.Repeat(2),
		alphabet.QLetter{L: 'C'}.Repeat(2),
		alphabet.QLetter{L: 'G'}.Repeat(2),
		alphabet.QLetter{L: 'T'}.Repeat(2),
	}

	set.AppendEach(as)

	for _, s := range set {
		fmt.Println(s)
	}
	// Output:
	// ACGCTGACTTGGTGCACGTAA
	// ACGACTGGGACGTCC
	// ACGCTGACTGGCCGTGG
	// GCCTTTGCACGTTT
}

func ExampleSet_Count() {
	fmt.Println(set.Count())
	// Output:
	// 4
}

func ExampleSet_Get() {
	fmt.Println(set.Get(2))
	// Output:
	// ACGCTGACTGGCCGTGG
}

func ExampleSet_Len() {
	fmt.Println(set.Len())
	// Output:
	// 21
}

func ExampleSet_RevComp() {
	if err := set.RevComp(); err == nil {
		for _, s := range set {
			fmt.Println(s)
		}
	}
}

func ExampleSet_Reverse() {
	set.RevComp()
	for _, s := range set {
		fmt.Println(s)
	}
}
