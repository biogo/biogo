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

// Package packed provides support for manipulation of single nucleic acid
// sequences with and without quality data.
//
// Two basic nucleic acid sequence types are provided, Seq and QSeq.
package packed

import (
	"fmt"
	"github.com/kortschak/biogo/bio"
	"github.com/kortschak/biogo/exp/alphabet"
	"github.com/kortschak/biogo/exp/seq"
	"github.com/kortschak/biogo/exp/seq/nucleic"
	"github.com/kortschak/biogo/exp/seq/sequtils"
	"github.com/kortschak/biogo/feat"
	"github.com/kortschak/biogo/interval"
	"github.com/kortschak/biogo/util"
)

var emptyString = ""

// Packing is a type holding bit packed letters and padding offsets.
type Packing struct {
	Letters           []alphabet.Pack // Big-endian packing.
	LeftPad, RightPad int8
}

// Align the Packing to the specified end.
func (self *Packing) Align(where int) {
	switch where {
	case seq.Start:
		self.shiftLeft(self.Letters, self.LeftPad)
		if self.LeftPad+self.RightPad >= 4 {
			self.Letters = self.Letters[:len(self.Letters)-1]
		}
		self.RightPad += self.LeftPad
		self.RightPad &= 0x3
		self.LeftPad -= self.LeftPad
	case seq.End:
		self.shiftRight(self.Letters, self.RightPad)
		if self.LeftPad+self.RightPad >= 4 {
			self.Letters = self.Letters[1:]
		}
		self.LeftPad += self.RightPad
		self.LeftPad &= 0x3
		self.RightPad -= self.RightPad
	}
}

func (self *Packing) shiftLeft(s []alphabet.Pack, count int8) {
	if count > self.LeftPad || count < 0 {
		panic("packed: illegal shift")
	}
	if count == 0 {
		return
	}
	c := byte(count)
	for i := range s[:len(s)-1] {
		s[i] <<= c << 1
		s[i] |= s[i+1] >> (((4 - c) & 3) << 1)
	}
	s[len(s)-1] <<= c << 1
}

func (self *Packing) shiftRight(s []alphabet.Pack, count int8) {
	if count > self.RightPad || count < 0 {
		panic("packed: illegal shift")
	}
	if count == 0 {
		return
	}
	c := byte(count)
	for i := len(s) - 1; i > 0; i-- {
		s[i] >>= c << 1
		s[i] |= s[i-1] << (((4 - c) & 3) << 1)
	}
	s[0] >>= c << 1
}

// Pack bytes that conform to a into a slice of alphabet.Pack. Panics if a byte in s does not conform.
func PackLetters(a alphabet.Nucleic, s ...alphabet.Letter) (p *Packing) {
	p = &Packing{
		Letters:  make([]alphabet.Pack, (len(s)+3)/4),
		RightPad: int8(4-len(s)&3) & 3,
	}

	for i, c := range s {
		if !a.IsValid(c) {
			panic("packed: invalid letter")
		}
		p.Letters[i/4] <<= 2
		p.Letters[i/4] |= alphabet.Pack(a.IndexOf(c))
	}
	if sc := uint(len(s)) & 3; sc != 0 {
		p.Letters[len(p.Letters)-1] <<= (4 - sc) << 1
	}

	return
}

// Pack bytes that conform to a into a slice of alphabet.Pack. Panics if a byte in s does not conform.
func PackQLetters(a alphabet.Nucleic, s ...alphabet.QLetter) (p *Packing) {
	p = &Packing{
		Letters:  make([]alphabet.Pack, (len(s)+3)/4),
		RightPad: int8(4-len(s)&3) & 3,
	}

	for i, c := range s {
		if !a.IsValid(c.L) {
			panic("packed: invalid letter")
		}
		p.Letters[i/4] <<= 2
		p.Letters[i/4] |= alphabet.Pack(a.IndexOf(c.L))
	}
	if sc := uint(len(s)) & 3; sc != 0 {
		p.Letters[len(p.Letters)-1] <<= (4 - sc) << 1
	}

	return
}

