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

// A QSeq is an aligned sequence with quality scores.
type QSeq struct {
	seq.Annotation
	SubIDs     []string
	Seq        alphabet.QColumns
	Consensify seq.ConsenseFunc
	Threshold  alphabet.Qphred // Threshold for returning valid letter.
	LowQFilter seq.Filter      // How to represent below threshold letter.
	Encode     alphabet.Encoding
}

// NewSeq creates a new Seq with the given id, letter sequence and alphabet.
func NewQSeq(id string, subids []string, ql [][]alphabet.QLetter, alpha alphabet.Alphabet, enc alphabet.Encoding, cons seq.ConsenseFunc) (*QSeq, error) {
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
		return nil, errors.New("alignment: id/seq number mismatch")
	}

	return &QSeq{
		Annotation: seq.Annotation{
			ID:    id,
			Alpha: alpha,
		},
		SubIDs:     append([]string(nil), subids...),
		Seq:        append([][]alphabet.QLetter(nil), ql...),
		Encode:     enc,
		Consensify: cons,
		Threshold:  2,
		LowQFilter: linear.LowQFilter,
	}, nil
}

// Interface guarantees
var (
	_ feat.Feature = &QSeq{}
	_ seq.Sequence = &QSeq{}
	_ seq.Sequence = &QSeq{}
)

// Slice returns the sequence data as a alphabet.Slice.
func (s *QSeq) Slice() alphabet.Slice { return s.Seq }

// SetSlice sets the sequence data represented by the Seq. SetSlice will panic if sl
// is not a QColumns.
func (s *QSeq) SetSlice(sl alphabet.Slice) { s.Seq = sl.(alphabet.QColumns) }

// At returns the letter at position pos.
func (s *QSeq) At(pos seq.Position) alphabet.QLetter {
	return s.Seq[pos.Col-s.Offset][pos.Row]
}

// Set sets the letter at position pos to l.
func (s *QSeq) Set(pos seq.Position, l alphabet.QLetter) {
	s.Seq[pos.Col-s.Offset][pos.Row] = l
}

// SetE sets the quality at position pos to e to reflect the given p(Error).
func (s *QSeq) SetE(pos seq.Position, e float64) {
	s.Seq[pos.Col-s.Offset][pos.Row].Q = alphabet.Ephred(e)
}

// QEncode encodes the quality at position pos to a letter based on the sequence encoding setting.
func (s *QSeq) QEncode(pos seq.Position) byte {
	return s.Seq[pos.Col-s.Offset][pos.Row].Q.Encode(s.Encode)
}

// Encoding returns the quality encoding scheme.
func (s *QSeq) Encoding() alphabet.Encoding { return s.Encode }

// SetEncoding sets the quality encoding scheme to e.
func (s *QSeq) SetEncoding(e alphabet.Encoding) { s.Encode = e }

// EAt returns the probability of a sequence error at position pos.
func (s *QSeq) EAt(pos seq.Position) float64 {
	return s.Seq[pos.Col-s.Offset][pos.Row].Q.ProbE()
}

// Len returns the length of the alignment.
func (s *QSeq) Len() int { return len(s.Seq) }

// Rows returns the number of rows in the alignment.
func (s *QSeq) Rows() int { return s.Seq.Rows() }

// Start returns the start position of the sequence in global coordinates.
func (s *QSeq) Start() int { return s.Offset }

// End returns the end position of the sequence in global coordinates.
func (s *QSeq) End() int { return s.Offset + s.Len() }

// Copy returns a copy of the sequence.
func (s *QSeq) Copy() seq.Sequence {
	c := *s
	c.Seq = make([][]alphabet.QLetter, len(s.Seq))
	for i, s := range s.Seq {
		c.Seq[i] = append([]alphabet.QLetter(nil), s...)
	}

	return &c
}

// Return an empty sequence.
func (s *QSeq) New() seq.Sequence {
	return &QSeq{}
}

// RevComp reverse complements the sequence. RevComp will panic if the alphabet used by
// the receiver is not a Complementor.
func (s *QSeq) RevComp() {
	rs, comp := s.Seq, s.Alpha.(alphabet.Complementor).ComplementTable()
	i, j := 0, len(rs)-1
	for ; i < j; i, j = i+1, j-1 {
		for r := range rs[i] {
			rs[i][r].L, rs[j][r].L = comp[rs[j][r].L], comp[rs[i][r].L]
			rs[i][r].Q, rs[j][r].Q = rs[j][r].Q, rs[i][r].Q
		}
	}
	if i == j {
		for r := range rs[i] {
			rs[i][r].L = comp[rs[i][r].L]
		}
	}
	s.Strand = -s.Strand
}

