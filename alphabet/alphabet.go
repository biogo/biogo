// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package alphabet describes biological sequence letters, including quality scores.
package alphabet

import (
	"github.com/biogo/biogo/feat"

	"errors"
	"fmt"
	"strings"
	"unicode"
)

const (
	CaseSensitive = true
)

// Package alphabet provides default Alphabets for DNA, RNA and Protein. These
// alphabets are case insensitive and for the non-redundant nucleic acid alphabets
// satisfy the condition that the index of a letter is equal to the bitwise-complement
// of the index of the base-complement, modulo 4.
var (
	DNA = MustComplement(NewComplementor(
		"acgt",
		feat.DNA,
		MustPair(NewPairing("acgtnxACGTNX-", "tgcanxTGCANX-")),
		'-', 'n',
		!CaseSensitive,
	))

	DNAgapped = MustComplement(NewComplementor(
		"-acgt",
		feat.DNA,
		MustPair(NewPairing("acgtnxACGTNX-", "tgcanxTGCANX-")),
		'-', 'n',
		!CaseSensitive,
	))

	DNAredundant = MustComplement(NewComplementor(
		"-acmgrsvtwyhkdbn",
		feat.DNA,
		MustPair(NewPairing("acmgrsvtwyhkdbnxACMGRSVTWYHKDBNX-", "tgkcysbawrdmhvnxTGKCYSBAWRDMHVNX-")),
		'-', 'n',
		!CaseSensitive,
	))

	RNA = MustComplement(NewComplementor(
		"acgu",
		feat.RNA,
		MustPair(NewPairing("acgunxACGUNX-", "ugcanxUGCANX-")),
		'-', 'n',
		!CaseSensitive,
	))

	RNAgapped = MustComplement(NewComplementor(
		"-acgu",
		feat.RNA,
		MustPair(NewPairing("acgunxACGUNX-", "ugcanxUGCANX-")),
		'-', 'n',
		!CaseSensitive,
	))

	RNAredundant = MustComplement(NewComplementor(
		"-acmgrsvuwyhkdbn",
		feat.RNA,
		MustPair(NewPairing("acmgrsvuwyhkdbnxACMGRSVUWYHKDBNX-", "ugkcysbawrdmhvnxUGKCYSBAWRDMHVNX-")),
		'-', 'n',
		!CaseSensitive,
	))

	Protein = Must(NewAlphabet(
		"-abcdefghijklmnpqrstvwxyz*",
		feat.Protein,
		'-', 'x',
		!CaseSensitive,
	))
)

// Must is a helper that wraps a call to a function returning (Alphabet, error)
// and panics if the error is non-nil. It is intended for use in variable
// initializations.
func Must(a Alphabet, err error) Alphabet {
	if err != nil {
		panic(err)
	}
	return a
}

// MustComplement is a helper that wraps a call to a function returning (Complementor, error)
// and panics if the error is non-nil. It is intended for use in variable
// initializations.
func MustComplement(c Complementor, err error) Complementor {
	if err != nil {
		panic(err)
	}
	return c
}

// MustPair is a helper that wraps a call to a function returning (*Pairing, error)
// and panics if the error is non-nil. It is intended for use in variable
// initializations.
func MustPair(p *Pairing, err error) *Pairing {
	if err != nil {
		panic(err)
	}
	return p
}

// Type Index is a pointer to an index table.
type Index *[256]int

// An Alphabet describes valid single character letters within a sequence.
type Alphabet interface {
	// IsValid reports whether a letter conforms to the alphabet.
	IsValid(Letter) bool

	// AllValid reports whether a slice of bytes conforms to the alphabet.
	// It returns the index of the first invalid byte,
	// or a negative int if all bytes are valid.
	AllValid([]Letter) (ok bool, pos int)

	// AllValidQLetter reports whether a slice of bytes conforms to the alphabet.
	// It returns the index of the first invalid byte,
	// or a negative int if all bytes are valid.
	AllValidQLetter([]QLetter) (ok bool, pos int)

	// Len returns the number of distinct valid letters in the alphabet.
	Len() int

	// IndexOf returns the index of a given letter.
	IndexOf(Letter) int

	// Letter returns the letter corresponding to the given index.
	Letter(int) Letter

	// LetterIndex returns a pointer to the internal array specifying
	// letter to index conversion. The returned index should not be altered.
	LetterIndex() Index

	// Letters returns a string of letters conforming to the alphabet in index
	// order. In case insensitive alphabets, both cases are presented.
	Letters() string

	// ValidLetters returns a slice of the internal []bool indicating valid
	// letters. The returned slice should not be altered.
	ValidLetters() []bool

	// Gap returns the gap character used by the alphabet.
	Gap() Letter

	// Ambiguous returns the character representing an ambiguous letter.
	Ambiguous() Letter

	// Moltype returns the molecule type of the alphabet.
	Moltype() feat.Moltype

	// IsCased returns whether the alphabet is case sensitive.
	IsCased() bool
}

// A Complementor is an Alphabet that describes the complementation relationships
// between letters.
type Complementor interface {
	Alphabet
	Complement(Letter) (Letter, bool)
	ComplementTable() []Letter
}

// Single letter alphabet type.
type alpha struct {
	letters        string
	length         int
	valid          [256]bool
	index          [256]int
	gap, ambiguous Letter
	caseSensitive  bool
	molType        feat.Moltype
}