// Seq is a nucleic sequence packed 4 bases per byte.
type Seq struct {
	ID        string
	Desc      string
	Loc       string
	S         *Packing
	Strand    nucleic.Strand
	Stringify seq.Stringify // Function allowing user specified string representation.
	Meta      interface{}   // No operation implicitly copies or changes the contents of Meta.
	alphabet  alphabet.Nucleic
	circular  bool
	offset    int
}

func checkPackedAlpha(alpha alphabet.Nucleic) error {
	if alpha.Len() != 4 {
		return bio.NewError("Cannot create packed sequence with alphabet length != 4", 0, alpha)
	}
	for _, v := range alphabet.BytesToLetters([]byte(alpha.String())) {
		if c, ok := alpha.Complement(v); ok && alpha.IndexOf(v) != alpha.IndexOf(c)^0x3 {
			// TODO: Resolution to the following problem:
			// Normal nucleotide alphabets (ACGT/ACGU) are safe with this in either case sensitive or
			// insensitive. Other alphabets may not be, in this case specify case sensitive.
			return bio.NewError("alphabet order not consistent with bit operations for packed.", 0, alpha)
		}
	}

	return nil
}

// Create a new Seq with the given id, letter sequence and alphabet.
func NewSeq(id string, b []alphabet.Letter, alpha alphabet.Nucleic) (p *Seq, err error) {
	defer func() {
		if r := recover(); r != nil {
			_, pos := alpha.AllValid(b)
			err = bio.NewError(fmt.Sprintf("Encoding error: %s %q at position %d.", r, b[pos], pos), 1, b)
		}
	}()

	err = checkPackedAlpha(alpha)
	if err != nil {
		return
	}

	p = &Seq{
		ID:        id,
		S:         PackLetters(alpha, b...),
		alphabet:  alpha,
		Strand:    1,
		Stringify: Stringify,
	}

	return
}

// Interface guarantees:
var (
	_ seq.Polymer      = &Seq{}
	_ seq.Sequence     = &Seq{}
	_ seq.Appender     = &Seq{}
	_ nucleic.Sequence = &Seq{}
)

// Required to satisfy nucleic.Sequence interface.
func (self *Seq) Nucleic() {}

// Name returns a pointer to the ID string of the sequence.
func (self *Seq) Name() *string { return &self.ID }

// Description returns a pointer to the Desc string of the sequence.
func (self *Seq) Description() *string { return &self.Desc }

// Location returns a pointer to the Loc string of the sequence.
func (self *Seq) Location() *string { return &self.Loc }

// Raw returns the underlying *Packing struct pointer.
func (self *Seq) Raw() interface{} { return self.S }

// Append Letters to the sequence.
func (self *Seq) AppendLetters(a ...alphabet.Letter) (err error) {
	defer func() {
		if r := recover(); r != nil {
			_, pos := self.alphabet.AllValid(a)
			err = bio.NewError(fmt.Sprintf("Encoding error: %s %q at position %d.", r, a[pos], pos), 1, a)
		}
	}()

	i := 0
	for ; self.S.RightPad > 0 && i < len(a); i, self.S.RightPad = i+1, self.S.RightPad-1 {
		if !self.alphabet.IsValid(a[i]) {
			return bio.NewError(fmt.Sprintf("Invalid letter %q at position %d.", a[i], i), 0, nil)
		}
		self.S.Letters[len(self.S.Letters)-1] |= alphabet.Pack(self.alphabet.IndexOf(a[i])) << (4 - byte(self.S.RightPad))
	}
	self.S.Letters = append(self.S.Letters, PackLetters(self.alphabet, a[i:]...).Letters...)

	return
}

// Append QLetters to the sequence.
func (self *Seq) AppendQLetters(a ...alphabet.QLetter) (err error) {
	defer func() {
		if r := recover(); r != nil {
			_, pos := self.alphabet.AllValidQLetter(a)
			err = bio.NewError(fmt.Sprintf("Encoding error: %s %q at position %d.", r, a[pos], pos), 1, a)
		}
	}()

	i := 0
	for ; self.S.RightPad > 0 && i < len(a); i, self.S.RightPad = i+1, self.S.RightPad-1 {
		if !self.alphabet.IsValid(a[i].L) {
			return bio.NewError(fmt.Sprintf("Invalid letter %q at position %d.", a[i], i), 0, nil)
		}
		self.S.Letters[len(self.S.Letters)-1] |= alphabet.Pack(self.alphabet.IndexOf(a[i].L)) << (4 - byte(self.S.RightPad))
	}
	self.S.Letters = append(self.S.Letters, PackQLetters(self.alphabet, a[i:]...).Letters...)

	return
}

