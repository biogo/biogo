package alphabet
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
	N        = "acgt"
	Npairing = [2]string{"acgtnxACGTNX-", "tgcanxTGCANX-"}
	R        = "acgu"
	Rpairing = [2]string{"acgunxACGUNX-", "ugcanxUGCANX-"}
	P        = "abcdefghijklmnpqrstvxyz*"
)

var (
	DNA     *Deoxyribonucleic
	RNA     *Ribonucleic
	Protein *Peptide
)

func init() {
	if err := Init(); err != nil {
		panic(err)
	}
}

// Provide default Alphabets.
func Init() (err error) {
	var pairing *Pairing
	if pairing, err = NewPairing(Npairing[0], Npairing[1]); err != nil {
		return
	} else if DNA, err = NewDeoxyribonucleic(N, pairing, !CaseSensitive); err != nil {
		return

	}
	if pairing, err = NewPairing(Rpairing[0], Rpairing[1]); err != nil {
		return
	} else if RNA, err = NewRibonucleic(R, pairing, !CaseSensitive); err != nil {
		return

	}
	if Protein, err = NewPeptide(P, !CaseSensitive); err != nil {
		return
	}

	return
}

// Minimum requirements for an Alphabet.
type Alphabet interface {
	IsValid(byte) bool
	AllValid([]byte) (bool, int)
	Len() int
	ValidLetters() []bool
	LetterIndex() []int
	String() string
}

// Nucleic alphabets are able to complement their values.
type Complementable interface {
	Alphabet
	ComplementOf(byte) (byte, bool)
	ComplementTable() []byte
}

// Single letter alphabet type.
type Generic struct {
	letters       string
	valid         [256]bool
	index         [256]int
	caseSensitive bool
}

// Return a new alphabet. Index values for letters reflect order of the letters parameter if Generic is case sensitive,
// otherwise index values will reflect ASCII sort order. Letters must be within the ASCII range.
func NewGeneric(letters string, caseSensitive bool) (a *Generic, err error) {
	if strings.IndexFunc(letters, func(r rune) bool { return r < 0 || r > unicode.MaxASCII }) > -1 {
		return nil, errors.New("letters contains non-ASCII rune.")
	}

	a = &Generic{
		caseSensitive: caseSensitive,
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
		let := make([]byte, 0, size)
		for _, l := range ll {
			let = append(let, byte(l))
		}
		a.letters = string(let)
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

// Return the number of distinct valid letters in the alphabet.
func (self *Generic) Len() int { return len(self.letters) }

// Return whether the alphabet is case sensitive.
func (self *Generic) IsCaseSensitive() bool { return self.caseSensitive }

// Check that a slice of bytes conforms to the alphabet, returning false
// and the position of the first invalid byte if invalid and true and a negative
// int if valid.
func (self *Generic) AllValid(n []byte) (valid bool, pos int) {
	for i, v := range n {
		if !self.valid[v] {
			return false, i
		}
	}

	return true, -1
}

// Check that a byte conforms to the alphabet.
func (self *Generic) IsValid(n byte) bool {
	return self.valid[n]
}

// Return the index of a letter.
func (self *Generic) IndexOf(n byte) int {
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
	pair []byte
	ok   []bool
}

// Create a new Pairing from a pair of strings. 
func NewPairing(s, c string) (p *Pairing, err error) {
	if len(s) != len(c) {
		return nil, errors.New("Length of pairing definitions do not match.")
	}

	p = &Pairing{
		pair: make([]byte, 256),
		ok:   make([]bool, 256),
	}

	for i := range p.pair {
		p.pair[i] = byte(i)
	}

	cr := []rune(c)
	for i, v := range s {
		if v < 0 || cr[i] < 0 || v > unicode.MaxASCII || cr[i] > unicode.MaxASCII {
			return nil, errors.New("Pairing definition contains non-ASCII rune.")
		}
		p.pair[v] = byte(cr[i])
		p.ok[v] = true
	}

	return
}

// Returns the complement of a letter and true if the complement is a valid letter otherwise unchanged and false.
func (self *Pairing) ComplementOf(l byte) (c byte, ok bool) { return self.pair[l], self.ok[l] }

// Returns a complementation table based on the internal representation. Invalid pairs hold a value outside the ASCII range.
func (self *Pairing) ComplementTable() (t []byte) {
	t = make([]byte, 256)
	copy(t, self.pair)
	for i, ok := range self.ok {
		if !ok {
			t[i] |= unicode.MaxASCII + 1
		}
	}

	return
}

// The Nucleic type incorporates a Generic alphabet with the capacity to return a complement.
type Nucleic struct {
	*Generic
	*Pairing
}

// Create a generalised Nucleic alphabet. The Complement table is checked for validity and an error is returned if an invalid complement pair is found.
// Pairings that result in no change but would otherwise be invalid are allowed. If invalid pairings are required, the Pairing should be provided after
// creating the Nucleic struct.
func NewNucleic(letters string, pairs *Pairing, caseSensitive bool) (n *Nucleic, err error) {
	var g *Generic
	if g, err = NewGeneric(letters, caseSensitive); err != nil {
		return nil, err
	}

	if pairs != nil {
		for i, v := range pairs.pair {
			if !(pairs.ok[i] || byte(i&unicode.MaxASCII) == v&unicode.MaxASCII) && !(g.valid[i] && g.valid[v]) {
				return nil, errors.New(fmt.Sprintf("Invalid pairing: %c (%d) -> %c (%d)", i, i, v, v))
			}
		}
	}

	return &Nucleic{
		Generic: g,
		Pairing: pairs,
	}, nil
}

// Deoxyribonucleic is an alias to Nucleic to provide type restrictions.
type Deoxyribonucleic Nucleic

// Return a new Deoxyribonucleic alphabet.
func NewDeoxyribonucleic(letters string, pairs *Pairing, caseSensitive bool) (d *Deoxyribonucleic, err error) {
	var n *Nucleic
	if n, err = NewNucleic(letters, pairs, caseSensitive); err != nil {
		return nil, err
	}
	return (*Deoxyribonucleic)(n), nil
}

// Ribonucleic is an alias to Nucleic to provide type restrictions.
type Ribonucleic Nucleic

// Return a new Ribonucleic alphabet.
func NewRibonucleic(letters string, pairs *Pairing, caseSensitive bool) (r *Ribonucleic, err error) {
	var n *Nucleic
	if n, err = NewNucleic(letters, pairs, caseSensitive); err != nil {
		return nil, err
	}
	return (*Ribonucleic)(n), nil
}

// Peptide wraps Generic to provide type restrictions.
type Peptide struct {
	*Generic
}

// Return a new Peptide alphabet.
func NewPeptide(letters string, caseSensitive bool) (p *Peptide, err error) {
	var g *Generic
	if g, err = NewGeneric(letters, caseSensitive); err != nil {
		return nil, err
	}
	return &Peptide{g}, nil
}
