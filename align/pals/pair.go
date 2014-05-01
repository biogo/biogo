// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pals

import (
	"code.google.com/p/biogo/align/pals/dp"
	"code.google.com/p/biogo/io/featio/gff"
	"code.google.com/p/biogo/seq"

	"fmt"
	"strconv"
	"strings"
)

// A Pair holds a pair of features with additional information relating the two.
type Pair struct {
	A, B   *Feature
	Score  int        // Score of alignment between features.
	Error  float64    // Identity difference between feature sequences.
	Strand seq.Strand // Strand relationship: seq.Plus indicates same strand, seq.Minus indicates opposite strand.
}

func (fp *Pair) String() string {
	if fp == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s/%s[%d,%d)--%s/%s[%d,%d)",
		fp.A.Location().Name(), fp.A.Name(), fp.A.Start(), fp.A.End(),
		fp.B.Location().Name(), fp.B.Name(), fp.B.Start(), fp.B.End(),
	)
}

// NewPair converts a DPHit and two packed sequences into a Pair.
func NewPair(target, query *Packed, hit dp.DPHit, comp bool) (*Pair, error) {
	t, err := target.feature(hit.Abpos, hit.Aepos, false)
	if err != nil {
		return nil, err
	}
	q, err := query.feature(hit.Bbpos, hit.Bepos, comp)
	if err != nil {
		return nil, err
	}

	var strand seq.Strand
	if comp {
		strand = -1
	} else {
		strand = 1
	}

	return &Pair{
		A:      t,
		B:      q,
		Score:  hit.Score,
		Error:  hit.Error,
		Strand: strand,
	}, nil
}

// ExpandFeature converts a *gff.Feature containing PALS-type feature attributes into a Pair.
func ExpandFeature(f *gff.Feature) (*Pair, error) {
	targ := f.FeatAttributes.Get("Target")
	if targ == "" {
		return nil, fmt.Errorf("pals: not a feature pair")
	}
	fields := strings.Fields(targ)
	if len(fields) != 3 {
		return nil, fmt.Errorf("pals: not a feature pair")
	}

	s, err := strconv.Atoi(fields[1])
	if err != nil {
		return nil, err
	}
	s--
	e, err := strconv.Atoi(fields[2])
	if err != nil {
		return nil, err
	}

	maxe, err := strconv.ParseFloat(f.FeatAttributes.Get("maxe"), 64)
	if err != nil {
		return nil, err
	}

	fp := &Pair{
		A: &Feature{
			ID:   fmt.Sprintf("%s:%d..%d", f.SeqName, f.FeatStart, f.FeatEnd),
			Loc:  Contig(f.SeqName),
			From: f.FeatStart,
			To:   f.FeatEnd,
		},
		B: &Feature{
			ID:   fmt.Sprintf("%s:%d..%s", fields[0], s, fields[2]),
			Loc:  Contig(fields[0]),
			From: s,
			To:   e,
		},
		Score:  int(*f.FeatScore),
		Error:  maxe,
		Strand: f.FeatStrand,
	}
	f.FeatScore = nil
	f.FeatAttributes = nil
	f.FeatStrand = seq.None

	return fp, nil
}

// Invert returns a reversed copy of the feature pair such that A', B' = B, A.
func (fp *Pair) Invert() *Pair {
	return &Pair{
		A:      fp.B,
		B:      fp.A,
		Score:  fp.Score,
		Error:  fp.Error,
		Strand: fp.Strand,
	}
}
