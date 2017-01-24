// Copyright ©2015 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gene

import (
	"github.com/biogo/biogo/feat"

	"testing"

	"gopkg.in/check.v1"
)

// Assert that interfaces are satisfied
var (
	_ feat.Feature = (*Gene)(nil)
	_ feat.Feature = (*NonCodingTranscript)(nil)
	_ feat.Feature = (*CodingTranscript)(nil)
	_ feat.Feature = (*Exon)(nil)
	_ feat.Feature = (*Intron)(nil)
	_ feat.Feature = (*TranscriptFeature)(nil)

	_ featureOrienter = (*Gene)(nil)
	_ featureOrienter = (*NonCodingTranscript)(nil)
	_ featureOrienter = (*CodingTranscript)(nil)
	_ featureOrienter = (*Exon)(nil)
	_ featureOrienter = (*Intron)(nil)
	_ featureOrienter = (*TranscriptFeature)(nil)

	_ Transcript = (*NonCodingTranscript)(nil)
	_ Transcript = (*CodingTranscript)(nil)

	_ Interface = (*Gene)(nil)
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

// Create the test suite
type S struct{}

var _ = check.Suite(&S{})

// Chr implements feat.Feature and is used as location for the test objects.
type Chr string

func (c Chr) Start() int             { return 0 }
func (c Chr) End() int               { return 0 }
func (c Chr) Len() int               { return 0 }
func (c Chr) Name() string           { return string(c) }
func (c Chr) Description() string    { return "chrom" }
func (c Chr) Location() feat.Feature { return nil }

// ori implements feat.Feature and is used as location for test objects.
type ori struct {
	start, end int
	name       string
	desc       string
	loc        feat.Feature
	orient     feat.Orientation
}

func (o ori) Name() string                  { return o.name }
func (o ori) Description() string           { return o.desc }
func (o ori) Start() int                    { return o.start }
func (o ori) End() int                      { return o.end }
func (o ori) Len() int                      { return o.end - o.start }
func (o ori) Location() feat.Feature        { return o.loc }
func (o ori) Orientation() feat.Orientation { return o.orient }

// Define some test objects that will be used in the actual tests
var (
	geneA = Gene{
		ID:     "geneA",
		Chrom:  Chr("Y"),
		Offset: 100,
		Orient: feat.Forward,
		Desc:   "forward gene",
	}
	geneB = Gene{
		ID:     "geneB",
		Chrom:  Chr("X"),
		Offset: 100,
		Orient: feat.Reverse,
		Desc:   "reverse gene",
	}
	geneC = Gene{
		ID: "geneC",
		Chrom: ori{
			start:  0,
			end:    800,
			orient: feat.Reverse,
			loc: ori{
				start:  0,
				end:    900,
				orient: feat.Forward,
				loc: ori{
					start:  0,
					end:    1000,
					orient: feat.Reverse,
				}}},
		Offset: 100,
		Orient: feat.Reverse,
		Desc:   "reverse gene on a contig on a supercontig on an ultra contig.",
	}
	codingTranscriptA = CodingTranscript{
		ID:       "codingTranscriptA",
		Loc:      Chr("Y"),
		Offset:   100,
		CDSstart: 100,
		CDSend:   600,
		Orient:   feat.Forward,
		Desc:     "forward transcript with cds",
	}
	codingTranscriptB = CodingTranscript{
		ID:       "codingTranscriptB",
		Loc:      Chr("X"),
		Offset:   500,
		CDSstart: 300,
		CDSend:   1300,
		Orient:   feat.Reverse,
		Desc:     "reverse transcript with cds",
	}
	codingTranscriptC = CodingTranscript{
		ID:       "codingTranscriptC",
		Loc:      &geneC,
		Offset:   20,
		CDSstart: 100,
		CDSend:   500,
		Orient:   feat.Forward,
		Desc:     "forward transcript with cds on reverse gene",
	}
	nonCodingTranscriptA = NonCodingTranscript{
		ID:     "nonCodingTranscriptA",
		Loc:    Chr("Y"),
		Offset: 100,
		Orient: feat.Forward,
		Desc:   "forward non coding transcript",
	}
	nonCodingTranscriptB = NonCodingTranscript{
		ID:     "nonCodingTranscriptB",
		Loc:    Chr("X"),
		Offset: 500,
		Orient: feat.Reverse,
		Desc:   "reverse non coding transcript",
	}
)

// Tests for Gene
var geneTests = []struct {
	Test        string
	Gene        Interface
	Name        string
	Chrom       string
	Start, End  int
	Len         int
	Orientation feat.Orientation
	Feats       []feat.Feature
	SetErr      string
	TransCount  int
}{
	{
		Test:        "forward gene with valid feats",
		Gene:        &geneA,
		Name:        "geneA",
		Chrom:       "Y",
		Start:       100,
		End:         120,
		Len:         20,
		Orientation: feat.Forward,
		Feats: []feat.Feature{
			&NonCodingTranscript{Loc: &geneA, exons: []Exon{{Length: 20}}},
			&NonCodingTranscript{Loc: &geneA},
		},
		TransCount: 2,
	},
	{
		Test:        "reverse gene with valid feats",
		Gene:        &geneB,
		Name:        "geneB",
		Chrom:       "X",
		Start:       100,
		End:         110,
		Len:         10,
		Orientation: feat.Reverse,
		Feats: []feat.Feature{
			&NonCodingTranscript{Loc: &geneB, exons: []Exon{{Length: 10}}},
			&NonCodingTranscript{Loc: &geneB},
		},
		TransCount: 2,
	},
	{
		Test:   "forward gene with feat on wrong location",
		Gene:   &geneA,
		Feats:  []feat.Feature{&NonCodingTranscript{Loc: &geneB}},
		SetErr: "transcript location does not match the gene",
	},
	{
		Test:   "reverse gene with no feat from 0",
		Gene:   &geneB,
		Feats:  []feat.Feature{&NonCodingTranscript{Loc: &geneB, Offset: 5}},
		SetErr: "no transcript with 0 start on gene",
	},
}

func (s *S) TestGene(c *check.C) {
	for _, d := range geneTests {
		g := d.Gene

		// Test SetFeatures
		if err := g.SetFeatures(d.Feats...); err != nil {
			c.Assert(err, check.ErrorMatches, d.SetErr)
		} else {
			c.Check(g.Name(), check.Equals, d.Name)
			c.Check(g.Start(), check.Equals, d.Start)
			c.Check(g.End(), check.Equals, d.End)
			c.Check(g.Len(), check.Equals, d.Len)
			c.Check(g.Location().Name(), check.Equals, d.Chrom)
			c.Check(g.Orientation(), check.Equals, d.Orientation)
			c.Check(len(TranscriptsOf(g)), check.Equals, d.TransCount)
		}
	}
}

// Tests for Transcript
var transcriptTests = []struct {
	Test               string
	Transcript         Transcript
	Name               string
	Loc                feat.Feature
	Start, End         int
	UTR5start, UTR5end int
	CDSstart, CDSend   int
	UTR3start, UTR3end int
	Len                int
	Orientation        feat.Orientation
	Exons              []Exon
	AddErr             string
	ExonicLen          int
}{
	{
		Test:        "forward transcript with cds and valid exons",
		Transcript:  &codingTranscriptA,
		Name:        "codingTranscriptA",
		Loc:         Chr("Y"),
		Orientation: feat.Forward,
		Exons: []Exon{
			{Transcript: &codingTranscriptA, Offset: 0, Length: 300},
			{Transcript: &codingTranscriptA, Offset: 600, Length: 200}},
		Start:     100,
		End:       900,
		UTR5start: 0,
		UTR5end:   100,
		CDSstart:  100,
		CDSend:    600,
		UTR3start: 600,
		UTR3end:   800,
		Len:       800,
		ExonicLen: 500,
	},
	{
		Test:        "reverse transcript with cds and valid exons",
		Transcript:  &codingTranscriptB,
		Name:        "codingTranscriptB",
		Loc:         Chr("X"),
		Orientation: feat.Reverse,
		Exons: []Exon{
			{Transcript: &codingTranscriptB, Offset: 0, Length: 600},
			{Transcript: &codingTranscriptB, Offset: 900, Length: 600}},
		Start:     500,
		End:       2000,
		UTR3start: 0,
		UTR3end:   300,
		CDSstart:  300,
		CDSend:    1300,
		UTR5start: 1300,
		UTR5end:   1500,
		Len:       1500,
		ExonicLen: 1200,
	},
	{
		Test:        "forward transcript with cds and valid exons on reverse gene on a contig on a supercontig on an ultra contig.",
		Transcript:  &codingTranscriptC,
		Name:        "codingTranscriptC",
		Loc:         &geneC,
		Orientation: feat.Forward,
		Exons: []Exon{
			{Transcript: &codingTranscriptC, Offset: 0, Length: 500},
			{Transcript: &codingTranscriptC, Offset: 600, Length: 100}},
		Start:     20,
		End:       720,
		UTR3start: 0,
		UTR3end:   100,
		CDSstart:  100,
		CDSend:    500,
		UTR5start: 500,
		UTR5end:   700,
		Len:       700,
		ExonicLen: 600,
	},
	{
		Test:        "forward non-coding transcript with valid exons",
		Transcript:  &nonCodingTranscriptA,
		Name:        "nonCodingTranscriptA",
		Loc:         Chr("Y"),
		Orientation: feat.Forward,
		Exons: []Exon{
			{Transcript: &nonCodingTranscriptA, Offset: 0, Length: 300},
			{Transcript: &nonCodingTranscriptA, Offset: 600, Length: 200}},
		Start:     100,
		End:       900,
		Len:       800,
		ExonicLen: 500,
	},
	{
		Test:        "reverse non-coding transcript without exon at 0",
		Transcript:  &nonCodingTranscriptB,
		Orientation: feat.Reverse,
		Exons:       []Exon{{Transcript: &nonCodingTranscriptB, Offset: 10}},
		AddErr:      "no exon with a zero start",
	},
	{
		Test:        "reverse non-coding transcript with wrong exon location",
		Transcript:  &nonCodingTranscriptB,
		Orientation: feat.Reverse,
		Exons:       []Exon{{Offset: 0, Length: 10000}},
		AddErr:      "exon location is not the transcript",
	},
}

func (s *S) TestTranscript(c *check.C) {
	for _, d := range transcriptTests {
		t := d.Transcript

		// Test SetExons
		if err := t.SetExons(d.Exons...); err != nil {
			c.Assert(err, check.ErrorMatches, d.AddErr)
		} else {
			t.Exons()[0].Offset = 1000000 // should have no effect on t

			c.Check(t.Name(), check.Equals, d.Name)
			c.Check(t.Start(), check.Equals, d.Start)
			c.Check(t.End(), check.Equals, d.End)
			c.Check(t.Len(), check.Equals, d.Len)
			c.Check(t.Location(), check.Equals, d.Loc)
			c.Check(t.Orientation(), check.Equals, d.Orientation)
			c.Check(t.Exons().SplicedLen(), check.Equals, d.ExonicLen)

			// Test CodingTranscript specifics
			if t, ok := t.(*CodingTranscript); ok {
				utr5, cds, utr3 := t.UTR5(), t.CDS(), t.UTR3()
				c.Check(utr5.Start(), check.Equals, d.UTR5start)
				c.Check(utr5.End(), check.Equals, d.UTR5end)
				c.Check(utr5.Location(), check.Equals, t)
				c.Check(cds.Start(), check.Equals, d.CDSstart)
				c.Check(cds.End(), check.Equals, d.CDSend)
				c.Check(cds.Location(), check.Equals, t)
				c.Check(utr3.Start(), check.Equals, d.UTR3start)
				c.Check(utr3.End(), check.Equals, d.UTR3end)
				c.Check(utr3.Location(), check.Equals, t)

				c.Check(t.CDSstart, check.Equals, d.CDSstart)
				c.Check(t.CDSend, check.Equals, d.CDSend)
				c.Check(t.UTR5start(), check.Equals, d.UTR5start)
				c.Check(t.UTR5end(), check.Equals, d.UTR5end)
				c.Check(t.UTR3start(), check.Equals, d.UTR3start)
				c.Check(t.UTR3end(), check.Equals, d.UTR3end)
			}
		}
	}
}

// Tests for Exon and Intron
type featureOrienter interface {
	feat.Orienter
	feat.Feature
}

var exonIntronTests = []struct {
	Test        string
	Feat        featureOrienter
	Start, End  int
	Len         int
	Transcript  feat.Feature
	Orientation feat.Orientation
}{
	{
		Test:        "Exon on transcript",
		Feat:        Exon{Offset: 200, Length: 200},
		Start:       200,
		End:         400,
		Len:         200,
		Orientation: feat.Forward,
	},
	{
		Test:        "Intron on transcript",
		Feat:        Intron{Offset: 300, Length: 500},
		Start:       300,
		End:         800,
		Len:         500,
		Orientation: feat.Forward,
	},
}

func (s *S) TestExonIntron(c *check.C) {
	for _, d := range exonIntronTests {
		e := d.Feat

		c.Check(e.Start(), check.Equals, d.Start)
		c.Check(e.End(), check.Equals, d.End)
		c.Check(e.Len(), check.Equals, d.Len)
		c.Check(e.Location(), check.DeepEquals, d.Transcript)
		c.Check(e.Orientation(), check.Equals, d.Orientation)
	}
}

// Tests for Exons
var exonsTests = []struct {
	Test                        string
	InputExons                  []Exon
	Location                    feat.Feature
	Start, End, Len, SplicedLen int
	AddErr                      string
	MadeIntrons                 Introns
}{
	{
		Test: "Exons not in order",
		InputExons: []Exon{
			{Offset: 300, Length: 100},
			{Offset: 0, Length: 100},
		},
		Start:      0,
		End:        400,
		Len:        2,
		SplicedLen: 200,
		MadeIntrons: Introns{
			Intron{Offset: 100, Length: 200},
		},
	},
	{
		Test: "Exons overlap",
		InputExons: []Exon{
			{Offset: 0, Length: 100},
			{Offset: 50, Length: 100},
		},
		AddErr: "exons overlap",
	},
	{
		Test: "Exons on different transcripts",
		InputExons: []Exon{
			{Transcript: &codingTranscriptA},
			{Transcript: &codingTranscriptB},
		},
		AddErr: "exons location differ",
	},
}

func (s *S) TestExons(c *check.C) {
	for _, d := range exonsTests {
		var e Exons
		ie := d.InputExons

		// Test SetExons
		if e, err := e.Add(ie...); err != nil {
			c.Assert(err, check.ErrorMatches, d.AddErr)
		} else {
			c.Check(e.Location(), check.DeepEquals, d.Location)
			c.Check(e.Start(), check.Equals, d.Start)
			c.Check(e.End(), check.Equals, d.End)
			c.Check(e.Len(), check.Equals, d.Len)
			c.Check(e.SplicedLen(), check.Equals, d.SplicedLen)
			c.Check(e.Introns(), check.DeepEquals, d.MadeIntrons)
		}
	}
}
