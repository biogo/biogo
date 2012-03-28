package protein

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
	"github.com/kortschak/biogo/bio"
	"github.com/kortschak/biogo/exp/alphabet"
	"github.com/kortschak/biogo/exp/seq"
	"github.com/kortschak/biogo/exp/seq/sequtils"
	"github.com/kortschak/biogo/feat"
)

// Seq is a basic protein sequence.
type Seq struct {
	ID        string
	Desc      string
	Loc       string
	S         []alphabet.Letter
	Stringify seq.Stringify // Function allowing user specified string representation.
	Meta      interface{}   // No operation implicitly copies or changes the contents of Meta.
	alphabet  alphabet.Peptide
	circular  bool
	offset    int
}

// Create a new Seq with the given id, letter sequence and alphabet.
func NewSeq(id string, b []alphabet.Letter, alpha alphabet.Peptide) *Seq {
	return &Seq{
		ID:        id,
		S:         append([]alphabet.Letter{}, b...),
		alphabet:  alpha,
		Stringify: Stringify,
	}
}

// Interface guarantees:
var (
	_ seq.Polymer  = &Seq{}
	_ seq.Sequence = &Seq{}
	_ seq.Appender = &Seq{}
	_ Sequence     = &Seq{}
)

// Required to satisfy protein.Sequence interface.
func (self *Seq) Protein() {}

// Name returns a pointer to the ID string of the sequence.
func (self *Seq) Name() *string { return &self.ID }

// Description returns a pointer to the Desc string of the sequence.
func (self *Seq) Description() *string { return &self.Desc }

// Location returns a pointer to the Loc string of the sequence.
func (self *Seq) Location() *string { return &self.Loc }

// Raw returns a pointer to the the underlying []byte slice.
func (self *Seq) Raw() interface{} { return &self.S }

// Append letters to the sequence.
func (self *Seq) AppendQLetters(a ...alphabet.QLetter) (err error) {
	l := self.Len()
	self.S = append(self.S, make([]alphabet.Letter, len(a))...)[:l]
	for _, v := range a {
		self.S = append(self.S, v.L)
	}

	return
}

// Append Letters to the sequence.
func (self *Seq) AppendLetters(a ...alphabet.Letter) (err error) {
	self.S = append(self.S, a...)

	return
}

// Return the Alphabet used by the sequence.
func (self *Seq) Alphabet() alphabet.Alphabet { return self.alphabet }

// Return the letter at position pos.
func (self *Seq) At(pos seq.Position) alphabet.QLetter {
	if pos.Ind != 0 {
		panic("protein: index out of range")
	}
	return alphabet.QLetter{
		L: self.S[pos.Pos-self.offset],
		Q: DefaultQphred,
	}
}

// Set the letter at position pos to l.
func (self *Seq) Set(pos seq.Position, l alphabet.QLetter) {
	if pos.Ind != 0 {
		panic("protein: index out of range")
	}
	self.S[pos.Pos-self.offset] = l.L
}

// Return the length of the sequence.
func (self *Seq) Len() int { return len(self.S) }

// Satisfy Counter.
func (self *Seq) Count() int { return 1 }

// Set the global offset of the sequence to o.
func (self *Seq) Offset(o int) { self.offset = o }

// Return the start position of the sequence in global coordinates.
func (self *Seq) Start() int { return self.offset }

// Return the end position of the sequence in global coordinates.
func (self *Seq) End() int { return self.offset + self.Len() }

// Return the molecule type of the sequence.
func (self *Seq) Moltype() bio.Moltype { return self.alphabet.Moltype() }

// Validate the letters of the sequence according to the specified alphabet.
func (self *Seq) Validate() (bool, int) { return self.alphabet.AllValid(self.S) }

// Return a copy of the sequence.
func (self *Seq) Copy() seq.Sequence {
	c := *self
	c.S = append([]alphabet.Letter{}, self.S...)
	c.Meta = nil

	return &c
}

// Reverse the sequence.
func (self *Seq) Reverse() { self.S = sequtils.Reverse(self.S).([]alphabet.Letter) }

// Specify that the sequence is circular.
func (self *Seq) Circular(c bool) { self.circular = c }

// Return whether the sequence is circular.
func (self *Seq) IsCircular() bool { return self.circular }

// Return a subsequence from start to end, wrapping if the sequence is circular.
func (self *Seq) Subseq(start int, end int) (sub seq.Sequence, err error) {
	var (
		s  *Seq
		tt interface{}
	)

	if tt, err = sequtils.Truncate(self.S, start-self.offset, end-self.offset, self.circular); err == nil {
		s = &Seq{}
		*s = *self
		s.S = tt.([]alphabet.Letter)
		s.S = nil
		s.Meta = nil
		s.offset = start
		s.circular = false
	}

	return s, nil
}

// Truncate the sequence from start to end, wrapping if the sequence is circular.
func (self *Seq) Truncate(start int, end int) (err error) {
	var tt interface{}

	if tt, err = sequtils.Truncate(self.S, start-self.offset, end-self.offset, self.circular); err == nil {
		self.S = tt.([]alphabet.Letter)
		self.offset = start
		self.circular = false
	}

	return
}

// Join p to the sequence at the end specified by where.
func (self *Seq) Join(p *Seq, where int) (err error) {
	if self.circular {
		return bio.NewError("Cannot join circular sequence: receiver.", 1, self)
	} else if p.circular {
		return bio.NewError("Cannot join circular sequence: parameter.", 1, p)
	}

	tt, offset := sequtils.Join(self.S, p.S, where)
	self.offset = offset
	self.S = tt.([]alphabet.Letter)

	return
}

// Join sequentially order disjunct segments of the sequence, returning any error.
func (self *Seq) Stitch(f feat.FeatureSet) (err error) {
	var tt interface{}

	if tt, err = sequtils.Stitch(self.S, self.offset, f); err == nil {
		self.S = tt.([]alphabet.Letter)
		self.circular = false
		self.offset = 0
	}

	return
}

// Join segments of the sequence, returning any error.
func (self *Seq) Compose(f feat.FeatureSet) (err error) {
	var tt []interface{}

	if tt, err = sequtils.Compose(self.S, self.offset, f); err == nil {
		s := []alphabet.Letter{}
		for _, ts := range tt {
			s = append(s, ts.([]alphabet.Letter)...)
		}

		self.S = s
		self.circular = false
		self.offset = 0
	}

	return
}

// Return a string representation of the sequence. Representation is determined by the Stringify field.
func (self *Seq) String() string { return self.Stringify(self) }

// The default Stringify function for Seq.
var Stringify = func(s seq.Polymer) string { return alphabet.Letters(s.(*Seq).S).String() }
