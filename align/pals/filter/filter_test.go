// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package filter

import (
	"testing"

	"gopkg.in/check.v1"

	"github.com/biogo/biogo/alphabet"
	"github.com/biogo/biogo/index/kmerindex"
	"github.com/biogo/biogo/morass"
	"github.com/biogo/biogo/seq/linear"
	"github.com/biogo/biogo/util"
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
	if sorter, err = morass.New(Hit{}, "", "", 2<<20, false); err != nil {
		c.Fatalf("Failed to create morass: %v", err)
	}
	f.Filter(b, false, false, sorter)
	c.Check(sorter.Len(), check.Equals, int64(12))
	r := make([]Hit, 1, sorter.Len())
	for {
		err = sorter.Pull(&r[len(r)-1])
		if err != nil {
			r = r[:len(r)-1]
			break
		}
		r = append(r, Hit{})
	}
	c.Check(r, check.DeepEquals, []Hit{
		{From: 0, To: 163, Diagonal: 32},
		{From: 141, To: 247, Diagonal: 64},
		{From: 237, To: 433, Diagonal: 1120},
		{From: 241, To: 347, Diagonal: 96},
		{From: 341, To: 452, Diagonal: 128},
		{From: 447, To: 565, Diagonal: 1952},
		{From: 542, To: 628, Diagonal: 1984},
		{From: 627, To: 814, Diagonal: 2592},
		{From: 786, To: 898, Diagonal: 2624},
		{From: 868, To: 939, Diagonal: 2880},
		{From: 938, To: 997, Diagonal: 3040},
		{From: 938, To: 1024, Diagonal: 3072},
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
