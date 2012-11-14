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

package align

import (
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/feat"
)

// SW is the Smith-Waterman aligner type.
// Matrix is a square scoring matrix with the last column and last row specifying gap penalties.
// Currently gap opening is not considered.
type SW Linear

// Align aligns two sequences using the Smith-Waterman algorithm. It returns an alignment description
// or an error if the scoring matrix is not square, or the sequence data types or alphabets do not match.
func (a SW) Align(reference, query AlphabetSlicer) ([]feat.Pair, error) {
	alpha := reference.Alphabet()
	if alpha == nil {
		return nil, ErrNoAlphabet
	}
	if alpha != query.Alphabet() {
		return nil, ErrMismatchedAlphabets
	}
	switch rSeq := reference.Slice().(type) {
	case alphabet.Letters:
		qSeq, ok := query.Slice().(alphabet.Letters)
		if !ok {
			return nil, ErrMismatchedTypes
		}
		return a.alignLetters(rSeq, qSeq, alpha)
	case alphabet.QLetters:
		qSeq, ok := query.Slice().(alphabet.QLetters)
		if !ok {
			return nil, ErrMismatchedTypes
		}
		return a.alignQLetters(rSeq, qSeq, alpha)
	default:
		return nil, ErrTypeNotHandled
	}

	panic("cannot reach")
}