func newAlphabet(letters string, molType feat.Moltype, gap, ambiguous Letter, caseSensitive bool) (*alpha, error) {
	if strings.IndexFunc(letters, func(r rune) bool { return r < 0 || r > unicode.MaxASCII }) > -1 {
		return nil, errors.New("alphabet: letters contains non-ASCII rune")
	}

	a := &alpha{
		length:        len(letters),
		gap:           gap,
		ambiguous:     ambiguous,
		caseSensitive: caseSensitive,
		molType:       molType,
	}

	for i := range a.index {
		a.index[i] = -1
	}

	if caseSensitive {
		a.letters = letters
		for i, l := range a.letters {
			a.valid[l] = true
			a.index[l] = i
		}
		return a, nil
	}

	a.letters = strings.ToLower(letters) + strings.ToUpper(letters)
	for i, l := range a.letters[:len(letters)] {
		a.valid[l] = true
		a.index[l] = i
	}
	for i, l := range a.letters[len(letters):] {
		a.valid[l] = true
		a.index[l] = a.index[a.letters[i]]
	}

	return a, nil
}

func (a *alpha) Moltype() feat.Moltype { return a.molType }
func (a *alpha) Len() int              { return a.length }
func (a *alpha) IsCased() bool         { return a.caseSensitive }
func (a *alpha) Gap() Letter           { return a.gap }
func (a *alpha) Ambiguous() Letter     { return a.ambiguous }
func (a *alpha) AllValidQLetter(n []QLetter) (bool, int) {
	for i, v := range n {
		if !a.valid[v.L] {
			return false, i
		}
	}

	return true, -1
}
func (a *alpha) AllValid(n []Letter) (bool, int) {
	for i, v := range n {
		if !a.valid[v] {
			return false, i
		}
	}

	return true, -1
}
func (a *alpha) IsValid(n Letter) bool {
	return a.valid[n]
}
func (a *alpha) Letter(i int) Letter {
	return Letter(a.letters[:a.length][i])
}
func (a *alpha) IndexOf(n Letter) int {
	return a.index[n]
}
func (a *alpha) ValidLetters() []bool { return a.valid[:] }
func (a *alpha) LetterIndex() Index   { return Index(&a.index) }
func (a *alpha) Letters() string      { return a.letters }

// A Pairing provides a lookup table between a letter and its complement.
type Pairing struct {
	pair        []Letter
	ok          []bool
	complements [256]Letter
}

// NewPairing create a new Pairing from a pair of strings. Pairing definitions must be
// a bijection and must contain only ASCII characters.
func NewPairing(s, c string) (*Pairing, error) {
	if len(s) != len(c) {
		return nil, errors.New("alphabet: length of pairing definitions do not match")
	}

	p := &Pairing{
		pair: make([]Letter, 256),
		ok:   make([]bool, 256),
	}

	for i := range p.pair {
		p.pair[i] = Letter(i)
	}

	cr := []rune(c)
	for i, v := range s {
		if v < 0 || cr[i] < 0 || v > unicode.MaxASCII || cr[i] > unicode.MaxASCII {
			return nil, errors.New("alphabet: pairing definition contains non-ASCII rune")
		}
		p.pair[v] = Letter(cr[i])
		p.ok[v] = true
	}
	for i, l := range s {
		if Letter(l) != p.pair[p.pair[l]] {
			return nil, errors.New("alphabet: pairing definition is not a bijection")
		}
		if Letter(c[i]) != p.pair[p.pair[c[i]]] {
			return nil, errors.New("alphabet: pairing definition is not a bijection")
		}
	}
	copy(p.complements[:], p.pair)
	for i, ok := range p.ok {
		if !ok {
			p.complements[i] |= unicode.MaxASCII + 1
		}
	}
	return p, nil
}

// Returns the complement of a letter and true if the complement is a valid letter otherwise unchanged and false.
func (p *Pairing) Complement(l Letter) (c Letter, ok bool) { return p.pair[l], p.ok[l] }

// Returns a complementation table based on the internal representation. Invalid pairs hold a value outside the ASCII range.
// The caller must not modify the returned table.
func (p *Pairing) ComplementTable() []Letter {
	return p.complements[:]
}

type nucleic struct {
	*alpha
	*Pairing
}

// NewComplementor returns a complementing alphabet. The Complement table is checked for
// validity and an error is returned if an invalid complement pair is found. Pairings
// that result in no change but would otherwise be invalid are allowed. Letter parameter
// handling is the same as for NewAlphabet.
func NewComplementor(letters string, molType feat.Moltype, pairs *Pairing, gap, ambiguous Letter, caseSensitive bool) (Complementor, error) {
	a, err := newAlphabet(letters, molType, gap, ambiguous, caseSensitive)
	if err != nil {
		return nil, err
	}

	if pairs != nil {
		for i, v := range pairs.pair {
			if !(pairs.ok[i] || Letter(i&unicode.MaxASCII) == v&unicode.MaxASCII) && !(a.valid[i] && a.valid[v]) {
				return nil, fmt.Errorf("alphabet: invalid pairing: %c (%d) -> %c (%d)", i, i, v, v)
			}
		}
	}

	return &nucleic{
		alpha:   a,
		Pairing: pairs,
	}, nil
}

// NewAlphabet returns a new Alphabet based on the provided definitions. Index values
// for letters reflect order of the letters parameter. Letters must be within the
// ASCII range. No check is performed to determine whether letters appear more than once,
// the index of a letter will be the position of the last occurrence of that letter in the
// letters parameter.
func NewAlphabet(letters string, molType feat.Moltype, gap, ambiguous Letter, caseSensitive bool) (Alphabet, error) {
	return newAlphabet(letters, molType, gap, ambiguous, caseSensitive)
}
