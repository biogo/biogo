// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package seq provides the base for storage and manipulation of biological sequence information.
// 
// A variety of sequence types are provided by derived packages including linear and protein sequence
// with and without quality scores. Multiple sequence data is also supported as unaligned sets and aligned sequences.
// 
// Quality scoring is based on Phred scores, although there is the capacity to interconvert between Phred and
// Solexa scores and a Solexa quality package is provide, though not integrated.
package seq

import (
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/feat"
	"fmt"
	"math"
)

const (
	Start = 1 << iota
	End
)

var (
	// The default value for Qphred scores from non-quality sequences.
	DefaultQphred alphabet.Qphred = 40
	// The default encoding for Qphred scores from non-quality sequences.
	DefaultEncoding alphabet.Encoding = alphabet.Sanger
)

type Alphabeter interface {
	Alphabet() alphabet.Alphabet
}

// A QFilter returns a letter based on an alphabet, quality letter and quality threshold.
type QFilter func(a alphabet.Alphabet, thresh alphabet.Qphred, ql alphabet.QLetter) alphabet.Letter

var (
	// AmbigFilter is a QFilter function that returns the given alphabet's ambiguous position
	// letter for quality letters with a quality score below the specified threshold.
	AmbigFilter QFilter = func(a alphabet.Alphabet, thresh alphabet.Qphred, l alphabet.QLetter) alphabet.Letter {
		if l.L == a.Gap() || l.Q >= thresh {
			return l.L
		}
		return a.Ambiguous()
	}

	// CaseFilter is a QFilter function that returns a lower case letter for quality letters
	// with a quality score below the specified threshold and upper case equal to or above the threshold.
	CaseFilter QFilter = func(a alphabet.Alphabet, thresh alphabet.Qphred, l alphabet.QLetter) alphabet.Letter {
		switch {
		case l.L == a.Gap():
			return l.L
		case l.Q >= thresh:
			return l.L &^ ('a' - 'A')
		}
		return l.L | ('a' - 'A')
	}
)

// A Sequence is a feature that stores sequence information.
type Sequence interface {
	Feature
	At(int) alphabet.QLetter      // Return the letter at a specific position.
	Set(int, alphabet.QLetter)    // Set the letter at a specific position.
	Alphabet() alphabet.Alphabet  // Return the Alphabet being used.
	RevComp()                     // Reverse complement the sequence.
	Reverse()                     // Reverse the order of elements in the sequence.
	New() Sequence                // Return a pointer to the zero value of the concrete type.
	Clone() Sequence              // Return a copy of the Sequence.
	CloneAnnotation() *Annotation // Return a copy of the sequence's annotation.
	Slicer
	Conformationer
	ConformationSetter
	fmt.Formatter
}

// A Feature describes the basis for sequence features.
type Feature interface {
	feat.Feature
	feat.Offsetter
}

// A Conformationer can give information regarding the sequence's conformation. For the
// purposes of sequtils, types that are not a Conformationer are treated as linear.
type Conformationer interface {
	Conformation() feat.Conformation
}

// A ConformationSetter can set its sequence conformation.
type ConformationSetter interface {
	SetConformation(feat.Conformation)
}

// A Slicer returns and sets a Slice. 
type Slicer interface {
	Slice() alphabet.Slice
	SetSlice(alphabet.Slice)
}

// A Scorer is a sequence type that provides Phred-based scoring information.
type Scorer interface {
	EAt(int) float64               // Return the p(Error) for a specific position.
	SetE(int, float64)             // Set the p(Error) for a specific position.
	Encoding() alphabet.Encoding   // Return the score encoding scheme.
	SetEncoding(alphabet.Encoding) // Set the score encoding scheme.
	QEncode(int) byte              // Encode the quality at the specified position according the the encoding scheme.
}

// A Quality is a feature whose elements are Phred scores.
type Quality interface {
	Scorer
	Copy() Quality // Return a copy of the Quality.
}

// Rower describes the interface for sets of sequences or aligned multiple sequences.
type Rower interface {
	Rows() int
	Row(i int) Sequence
}

// RowAppender is a type for sets of sequences or aligned multiple sequences that can append letters to individual or grouped sequences.
type RowAppender interface {
	Rower
	AppendEach(a [][]alphabet.QLetter) (err error)
}

// An Appender can append letters.
type Appender interface {
	AppendLetters(...alphabet.Letter) error
	AppendQLetters(...alphabet.QLetter) error
}

// Aligned describes the interface for aligned multiple sequences.
type Aligned interface {
	Start() int
	End() int
	Rows() int
	Column(pos int, fill bool) []alphabet.Letter
	ColumnQL(pos int, fill bool) []alphabet.QLetter
}

// An AlignedAppenderis a multiple sequence alignment that can append letters.
type AlignedAppender interface {
	Aligned
	AppendColumns(a ...[]alphabet.QLetter) (err error)
	AppendEach(a [][]alphabet.QLetter) (err error)
}

// ConsenseFunc is a function type that returns the consensus letter for a column of an alignment.
type ConsenseFunc func(a Aligned, alpha alphabet.Alphabet, pos int, fill bool) alphabet.QLetter

var (
	// The default ConsenseFunc function.
	DefaultConsensus ConsenseFunc = func(a Aligned, alpha alphabet.Alphabet, pos int, fill bool) alphabet.QLetter {
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

	// A default ConsenseFunc function that takes letter quality into account.
	// http://staden.sourceforge.net/manual/gap4_unix_120.html
	DefaultQConsensus ConsenseFunc = func(a Aligned, alpha alphabet.Alphabet, pos int, fill bool) alphabet.QLetter {
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
)

// Tolerance on float comparison for DefaultQConsensus.
var FloatTolerance float64 = 1e-10
