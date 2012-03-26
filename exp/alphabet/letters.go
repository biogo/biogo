package alphabet

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
	"unsafe"
)

type QLetter struct {
	L Letter
	Q Qphred
}

// Pack a QLetter into a QPack. a.Len() == 4.
func (self QLetter) Pack(a Nucleic) (QPack, error) {
	if a.Len() != 4 {
		return 0, bio.NewError("Invalid alphabet", 0, self)
	}
	if !a.IsValid(self.L) {
		return QPack(byte(self.Q) << 2), bio.NewError("Invalid letter", 0, self)
	}
	return QPack(byte(self.Q)<<2 | byte(a.IndexOf(self.L)&0x3)), nil
}

const logThreshQL = 1e2 // Approximate count where range loop becomes slower than copy

// Repeat a QLetter count times.
func (self QLetter) Repeat(count int) (r []QLetter) {
	r = make([]QLetter, count)
	switch {
	case count == 0:
	case count < logThreshQL:
		for i := range r {
			r[i] = self
		}
	default:
		r[0] = self
		for i := 1; i < len(r); {
			i += copy(r[i:], r[:i])
		}
	}

	return
}

type Encoding int8

const (
	Sanger Encoding = iota
	Solexa
	Illumina1_3
	Illumina1_5
	Illumina1_8
	Illumina1_9
)

type Letter byte

const logThreshL = 2e2 // Approximate count where range loop becomes slower than copy

// Repeat a Letter count times.
func (self Letter) Repeat(count int) (r []Letter) {
	r = make([]Letter, count)
	switch {
	case count == 0:
	case count < logThreshL:
		for i := range r {
			r[i] = self
		}
	default:
		r[0] = self
		for i := 1; i < len(r); {
			i += copy(r[i:], r[:i])
		}
	}

	return
}

func BytesToLetters(b []byte) []Letter { return *(*[]Letter)(unsafe.Pointer(&b)) }

func LettersToBytes(l []Letter) []byte { return *(*[]byte)(unsafe.Pointer(&l)) }

type Letters []Letter

func (self Letters) String() string { return string(LettersToBytes(self)) }

type Qscore interface {
	ProbE() float64
	Encode(Encoding) byte
	String() string
}

type Qphred int8

func DecodeToQphred(q byte, encoding Encoding) (p Qphred) {
	switch encoding {
	case Sanger, Illumina1_8, Illumina1_9:
		return Qphred(q) - 33
	case Illumina1_3, Illumina1_5:
		return Qphred(q) - 64
	case Solexa:
		return (Qsolexa(q) - 64).Qphred()
	}

	panic("cannot reach")
}

func Ephred(p float64) Qphred {
	if p == 0 {
		return 127
	}
	Q := -10 * math.Log10(p)
	if Q > 127 {
		Q = 127
	}
	return Qphred(Q)
}

func (self Qphred) ProbE() float64 {
	if self == 127 {
		return 0
	}
	return math.Pow(10, -(float64(self) / 10))
}

func (self Qphred) Qsolexa() Qsolexa {
	if self == 127 {
		return 127
	}
	return Qsolexa(10 * math.Log10(math.Pow(10, float64(self)/10)-1))
}

func (self Qphred) Encode(encoding Encoding) (q byte) {
	if self == 127 {
		return 126
	}
	switch encoding {
	case Sanger, Illumina1_8, Illumina1_9:
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
		return q
	case Solexa:
		q = byte(self.Qsolexa())
		if q <= 62 {
			q += 64
		}
	}

	return
}

func (self Qphred) String() string {
	if self < 127 {
		return string([]byte{byte(self)})
	}
	return "\u221e"
}

type Qsolexa int8

func DecodeToQsolexa(q byte, encoding Encoding) (p Qsolexa) {
	switch encoding {
	case Sanger, Illumina1_8, Illumina1_9:
		return (Qphred(q) - 33).Qsolexa()
	case Illumina1_3, Illumina1_5:
		return (Qphred(q) - 64).Qsolexa()
	case Solexa:
		return Qsolexa(q) - 64
	}

	panic("cannot reach")
}

func Esolexa(p float64) Qsolexa {
	if p == 0 {
		return 127
	}
	return Qsolexa(-10 * math.Log10(p/(1-p)))
}

func (self Qsolexa) ProbE() float64 {
	if self == 127 {
		return 0
	}
	pq := math.Pow(10, -(float64(self) / 10))
	return pq / (1 + pq)
}

func (self Qsolexa) Qphred() Qphred {
	if self == 127 {
		return 127
	}
	return Qphred(10 * math.Log10(math.Pow(10, float64(self)/10)+1))
}

func (self Qsolexa) Encode(encoding Encoding) (q byte) {
	if self == 127 {
		return 126
	}
	switch encoding {
	case Sanger, Illumina1_8:
		q = byte(self.Qphred())
		if q <= 93 {
			q += 33
		}
	case Illumina1_3:
		q = byte(self.Qphred())
		if q <= 62 {
			q += 64
		}
	case Illumina1_5:
		q = byte(self.Qphred())
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

	return
}

func (self Qsolexa) String() string {
	if self < 127 {
		return string([]byte{byte(self)})
	}
	return "\u221e"
}

// Bitpacked sequence. 2 bits per base.
type Pack byte

// Bitpacked quality base. Bits 0 and 1 encode base, 2-7 encode quality.
// Behaviour is undefined if alphabet used does not satisfy Len() == 4.
type QPack byte

// Upack a QPack to a QLetter.
func (self QPack) Unpack(a Nucleic) QLetter {
	return QLetter{
		L: a.Letter(int(self & 0x3)),
		Q: Qphred(self >> 2),
	}
}
