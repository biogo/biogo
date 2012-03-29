// Package multi handles collections of sequences as alignments or sets.
package multi

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
	"github.com/kortschak/biogo/exp/seq/nucleic/packed"
	"github.com/kortschak/biogo/feat"
	"github.com/kortschak/biogo/interval"
	"github.com/kortschak/biogo/util"
	"reflect"
	"sync"
)

func init() {
	joinerRegistryLock = &sync.RWMutex{}
	joinerRegistry = make(map[reflect.Type]JoinFunc)
}

var (
	emptyString        = ""
	joinerRegistryLock *sync.RWMutex
	joinerRegistry     map[reflect.Type]JoinFunc
)

type Multi struct {
	ID         string
	Desc       string
	Loc        string
	S          []nucleic.Sequence
	Consensify nucleic.Consensifyer
	Stringify  seq.Stringify
	Meta       interface{} // No operation implicitly copies or changes the contents of Meta.
	alphabet   alphabet.Nucleic
	offset     int
	circular   bool
	encoding   alphabet.Encoding
}

// Create a new Multi sequence. Including Alignment or QAlignment types in n will result in undefined behaviour.
func NewMulti(id string, n []nucleic.Sequence, cons nucleic.Consensifyer) (m *Multi, err error) {
	var alpha alphabet.Nucleic
	for _, s := range n {
		if alpha != nil && s.Alphabet() != alpha {
			return nil, bio.NewError("Inconsistent alphabets", 0, n)
		} else if alpha == nil {
			alpha = s.Alphabet().(alphabet.Nucleic)
		}
	}
	m = &Multi{
		ID:         id,
		S:          n,
		alphabet:   alpha,
		Consensify: cons,
		Stringify: func(s seq.Polymer) string {
			t := s.(*Multi).Consensus(false)
			return t.String()
		},
	}

	return
}

// Interface guarantees:
var (
	_ seq.Polymer             = &Multi{}
	_ seq.Sequence            = &Multi{}
	_ seq.Scorer              = &Multi{}
	_ nucleic.Sequence        = &Multi{}
	_ nucleic.Quality         = &Multi{}
	_ nucleic.Getter          = &Multi{}
	_ nucleic.GetterAppender  = &Multi{}
	_ nucleic.Aligned         = &Multi{}
	_ nucleic.AlignedAppender = &Multi{}
)

// Required to satisfy nucleic Sequence interface.
func (self *Multi) Nucleic() {}

// Name returns a pointer to the ID string of the sequence.
func (self *Multi) Name() *string { return &self.ID }

// Description returns a pointer to the Desc string of the sequence.
func (self *Multi) Description() *string { return &self.Desc }

// Location returns a pointer to the Loc string of the sequence.
func (self *Multi) Location() *string { return &self.Loc }

// TODO
// func (self *Multi) Delete(i int) {}

func (self *Multi) Add(n ...nucleic.Sequence) (err error) {
	for _, s := range n {
		if s.Alphabet() != self.alphabet {
			return bio.NewError("Inconsistent alphabets", 0, self, s)
		}
	}
	self.S = append(self.S, n...)

	return
}

// Raw returns a pointer to the underlying []nucleic.Sequence slice.
func (self *Multi) Raw() interface{} { return &self.S }

// Append a to the ith sequence in the reciever.
func (self *Multi) Append(i int, a ...alphabet.QLetter) (err error) {
	return self.Get(i).(seq.Appender).Append(a...)
}

// Append each byte of each a to the appropriate sequence in the reciever.
func (self *Multi) AppendColumns(a ...[]alphabet.QLetter) (err error) {
	for i, s := range a {
		if len(s) != self.Count() {
			return bio.NewError(fmt.Sprintf("Column %d does not match Count(): %d != %d.", i, len(s), self.Count()), 0, a)
		}
	}
	for i, b := 0, make([]alphabet.QLetter, 0, len(a)); i < self.Count(); i, b = i+1, b[:0] {
		for _, s := range a {
			b = append(b, s[i])
		}
		self.Append(i, b...)
	}

	return
}

// Append each []byte in a to the appropriate sequence in the reciever.
func (self *Multi) AppendEach(a [][]alphabet.QLetter) (err error) {
	if len(a) != self.Count() {
		return bio.NewError(fmt.Sprintf("Number of sequences does not match Count(): %d != %d.", len(a), self.Count()), 0, a)
	}
	var i int
	for _, s := range self.S {
		if al, ok := s.(nucleic.AlignedAppender); ok {
			count := al.Count()
			if al.AppendEach(a[i:i+count]) != nil {
				panic("internal size mismatch")
			}
			i += count
		} else {
			s.(seq.Appender).Append(a[i]...)
			i++
		}
	}

	return
}

