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
	"github.com/kortschak/BioGo/feat"
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
	Inplace  bool
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
		Inplace:  false,
	}
}

// Set Inplace and return self for chaining.
func (self *Seq) WorkInplace(b bool) *Seq {
	self.Inplace = b
	return self
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

	if !self.Inplace && self.Quality != nil && self.Quality.Inplace {
		return nil, bio.NewError("Inplace operation on Quality with non-Inplace operation on parent Seq.", 0, self)
	}

	if start < self.Offset || end < self.Offset ||
		start > len(self.Seq)+self.Offset || end > len(self.Seq)+self.Offset {
		return nil, bio.NewError("Start or end position out of range.", 0, self)
	}

	if start <= end {
		if self.Inplace {
			ts = self.Seq[start-self.Offset : end-self.Offset]
		} else {
			ts = append([]byte{}, self.Seq[start-self.Offset:end-self.Offset]...)
		}
	} else if self.Circular {
		if self.Inplace {
			ts = append(self.Seq[start-self.Offset:], self.Seq[:end-self.Offset]...) // not quite inplace for this op
		} else {
			ts = make([]byte, len(self.Seq)-start-self.Offset, len(self.Seq)+end-start)
			copy(ts, self.Seq[start-self.Offset:])
			ts = append(ts, self.Seq[:end-self.Offset]...)
		}
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

	if self.Inplace {
		s = self
		s.Seq = ts
		s.Circular = false
		s.Quality = q
	} else {
		s = &Seq{
			ID:       self.ID,
			Seq:      ts,
			Offset:   start,
			Strand:   self.Strand,
			Circular: false,
			Moltype:  self.Moltype,
			Quality:  q,
		}
	}

	return
}

func (self *Seq) RevComp() (s *Seq, err error) {
	var rs []byte
	if self.Inplace {
		rs = self.Seq
	} else {
		if self.Quality != nil && self.Quality.Inplace {
			return nil, bio.NewError("Inplace operation on Quality with non-Inplace operation on parent Seq.", 0, self)
		}
		rs = make([]byte, len(self.Seq))
	}

	if self.Moltype == bio.DNA || self.Moltype == bio.RNA {
		i, j := 0, len(self.Seq)-1
		for ; i < j; i, j = i+1, j-1 {
			rs[i], rs[j] = complement[self.Moltype][self.Seq[j]], complement[self.Moltype][self.Seq[i]]
		}
		if i == j {
			rs[i] = complement[self.Moltype][self.Seq[i]]
		}
	} else {
		return nil, bio.NewError("Cannot reverse-complement protein.", 0, self)
	}

	var q *Quality
	if self.Quality != nil {
		q = self.Quality.Reverse()
	}

	if self.Inplace {
		s = self
		s.Quality = q
	} else {
		s = &Seq{
			ID:       self.ID,
			Seq:      rs,
			Offset:   self.Offset + len(self.Seq),
			Strand:   -self.Strand,
			Circular: self.Circular,
			Moltype:  self.Moltype,
			Quality:  q,
		}
	}

	return
}

func (self *Seq) Join(s *Seq, where int) (j *Seq, err error) {
	var (
		ts []byte
		ID string
	)

	if self.Circular {
		return nil, bio.NewError("Cannot join circular molecule.", 0, self)
	}

	if !self.Inplace && self.Quality != nil && self.Quality.Inplace {
		return nil, bio.NewError("Inplace operation on Quality with non-Inplace operation on parent Seq.", 0, self)
	}

	switch where {
	case Prepend:
		ID = s.ID + "+" + self.ID
		ts = make([]byte, len(s.Seq), len(s.Seq)+len(self.Seq))
		copy(ts, s.Seq)
		ts = append(ts, self.Seq...)
	case Append:
		ID = self.ID + "+" + s.ID
		if self.Inplace {
			ts = append(self.Seq, s.Seq...)
		} else {
			ts = make([]byte, len(self.Seq), len(s.Seq)+len(self.Seq))
			copy(ts, self.Seq)
			ts = append(ts, s.Seq...)
		}
	}

	var q *Quality
	if self.Quality != nil && s.Quality != nil {
		if q, err = self.Quality.Join(s.Quality, where); err != nil {
			return nil, err
		}
	}

	if self.Inplace {
		j = self
		j.Seq = ts
		j.Quality = q // self.Quality will become nil if either sequence lacks Quality
	} else {
		j = &Seq{
			ID:      ID,
			Seq:     ts,
			Strand:  self.Strand,
			Moltype: self.Moltype,
			Quality: q,
		}
	}
	if where == Prepend {
		j.Offset -= s.Len()
	}

	return
}

func (self *Seq) Stitch(f feat.FeatureSet) (s *Seq, err error) {
	if !self.Inplace && self.Quality != nil && self.Quality.Inplace {
		return nil, bio.NewError("Inplace operation on Quality with non-Inplace operation on parent Seq.", 0, self)
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

	span, err := interval.New("", self.Start(), self.End(), 0, nil)
	if err != nil {
		panic("Seq.End() < Seq.Start()")
	}
	fs, _ := t.Flatten(span, 0, 0)

	if self.Inplace {
		self.Seq = self.stitch(fs)
		self.Offset = 0
		self.Circular = false
		if self.Quality != nil {
			var q *Quality
			if !self.Quality.Inplace {
				q = &Quality{ID: self.Quality.ID}
			}
			q.Qual = self.Quality.stitch(fs)
			q.Offset = 0
			q.Circular = false
			self.Quality = q
		}
		s = self
	} else {
		var q *Quality
		if self.Quality != nil {
			q = &Quality{
				ID:       self.Quality.ID,
				Qual:     self.Quality.stitch(fs),
				Offset:   0,
				Circular: false,
			}
		}
		s = &Seq{
			ID:       self.ID,
			Seq:      self.stitch(fs),
			Offset:   0,
			Strand:   self.Strand,
			Circular: false,
			Moltype:  self.Moltype,
			Quality:  q,
		}
	}

	return
}

func (self *Seq) stitch(f []*interval.Interval) (ts []byte) {
	for _, seg := range f {
		ts = append(ts, self.Seq[util.Max(seg.Start()-self.Offset, 0):util.Min(seg.End()-self.Offset, len(self.Seq))]...)
	}

	return
}

var defaultStringFunc = func(s *Seq) string { return string(s.Seq) }

var StringFunc = defaultStringFunc

func (self *Seq) String() string {
	return StringFunc(self)
}
