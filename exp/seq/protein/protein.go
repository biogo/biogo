// Package protein provides support for manipulation of single protein
// sequences with and without quality data.
//
// Two basic protein sequence types are provided, Seq and QSeq. Interfaces
// for more complex sequence types are also defined.
package protein

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
	"github.com/kortschak/biogo/exp/alphabet"
	"github.com/kortschak/biogo/exp/seq"
	"math"
)

var emptyString = ""

// The default value for Qphred scores from non-quality sequences.
var DefaultQphred alphabet.Qphred = 40

// Sequence describes the interface for protein sequences.
type Sequence interface {
	seq.Polymer
	seq.Sequence
	Protein() // No op function to tag protein type sequence data.
}

// Quality describes the interface for protein sequences with quality scores.
type Quality interface {
	seq.Polymer
	seq.Sequence
	seq.Scorer
	Protein()
}

// Aligned describes the interface for aligned multiple sequences.
type Aligned interface {
	Sequence
	Column(pos int, fill bool) []alphabet.Letter
	ColumnQL(pos int, fill bool) []alphabet.QLetter
	Consensus(fill bool) *QSeq
}

// An AlignedAppenderis a multiple sequence alignment that can append letters.
type AlignedAppender interface {
	Aligned
	AppendColumns(a ...[]alphabet.QLetter) (err error)
	AppendEach(a [][]alphabet.QLetter) (err error)
}

// Extracter describes the interface for column based aligned multiple sequences.
type Extracter interface {
	Sequence
	Extract(i int) Sequence
}

// Getter describes the interface for sets of sequences or aligned multiple sequences.
type Getter interface {
	Sequence
	Get(i int) Sequence
}

// GetterAppender is a type for sets of sequences or aligned multiple sequences that can append letters to individual or grouped seqeunces.
type GetterAppender interface {
	Getter
	AppendEach(a [][]alphabet.QLetter) (err error)
}

// Consensifyer is a function type that returns the consensus letter for a column of an alignment.
type Consensifyer func(a Aligned, pos int, fill bool) alphabet.QLetter

// The default Consensifyer function.
var Consensify = func(a Aligned, pos int, fill bool) alphabet.QLetter {
	alpha := a.Alphabet()
	w := make([]int, alpha.Len())
	c := a.Column(pos, fill)

	for _, l := range c {
		if alpha.IsValid(l) {
			w[alpha.IndexOf(l)]++
		}
	}

	var max, maxi int
	for i, v := range w {
		if v > max {
			max, maxi = v, i
		}
	}

	return alphabet.QLetter{
		L: alpha.Letter(maxi),
		Q: alphabet.Ephred(1 - (float64(max) / float64(len(c)))),
	}
}

// Tolerance on float comparison for QConsensify
var FloatTolerance float64 = 1e-10

// A default Consensifyer function that takes letter quality into account.
// http://staden.sourceforge.net/manual/gap4_unix_120.html
var QConsensify = func(a Aligned, pos int, fill bool) alphabet.QLetter {
	alpha := a.Alphabet()

	w := make([]float64, alpha.Len())
	for i := range w {
		w[i] = 1
	}

	others := float64(alpha.Len() - 1)
	c := a.ColumnQL(pos, fill)
	for _, l := range c {
		if alpha.IsValid(l.L) {
			i, alt := alpha.IndexOf(l.L), l.Q.ProbE()
			p := (1 - alt)
			alt /= others
			for b := range w {
				if i == b {
					w[b] *= p
				} else {
					w[b] *= alt
				}
			}
		}
	}

	var (
		max         = 0.
		sum         float64
		best, count int
	)
	for _, p := range w {
		sum += p
	}
	for i, v := range w {
		if v /= sum; v > max {
			max, best = v, i
			count = 0
		}
		if v == max || math.Abs(max-v) < FloatTolerance {
			count++
		}
	}

	if count > 1 {
		return alphabet.QLetter{
			L: alpha.Ambiguous(),
			Q: 0,
		}
	}

	return alphabet.QLetter{
		L: alpha.Letter(best),
		Q: alphabet.Ephred(1 - max),
	}
}
