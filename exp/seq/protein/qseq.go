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

// QSeq is a basic protein sequence with Phred quality scores.
type QSeq struct {
	ID         string
	Desc       string
	Loc        string
	S          []alphabet.QLetter
	Threshold  alphabet.Qphred // Threshold for returning valid letter.
	LowQFilter seq.Filter      // How to represent below threshold letter.
	Stringify  seq.Stringify   // Function allowing user specified string representation.
	Meta       interface{}     // No operation implicitly copies or changes the contents of Meta.
	alphabet   alphabet.Peptide
	circular   bool
	offset     int
	encoding   alphabet.Encoding
}

// Create a new QSeq with the given id, letter sequence, alphabet and quality encoding.
func NewQSeq(id string, ql []alphabet.QLetter, alpha alphabet.Peptide, encode alphabet.Encoding) *QSeq {
	return &QSeq{
		ID:         id,
		S:          append([]alphabet.QLetter{}, ql...),
		alphabet:   alpha,
		encoding:   encode,
		Threshold:  2,
		LowQFilter: func(s seq.Sequence, _ alphabet.Letter) alphabet.Letter { return s.(*QSeq).alphabet.Ambiguous() },
		Stringify:  QStringify,
	}
}

// Interface guarantees:
var (
	_ seq.Polymer  = &QSeq{}
	_ seq.Sequence = &QSeq{}
	_ seq.Scorer   = &QSeq{}
	_ seq.Appender = &QSeq{}
	_ Sequence     = &QSeq{}
	_ Quality      = &QSeq{}
)

// Required to satisfy protein.Sequence interface.
func (self *QSeq) Protein() {}

// Name returns a pointer to the ID string of the sequence.
func (self *QSeq) Name() *string { return &self.ID }

// Description returns a pointer to the Desc string of the sequence.
func (self *QSeq) Description() *string { return &self.Desc }

// Location returns a pointer to the Loc string of the sequence.
func (self *QSeq) Location() *string { return &self.Loc }

// Raw returns a pointer to the underlying []Qphred slice.
func (self *QSeq) Raw() interface{} { return &self.S }

// Append letters with quality scores to the seq.
func (self *QSeq) Append(a ...alphabet.QLetter) (err error) { self.S = append(self.S, a...); return }

// Return the Alphabet used by the sequence.
func (self *QSeq) Alphabet() alphabet.Alphabet { return self.alphabet }

// Return the letter at position pos.
func (self *QSeq) At(pos seq.Position) alphabet.QLetter {
	if pos.Ind != 0 {
		panic("protein: index out of range")
	}
	return self.S[pos.Pos-self.offset]
}

// Encode the quality at position pos to a letter based on the sequence encoding setting.
func (self *QSeq) QEncode(pos seq.Position) byte {
	if pos.Ind != 0 {
		panic("protein: index out of range")
	}
	return self.S[pos.Pos-self.offset].Q.Encode(self.encoding)
}

// Decode a quality letter to a phred score based on the sequence encoding setting.
func (self *QSeq) QDecode(l byte) alphabet.Qphred { return alphabet.DecodeToQphred(l, self.encoding) }

// Return the quality encoding type.
func (self *QSeq) Encoding() alphabet.Encoding { return self.encoding }

// Set the quality encoding type to e.
func (self *QSeq) SetEncoding(e alphabet.Encoding) { self.encoding = e }

// Return the probability of a sequence error at position pos.
func (self *QSeq) EAt(pos seq.Position) float64 {
	if pos.Ind != 0 {
		panic("protein: index out of range")
	}
	return self.S[pos.Pos-self.offset].Q.ProbE()
}

// Set the letter at position pos to l.
func (self *QSeq) Set(pos seq.Position, l alphabet.QLetter) {
	if pos.Ind != 0 {
		panic("protein: index out of range")
	}
	self.S[pos.Pos-self.offset] = l
}

// Set the quality at position pos to l to reflect the given p(Error).
func (self *QSeq) SetE(pos seq.Position, e float64) {
	if pos.Ind != 0 {
		panic("protein: index out of range")
	}
	self.S[pos.Pos-self.offset].Q = alphabet.Ephred(e)
}

// Return the length of the sequence.
func (self *QSeq) Len() int { return len(self.S) }

// Satisfy Counter.
func (self *QSeq) Count() int { return 1 }

