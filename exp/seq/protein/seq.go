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

package protein

import (
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/feat"
	"code.google.com/p/biogo/exp/seq"
)

// A Seq is a basic protein acid sequence.
type Seq struct {
	Annotation
	Seq alphabet.Letters
}

// Interface guarantees
var (
	_ feat.Feature = &Seq{}
	_ seq.Sequence = &Seq{}
	_ Sequence     = &Seq{}
)

// NewSeq creates a new Seq with the given id, letter sequence and alphabet.
func NewSeq(id string, b []alphabet.Letter, alpha alphabet.Peptide) *Seq {
	return &Seq{
		Annotation: Annotation{
			ID:    id,
			Alpha: alpha,
		},
		Seq: append(alphabet.Letters(nil), b...),
	}
}

// Append append QLetters to the sequence, ignoring Q component.
func (s *Seq) AppendQLetters(a ...alphabet.QLetter) error {
	l := s.Len()
	s.Seq = append(s.Seq, make([]alphabet.Letter, len(a))...)[:l]
	for _, v := range a {
		s.Seq = append(s.Seq, v.L)
	}
	return nil
}

// Append appends Letters to the sequence.
func (s *Seq) AppendLetters(a ...alphabet.Letter) error {
	s.Seq = append(s.Seq, a...)
	return nil
}

// Slice returns the sequence data as a alphabet.Slice.
func (s *Seq) Slice() alphabet.Slice { return s.Seq }

// SetSlice sets the sequence data represented by the sequence. SetSlice will panic if sl
// is not a alphabet.Letters.
func (s *Seq) SetSlice(sl alphabet.Slice) { s.Seq = sl.(alphabet.Letters) }

// At returns the letter at position pos.
func (s *Seq) At(pos seq.Position) alphabet.QLetter {
	if pos.Row != 0 {
		panic("protein: index out of range")
	}
	return alphabet.QLetter{
		L: s.Seq[pos.Col-s.Offset],
		Q: DefaultQphred,
	}
}

// Set sets the letter at position pos to l.
func (s *Seq) Set(pos seq.Position, l alphabet.QLetter) {
	if pos.Row != 0 {
		panic("protein: index out of range")
	}
	s.Seq[pos.Col-s.Offset] = l.L
}

// Len returns the length of the sequence.
func (s *Seq) Len() int { return len(s.Seq) }

// Start returns the start position of the sequence in global coordinates.
func (s *Seq) Start() int { return s.Offset }

// End returns the end position of the sequence in global coordinates.
func (s *Seq) End() int { return s.Offset + s.Len() }

// Validate validates the letters of the sequence according to the sequence alphabet.
func (s *Seq) Validate() (bool, int) { return s.Alpha.AllValid(s.Seq) }

// Copy returns a copy of the sequence.
func (s *Seq) Copy() seq.Sequence {
	c := *s
	c.Seq = append([]alphabet.Letter(nil), s.Seq...)
	return &c
}

// New returns an empty *Seq sequence.
func (s *Seq) New() seq.Sequence {
	return &Seq{}
}

// Reverse reverses the order of letters in the the sequence without complementing them.
func (s *Seq) Reverse() {
	l := s.Seq
	for i, j := 0, len(l)-1; i < j; i, j = i+1, j-1 {
		l[i], l[j] = l[j], l[i]
	}
}

func (s *Seq) String() string { return alphabet.Letters(s.Seq).String() }
