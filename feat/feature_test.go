// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package feat_test

import (
	"github.com/biogo/biogo/feat"
	"github.com/biogo/biogo/seq/alignment"
	"github.com/biogo/biogo/seq/linear"
	"github.com/biogo/biogo/seq/multi"

	"gopkg.in/check.v1"
)

var (
	_ feat.Feature = (*linear.Seq)(nil)
	_ feat.Feature = (*linear.QSeq)(nil)
	_ feat.Feature = (*alignment.Seq)(nil)
	_ feat.Feature = (*alignment.QSeq)(nil)
	_ feat.Feature = (*multi.Multi)(nil)

	_ feat.Offsetter = (*linear.Seq)(nil)
	_ feat.Offsetter = (*linear.QSeq)(nil)
	_ feat.Offsetter = (*alignment.Seq)(nil)
	_ feat.Offsetter = (*alignment.QSeq)(nil)
	_ feat.Offsetter = (*multi.Multi)(nil)

	_ feat.LocationSetter = (*linear.Seq)(nil)
	_ feat.LocationSetter = (*linear.QSeq)(nil)
	_ feat.LocationSetter = (*alignment.Seq)(nil)
	_ feat.LocationSetter = (*alignment.QSeq)(nil)
	_ feat.LocationSetter = (*multi.Multi)(nil)
)

type S struct{}

var _ = check.Suite(&S{})

type chrom int

func (c chrom) Name() string           { return "test" }
func (c chrom) Description() string    { return "chromosome" }
func (c chrom) Start() int             { return 0 }
func (c chrom) End() int               { return int(c) }
func (c chrom) Len() int               { return int(c) }
func (c chrom) Location() feat.Feature { return nil }

type nonOri struct {
	start, end int
	name       string
	desc       string
	loc        feat.Feature
}

func (o nonOri) Name() string           { return o.name }
func (o nonOri) Description() string    { return o.desc }
func (o nonOri) Start() int             { return o.start }
func (o nonOri) End() int               { return o.end }
func (o nonOri) Len() int               { return o.end - o.start }
func (o nonOri) Location() feat.Feature { return o.loc }

type ori struct {
	nonOri
	orient feat.Orientation
}

func (o ori) Orientation() feat.Orientation { return o.orient }

var (
	chrom1 = chrom(1000)
	chrom2 = chrom(500)
	geneA  = ori{
		nonOri: nonOri{
			start: 10, end: 50,
			name: "genA",
			desc: "gene",
			loc:  chrom1,
		},
		orient: feat.Forward,
	}
	proA = ori{
		nonOri: nonOri{
			start: 10, end: 20,
			name: "genA",
			desc: "promoter",
			loc:  geneA,
		},
		orient: feat.Forward,
	}
	pribA = ori{
		nonOri: nonOri{
			start: 15, end: 20,
			name: "pribA",
			desc: "promoter box",
			loc:  proA,
		},
		orient: feat.Forward,
	}
	opA = nonOri{
		start: 12, end: 16,
		name: "genb",
		desc: "operator",
		loc:  proA,
	}
	orfA = ori{
		nonOri: nonOri{
			start: 15, end: 30,
			name: "genA",
			desc: "orf",
			loc:  geneA,
		},
		orient: feat.Forward,
	}
	antiA = ori{
		nonOri: nonOri{
			start: 45, end: 50,
			name: "genA",
			desc: "antisense",
			loc:  geneA,
		},
		orient: feat.Reverse,
	}

	geneB = ori{
		nonOri: nonOri{
			start: 60, end: 100,
			name: "genB",
			desc: "gene",
			loc:  chrom1,
		},
		orient: feat.Reverse,
	}
	proB = ori{
		nonOri: nonOri{
			start: 90, end: 100,
			name: "genB",
			desc: "promoter",
			loc:  geneB,
		},
		orient: feat.Forward,
	}
	opB = nonOri{
		start: 94, end: 98,
		name: "genB",
		desc: "operator",
		loc:  proB,
	}
	orfB = ori{
		nonOri: nonOri{
			start: 15, end: 30,
			name: "genb",
			desc: "orf",
			loc:  geneB,
		},
		orient: feat.Forward,
	}

	pal = nonOri{
		start: 300, end: 320,
		name: "palA",
		desc: "palindrome",
		loc:  chrom1,
	}

	freeOri = ori{
		nonOri: nonOri{
			start: 10, end: 100,
			name: "frag",
			desc: "fragment",
		},
		orient: feat.Reverse,
	}

	orientationTests = []struct {
		f   feat.Feature
		ori feat.Orientation
		ref feat.Feature
	}{
		{
			f:   chrom1,
			ori: feat.Forward,
			ref: chrom1,
		},
		{
			f:   geneA,
			ori: feat.Forward,
			ref: chrom1,
		},
		{
			f:   orfA,
			ori: feat.Forward,
			ref: chrom1,
		},
		{
			f:   antiA,
			ori: feat.Reverse,
			ref: chrom1,
		},
		{
			f:   proA,
			ori: feat.Forward,
			ref: chrom1,
		},
		{
			f:   opA,
			ori: feat.Forward,
			ref: opA,
		},
		{
			f:   geneB,
			ori: feat.Reverse,
			ref: chrom1,
		},
		{
			f:   orfB,
			ori: feat.Reverse,
			ref: chrom1,
		},
		{
			f:   proB,
			ori: feat.Reverse,
			ref: chrom1,
		},
		{
			f:   opB,
			ori: feat.Forward,
			ref: opB,
		},
		{
			f:   pal,
			ori: feat.Forward,
			ref: pal,
		},
		{
			f:   freeOri,
			ori: feat.Reverse,
			ref: nil,
		},
	}
)

