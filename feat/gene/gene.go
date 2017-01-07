// Copyright ©2015 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gene contains the types and methods to handle the definition of a
// gene. A gene is a union of genomic sequences encoding a coherent set of
// potentially overlapping functional products. Since the package is located
// under the feat namespace, we define gene to correspond to a specific
// genomic region (has genomic coordinates).
//
// The package also contain types to describe gene transcripts. Transcripts
// can be coding and non-coding. Coding transcripts have functional regions
// (5'UTR, CDS and 3'UTR) and consist of exons.
package gene

import (
	"github.com/biogo/biogo/feat"

	"errors"
	"sort"
)

const maxInt = int(^uint(0) >> 1) // The maximum int value.

// Interface defines the gene interface.
type Interface interface {
	feat.Feature
	feat.Orienter
	feat.Set
	SetFeatures(...feat.Feature) error
}

// Transcript is the interface for a gene transcript.
type Transcript interface {
	feat.Feature
	feat.Orienter
	Exons() Exons
	Introns() Introns
	SetExons(...Exon) error
}

// TranscriptsOf scans a feat.Set and returns any Transcripts that it finds.
func TranscriptsOf(s feat.Set) []Transcript {
	var ts []Transcript
	for _, f := range s.Features() {
		if t, ok := f.(Transcript); ok {
			ts = append(ts, t)
		}
	}
	return ts
}

// A Gene occupies a specific region on the genome and may have 0 or more
// features, including transcripts, associated with it. The gene is tightly
// coupled with its features in the sense that the gene boundaries are defined
// by the features. By definition one of the features must always start at
// position 0 relative to the gene and this or another one has to end at the
// end of the gene. The former is asserted when features are set and the
// latter is guaranteed by setting the gene end at the largest end of the
// features.
type Gene struct {
	ID     string
	Chrom  feat.Feature
	Offset int
	Orient feat.Orientation
	Desc   string
	length int
	feats  []feat.Feature
}

// Start returns the gene start on the chromosome.
func (g *Gene) Start() int { return g.Offset }

// End returns the gene end on the chromosome.
func (g *Gene) End() int { return g.Offset + g.Len() }

// Len returns the length of the gene.
func (g *Gene) Len() int { return g.length }

// Name returns the gene name. Currently the same as the id.
func (g *Gene) Name() string { return g.ID }

// Description returns a description for the gene.
func (g *Gene) Description() string { return g.Desc }

// Location returns the location of the gene. Namely the chromosome.
func (g *Gene) Location() feat.Feature { return g.Chrom }

// Orientation returns the orientation of the gene relative to the chromosome.
func (g *Gene) Orientation() feat.Orientation { return g.Orient }

// Features returns all features added to the gene.
func (g *Gene) Features() []feat.Feature { return g.feats }

// SetFeatures sets the gene features. Internally, it verifies that their
// Location is the gene and that one of them has zero Start. If an error
// occurs it is returned and the features are not set.
func (g *Gene) SetFeatures(feats ...feat.Feature) error {
	pos := maxInt
	end := 0
	for _, f := range feats {
		if f.Location() != g {
			return errors.New("transcript location does not match the gene")
		}
		if f.Start() < pos {
			pos = f.Start()
		}
		if f.End() > end {
			end = f.End()
		}
	}
	if pos != 0 {
		return errors.New("no transcript with 0 start on gene")
	}
	g.length = end - pos
	g.feats = feats
	return nil
}

// A NonCodingTranscript is a gene transcript that has no coding potential. It
// can be located on any feat.Feature such as a gene or a chromosome. The
// concept of exons is tightly coupled with the NonCodingTranscript in the
// sense that the transcript borders are basically defined by the contained
// exons. By definition one of the exons must always start at position 0
// relative to the transcript and this or another one must end at the end of
// transcript. The former is asserted when exons are set and the latter is
// guaranteed by setting the transcript end at the end of the last exon.
type NonCodingTranscript struct {
	ID     string
	Loc    feat.Feature
	Offset int
	Orient feat.Orientation
	Desc   string
	exons  Exons
}

