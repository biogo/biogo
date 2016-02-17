// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linear

import (
	"github.com/biogo/biogo/alphabet"
	"github.com/biogo/biogo/feat"
	"github.com/biogo/biogo/seq"

	"fmt"
	"unicode"
)

// A Seq is a basic linear sequence.
type Seq struct {
	seq.Annotation
	Seq alphabet.Letters
}

// Interface guarantees
var (
	_ feat.Feature = (*Seq)(nil)
	_ seq.Sequence = (*Seq)(nil)
)

// NewSeq creates a new Seq with the given id, letter sequence and alphabet.
func NewSeq(id string, b []alphabet.Letter, alpha alphabet.Alphabet) *Seq {
	return &Seq{
		Annotation: seq.Annotation{
			ID:     id,
			Alpha:  alpha,
			Strand: seq.Plus,
		},
		Seq: append(alphabet.Letters(nil), b...),
	}
}

// Append append QLetters to the sequence, ignoring Q component.
func (s *Seq) AppendQLetters(a ...alphabet.QLetter) error {
	l := s.Len()
	s.Seq = append(s.Seq, make([]alphabet.Letter, len(a))...)[:l]
	for _, v := range a {
		s.Seq = append(s.Seq, v.L)
	}
	return nil
}

// Append appends Letters to the sequence.
func (s *Seq) AppendLetters(a ...alphabet.Letter) error {
	s.Seq = append(s.Seq, a...)
	return nil
}

// Slice returns the sequence data as a alphabet.Slice.
func (s *Seq) Slice() alphabet.Slice { return s.Seq }

// SetSlice sets the sequence data represented by the sequence. SetSlice will panic if sl
// is not a alphabet.Letters.
func (s *Seq) SetSlice(sl alphabet.Slice) { s.Seq = sl.(alphabet.Letters) }

// At returns the letter at position pos.
func (s *Seq) At(i int) alphabet.QLetter {
	return alphabet.QLetter{
		L: s.Seq[i-s.Offset],
		Q: seq.DefaultQphred,
	}
}

// Set sets the letter at position pos to l.
func (s *Seq) Set(i int, l alphabet.QLetter) error {
	s.Seq[i-s.Offset] = l.L
	return nil
}

// Len returns the length of the sequence.
func (s *Seq) Len() int { return len(s.Seq) }

// Start returns the start position of the sequence in coordinates relative to the sequence
// location.
func (s *Seq) Start() int { return s.Offset }

// End returns the end position of the sequence in coordinates relative to the sequence
// location.
func (s *Seq) End() int { return s.Offset + s.Len() }

// Validate validates the letters of the sequence according to the sequence alphabet.
func (s *Seq) Validate() (bool, int) { return s.Alpha.AllValid(s.Seq) }

// Clone returns a copy of the sequence.
func (s *Seq) Clone() seq.Sequence {
	c := *s
	c.Seq = append([]alphabet.Letter(nil), s.Seq...)
	return &c
}

// New returns an empty *Seq sequence with the same alphabet.
func (s *Seq) New() seq.Sequence {
	return &Seq{Annotation: seq.Annotation{Alpha: s.Alpha}}
}

// RevComp reverse complements the sequence. RevComp will panic if the alphabet used by
// the receiver is not a Complementor.
func (s *Seq) RevComp() {
	l, comp := s.Seq, s.Alphabet().(alphabet.Complementor).ComplementTable()
	i, j := 0, len(l)-1
	for ; i < j; i, j = i+1, j-1 {
		l[i], l[j] = comp[l[j]], comp[l[i]]
	}
	if i == j {
		l[i] = comp[l[i]]
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

// String returns a string representation of the sequence data only.
func (s *Seq) String() string { return alphabet.Letters(s.Seq).String() }

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
	var (
		w, wOk = fs.Width()
		p, pOk = fs.Precision()
		buf    []alphabet.Letter
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
		for _, l := range buf {
			fmt.Fprintf(fs, "%c", l)
		}
		if pOk && p < s.Len() {
			fmt.Fprint(fs, "...")
		}
	case 'a':
		s.formatDescLineTo(fs, '>')
		for i, l := range buf {
			fmt.Fprintf(fs, "%c", l)
			if wOk && i < s.Len()-1 && i%w == w-1 {
				fmt.Fprintln(fs)
			}
		}
		if pOk && p < s.Len() {
			fmt.Fprint(fs, "...")
		}
	case 'q':
		s.formatDescLineTo(fs, '@')
		for _, l := range buf {
			fmt.Fprintf(fs, "%c", l)
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
		e := seq.DefaultQphred.Encode(seq.DefaultEncoding)
		if e >= unicode.MaxASCII {
			e = unicode.MaxASCII - 1
		}
		for range buf {
			fmt.Fprintf(fs, "%c", e)
		}
		if pOk && p < s.Len() {
			fmt.Fprint(fs, "...")
		}
	default:
		fmt.Fprintf(fs, "%%!%c(linear.Seq=%.10s)", c, s)
	}
}

func (s *Seq) formatDescLineTo(fs fmt.State, p rune) {
	fmt.Fprintf(fs, "%c%s", p, s.ID)
	if s.Desc != "" {
		fmt.Fprintf(fs, " %s", s.Desc)
	}
	fmt.Fprintln(fs)
}
