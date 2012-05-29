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

package packed

import (
	"fmt"
	"github.com/kortschak/biogo/bio"
	"github.com/kortschak/biogo/exp/alphabet"
	"github.com/kortschak/biogo/exp/seq"
	"github.com/kortschak/biogo/exp/seq/nucleic"
	"github.com/kortschak/biogo/exp/seq/quality"
	"github.com/kortschak/biogo/exp/seq/sequtils"
	"github.com/kortschak/biogo/feat"
)

// QSeq is a packed nucleic acid with Phred quality scores allowing one byte per quality base.
type QSeq struct {
	ID         string
	Desc       string
	Loc        string
	S          []alphabet.QPack
	Strand     nucleic.Strand
	Threshold  alphabet.Qphred // Threshold for returning valid letter.
	LowQFilter seq.Filter      // How to represent below threshold letter.
	Stringify  seq.Stringify   // Function allowing user specified string representation.
	Meta       interface{}     // No operation implicitly copies or changes the contents of Meta.
	alphabet   alphabet.Nucleic
	circular   bool
	offset     int
	encoding   alphabet.Encoding
}

// Create a new QSeq with the given id, letter sequence, alphabet and quality encoding.
func NewQSeq(id string, qp []alphabet.QPack, alpha alphabet.Nucleic, encode alphabet.Encoding) (p *QSeq, err error) {
	err = checkPackedAlpha(alpha)
	if err != nil {
		return
	}

	p = &QSeq{
		ID:         id,
		S:          append([]alphabet.QPack(nil), qp...),
		alphabet:   alpha,
		encoding:   encode,
		Strand:     1,
		Threshold:  2,
		LowQFilter: LowQFilter,
		Stringify:  QStringify,
	}

	return
}

// Interface guarantees:
var (
	_ seq.Polymer      = &QSeq{}
	_ seq.Sequence     = &QSeq{}
	_ seq.Scorer       = &QSeq{}
	_ seq.Appender     = &QSeq{}
	_ nucleic.Sequence = &QSeq{}
	_ nucleic.Quality  = &QSeq{}
)

// Required to satisfy nucleic.Sequence interface.
func (self *QSeq) Nucleic() {}

// Name returns a pointer to the ID string of the sequence.
func (self *QSeq) Name() *string { return &self.ID }

// Description returns a pointer to the Desc string of the sequence.
func (self *QSeq) Description() *string { return &self.Desc }

// Location returns a pointer to the Loc string of the sequence.
func (self *QSeq) Location() *string { return &self.Loc }

// Raw returns a pointer to the underlying []alphabet.QPack slice.
func (self *QSeq) Raw() interface{} { return &self.S }

// Append QLetters to the sequence, the DefaultQphred value is used for quality scores.
func (self *QSeq) AppendLetters(a ...alphabet.Letter) (err error) {
	l := self.Len()
	self.S = append(self.S, make([]alphabet.QPack, len(a))...)[:l]
	alpha := self.alphabet
	var p alphabet.QPack
	for _, v := range a {
		p, err = (alphabet.QLetter{L: v, Q: nucleic.DefaultQphred}.Pack(alpha))
		if err != nil {
			self.S = self.S[:l]
			return
		}
		self.S = append(self.S, p)
	}

	return
}

// Append QLetters to the seq. Qualities are set to the default 0.
func (self *QSeq) AppendQLetters(a ...alphabet.QLetter) (err error) {
	self.S = append(self.S, make([]alphabet.QPack, len(a))...)[:len(self.S)]
	var qp alphabet.QPack
	for i, ql := range a {
		qp, err = ql.Pack(self.alphabet)
		if err != nil {
			if ql.Q > self.Threshold {
				return bio.NewError(fmt.Sprintf("%s %q at position %d.", err.Error(), err.(bio.Error).Items(), i), 0)
			}
		}

		self.S = append(self.S, qp)
	}

	return
}

// Return the Alphabet used by the sequence.
func (self *QSeq) Alphabet() alphabet.Alphabet { return self.alphabet }

// Return the letter as position pos.
func (self *QSeq) At(pos seq.Position) alphabet.QLetter {
	if pos.Ind != 0 {
		panic("packed: index out of range")
	}
	var q alphabet.Qphred
	if q = self.At(pos).Q; q > self.Threshold {
		return alphabet.QLetter{
			L: self.alphabet.Letter(int(self.S[pos.Pos-self.offset] & 0x3)),
			Q: q,
		}
	}
	return alphabet.QLetter{
		L: self.LowQFilter(self, self.alphabet.Letter(int(self.S[pos.Pos-self.offset]&0x3))),
		Q: q,
	}
}

// Encode the quality at position pos to a letter based on the sequence encoding setting.
func (self *QSeq) QEncode(pos seq.Position) byte {
	if pos.Ind != 0 {
		panic("packed: index out of range")
	}
	return alphabet.Qphred(self.S[pos.Pos-self.offset] >> 2).Encode(self.encoding)
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
		panic("packed: index out of range")
	}
	return alphabet.Qphred(self.S[pos.Pos-self.offset] >> 2).ProbE()
}

