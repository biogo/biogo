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

package quality

import (
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/exp/seq/nucleic"
)

// A slice of quality scores that satisfies the alphabet.Slice interface.
type Qsolexas []alphabet.Qsolexa

func (q Qsolexas) Make(len, cap int) alphabet.Slice    { return make(Qsolexas, len, cap) }
func (q Qsolexas) Len() int                            { return len(q) }
func (q Qsolexas) Cap() int                            { return cap(q) }
func (q Qsolexas) Slice(start, end int) alphabet.Slice { return q[start:end] }
func (q Qsolexas) Append(a alphabet.Slice) alphabet.Slice {
	return append(q, a.(Qsolexas)...)
}
func (q Qsolexas) Copy(a alphabet.Slice) int { return copy(q, a.(Qsolexas)) }

type Solexa struct {
	nucleic.Annotation
	Qual   Qsolexas
	Encode alphabet.Encoding
}

// Create a new scoring type.
func NewSolexa(id string, q []alphabet.Qsolexa, encode alphabet.Encoding) *Solexa {
	return &Solexa{
		Annotation: nucleic.Annotation{ID: id},
		Qual:       append([]alphabet.Qsolexa(nil), q...),
		Encode:     encode,
	}
}

// Returns the underlying quality score slice.
func (q *Solexa) Slice() alphabet.Slice { return q.Qual }

// Set the underlying quality score slice.
func (q *Solexa) SetSlice(sl alphabet.Slice) { q.Qual = sl.(Qsolexas) }

// Append to the scores.
func (q *Solexa) Append(a ...alphabet.Qsolexa) { q.Qual = append(q.Qual, a...) }

// Return the raw score at position pos.
func (q *Solexa) At(pos seq.Position) alphabet.Qsolexa { return q.Qual[pos.Col-q.Offset] }

// Return the error probability at position pos.
func (q *Solexa) EAt(pos seq.Position) float64 { return q.Qual[pos.Col-q.Offset].ProbE() }

// Set the raw score at position pos to qual.
func (q *Solexa) Set(pos seq.Position, qual alphabet.Qsolexa) { q.Qual[pos.Col-q.Offset] = qual }

// Set the error probability to e at position pos.
func (q *Solexa) SetE(pos seq.Position, e float64) {
	q.Qual[pos.Col-q.Offset] = alphabet.Esolexa(e)
}

// Encode the quality at position pos to a letter based on the sequence Encode setting.
func (q *Solexa) QEncode(pos seq.Position) byte {
	return q.Qual[pos.Col-q.Offset].Encode(q.Encode)
}

// Decode a quality letter to a phred score based on the sequence Encode setting.
func (q *Solexa) QDecode(l byte) alphabet.Qsolexa { return q.Encode.DecodeToQsolexa(l) }

// Return the quality Encode type.
func (q *Solexa) Encoding() alphabet.Encoding { return q.Encode }

// Set the quality Encode type to e.
func (q *Solexa) SetEncoding(e alphabet.Encoding) { q.Encode = e }

// Return the lenght of the score sequence.
func (q *Solexa) Len() int { return len(q.Qual) }

// Return the start position of the score sequence.
func (q *Solexa) Start() int { return q.Offset }

// Return the end position of the score sequence.
func (q *Solexa) End() int { return q.Offset + q.Len() }

// Return a copy of the quality sequence.
func (q *Solexa) Copy() seq.Quality {
	c := *q
	c.Qual = append([]alphabet.Qsolexa(nil), q.Qual...)

	return &c
}

// Reverse the order of elements in the sequence.
func (q *Solexa) Reverse() {
	l := q.Qual
	for i, j := 0, len(l)-1; i < j; i, j = i+1, j-1 {
		l[i], l[j] = l[j], l[i]
	}
}

func (q *Solexa) String() string {
	qs := make([]byte, 0, len(q.Qual))
	for _, s := range q.Qual {
		qs = append(qs, s.Encode(q.Encode))
	}
	return string(qs)
}