// Start returns the transcript start relative to Location.
func (t *NonCodingTranscript) Start() int { return t.Offset }

// End returns the transcript end relative to Location.
func (t *NonCodingTranscript) End() int { return t.Offset + t.exons.End() }

// Len returns the length of the transcript.
func (t *NonCodingTranscript) Len() int { return t.End() - t.Start() }

// Name returns the transcript name. Currently the same as the id.
func (t *NonCodingTranscript) Name() string { return t.ID }

// Description returns a description for the transcript.
func (t *NonCodingTranscript) Description() string { return t.Desc }

// Location returns the location of the transcript. Can be any feat.Feature
// such as a gene or a chromosome.
func (t *NonCodingTranscript) Location() feat.Feature { return t.Loc }

// Orientation returns the orientation of the transcript relative to Location.
func (t *NonCodingTranscript) Orientation() feat.Orientation { return t.Orient }

// Exons returns a typed slice with the transcript exons.
func (t *NonCodingTranscript) Exons() Exons { return t.exons }

// Introns returns a typed slice with the transcript introns.
func (t *NonCodingTranscript) Introns() Introns { return t.exons.Introns() }

// SetExons sets the transcript exons. Internally, it sorts exons by Start,
// verifies that their Location is the transcript, that they are not
// overlapping and that one has zero Start. If an error occurs it is returned
// and the exons are not set.
func (t *NonCodingTranscript) SetExons(exons ...Exon) error {
	exons, err := buildExonsFor(t, exons...)
	if err != nil {
		return err
	}
	t.exons = exons
	return nil
}

// A CodingTranscript is a gene transcript that has coding potential. It can
// be located on any feat.Feature such as a gene or a chromosome. The concept
// of exons is tightly coupled with the CodingTranscript in the sense that
// the transcript borders are basically defined by the contained exons. By
// definition one of the exons must always start at position 0 relative to the
// transcript and this or another one must end at the transcript end. The
// former is asserted when exons are set and the latter is guaranteed by
// setting the transcript end at the end of the last exon.
type CodingTranscript struct {
	ID       string
	Loc      feat.Feature
	Offset   int
	Orient   feat.Orientation
	Desc     string
	CDSstart int
	CDSend   int
	exons    Exons
}

// Start returns the transcript start relative to Location.
func (t *CodingTranscript) Start() int { return t.Offset }

// End returns the transcript end relative to Location.
func (t *CodingTranscript) End() int { return t.Offset + t.exons.End() }

// Len returns the length of the transcript.
func (t *CodingTranscript) Len() int { return t.End() - t.Start() }

// Name returns the transcript name. Currently the same as the id.
func (t *CodingTranscript) Name() string { return t.ID }

// Description returns a description for the transcript.
func (t *CodingTranscript) Description() string { return t.Desc }

// Location returns the location of the transcript. Can be any feat.Feature
// such as a gene or a chromosome.
func (t *CodingTranscript) Location() feat.Feature { return t.Loc }

// Orientation returns the orientation of the transcript relative to Location.
func (t *CodingTranscript) Orientation() feat.Orientation {
	return t.Orient
}

// UTR5 returns a feat.Feature that corresponds to the 5'UTR of the
// transcript.
func (t *CodingTranscript) UTR5() feat.Feature {
	var start, end int
	ori, _ := feat.BaseOrientationOf(t)
	switch ori {
	case feat.Forward:
		start = 0
		end = t.CDSstart
	case feat.Reverse:
		start = t.CDSend
		end = t.Len()
	default:
		panic("gene: zero orientation for transcript")
	}
	return &TranscriptFeature{
		Transcript: t,
		Offset:     start,
		Length:     end - start,
		Orient:     feat.Forward,
	}
}

// CDS returns a feat.Feature that corresponds to the coding region of the
// transcript.
func (t *CodingTranscript) CDS() feat.Feature {
	return &TranscriptFeature{
		Transcript: t,
		Offset:     t.CDSstart,
		Length:     t.CDSend - t.CDSstart,
		Orient:     feat.Forward,
	}
}

