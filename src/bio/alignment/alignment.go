// Basic sequence alignment package
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
	"bytes"
	"github.com/kortschak/BioGo/bio"
	"github.com/kortschak/BioGo/bio/featgroup"
	"github.com/kortschak/BioGo/bio/interval"
	"github.com/kortschak/BioGo/bio/seq"
	"github.com/kortschak/BioGo/bio/util"
	"strings"
)

const (
	Left = 1 << iota
	Right
)

type Alignment []*seq.Seq

func (self *Alignment) Add(s *seq.Seq) (a *Alignment) {
	*self = append(*self, s)
	return self
}

func (self *Alignment) Len() int {
	var (
		min  int = util.MaxInt
		max  int = util.MinInt
		span int
	)

	for _, s := range *self {
		if s.Offset < min {
			min = s.Offset
		}
		if span = s.Offset + s.Len(); span > max {
			max = span
		}
	}

	return max - min
}

func (self *Alignment) Start() (start int) {
	start = util.MaxInt

	for _, s := range *self {
		if s.Offset < start {
			start = s.Offset
		}
	}

	return
}

func (self *Alignment) End() (end int) {
	end = util.MinInt
	var right int

	for _, s := range *self {
		if right = s.Offset + s.Len(); right > end {
			end = right
		}
	}

	return
}

func (self *Alignment) IsFlush(where int) bool {
	for i, s, e := 0, (*self)[0].Offset, (*self)[0].Len(); i < len(*self); i++ {
		if ((*self)[i].Offset != s && (where&Left) != 0) || ((*self)[i].Len() != e && (where&Right) != 0) {
			return false
		}
	}

	return true
}

func (self *Alignment) Flush(where int, fill byte) *Alignment {
	if where&Right != 0 {
		end := self.End()
		for _, s := range *self {
			s.Seq = append(s.Seq, bytes.Repeat([]byte{fill}, end-(s.Offset+s.Len()))...)
		}
	}
	if where&Left != 0 {
		start := self.Start()
		for _, s := range *self {
			if diff := s.Offset - start; diff > 0 {
				b := make([]byte, diff, diff+s.Len())
				copy(b, bytes.Repeat([]byte{fill}, diff))
				s.Seq = append(b, s.Seq...)
				s.Offset = start
			}
		}
	}

	return self
}

func (self *Alignment) Trunc(start, end int) (a *Alignment, err error) {
	var t *seq.Seq
	a = &Alignment{}
	for _, s := range *self {
		if t, err = s.Trunc(start, end); err != nil {
			return nil, err
		}
		a.Add(t)
	}
	return
}

func (self *Alignment) RevComp() (a *Alignment, err error) {
	var t *seq.Seq
	a = &Alignment{}
	for _, s := range *self {
		if t, err = s.RevComp(); err != nil {
			return nil, err
		}
		a.Add(t)
	}
	return
}

func (self *Alignment) Join(a *Alignment, fill byte, where int) (b *Alignment, err error) {
	if len(*self) != len(*a) {
		return nil, bio.NewError("Alignments do not hold the same number of sequences", 0, []*Alignment{self, a})
	}
	switch {
	case fill == seq.Prepend:

		if !a.IsFlush(Right) {
			a.Flush(Right, fill)
		}
		if !self.IsFlush(Left) {
			self.Flush(Left, fill)
		}
		for i, s := range *a {
			s.Seq = append(s.Seq, (*self)[i].Seq...)
		}
		self = a
	case fill == seq.Append:
		if !a.IsFlush(Left) {
			a.Flush(Left, fill)
		}
		if !self.IsFlush(Right) {
			self.Flush(Right, fill)
		}
		for i, s := range *self {
			s.Seq = append(s.Seq, (*a)[i].Seq...)
		}
	}
	return self, nil
}

func (self *Alignment) Stitch(f *featgroup.FeatureGroup) (a *Alignment, err error) {
	t := interval.NewTree()
	var i *interval.Interval

	for _, feature := range *f {
		if i, err = interval.New("", feature.Start, feature.End, 0, nil); err != nil {
			return nil, err
		} else {
			t.Insert(i)
		}
	}
	// Mark the end of the sequence
	i, _ = interval.New("EOA", self.End()+1, self.End()+1, 0, nil)
	t.Insert(i)

	a = &Alignment{}
	for i := 0; i < len(*self); i++ {
		*a = append(*a, &seq.Seq{})
	}

	var (
		start int
		last  *interval.Interval
	)

	left := self.Start()
	for chromosome := range t {
		for segment := range t.Traverse(chromosome) {
			if segment.End() < left {
				continue
			}
			if last == nil { // start of the features
				start = segment.Start()
			}
			if last.End() < segment.Start()-1 { // at least one position gap between this feature and the last
				for i, sequence := range *self {
					(*a)[i].Seq = append((*a)[i].Seq, sequence.Seq[util.Max(0, start-sequence.Offset):last.End()-sequence.Offset]...)
				}
				start = segment.Start()
			}
			if segment.End() > self.End() { // this is the last useful segment
				for i, sequence := range *self {
					(*a)[i].Seq = append((*a)[i].Seq, sequence.Seq[util.Max(0, start-sequence.Offset):util.Min(len(sequence.Seq), last.End()-sequence.Offset)]...)
				}
				break
			}
			last = segment
		}
	}

	for i := 0; i < len(*self); i++ {
		(*self)[i].Seq = (*a)[i].Seq
		(*self)[i].Offset = 0
	}

	return self, nil
}

type ConsFunc func(value []byte) byte

func (self *Alignment) Consensus(f ConsFunc, fill byte) (c *seq.Seq, err error) {
	start := self.Start()
	end := self.End()
	c = &seq.Seq{Offset: start}
	stripe := make([]byte, len(*self))
	for i := start; i < end; i++ {
		for j, s := range *self {
			if i >= s.Offset && i < s.Offset+s.Len() {
				stripe[j] = s.Seq[i-start]
			} else {
				stripe[j] = fill
			}
		}
		c.Seq = append(c.Seq, f(stripe))
	}
	return
}

func (self *Alignment) Column(pos int) (c []byte, err error) {
	c = make([]byte, 0, len(*self))
	for _, s := range *self {
		if pos < 0 || pos >= s.Len() {
			return nil, bio.NewError("Column out of range", 0, nil)
		}
		c = append(c, s.Seq[pos])
	}
	return
}

var DefaultStringFunc = func(a *Alignment) string {
	var b string
	start := a.Start()
	for _, s := range *a {
		b += strings.Repeat(" ", s.Offset-start) + s.String() + "\n"
	}
	return b
}

var StringFunc = DefaultStringFunc

func (self *Alignment) String() string {
	return StringFunc(self)
}
