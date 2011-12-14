package seq
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
	"github.com/kortschak/BioGo/feat"
	"github.com/kortschak/BioGo/interval"
	"github.com/kortschak/BioGo/util"
	"strings"
)

const (
	Left = 1 << iota
	Right
)

type Alignment []*Seq

func (self Alignment) Add(s *Seq) (a Alignment) {
	self = append(self, s)
	return self
}

func (self Alignment) Len() int {
	var (
		min  int = util.MaxInt
		max  int = util.MinInt
		span int
	)

	for _, s := range self {
		if s.Offset < min {
			min = s.Offset
		}
		if span = s.Offset + s.Len(); span > max {
			max = span
		}
	}

	return max - min
}

func (self Alignment) Start() (start int) {
	start = util.MaxInt

	for _, s := range self {
		if s.Offset < start {
			start = s.Offset
		}
	}

	return
}

func (self Alignment) End() (end int) {
	end = util.MinInt
	var right int

	for _, s := range self {
		if right = s.Offset + s.Len(); right > end {
			end = right
		}
	}

	return
}

func (self Alignment) IsFlush(where int) bool {
	for i, s, e := 0, self[0].Offset, self[0].Len(); i < len(self); i++ {
		if (self[i].Offset != s && (where&Left) != 0) || (self[i].Len() != e && (where&Right) != 0) {
			return false
		}
	}

	return true
}

func (self Alignment) Flush(where int, fill byte) Alignment {
	if where&Right != 0 {
		end := self.End()
		for _, s := range self {
			s.Seq = append(s.Seq, bytes.Repeat([]byte{fill}, end-(s.Offset+s.Len()))...)
		}
	}
	if where&Left != 0 {
		start := self.Start()
		for _, s := range self {
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

func (self Alignment) Trunc(start, end int) (a Alignment, err error) {
	var t *Seq
	a = Alignment{}
	for _, s := range self {
		if t, err = s.Trunc(start, end); err != nil {
			return nil, err
		}
		a.Add(t)
	}
	return
}

func (self Alignment) RevComp() (a Alignment, err error) {
	var t *Seq
	a = Alignment{}
	for _, s := range self {
		if t, err = s.RevComp(); err != nil {
			return nil, err
		}
		a.Add(t)
	}
	return
}

func (self Alignment) Join(a Alignment, fill byte, where int) (b Alignment, err error) {
	if len(self) != len(a) {
		return nil, bio.NewError("Alignments do not hold the same number of sequences", 0, []Alignment{self, a})
	}

	switch where {
	case Prepend:
		if !a.IsFlush(Right) {
			a.Flush(Right, fill)
		}
		if !self.IsFlush(Left) {
			self.Flush(Left, fill)
		}
		for i, s := range a {
			s.Seq = append(s.Seq, self[i].Seq...)
		}
		self = a
	case Append:
		if !a.IsFlush(Left) {
			a.Flush(Left, fill)
		}
		if !self.IsFlush(Right) {
			self.Flush(Right, fill)
		}
		for i, s := range self {
			s.Seq = append(s.Seq, a[i].Seq...)
		}
	}

	return self, nil
}

func (self Alignment) Stitch(f feat.FeatureSet) (a Alignment, err error) {
	for _, s := range self {
		if !s.Inplace && s.Quality != nil && s.Quality.Inplace {
			return nil, bio.NewError("Inplace operation on Quality with non-Inplace operation on parent Seq.", 0, s)
		}
	}

	t := interval.NewTree()
	var i *interval.Interval

	for _, feature := range f {
		if i, err = interval.New("", feature.Start, feature.End, 0, nil); err != nil {
			return nil, err
		} else {
			t.Insert(i)
		}
	}

	start := self.Start()
	a = make(Alignment, len(self))
	span, err := interval.New("", start, self.End(), 0, nil)
	if err != nil {
		panic("Seq.End() < Seq.Start()")
	}
	fs, _ := t.Flatten(span, 0, 0)

	var offset int
	for i, s := range self {
		if s.Inplace {
			s.Seq = s.stitch(fs)
			if s.Offset -= fs[0].Start(); offset < 0 {
				s.Offset = 0
			}
			s.Circular = false
			if s.Quality != nil {
				var q *Quality
				if s.Quality.Inplace {
					q = s.Quality
				} else {
					q = &Quality{ID: s.Quality.ID}
				}
				q.Qual = s.Quality.stitch(fs)
				if q.Offset = s.Quality.Offset - fs[0].Start(); q.Offset < 0 {
					q.Offset = 0
				}
				q.Circular = false
				s.Quality = q
			}
			a[i] = s
		} else {
			var q *Quality
			if s.Quality != nil {
				if offset = s.Quality.Offset - fs[0].Start(); offset < 0 {
					offset = 0
				}
				q = &Quality{
					ID:       s.Quality.ID,
					Qual:     s.Quality.stitch(fs),
					Offset:   offset,
					Circular: false,
				}
			}
			if offset = s.Offset - fs[0].Start(); offset < 0 {
				offset = 0
			}
			a[i] = &Seq{
				ID:       s.ID,
				Seq:      s.stitch(fs),
				Offset:   offset,
				Strand:   s.Strand,
				Circular: false,
				Moltype:  s.Moltype,
				Quality:  q,
			}
		}
	}

	return
}

type ConsFunc func(value []byte) byte

func (self Alignment) Consensus(f ConsFunc, fill byte) (c *Seq, err error) {
	start := self.Start()
	end := self.End()
	c = &Seq{Offset: start}
	stripe := make([]byte, len(self))
	for i := start; i < end; i++ {
		for j, s := range self {
			if i-s.Offset >= 0 || i-s.Offset < s.Offset+s.Len() {
				stripe[j] = s.Seq[i]
			} else {
				stripe[j] = fill
			}
		}
		c.Seq = append(c.Seq, f(stripe))
	}
	return
}

func (self Alignment) Column(pos int, fill byte) (c []byte, err error) {
	if pos < self.Start() || pos >= self.End() {
		return nil, bio.NewError("Column out of range", 0, nil)
	}
	c = make([]byte, len(self))
	for i, s := range self {
		if pos-s.Offset >= 0 || pos-s.Offset < s.Offset+s.Len() {
			c[i] = s.Seq[pos]
		} else {
			c[i] = fill
		}
	}

	return
}

var DefaultStringFunc = func(a Alignment) string {
	var b string
	start := a.Start()
	for _, s := range a {
		b += strings.Repeat(" ", s.Offset-start) + s.String() + "\n"
	}
	return b
}

var AlignmentStringFunc = DefaultStringFunc

func (self Alignment) String() string {
	return AlignmentStringFunc(self)
}
