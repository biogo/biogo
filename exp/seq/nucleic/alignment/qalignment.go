package alignment

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
	"fmt"
	"github.com/kortschak/biogo/bio"
	"github.com/kortschak/biogo/exp/alphabet"
	"github.com/kortschak/biogo/exp/seq"
	"github.com/kortschak/biogo/exp/seq/nucleic"
	"github.com/kortschak/biogo/exp/seq/sequtils"
	"github.com/kortschak/biogo/feat"
	"github.com/kortschak/biogo/util"
)

// Aligned nucleic acid with quality scores.
type QSeq struct {
	ID         string
	Desc       string
	Loc        string
	SubIDs     []string
	S          [][]alphabet.QLetter
	Consensify nucleic.Consensifyer
	Strand     nucleic.Strand
	Threshold  alphabet.Qphred // Threshold for returning valid letter.
	LowQFilter seq.Filter      // How to represent below threshold letter.
	Stringify  seq.Stringify
	Meta       interface{} // No operation implicitly copies or changes the contents of Meta.
	alphabet   alphabet.Nucleic
	circular   bool
	offset     int
	encoding   alphabet.Encoding
}

func NewQSeq(id string, subids []string, ql [][]alphabet.QLetter, alpha alphabet.Nucleic, encode alphabet.Encoding, cons nucleic.Consensifyer) (*QSeq, error) {
	switch lids, lseq := len(subids), len(ql); {
	case lids == 0 && len(ql) == 0:
	case lseq != 0 && lids == len(ql[0]):
		if lids == 0 {
			subids = make([]string, len(ql[0]))
			for i := range subids {
				subids[i] = fmt.Sprintf("%s:%d", id, i)
			}
		}
	default:
		return nil, bio.NewError("alignment: id/seq number mismatch", 0)
	}

	return &QSeq{
		ID:         id,
		SubIDs:     append([]string{}, subids...),
		S:          append([][]alphabet.QLetter{}, ql...),
		alphabet:   alpha,
		encoding:   encode,
		Strand:     1,
		Consensify: cons,
		Threshold:  2,
		LowQFilter: func(s seq.Sequence, _ alphabet.Letter) alphabet.Letter { return s.Alphabet().Ambiguous() },
		Stringify: func(s seq.Polymer) string {
			t := s.(*QSeq).Consensus(false)
			t.Threshold = s.(*QSeq).Threshold
			t.LowQFilter = s.(*QSeq).LowQFilter
			return t.String()
		},
	}, nil
}

// Interface guarantees:
var (
	_ seq.Polymer             = &QSeq{}
	_ seq.Sequence            = &QSeq{}
	_ seq.Scorer              = &QSeq{}
	_ nucleic.Sequence        = &QSeq{}
	_ nucleic.Quality         = &QSeq{}
	_ nucleic.Extracter       = &QSeq{}
	_ nucleic.Aligned         = &QSeq{}
	_ nucleic.AlignedAppender = &QSeq{}
)

// Required to satisfy nucleic.Sequence interface.
func (self *QSeq) Nucleic() {}

// Name returns a pointer to the ID string of the sequence.
func (self *QSeq) Name() *string { return &self.ID }

// Description returns a pointer to the Desc string of the sequence.
func (self *QSeq) Description() *string { return &self.Desc }

// Location returns a pointer to the Loc string of the sequence.
func (self *QSeq) Location() *string { return &self.Loc }

// Raw returns a pointer to the underlying [][]alphabet.QLetter slice.
func (self *QSeq) Raw() interface{} { return &self.S }

// Append each byte of each a to the appropriate sequence in the reciever.
func (self *QSeq) AppendColumns(a ...[]alphabet.QLetter) (err error) {
	for i, s := range a {
		if len(s) != self.Count() {
			return bio.NewError(fmt.Sprintf("Column %d does not match Count(): %d != %d.", i, len(s), self.Count()), 0, a)
		}
	}

	self.S = append(self.S, a...)

	return
}

