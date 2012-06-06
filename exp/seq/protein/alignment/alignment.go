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

// Package alignment handles aligned sequences stored as columns.
package alignment

import (
	"code.google.com/p/biogo/bio"
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/exp/seq/protein"
	"code.google.com/p/biogo/exp/seq/sequtils"
	"code.google.com/p/biogo/feat"
	"code.google.com/p/biogo/util"
	"fmt"
)

// Alignment protein sequence.
type Seq struct {
	ID         string
	Desc       string
	Loc        string
	SubIDs     []string
	S          [][]alphabet.Letter
	Consensify protein.Consensifyer
	Stringify  seq.Stringify
	Meta       interface{} // No operation implicitly copies or changes the contents of Meta.
	alphabet   alphabet.Peptide
	circular   bool
	offset     int
}

func NewSeq(id string, subids []string, b [][]alphabet.Letter, alpha alphabet.Peptide, cons protein.Consensifyer) (*Seq, error) {
	switch lids, lseq := len(subids), len(b); {
	case lids == 0 && len(b) == 0:
	case lseq != 0 && lids == len(b[0]):
		if lids == 0 {
			subids = make([]string, len(b[0]))
			for i := range subids {
				subids[i] = fmt.Sprintf("%s:%d", id, i)
			}
		}
	default:
		return nil, bio.NewError("alignment: id/seq number mismatch", 0)
	}

	return &Seq{
		ID:         id,
		SubIDs:     append([]string(nil), subids...),
		S:          append([][]alphabet.Letter(nil), b...),
		alphabet:   alpha,
		Consensify: cons,
		Stringify: func(s seq.Polymer) string {
			t := s.(*Seq).Consensus(false)
			return t.String()
		},
	}, nil
}

// Interface guarantees:
var (
	_ seq.Polymer             = &Seq{}
	_ seq.Sequence            = &Seq{}
	_ protein.Sequence        = &Seq{}
	_ protein.Extracter       = &Seq{}
	_ protein.Aligned         = &Seq{}
	_ protein.AlignedAppender = &Seq{}
)

// Required to satisfy protein.Sequence interface.
func (self *Seq) Protein() {}

// Raw returns a pointer to the underlying [][]byte slice.
func (self *Seq) Raw() interface{} { return &self.S }

// Append each byte of each a to the appropriate sequence in the reciever.
func (self *Seq) AppendColumns(a ...[]alphabet.QLetter) (err error) {
	for i, s := range a {
		if len(s) != self.Count() {
			return bio.NewError(fmt.Sprintf("Column %d does not match Count(): %d != %d.", i, len(s), self.Count()), 0, a)
		}
	}

	self.S = append(self.S, make([][]alphabet.Letter, len(a))...)[:len(self.S)]
	for _, s := range a {
		c := make([]alphabet.Letter, len(s))
		for i := range s {
			c[i] = s[i].L
		}
		self.S = append(self.S, c)
	}

	return
}

