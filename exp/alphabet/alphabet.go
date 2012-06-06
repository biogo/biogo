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

// Package alphabet describes biological sequence letters, including quality scores.
package alphabet

import (
	"code.google.com/p/biogo/bio"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"
)

const (
	CaseSensitive = true
)

var (
	N                 = "acgt"
	Npairing          = [2]string{"acgtnxACGTNX-", "tgcanxTGCANX-"}
	R                 = "acgu"
	Rpairing          = [2]string{"acgunxACGUNX-", "ugcanxUGCANX-"}
	Nambiguous Letter = 'n'
	P                 = "abcdefghijklmnpqrstvxyz*"
	Pambiguous Letter = 'x'
	Gap        Letter = '-'
)

var (
	DNA, RNA Nucleic
	Protein  Peptide
)

func init() {
	if err := Init(); err != nil {
		panic(err)
	}
}

// Provide default Alphabets.
func Init() (err error) {
	pairing, err := NewPairing(Npairing[0], Npairing[1])
	if err != nil {
		return
	}
	DNA, err = NewNucleic(N, bio.DNA, pairing, Gap, Nambiguous, !CaseSensitive)
	if err != nil {
		return
	}

	pairing, err = NewPairing(Rpairing[0], Rpairing[1])
	if err != nil {
		return
	}
	RNA, err = NewNucleic(R, bio.RNA, pairing, Gap, Nambiguous, !CaseSensitive)
	if err != nil {
		return
	}

	Protein, err = NewPeptide(P, Gap, Pambiguous, !CaseSensitive)
	if err != nil {
		return
	}

	return
}

// Minimum requirements for an Alphabet.
type Alphabet interface {
	IsValid(Letter) bool
	AllValid([]Letter) (bool, int)
	AllValidQLetter([]QLetter) (bool, int)
	Len() int
	Moltype() bio.Moltype
	IndexOf(Letter) int
	Letter(int) Letter
	ValidLetters() []bool
	LetterIndex() []int
	Gap() Letter
	Ambiguous() Letter
	String() string
}

// Nucleic alphabets are able to complement their values.
type Complementor interface {
	Complement(Letter) (Letter, bool)
	ComplementTable() []Letter
}

// Single letter alphabet type.
type Generic struct {
	letters        string
	valid          [256]bool
	index          [256]int
	gap, ambiguous Letter
	caseSensitive  bool
	molType        bio.Moltype
}

// Return a new alphabet. Index values for letters reflect order of the letters parameter if Generic is case sensitive,
// otherwise index values will reflect ASCII sort order. Letters must be within the ASCII range.
func NewGeneric(letters string, molType bio.Moltype, gap, ambiguous Letter, caseSensitive bool) (a *Generic, err error) {
	if strings.IndexFunc(letters, func(r rune) bool { return r < 0 || r > unicode.MaxASCII }) > -1 {
		return nil, errors.New("letters contains non-ASCII rune.")
	}

	a = &Generic{
		gap:           gap,
		ambiguous:     ambiguous,
		caseSensitive: caseSensitive,
		molType:       molType,
	}

	if caseSensitive {
		a.letters = letters
	} else {
		set := make(map[rune]struct{}, len(letters))
		for _, l := range letters {
			set[unicode.ToLower(l)] = struct{}{}
		}
		size := len(set)
		ll := make([]int, 0, size)
		for l := range set {
			ll = append(ll, int(l))
		}
		sort.Ints(ll)
		let := make([]Letter, 0, size)
		for _, l := range ll {
			let = append(let, Letter(l))
		}
		a.letters = string(LettersToBytes(let))
	}

	for i := range a.index {
		a.index[i] = -1
	}

	for i, l := range a.letters {
		a.valid[l] = true
		a.index[l] = i
		if !caseSensitive {
			a.valid[unicode.ToUpper(l)] = true
			a.index[unicode.ToUpper(l)] = i
		}
	}

	return
}

// Return the molecule type of the alphabet.
func (self *Generic) Moltype() bio.Moltype { return self.molType }

// Return the number of distinct valid letters in the alphabet.
func (self *Generic) Len() int { return len(self.letters) }

// Return whether the alphabet is case sensitive.
func (self *Generic) IsCaseSensitive() bool { return self.caseSensitive }

// Return the gap character.
func (self *Generic) Gap() Letter { return self.gap }

// Return the character representing an ambiguous letter.
func (self *Generic) Ambiguous() Letter { return self.ambiguous }

// Check that a slice of bytes conforms to the alphabet, returning false
// and the position of the first invalid byte if invalid and true and a negative
// int if valid.
func (self *Generic) AllValidQLetter(n []QLetter) (valid bool, pos int) {
	for i, v := range n {
		if !self.valid[v.L] {
			return false, i
		}
	}

	return true, -1
}