// Return the Alphabet used by the sequence.
func (self *Seq) Alphabet() alphabet.Alphabet { return self.alphabet }

// Return the letter at position pos.
func (self *Seq) At(pos seq.Position) alphabet.QLetter {
	if pos.Ind != 0 || pos.Pos < self.offset || pos.Pos >= self.End() {
		panic("packed: index out of range")
	}
	p := pos.Pos + int(self.S.LeftPad) - self.offset
	pIndex, pOffset := p/4, uint(p%4)
	code := self.S.Letters[pIndex] >> (2 * (3 - pOffset)) & 0x3
	return alphabet.QLetter{
		L: self.alphabet.Letter(int(code)),
		Q: nucleic.DefaultQphred,
	}
}

// Set the letter at position pos to l.
func (self *Seq) Set(pos seq.Position, l alphabet.QLetter) {
	if pos.Ind != 0 || pos.Pos < self.offset || pos.Pos >= self.End() {
		panic("packed: index out of range")
	}
	p := pos.Pos + int(self.S.LeftPad) - self.offset
	pIndex, pOffset := p/4, uint(p%4)
	self.S.Letters[pIndex] &^= 0x3 << (2 * pOffset)
	self.S.Letters[pIndex] |= alphabet.Pack(self.alphabet.IndexOf(l.L)) << (2 * pOffset)
}

// Return the length of the sequence.
func (self *Seq) Len() int {
	return len(self.S.Letters)*4 - int(self.S.LeftPad) - int(self.S.RightPad)
}

// Satisfy Counter.
func (self *Seq) Count() int { return 1 }

// Set the global offset of the sequence to o.
func (self *Seq) Offset(o int) { self.offset = o }

// Return the start position of the sequence in global coordinates.
func (self *Seq) Start() int { return self.offset }

// Return the end position of the sequence in global coordinates.
func (self *Seq) End() int { return self.offset + self.Len() }

// Return the molecule type of the sequence.
func (self *Seq) Moltype() bio.Moltype { return self.alphabet.Moltype() }

// Validate the letters of the sequence according to the specified alphabet. This is always successful as encoding does not allow invalid letters.
func (self *Seq) Validate() (bool, int) { return true, -1 }

// Return a copy of the sequence.
func (self *Seq) Copy() seq.Sequence {
	c := &Seq{}
	*c = *self
	c.S = &Packing{}
	*c.S = *self.S
	c.S.Letters = append([]alphabet.Pack(nil), self.S.Letters...)

	return c
}

// Reverse complement the sequence.
func (self *Seq) RevComp() {
	// depends on complements being reversed bit order - ensure this at construction
	self.S.Letters = sequtils.Reverse(self.S.Letters).([]alphabet.Pack)
	for i := range self.S.Letters {
		v := &self.S.Letters[i]
		*v = ((*v >> 2) & 0x33) | ((*v & 0x33) << 2)
		*v = ((*v >> 4) & 0x0f) | ((*v & 0x0f) << 4)
		*v ^= 0xff
	}
	self.S.LeftPad, self.S.RightPad = self.S.RightPad, self.S.LeftPad
	self.Strand = -self.Strand
}

// Reverse the sequence.
func (self *Seq) Reverse() {
	self.S.Letters = sequtils.Reverse(self.S.Letters).([]alphabet.Pack)
	for i := range self.S.Letters {
		v := &self.S.Letters[i]
		*v = ((*v >> 2) & 0x33) | ((*v & 0x33) << 2)
		*v = ((*v >> 4) & 0x0f) | ((*v & 0x0f) << 4)
	}
	self.S.LeftPad, self.S.RightPad = self.S.RightPad, self.S.LeftPad
}

// Specify that the sequence is circular.
func (self *Seq) Circular(c bool) { self.circular = c }

// Return whether the sequence is circular.
func (self *Seq) IsCircular() bool { return self.circular }

