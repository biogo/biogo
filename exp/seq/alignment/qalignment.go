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
	"strings"
	"unicode"
)

// A QSeq is an aligned sequence with quality scores.
type QSeq struct {
	seq.Annotation
	SubAnnotations []seq.Annotation
	Seq            alphabet.QColumns
	ColumnConsense seq.ConsenseFunc
	Threshold      alphabet.Qphred // Threshold for returning valid letter.
	QFilter        seq.QFilter     // How to represent below threshold letter.
	Encode         alphabet.Encoding
}

// NewSeq creates a new Seq with the given id, letter sequence and alphabet.
func NewQSeq(id string, subids []string, ql [][]alphabet.QLetter, alpha alphabet.Alphabet, enc alphabet.Encoding, cons seq.ConsenseFunc) (*QSeq, error) {
	var (
		lids, lseq = len(subids), len(ql)
		subann     []seq.Annotation
	)
	switch {
	case lids == 0 && len(ql) == 0:
	case lseq != 0 && lids == len(ql[0]):
		if lids == 0 {
			subann = make([]seq.Annotation, len(ql[0]))
			for i := range subids {
				subann[i].ID = fmt.Sprintf("%s:%d", id, i)
			}
		} else {
			subann = make([]seq.Annotation, lids)
			for i, sid := range subids {
				subann[i].ID = sid
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
		SubAnnotations: subann,
		Seq:            append([][]alphabet.QLetter(nil), ql...),
		Encode:         enc,
		ColumnConsense: cons,
		Threshold:      2,
		QFilter:        seq.AmbigFilter,
	}, nil
}

// Interface guarantees
var (
	_ feat.Feature = &QSeq{}
	_ feat.Feature = QRow{}
	_ seq.Sequence = QRow{}
	_ seq.Scorer   = QRow{}
)

// Slice returns the sequence data as a alphabet.Slice.
func (s *QSeq) Slice() alphabet.Slice { return s.Seq }

// SetSlice sets the sequence data represented by the Seq. SetSlice will panic if sl
// is not a QColumns.
func (s *QSeq) SetSlice(sl alphabet.Slice) { s.Seq = sl.(alphabet.QColumns) }

// Encoding returns the quality encoding scheme.
func (s *QSeq) Encoding() alphabet.Encoding { return s.Encode }

// SetEncoding sets the quality encoding scheme to e.
func (s *QSeq) SetEncoding(e alphabet.Encoding) { s.Encode = e }

// Len returns the length of the alignment.
func (s *QSeq) Len() int { return len(s.Seq) }

// Rows returns the number of rows in the alignment.
func (s *QSeq) Rows() int { return s.Seq.Rows() }

// Start returns the start position of the sequence in global coordinates.
func (s *QSeq) Start() int { return s.Offset }

// End returns the end position of the sequence in global coordinates.
func (s *QSeq) End() int { return s.Offset + s.Len() }

// Clone returns a copy of the sequence.
func (s *QSeq) Clone() seq.Rower {
	c := *s
	c.Seq = make([][]alphabet.QLetter, len(s.Seq))
	for i, s := range s.Seq {
		c.Seq[i] = append([]alphabet.QLetter(nil), s...)
	}

	return &c
}

// Return an empty sequence.
func (s *QSeq) New() *QSeq {
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
	t.QFilter = s.QFilter
	return t.String()
}

// Add sequences n to Alignment. Sequences in n must align start and end with the receiving alignment.
// Additional sequence will be clipped.
func (s *QSeq) Add(n ...seq.Sequence) error {
	for i := s.Start(); i < s.End(); i++ {
		s.Seq[i] = append(s.Seq[i], s.column(n, i)...)
	}
	for i := range n {
		s.SubAnnotations = append(s.SubAnnotations, *n[i].CloneAnnotation())
	}

	return nil
}

func (s *QSeq) column(m []seq.Sequence, pos int) []alphabet.QLetter {
	c := make([]alphabet.QLetter, 0, s.Rows())

	for _, r := range m {
		if a, ok := r.(seq.Aligned); ok {
			if a.Start() <= pos && pos < a.End() {
				c = append(c, a.ColumnQL(pos, true)...)
			} else {
				c = append(c, alphabet.QLetter{L: s.Alpha.Gap()}.Repeat(a.Rows())...)
			}
		} else {
			if r.Start() <= pos && pos < r.End() {
				c = append(c, r.At(pos))
			} else {
				c = append(c, alphabet.QLetter{L: s.Alpha.Gap()})
			}
		}
	}

	return c
}

// Delete removes the sequence represented at row i of the alignment. It panics if i is out of range.
func (s *QSeq) Delete(i int) {
	if i >= s.Rows() {
		panic("alignment: index out of range")
	}
	cs := s.Seq
	for j, c := range cs {
		cs[j] = c[:i+copy(c[i:], c[i+1:])]
	}
	sa := s.SubAnnotations
	s.SubAnnotations = sa[:i+copy(sa[i:], sa[i+1:])]
}

// Row returns the sequence represented at row i of the alignment. It panics is i is out of range.
func (s *QSeq) Row(i int) seq.Sequence {
	if i < 0 || i >= s.Rows() {
		panic("alignment: index out of range")
	}
	return QRow{Align: s, Row: i}
}

// AppendColumns appends each Qletter of each element of a to the appropriate sequence in the receiver.
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
		if l.Q >= s.Threshold {
			c[i] = l.L
		} else {
			c[i] = s.QFilter(s.Alpha, 255, alphabet.QLetter{})
		}
	}

	return c
}

