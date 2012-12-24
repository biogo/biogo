// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package seq

import (
	"code.google.com/p/biogo/bio"
	"code.google.com/p/biogo/feat"
	"code.google.com/p/biogo/interval"
	"code.google.com/p/biogo/util"
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
			tq = append([]Qsanger(nil), self.Qual[start-self.Offset:end-self.Offset]...)
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
		i, err = interval.New(emptyString, feature.Start, feature.End, 0, nil)
		if err != nil {
			return
		} else {
			t.Insert(i)
		}
	}

	span, err := interval.New(emptyString, self.Start(), self.End(), 0, nil)
	if err != nil {
		panic("Seq.End() < Seq.Start()")
	}

	fs, _ := t.Flatten(span, 0, 0)
	if self.Inplace {
		q = self
		q.Qual = self.stitch(fs)
		q.Offset = 0
		q.Circular = false
	} else {
		q = &Quality{
			ID:       self.ID,
			Qual:     self.stitch(fs),
			Offset:   0,
			Strand:   self.Strand,
			Circular: false,
		}
	}

	return
}

func (self *Quality) stitch(fs []*interval.Interval) (tq []Qsanger) {
	for _, seg := range fs {
		tq = append(tq, self.Qual[util.Max(seg.Start()-self.Offset, 0):util.Min(seg.End()-self.Offset, len(self.Qual))]...)
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