func (s *S) TestBaseOrientationOf(c *check.C) {
	for _, t := range orientationTests {
		ori, ref := feat.BaseOrientationOf(t.f)
		c.Check(ori, check.Equals, t.ori)
		c.Check(ref, check.Equals, t.ref)
	}

	// Check that we find the same reference where possible.
	_, refGeneA := feat.BaseOrientationOf(geneA)
	_, refGeneB := feat.BaseOrientationOf(geneB)
	c.Check(refGeneA, check.Equals, refGeneB)

	// Check that unorientable features return different reference features.
	_, refOpA := feat.BaseOrientationOf(opA)
	_, refOpB := feat.BaseOrientationOf(opB)
	c.Check(refOpA, check.Not(check.Equals), refOpB)
	_, refChrom1 := feat.BaseOrientationOf(chrom1)
	_, refChrom2 := feat.BaseOrientationOf(chrom2)
	c.Check(refChrom1, check.Not(check.Equals), refChrom2)

	// Check we detect cycles.
	var cycle ori
	cycle.orient = feat.Forward
	cycle.loc = &cycle
	c.Check(func() { feat.BaseOrientationOf(cycle) }, check.Panics, "feat: feature chain too long")
}

func (s *S) TestOrientationWithin(c *check.C) {
	for _, t := range orientationTests {
		c.Check(feat.OrientationWithin(t.f, t.ref), check.Equals, t.ori)
	}

	// Check we can find things that are all along the feature chain.
	for ref := feat.Feature(pribA); ref != nil; ref = ref.Location() {
		c.Check(feat.OrientationWithin(pribA, ref), check.Equals, feat.Forward)
	}

	// Check that unorientable features return different reference features.
	c.Check(feat.OrientationWithin(pribA, nil), check.Equals, feat.NotOriented)
	c.Check(feat.OrientationWithin(opA, chrom2), check.Equals, feat.NotOriented)
	c.Check(feat.OrientationWithin(opA, geneB), check.Equals, feat.NotOriented)

	// Check we detect cycles.
	var cycle ori
	cycle.orient = feat.Forward
	cycle.loc = &cycle
	c.Check(func() { feat.OrientationWithin(cycle, chrom1) }, check.Panics, "feat: feature chain too long")
}

// Tests for BasePositionOf and PositionWithin.
var baseCoordsTests = []struct {
	f, ref                  feat.Feature
	pos, basePos, posWithin int
}{
	{
		f:         chrom1,
		pos:       20,
		basePos:   20,
		posWithin: 20,
		ref:       chrom1,
	},
	{
		f:         geneA,
		pos:       30,
		basePos:   40,
		posWithin: 40,
		ref:       chrom1,
	},
	{
		f:         proA,
		pos:       40,
		basePos:   60,
		posWithin: 60,
		ref:       chrom1,
	},
	{
		f:         opA,
		pos:       20,
		basePos:   52,
		posWithin: 52,
		ref:       chrom1,
	},
	{
		f:         antiA,
		pos:       20,
		basePos:   75,
		posWithin: 75,
		ref:       chrom1,
	},
	{
		f:         freeOri,
		pos:       0,
		basePos:   10,
		posWithin: 0,
		ref:       freeOri,
	},
}

func (s *S) TestBasePositionOf(c *check.C) {
	for _, t := range baseCoordsTests {
		pos, ref := feat.BasePositionOf(t.f, t.pos)
		c.Check(pos, check.Equals, t.basePos)
		c.Check(ref, check.Equals, t.ref)
	}

	// Check that we find the same reference where possible.
	_, refGeneA := feat.BasePositionOf(geneA, 0)
	_, refGeneB := feat.BasePositionOf(geneB, 0)
	c.Check(refGeneA, check.Equals, refGeneB)

	// Check we detect cycles.
	var cycle ori
	cycle.loc = &cycle
	c.Check(func() { feat.BasePositionOf(cycle, 10) }, check.Panics, "feat: feature chain too long")
}

func (s *S) TestPositionWithin(c *check.C) {
	for _, t := range baseCoordsTests {
		pos, ok := feat.PositionWithin(t.f, t.ref, t.pos)
		c.Check(pos, check.Equals, t.posWithin)
		c.Check(ok, check.Equals, true)
	}

	// Check unorthodox tree structures.
	_, ok := feat.PositionWithin(opA, chrom2, 10)
	c.Check(ok, check.Equals, false)
	_, ok = feat.PositionWithin(nil, nil, 10)
	c.Check(ok, check.Equals, false)

	// Check we detect cycles.
	var cycle ori
	cycle.loc = &cycle
	c.Check(func() { feat.PositionWithin(cycle, chrom1, 10) }, check.Panics, "feat: feature chain too long")
}