// ColumnQL returns a slice of quality letters reflecting the column at pos.
func (s *QSeq) ColumnQL(pos int, _ bool) []alphabet.QLetter { return s.Seq[pos] }

// Consensus returns a quality sequence reflecting the consensus of the receiver determined by the
// ColumnConsense field.
func (s *QSeq) Consensus(_ bool) *linear.QSeq {
	cs := make([]alphabet.QLetter, 0, s.Len())
	alpha := s.Alphabet()
	for i := range s.Seq {
		cs = append(cs, s.ColumnConsense(s, alpha, i, false))
	}

	qs := linear.NewQSeq("Consensus:"+s.ID, cs, s.Alpha, alphabet.Sanger)
	qs.Strand = s.Strand
	qs.SetOffset(s.Offset)
	qs.Conform = s.Conform

	return qs
}

// Format is a support routine for fmt.Formatter. It accepts the formats 'v' and 's'
// (string), 'a' (fasta) and 'q' (fastq). String, fasta and fastq formats support
// truncated output via the verb's precision. Fasta format supports sequence line
// specification via the verb's width field. Fastq format supports optional inclusion
// of the '+' line descriptor line with the '+' flag. The 'v' verb supports the '#'
// flag for Go syntax output. The 's' and 'v' formats support the '-' flag for
// omission of the sequence name.
func (s *QSeq) Format(fs fmt.State, c rune) {
	if s == nil {
		fmt.Fprint(fs, "<nil>")
		return
	}
	switch c {
	case 'v':
		if fs.Flag('#') {
			fmt.Fprintf(fs, "&%#v", *s)
			return
		}
		fallthrough
	case 's', 'a', 'q':
		r := QRow{Align: s}
		for r.Row = 0; r.Row < s.Rows(); r.Row++ {
			r.Format(fs, c)
			if r.Row < s.Rows()-1 {
				fmt.Fprintln(fs)
			}
		}
	default:
		fmt.Fprintf(fs, "%%!%c(*alignment.QSeq=%.10s)", c, s)
	}
}

// A Row is a pointer into an alignment that satisfies the seq.Sequence and seq.Scorer interfaces.
type QRow struct {
	Align *QSeq
	Row   int
}

// At returns the letter at position i.
func (r QRow) At(i int) alphabet.QLetter {
	return r.Align.Seq[i-r.Align.Offset][r.Row]
}

// Set sets the letter at position i to l.
func (r QRow) Set(i int, l alphabet.QLetter) {
	r.Align.Seq[i-r.Align.Offset][r.Row] = l
}

// Len returns the length of the alignment.
func (r QRow) Len() int { return len(r.Align.Seq) }

// Start returns the start position of the sequence in global coordinates.
func (r QRow) Start() int { return r.Align.SubAnnotations[r.Row].Offset }

// End returns the end position of the sequence in global coordinates.
func (r QRow) End() int { return r.Start() + r.Len() }

// Location returns the feature containing the row's sequence.
func (r QRow) Location() feat.Feature { return r.Align.SubAnnotations[r.Row].Loc }

// SetE sets the quality at position i to e to reflect the given p(Error).
func (r QRow) SetE(i int, e float64) {
	r.Align.Seq[i-r.Align.Offset][r.Row].Q = alphabet.Ephred(e)
}

// QEncode encodes the quality at position i to a letter based on the sequence encoding setting.
func (r QRow) QEncode(i int) byte {
	return r.Align.Seq[i-r.Align.Offset][r.Row].Q.Encode(r.Encoding())
}

// EAt returns the probability of a sequence error at position i.
func (r QRow) EAt(i int) float64 {
	return r.Align.Seq[i-r.Align.Offset][r.Row].Q.ProbE()
}

func (r QRow) Alphabet() alphabet.Alphabet         { return r.Align.Alpha }
func (r QRow) Encoding() alphabet.Encoding         { return r.Align.Encoding() }
func (r QRow) SetEncoding(e alphabet.Encoding)     { r.Align.SetEncoding(e) }
func (r QRow) Conformation() feat.Conformation     { return r.Align.SubAnnotations[r.Row].Conform }
func (r QRow) SetConformation(c feat.Conformation) { r.Align.SubAnnotations[r.Row].Conform = c }
func (r QRow) Name() string                        { return r.Align.SubAnnotations[r.Row].ID }
func (r QRow) Description() string                 { return r.Align.SubAnnotations[r.Row].Desc }
func (r QRow) SetOffset(o int)                     { r.Align.SubAnnotations[r.Row].Offset = o }

