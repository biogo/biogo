// Package alignment handles aligned sequences stored as columns.
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

// Alignment nucleic acid.
type Alignment struct {
	ID         string
	Desc       string
	Loc        string
	SubIDs     []string
	S          [][]alphabet.Letter
	Consensify nucleic.Consensifyer
	Strand     nucleic.Strand
	Stringify  seq.Stringify
	Meta       interface{} // No operation implicitly copies or changes the contents of Meta.
	alphabet   alphabet.Nucleic
	circular   bool
	offset     int
}

func NewAlignment(id string, subids []string, b [][]alphabet.Letter, alpha alphabet.Nucleic, cons nucleic.Consensifyer) (*Alignment, error) {
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

	return &Alignment{
		ID:         id,
		SubIDs:     append([]string{}, subids...),
		S:          append([][]alphabet.Letter{}, b...),
		alphabet:   alpha,
		Strand:     1,
		Consensify: cons,
		Stringify: func(s seq.Polymer) string {
			t := s.(*Alignment).Consensus(false)
			return t.String()
		},
	}, nil
}

// Interface guarantees:
var (
	_ seq.Polymer       = &Alignment{}
	_ seq.Sequence      = &Alignment{}
	_ nucleic.Sequence  = &Alignment{}
	_ nucleic.Extracter = &Alignment{}
	_ nucleic.Aligned   = &Alignment{}
	_ nucleic.AlignedAppender  = &Alignment{} // TODO
)

// Required to satisfy nucleic.Sequence interface.
func (self *Alignment) Nucleic() {}

// Raw returns a pointer to the underlying [][]byte slice.
func (self *Alignment) Raw() interface{} { return &self.S }

