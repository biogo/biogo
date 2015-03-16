// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pals

import (
	"github.com/biogo/biogo/align/pals/dp"
	"github.com/biogo/biogo/align/pals/filter"
	"github.com/biogo/biogo/alphabet"
	"github.com/biogo/biogo/seq"
	"github.com/biogo/biogo/seq/linear"
	"github.com/biogo/biogo/util"

	"bytes"
	"fmt"
	"math"
	"testing"

	"gopkg.in/check.v1"
)

const (
	Q = 4
)

var (
	maxk byte    = 8
	l    [Q]byte = [Q]byte{'A', 'C', 'G', 'T'}
	ps   *Packed
)

// Helpers
type B struct {
	*bytes.Buffer
}

func (b *B) Close() error { return nil }

// Checkers
type floatApproxChecker struct {
	*check.CheckerInfo
}

var floatApprox check.Checker = &floatApproxChecker{
	&check.CheckerInfo{Name: "Approximately", Params: []string{"obtained", "expected", "epsilon"}},
}

func (checker *floatApproxChecker) Check(params []interface{}, names []string) (result bool, error string) {
	return math.Abs(params[0].(float64)-params[1].(float64))/params[0].(float64) < params[2].(float64), ""
}

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type ft struct {
	start, end int
	result     *Feature
}

//	1:            deBruijn1          4            0-0
//	2:            deBruijn2         16            1-1
//	3:            deBruijn3         64            2-2
//	4:            deBruijn4        256            3-3
//	5:            deBruijn5       1024            4-5
//	6:            deBruijn6       4096            6-10
//	7:            deBruijn7      16384           11-27
//	8:            deBruijn8      65536           28-92
var T []ft = []ft{
	{1020, 1030, &Feature{"deBruijn2", 0, 6, Contig("deBruijn2"), nil}},
	{1025, 1030, &Feature{"deBruijn2", 1, 6, Contig("deBruijn2"), nil}},
	{1010, 1060, &Feature{"deBruijn2", 0, 16, Contig("deBruijn2"), nil}},
	{0, 1060, &Feature{"deBruijn1", 0, 4, Contig("deBruijn1"), nil}},
	{4 * binSize, 4*binSize + 904, &Feature{"deBruijn5", 0, 904, Contig("deBruijn5"), nil}},
	{29 * binSize, 32*binSize - 1, &Feature{"deBruijn8", 1024, 4095, Contig("deBruijn8"), nil}},
}

type pt struct {
	l          int
	id         float64
	k, s, d, t int
	list       float64
}

var P []pt = []pt{
	{l: 50, id: 0.1, k: 6, s: 6, d: 0, t: 32, list: 7.3},
	{l: 60, id: 0.1, k: 7, s: 7, d: 0, t: 32, list: 1.8},
	{l: 70, id: 0.1, k: 8, s: 8, d: 0, t: 32, list: 0.46},
	{l: 80, id: 0.1, k: 10, s: 10, d: 0, t: 32, list: 0.029},
	{l: 90, id: 0.1, k: 11, s: 11, d: 0, t: 32, list: 0.0071},
	{l: 100, id: 0.1, k: 6, s: 12, d: 1, t: 33, list: 7.3},
	{l: 200, id: 0.5, k: 6, s: 25, d: 3, t: 35, list: 7.3},
	{l: 200, id: 0.9, k: 10, s: 200, d: 19, t: 51, list: 0.029},
	{l: 400, id: 0.8, k: 6, s: 50, d: 7, t: 39, list: 7.3},
	{l: 400, id: 0.9, k: 10, s: 400, d: 39, t: 71, list: 0.029},
	{l: 400, id: 0.99, k: 15, s: 400, d: 4, t: 36, list: 2.8e-05},
}

type S struct{}

var _ = check.Suite(&S{})

func (s *S) SetUpSuite(c *check.C) {
	p := NewPacker("")
	for k := byte(1); k <= maxk; k++ {
		a := &linear.Seq{
			Annotation: seq.Annotation{
				ID: fmt.Sprintf("deBruijn%d", k),
			},
			Seq: make(alphabet.Letters, 0, util.Pow(Q, k)),
		}
		for _, i := range util.DeBruijn(byte(Q), k) {
			a.Seq = append(a.Seq, alphabet.Letter(l[i]))
		}
		a.Loc = Contig(a.ID)
		p.Pack(a)
	}
	ps = p.FinalisePack()
}

func (s *S) TestOptimise(c *check.C) {
	// minHitLen int, minId float64, target, query *seq.Seq, tubeOffset int, maxMem uint64
	t := &linear.Seq{Seq: make(alphabet.Letters, 29940)}
	for _, p := range P {
		pa := New(t, t, true, nil, 1, 0, nil, nil)
		err := pa.Optimise(p.l, p.id)
		if err == nil {
			c.Check(*pa.FilterParams, check.Equals, filter.Params{WordSize: p.k, MinMatch: p.s, MaxError: p.d, TubeOffset: p.t})
			c.Check(*pa.DPParams, check.Equals, dp.Params{MinHitLength: p.l, MinId: p.id})
			c.Check(pa.AvgIndexListLength(pa.FilterParams), floatApprox, p.list, 0.05)
		}
	}
}

func (s *S) TestPack(c *check.C) {
	p := NewPacker("")
	for k := byte(1); k <= maxk; k++ {
		a := &linear.Seq{
			Annotation: seq.Annotation{
				ID: fmt.Sprintf("deBruijn%d", k),
			},
			Seq: make(alphabet.Letters, 0, util.Pow(Q, k)),
		}
		for _, i := range util.DeBruijn(byte(Q), k) {
			a.Seq = append(a.Seq, alphabet.Letter(l[i]))
		}
		ps, _ := p.Pack(a)
		c.Logf("%d: %s", k, ps)
	}
	c.Check(p.FinalisePack().Len(), check.Equals, 94208)
}

func (s *S) TestFeaturise(c *check.C) {
	for _, t := range T {
		f, err := ps.feature(t.start, t.end, false)
		if err != nil {
			c.Fatal(err)
		}
		c.Check(f, check.DeepEquals, t.result)
	}
}

func (s *S) TestWrite(c *check.C) {
	b := &B{&bytes.Buffer{}}
	w := NewWriter(b, 4, 60, false)
	for _, t := range T {
		if f1, err := ps.feature(t.start, t.end, false); err != nil {
			c.Fatal(err)
		} else {
			if f2, err := ps.feature(t.start, t.end, false); err != nil {
				c.Fatal(err)
			} else {
				n, err := w.Write(&Pair{A: f1, B: f2})
				c.Check(n, check.Not(check.Equals), 0)
				c.Check(err, check.Equals, nil)
			}
		}
	}
	c.Check(string(b.Bytes()), check.Equals,
		`deBruijn2	pals	hit	1	6	0.0000	.	.	Target deBruijn2 1 6; maxe 0
deBruijn2	pals	hit	2	6	0.0000	.	.	Target deBruijn2 2 6; maxe 0
deBruijn2	pals	hit	1	16	0.0000	.	.	Target deBruijn2 1 16; maxe 0
deBruijn1	pals	hit	1	4	0.0000	.	.	Target deBruijn1 1 4; maxe 0
deBruijn5	pals	hit	1	904	0.0000	.	.	Target deBruijn5 1 904; maxe 0
deBruijn8	pals	hit	1025	4095	0.0000	.	.	Target deBruijn8 1025 4095; maxe 0
`)
}