func (r QRow) RevComp() {
	rs, comp := r.Align.Seq, r.Alphabet().(alphabet.Complementor).ComplementTable()
	i, j := 0, len(rs)-1
	for ; i < j; i, j = i+1, j-1 {
		rs[i][r.Row].L, rs[j][r.Row].L = comp[rs[j][r.Row].L], comp[rs[i][r.Row].L]
		rs[i][r.Row].Q, rs[j][r.Row].Q = rs[j][r.Row].Q, rs[i][r.Row].Q
	}
	if i == j {
		rs[i][r.Row].L = comp[rs[i][r.Row].L]
	}
	r.Align.SubAnnotations[r.Row].Strand = -r.Align.SubAnnotations[r.Row].Strand
}
func (r QRow) Reverse() {
	l := r.Align.Seq
	for i, j := 0, len(l)-1; i < j; i, j = i+1, j-1 {
		l[i][r.Row], l[j][r.Row] = l[j][r.Row], l[i][r.Row]
	}
	r.Align.SubAnnotations[r.Row].Strand = seq.None
}
func (r QRow) New() seq.Sequence { return QRow{} }
func (r QRow) Clone() seq.Sequence {
	b := make([]alphabet.QLetter, r.Len())
	for i, c := range r.Align.Seq {
		b[i] = c[r.Row]
	}
	return linear.NewQSeq(r.Name(), b, r.Alphabet(), r.Align.Encoding())
}
func (r QRow) CloneAnnotation() *seq.Annotation {
	return r.Align.SubAnnotations[r.Row].CloneAnnotation()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// String returns a string representation of the sequence data only.
func (r QRow) String() string { return fmt.Sprintf("%-s", r) }

// Format is a support routine for fmt.Formatter. It accepts the formats 'v' and 's'
// (string), 'a' (fasta) and 'q' (fastq). String, fasta and fastq formats support
// truncated output via the verb's precision. Fasta format supports sequence line
// specification via the verb's width field. Fastq format supports optional inclusion
// of the '+' line descriptor line with the '+' flag. The 'v' verb supports the '#'
// flag for Go syntax output. The 's' and 'v' formats support the '-' flag for
// omission of the sequence name.
func (r QRow) Format(fs fmt.State, c rune) {
	var (
		s      = r.Align
		w, wOk = fs.Width()
		p, pOk = fs.Precision()
		buf    alphabet.QColumns
	)
	if s != nil {
		if pOk {
			buf = s.Seq[:min(p, len(s.Seq))]
		} else {
			buf = s.Seq
		}
	}

	switch c {
	case 'v':
		if fs.Flag('#') {
			type shadowQRow QRow
			sr := fmt.Sprintf("%#v", shadowQRow(r))
			fmt.Fprintf(fs, "%T%s", r, sr[strings.Index(sr, "{"):])
			return
		}
		fallthrough
	case 's':
		if s == nil {
			fmt.Fprint(fs, "<nil>")
			return
		}
		if !fs.Flag('-') {
			fmt.Fprintf(fs, "%q ", r.Name())
		}
		for _, qc := range buf {
			fmt.Fprintf(fs, "%c", s.QFilter(s.Alpha, s.Threshold, qc[r.Row]))
		}
		if pOk && p < s.Len() {
			fmt.Fprint(fs, "...")
		}
	case 'a':
		if s == nil {
			return
		}
		r.formatDescLineTo(fs, '>')
		for i, qc := range buf {
			fmt.Fprintf(fs, "%c", s.QFilter(s.Alpha, s.Threshold, qc[r.Row]))
			if wOk && i < s.Len()-1 && i%w == w-1 {
				fmt.Fprintln(fs)
			}
		}
		if pOk && p < s.Len() {
			fmt.Fprint(fs, "...")
		}
	case 'q':
		if s == nil {
			return
		}
		r.formatDescLineTo(fs, '@')
		for _, qc := range buf {
			fmt.Fprintf(fs, "%c", qc[r.Row].L)
		}
		if pOk && p < s.Len() {
			fmt.Fprintln(fs, "...")
		} else {
			fmt.Fprintln(fs)
		}
		if fs.Flag('+') {
			r.formatDescLineTo(fs, '+')
		} else {
			fmt.Fprintln(fs, "+")
		}
		for _, qc := range buf {
			e := qc[r.Row].Q.Encode(s.Encode)
			if e >= unicode.MaxASCII {
				e = unicode.MaxASCII - 1
			}
			fmt.Fprintf(fs, "%c", e)
		}
		if pOk && p < s.Len() {
			fmt.Fprint(fs, "...")
		}
	default:
		fmt.Fprintf(fs, "%%!%c(alignment.QRow=%.10s)", c, s)
	}
}

func (r QRow) formatDescLineTo(fs fmt.State, p rune) {
	fmt.Fprintf(fs, "%c%s", p, r.Name())
	if d := r.Description(); d != "" {
		fmt.Fprintf(fs, " %s", d)
	}
	fmt.Fprintln(fs)
}

// SetSlice unconditionally panics.
func (r QRow) SetSlice(_ alphabet.Slice) { panic("alignment: cannot alter row slice") }

// Slice unconditionally panics.
func (r QRow) Slice() alphabet.Slice { panic("alignment: cannot get row slice") }
