package dp

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
	"github.com/kortschak/BioGo/align/pals/filter"
	"github.com/kortschak/BioGo/align/sw"
	"github.com/kortschak/BioGo/seq"
	"github.com/kortschak/BioGo/util"
	check "launchpad.net/gocheck"
	"testing"
)

var (
	k byte              = 6
	T filter.Trapezoids = filter.Trapezoids{
		{Next: nil, Top: 452, Bottom: 0, Left: -128, Right: 3},
		{Next: nil, Top: 433, Bottom: 237, Left: -1120, Right: -1085},
		{Next: nil, Top: 628, Bottom: 447, Left: -1984, Right: -1917},
		{Next: nil, Top: 898, Bottom: 627, Left: -2624, Right: -2557},
		{Next: nil, Top: 939, Bottom: 868, Left: -2880, Right: -2845},
		{Next: nil, Top: 1024, Bottom: 938, Left: -3072, Right: -3005},
	}
	H DPHits = DPHits{
		{Abpos: 1, Bbpos: 0, Aepos: 290, Bepos: 242, LowDiagonal: -3, HighDiagonal: 54, Score: 101, Error: 0.19421487603305784},
		{Abpos: 365, Bbpos: 286, Aepos: 435, Bepos: 345, LowDiagonal: 74, HighDiagonal: 96, Score: 26, Error: 0.1864406779661017},
		{Abpos: 437, Bbpos: 341, Aepos: 507, Bepos: 400, LowDiagonal: 91, HighDiagonal: 113, Score: 26, Error: 0.1864406779661017},
		{Abpos: 3201, Bbpos: 642, Aepos: 3477, Bepos: 873, LowDiagonal: 2553, HighDiagonal: 2610, Score: 96, Error: 0.19480519480519481},
		{Abpos: 3980, Bbpos: 948, Aepos: 4066, Bepos: 1021, LowDiagonal: 3026, HighDiagonal: 3054, Score: 30, Error: 0.1917808219178082},
	}
)

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

func (s *S) TestAlignment(c *check.C) {
	l := [...]byte{'A', 'C', 'G', 'T'}
	Q := len(l)
	a := &seq.Seq{Seq: make([]byte, 0, util.Pow(Q, k))}
	for _, i := range util.DeBruijn(byte(Q), k) {
		a.Seq = append(a.Seq, l[i])
	}
	b := &seq.Seq{Seq: make([]byte, 0, util.Pow(Q, k-1))}
	for _, i := range util.DeBruijn(byte(Q), k-1) {
		b.Seq = append(b.Seq, l[i])
	}
	aligner := NewAligner(a, b, int(k), 50, 0.80, 4)
	hits := aligner.AlignTraps(T)
	c.Check(hits, check.DeepEquals, H)
	la, lb, err := hits.Sum()
	c.Check(la, check.Equals, 791)
	c.Check(lb, check.Equals, 664)
	c.Check(err, check.Equals, nil)
	for _, h := range H {
		sa, sb := &seq.Seq{Seq: a.Seq[h.Abpos:h.Aepos]}, &seq.Seq{Seq: b.Seq[h.Bbpos:h.Bepos]}
		swm := [][]int{
			{2, -1, -1, -1, -1},
			{-1, 2, -1, -1, -1},
			{-1, -1, 2, -1, -1},
			{-1, -1, -1, 2, -1},
			{-1, -1, -1, -1, 0},
		}

		smith := &sw.Aligner{Matrix: swm, LookUp: sw.LookUpN, GapChar: '-'}
		swa, _ := smith.Align(sa, sb)
		c.Logf("a: %s\nb: %s\n", swa[0], swa[1])
	}
}
