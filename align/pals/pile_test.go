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
	"code.google.com/p/biogo/feat"
	check "launchpad.net/gocheck"
)

var (
	testPairs = []*FeaturePair{
		&FeaturePair{
			A:     &feat.Feature{ID: "a", Location: "testchr", Start: 2, End: 4},
			B:     &feat.Feature{ID: "g", Location: "testchr", Start: 7, End: 9},
			Score: 1,
		},
		&FeaturePair{
			A:     &feat.Feature{ID: "b", Location: "testchr", Start: 3, End: 4},
			B:     &feat.Feature{ID: "i", Location: "testchr", Start: 7, End: 8},
			Score: 1,
		},
		&FeaturePair{
			A:     &feat.Feature{ID: "c", Location: "testchr", Start: 1, End: 3},
			B:     &feat.Feature{ID: "j", Location: "testchr", Start: 8, End: 9},
			Score: 1,
		},
		&FeaturePair{
			A:     &feat.Feature{ID: "d", Location: "testchr", Start: 1, End: 4},
			B:     &feat.Feature{ID: "f", Location: "testchr", Start: 6, End: 9},
			Score: 1,
		},
		&FeaturePair{
			B:     &feat.Feature{ID: "e", Location: "testchr", Start: 4, End: 5},
			A:     &feat.Feature{ID: "k", Location: "testchr", Start: 10, End: 11},
			Score: 1,
		},
	}
)

func (s *S) TestPiler(c *check.C) {
	epsilon := 0.95
	for _, f := range []PileFilter{
		nil,
		func(a, b *feat.Feature, pa, pb *PileInterval) bool {
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
				c.Check(fp.A.Meta, check.DeepEquals, pi)
			}
		}
	}
}