// Return a subsequence from start to end, wrapping if the sequence is circular.
func (self *Seq) Subseq(start int, end int) (sub seq.Sequence, err error) {
	var (
		ps *Seq
		tt interface{}
	)

	sl, sr := (self.Len()-start-self.offset+int(self.S.RightPad)+3)/4, (end-self.offset+int(self.S.LeftPad)+3)/4

	if s, e := (start-self.offset+int(self.S.LeftPad))/4, (end-self.offset+int(self.S.LeftPad)+3)/4; s != 0 || e != len(self.S.Letters) {
		tt, err = sequtils.Truncate(self.S.Letters, s, e, self.circular)
		if err == nil {
			ps = &Seq{}
			*ps = *self
			ps.S = &Packing{}
			*ps.S = *self.S
			ps.S.Letters = tt.([]alphabet.Pack)
			ps.Meta = nil
		} else {
			return
		}
	}

	if ps.circular && start > end {
		if sl+sr != len(ps.S.Letters) {
			panic(fmt.Sprintf("internal inconsistency %d + %d != %d", (sl+3)/4, (sr+3)/4, len(ps.S.Letters)))
		}
		ps.S.shiftRight(ps.S.Letters[:sl], ps.S.RightPad)
		ps.S.shiftLeft(ps.S.Letters[len(ps.S.Letters)-sr:], ps.S.LeftPad)

		ps.S.LeftPad, ps.S.RightPad = int8(4-start-ps.offset+int(ps.S.RightPad))&3, int8(4-end-ps.offset+int(ps.S.LeftPad))&3
		ps.offset = 0
	} else {
		ps.S.LeftPad, ps.S.RightPad = int8(start-ps.offset+int(ps.S.LeftPad))&3, int8(4-end-ps.offset+int(ps.S.LeftPad))&3
		ps.offset = start
	}
	ps.circular = false

	return ps, err
}

// Truncate the sequence from start to end, wrapping if the sequence is circular.
func (self *Seq) Truncate(start int, end int) (err error) {
	var tt interface{}

	sl, sr := (self.Len()-start-self.offset+int(self.S.RightPad)+3)/4, (end-self.offset+int(self.S.LeftPad)+3)/4

	if s, e := (start-self.offset+int(self.S.LeftPad))/4, (end-self.offset+int(self.S.LeftPad)+3)/4; s != 0 || e != len(self.S.Letters) {
		tt, err = sequtils.Truncate(self.S.Letters, s, e, self.circular)
		if err == nil {
			self.S.Letters = tt.([]alphabet.Pack)
		} else {
			return
		}
	}

	if self.circular && start > end {
		if sl+sr != len(self.S.Letters) {
			panic(fmt.Sprintf("packed: internal inconsistency %d + %d != %d", (sl+3)/4, (sr+3)/4, len(self.S.Letters)))
		}
		self.S.shiftRight(self.S.Letters[:sl], self.S.RightPad)
		self.S.shiftLeft(self.S.Letters[len(self.S.Letters)-sr:], self.S.LeftPad)

		self.S.LeftPad, self.S.RightPad = int8(4-start-self.offset+int(self.S.RightPad))&3, int8(4-end-self.offset+int(self.S.LeftPad))&3
		self.offset = 0
	} else {
		self.S.LeftPad, self.S.RightPad = int8(start-self.offset+int(self.S.LeftPad))&3, int8(4-end-self.offset+int(self.S.LeftPad))&3
		self.offset = start
	}
	self.circular = false

	return
}

// Join p to the sequence at the end specified by where.
func (self *Seq) Join(p *Seq, where int) (err error) {
	if self.circular {
		return bio.NewError("Cannot join circular sequence: receiver.", 0, self)
	} else if p.circular {
		return bio.NewError("Cannot join circular sequence: parameter.", 0, p)
	}

	switch where {
	case seq.Start:
		p = p.Copy().(*Seq)
		p.S.Align(seq.End)
		self.S.Align(seq.Start)
		self.S.LeftPad = p.S.LeftPad
	case seq.End:
		p = p.Copy().(*Seq)
		p.S.Align(seq.Start)
		self.S.Align(seq.End)
		self.S.RightPad = p.S.RightPad
	default:
		return bio.NewError("Undefined location.", 0, where)
	}

	tt, offset := sequtils.Join(self.S.Letters, p.S.Letters, where)
	self.offset = offset
	self.S.Letters = tt.([]alphabet.Pack)

	return
}