// Append each byte of each a to the appropriate sequence in the reciever.
func (self *Alignment) AppendColumns(a ...[]alphabet.QLetter) (err error) {
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
func (self *Alignment) AppendEach(a [][]alphabet.QLetter) (err error) {
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
func (self *Alignment) Name() *string { return &self.ID }

// Description returns a pointer to the Desc string of the sequence.
func (self *Alignment) Description() *string { return &self.Desc }

// Location returns a pointer to the Loc string of the sequence.
func (self *Alignment) Location() *string { return &self.Loc }

func (self *Alignment) column(m []nucleic.Sequence, pos int) (c []alphabet.Letter) {
	count := 0
	for _, s := range m {
		count += s.Count()
	}

	c = make([]alphabet.Letter, 0, count)

	for _, s := range m {
		if a, ok := s.(nucleic.Aligned); ok {
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
// func (self *Alignment) Delete(i int) {}

// Add sequences n to Alignment. Sequences in n must align start and end with the receiving alignment.
// Additional sequence will be clipped.
func (self *Alignment) Add(n ...nucleic.Sequence) (err error) {
	for i := self.Start(); i < self.End(); i++ {
		self.S[i] = append(self.S[i], self.column(n, i)...)
	}
	for i := range n {
		self.SubIDs = append(self.SubIDs, *n[i].Name())
	}

	return
}

func (self *Alignment) Extract(i int) nucleic.Sequence {
	s := make([]alphabet.Letter, 0, self.Len())
	for _, c := range self.S {
		s = append(s, c[i])
	}

	return nucleic.NewSeq(self.SubIDs[i], s, self.alphabet)
}

func (self *Alignment) Alphabet() alphabet.Alphabet { return self.alphabet }

func (self *Alignment) At(pos seq.Position) alphabet.QLetter {
	return alphabet.QLetter{
		L: self.S[pos.Pos-self.offset][pos.Ind],
		Q: nucleic.DefaultQphred,
	}
}

func (self *Alignment) Set(pos seq.Position, l alphabet.QLetter) {
	self.S[pos.Pos-self.offset][pos.Ind] = l.L
}

func (self *Alignment) Column(pos int, _ bool) (c []alphabet.Letter) {
	c = make([]alphabet.Letter, self.Count())
	copy(c, self.S[pos])

	return
}

func (self *Alignment) ColumnQL(pos int, _ bool) (c []alphabet.QLetter) {
	c = make([]alphabet.QLetter, self.Count())
	for i, l := range self.S[pos] {
		c[i] = alphabet.QLetter{
			L: l,
			Q: nucleic.DefaultQphred,
		}
	}

	return
}

func (self *Alignment) Len() int { return len(self.S) }

func (self *Alignment) Count() int { return len(self.S[0]) }

func (self *Alignment) Offset(o int) { self.offset = o }

func (self *Alignment) Start() int { return self.offset }

func (self *Alignment) End() int { return self.offset + self.Len() }

func (self *Alignment) Copy() seq.Sequence {
	c := *self
	c.S = make([][]alphabet.Letter, len(self.S))
	for i, s := range self.S {
		c.S[i] = append([]alphabet.Letter{}, s...)
	}
	c.Meta = nil

	return &c
}

func (self *Alignment) RevComp() {
	self.S = self.revComp(self.S, self.alphabet.ComplementTable())
	self.Strand = -self.Strand
}

func (self *Alignment) revComp(rs [][]alphabet.Letter, complement []alphabet.Letter) [][]alphabet.Letter {
	i, j := 0, len(rs)-1
	for ; i < j; i, j = i+1, j-1 {
		for s := range rs[i] {
			rs[i][s], rs[j][s] = complement[rs[j][s]], complement[rs[i][s]]
		}
	}
	if i == j {
		for s := range rs[i] {
			rs[i][s] = complement[rs[i][s]]
		}
	}

	return rs
}

func (self *Alignment) Reverse() { self.S = sequtils.Reverse(self.S).([][]alphabet.Letter) }

func (self *Alignment) Circular(c bool) { self.circular = c }

func (self *Alignment) IsCircular() bool { return self.circular }

// Return a subsequence from start to end, wrapping if the sequence is circular.
func (self *Alignment) Subseq(start int, end int) (sub seq.Sequence, err error) {
	var (
		s  *Alignment
		tt interface{}
	)

	if tt, err = sequtils.Truncate(self.S, start-self.offset, end-self.offset, self.circular); err == nil {
		s = &Alignment{}
		*s = *self
		s.S = tt.([][]alphabet.Letter)
		s.S = nil
		s.Meta = nil
		s.offset = start
		s.circular = false
	}

	return s, nil
}

func (self *Alignment) Truncate(start int, end int) (err error) {
	var tt interface{}

	if tt, err = sequtils.Truncate(self.S, start-self.offset, end-self.offset, self.circular); err == nil {
		self.S = tt.([][]alphabet.Letter)
		self.offset = start
		self.circular = false
	}

	return
}

func (self *Alignment) Join(p *Alignment, where int) (err error) {
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

func (self *Alignment) Stitch(f feat.FeatureSet) (err error) {
	var tt interface{}

	if tt, err = sequtils.Stitch(self.S, self.offset, f); err == nil {
		self.S = tt.([][]alphabet.Letter)
		self.circular = false
		self.offset = 0
	}

	return
}

func (self *Alignment) Compose(f feat.FeatureSet) (err error) {
	var tt []interface{}

	if tt, err = sequtils.Compose(self.S, self.offset, f); err == nil {
		s := [][]alphabet.Letter{}
		complement := self.alphabet.ComplementTable()
		for i, ts := range tt {
			if f[i].Strand == -1 {
				s = append(s, self.revComp(ts.([][]alphabet.Letter), complement)...)
			} else {
				s = append(s, ts.([][]alphabet.Letter)...)
			}
		}

		self.S = s
		self.circular = false
		self.offset = 0
	}

	return
}

func (self *Alignment) String() string { return self.Stringify(self) }

func (self *Alignment) Consensus(_ bool) (qs *nucleic.QSeq) {
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