// UTR3 returns a feat.Feature that corresponds to the 3'UTR of the
// transcript.
func (t *CodingTranscript) UTR3() feat.Feature {
	var start, end int
	ori, _ := feat.BaseOrientationOf(t)
	switch ori {
	case feat.Forward:
		start = t.CDSend
		end = t.Len()
	case feat.Reverse:
		start = 0
		end = t.CDSstart
	default:
		panic("gene: zero orientation for transcript")
	}
	return &TranscriptFeature{
		Transcript: t,
		Offset:     start,
		Length:     end - start,
		Orient:     feat.Forward,
	}
}

// UTR5start returns the start of the 5'UTR relative to the transcript.
func (t *CodingTranscript) UTR5start() int {
	return t.UTR5().Start()
}

// UTR5end returns the end of the 5'UTR relative to the transcript.
func (t *CodingTranscript) UTR5end() int {
	return t.UTR5().End()
}

// UTR3start returns the start of the 3'UTR relative to the transcript.
func (t *CodingTranscript) UTR3start() int {
	return t.UTR3().Start()
}

// UTR3end returns the end of the 3'UTR relative to the transcript.
func (t *CodingTranscript) UTR3end() int {
	return t.UTR3().End()
}

// Exons returns a typed slice with the transcript exons.
func (t *CodingTranscript) Exons() Exons { return t.exons }

// Introns returns a typed slice with the transcript introns.
func (t *CodingTranscript) Introns() Introns { return t.exons.Introns() }

// SetExons sets the transcript exons. Internally, it sorts exons by Start,
// verifies that their Location is the transcript, that they are not
// overlapping and that one has zero Start. If an error occurs it is returned
// and the exons are not set.
func (t *CodingTranscript) SetExons(exons ...Exon) error {
	newExons, err := buildExonsFor(t, exons...)
	if err != nil {
		return err
	}
	t.exons = newExons
	return nil
}

// TranscriptFeature defines a feature on a transcript.
type TranscriptFeature struct {
	Transcript Transcript
	Offset     int
	Length     int
	Orient     feat.Orientation
	Desc       string
}

// Start returns the feature start relative to Transcript.
func (t *TranscriptFeature) Start() int { return t.Offset }

// End returns the feature end relative to TranscriptLocation.
func (t *TranscriptFeature) End() int { return t.Offset + t.Length }

// Len returns the length of the feature.
func (t *TranscriptFeature) Len() int { return t.Length }

// Name returns an empty string.
func (t *TranscriptFeature) Name() string { return "" }

// Description returns the feature description.
func (t *TranscriptFeature) Description() string { return t.Desc }

// Location returns the Transcript.
func (t *TranscriptFeature) Location() feat.Feature { return t.Transcript }

// Orientation returns the orientation of the feature relative to Transcript.
func (t *TranscriptFeature) Orientation() feat.Orientation {
	return t.Orient
}

// Exons is a typed slice of Exon. It guarantees that exons are always sorted
// by Start, are all located on the same feature and are non overlapping.
type Exons []Exon

// SplicedLen returns the total length of the exons.
func (s Exons) SplicedLen() int {
	length := 0
	for _, e := range s {
		length += e.Len()
	}
	return length
}

// Add adds exons to the slice and safeguards the types contracts. It returns
// a new slice with the added exons. It checks for sorting, overlap, and
// location match.  If and error occurs it returns the old slice (without the
// new exons) and the error.
func (s Exons) Add(exons ...Exon) (Exons, error) {
	newSlice := append(s, exons...)
	sort.Sort(newSlice)
	for i, e := range newSlice {
		if i != 0 && e.Start() < newSlice[i-1].End() {
			return s, errors.New("exons overlap")
		}
		if i != 0 && e.Location() != newSlice[i-1].Location() {
			return s, errors.New("exons location differ")
		}

	}
	if s.Location() != nil && s.Location() != newSlice.Location() {
		return s, errors.New("new exons locations differ from old ones")
	}
	return newSlice, nil
}