// Append each []byte in a to the appropriate sequence in the reciever.
func (self *Seq) AppendEach(a [][]alphabet.QLetter) (err error) {
	if len(a) != self.Count() {
		return bio.NewError(fmt.Sprintf("Number of sequences does not match Count(): %d != %d.", len(a), self.Count()), 0, a)
	}
	max := util.MinInt
	for _, s := range a {
		if l := len(s); l > max {
			max = l
		}
	}
	self.S = append(self.S, make([][]alphabet.Letter, max)...)[:len(self.S)]
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

// Name returns a pointer to the ID string of the sequence.
func (self *Seq) Name() *string { return &self.ID }

// Description returns a pointer to the Desc string of the sequence.
func (self *Seq) Description() *string { return &self.Desc }

// Location returns a pointer to the Loc string of the sequence.
func (self *Seq) Location() *string { return &self.Loc }

func (self *Seq) column(m []protein.Sequence, pos int) (c []alphabet.Letter) {
	count := 0
	for _, s := range m {
		count += s.Count()
	}

	c = make([]alphabet.Letter, 0, count)

	for _, s := range m {
		if a, ok := s.(protein.Aligned); ok {
			if a.Start() <= pos && pos < a.End() {
				c = append(c, a.Column(pos, true)...)
			} else {
				c = append(c, self.alphabet.Gap().Repeat(a.Count())...)
			}
		} else {
			if s.Start() <= pos && pos < s.End() {
				c = append(c, s.At(seq.Position{Pos: pos}).L)
			} else {
				c = append(c, self.alphabet.Gap())
			}
		}
	}

	return
}

// TODO
// func (self *Seq) Delete(i int) {}

// Add sequences n to Seq. Sequences in n must align start and end with the receiving alignment.
// Additional sequence will be clipped.
func (self *Seq) Add(n ...protein.Sequence) (err error) {
	for i := self.Start(); i < self.End(); i++ {
		self.S[i] = append(self.S[i], self.column(n, i)...)
	}
	for i := range n {
		self.SubIDs = append(self.SubIDs, *n[i].Name())
	}

	return
}

func (self *Seq) Extract(i int) protein.Sequence {
	s := make([]alphabet.Letter, 0, self.Len())
	for _, c := range self.S {
		s = append(s, c[i])
	}

	return protein.NewSeq(self.SubIDs[i], s, self.alphabet)
}

func (self *Seq) Alphabet() alphabet.Alphabet { return self.alphabet }

func (self *Seq) At(pos seq.Position) alphabet.QLetter {
	return alphabet.QLetter{
		L: self.S[pos.Pos-self.offset][pos.Ind],
		Q: protein.DefaultQphred,
	}
}

func (self *Seq) Set(pos seq.Position, l alphabet.QLetter) {
	self.S[pos.Pos-self.offset][pos.Ind] = l.L
}

func (self *Seq) Column(pos int, _ bool) (c []alphabet.Letter) {
	c = make([]alphabet.Letter, self.Count())
	copy(c, self.S[pos])

	return
}

func (self *Seq) ColumnQL(pos int, _ bool) (c []alphabet.QLetter) {
	c = make([]alphabet.QLetter, self.Count())
	for i, l := range self.S[pos] {
		c[i] = alphabet.QLetter{
			L: l,
			Q: protein.DefaultQphred,
		}
	}

	return
}

func (self *Seq) Len() int { return len(self.S) }

func (self *Seq) Count() int { return len(self.S[0]) }

func (self *Seq) Offset(o int) { self.offset = o }

func (self *Seq) Start() int { return self.offset }

func (self *Seq) End() int { return self.offset + self.Len() }

func (self *Seq) Copy() seq.Sequence {
	c := *self
	c.S = make([][]alphabet.Letter, len(self.S))
	for i, s := range self.S {
		c.S[i] = append([]alphabet.Letter(nil), s...)
	}
	c.Meta = nil

	return &c
}

func (self *Seq) Reverse() { self.S = sequtils.Reverse(self.S).([][]alphabet.Letter) }

func (self *Seq) Circular(c bool) { self.circular = c }

func (self *Seq) IsCircular() bool { return self.circular }

// Return a subsequence from start to end, wrapping if the sequence is circular.
func (self *Seq) Subseq(start int, end int) (sub seq.Sequence, err error) {
	var s *Seq

	tt, err := sequtils.Truncate(self.S, start-self.offset, end-self.offset, self.circular)
	if err == nil {
		s = &Seq{}
		*s = *self
		s.S = tt.([][]alphabet.Letter)
		s.S = nil
		s.Meta = nil
		s.offset = start
		s.circular = false
	}

	return s, nil
}

func (self *Seq) Truncate(start int, end int) (err error) {
	tt, err := sequtils.Truncate(self.S, start-self.offset, end-self.offset, self.circular)
	if err == nil {
		self.S = tt.([][]alphabet.Letter)
		self.offset = start
		self.circular = false
	}

	return
}

func (self *Seq) Join(p *Seq, where int) (err error) {
	if self.circular {
		return bio.NewError("Cannot join circular sequence: receiver.", 1, self)
	} else if p.circular {
		return bio.NewError("Cannot join circular sequence: parameter.", 1, p)
	}

	var tt interface{}

	tt, self.offset = sequtils.Join(self.S, p.S, where)
	self.S = tt.([][]alphabet.Letter)

	return
}

func (self *Seq) Stitch(f feat.FeatureSet) (err error) {
	tt, err := sequtils.Stitch(self.S, self.offset, f)
	if err == nil {
		self.S = tt.([][]alphabet.Letter)
		self.circular = false
		self.offset = 0
	}

	return
}

func (self *Seq) Compose(f feat.FeatureSet) (err error) {
	tt, err := sequtils.Compose(self.S, self.offset, f)
	if err == nil {
		s := [][]alphabet.Letter{}
		for _, ts := range tt {
			s = append(s, ts.([][]alphabet.Letter)...)
		}

		self.S = s
		self.circular = false
		self.offset = 0
	}

	return
}

func (self *Seq) String() string { return self.Stringify(self) }

func (self *Seq) Consensus(_ bool) (qs *protein.QSeq) {
	cs := make([]alphabet.QLetter, 0, self.Len())
	for i := range self.S {
		cs = append(cs, self.Consensify(self, i, false))
	}

	qs = protein.NewQSeq("Consensus:"+self.ID, cs, self.alphabet, alphabet.Sanger)
	qs.Offset(self.offset)
	qs.Circular(self.circular)

	return
}
