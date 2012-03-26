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
	"github.com/kortschak/BioGo/bio"
	"github.com/kortschak/BioGo/exp/alphabet"
	"github.com/kortschak/BioGo/exp/seq"
	"github.com/kortschak/BioGo/feat"
)

func ExampleNewSeq_1() {
	if d, err := NewSeq("example DNA", []alphabet.Letter("ACGCTGACTTGGTGCACGTA"), alphabet.DNA); err == nil {
		fmt.Println(d, d.Moltype())
	}
	// Output:
	// acgctgacttggtgcacgta DNA
}

func ExampleNewSeq_2() {
	if _, err := NewSeq("example RNA", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.RNA); err != nil {
		fmt.Printf("%v: %v\n", err, alphabet.Letters(err.(bio.Error).Items()[0].([]alphabet.Letter)))
	}
	// Output:
	// Encoding error: packed: invalid letter 'T' at position 4.: ACGCTGACTTGGTGCACGT
}

func ExampleSeq_Truncate_1() {
	if s, err := NewSeq("example DNA", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.DNA); err == nil {
		fmt.Println(s)
		if err := s.Truncate(5, 12); err == nil {
			fmt.Println(s)
		}
	}
	// Output:
	// acgctgacttggtgcacgt
	// gacttgg
}

func ExampleSeq_Truncate_2() {
	var (
		s   *Seq
		err error
	)

	if s, err = NewSeq("example DNA", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.DNA); err == nil {
		s.Circular(true)
		fmt.Printf("%s Circular = %v\n", s, s.IsCircular())
		if err := s.Truncate(12, 5); err == nil {
			fmt.Printf("%s Circular = %v\n", s, s.IsCircular())
		} else {
			fmt.Println("Error:", err)
		}
	}

	if s, err = NewSeq("example DNA", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.DNA); err == nil {
		fmt.Printf("%s Circular = %v\n", s, s.IsCircular())
		if err := s.Truncate(12, 5); err == nil {
			fmt.Printf("%s Circular = %v\n", s, s.IsCircular())
		} else {
			fmt.Println("Error:", err)
		}
	}
	// Output:
	// acgctgacttggtgcacgt Circular = true
	// tgcacgtacgct Circular = false
	// acgctgacttggtgcacgt Circular = false
	// Error: Start position greater than end position for non-circular sequence.
}

func ExampleSeq_RevComp_1() {
	if s, err := NewSeq("example DNA", []alphabet.Letter("ATGCTGACTTGGTGCACGT"), alphabet.DNA); err == nil {
		fmt.Println(s)
		s.RevComp()
		fmt.Println(s)
	}
	// Output:
	// atgctgacttggtgcacgt
	// acgtgcaccaagtcagcat
}

func ExampleSeq_Join() {
	var (
		s1, s2 *Seq
		err    error
	)

	if s1, err = NewSeq("a", []alphabet.Letter("agctgtgctga"), alphabet.DNA); err != nil {
		return
	}
	if s2, err = NewSeq("b", []alphabet.Letter("CGTGCAGTCATGAGTGA"), alphabet.DNA); err != nil {
		return
	}
	fmt.Println(s1, s2)
	if err = s1.Join(s2, seq.Start); err == nil {
		fmt.Println(s1)
	}

	if s1, err = NewSeq("a", []alphabet.Letter("agctgtgctga"), alphabet.DNA); err != nil {
		return
	}
	if s2, err = NewSeq("b", []alphabet.Letter("CGTGCAGTCATGAGTGA"), alphabet.DNA); err != nil {
		return
	}
	if err = s1.Join(s2, seq.End); err == nil {
		fmt.Println(s1)
	}
	// Output:
	// agctgtgctga cgtgcagtcatgagtga
	// cgtgcagtcatgagtgaagctgtgctga
	// agctgtgctgacgtgcagtcatgagtga
}

func ExampleSeq_Stitch() {
	if s, err := NewSeq("example DNA", []alphabet.Letter("aAGTATAAgtcagtgcagtgtctggcagTGCTCGTGCgtagtgaagtagGGTTAGTTTa"), alphabet.DNA); err == nil {
		f := feat.FeatureSet{
			&feat.Feature{Start: 1, End: 8},
			&feat.Feature{Start: 28, End: 37},
			&feat.Feature{Start: 49, End: s.Len() - 1},
		}
		fmt.Println(s)
		if err := s.Stitch(f); err == nil {
			fmt.Println(s)
		}
	}
	// Output:
	// aagtataagtcagtgcagtgtctggcagtgctcgtgcgtagtgaagtagggttagttta
	// agtataatgctcgtgcggttagttt
}

func ExampleSeq_Compose() {
	if s, err := NewSeq("example DNA", []alphabet.Letter("aAGTATAAgtcagtgcagtgtctggcagTAgtagtgaagtagggttagttta"), alphabet.DNA); err == nil {
		f := feat.FeatureSet{
			&feat.Feature{Start: 0, End: 30},
			&feat.Feature{Start: 1, End: 8},
			&feat.Feature{Start: 28, End: 30},
			&feat.Feature{Start: 30, End: s.Len() - 1},
		}
		fmt.Println(s)
		if err := s.Compose(f); err == nil {
			fmt.Println(s)
		}
	}
	// Output:
	// aagtataagtcagtgcagtgtctggcagtagtagtgaagtagggttagttta
	// aagtataagtcagtgcagtgtctggcagtaagtataatagtagtgaagtagggttagttt
}
