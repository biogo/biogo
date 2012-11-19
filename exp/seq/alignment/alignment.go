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
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/feat"
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/exp/seq/linear"
	"code.google.com/p/biogo/util"
	"errors"
	"fmt"
)

type rowCounter interface {
	Rows() int
}

func rows(s seq.Sequence) int {
	row := 1
	if m, ok := s.(rowCounter); ok {
		row = m.Rows()
	}
	return row
}

// A Seq is an aligned sequence.
type Seq struct {
	seq.Annotation
	SubIDs     []string
	Seq        alphabet.Columns
	Consensify seq.ConsenseFunc
}

// NewSeq creates a new Seq with the given id, letter sequence and alphabet.
func NewSeq(id string, subids []string, b [][]alphabet.Letter, alpha alphabet.Alphabet, cons seq.ConsenseFunc) (*Seq, error) {
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
		return nil, errors.New("alignment: id/seq number mismatch")
	}

	return &Seq{
		Annotation: seq.Annotation{
			ID:    id,
			Alpha: alpha,
		},
		SubIDs:     append([]string(nil), subids...),
		Seq:        append([][]alphabet.Letter(nil), b...),
		Consensify: cons,
	}, nil
}

// Interface guarantees
var (
	_ feat.Feature = &Seq{}
	_ seq.Sequence = &Seq{}
	_ seq.Sequence = &Seq{}
)

// Slice returns the sequence data as a alphabet.Slice.
func (s *Seq) Slice() alphabet.Slice { return s.Seq }

// SetSlice sets the sequence data represented by the Seq. SetSlice will panic if sl
// is not a Columns.
func (s *Seq) SetSlice(sl alphabet.Slice) { s.Seq = sl.(alphabet.Columns) }

// At returns the letter at position pos.
func (s *Seq) At(pos seq.Position) alphabet.QLetter {
	return alphabet.QLetter{
		L: s.Seq[pos.Col-s.Offset][pos.Row],
		Q: seq.DefaultQphred,
	}
}

// Set sets the letter at position pos to l.
func (s *Seq) Set(pos seq.Position, l alphabet.QLetter) {
	s.Seq[pos.Col-s.Offset][pos.Row] = l.L
}

// Len returns the length of the alignment.
func (s *Seq) Len() int { return len(s.Seq) }

// Rows returns the number of rows in the alignment.
func (s *Seq) Rows() int { return s.Seq.Rows() }

// Start returns the start position of the sequence in global coordinates.
func (s *Seq) Start() int { return s.Offset }

// End returns the end position of the sequence in global coordinates.
func (s *Seq) End() int { return s.Offset + s.Len() }

// Copy returns a copy of the sequence.
func (s *Seq) Copy() seq.Sequence {
	c := *s
	c.Seq = make(alphabet.Columns, len(s.Seq))
	for i, cs := range s.Seq {
		c.Seq[i] = append([]alphabet.Letter(nil), cs...)
	}

	return &c
}

// New returns an empty *Seq sequence.
func (s *Seq) New() seq.Sequence {
	return &Seq{}
}

// RevComp reverse complements the sequence. RevComp will panic if the alphabet used by
// the receiver is not a Complementor.
func (s *Seq) RevComp() {
	rs, comp := s.Seq, s.Alpha.(alphabet.Complementor).ComplementTable()
	i, j := 0, len(rs)-1
	for ; i < j; i, j = i+1, j-1 {
		for r := range rs[i] {
			rs[i][r], rs[j][r] = comp[rs[j][r]], comp[rs[i][r]]
		}
	}
	if i == j {
		for r := range rs[i] {
			rs[i][r] = comp[rs[i][r]]
		}
	}
	s.Strand = -s.Strand
}

// Reverse reverses the order of letters in the the sequence without complementing them.
func (s *Seq) Reverse() {
	l := s.Seq
	for i, j := 0, len(l)-1; i < j; i, j = i+1, j-1 {
		l[i], l[j] = l[j], l[i]
	}
	s.Strand = seq.None
}

func (s *Seq) String() string {
	return s.Consensus(false).String()
}