// Join sequentially order disjunct segments of the sequence, returning any error.
func (self *Seq) Stitch(f feat.FeatureSet) (err error) {
	tr := interval.NewTree()
	var i *interval.Interval

	for _, feature := range f {
		i, err = interval.New(emptyString, feature.Start, feature.End, 0, nil)
		if err != nil {
			return
		} else {
			tr.Insert(i)
		}
	}

	span, err := interval.New(emptyString, self.offset, self.End(), 0, nil)
	if err != nil {
		panic("packed: Sequence.End() < Sequence.Start()")
	}
	fs, _ := tr.Flatten(span, 0, 0)
	l := 0

	for _, seg := range fs {
		l += util.Min(seg.End(), self.End()) - util.Max(seg.Start(), self.Start())
	}

	t := &Seq{}
	*t = *self
	t.S = &Packing{Letters: make([]alphabet.Pack, 0, (l+3)/4)}

	var tseg seq.Sequence
	for _, seg := range fs {
		tseg, err = self.Subseq(util.Max(seg.Start(), self.Start()), util.Min(seg.End(), self.End()))
		if err != nil {
			return
		}
		s := tseg.(*Seq).S
		s.Align(seq.Start)
		t.S.Align(seq.End)
		t.S.Letters = append(t.S.Letters, s.Letters...)
		t.S.RightPad = s.RightPad
	}

	*self = *t

	return
}

// Join segments of the sequence, returning any error.
func (self *Seq) Compose(f feat.FeatureSet) (err error) {
	l := 0
	for _, seg := range f {
		if seg.End < seg.Start {
			return bio.NewError("Feature end < start", 0, seg)
		}
		l += util.Min(seg.End, self.End()) - util.Max(seg.Start, self.Start())
	}

	t := &Seq{}
	*t = *self
	t.S = &Packing{Letters: make([]alphabet.Pack, 0, (l+3)/4)}

	var tseg seq.Sequence
	for _, seg := range f {
		tseg, err = self.Subseq(util.Max(seg.Start, self.Start()), util.Min(seg.End, self.End()))
		if err != nil {
			return
		}
		tseg := tseg.(*Seq)
		if seg.Strand == -1 {
			tseg.RevComp()
		}
		tseg.S.Align(seq.Start)
		t.S.Align(seq.End)
		t.S.Letters = append(t.S.Letters, tseg.S.Letters...)
		t.S.RightPad = tseg.S.RightPad
	}

	*self = *t

	return
}

// Return an unpacked sequence.
func (self *Seq) Unpack() (n *nucleic.Seq) {
	n = nucleic.NewSeq(self.ID, alphabet.BytesToLetters([]byte(self.String())), self.alphabet)
	n.Circular(self.circular)
	n.Offset(self.offset)

	return
}

// Return a string representation of the sequence. Representation is determined by the Stringify field.
func (self *Seq) String() string { return self.Stringify(self) }

// The default Stringify function for Seq.
var Stringify = func(s seq.Polymer) string {
	cs := make([]alphabet.Letter, 0, s.Len())
	switch s := s.(*Seq); len(s.S.Letters) {
	case 0:
		break
	case 1, 2:
		for i := s.Start(); i < s.End(); i++ {
			cs = append(cs, s.At(seq.Position{Pos: i}).L)
		}
	default:
		// first byte
		for p, i := s.S.Letters[0], 3-int(s.S.LeftPad); i >= 0; i-- {
			cs = append(cs, s.alphabet.Letter(int(p>>(uint(i)<<1)&0x3)))
		}
		// middle bytes
		for _, p := range s.S.Letters[1 : len(s.S.Letters)-1] {
			for i := 3; i >= 0; i-- {
				cs = append(cs, s.alphabet.Letter(int(p>>(uint(i)<<1)&0x3)))
			}
		}
		// last byte
		for p, i := s.S.Letters[len(s.S.Letters)-1], 3; i >= int(s.S.RightPad); i-- {
			cs = append(cs, s.alphabet.Letter(int(p>>(uint(i)<<1)&0x3)))
		}
	}

	return alphabet.Letters(cs).String()
}