// Set the global offset of the sequence to o.
func (self *QSeq) Offset(o int) { self.offset = o }

// Return the start position of the sequence in global coordinates.
func (self *QSeq) Start() int { return self.offset }

// Return the end position of the sequence in global coordinates.
func (self *QSeq) End() int { return self.offset + self.Len() }

// Return the molecule type of the sequence.
func (self *QSeq) Moltype() bio.Moltype { return self.alphabet.Moltype() }

// Validate the letters of the sequence according to the specified alphabet.
func (self *QSeq) Validate() (bool, int) {
	for i, ql := range self.S {
		if !self.alphabet.IsValid(ql.L) {
			return false, i
		}
	}

	return true, -1
}

// Return a copy of the sequence.
func (self *QSeq) Copy() seq.Sequence {
	c := *self
	c.S = append([]alphabet.QLetter{}, self.S...)
	c.Meta = nil

	return &c
}

// Reverse the sequence.
func (self *QSeq) Reverse() { self.S = sequtils.Reverse(self.S).([]alphabet.QLetter) }

// Specify that the sequence is circular.
func (self *QSeq) Circular(c bool) { self.circular = c }

// Return whether the sequence is circular.
func (self *QSeq) IsCircular() bool { return self.circular }

// Return a subsequence from start to end, wrapping if the sequence is circular.
func (self *QSeq) Subseq(start int, end int) (sub seq.Sequence, err error) {
	var (
		s  *QSeq
		tt interface{}
	)

	if tt, err = sequtils.Truncate(self.S, start-self.offset, end-self.offset, self.circular); err == nil {
		s = &QSeq{}
		*s = *self
		s.S = tt.([]alphabet.QLetter)
		s.S = nil
		s.Meta = nil
		s.offset = start
		s.circular = false
	}

	return s, nil
}

// Truncate the sequenc from start to end, wrapping if the sequence is circular.
func (self *QSeq) Truncate(start int, end int) (err error) {
	var tt interface{}

	if tt, err = sequtils.Truncate(self.S, start-self.offset, end-self.offset, self.circular); err == nil {
		self.S = tt.([]alphabet.QLetter)
		self.offset = start
		self.circular = false
	}

	return
}

// Join p to the sequence at the end specified by where.
func (self *QSeq) Join(p *QSeq, where int) (err error) {
	if self.circular {
		return bio.NewError("Cannot join circular sequence: receiver.", 1, self)
	} else if p.circular {
		return bio.NewError("Cannot join circular sequence: parameter.", 1, p)
	}

	var tt interface{}

	tt, self.offset = sequtils.Join(self.S, p.S, where)
	self.S = tt.([]alphabet.QLetter)

	return
}

// Join sequentially order disjunct segments of the sequence, returning any error.
func (self *QSeq) Stitch(f feat.FeatureSet) (err error) {
	var tt interface{}

	if tt, err = sequtils.Stitch(self.S, self.offset, f); err == nil {
		self.S = tt.([]alphabet.QLetter)
		self.circular = false
		self.offset = 0
	}

	return
}

// Join segments of the sequence, returning any error.
func (self *QSeq) Compose(f feat.FeatureSet) (err error) {
	var tt []interface{}

	if tt, err = sequtils.Compose(self.S, self.offset, f); err == nil {
		s := []alphabet.QLetter{}
		for _, ts := range tt {
			s = append(s, ts.([]alphabet.QLetter)...)
		}

		self.S = s
		self.circular = false
		self.offset = 0
	}

	return
}

// Return a string representation of the sequence. Representation is determined by the Stringify field.
func (self *QSeq) String() string { return self.Stringify(self) }

// The default Stringify function for QSeq.
var QStringify = func(s seq.Polymer) string {
	t := s.(*QSeq)
	gap := t.Alphabet().Gap()
	cs := make([]alphabet.Letter, 0, len(t.S))
	for _, ql := range t.S {
		if alphabet.Qphred(ql.Q) > t.Threshold || ql.L == gap {
			cs = append(cs, ql.L)
		} else {
			cs = append(cs, t.LowQFilter(t, ql.L))
		}
	}

	return alphabet.Letters(cs).String()
}

// The default LowQFilter function for QSeq.
var LowQFilter = func(s seq.Sequence, _ alphabet.Letter) alphabet.Letter { return s.(*QSeq).alphabet.Ambiguous() }