func (self *Multi) Get(i int) nucleic.Sequence {
	var count int
	for _, s := range self.S {
		if m, ok := s.(nucleic.Getter); ok {
			count = m.Count()
			if i < count {
				return m.Get(i)
			}
		} else {
			count = 1
			if i == 0 {
				return s
			}
		}
		i -= count
	}

	panic("index out of range")
}

func (self *Multi) Alphabet() alphabet.Alphabet { return self.alphabet }

func (self *Multi) At(pos seq.Position) alphabet.QLetter {
	var count int
	for _, s := range self.S {
		count = s.Count()
		if pos.Ind < count {
			return s.At(pos)
		}
		pos.Ind -= count
	}

	panic("index out of range")
}

// Encode the quality at position pos to a letter based on the sequence encoding setting.
func (self *Multi) QEncode(pos seq.Position) byte {
	return self.At(pos).Q.Encode(self.encoding)
}

// Decode a quality letter to a phred score based on the sequence encoding setting.
func (self *Multi) QDecode(l byte) alphabet.Qphred { return alphabet.DecodeToQphred(l, self.encoding) }

// Return the quality encoding type.
func (self *Multi) Encoding() alphabet.Encoding { return self.encoding }

// Set the quality encoding type to e.
func (self *Multi) SetEncoding(e alphabet.Encoding) {
	for _, s := range self.S {
		if enc, ok := s.(seq.Encoder); ok {
			enc.SetEncoding(e)
		}
	}
	self.encoding = e
}

func (self *Multi) EAt(pos seq.Position) float64 {
	var count int
	for _, s := range self.S {
		if a, ok := s.(seq.Counter); ok {
			count = a.Count()
		} else {
			count = 1
		}
		if pos.Ind < count {
			if qs, ok := s.(seq.Quality); ok {
				return qs.EAt(pos)
			} else {
				return nucleic.DefaultQphred.ProbE()
			}
		}
		pos.Ind -= count
	}

	panic("index out of range")
}

func (self *Multi) SetE(pos seq.Position, q float64) {
	var count int
	for _, s := range self.S {
		if a, ok := s.(seq.Counter); ok {
			count = a.Count()
		} else {
			count = 1
		}
		if pos.Ind < count {
			if qs, ok := s.(seq.Quality); ok {
				qs.SetE(pos, q)
				return
			}
		}
		pos.Ind -= count
	}

	panic("index out of range")
}

func (self *Multi) Set(pos seq.Position, l alphabet.QLetter) {
	var count int
	for _, s := range self.S {
		count = s.Count()
		if pos.Ind < count {
			s.Set(pos, l)
			return
		}
		pos.Ind -= count
	}

	panic("index out of range")
}

func (self *Multi) Column(pos int, fill bool) []alphabet.Letter {
	if pos < self.Start() || pos >= self.End() {
		panic("index out of range")
	}

	var c []alphabet.Letter
	if fill {
		c = make([]alphabet.Letter, 0, self.Count())
	} else {
		c = []alphabet.Letter{}
	}

	for _, s := range self.S {
		if a, ok := s.(nucleic.Aligned); ok {
			if a.Start() <= pos && pos < a.End() {
				c = append(c, a.Column(pos, fill)...)
			} else if fill {
				c = append(c, self.alphabet.Gap().Repeat(a.Count())...)
			}
		} else {
			if s.Start() <= pos && pos < s.End() {
				c = append(c, s.At(seq.Position{Pos: pos}).L)
			} else if fill {
				c = append(c, self.alphabet.Gap())
			}
		}
	}

	return c
}

func (self *Multi) ColumnQL(pos int, fill bool) []alphabet.QLetter {
	if pos < self.Start() || pos >= self.End() {
		panic("index out of range")
	}

	var c []alphabet.QLetter
	if fill {
		c = make([]alphabet.QLetter, 0, self.Count())
	} else {
		c = []alphabet.QLetter{}
	}

	for _, s := range self.S {
		if a, ok := s.(nucleic.Aligned); ok {
			if a.Start() <= pos && pos < a.End() {
				c = append(c, a.ColumnQL(pos, fill)...)
			} else if fill {
				c = append(c, alphabet.QLetter{L: self.alphabet.Gap()}.Repeat(a.Count())...)
			}
		} else {
			if s.Start() <= pos && pos < s.End() {
				c = append(c, s.At(seq.Position{Pos: pos}))
			} else if fill {
				c = append(c, alphabet.QLetter{L: self.alphabet.Gap()})
			}
		}
	}

	return c
}

