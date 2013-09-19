// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package featio_test

import (
	"code.google.com/p/biogo/io/featio"
	"code.google.com/p/biogo/io/featio/gff"
	"code.google.com/p/biogo/seq"

	"bytes"
	check "launchpad.net/gocheck"
	"testing"
)

// Helpers
func floatPtr(f float64) *float64 { return &f }

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

func (s *S) TestReadGFF(c *check.C) {
	for i, g := range []struct {
		gff  string
		feat []*gff.Feature
	}{
		{
			gff: `SEQ1	EMBL	atg	103	105	.	+	0
SEQ1	EMBL	exon	103	172	.	+	0
SEQ1	EMBL	splice5	172	173	.	+	.
SEQ1	netgene	splice5	172	173	0.94	+	.
SEQ1	genie	sp5-20	163	182	2.3	+	.
SEQ1	genie	sp5-10	168	177	2.1	+	.
SEQ2	grail	ATG	17	19	2.1	-	0
`,
			feat: []*gff.Feature{
				{SeqName: "SEQ1", Source: "EMBL", Feature: "atg", FeatStart: 102, FeatEnd: 105, FeatScore: nil, FeatFrame: gff.Frame0, FeatStrand: seq.Plus},
				{SeqName: "SEQ1", Source: "EMBL", Feature: "exon", FeatStart: 102, FeatEnd: 172, FeatScore: nil, FeatFrame: gff.Frame0, FeatStrand: seq.Plus},
				{SeqName: "SEQ1", Source: "EMBL", Feature: "splice5", FeatStart: 171, FeatEnd: 173, FeatScore: nil, FeatFrame: gff.NoFrame, FeatStrand: seq.Plus},
				{SeqName: "SEQ1", Source: "netgene", Feature: "splice5", FeatStart: 171, FeatEnd: 173, FeatScore: floatPtr(0.94), FeatFrame: gff.NoFrame, FeatStrand: seq.Plus},
				{SeqName: "SEQ1", Source: "genie", Feature: "sp5-20", FeatStart: 162, FeatEnd: 182, FeatScore: floatPtr(2.3), FeatFrame: gff.NoFrame, FeatStrand: seq.Plus},
				{SeqName: "SEQ1", Source: "genie", Feature: "sp5-10", FeatStart: 167, FeatEnd: 177, FeatScore: floatPtr(2.1), FeatFrame: gff.NoFrame, FeatStrand: seq.Plus},
				{SeqName: "SEQ2", Source: "grail", Feature: "ATG", FeatStart: 16, FeatEnd: 19, FeatScore: floatPtr(2.1), FeatFrame: gff.Frame0, FeatStrand: seq.Minus},
			},
		},
	} {
		sc := featio.NewScanner(
			gff.NewReader(
				bytes.NewBufferString(g.gff),
			),
		)

		var j int
		for sc.Next() {
			f := sc.Feat()
			c.Check(f, check.DeepEquals, g.feat[j], check.Commentf("Test: %d Line: %d", i, j+1))
			j++
		}
		c.Check(sc.Error(), check.Equals, nil)
		c.Check(j, check.Equals, len(g.feat))
	}
}

func (s *S) TestReadFromFunc(c *check.C) {
	for i, g := range []struct {
		gff  string
		feat []*gff.Feature
	}{
		{
			gff: `SEQ1	EMBL	atg	103	105	.	+	0
SEQ1	EMBL	exon	103	172	.	+	0
SEQ1	EMBL	splice5	172	173	.	+	.
SEQ1	netgene	splice5	172	173	0.94	+	.
SEQ1	genie	sp5-20	163	182	2.3	+	.
SEQ1	genie	sp5-10	168	177	2.1	+	.
SEQ2	grail	ATG	17	19	2.1	-	0
`,
			feat: []*gff.Feature{
				{SeqName: "SEQ1", Source: "EMBL", Feature: "atg", FeatStart: 102, FeatEnd: 105, FeatScore: nil, FeatFrame: gff.Frame0, FeatStrand: seq.Plus},
				{SeqName: "SEQ1", Source: "EMBL", Feature: "exon", FeatStart: 102, FeatEnd: 172, FeatScore: nil, FeatFrame: gff.Frame0, FeatStrand: seq.Plus},
				{SeqName: "SEQ1", Source: "EMBL", Feature: "splice5", FeatStart: 171, FeatEnd: 173, FeatScore: nil, FeatFrame: gff.NoFrame, FeatStrand: seq.Plus},
				{SeqName: "SEQ1", Source: "netgene", Feature: "splice5", FeatStart: 171, FeatEnd: 173, FeatScore: floatPtr(0.94), FeatFrame: gff.NoFrame, FeatStrand: seq.Plus},
				{SeqName: "SEQ1", Source: "genie", Feature: "sp5-20", FeatStart: 162, FeatEnd: 182, FeatScore: floatPtr(2.3), FeatFrame: gff.NoFrame, FeatStrand: seq.Plus},
				{SeqName: "SEQ1", Source: "genie", Feature: "sp5-10", FeatStart: 167, FeatEnd: 177, FeatScore: floatPtr(2.1), FeatFrame: gff.NoFrame, FeatStrand: seq.Plus},
				{SeqName: "SEQ2", Source: "grail", Feature: "ATG", FeatStart: 16, FeatEnd: 19, FeatScore: floatPtr(2.1), FeatFrame: gff.Frame0, FeatStrand: seq.Minus},
			},
		},
	} {
		sc := featio.NewScannerFromFunc(
			gff.NewReader(
				bytes.NewBufferString(g.gff),
			).Read,
		)

		var j int
		for sc.Next() {
			f := sc.Feat()
			c.Check(f, check.DeepEquals, g.feat[j], check.Commentf("Test: %d Line: %d", i, j+1))
			j++
		}
		c.Check(sc.Error(), check.Equals, nil)
		c.Check(j, check.Equals, len(g.feat))
	}
}
