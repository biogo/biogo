// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package seq

import (
	"math"
)

type Encoding byte

const (
	Sanger Encoding = iota
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
