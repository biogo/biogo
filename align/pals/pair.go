// Copyright ©2011-2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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

package pals

import (
	"code.google.com/p/biogo/align/pals/dp"
	"code.google.com/p/biogo/exp/feat"
	gff "code.google.com/p/biogo/feat"
	"fmt"
	"strconv"
	"strings"
)

// A Pair holds a pair of features with additional information relating the two.
type Pair struct {
	A, B   feat.Feature
	Score  int     // Score of alignment between features.
	Error  float64 // Identity difference between feature sequences.
	Strand int8    // Strand relationship: positive indicates same strand, negative indicates opposite strand.
}

func (fp *Pair) String() string {
	return fmt.Sprintf("%s/%s[%d,%d)--%s/%s[%d,%d)",
		fp.A.Location().Location().Name(), fp.A.Name(), fp.A.Start(), fp.A.End(),
		fp.B.Location().Location().Name(), fp.B.Name(), fp.B.Start(), fp.B.End(),
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

	var strand int8
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

// ExpandFeature converts an old-style *feat.Feature (package temporarily renamed to gff for collision avoidance) containing a PALS-type feature attribute
// into a Pair.
func ExpandFeature(f *gff.Feature) (*Pair, error) {
	if len(f.Attributes) < 7 || f.Attributes[:7] != "Target " {
		return nil, fmt.Errorf("pals: not a feature pair")
	}
	fields := strings.Fields(f.Attributes)
	if len(fields) != 6 {
		return nil, fmt.Errorf("pals: not a feature pair")
	}

	s, err := strconv.Atoi(fields[2])
	if err != nil {
		return nil, err
	}
	s--
	e, err := strconv.Atoi(fields[3][:len(fields[3])-1])
	if err != nil {
		return nil, err
	}

	maxe, err := strconv.ParseFloat(fields[5], 64)
	if err != nil {
		return nil, err
	}

	fp := &Pair{
		A: &Feature{
			ID:   f.ID,
			Loc:  Contig(f.Location),
			From: f.Start,
			To:   f.End,
		},
		B: &Feature{
			ID:   fmt.Sprintf("%s:%d..%d", fields[1], s, e),
			Loc:  Contig(fields[1]),
			From: s,
			To:   e},
		Score:  int(*f.Score),
		Error:  maxe,
		Strand: f.Strand,
	}
	f.Score = nil
	f.Attributes = ""
	f.Strand = 0

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
