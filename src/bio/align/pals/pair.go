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
//
import (
	"bio"
	"bio/align/pals/dp"
	"bio/feat"
	"bio/seq"
	"fmt"
)

func FeatureOf(contigs *seq.Seq, from, to int, comp bool) (feature *feat.Feature, err error) {
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

	contigIndex := contigs.Meta.(SeqMap).binMap[bin]

	if contigIndex < 0 || contigIndex >= len(contigs.Meta.(SeqMap).contigs) {
		return nil, bio.NewError(fmt.Sprintf("%s: contig index %d our of range 0..%d", contigs.ID, contigIndex, len(contigs.Meta.(SeqMap).contigs)), 0, nil)
	}

	length := to - from + 1

	if length < 0 {
		return nil, bio.NewError(fmt.Sprintf("%s: length < 0", contigs.ID), 0, nil)
	}

	contig := contigs.Meta.(SeqMap).contigs[contigIndex]
	contigFrom := from - contig.from + 1
	contigTo := contigFrom + length - 1

	if contigFrom < 1 {
		contigFrom = 1
	}

	if contigTo > contig.seq.Len() {
		contigTo = contig.seq.Len()
	}

	return &feat.Feature{
		ID:    append([]byte{}, contig.seq.ID...),
		Start: contigFrom,
		End:   contigTo,
	}, nil
}

func FeaturePairOf(target, query *seq.Seq, hit dp.DPHit, comp bool) (pair *FeaturePair, err error) {
	var (
		t, q   *feat.Feature
		strand int8
	)

	if t, err = FeatureOf(target, hit.Abpos, hit.Aepos, false); err != nil {
		return
	}
	if q, err = FeatureOf(query, hit.Bbpos, hit.Bepos, comp); err != nil {
		return
	}

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
