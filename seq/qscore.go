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
	"math"
)

type Encoding byte

const (
	Sanger = iota
	Solexa
	Illumina1_3
	Illumina1_5
	Illumina1_8
)

type Qscore interface {
	Encode(Encoding) byte
	ProbE() float64
}

type Qsanger int8

func (self Qsanger) ToSolexa() Qsolexa {
	return Qsolexa(10 * math.Log10(math.Pow(10, float64(self)/10)-1))
}

func (self Qsanger) Encode(encoding Encoding) (q byte) {
	switch encoding {
	case Sanger, Illumina1_8:
		q = byte(self)
		if q <= 93 {
			q += 33
		}
	case Illumina1_3:
		q = byte(self)
		if q <= 62 {
			q += 64
		}
	case Illumina1_5:
		q = byte(self)
		if q <= 62 {
			q += 64
		}
		if q < 'B' {
			q = 'B'
		}
	case Solexa:
		q = byte(self.ToSolexa())
		if q <= 62 {
			q += 64
		}
	}
	return q
}

func (self Qsanger) ProbE() float64 {
	return math.Pow(10, -(float64(self) / 10))
}

type Qsolexa int8

func (self Qsolexa) ToSanger() Qsanger {
	return Qsanger(10 * math.Log10(math.Pow(10, float64(self)/10)+1))
}

func (self Qsolexa) Encode(encoding Encoding) (q byte) {
	switch encoding {
	case Sanger, Illumina1_8:
		q = byte(self.ToSanger())
		if q <= 93 {
			q += 33
		}
	case Illumina1_3:
		q = byte(self.ToSanger())
		if q <= 62 {
			q += 64
		}
	case Illumina1_5:
		q = byte(self.ToSanger())
		if q <= 62 {
			q += 64
		}
		if q < 'B' {
			q = 'B'
		}
	case Solexa:
		q = byte(self)
		if q <= 62 {
			q += 64
		}
	}
	return q
}

func (self Qsolexa) ProbE() float64 {
	pq := math.Pow(10, -(float64(self) / 10))
	return pq / (1 + pq)
}
