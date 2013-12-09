// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dp

import (
	"code.google.com/p/biogo/align"
	"code.google.com/p/biogo/align/pals/filter"
	"code.google.com/p/biogo/alphabet"
	"code.google.com/p/biogo/seq/linear"
	"code.google.com/p/biogo/util"

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

var (
	maxIGap    = 5
	diffCost   = 3
	sameCost   = 1
	matchCost  = diffCost + sameCost
	blockCost  = diffCost * maxIGap
	rMatchCost = float64(diffCost) + 1
)

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

func (s *S) TestAlignment(c *check.C) {
	l := [...]byte{'A', 'C', 'G', 'T'}
	Q := len(l)
	a := &linear.Seq{Seq: make(alphabet.Letters, 0, util.Pow(Q, k))}
	a.Alpha = alphabet.DNA
	for _, i := range util.DeBruijn(byte(Q), k) {
		a.Seq = append(a.Seq, alphabet.Letter(l[i]))
	}
	b := &linear.Seq{Seq: make(alphabet.Letters, 0, util.Pow(Q, k-1))}
	b.Alpha = alphabet.DNA
	for _, i := range util.DeBruijn(byte(Q), k-1) {
		b.Seq = append(b.Seq, alphabet.Letter(l[i]))
	}
	aligner := NewAligner(a, b, int(k), 50, 0.80)
	aligner.Costs = &Costs{
		MaxIGap:    maxIGap,
		DiffCost:   diffCost,
		SameCost:   sameCost,
		MatchCost:  matchCost,
		BlockCost:  blockCost,
		RMatchCost: rMatchCost,
	}
	hits := aligner.AlignTraps(T)
	c.Check(hits, check.DeepEquals, H)
	la, lb, err := hits.Sum()
	c.Check(la, check.Equals, 791)
	c.Check(lb, check.Equals, 664)
	c.Check(err, check.Equals, nil)
	for _, h := range H {
		sa, sb := &linear.Seq{Seq: a.Seq[h.Abpos:h.Aepos]}, &linear.Seq{Seq: b.Seq[h.Bbpos:h.Bepos]}
		sa.Alpha = alphabet.DNAgapped
		sb.Alpha = alphabet.DNAgapped
		smith := align.SW{
			{0, -1, -1, -1, -1},
			{-1, 2, -1, -1, -1},
			{-1, -1, 2, -1, -1},
			{-1, -1, -1, 2, -1},
			{-1, -1, -1, -1, 2},
		}
		swa, _ := smith.Align(sa, sb)
		fa := align.Format(sa, sb, swa, sa.Alpha.Gap())
		c.Logf("%v\n", swa)
		c.Logf("%s\n%s\n", fa[0], fa[1])
	}
}
