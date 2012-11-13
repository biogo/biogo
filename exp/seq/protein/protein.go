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

// Package protein provides support for manipulation of single protein
// sequences with and without quality data.
//
// Two basic protein sequence types are provided, Seq and QSeq. Interfaces
// for more complex sequence types are also defined.
package protein

import (
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/exp/seq/sequtils"
)

// The default value for Qphred scores from non-quality sequences.
var DefaultQphred alphabet.Qphred = 40

// Sequence describes the interface for protein sequences.
type Sequence interface {
	seq.Sequence
	seq.Reverser
	sequtils.Slicer
	Protein() // No op function to tag protein type sequence data.
}

// Quality describes the interface for protein sequences with quality scores.
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
