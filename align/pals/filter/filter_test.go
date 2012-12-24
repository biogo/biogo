// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package filter

import (
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/seq/linear"
	"code.google.com/p/biogo/index/kmerindex"
	"code.google.com/p/biogo/morass"
	"code.google.com/p/biogo/util"
	check "launchpad.net/gocheck"
	"testing"
)

var k byte = 6

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

func (s *S) TestFilterAndMerge(c *check.C) {
	l := [...]byte{'A', 'C', 'G', 'T'}
	Q := len(l)
	a := &linear.Seq{Seq: make(alphabet.Letters, 0, util.Pow(Q, k))}
	a.Alpha = alphabet.DNA
	for _, i := range util.DeBruijn(byte(Q), k) {
		a.Seq = append(a.Seq, alphabet.Letter(l[i]))
	}
	b := &linear.Seq{Seq: make(alphabet.Letters, 0, util.Pow(Q, k-1))}
	// b.Alpha = alphabet.DNA // Not actually required for this use.
	for _, i := range util.DeBruijn(byte(Q), k-1) {
		b.Seq = append(b.Seq, alphabet.Letter(l[i]))
	}
	i, err := kmerindex.New(int(k), a)
	if err != nil {
		c.Fatalf("Failed to create kmerindex: %v", err)
	}
	i.Build()
	p := &Params{WordSize: int(k), MinMatch: 50, MaxError: 4, TubeOffset: 32}
	f := New(i, p)
	var sorter *morass.Morass
	if sorter, err = morass.New(FilterHit{}, "", "", 2<<20, false); err != nil {
		c.Fatalf("Failed to create morass: %v", err)
	}
	f.Filter(b, false, false, sorter)
	c.Check(sorter.Len(), check.Equals, int64(12))
	r := make([]FilterHit, 1, sorter.Len())
	for {
		err = sorter.Pull(&r[len(r)-1])
		if err != nil {
			r = r[:len(r)-1]
			break
		}
		r = append(r, FilterHit{})
	}
	c.Check(r, check.DeepEquals, []FilterHit{
		{QFrom: 0, QTo: 163, DiagIndex: 32},
		{QFrom: 141, QTo: 247, DiagIndex: 64},
		{QFrom: 237, QTo: 433, DiagIndex: 1120},
		{QFrom: 241, QTo: 347, DiagIndex: 96},
		{QFrom: 341, QTo: 452, DiagIndex: 128},
		{QFrom: 447, QTo: 565, DiagIndex: 1952},
		{QFrom: 542, QTo: 628, DiagIndex: 1984},
		{QFrom: 627, QTo: 814, DiagIndex: 2592},
		{QFrom: 786, QTo: 898, DiagIndex: 2624},
		{QFrom: 868, QTo: 939, DiagIndex: 2880},
		{QFrom: 938, QTo: 997, DiagIndex: 3040},
		{QFrom: 938, QTo: 1024, DiagIndex: 3072},
	})
	m := NewMerger(i, b, p, 5, false)
	for _, h := range r {
		m.MergeFilterHit(&h)
	}
	t := m.FinaliseMerge()
	sorter.CleanUp()
	c.Check(len(t), check.Equals, 6)
	la, lb := t.Sum()
	c.Check(la, check.Equals, 1257)
	c.Check(lb, check.Equals, 402)
	c.Check(t, check.DeepEquals, Trapezoids{
		{Next: nil, Top: 452, Bottom: 0, Left: -128, Right: 3},
		{Next: nil, Top: 433, Bottom: 237, Left: -1120, Right: -1085},
		{Next: nil, Top: 628, Bottom: 447, Left: -1984, Right: -1917},
		{Next: nil, Top: 898, Bottom: 627, Left: -2624, Right: -2557},
		{Next: nil, Top: 939, Bottom: 868, Left: -2880, Right: -2845},
		{Next: nil, Top: 1024, Bottom: 938, Left: -3072, Right: -3005},
	})
}