// Reverse reverses the order of letters in the the sequence without complementing them.
func (s *QSeq) Reverse() {
	l := s.Seq
	for i, j := 0, len(l)-1; i < j; i, j = i+1, j-1 {
		l[i], l[j] = l[j], l[i]
	}
	s.Strand = seq.None
}

func (s *QSeq) String() string {
	t := s.Consensus(false)
	t.Threshold = s.Threshold
	t.LowQFilter = s.LowQFilter
	return t.String()
}

// Add sequences n to Alignment. Sequences in n must align start and end with the receiving alignment.
// Additional sequence will be clipped.
func (s *QSeq) Add(n ...seq.Sequence) error {
	for i := s.Start(); i < s.End(); i++ {
		s.Seq[i] = append(s.Seq[i], s.column(n, i)...)
	}
	for i := range n {
		s.SubIDs = append(s.SubIDs, n[i].Name())
	}

	return nil
}

func (s *QSeq) column(m []seq.Sequence, pos int) []alphabet.QLetter {
	var row int
	for _, ss := range m {
		row += rows(ss)
	}

	c := make([]alphabet.QLetter, 0, row)

	for _, r := range m {
		if a, ok := r.(seq.Aligned); ok {
			if a.Start() <= pos && pos < a.End() {
				c = append(c, a.ColumnQL(pos, true)...)
			} else {
				c = append(c, alphabet.QLetter{L: s.Alpha.Gap()}.Repeat(a.Rows())...)
			}
		} else {
			if r.Start() <= pos && pos < r.End() {
				c = append(c, r.At(seq.Position{Col: pos}))
			} else {
				c = append(c, alphabet.QLetter{L: s.Alpha.Gap()})
			}
		}
	}

	return c
}

// TODO
func (s *QSeq) Delete(i int) {}

// Get returns the sequence corresponding to the ith row of the Seq.
func (s *QSeq) Get(i int) seq.Sequence {
	t := make([]alphabet.QLetter, 0, s.Len())
	for _, c := range s.Seq {
		t = append(t, c[i])
	}

	return linear.NewQSeq(s.SubIDs[i], t, s.Alpha, s.Encode)
}

// AppendColumns appends each Qletter of each element of a to the appropriate sequence in the reciever.
func (s *QSeq) AppendColumns(a ...[]alphabet.QLetter) error {
	for i, c := range a {
		if len(c) != s.Rows() {
			return fmt.Errorf("alignment: column %d does not match Rows(): %d != %d.", i, len(c), s.Rows())
		}
	}

	s.Seq = append(s.Seq, a...)

	return nil
}

// AppendEach appends each []alphabet.QLetter in a to the appropriate sequence in the receiver.
func (s *QSeq) AppendEach(a [][]alphabet.QLetter) error {
	if len(a) != s.Rows() {
		return fmt.Errorf("alignment: number of sequences does not match Rows(): %d != %d.", len(a), s.Rows())
	}
	max := util.MinInt
	for _, r := range a {
		if l := len(r); l > max {
			max = l
		}
	}
	s.Seq = append(s.Seq, make([][]alphabet.QLetter, max)...)[:len(s.Seq)]
	for i, b := 0, make([]alphabet.QLetter, 0, len(a)); i < max; i, b = i+1, b[:0] {
		for _, r := range a {
			if i < len(r) {
				b = append(b, r[i])
			} else {
				b = append(b, alphabet.QLetter{L: s.Alpha.Gap()})
			}
		}
		s.AppendColumns(b)
	}

	return nil
}

// Column returns a slice of letters reflecting the column at pos.
func (s *QSeq) Column(pos int, _ bool) []alphabet.Letter {
	c := make([]alphabet.Letter, s.Rows())
	for i, l := range s.Seq[pos] {
		if l.Q > s.Threshold {
			c[i] = l.L
		} else {
			c[i] = s.LowQFilter(s, 0)
		}
	}

	return c
}

// ColumnQL returns a slice of quality letters reflecting the column at pos.
func (s *QSeq) ColumnQL(pos int, _ bool) []alphabet.QLetter { return s.Seq[pos] }

// Consensus returns a quality sequence reflecting the consensus of the receiver determined by the
// Consensify field.
func (s *QSeq) Consensus(_ bool) *linear.QSeq {
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
