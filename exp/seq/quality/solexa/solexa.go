// Package solexa provides support for manipulation of quality data in Solexa format.
//
// This package is not used directly by any other quality containing sequence types.
package quality

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
	"github.com/kortschak/biogo/bio"
	"github.com/kortschak/biogo/exp/alphabet"
	"github.com/kortschak/biogo/exp/seq"
	"github.com/kortschak/biogo/exp/seq/sequtils"
	"github.com/kortschak/biogo/feat"
)

type Appender interface {
	Append(a ...alphabet.Qsolexa) (err error)
}

// Solexa quality data - does not satisfy Quality interface, but included for completeness.
type Solexa struct {
	ID        string
	Desc      string
	Loc       string
	S         []alphabet.Qsolexa
	Stringify seq.Stringify
	Meta      interface{} // No operation implicitly copies or changes the contents of Meta.
	offset    int
	circular  bool
}

func NewSolexa(id string, q []alphabet.Qsolexa) *Solexa {
	return &Solexa{
		ID: id,
		S:  append([]alphabet.Qsolexa(nil), q...),
		Stringify: func(q seq.Polymer) string {
			t := q.(*Solexa)
			qs := make([]byte, 0, len(t.S))
			for _, s := range t.S {
				qs = append(qs, s.Encode(alphabet.Solexa))
			}
			return string(qs)
		},
	}
}

// Name returns a pointer to the ID string of the sequence.
func (self *Solexa) Name() *string { return &self.ID }

// Description returns a pointer to the Desc string of the sequence.
func (self *Solexa) Description() *string { return &self.Desc }

// Location returns a pointer to the Loc string of the sequence.
func (self *Solexa) Location() *string { return &self.Loc }

// Raw returns the underlying []alphabet.Qsolexa slice.
func (self *Solexa) Raw() interface{} { return self.S }

func (self *Solexa) Append(a ...alphabet.Qsolexa) { self.S = append(self.S, a...) }

func (self *Solexa) At(pos seq.Position) alphabet.Qsolexa { return self.S[pos.Pos-self.offset] }

func (self *Solexa) EAt(pos seq.Position) float64 { return self.S[pos.Pos-self.offset].ProbE() }

func (self *Solexa) Set(pos seq.Position, l alphabet.Qsolexa) { self.S[pos.Pos-self.offset] = l }

func (self *Solexa) SetE(pos seq.Position, l float64) {
	self.S[pos.Pos-self.offset] = alphabet.Esolexa(l)
}

// Encode the quality at position pos to a letter based on the sequence encoding setting. Only encodes to Solexa.
func (self *Solexa) QEncode(pos seq.Position) byte {
	return self.S[pos.Pos-self.offset].Encode(alphabet.Solexa)
}

// Decode a quality letter to a phred score based on the sequence encoding setting. Only decodes from Solexa.
func (self *Solexa) QDecode(l byte) alphabet.Qsolexa {
	return alphabet.DecodeToQsolexa(l, alphabet.Solexa)
}

// Return the quality encoding type.
func (self *Solexa) Encoding() alphabet.Encoding { return alphabet.Solexa }

// Set the quality encoding type to e. No-op at this stage.
func (self *Solexa) SetEncoding(e alphabet.Encoding) {}

func (self *Solexa) Len() int { return len(self.S) }

func (self *Solexa) Offset(o int) { self.offset = o }

func (self *Solexa) Start() int { return self.offset }

func (self *Solexa) End() int { return self.offset + self.Len() }

func (self *Solexa) Count() int { return 1 }

func (self *Solexa) Copy() *Solexa {
	c := *self
	c.S = append([]alphabet.Qsolexa(nil), self.S...)
	c.Meta = nil

	return &c
}

func (self *Solexa) Reverse() { self.S = sequtils.Reverse(self.S).([]alphabet.Qsolexa) }

func (self *Solexa) Circular(c bool) { self.circular = c }

func (self *Solexa) IsCircular() bool { return self.circular }

// Return a subsequence from start to end, wrapping if the sequence is circular.
func (self *Solexa) Subseq(start int, end int) (q *Solexa, err error) {
	tt, err := sequtils.Truncate(self.S, start-self.offset, end-self.offset, self.circular)
	if err == nil {
		q = &Solexa{}
		*q = *self
		q.S = tt.([]alphabet.Qsolexa)
		q.S = nil
		q.Meta = nil
		q.offset = start
		q.circular = false
	}

	return
}

func (self *Solexa) Truncate(start int, end int) (err error) {
	tt, err := sequtils.Truncate(self.S, start-self.offset, end-self.offset, self.circular)
	if err == nil {
		self.S = tt.([]alphabet.Qsolexa)
		self.offset = start
		self.circular = false
	}

	return
}

func (self *Solexa) Join(p *Solexa, where int) (err error) {
	if self.circular {
		return bio.NewError("Cannot join circular sequence: receiver.", 1, self)
	} else if p.circular {
		return bio.NewError("Cannot join circular sequence: parameter.", 1, p)
	}

	var tt interface{}

	tt, self.offset = sequtils.Join(self.S, p.S, where)
	self.S = tt.([]alphabet.Qsolexa)

	return
}

func (self *Solexa) Stitch(f feat.FeatureSet) (err error) {
	tt, err := sequtils.Stitch(self.S, self.offset, f)
	if err == nil {
		self.S = tt.([]alphabet.Qsolexa)
		self.circular = false
		self.offset = 0
	}

	return
}

func (self *Solexa) Compose(f feat.FeatureSet) (err error) {
	tt, err := sequtils.Compose(self.S, self.offset, f)
	if err == nil {
		s := []alphabet.Qsolexa{}
		for i, ts := range tt {
			if f[i].Strand == -1 {
				s = append(s, sequtils.Reverse(ts).([]alphabet.Qsolexa)...)
			} else {
				s = append(s, ts.([]alphabet.Qsolexa)...)
			}
		}

		self.S = s
		self.circular = false
		self.offset = 0
	}

	return
}

func (self *Solexa) String() string { return self.Stringify(self) }
