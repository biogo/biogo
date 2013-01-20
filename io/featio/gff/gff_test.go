// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gff

import (
	"bytes"
	"code.google.com/p/biogo/alphabet"
	"code.google.com/p/biogo/bio"
	"code.google.com/p/biogo/feat"
	"code.google.com/p/biogo/seq"
	"code.google.com/p/biogo/seq/linear"
	"io"
	check "launchpad.net/gocheck"
	"strings"
	"testing"
	"time"
)

// Helpers
func floatPtr(f float64) *float64 { return &f }

func mustTime(t time.Time, err error) time.Time {
	if err != nil {
		panic(err)
	}
	return t
}

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

var (
	gffTests = []struct {
		gff  string
		feat []*Feature
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
			feat: []*Feature{
				{SeqName: "SEQ1", Source: "EMBL", Feature: "atg", FeatStart: 102, FeatEnd: 105, FeatScore: nil, FeatFrame: Frame0, FeatStrand: seq.Plus},
				{SeqName: "SEQ1", Source: "EMBL", Feature: "exon", FeatStart: 102, FeatEnd: 172, FeatScore: nil, FeatFrame: Frame0, FeatStrand: seq.Plus},
				{SeqName: "SEQ1", Source: "EMBL", Feature: "splice5", FeatStart: 171, FeatEnd: 173, FeatScore: nil, FeatFrame: NoFrame, FeatStrand: seq.Plus},
				{SeqName: "SEQ1", Source: "netgene", Feature: "splice5", FeatStart: 171, FeatEnd: 173, FeatScore: floatPtr(0.94), FeatFrame: NoFrame, FeatStrand: seq.Plus},
				{SeqName: "SEQ1", Source: "genie", Feature: "sp5-20", FeatStart: 162, FeatEnd: 182, FeatScore: floatPtr(2.3), FeatFrame: NoFrame, FeatStrand: seq.Plus},
				{SeqName: "SEQ1", Source: "genie", Feature: "sp5-10", FeatStart: 167, FeatEnd: 177, FeatScore: floatPtr(2.1), FeatFrame: NoFrame, FeatStrand: seq.Plus},
				{SeqName: "SEQ2", Source: "grail", Feature: "ATG", FeatStart: 16, FeatEnd: 19, FeatScore: floatPtr(2.1), FeatFrame: Frame0, FeatStrand: seq.Minus},
			},
		},
	}

	metaTests = []struct {
		date          string
		format        string
		sourceVersion string
		version       int
		name          string
		molType       bio.Moltype
		gff           string
		feat          []feat.Feature
		write         []interface{}
	}{
		{
			date:          "1997-11-08",
			format:        Astronomical,
			sourceVersion: "<source> <version-text>",
			version:       2,
			name:          "<seqname>",
			molType:       bio.DNA,

			gff: `##gff-version 2
##source-version <source> <version-text>
##date 1997-11-08
##Type DNA <seqname>
##DNA <seqname>
##acggctcggattggcgctggatgatagatcagacgac
##...
##end-DNA
##RNA <seqname>
##acggcucggauuggcgcuggaugauagaucagacgac
##...
##end-RNA
##Protein <seqname>
##MVLSPADKTNVKAAWGKVGAHAGEYGAEALERMFLSF
##...
##end-Protein
##sequence-region <seqname> 1 5
`,
			feat: []feat.Feature{
				linear.NewSeq("<seqname>", alphabet.BytesToLetters([]byte("acggctcggattggcgctggatgatagatcagacgac...")), alphabet.DNA),
				linear.NewSeq("<seqname>", alphabet.BytesToLetters([]byte("acggcucggauuggcgcuggaugauagaucagacgac...")), alphabet.RNA),
				linear.NewSeq("<seqname>", alphabet.BytesToLetters([]byte("MVLSPADKTNVKAAWGKVGAHAGEYGAEALERMFLSF...")), alphabet.Protein),
				&Region{Sequence: Sequence{SeqName: "<seqname>", Type: bio.DNA}, RegionStart: 0, RegionEnd: 5},
			},
			write: []interface{}{
				2,
				"source-version <source> <version-text>",
				mustTime(time.Parse(Astronomical, "1997-11-08")),
				Sequence{SeqName: "<seqname>", Type: bio.DNA},
				linear.NewSeq("<seqname>", alphabet.BytesToLetters([]byte("acggctcggattggcgctggatgatagatcagacgac...")), alphabet.DNA),
				linear.NewSeq("<seqname>", alphabet.BytesToLetters([]byte("acggcucggauuggcgcuggaugauagaucagacgac...")), alphabet.RNA),
				linear.NewSeq("<seqname>", alphabet.BytesToLetters([]byte("MVLSPADKTNVKAAWGKVGAHAGEYGAEALERMFLSF...")), alphabet.Protein),
				&Region{Sequence: Sequence{SeqName: "<seqname>"}, RegionStart: 0, RegionEnd: 5},
			},
		},
	}
)

func (s *S) TestReadGFF(c *check.C) {
	for i, g := range gffTests {
		buf := strings.NewReader(g.gff)
		r := NewReader(buf)
		for j := 0; ; j++ {
			f, err := r.Read()
			if err == io.EOF {
				c.Check(j, check.Equals, len(g.feat))
				break
			}
			c.Check(f, check.DeepEquals, g.feat[j], check.Commentf("Test: %d Line: %d", i, j+1))
			c.Check(err, check.Equals, nil)
		}
	}
}

func (s *S) TestReadMetaline(c *check.C) {
	for i, g := range metaTests {
		buf := strings.NewReader(g.gff)
		r := NewReader(buf)
		for j := 0; ; j++ {
			f, err := r.Read()
			if err == io.EOF {
				c.Check(j, check.Equals, len(g.feat))
				break
			}
			c.Check(f, check.DeepEquals, g.feat[j], check.Commentf("Test: %d Line: %d", i, j+1))
			c.Check(err, check.Equals, nil)
		}
		c.Check(r.Version, check.Equals, g.version)
		date, err := time.Parse(g.format, g.date)
		c.Assert(err, check.Equals, nil)
		c.Check(r.Name, check.Equals, g.name)
		c.Check(r.Type, check.Equals, g.molType)
		c.Check(r.Date, check.Equals, date)
		c.Check(r.SourceVersion, check.Equals, g.sourceVersion)
	}
}

const width = 37 // Not the normal fasta width - this matches the examples from the GFF spec page.

func (s *S) TestWriteGff(c *check.C) {
	for i, g := range gffTests {
		buf := &bytes.Buffer{}
		w := NewWriter(buf, width, false)
		for _, f := range g.feat {
			_, err := w.Write(f)
			c.Check(err, check.Equals, nil)
		}
		w.Close()
		c.Check(buf.String(), check.Equals, g.gff, check.Commentf("Test: %d", i))
	}
}

func (s *S) TestWriteMetadata(c *check.C) {
	for i, g := range metaTests {
		buf := &bytes.Buffer{}
		w := NewWriter(buf, width, false)
		for _, d := range g.write {
			_, err := w.WriteMetaData(d)
			c.Check(err, check.Equals, nil)
		}
		w.Close()
		c.Check(buf.String(), check.Equals, g.gff, check.Commentf("Test: %d", i))
	}
}
