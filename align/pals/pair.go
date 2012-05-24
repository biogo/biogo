package pals

// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

import (
	"fmt"
	"github.com/kortschak/biogo/align/pals/dp"
	"github.com/kortschak/biogo/bio"
	"github.com/kortschak/biogo/feat"
	"github.com/kortschak/biogo/seq"
	"strconv"
	"strings"
)

// A FeaturePair holds a pair of features with additional information relating the two.
type FeaturePair struct {
	A, B   *feat.Feature
	Score  int     // Score of alignment between features.
	Error  float64 // Identity difference between feature sequences.
	Strand int8    // Strand relationship: positive indicates same strand, negative indicates opposite strand.
}

// Convert coordinates in a packed sequence into a feat.Feature.
func featureOf(contigs *seq.Seq, from, to int, comp bool) (feature *feat.Feature, err error) {
	if comp {
		from, to = contigs.Len()-to, contigs.Len()-from
	}
	if from >= to {
		return nil, bio.NewError(fmt.Sprintf("%s: from > to", contigs.ID), 0, nil)
	}

	// DPHit coordinates sometimes over/underflow.
	// This is a lazy hack to work around it, should really figure
	// out what is going on.
	if from < 0 {
		from = 0
	}
	if to > contigs.Len() {
		to = contigs.Len()
	}

	// Take midpoint of segment -- lazy hack again, endpoints
	// sometimes under / overflow
	bin := (from + to) / (2 * binSize)
	binCount := (contigs.Len() + binSize - 1) / binSize

	if bin < 0 || bin >= binCount {
		return nil, bio.NewError(fmt.Sprintf("%s: bin %d out of range 0..%d", contigs.ID, bin, binCount-1), 0, nil)
	}

	contigIndex := contigs.Meta.(seqMap).binMap[bin]

	if contigIndex < 0 || contigIndex >= len(contigs.Meta.(seqMap).contigs) {
		return nil, bio.NewError(fmt.Sprintf("%s: contig index %d out of range 0..%d", contigs.ID, contigIndex, len(contigs.Meta.(seqMap).contigs)), 0, nil)
	}

	length := to - from

	if length < 0 {
		return nil, bio.NewError(fmt.Sprintf("%s: length < 0", contigs.ID), 0, nil)
	}

	contig := contigs.Meta.(seqMap).contigs[contigIndex]
	contigFrom := from - contig.from
	contigTo := contigFrom + length

	if contigFrom < 0 {
		contigFrom = 0
	}

	if contigTo > contig.seq.Len() {
		contigTo = contig.seq.Len()
	}

	return &feat.Feature{
		ID:    contig.seq.ID,
		Start: contigFrom,
		End:   contigTo,
	}, nil
}

// Convert a DPHit and two packed sequences into a FeaturePair.
func NewFeaturePair(target, query *seq.Seq, hit dp.DPHit, comp bool) (pair *FeaturePair, err error) {
	t, err := featureOf(target, hit.Abpos, hit.Aepos, false)
	if err != nil {
		return
	}
	q, err := featureOf(query, hit.Bbpos, hit.Bepos, comp)
	if err != nil {
		return
	}

	var strand int8
	if comp {
		strand = -1
	} else {
		strand = 1
	}

	return &FeaturePair{
		A:      t,
		B:      q,
		Score:  hit.Score,
		Error:  hit.Error,
		Strand: strand,
	}, nil
}

// Expand a feat.Feature containing a PALS-type feature attribute into a FeaturePair.
func ExpandFeature(f *feat.Feature) (pair *FeaturePair, err error) {
	if len(f.Attributes) < 7 || f.Attributes[:7] != "Target " {
		return nil, fmt.Errorf("pals: not a feature pair")
	}
	fields := strings.Fields(f.Attributes)
	if len(fields) != 6 {
		return nil, fmt.Errorf("pals: not a feature pair")
	}

	s, err := strconv.Atoi(fields[2])
	if err != nil {
		return
	}
	s--
	e, err := strconv.Atoi(fields[3][:len(fields[3])-1])
	if err != nil {
		return
	}

	maxe, err := strconv.ParseFloat(fields[5], 64)
	if err != nil {
		return
	}

	pair = &FeaturePair{
		A: f,
		B: &feat.Feature{
			ID:       fmt.Sprintf("%s:%d..%d", fields[1], s, e),
			Location: fields[1],
			Start:    s,
			End:      e},
		Score:  int(*f.Score),
		Error:  maxe,
		Strand: f.Strand,
	}
	f.Score = nil
	f.Attributes = ""
	f.Strand = 0

	return

}

// Invert returns a reversed copy of the feature pair such that A', B' = B, A.
func (self *FeaturePair) Invert() *FeaturePair {
	return &FeaturePair{
		A:      self.B,
		B:      self.A,
		Score:  self.Score,
		Error:  self.Error,
		Strand: self.Strand,
	}
}
