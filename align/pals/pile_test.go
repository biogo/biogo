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

package pals

import (
	"code.google.com/p/biogo/exp/feat"
	check "launchpad.net/gocheck"
)

var (
	testChr   = Contig("testchr")
	testPairs = []*Pair{
		&Pair{
			A:     &Feature{ID: "a", Loc: testChr, From: 2, To: 4},
			B:     &Feature{ID: "g", Loc: testChr, From: 7, To: 9},
			Score: 1,
		},
		&Pair{
			A:     &Feature{ID: "b", Loc: testChr, From: 3, To: 4},
			B:     &Feature{ID: "i", Loc: testChr, From: 7, To: 8},
			Score: 1,
		},
		&Pair{
			A:     &Feature{ID: "c", Loc: testChr, From: 1, To: 3},
			B:     &Feature{ID: "j", Loc: testChr, From: 8, To: 9},
			Score: 1,
		},
		&Pair{
			A:     &Feature{ID: "d", Loc: testChr, From: 1, To: 4},
			B:     &Feature{ID: "f", Loc: testChr, From: 6, To: 9},
			Score: 1,
		},
		&Pair{
			A:     &Feature{ID: "k", Loc: testChr, From: 10, To: 11},
			B:     &Feature{ID: "e", Loc: testChr, From: 4, To: 5},
			Score: 1,
		},
	}
)

func (s *S) TestPiler(c *check.C) {
	epsilon := 0.95
	for _, f := range []PileFilter{
		nil,
		func(a, b feat.Feature, pa, pb *PileInterval) bool {
			lpa := float64(pa.End - pa.Start)
			lpb := float64(pb.End - pb.Start)

			return float64(a.Len()) >= lpa*epsilon || float64(b.Len()) >= lpb*epsilon
		},
	} {
		p := NewPiler(0)
		for _, fp := range testPairs {
			err := p.Add(fp)
			if err != nil {
				c.Fatal(err)
			}
		}

		piles, err := p.Piles(f)
		if err != nil {
			c.Fatal(err)
		}

		for i, pi := range piles {
			c.Logf("%d %v", i, pi)
			for _, fp := range pi.Images {
				c.Logf("\t%v", fp)
				c.Check(fp.A.Location(), check.DeepEquals, pi)
			}
		}
	}
}
