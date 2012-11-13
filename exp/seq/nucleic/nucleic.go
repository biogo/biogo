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

// Package nucleic provides support for manipulation of single nucleic acid
// sequences with and without quality data.
//
// Two basic nucleic acid sequence types are provided, Seq and QSeq. Interfaces
// for more complex sequence types are also defined.
package nucleic

import (
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/exp/seq/sequtils"
)

// Strand stores nucleic acid sequence strand information.
type Strand int8

const (
	Minus Strand = iota - 1
	None
	Plus
)

// The default value for Qphred scores from non-quality sequences.
var DefaultQphred alphabet.Qphred = 40

func (s Strand) String() string {
	switch s {
	case Plus:
		return "(+)"
	case None:
		return "."
	case Minus:
		return "(-)"
	}
	return "undefined"
}

// Sequence describes the interface for nucleic acid sequences.
type Sequence interface {
	seq.Sequence
	seq.Reverser
	seq.Complementer
	sequtils.Slicer
	sequtils.Conformationer
	sequtils.ConformationSetter
	Nucleic() // No op function to tag nucleic type sequence data.
}

// Quality describes the interface for nucleic acid sequences with quality scores.
type Quality interface {
	Sequence
	EAt(seq.Position) float64      // Return the p(Error) for a specific position.
	SetE(seq.Position, float64)    // Set the p(Error) for a specific position.
	Encoding() alphabet.Encoding   // Return the score encoding scheme.
	SetEncoding(alphabet.Encoding) // Set the score encoding scheme.
	QEncode(seq.Position) byte     // Encode the quality at pos according the the encoding scheme.
}

// An AlignedAppenderis a multiple sequence alignment that can append letters.
type AlignedAppender interface {
	seq.Aligned
	AppendColumns(a ...[]alphabet.QLetter) (err error)
	AppendEach(a [][]alphabet.QLetter) (err error)
}

// Getter describes the interface for sets of sequences or aligned multiple sequences.
type Getter interface {
	Rows() int
	Get(i int) Sequence
}

// GetterAppender is a type for sets of sequences or aligned multiple sequences that can append letters to individual or grouped sequences.
type GetterAppender interface {
	Getter
	AppendEach(a [][]alphabet.QLetter) (err error)
}
