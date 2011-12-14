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
	"github.com/kortschak/BioGo/feat"
	"github.com/kortschak/BioGo/interval"
	"github.com/kortschak/BioGo/util"
)

type Quality struct {
	ID       string
	Qual     []Qsanger
	Offset   int
	Strand   int8
	Circular bool
	Inplace  bool
}

func NewQuality(id string, q []Qsanger) *Quality {
	return &Quality{
		ID:       id,
		Qual:     q,
		Offset:   0,
		Strand:   1,
		Circular: false,
		Inplace:  false,
	}
}

// Set Inplace and return self for chaining.
func (self *Quality) WorkInplace(b bool) *Quality {
	self.Inplace = b
	return self
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
	var tq []Qsanger

	if start < self.Offset || end < self.Offset ||
		start > len(self.Qual)+self.Offset || end > len(self.Qual)+self.Offset {
		return nil, bio.NewError("Start or end position out of range.", 0, self)
	}

	if start <= end {
		if self.Inplace {
			tq = self.Qual[start-self.Offset : end-self.Offset]
		} else {
			tq = append([]Qsanger{}, self.Qual[start-self.Offset:end-self.Offset]...)
		}
	} else if self.Circular {
		if self.Inplace {
			tq = append(self.Qual[start-self.Offset:], self.Qual[:end-self.Offset]...) // not quite inplace for this op
		} else {
			tq = make([]Qsanger, len(self.Qual)-start-self.Offset, len(self.Qual)+end-start)
			copy(tq, self.Qual[start-self.Offset:])
			tq = append(tq, self.Qual[:end-self.Offset]...)
		}
	} else {
		return nil, bio.NewError("Start position greater than end position for non-circular molecule.", 0, self)
	}

	if self.Inplace {
		q = self
		q.Qual = tq
		q.Circular = false
	} else {
		q = &Quality{
			ID:       self.ID,
			Qual:     tq,
			Offset:   start,
			Strand:   self.Strand,
			Circular: false,
		}
	}

	return
}

func (self *Quality) Reverse() (q *Quality) {
	var rq []Qsanger
	if self.Inplace {
		rq = self.Qual
	} else {
		rq = make([]Qsanger, len(self.Qual))
	}

	i, j := 0, len(self.Qual)-1
	for ; i < j; i, j = i+1, j-1 {
		rq[i], rq[j] = self.Qual[j], self.Qual[i]
	}
	if i == j {
		rq[i] = self.Qual[i]
	}

	if self.Inplace {
		q = self
	} else {
		q = &Quality{
			ID:       self.ID,
			Qual:     rq,
			Offset:   self.Offset + len(self.Qual),
			Strand:   -self.Strand,
			Circular: self.Circular,
		}
	}

	return
}

func (self *Quality) Join(q *Quality, where int) (j *Quality, err error) {
	var (
		tq []Qsanger
		ID string
	)

	if self.Circular {
		return nil, bio.NewError("Cannot join circular molecule.", 0, self)
	}

	switch where {
	case Prepend:
		ID = q.ID + "+" + self.ID
		tq = make([]Qsanger, len(q.Qual), len(q.Qual)+len(self.Qual))
		copy(tq, q.Qual)
		tq = append(tq, self.Qual...)
	case Append:
		ID = self.ID + "+" + q.ID
		if self.Inplace {
			tq = append(self.Qual, q.Qual...)
		} else {
			tq = make([]Qsanger, len(self.Qual), len(q.Qual)+len(self.Qual))
			copy(tq, self.Qual)
			tq = append(tq, q.Qual...)
		}
	}

	if self.Inplace {
		j = self
		j.Qual = tq
	} else {
		j = &Quality{
			ID:   ID,
			Qual: tq,
		}
	}
	if where == Prepend {
		j.Offset -= j.Len()
	}

	return
}

func (self *Quality) Stitch(f feat.FeatureSet) (q *Quality, err error) {
	t := interval.NewTree()
	var i *interval.Interval

	for _, feature := range f {
		if i, err = interval.New("", feature.Start, feature.End, 0, nil); err != nil {
			return nil, err
		} else {
			t.Insert(i)
		}
	}

	tq := []Qsanger{}
	if span, err := interval.New("", self.Start(), self.End(), 0, nil); err != nil {
		panic("Seq.End() < Seq.Start()")
	} else {
		f, _ := t.Flatten(span, 0, 0)
		for _, seg := range f {
			tq = append(tq, self.Qual[util.Max(seg.Start()-self.Offset, 0):util.Min(seg.End()-self.Offset, len(self.Qual))]...)
		}
	}

	if self.Inplace {
		q = self
		q.Qual = tq
		q.Offset = 0
		q.Circular = false
	} else {
		q = &Quality{
			ID:       self.ID,
			Qual:     tq,
			Offset:   0,
			Strand:   self.Strand,
			Circular: false,
		}
	}

	return
}

// Return the quality as a Sanger quality string
func (self *Quality) String() string {
	qs := make([]byte, 0, len(self.Qual))
	for _, q := range self.Qual {
		qs = append(qs, q.Encode(Sanger))
	}
	return string(qs)
}