// Check that a slice of bytes conforms to the alphabet, returning false
// and the position of the first invalid byte if invalid and true and a negative
// int if valid.
func (self *Generic) AllValid(n []Letter) (valid bool, pos int) {
	for i, v := range n {
		if !self.valid[v] {
			return false, i
		}
	}

	return true, -1
}

// Check that a byte conforms to the alphabet.
func (self *Generic) IsValid(n Letter) bool {
	return self.valid[n]
}

// Return the letter for and index.
func (self *Generic) Letter(i int) Letter {
	if !self.caseSensitive {
		return Letter(unicode.ToLower(rune(self.letters[i])))
	}
	return Letter(self.letters[i])
}

// Return the index of a letter.
func (self *Generic) IndexOf(n Letter) int {
	return self.index[n]
}

// Return a copy of the internal []bool indicating valid letters.
func (self *Generic) ValidLetters() (v []bool) {
	v = make([]bool, 256)
	copy(v, self.valid[:])
	return
}

// Return a copy of the internal []int specifying letter to index conversion.
func (self *Generic) LetterIndex() (i []int) {
	i = make([]int, 256)
	copy(i, self.index[:])

	return
}

// Return a string indicating characters accepted as valid by the Validator.
func (self *Generic) String() (s string) {
	s = self.letters

	if !self.caseSensitive {
		s += strings.ToUpper(s)
	}

	return
}

// Pairing provides a lookup table between a letter and its complement.
type Pairing struct {
	pair []Letter
	ok   []bool
}

// Create a new Pairing from a pair of strings. 
func NewPairing(s, c string) (p *Pairing, err error) {
	if len(s) != len(c) {
		return nil, errors.New("Length of pairing definitions do not match.")
	}

	p = &Pairing{
		pair: make([]Letter, 256),
		ok:   make([]bool, 256),
	}

	for i := range p.pair {
		p.pair[i] = Letter(i)
	}

	cr := []rune(c)
	for i, v := range s {
		if v < 0 || cr[i] < 0 || v > unicode.MaxASCII || cr[i] > unicode.MaxASCII {
			return nil, errors.New("Pairing definition contains non-ASCII rune.")
		}
		p.pair[v] = Letter(cr[i])
		p.ok[v] = true
	}

	return
}

// Returns the complement of a letter and true if the complement is a valid letter otherwise unchanged and false.
func (self *Pairing) Complement(l Letter) (c Letter, ok bool) { return self.pair[l], self.ok[l] }

// Returns a complementation table based on the internal representation. Invalid pairs hold a value outside the ASCII range.
func (self *Pairing) ComplementTable() (t []Letter) {
	t = make([]Letter, 256)
	copy(t, self.pair)
	for i, ok := range self.ok {
		if !ok {
			t[i] |= unicode.MaxASCII + 1
		}
	}

	return
}

// The Nucleic type incorporates a Generic alphabet with the capacity to return a complement.
type Nucleic interface {
	Alphabet
	Complementor
	nucleic()
}

type nucleic struct {
	*Generic
	*Pairing
}

func (n nucleic) nucleic() {}

// Create a generalised Nucleic alphabet. The Complement table is checked for validity and an error is returned if an invalid complement pair is found.
// Pairings that result in no change but would otherwise be invalid are allowed. If invalid pairings are required, the Pairing should be provided after
// creating the Nucleic struct.
func NewNucleic(letters string, molType bio.Moltype, pairs *Pairing, gap, ambiguous Letter, caseSensitive bool) (n Nucleic, err error) {
	g, err := NewGeneric(letters, molType, gap, ambiguous, caseSensitive)
	if err != nil {
		return
	}

	if pairs != nil {
		for i, v := range pairs.pair {
			if !(pairs.ok[i] || Letter(i&unicode.MaxASCII) == v&unicode.MaxASCII) && !(g.valid[i] && g.valid[v]) {
				return nil, errors.New(fmt.Sprintf("Invalid pairing: %c (%d) -> %c (%d)", i, i, v, v))
			}
		}
	}

	return &nucleic{
		Generic: g,
		Pairing: pairs,
	}, nil
}

// Peptide wraps Generic to provide type restrictions.
type Peptide interface {
	Alphabet
	peptide()
}

type peptide struct {
	*Generic
}

func (p peptide) peptide() {}

// Return a new Peptide alphabet.
func NewPeptide(letters string, gap, ambiguous Letter, caseSensitive bool) (p Peptide, err error) {
	g, err := NewGeneric(letters, bio.Protein, gap, ambiguous, caseSensitive)
	if err != nil {
		return
	}
	return &peptide{g}, nil
}