// Add adds the sequences n to Seq. Sequences in n should align start and end with the receiving alignment.
// Additional sequence will be clipped and missing sequence will be filled with the gap letter.
func (s *Seq) Add(n ...seq.Sequence) error {
	for i := s.Start(); i < s.End(); i++ {
		s.Seq[i] = append(s.Seq[i], s.column(n, i)...)
	}
	for i := range n {
		s.SubIDs = append(s.SubIDs, n[i].Name())
	}

	return nil
}

func (s *Seq) column(m []seq.Sequence, pos int) []alphabet.Letter {
	row := 0
	for _, ss := range m {
		row += rows(ss)
	}

	c := make([]alphabet.Letter, 0, row)

	for _, ss := range m {
		if a, ok := ss.(seq.Aligned); ok {
			if a.Start() <= pos && pos < a.End() {
				c = append(c, a.Column(pos, true)...)
			} else {
				c = append(c, s.Alpha.Gap().Repeat(a.Rows())...)
			}
		} else {
			if ss.Start() <= pos && pos < ss.End() {
				c = append(c, ss.At(seq.Position{Col: pos}).L)
			} else {
				c = append(c, s.Alpha.Gap())
			}
		}
	}

	return c
}

// TODO
func (s *Seq) Delete(i int) {}

// Get returns the sequence corresponding to the ith row of the Seq.
func (s *Seq) Get(i int) seq.Sequence {
	c := make([]alphabet.Letter, 0, s.Len())
	for _, l := range s.Seq {
		c = append(c, l[i])
	}

	return linear.NewSeq(s.SubIDs[i], c, s.Alpha)
}

// AppendColumns appends each Qletter of each element of a to the appropriate sequence in the reciever.
func (s *Seq) AppendColumns(a ...[]alphabet.QLetter) error {
	for i, r := range a {
		if len(r) != s.Rows() {
			return fmt.Errorf("alignment: column %d does not match Rows(): %d != %d.", i, len(r), s.Rows())
		}
	}

	s.Seq = append(s.Seq, make([][]alphabet.Letter, len(a))...)[:len(s.Seq)]
	for _, r := range a {
		c := make([]alphabet.Letter, len(r))
		for i := range r {
			c[i] = r[i].L
		}
		s.Seq = append(s.Seq, c)
	}

	return nil
}

// AppendEach appends each []alphabet.QLetter in a to the appropriate sequence in the receiver.
func (s *Seq) AppendEach(a [][]alphabet.QLetter) error {
	if len(a) != s.Rows() {
		return fmt.Errorf("alignment: number of sequences does not match Rows(): %d != %d.", len(a), s.Rows())
	}
	max := util.MinInt
	for _, ss := range a {
		if l := len(ss); l > max {
			max = l
		}
	}
	s.Seq = append(s.Seq, make([][]alphabet.Letter, max)...)[:len(s.Seq)]
	for i, b := 0, make([]alphabet.QLetter, 0, len(a)); i < max; i, b = i+1, b[:0] {
		for _, ss := range a {
			if i < len(ss) {
				b = append(b, ss[i])
			} else {
				b = append(b, alphabet.QLetter{L: s.Alpha.Gap()})
			}
		}
		s.AppendColumns(b)
	}

	return nil
}

// Column returns a slice of letters reflecting the column at pos.
func (s *Seq) Column(pos int, _ bool) []alphabet.Letter {
	return s.Seq[pos]
}

// ColumnQL returns a slice of quality letters reflecting the column at pos.
func (s *Seq) ColumnQL(pos int, _ bool) []alphabet.QLetter {
	c := make([]alphabet.QLetter, s.Rows())
	for i, l := range s.Seq[pos] {
		c[i] = alphabet.QLetter{
			L: l,
			Q: seq.DefaultQphred,
		}
	}

	return c
}

// Consensus returns a quality sequence reflecting the consensus of the receiver determined by the
// Consensify field.
func (s *Seq) Consensus(_ bool) *linear.QSeq {
	cs := make([]alphabet.QLetter, 0, s.Len())
	alpha := s.Alphabet()
	for i := range s.Seq {
		cs = append(cs, s.Consensify(s, alpha, i, false))
	}

	qs := linear.NewQSeq("Consensus:"+s.ID, cs, s.Alpha, alphabet.Sanger)
	qs.Strand = s.Strand
	qs.SetOffset(s.Offset)
	qs.Conform = s.Conform

	return qs
}
