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

// Package seq provides the base for storage and manipulation of biological sequence information.
// 
// A variety of sequence types are provided by derived packages including nucleic and protein sequence
// with and without quality scores. Multiple sequence data is also supported as unaligned sets and aligned sequences.
// 
// Quality scoring is based on Phred scores, although there is the capacity to interconvert between Phred and
// Solexa scores and a Solexa quality package is provide, though not integrated.
package seq

import (
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/feat"
	"math"
)

const (
	Start = 1 << iota
	End
)

type Alphabeter interface {
	Alphabet() alphabet.Alphabet
}

type Filter func(Alphabeter, alphabet.Letter) alphabet.Letter

// A Position holds a sequence position for all sequence types.
type Position struct {
	Col int // The index of a letter within a sequence.
	Row int // The specific sequence within a multiple sequence.
}

// An Appender can append letters.
type Appender interface {
	AppendLetters(...alphabet.Letter) error
	AppendQLetters(...alphabet.QLetter) error
}

// A Feature describes the basis for sequence features.
type Feature interface {
	feat.Feature
	feat.Offsetter
}

// A Sequence is a feature that stores sequence information.
type Sequence interface {
	Feature
	At(Position) alphabet.QLetter   // Return the letter at a specific position.
	Set(Position, alphabet.QLetter) // Set the letter at a specific position.
	Alphabet() alphabet.Alphabet    // Return the Alphabet being used.
	New() Sequence                  // Return a pointer to the zero value of the concrete type.
	Copy() Sequence                 // Return a copy of the Sequence.
}

// A Scorer is a sequence type that provides Phred-based scoring information.
type Scorer interface {
	EAt(Position) float64          // Return the p(Error) for a specific position.
	SetE(Position, float64)        // Set the p(Error) for a specific position.
	Encoding() alphabet.Encoding   // Return the score encoding scheme.
	SetEncoding(alphabet.Encoding) // Set the score encoding scheme.
	QEncode(pos Position) byte     // Encode the quality at pos according the the encoding scheme.
}

// A Quality is a feature whose elements are Phred scores.
type Quality interface {
	Scorer
	Copy() Quality // Return a copy of the Quality.
}

// A Complementer type can be reverse complemented.
type Complementer interface {
	RevComp() // Reverse complement the sequence.
}

// A Reverser type can be reversed.
type Reverser interface {
	Reverse() // Reverse the order of elements in the sequence.
}

// Aligned describes the interface for aligned multiple sequences.
type Aligned interface {
	Start() int
	End() int
	Rows() int
	Column(pos int, fill bool) []alphabet.Letter
	ColumnQL(pos int, fill bool) []alphabet.QLetter
}

// ConsenseFunc is a function type that returns the consensus letter for a column of an alignment.
type ConsenseFunc func(a Aligned, alpha alphabet.Alphabet, pos int, fill bool) alphabet.QLetter

// The default ConsenseFunc function.
var DefaultConsensus = func(a Aligned, alpha alphabet.Alphabet, pos int, fill bool) alphabet.QLetter {
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

// Tolerance on float comparison for QConsensify.
var FloatTolerance float64 = 1e-10

// A default ConsenseFunc function that takes letter quality into account.
// http://staden.sourceforge.net/manual/gap4_unix_120.html
var DefaultQConsensus = func(a Aligned, alpha alphabet.Alphabet, pos int, fill bool) alphabet.QLetter {
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