// Append each []byte in a to the appropriate sequence in the reciever.
func (self *QSeq) AppendEach(a [][]alphabet.QLetter) (err error) {
	if len(a) != self.Count() {
		return bio.NewError(fmt.Sprintf("Number of sequences does not match Count(): %d != %d.", len(a), self.Count()), 0, a)
	}
	max := util.MinInt
	for _, s := range a {
		if l := len(s); l > max {
			max = l
		}
	}
	self.S = append(self.S, make([][]alphabet.QLetter, max)...)[:len(self.S)]
	for i, b := 0, make([]alphabet.QLetter, 0, len(a)); i < max; i, b = i+1, b[:0] {
		for _, s := range a {
			if i < len(s) {
				b = append(b, s[i])
			} else {
				b = append(b, alphabet.QLetter{L: self.alphabet.Gap()})
			}
		}
		self.AppendColumns(b)
	}

	return
}

func (self *QSeq) column(m []nucleic.Sequence, pos int) (c []alphabet.QLetter) {
	count := 0
	for _, s := range m {
		count += s.Count()
	}

	c = make([]alphabet.QLetter, 0, count)

	for _, s := range m {
		if a, ok := s.(nucleic.Aligned); ok {
			if a.Start() <= pos && pos < a.End() {
				c = append(c, a.ColumnQL(pos, true)...)
			} else {
				c = append(c, alphabet.QLetter{L: self.alphabet.Gap()}.Repeat(a.Count())...)
			}
		} else {
			if s.Start() <= pos && pos < s.End() {
				c = append(c, s.At(seq.Position{Pos: pos}))
			} else {
				c = append(c, alphabet.QLetter{L: self.alphabet.Gap()})
			}
		}
	}

	return
}

// TODO
// func (self *QSeq) Delete(i int) {}

// Add sequences n to Alignment. Sequences in n must align start and end with the receiving alignment.
// Additional sequence will be clipped.
func (self *QSeq) Add(n ...nucleic.Sequence) (err error) {
	for i := self.Start(); i < self.End(); i++ {
		self.S[i] = append(self.S[i], self.column(n, i)...)
	}
	for i := range n {
		self.SubIDs = append(self.SubIDs, *n[i].Name())
	}

	return
}

func (self *QSeq) Extract(i int) nucleic.Sequence {
	s := make([]alphabet.QLetter, 0, self.Len())
	for _, c := range self.S {
		s = append(s, c[i])
	}

	return nucleic.NewQSeq(self.SubIDs[i], s, self.alphabet, self.encoding)
}

func (self *QSeq) Alphabet() alphabet.Alphabet { return self.alphabet }

func (self *QSeq) At(pos seq.Position) alphabet.QLetter {
	return self.S[pos.Pos-self.offset][pos.Ind]
}

func (self *QSeq) QEncode(pos seq.Position) byte {
	return self.S[pos.Pos-self.offset][pos.Ind].Q.Encode(self.encoding)
}

func (self *QSeq) QDecode(l byte) alphabet.Qphred {
	return alphabet.DecodeToQphred(l, self.encoding)
}

func (self *QSeq) Encoding() alphabet.Encoding { return self.encoding }

// Set the quality encoding type to e.
func (self *QSeq) SetEncoding(e alphabet.Encoding) { self.encoding = e }

func (self *QSeq) EAt(pos seq.Position) float64 {
	return self.S[pos.Pos-self.offset][pos.Ind].Q.ProbE()
}

func (self *QSeq) Set(pos seq.Position, l alphabet.QLetter) {
	self.S[pos.Pos-self.offset][pos.Ind] = l
}

func (self *QSeq) SetE(pos seq.Position, l float64) {
	self.S[pos.Pos-self.offset][pos.Ind].Q = alphabet.Ephred(l)
}

func (self *QSeq) Column(pos int, _ bool) (c []alphabet.Letter) {
	c = make([]alphabet.Letter, self.Count())
	for i, l := range self.S[pos] {
		if l.Q > self.Threshold {
			c[i] = l.L
		} else {
			c[i] = self.LowQFilter(self, 0)
		}
	}

	return
}

func (self *QSeq) ColumnQL(pos int, _ bool) []alphabet.QLetter { return self.S[pos] }

func (self *QSeq) Len() int { return len(self.S) }

func (self *QSeq) Count() int { return len(self.S[0]) }

