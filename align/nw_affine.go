// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package align

import (
	"code.google.com/p/biogo/alphabet"
	"code.google.com/p/biogo/feat"
)

// Setting debugNeedleAffine to true gives verbose scoring table output for the dynamic programming.
const debugNeedleAffine = false

// NWAffine is the affine gap penalty Needleman-Wunsch aligner type.
type NWAffine Affine

// Align aligns two sequences using the Needleman-Wunsch algorithm. It returns an alignment description
// or an error if the scoring matrix is not square, or the sequence data types or alphabets do not match.
func (a NWAffine) Align(reference, query AlphabetSlicer) ([]feat.Pair, error) {
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
}