// Set the letter at position pos to l.
func (self *QSeq) Set(pos seq.Position, l alphabet.QLetter) {
	if pos.Ind != 0 {
		panic("packed: index out of range")
	}
	self.S[pos.Pos-self.offset] = alphabet.QPack(self.alphabet.IndexOf(l.L)) | alphabet.QPack(l.Q)<<2
}

// Set the quality at position pos to e to reflect the given p(Error).
func (self *QSeq) SetE(pos seq.Position, e float64) {
	if pos.Ind != 0 {
		panic("packed: index out of range")
	}
	self.S[pos.Pos-self.offset] &^= ((1 << 6) - 1) << 2
	self.S[pos.Pos-self.offset] = alphabet.QPack(alphabet.Ephred(e)) << 2
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

// Validate the letters of the sequence according to the specified alphabet. This is always successful as encoding does not allow invalid letters.
func (self *QSeq) Validate() (bool, int) { return true, -1 }

// Return a copy of the sequence.
func (self *QSeq) Copy() seq.Sequence {
	c := *self
	c.S = append([]alphabet.QPack(nil), self.S...)
	c.Meta = nil

	return &c
}

// Reverse complement the sequence.
func (self *QSeq) RevComp() {
	self.S = self.revComp(self.S)
	self.Strand = -self.Strand
}

func (self *QSeq) revComp(s []alphabet.QPack) []alphabet.QPack {
	i, j := 0, len(s)-1
	for ; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j]^0x3, s[i]^0x3
	}
	if i == j {
		s[i] = s[i] ^ 0x3
	}

	return s
}

// Reverse the sequence.
func (self *QSeq) Reverse() { self.S = sequtils.Reverse(self.S).([]alphabet.QPack) }

// Specify that the sequence is circular.
func (self *QSeq) Circular(c bool) { self.circular = c }

// Return whether the sequence is circular.
func (self *QSeq) IsCircular() bool { return self.circular }

// Return a subsequence from start to end, wrapping if the sequence is circular.
func (self *QSeq) Subseq(start int, end int) (sub seq.Sequence, err error) {
	var s *QSeq

	tt, err := sequtils.Truncate(self.S, start-self.offset, end-self.offset, self.circular)
	if err == nil {
		s = &QSeq{}
		*s = *self
		s.S = tt.([]alphabet.QPack)
		s.S = nil
		s.Meta = nil
		s.offset = start
		s.circular = false
	}

	return s, nil
}

// Truncate the sequenc from start to end, wrapping if the sequence is circular.
func (self *QSeq) Truncate(start int, end int) (err error) {
	tt, err := sequtils.Truncate(self.S, start-self.offset, end-self.offset, self.circular)
	if err == nil {
		self.S = tt.([]alphabet.QPack)
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
	self.S = tt.([]alphabet.QPack)

	return
}

// Join sequentially order disjunct segments of the sequence, returning any error.
func (self *QSeq) Stitch(f feat.FeatureSet) (err error) {
	tt, err := sequtils.Stitch(self.S, self.offset, f)
	if err == nil {
		self.S = tt.([]alphabet.QPack)
		self.circular = false
		self.offset = 0
	}

	return
}

// Join segments of the sequence, returning any error.
func (self *QSeq) Compose(f feat.FeatureSet) (err error) {
	tt, err := sequtils.Compose(self.S, self.offset, f)
	if err == nil {
		s := []alphabet.QPack{}
		for i, ts := range tt {
			if f[i].Strand == -1 {
				s = append(s, self.revComp(ts.([]alphabet.QPack))...)
			} else {
				s = append(s, ts.([]alphabet.QPack)...)
			}
		}

		self.S = s
		self.circular = false
		self.offset = 0
	}

	return
}

// Return an unpacked sequence and quality.
func (self *QSeq) Unpack() (n *nucleic.Seq, q *quality.Phred) {
	n = nucleic.NewSeq(self.ID, alphabet.BytesToLetters([]byte(self.String())), self.alphabet)
	n.Circular(self.circular)
	n.Offset(self.offset)
	qb := make([]alphabet.Qphred, self.Len())
	for i, v := range self.S {
		qb[i] = alphabet.Qphred(v >> 2)
	}
	q = quality.NewPhred(self.ID, qb, self.encoding)
	q.Circular(self.circular)
	q.Offset(self.offset)

	return
}

// Return a string representation of the sequence. Representation is determined by the Stringify field.
func (self *QSeq) String() string { return self.Stringify(self) }

// The default Stringify function for QSeq.
var QStringify = func(s seq.Polymer) string {
	t := s.(*QSeq)
	cs := make([]alphabet.Letter, 0, len(t.S))
	for _, l := range t.S {
		if alphabet.Qphred(l>>2) > t.Threshold {
			cs = append(cs, t.alphabet.Letter(int(l&0x3)))
		} else {
			cs = append(cs, t.LowQFilter(t, t.alphabet.Letter(int(l&0x3))))
		}
	}

	return alphabet.Letters(cs).String()
}

// The default LowQFilter function for QSeq.
var LowQFilter = func(s seq.Sequence, _ alphabet.Letter) alphabet.Letter { return s.(*QSeq).alphabet.Ambiguous() }