func (self *Multi) Len() int {
	min := util.MaxInt
	max := util.MinInt

	for _, s := range self.S {
		if start := s.Start(); start < min {
			min = start
		}
		if end := s.End(); end > max {
			max = end
		}
	}

	return max - min
}

func (self *Multi) Count() (c int) {
	for _, s := range self.S {
		c += s.Count()
	}

	return
}

func (self *Multi) Offset(o int) {
	for _, s := range self.S {
		s.Offset(s.Start() - self.offset + o)
	}
	self.offset = o
}

func (self *Multi) Start() (start int) {
	start = util.MaxInt

	for _, s := range self.S {
		if l := s.Start(); l < start {
			start = l
		}
	}

	return
}

func (self *Multi) End() (end int) {
	end = util.MinInt

	for _, s := range self.S {
		if r := s.End(); r > end {
			end = r
		}
	}

	return
}

func (self *Multi) Copy() seq.Sequence {
	c := &Multi{}
	*c = *self
	c.Meta = nil
	c.S = make([]nucleic.Sequence, len(self.S))
	for i, s := range self.S {
		c.S[i] = s.Copy().(nucleic.Sequence)
	}

	return c
}

func (self *Multi) RevComp() {
	end := self.End()
	for _, s := range self.S {
		s.RevComp()
		s.Offset(end - s.End())
	}

	return
}

func (self *Multi) Reverse() {
	end := self.End()
	for _, s := range self.S {
		s.Reverse()
		s.Offset(end - s.End())
	}
}

func (self *Multi) Circular(c bool) { self.circular = c }

func (self *Multi) IsCircular() bool { return self.circular }

func (self *Multi) IsFlush(where int) bool {
	if self.Count() == 0 {
		return true
	}
	var start, end int
	for i, s := range self.S {
		if l, r := s.Start(), s.End(); i > 0 &&
			((l != start && where&seq.Start != 0) ||
				(r != end && where&seq.End != 0)) {
			return false
		} else if i == 0 {
			start, end = l, r
		}
	}
	return true
}

func (self *Multi) Flush(where int, fill alphabet.Letter) {
	if self.IsFlush(where) {
		return
	}

	if where&seq.Start != 0 {
		start := self.Start()
		for _, s := range self.S {
			if s.Start()-start < 1 {
				continue
			}
			if m, ok := s.(*Multi); ok {
				m.Flush(where, fill)
				continue
			}
			S := s.Raw()
			switch S.(type) {
			case *[]alphabet.Letter:
				uS := S.(*[]alphabet.Letter)
				*uS = append(fill.Repeat(s.Start()-start), *uS...)
			case *[]alphabet.QLetter:
				uS := S.(*[]alphabet.QLetter)
				*uS = append(alphabet.QLetter{L: fill}.Repeat(s.Start()-start), *uS...)
			case packed.Packing, *[]alphabet.QPack:
				panic("not implemented") // and never will be
				// packed.Seq cannot hold letters beyond the 4 letters in the
				// alphabet so it cannot have gaps.
				// Perhaps a bitmap of valid bases may be considered though
				// I can't see any particularly strong argument for this.
				// packed.QSeq could have gaps, by assigning 0 Qphred, but this
				// opens up possibility for abuse unless a valid bitmap is also
				// inlcuded for this type.
			}
			s.Offset(start)
		}
	}
	if where&seq.End != 0 {
		end := self.End()
		for i := 0; i < self.Count(); i++ {
			s := self.Get(i)
			if end-s.End() < 1 {
				continue
			}
			s.(seq.Appender).Append(alphabet.QLetter{L: fill}.Repeat(end - s.End())...)
		}
	}
}

func (self *Multi) Subseq(start, end int) (sub seq.Sequence, err error) {
	var (
		ns []nucleic.Sequence
		s  *Multi
	)

	for _, s := range self.S {
		rs, err := s.Subseq(start, end)
		if err != nil {
			return nil, err
		}
		ns = append(ns, rs.(nucleic.Sequence))
	}

	s = &Multi{}
	*s = *self
	s.S = ns

	return s, nil
}

func (self *Multi) Truncate(start, end int) (err error) {
	for _, s := range self.S {
		if err = s.Truncate(start, end); err != nil {
			return err
		}
	}

	return
}

