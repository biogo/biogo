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
	"code.google.com/p/biogo/exp/seq/linear"
)

// A slice of quality scores that satisfies the alphabet.Slice interface.
type Qphreds []alphabet.Qphred

func (q Qphreds) Make(len, cap int) alphabet.Slice    { return make(Qphreds, len, cap) }
func (q Qphreds) Len() int                            { return len(q) }
func (q Qphreds) Cap() int                            { return cap(q) }
func (q Qphreds) Slice(start, end int) alphabet.Slice { return q[start:end] }
func (q Qphreds) Append(a alphabet.Slice) alphabet.Slice {
	return append(q, a.(Qphreds)...)
}
func (q Qphreds) Copy(a alphabet.Slice) int { return copy(q, a.(Qphreds)) }

type Phred struct {
	linear.Annotation
	Qual   Qphreds
	Encode alphabet.Encoding
}

// Create a new scoring type.
func NewPhred(id string, q []alphabet.Qphred, encode alphabet.Encoding) *Phred {
	return &Phred{
		Annotation: linear.Annotation{ID: id},
		Qual:       append([]alphabet.Qphred(nil), q...),
		Encode:     encode,
	}
}

// Returns the underlying quality score slice.
func (q *Phred) Slice() alphabet.Slice { return q.Qual }

// Set the underlying quality score slice.
func (q *Phred) SetSlice(sl alphabet.Slice) { q.Qual = sl.(Qphreds) }

// Append to the scores.
func (q *Phred) Append(a ...alphabet.Qphred) { q.Qual = append(q.Qual, a...) }

// Return the raw score at position pos.
func (q *Phred) At(pos seq.Position) alphabet.Qphred { return q.Qual[pos.Col-q.Offset] }

// Return the error probability at position pos.
func (q *Phred) EAt(pos seq.Position) float64 { return q.Qual[pos.Col-q.Offset].ProbE() }

// Set the raw score at position pos to qual.
func (q *Phred) Set(pos seq.Position, qual alphabet.Qphred) { q.Qual[pos.Col-q.Offset] = qual }

// Set the error probability to e at position pos.
func (q *Phred) SetE(pos seq.Position, e float64) {
	q.Qual[pos.Col-q.Offset] = alphabet.Ephred(e)
}

// Encode the quality at position pos to a letter based on the sequence Encode setting.
func (q *Phred) QEncode(pos seq.Position) byte {
	return q.Qual[pos.Col-q.Offset].Encode(q.Encode)
}

// Decode a quality letter to a phred score based on the sequence Encode setting.
func (q *Phred) QDecode(l byte) alphabet.Qphred { return q.Encode.DecodeToQphred(l) }

// Return the quality Encode type.
func (q *Phred) Encoding() alphabet.Encoding { return q.Encode }

// Set the quality Encode type to e.
func (q *Phred) SetEncoding(e alphabet.Encoding) { q.Encode = e }

// Return the lenght of the score sequence.
func (q *Phred) Len() int { return len(q.Qual) }

// Return the start position of the score sequence.
func (q *Phred) Start() int { return q.Offset }

// Return the end position of the score sequence.
func (q *Phred) End() int { return q.Offset + q.Len() }

// Return a copy of the quality sequence.
func (q *Phred) Copy() seq.Quality {
	c := *q
	c.Qual = append([]alphabet.Qphred(nil), q.Qual...)

	return &c
}

// Reverse the order of elements in the sequence.
func (q *Phred) Reverse() {
	l := q.Qual
	for i, j := 0, len(l)-1; i < j; i, j = i+1, j-1 {
		l[i], l[j] = l[j], l[i]
	}
}

func (q *Phred) String() string {
	qs := make([]byte, 0, len(q.Qual))
	for _, s := range q.Qual {
		qs = append(qs, s.Encode(q.Encode))
	}
	return string(qs)
}
