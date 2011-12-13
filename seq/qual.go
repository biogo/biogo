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
	"github.com/kortschak/BioGo/bio"
	"github.com/kortschak/BioGo/featgroup"
	"github.com/kortschak/BioGo/interval"
	"github.com/kortschak/BioGo/util"
	"math"
)

type Quality struct {
	ID       string
	Qual     []int8
	Offset   int
	Strand   int8
	Circular bool
}

func NewQuality(id string, seq []int8) *Quality {
	return &Quality{
		ID:       id,
		Qual:     seq,
		Offset:   0,
		Strand:   1,
		Circular: false,
	}
}

func (self *Quality) Len() int {
	return len(self.Qual)
}

func (self *Quality) Start() int {
	return self.Offset
}

func (self *Quality) End() int {
	return self.Offset + len(self.Qual)
}

func (self *Quality) Trunc(start, end int) (q *Quality, err error) {
	var ts []int8

	if start < self.Offset || end < self.Offset ||
		start > len(self.Qual)+self.Offset || end > len(self.Qual)+self.Offset {
		return nil, bio.NewError("Start or end position out of range.", 0, self)
	}

	if start <= end {
		ts = append([]int8{}, self.Qual[start-self.Offset:end-self.Offset]...)
	} else if self.Circular {
		ts = make([]int8, len(self.Qual)-start-self.Offset, len(self.Qual)+end-start)
		copy(ts, self.Qual[start-self.Offset:])
		ts = append(ts, self.Qual[:end-self.Offset]...)
	} else {
		return nil, bio.NewError("Start position greater than end position for non-circular molecule.", 0, self)
	}

	return &Quality{
		ID:       self.ID,
		Qual:     ts,
		Offset:   start,
		Strand:   self.Strand,
		Circular: false,
	}, nil
}

func (self *Quality) Reverse() *Quality {
	rs := make([]int8, len(self.Qual))

	for i, j := 0, len(self.Qual)-1; i < j; i, j = i+1, j-1 {
		rs[i], rs[j] = self.Qual[j], self.Qual[i]
	}

	return &Quality{
		ID:       self.ID,
		Qual:     rs,
		Offset:   0,
		Strand:   -self.Strand,
		Circular: self.Circular,
	}
}

func (self *Quality) Join(q *Quality, where int) (err error) {
	if self.Circular {
		return bio.NewError("Cannot join circular molecule.", 0, self)
	}
	switch where {
	case Prepend:
		ts := make([]int8, len(q.Qual), len(q.Qual)+len(self.Qual))
		copy(ts, q.Qual)
		self.Qual = append(ts, self.Qual...)
	case Append:
		self.Qual = append(self.Qual, q.Qual...)
	}

	return
}

func (self *Quality) Stitch(f *featgroup.FeatureGroup) (q *Quality, err error) {
	t := interval.NewTree()
	var i *interval.Interval

	for _, feature := range *f {
		if i, err = interval.New("", feature.Start, feature.End, 0, nil); err != nil {
			return nil, err
		} else {
			t.Insert(i)
		}
	}

	tq := []int8{}
	if span, err := interval.New("", self.Start(), self.End(), 0, nil); err != nil {
		panic("Seq.End() < Seq.Start()")
	} else {
		f, _ := t.Flatten(span, 0, 0)
		for _, seg := range f {
			tq = append(tq, self.Qual[util.Max(seg.Start()-self.Offset, 0):util.Min(seg.End()-self.Offset, len(self.Qual))]...)
		}
	}

	self.Qual = tq
	self.Offset = 0

	return self, nil
}
func SangerToSolexa(q int8) int8 {
	return int8(10 * math.Log10(math.Pow(10, float64(q)/10)-1))
}

func SolexaToSanger(q int8) int8 {
	return int8(10 * math.Log10(math.Pow(10, float64(q)/10)+1))
}

// Return the quality as a Sanger quality string
func (self *Quality) String() string {
	qs := make([]byte, 0, len(self.Qual))
	for _, q := range self.Qual {
		if q <= 93 {
			q += 33
		}
		qs = append(qs, byte(q))
	}
	return string(qs)
}
