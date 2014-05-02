// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pals

import (
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
	for _, f := range []PairFilter{
		nil,
		func(p *Pair) bool {
			return float64(p.A.Len()) >= float64(p.A.Loc.Len())*epsilon ||
				float64(p.B.Len()) >= float64(p.B.Loc.Len())*epsilon
		},
	} {
		p := NewPiler(0)
		for _, fp := range testPairs {
			fp.A.Pair = fp
			fp.B.Pair = fp
			err := p.Add(fp)
			if err != nil {
				c.Fatal(err)
			}
		}

		for i, pi := range p.Piles(f) {
			c.Logf("%d %v", i, pi)
			for _, f := range pi.Images {
				c.Logf("\t%v", f.Pair)
				c.Check(f.Location(), check.DeepEquals, pi)
			}
		}
	}
}