// Location returns the common location of all the exons.
func (s Exons) Location() feat.Feature {
	if len(s) == 0 {
		return nil
	}
	return s[0].Location()
}

// Len returns the number of exons in the slice.
func (s Exons) Len() int {
	return len(s)
}

// Less returns whether the exon with index i should sort before
// the exon with index j.
func (s Exons) Less(i, j int) bool {
	return s[i].Start() < s[j].Start()
}

// Swap swaps the exons with indexes i and j.
func (s Exons) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// End returns the maximum End of all exons. Since exons are sorted and non
// overlapping this matches the End of the last exon in the slice.
func (s Exons) End() int {
	if len(s) == 0 {
		return 0
	}
	return s[len(s)-1].End()
}

// Start returns the minimum Start of all exons. Since exons are sorted and
// non overlapping this matches the Start of the first exon in the slice.
func (s Exons) Start() int {
	if len(s) == 0 {
		return 0
	}
	return s[0].Start()
}

// Introns returns a typed slice of Introns. Introns are built dynamically.
func (s Exons) Introns() Introns {
	var introns Introns
	if s.Len() < 2 {
		return introns
	}
	for i := 1; i < s.Len(); i++ {
		intron := Intron{
			Transcript: s[i].Transcript,
			Offset:     s[i-1].End(),
			Length:     s[i].Start() - s[i-1].End(),
		}
		introns = append(introns, intron)
	}
	return introns
}

// An Exon is the part of a transcript that remains present in the final
// mature RNA product after splicing.
type Exon struct {
	Transcript Transcript
	Offset     int
	Length     int
	Desc       string
}

// Start returns the start position of the exon relative to Transcript.
func (e Exon) Start() int { return e.Offset }

// End returns the end position of the exon relative to Transcript.
func (e Exon) End() int { return e.Offset + e.Length }

// Len returns the length of the exon.
func (e Exon) Len() int { return e.Length }

// Location returns the location of the exon - the transcript.
func (e Exon) Location() feat.Feature { return e.Transcript }

// Name returns an empty string.
func (e Exon) Name() string { return "" }

// Description returns a description for the exon.
func (e Exon) Description() string { return e.Desc }

// Orientation always returns Forward.
func (e Exon) Orientation() feat.Orientation {
	return feat.Forward
}

// Introns corresponds to a collection of introns.
type Introns []Intron

// An Intron is the part of a transcript that is removed during splicing
// and is not part of the final mature RNA product.
type Intron struct {
	Transcript Transcript
	Offset     int
	Length     int
	Desc       string
}

// Start returns the start position of the intron relative to Transcript.
func (i Intron) Start() int { return i.Offset }

// End returns the end position of the intron relative to Transcript.
func (i Intron) End() int { return i.Offset + i.Length }

// Len returns the length of the intron.
func (i Intron) Len() int { return i.Length }

// Location returns the location of the intron - the transcript.
func (i Intron) Location() feat.Feature { return i.Transcript }

// Name returns an empty string.
func (i Intron) Name() string { return "" }

// Description returns a description for the intron.
func (i Intron) Description() string { return i.Desc }

// Orientation always returns Forward.
func (i Intron) Orientation() feat.Orientation {
	return feat.Forward
}

// buildExonsFor is a helper function that will check if exons are compatible
// with a transcript and return a typed slice of exons. If it encounters an
// error or the exons are not compatible with the transcript it will return
// the error and a possibly partially filled slice. It is not safe to use the
// slice if the error is not nil.
func buildExonsFor(t Transcript, exons ...Exon) (Exons, error) {
	var newExons Exons
	newExons, err := newExons.Add(exons...)
	if err != nil {
		return newExons, err
	}
	if newExons.Location() != t {
		return newExons, errors.New("exon location is not the transcript")
	}
	if newExons.Start() != 0 {
		return newExons, errors.New("no exon with a zero start")
	}
	return newExons, nil
}
