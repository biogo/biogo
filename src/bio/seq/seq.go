// Basic sequence package
//
// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
//   This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU General Public License for more details.
//
//   You should have received a copy of the GNU General Public License
//   along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
package seq

import (
	"bio"
	"bio/util"
	"bio/featgroup"
	"bio/interval"
)

const (
	Prepend = iota
	Append
)

var (
	moltypesToString = [...]string{
		"dna", "rna", "protein",
	}

	complement = [...]map[byte]byte{
		{'a': 't', 'c': 'g', 'g': 'c', 't': 'a', 'n': 'n',
			'A': 'T', 'C': 'G', 'G': 'C', 'T': 'A', 'N': 'N'}, // DNA rules
		{'a': 'u', 'c': 'g', 'g': 'c', 'u': 'a', 'n': 'n',
			'A': 'U', 'C': 'G', 'G': 'C', 'U': 'A', 'N': 'N'}, // RNA rules
	}
)

type MetaFunc func(interface{}) interface{}

type Seq struct {
	ID       []byte
	Seq      []byte
	Offset   int
	Strand   int8
	Circular bool
	Moltype  byte
	Quality  *Quality
	Meta     interface{}
	MetaHook map[string]MetaFunc // at this point RevComp and Trunc call function on Meta with those keys if they exist
}

func New(id, seq []byte, qual *Quality) *Seq {
	return &Seq{
		ID:       id,
		Seq:      seq,
		Offset:   0,
		Strand:   1,
		Circular: false,
		Moltype:  0,
		Quality:  qual,
	}
}

func (self *Seq) Len() int {
	return len(self.Seq)
}

func (self *Seq) Start() int {
	return self.Offset
}

func (self *Seq) End() int {
	return self.Offset + len(self.Seq)
}

func (self *Seq) MoltypeAsString() string {
	return moltypesToString[self.Moltype]
}

func (self *Seq) Trunc(start, end int) (s *Seq, err error) {
	var ts []byte

	if start < self.Offset || end < self.Offset ||
		start > len(self.Seq)+self.Offset || end > len(self.Seq)+self.Offset {
		return nil, bio.NewError("Start or end position out of range", 0, self)
	}

	if !self.Circular || start <= end {
		ts = make([]byte, end-start)
		copy(ts, self.Seq[start-self.Offset:end-self.Offset])
	} else if self.Circular {
		ts = make([]byte, len(self.Seq)-end, end-start)
		copy(ts, self.Seq[start-self.Offset:])
		ts = append(ts, self.Seq[:end-self.Offset]...)
	} else {
		return nil, bio.NewError("Start position greater than end position for non-circular molecule", 0, self)
	}

	var q *Quality
	if self.Quality != nil {
		if q, err = self.Quality.Trunc(start, end); err != nil {
			err = bio.NewError("Quality.Trunc() returned error", 0, err)
			return
		}
	}

	var meta interface{}
	if self.MetaHook != nil {
		method := util.Name(1).Function
		if _, ok := self.MetaHook[method]; ok {
			meta = self.MetaHook[method](self.Meta)
		} else {
			meta = self.Meta
		}
	} else {
		meta = self.Meta
	}

	return &Seq{
		ID:       self.ID,
		Seq:      ts,
		Offset:   start,
		Strand:   self.Strand,
		Circular: false,
		Moltype:  self.Moltype,
		Quality:  q,
		Meta:     meta,
	}, nil
}

func (self *Seq) RevComp() (s *Seq, err error) {
	rs := make([]byte, len(self.Seq))

	if self.Moltype == bio.DNA || self.Moltype == bio.RNA {
		for i, j := 0, len(self.Seq)-1; i < len(self.Seq); i, j = i+1, j-1 {
			rs[i] = complement[self.Moltype][self.Seq[j]]
		}
	} else {
		return nil, bio.NewError("Cannot reverse complement protein", 0, self)
	}

	var q *Quality
	if self.Quality != nil {
		q = self.Quality.Reverse()
	}

	var meta interface{}
	if self.MetaHook != nil {
		method := util.Name(1).Function
		if _, ok := self.MetaHook[method]; ok {
			meta = self.MetaHook[method](self.Meta)
		} else {
			meta = self.Meta
		}
	} else {
		meta = self.Meta
	}

	return &Seq{
		ID:       self.ID,
		Seq:      rs,
		Offset:   0,
		Strand:   -self.Strand,
		Circular: self.Circular,
		Moltype:  self.Moltype,
		Quality:  q,
		Meta:     meta,
	}, nil
}

func (self *Seq) Join(s *Seq, where int) {
	switch where {
	case Prepend:
		s := make([]byte, len(s.Seq), len(s.Seq)+len(self.Seq))
		copy(s, self.Seq)
		self.Seq = append(s, self.Seq...)
		// do anything to Meta of self that would be offset dependent
	case Append:
		self.Seq = append(self.Seq, s.Seq...)
		// do anything to Meta from s that would be offset dependent
		// and copy to into self.Meta
	}
}

func (self *Seq) Stitch(f *featgroup.FeatureGroup) (s *Seq, err error) {
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
	i, _ = interval.New("EOS", self.End()+1, self.End()+1, 0, nil)
	t.Insert(i)

	s = &Seq{}
	var (
		start int
		last  *interval.Interval
	)

	if span, err := interval.New("Sequence", self.Start(), self.End(), 0, nil); err != nil {
		return nil, bio.NewError("Seq.End() < Seq.Start()", 0, self)
	} else {
		for segment := range t.Intersect(span, 1) {
			if last == nil { // start of the features
				start = segment.Start()
			}
			if last.End() < segment.Start()-1 { // at least one position gap between this feature and the last
				s.Seq = append(s.Seq, self.Seq[util.Max(0, start-self.Offset):last.End()-self.Offset]...)
				start = segment.Start()
			}
			if segment.End() > self.End() { // this is the last useful segment
				s.Seq = append(s.Seq, self.Seq[util.Max(0, start-self.Offset):util.Min(len(self.Seq), last.End()-self.Offset)]...)
				break
			}
			last = segment
		}
	}

	self.Seq = s.Seq
	self.Offset = 0

	return self, nil
}

var defaultStringFunc = func(s *Seq) string { return string(s.Seq) }

var StringFunc = defaultStringFunc

func (self *Seq) String() string {
	return StringFunc(self)
}
