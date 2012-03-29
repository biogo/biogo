/*
Package seq provides the base for storage and manipulation of biological sequence information.

A variety of sequence types are provided by derived packages including nucleic and protein sequence
with and without quality scores. Multiple sequence data is also supported as unaligned sets and aligned sequences.

Quality scoring is based on Phred scores, although there is the capacity to interconvert between Phred and
Solexa scores and a Solexa quality package is provide, though not integrated.
*/
package seq

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
	"github.com/kortschak/biogo/feat"
)

const (
	Start = 1 << iota
	End
)

var emptyString = ""

type Filter func(Sequence, alphabet.Letter) alphabet.Letter

type Stringify func(Polymer) string

// A Position holds a sequence position for all sequence types.
type Position struct {
	Pos int // The index of a letter within a sequence.
	Ind int // The specific sequence within a multiple sequence. Ignored by single sequence types.
}

// Polymer is the base type for sequences.
type Polymer interface {
	Raw() interface{} // Return a pointer the underlying polymer data unless the data is a pointer type, then return the data.
	Feature
	Offset(int) // Set the offset of the polymer.
	Circular
	Reverser
	Composer
	Stitcher
	Truncator
	String() string
}

// A Feature is an entity with a name, description, size and location.
// At some point this will move into new Feature package.
type Feature interface {
	Name() *string        // Return the ID of the polymer.
	Description() *string // Return the description of the polymer.
	Location() *string    // Return the location of the polymer.
	Start() int           // Return the start position of the polymer.
	End() int             // Return the end position of the polymer.
	Len() int             // Return the length of the polymer.
}

// The Circular interface includes the methods required for sequences that can be circular.
type Circular interface {
	IsCircular() bool // Return whether the sequence is circular.
	Circular(bool)    // Specify whether the sequence is circular.
}

// An Appender can append letters.
type Appender interface {
	Append(...alphabet.QLetter) error
}

// A Letterer gets and sets letters.
type Letterer interface {
	At(Position) alphabet.QLetter   // Return the letter at a specific position.
	Set(Position, alphabet.QLetter) // Set the letter at a specific position.
}

// A Scorer is a type that provides Phred-based scoring information.
type Scorer interface {
	EAt(Position) float64   // Return the p(Error) for a specific position.
	SetE(Position, float64) // Set the p(Error) for a specific position.
	Encoder
}

// A Sequence is a polymer whose elements are letters.
type Sequence interface {
	Letterer
	Counter
	Alphabet() alphabet.Alphabet             // Return the Alphabet being used.
	Subseq(start, end int) (Sequence, error) // Return a subsequence of the polymer from start to end or an error.
	Copy() Sequence                          // Return a copy of the Sequence.
}

// A Counter is a polymer type returns the number of polymers it contains.
type Counter interface {
	Count() int // Return the number of sub-polymers recursively.
}

// An Aligner is a polymer type whose sub-polymers are aligned.
type Aligner interface {
	Column(pos int) []alphabet.Letter    // Return the letter elements for all sub-polymers at a specific index.
	ColumnQL(pos int) []alphabet.QLetter // Return the quality letter elements for all sub-polymers at a specific index.
}

// An Encoder is a type that can encode Phred-based scoring information.
type Encoder interface {
	Encoding() alphabet.Encoding    // Return the score encoding scheme.
	SetEncoding(alphabet.Encoding)  // Set the score encoding scheme.
	QEncode(pos Position) byte      // Encode the quality at pos according the the encoding scheme.
	QDecode(l byte) alphabet.Qphred // Decode the l into a Qphred according the the encoding scheme.
}

// A Quality is a Polymer whose elements are Phred scores.
type Quality interface {
	Scorer
	At(Position) alphabet.Qphred
	Set(Position, alphabet.Qphred)
	Subseq(start, end int) (Quality, error) // Return a subsequence of the polymer from start to end or an error.
	Copy() Quality                          // Return a copy of the Quality.
}

// A Complementer type can be reverse complemented.
type Complementer interface {
	RevComp() // Reverse complement the polymer.
}

// A Reverse type can Reverse itself.
type Reverser interface {
	Reverse() // Reverse the order of elements in the polymer.
}

// A Truncator type can truncate itself.
type Truncator interface {
	Truncate(start, end int) error // Truncate the polymer from start to end, returning any error.
}

// A Joiner would be a type that can join another polymer. This interface is advisory.
// All polymers should satisfy the intent of this interface taking their own type only.
type Joiner interface {
	Join(Joiner, int) error // Join another polymer at end specified, returning any error.
}

// A Stitcher can join together sequentially ordered disjunct segments of the polymer.
type Stitcher interface {
	Stitch(f feat.FeatureSet) error // Join segments described by f, returning any error.
}

// A Compose can join together segments of the polymer in any order, potentially repeatedly.
type Composer interface {
	Compose(f feat.FeatureSet) error // Join segments described by f, returning any error.
}
