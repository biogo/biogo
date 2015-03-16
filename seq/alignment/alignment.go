// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package alignment handles aligned sequences stored as columns.
package alignment

import (
	"github.com/biogo/biogo/alphabet"
	"github.com/biogo/biogo/feat"
	"github.com/biogo/biogo/seq"
	"github.com/biogo/biogo/seq/linear"
	"github.com/biogo/biogo/util"

	"errors"
	"fmt"
	"strings"
	"unicode"
)

// A Seq is an aligned sequence.
type Seq struct {
	seq.Annotation
	SubAnnotations []seq.Annotation
	Seq            alphabet.Columns
	ColumnConsense seq.ConsenseFunc
}

// NewSeq creates a new Seq with the given id, letter sequence and alphabet.
func NewSeq(id string, subids []string, b [][]alphabet.Letter, alpha alphabet.Alphabet, cons seq.ConsenseFunc) (*Seq, error) {
	var (
		lids, lseq = len(subids), len(b)
		subann     []seq.Annotation
	)
	switch {
	case lids == 0 && len(b) == 0:
	case lseq != 0 && lids == len(b[0]):
		if lids == 0 {
			subann = make([]seq.Annotation, len(b[0]))
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

	return &Seq{
		Annotation: seq.Annotation{
			ID:    id,
			Alpha: alpha,
		},
		SubAnnotations: subann,
		Seq:            append([][]alphabet.Letter(nil), b...),
		ColumnConsense: cons,
	}, nil
}

// Interface guarantees
var (
	_ feat.Feature = (*Seq)(nil)
	_ feat.Feature = Row{}
	_ seq.Sequence = Row{}
)

// Slice returns the sequence data as a alphabet.Slice.
func (s *Seq) Slice() alphabet.Slice { return s.Seq }

// SetSlice sets the sequence data represented by the Seq. SetSlice will panic if sl
// is not a Columns.
func (s *Seq) SetSlice(sl alphabet.Slice) { s.Seq = sl.(alphabet.Columns) }

// Len returns the length of the alignment.
func (s *Seq) Len() int { return len(s.Seq) }

// Rows returns the number of rows in the alignment.
func (s *Seq) Rows() int { return s.Seq.Rows() }

// Start returns the start position of the sequence in coordinates relative to the
// sequence location.
func (s *Seq) Start() int { return s.Offset }

// End returns the end position of the sequence in coordinates relative to the
// sequence location.
func (s *Seq) End() int { return s.Offset + s.Len() }

// Clone returns a copy of the sequence.
func (s *Seq) Clone() seq.Rower {
	c := *s
	c.Seq = make(alphabet.Columns, len(s.Seq))
	for i, cs := range s.Seq {
		c.Seq[i] = append([]alphabet.Letter(nil), cs...)
	}

	return &c
}

// New returns an empty *Seq sequence with the same alphabet.
func (s *Seq) New() *Seq {
	return &Seq{Annotation: seq.Annotation{Alpha: s.Alpha}}
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
		s.SubAnnotations = append(s.SubAnnotations, *n[i].CloneAnnotation())
	}

	return nil
}

func (s *Seq) column(m []seq.Sequence, pos int) []alphabet.Letter {
	c := make([]alphabet.Letter, 0, s.Rows())

	for _, ss := range m {
		if a, ok := ss.(seq.Aligned); ok {
			if a.Start() <= pos && pos < a.End() {
				c = append(c, a.Column(pos, true)...)
			} else {
				c = append(c, s.Alpha.Gap().Repeat(a.Rows())...)
			}
		} else {
			if ss.Start() <= pos && pos < ss.End() {
				c = append(c, ss.At(pos).L)
			} else {
				c = append(c, s.Alpha.Gap())
			}
		}
	}

	return c
}

// Delete removes the sequence represented at row i of the alignment. It panics if i is out of range.
func (s *Seq) Delete(i int) {
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
func (s *Seq) Row(i int) seq.Sequence {
	if i < 0 || i >= s.Rows() {
		panic("alignment: index out of range")
	}
	return Row{Align: s, Row: i}
}

// AppendColumns appends each Qletter of each element of a to the appropriate sequence in the receiver.
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
// ColumnConsense field.
func (s *Seq) Consensus(_ bool) *linear.QSeq {
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
func (s *Seq) Format(fs fmt.State, c rune) {
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
		r := Row{Align: s}
		for r.Row = 0; r.Row < s.Rows(); r.Row++ {
			r.Format(fs, c)
			if r.Row < s.Rows()-1 {
				fmt.Fprintln(fs)
			}
		}
	default:
		fmt.Fprintf(fs, "%%!%c(*alignment.Seq=%.10s)", c, s)
	}
}

// A Row is a pointer into an alignment that satisfies the seq.Sequence interface.
type Row struct {
	Align *Seq
	Row   int
}

// At returns the letter at position i.
func (r Row) At(i int) alphabet.QLetter {
	return alphabet.QLetter{
		L: r.Align.Seq[i-r.Align.Offset][r.Row],
		Q: seq.DefaultQphred,
	}
}

// Set sets the letter at position i to l.
func (r Row) Set(i int, l alphabet.QLetter) error {
	r.Align.Seq[i-r.Align.Offset][r.Row] = l.L
	return nil
}

// Len returns the length of the row.
func (r Row) Len() int { return len(r.Align.Seq) }

// Start returns the start position of the sequence in coordinates relative to the
// sequence location.
func (r Row) Start() int { return r.Align.SubAnnotations[r.Row].Offset }

// End returns the end position of the sequence in coordinates relative to the
// sequence location.
func (r Row) End() int { return r.Start() + r.Len() }

// Location returns the feature containing the row's sequence.
func (r Row) Location() feat.Feature { return r.Align.SubAnnotations[r.Row].Loc }

func (r Row) Alphabet() alphabet.Alphabet     { return r.Align.Alpha }
func (r Row) Conformation() feat.Conformation { return r.Align.Conform }
func (r Row) SetConformation(c feat.Conformation) error {
	r.Align.SubAnnotations[r.Row].Conform = c
	return nil
}
func (r Row) Name() string {
	return r.Align.SubAnnotations[r.Row].ID
}
func (r Row) Description() string   { return r.Align.SubAnnotations[r.Row].Desc }
func (r Row) SetOffset(o int) error { r.Align.SubAnnotations[r.Row].Offset = o; return nil }

func (r Row) RevComp() {
	rs, comp := r.Align.Seq, r.Alphabet().(alphabet.Complementor).ComplementTable()
	i, j := 0, len(rs)-1
	for ; i < j; i, j = i+1, j-1 {
		rs[i][r.Row], rs[j][r.Row] = comp[rs[j][r.Row]], comp[rs[i][r.Row]]
	}
	if i == j {
		rs[i][r.Row] = comp[rs[i][r.Row]]
	}
	r.Align.SubAnnotations[r.Row].Strand = -r.Align.SubAnnotations[r.Row].Strand
}
func (r Row) Reverse() {
	l := r.Align.Seq
	for i, j := 0, len(l)-1; i < j; i, j = i+1, j-1 {
		l[i][r.Row], l[j][r.Row] = l[j][r.Row], l[i][r.Row]
	}
	r.Align.SubAnnotations[r.Row].Strand = seq.None
}
func (r Row) New() seq.Sequence {
	return Row{Align: &Seq{Annotation: seq.Annotation{Alpha: r.Align.Alpha}}}
}
func (r Row) Clone() seq.Sequence {
	b := make([]alphabet.Letter, r.Len())
	for i, c := range r.Align.Seq {
		b[i] = c[r.Row]
	}
	switch {
	case r.Row < 0:
		panic("under")
	case r.Row >= r.Align.Rows():
		panic("bang over Rows()")
	case r.Row >= len(r.Align.SubAnnotations):

		panic(fmt.Sprintf("bang over len(SubAnns): %d %d", r.Row, len(r.Align.SubAnnotations)))
	}
	return linear.NewSeq(r.Name(), b, r.Alphabet())
}
func (r Row) CloneAnnotation() *seq.Annotation { return r.Align.SubAnnotations[r.Row].CloneAnnotation() }

// String returns a string representation of the sequence data only.
func (r Row) String() string { return fmt.Sprintf("%-s", r) }

// Format is a support routine for fmt.Formatter. It accepts the formats 'v' and 's'
// (string), 'a' (fasta) and 'q' (fastq). String, fasta and fastq formats support
// truncated output via the verb's precision. Fasta format supports sequence line
// specification via the verb's width field. Fastq format supports optional inclusion
// of the '+' line descriptor line with the '+' flag. The 'v' verb supports the '#'
// flag for Go syntax output. The 's' and 'v' formats support the '-' flag for
// omission of the sequence name.
func (r Row) Format(fs fmt.State, c rune) {
	var (
		s      = r.Align
		w, wOk = fs.Width()
		p, pOk = fs.Precision()
		buf    alphabet.Columns
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
			type shadowRow Row
			sr := fmt.Sprintf("%#v", shadowRow(r))
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
		for _, lc := range buf {
			fmt.Fprintf(fs, "%c", lc[r.Row])
		}
		if pOk && s != nil && p < s.Len() {
			fmt.Fprint(fs, "...")
		}
	case 'a':
		if s == nil {
			return
		}
		r.formatDescLineTo(fs, '>')
		for i, lc := range buf {
			fmt.Fprintf(fs, "%c", lc[r.Row])
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
		for _, lc := range buf {
			fmt.Fprintf(fs, "%c", lc[r.Row])
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
		e := seq.DefaultQphred.Encode(seq.DefaultEncoding)
		if e >= unicode.MaxASCII {
			e = unicode.MaxASCII - 1
		}
		for _ = range buf {
			fmt.Fprintf(fs, "%c", e)
		}
		if pOk && p < s.Len() {
			fmt.Fprint(fs, "...")
		}
	default:
		fmt.Fprintf(fs, "%%!%c(alignment.Row=%.10s)", c, s)
	}
}

func (r Row) formatDescLineTo(fs fmt.State, p rune) {
	fmt.Fprintf(fs, "%c%s", p, r.Name())
	if d := r.Description(); d != "" {
		fmt.Fprintf(fs, " %s", d)
	}
	fmt.Fprintln(fs)
}

// SetSlice unconditionally panics.
func (r Row) SetSlice(_ alphabet.Slice) { panic("alignment: cannot alter row slice") }

// Slice unconditionally panics.
func (r Row) Slice() alphabet.Slice { panic("alignment: cannot get row slice") }
