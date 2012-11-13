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

// A QSeq is a basic protein acid sequence with Phred quality scores.
type QSeq struct {
	Annotation
	Seq        alphabet.QLetters
	Threshold  alphabet.Qphred // Threshold for returning valid letter.
	LowQFilter seq.Filter      // How to represent below threshold letter.
	Encode     alphabet.Encoding
}

// Interface guarantees
var (
	_ feat.Feature = &QSeq{}
	_ seq.Sequence = &QSeq{}
	_ Sequence     = &QSeq{}
	_ seq.Scorer   = &QSeq{}
)

// NewQSeq create a new QSeq with the given id, letter sequence, alphabet and quality encoding.
func NewQSeq(id string, ql []alphabet.QLetter, alpha alphabet.Peptide, enc alphabet.Encoding) *QSeq {
	return &QSeq{
		Annotation: Annotation{
			ID:    id,
			Alpha: alpha,
		},
		Seq:        append(alphabet.QLetters(nil), ql...),
		Encode:     enc,
		Threshold:  2,
		LowQFilter: LowQFilter,
	}
}

// Append append Letters to the sequence, the DefaultQphred value is used for quality scores.
func (s *QSeq) AppendLetters(a ...alphabet.Letter) error {
	l := s.Len()
	s.Seq = append(s.Seq, make([]alphabet.QLetter, len(a))...)[:l]
	for _, v := range a {
		s.Seq = append(s.Seq, alphabet.QLetter{L: v, Q: DefaultQphred})
	}
	return nil
}

// Append appends QLetters to the sequence.
func (s *QSeq) AppendQLetters(a ...alphabet.QLetter) error {
	s.Seq = append(s.Seq, a...)
	return nil
}

// Slice returns the sequence data as a alphabet.Slice.
func (s *QSeq) Slice() alphabet.Slice { return s.Seq }

// SetSlice sets the sequence data represented by the sequence. SetSlice will panic if sl
// is not a alphabet.QLetters.
func (s *QSeq) SetSlice(sl alphabet.Slice) { s.Seq = sl.(alphabet.QLetters) }

// At returns the letter at position pos.
func (s *QSeq) At(pos seq.Position) alphabet.QLetter {
	if pos.Row != 0 {
		panic("protein: index out of range")
	}
	return s.Seq[pos.Col-s.Offset]
}

// QEncode encodes the quality at position pos to a letter based on the sequence encoding setting.
func (s *QSeq) QEncode(pos seq.Position) byte {
	if pos.Row != 0 {
		panic("protein: index out of range")
	}
	return s.Seq[pos.Col-s.Offset].Q.Encode(s.Encode)
}

// Encoding returns the quality encoding scheme.
func (s *QSeq) Encoding() alphabet.Encoding { return s.Encode }

// SetEncoding sets the quality encoding scheme to e.
func (s *QSeq) SetEncoding(e alphabet.Encoding) { s.Encode = e }

// EAt returns the probability of a sequence error at position pos.
func (s *QSeq) EAt(pos seq.Position) float64 {
	if pos.Row != 0 {
		panic("protein: index out of range")
	}
	return s.Seq[pos.Col-s.Offset].Q.ProbE()
}

// Set sets the letter at position pos to l.
func (s *QSeq) Set(pos seq.Position, l alphabet.QLetter) {
	if pos.Row != 0 {
		panic("protein: index out of range")
	}
	s.Seq[pos.Col-s.Offset] = l
}

// SetE sets the quality at position pos to e to reflect the given p(Error).
func (s *QSeq) SetE(pos seq.Position, e float64) {
	if pos.Row != 0 {
		panic("protein: index out of range")
	}
	s.Seq[pos.Col-s.Offset].Q = alphabet.Ephred(e)
}

// Len returns the length of the sequence.
func (s *QSeq) Len() int { return len(s.Seq) }

// Start return the start position of the sequence in global coordinates.
func (s *QSeq) Start() int { return s.Offset }

// End returns the end position of the sequence in global coordinates.
func (s *QSeq) End() int { return s.Offset + s.Len() }

// Validate validates the letters of the sequence according to the sequence alphabet.
func (s *QSeq) Validate() (bool, int) {
	for i, ql := range s.Seq {
		if !s.Alpha.IsValid(ql.L) {
			return false, i
		}
	}

	return true, -1
}

// Copy returns a copy of the sequence.
func (s *QSeq) Copy() seq.Sequence {
	c := *s
	c.Seq = append([]alphabet.QLetter(nil), s.Seq...)

	return &c
}

// New returns an empty *QSeq sequence.
func (s *QSeq) New() seq.Sequence {
	return &QSeq{}
}

// Reverse reverses the order of letters in the the sequence without complementing them.
func (s *QSeq) Reverse() {
	l := s.Seq
	for i, j := 0, len(l)-1; i < j; i, j = i+1, j-1 {
		l[i], l[j] = l[j], l[i]
	}
}

func (s *QSeq) String() string {
	gap := s.Alpha.Gap()
	cs := make([]alphabet.Letter, 0, len(s.Seq))
	for _, ql := range s.Seq {
		if alphabet.Qphred(ql.Q) > s.Threshold || ql.L == gap {
			cs = append(cs, ql.L)
		} else {
			cs = append(cs, s.LowQFilter(s, ql.L))
		}
	}

	return alphabet.Letters(cs).String()
}

// The default LowQFilter function for QSeq.
var LowQFilter = func(s seq.Alphabeter, _ alphabet.Letter) alphabet.Letter { return s.Alphabet().Ambiguous() }