func (self *Multi) Join(a *Multi, where int) (err error) {
	if self.Count() != a.Count() {
		return bio.NewError("Multis do not hold the same number of sequences", 0, []*Multi{self, a})
	}

	switch where {
	case seq.Start:
		if !a.IsFlush(seq.End) {
			a.Flush(seq.End, self.alphabet.Gap())
		}
		if !self.IsFlush(seq.Start) {
			self.Flush(seq.Start, self.alphabet.Gap())
		}
	case seq.End:
		if !a.IsFlush(seq.Start) {
			a.Flush(seq.Start, self.alphabet.Gap())
		}
		if !self.IsFlush(seq.End) {
			self.Flush(seq.End, self.alphabet.Gap())
		}
	}

	for i := 0; i < self.Count(); i++ {
		s := self.Get(i)
		as := a.Get(i)
		if err = joinOne(s, as, where); err != nil {
			return
		}
	}

	return
}

func joinOne(s, as nucleic.Sequence, where int) (err error) {
	switch s.(type) {
	case *nucleic.Seq:
		if t, ok := as.(*nucleic.Seq); !ok {
			err = joinFailure(s, t)
		} else {
			err = s.(*nucleic.Seq).Join(t, where)
		}
	case *nucleic.QSeq:
		if t, ok := as.(*nucleic.QSeq); !ok {
			err = joinFailure(s, t)
		} else {
			err = s.(*nucleic.QSeq).Join(t, where)
		}
	case *packed.Seq:
		if t, ok := as.(*packed.Seq); !ok {
			err = joinFailure(s, t)
		} else {
			err = s.(*packed.Seq).Join(t, where)
		}
	case *packed.QSeq:
		if t, ok := as.(*packed.QSeq); !ok {
			err = joinFailure(s, t)
		} else {
			err = s.(*packed.QSeq).Join(t, where)
		}
	case *Multi:
		if t, ok := as.(*Multi); !ok {
			err = joinFailure(s, t)
		} else {
			err = s.(*Multi).Join(t, where)
		}
	default:
		joinerRegistryLock.RLock()
		if joinerFunc, ok := joinerRegistry[reflect.TypeOf(s)]; ok {
			err = joinerFunc(s, as, where)
		} else {
			err = bio.NewError(fmt.Sprintf("Sequence type %T not handled.", s), 0, s)
		}
		joinerRegistryLock.RUnlock()
	}

	return
}

func joinFailure(s, as nucleic.Sequence) error {
	return bio.NewError(fmt.Sprintf("Sequence type mismatch: %T != %T.", s, as), 0, []nucleic.Sequence{s, as})
}

type JoinFunc func(a, b nucleic.Sequence, where int) (err error)

func RegisterJoiner(p seq.Polymer, f JoinFunc) {
	joinerRegistryLock.Lock()
	joinerRegistry[reflect.TypeOf(p)] = f
	joinerRegistryLock.Unlock()
}

func (self *Multi) Stitch(f feat.FeatureSet) (err error) {
	tr := interval.NewTree()
	var i *interval.Interval

	for _, feature := range f {
		if i, err = interval.New(emptyString, feature.Start, feature.End, 0, nil); err != nil {
			return
		} else {
			tr.Insert(i)
		}
	}

	span, err := interval.New(emptyString, self.Start(), self.End(), 0, nil)
	if err != nil {
		panic("Sequence.End() < Sequence.Start()")
	}
	fs, _ := tr.Flatten(span, 0, 0)

	ff := feat.FeatureSet{}
	for _, seg := range fs {
		ff = append(ff, &feat.Feature{
			Start: util.Max(seg.Start(), self.Start()),
			End:   util.Min(seg.End(), self.End()),
		})
	}

	return self.Compose(ff)
}

func (self *Multi) Compose(f feat.FeatureSet) (err error) {
	self.Flush(seq.Start|seq.End, self.alphabet.Gap())

	for _, s := range self.S {
		if err = s.Compose(f); err != nil {
			return err
		}
	}

	return
}

func (self *Multi) String() string { return self.Stringify(self) }

func (self *Multi) Consensus(includeMissing bool) (c *nucleic.QSeq) {
	cs := make([]alphabet.QLetter, 0, self.Len())
	for i := self.Start(); i < self.End(); i++ {
		cs = append(cs, self.Consensify(self, i, includeMissing))
	}

	c = nucleic.NewQSeq("Consensus:"+self.ID, cs, self.alphabet, self.encoding)
	c.Offset(self.offset)

	return
}
