// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linear

import (
	"code.google.com/p/biogo/alphabet"
	"code.google.com/p/biogo/feat"
	"code.google.com/p/biogo/seq"

	"fmt"
	"unicode"
)

// A QSeq is a basic linear sequence with Phred quality scores.
type QSeq struct {
	seq.Annotation
	Seq       alphabet.QLetters
	Threshold alphabet.Qphred // Threshold for returning valid letter.
	QFilter   seq.QFilter     // How to represent below threshold letter.
	Encode    alphabet.Encoding
}

// Interface guarantees
var (
	_ feat.Feature = (*QSeq)(nil)
	_ seq.Sequence = (*QSeq)(nil)
	_ seq.Scorer   = (*QSeq)(nil)
)

// NewQSeq create a new QSeq with the given id, letter sequence, alphabet and quality encoding.
func NewQSeq(id string, ql []alphabet.QLetter, alpha alphabet.Alphabet, enc alphabet.Encoding) *QSeq {
	return &QSeq{
		Annotation: seq.Annotation{
			ID:     id,
			Alpha:  alpha,
			Strand: seq.Plus,
		},
		Seq:       append(alphabet.QLetters(nil), ql...),
		Encode:    enc,
		Threshold: 3,
		QFilter:   seq.AmbigFilter,
	}
}

// Append append Letters to the sequence, the DefaultQphred value is used for quality scores.
func (s *QSeq) AppendLetters(a ...alphabet.Letter) error {
	l := s.Len()
	s.Seq = append(s.Seq, make([]alphabet.QLetter, len(a))...)[:l]
	for _, v := range a {
		s.Seq = append(s.Seq, alphabet.QLetter{L: v, Q: seq.DefaultQphred})
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
func (s *QSeq) At(i int) alphabet.QLetter {
	return s.Seq[i-s.Offset]
}

// QEncode encodes the quality at position pos to a letter based on the sequence encoding setting.
func (s *QSeq) QEncode(i int) byte {
	return s.Seq[i-s.Offset].Q.Encode(s.Encode)
}

// Encoding returns the quality encoding scheme.
func (s *QSeq) Encoding() alphabet.Encoding { return s.Encode }

// SetEncoding sets the quality encoding scheme to e.
func (s *QSeq) SetEncoding(e alphabet.Encoding) error { s.Encode = e; return nil }

// EAt returns the probability of a sequence error at position pos.
func (s *QSeq) EAt(i int) float64 {
	return s.Seq[i-s.Offset].Q.ProbE()
}

// Set sets the letter at position pos to l.
func (s *QSeq) Set(i int, l alphabet.QLetter) error {
	s.Seq[i-s.Offset] = l
	return nil
}

// SetE sets the quality at position pos to e to reflect the given p(Error).
func (s *QSeq) SetE(i int, e float64) error {
	s.Seq[i-s.Offset].Q = alphabet.Ephred(e)
	return nil
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

// Clone returns a copy of the sequence.
func (s *QSeq) Clone() seq.Sequence {
	c := *s
	c.Seq = append([]alphabet.QLetter(nil), s.Seq...)

	return &c
}

// New returns an empty *QSeq sequence with the same alphabet.
func (s *QSeq) New() seq.Sequence {
	return &QSeq{Annotation: seq.Annotation{Alpha: s.Alpha}}
}

// RevComp reverse complements the sequence. RevComp will panic if the alphabet used by
// the receiver is not a Complementor.
func (s *QSeq) RevComp() {
	l, comp := s.Seq, s.Alphabet().(alphabet.Complementor).ComplementTable()
	i, j := 0, len(l)-1
	for ; i < j; i, j = i+1, j-1 {
		l[i].L, l[j].L = comp[l[j].L], comp[l[i].L]
		l[i].Q, l[j].Q = l[j].Q, l[i].Q
	}
	if i == j {
		l[i].L = comp[l[i].L]
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

// String returns a string representation of the sequence data only.
func (s *QSeq) String() string {
	cs := make([]alphabet.Letter, 0, len(s.Seq))
	for _, ql := range s.Seq {
		cs = append(cs, s.QFilter(s.Alpha, s.Threshold, ql))
	}

	return alphabet.Letters(cs).String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
	var (
		w, wOk = fs.Width()
		p, pOk = fs.Precision()
		buf    []alphabet.QLetter
	)
	if pOk {
		buf = s.Seq[:min(p, len(s.Seq))]
	} else {
		buf = s.Seq
	}

	switch c {
	case 'v':
		if fs.Flag('#') {
			fmt.Fprintf(fs, "&%#v", *s)
			return
		}
		fallthrough
	case 's':
		if !fs.Flag('-') {
			fmt.Fprintf(fs, "%q ", s.ID)
		}
		for _, ql := range buf {
			fmt.Fprintf(fs, "%c", s.QFilter(s.Alpha, s.Threshold, ql))
		}
		if pOk && p < s.Len() {
			fmt.Fprint(fs, "...")
		}
	case 'a':
		s.formatDescLineTo(fs, '>')
		for i, ql := range buf {
			fmt.Fprintf(fs, "%c", s.QFilter(s.Alpha, s.Threshold, ql))
			if wOk && i < s.Len()-1 && i%w == w-1 {
				fmt.Fprintln(fs)
			}
		}
		if pOk && p < s.Len() {
			fmt.Fprint(fs, "...")
		}
	case 'q':
		s.formatDescLineTo(fs, '@')
		for _, ql := range buf {
			fmt.Fprintf(fs, "%c", s.QFilter(s.Alpha, s.Threshold, ql))
		}
		if pOk && p < s.Len() {
			fmt.Fprintln(fs, "...")
		} else {
			fmt.Fprintln(fs)
		}
		if fs.Flag('+') {
			s.formatDescLineTo(fs, '+')
		} else {
			fmt.Fprintln(fs, "+")
		}
		for _, ql := range buf {
			e := ql.Q.Encode(s.Encode)
			if e >= unicode.MaxASCII {
				e = unicode.MaxASCII - 1
			}
			fmt.Fprintf(fs, "%c", e)
		}
		if pOk && p < s.Len() {
			fmt.Fprint(fs, "...")
		}
	default:
		fmt.Fprintf(fs, "%%!%c(linear.QSeq=%.10s)", c, s)
	}
}

func (s *QSeq) formatDescLineTo(fs fmt.State, p rune) {
	fmt.Fprintf(fs, "%c%s", p, s.ID)
	if s.Desc != "" {
		fmt.Fprintf(fs, " %s", s.Desc)
	}
	fmt.Fprintln(fs)
}
