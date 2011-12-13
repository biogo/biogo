// Basic sequence package
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
)

const (
	Prepend = iota
	Append
)

var (
	complement = [...]map[byte]byte{
		{'a': 't', 'c': 'g', 'g': 'c', 't': 'a', 'n': 'n',
			'A': 'T', 'C': 'G', 'G': 'C', 'T': 'A', 'N': 'N'}, // DNA rules
		{'a': 'u', 'c': 'g', 'g': 'c', 'u': 'a', 'n': 'n',
			'A': 'U', 'C': 'G', 'G': 'C', 'U': 'A', 'N': 'N'}, // RNA rules
	}
)

type Seq struct {
	ID       string
	Seq      []byte
	Offset   int
	Strand   int8
	Circular bool
	Moltype  bio.Moltype
	Quality  *Quality
	Meta     interface{} // No operation on Seq objects implicitly copies or changes the contents of Meta.
}

func New(id string, seq []byte, qual *Quality) *Seq {
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

func (self *Seq) Trunc(start, end int) (s *Seq, err error) {
	var ts []byte

	if start < self.Offset || end < self.Offset ||
		start > len(self.Seq)+self.Offset || end > len(self.Seq)+self.Offset {
		return nil, bio.NewError("Start or end position out of range.", 0, self)
	}

	if start <= end {
		ts = append([]byte{}, self.Seq[start-self.Offset:end-self.Offset]...)
	} else if self.Circular {
		ts = make([]byte, len(self.Seq)-start-self.Offset, len(self.Seq)+end-start)
		copy(ts, self.Seq[start-self.Offset:])
		ts = append(ts, self.Seq[:end-self.Offset]...)
	} else {
		return nil, bio.NewError("Start position greater than end position for non-circular molecule.", 0, self)
	}

	var q *Quality
	if self.Quality != nil {
		if q, err = self.Quality.Trunc(start, end); err != nil {
			err = bio.NewError("Quality.Trunc() returned error", 0, err)
			return
		}
	}

	return &Seq{
		ID:       self.ID,
		Seq:      ts,
		Offset:   start,
		Strand:   self.Strand,
		Circular: false,
		Moltype:  self.Moltype,
		Quality:  q,
	}, nil
}

func (self *Seq) RevComp() (s *Seq, err error) {
	rs := make([]byte, len(self.Seq))

	if self.Moltype == bio.DNA || self.Moltype == bio.RNA {
		for i, j := 0, len(self.Seq)-1; i < len(self.Seq); i, j = i+1, j-1 {
			rs[i] = complement[self.Moltype][self.Seq[j]]
		}
	} else {
		return nil, bio.NewError("Cannot reverse-complement protein.", 0, self)
	}

	var q *Quality
	if self.Quality != nil {
		q = self.Quality.Reverse()
	}

	return &Seq{
		ID:       self.ID,
		Seq:      rs,
		Offset:   0,
		Strand:   -self.Strand,
		Circular: self.Circular,
		Moltype:  self.Moltype,
		Quality:  q,
	}, nil
}

func (self *Seq) Join(s *Seq, where int) (err error) {
	if self.Circular {
		return bio.NewError("Cannot join circular molecule.", 0, self)
	}
	switch where {
	case Prepend:
		ts := make([]byte, len(s.Seq), len(s.Seq)+len(self.Seq))
		copy(ts, s.Seq)
		self.Seq = append(ts, self.Seq...)
	case Append:
		self.Seq = append(self.Seq, s.Seq...)
	}

	return
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

	ts := []byte{}
	if span, err := interval.New("", self.Start(), self.End(), 0, nil); err != nil {
		panic("Seq.End() < Seq.Start()")
	} else {
		f, _ := t.Flatten(span, 0, 0)
		for _, seg := range f {
			ts = append(ts, self.Seq[util.Max(seg.Start()-self.Offset, 0):util.Min(seg.End()-self.Offset, len(self.Seq))]...)
		}
	}

	if self.Quality != nil {
		self.Quality.Stitch(f)
	}

	self.Seq = ts
	self.Offset = 0

	return self, nil
}

var defaultStringFunc = func(s *Seq) string { return string(s.Seq) }

var StringFunc = defaultStringFunc

func (self *Seq) String() string {
	return StringFunc(self)
}