func (self *QSeq) Offset(o int) { self.offset = o }

func (self *QSeq) Start() int { return self.offset }

func (self *QSeq) End() int { return self.offset + self.Len() }

func (self *QSeq) Copy() seq.Sequence {
	c := *self
	c.S = make([][]alphabet.QLetter, len(self.S))
	for i, s := range self.S {
		c.S[i] = append([]alphabet.QLetter{}, s...)
	}
	c.Meta = nil

	return &c
}

func (self *QSeq) RevComp() {
	self.S = self.revComp(self.S, self.alphabet.ComplementTable())
	self.Strand = -self.Strand
}

func (self *QSeq) revComp(rs [][]alphabet.QLetter, complement []alphabet.Letter) [][]alphabet.QLetter {
	i, j := 0, len(rs)-1
	for ; i < j; i, j = i+1, j-1 {
		for s := range rs[i] {
			rs[i][s].L, rs[j][s].L = complement[rs[j][s].L], complement[rs[i][s].L]
			rs[i][s].Q, rs[j][s].Q = rs[j][s].Q, rs[i][s].Q
		}
	}
	if i == j {
		for s := range rs[i] {
			rs[i][s].L = complement[rs[i][s].L]
		}
	}

	return rs
}

func (self *QSeq) Reverse() { self.S = sequtils.Reverse(self.S).([][]alphabet.QLetter) }

func (self *QSeq) Circular(c bool) { self.circular = c }

func (self *QSeq) IsCircular() bool { return self.circular }

// Return a subsequence from start to end, wrapping if the sequence is circular.
func (self *QSeq) Subseq(start int, end int) (sub seq.Sequence, err error) {
	var s *QSeq

	tt, err := sequtils.Truncate(self.S, start-self.offset, end-self.offset, self.circular)
	if err == nil {
		s = &QSeq{}
		*s = *self
		s.S = tt.([][]alphabet.QLetter)
		s.S = nil
		s.Meta = nil
		s.offset = start
		s.circular = false
	}

	return s, nil
}

func (self *QSeq) Truncate(start int, end int) (err error) {
	tt, err := sequtils.Truncate(self.S, start-self.offset, end-self.offset, self.circular)
	if err == nil {
		self.S = tt.([][]alphabet.QLetter)
		self.offset = start
		self.circular = false
	}

	return
}

func (self *QSeq) Join(p *QSeq, where int) (err error) {
	if self.circular {
		return bio.NewError("Cannot join circular sequence: receiver.", 1, self)
	} else if p.circular {
		return bio.NewError("Cannot join circular sequence: parameter.", 1, p)
	}

	var tt interface{}

	tt, self.offset = sequtils.Join(self.S, p.S, where)
	self.S = tt.([][]alphabet.QLetter)

	return
}

func (self *QSeq) Stitch(f feat.FeatureSet) (err error) {
	tt, err := sequtils.Stitch(self.S, self.offset, f)
	if err == nil {
		self.S = tt.([][]alphabet.QLetter)
		self.circular = false
		self.offset = 0
	}

	return
}

func (self *QSeq) Compose(f feat.FeatureSet) (err error) {
	tt, err := sequtils.Compose(self.S, self.offset, f)
	if err == nil {
		s := [][]alphabet.QLetter{}
		complement := self.alphabet.ComplementTable()
		for i, ts := range tt {
			if f[i].Strand == -1 {
				s = append(s, self.revComp(ts.([][]alphabet.QLetter), complement)...)
			} else {
				s = append(s, ts.([][]alphabet.QLetter)...)
			}
		}

		self.S = s
		self.circular = false
		self.offset = 0
	}

	return
}

func (self *QSeq) String() string { return self.Stringify(self) }

func (self *QSeq) Consensus(_ bool) (qs *nucleic.QSeq) {
	cs := make([]alphabet.QLetter, 0, self.Len())
	for i := range self.S {
		cs = append(cs, self.Consensify(self, i, false))
	}

	qs = nucleic.NewQSeq("Consensus:"+self.ID, cs, self.alphabet, alphabet.Sanger)
	qs.Strand = self.Strand
	qs.Offset(self.offset)
	qs.Circular(self.circular)

	return
}
