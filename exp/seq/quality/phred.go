// Copyright ©2011-2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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

//Package quality provides support for manipulation of quality data in generic Phred format.
package quality

import (
	"code.google.com/p/biogo/bio"
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/exp/seq/sequtils"
	"code.google.com/p/biogo/feat"
)

type Appender interface {
	Append(a ...alphabet.Qphred) (err error)
}

type Phred struct {
	ID        string
	Desc      string
	Loc       string
	S         []alphabet.Qphred
	Stringify seq.Stringify
	Meta      interface{} // No operation implicitly copies or changes the contents of Meta.
	encoding  alphabet.Encoding
	offset    int
	circular  bool
}

func NewPhred(id string, q []alphabet.Qphred, encode alphabet.Encoding) *Phred {
	return &Phred{
		ID:       id,
		S:        append([]alphabet.Qphred(nil), q...),
		encoding: encode,
		Stringify: func(q seq.Polymer) string {
			t := q.(*Phred)
			qs := make([]byte, 0, len(t.S))
			for _, s := range t.S {
				qs = append(qs, s.Encode(t.encoding))
			}
			return string(qs)
		},
	}
}

// Name returns a pointer to the ID string of the sequence.
func (self *Phred) Name() *string { return &self.ID }

// Description returns a pointer to the Desc string of the sequence.
func (self *Phred) Description() *string { return &self.Desc }

// Location returns a pointer to the Loc string of the sequence.
func (self *Phred) Location() *string { return &self.Loc }

// Raw returns the underlying []alphabet.Qphred slice.
func (self *Phred) Raw() interface{} { return self.S }

func (self *Phred) Append(a ...alphabet.Qphred) { self.S = append(self.S, a...) }

func (self *Phred) At(pos seq.Position) alphabet.Qphred { return self.S[pos.Pos-self.offset] }

func (self *Phred) EAt(pos seq.Position) float64 { return self.S[pos.Pos-self.offset].ProbE() }

func (self *Phred) Set(pos seq.Position, q alphabet.Qphred) { self.S[pos.Pos-self.offset] = q }

func (self *Phred) SetE(pos seq.Position, e float64) {
	self.S[pos.Pos-self.offset] = alphabet.Ephred(e)
}

// Encode the quality at position pos to a letter based on the sequence encoding setting.
func (self *Phred) QEncode(pos seq.Position) byte {
	return self.S[pos.Pos-self.offset].Encode(self.encoding)
}

// Decode a quality letter to a phred score based on the sequence encoding setting.
func (self *Phred) QDecode(l byte) alphabet.Qphred { return alphabet.DecodeToQphred(l, self.encoding) }

// Return the quality encoding type.
func (self *Phred) Encoding() alphabet.Encoding { return self.encoding }

// Set the quality encoding type to e.
func (self *Phred) SetEncoding(e alphabet.Encoding) { self.encoding = e }

func (self *Phred) Len() int { return len(self.S) }

func (self *Phred) Offset(o int) { self.offset = o }

func (self *Phred) Start() int { return self.offset }

func (self *Phred) End() int { return self.offset + self.Len() }

func (self *Phred) Copy() seq.Quality {
	c := *self
	c.S = append([]alphabet.Qphred(nil), self.S...)
	c.Meta = nil

	return &c
}

func (self *Phred) Reverse() { self.S = sequtils.Reverse(self.S).([]alphabet.Qphred) }

func (self *Phred) Circular(c bool) { self.circular = c }

func (self *Phred) IsCircular() bool { return self.circular }

// Return a subsequence from start to end, wrapping if the sequence is circular.
func (self *Phred) Subseq(start int, end int) (sub seq.Quality, err error) {
	var q *Phred

	tt, err := sequtils.Truncate(self.S, start-self.offset, end-self.offset, self.circular)
	if err == nil {
		q = &Phred{}
		*q = *self
		q.S = tt.([]alphabet.Qphred)
		q.S = nil
		q.Meta = nil
		q.offset = start
		q.circular = false
	}

	return q, nil
}

func (self *Phred) Truncate(start int, end int) (err error) {
	tt, err := sequtils.Truncate(self.S, start-self.offset, end-self.offset, self.circular)
	if err == nil {
		self.S = tt.([]alphabet.Qphred)
		self.offset = start
		self.circular = false
	}

	return
}

func (self *Phred) Join(p *Phred, where int) (err error) {
	if self.circular {
		return bio.NewError("Cannot join circular sequence: receiver.", 1, self)
	} else if p.circular {
		return bio.NewError("Cannot join circular sequence: parameter.", 1, p)
	}

	var tt interface{}

	tt, self.offset = sequtils.Join(self.S, p.S, where)
	self.S = tt.([]alphabet.Qphred)

	return
}

func (self *Phred) Stitch(f feat.FeatureSet) (err error) {
	tt, err := sequtils.Stitch(self.S, self.offset, f)
	if err == nil {
		self.S = tt.([]alphabet.Qphred)
		self.circular = false
		self.offset = 0
	}

	return
}

func (self *Phred) Compose(f feat.FeatureSet) (err error) {
	tt, err := sequtils.Compose(self.S, self.offset, f)
	if err == nil {
		s := []alphabet.Qphred{}
		for i, ts := range tt {
			if f[i].Strand == -1 {
				s = append(s, sequtils.Reverse(ts).([]alphabet.Qphred)...)
			} else {
				s = append(s, ts.([]alphabet.Qphred)...)
			}
		}

		self.S = s
		self.circular = false
		self.offset = 0
	}

	return
}

func (self *Phred) String() string { return self.Stringify(self) }
