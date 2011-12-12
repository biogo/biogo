// Basic sequence quality package
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

	if start < self.Offset || end < self.Offset || start > len(self.Qual)+self.Offset || end > len(self.Qual)+self.Offset {
		return nil, bio.NewError("Start or end position out of range", 0, self)
	}

	if !self.Circular || start <= end {
		ts = make([]int8, end-start)
		copy(ts, self.Qual[start:end])
	} else if self.Circular {
		ts = make([]int8, len(self.Qual)-end, end-start)
		copy(ts, self.Qual[start:])
		ts = append(ts, self.Qual[:end]...)
	} else {
		return nil, bio.NewError("Start position greater than end position for non-circular molecule", 0, self)
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
