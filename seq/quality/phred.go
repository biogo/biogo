// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package quality

import (
	"code.google.com/p/biogo/alphabet"
	"code.google.com/p/biogo/seq"
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
	seq.Annotation
	Qual   Qphreds
	Encode alphabet.Encoding
}

// Create a new scoring type.
func NewPhred(id string, q []alphabet.Qphred, encode alphabet.Encoding) *Phred {
	return &Phred{
		Annotation: seq.Annotation{ID: id},
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
func (q *Phred) At(i int) alphabet.Qphred { return q.Qual[i-q.Offset] }

// Return the error probability at position pos.
func (q *Phred) EAt(i int) float64 { return q.Qual[i-q.Offset].ProbE() }

// Set the raw score at position pos to qual.
func (q *Phred) Set(i int, qual alphabet.Qphred) { q.Qual[i-q.Offset] = qual }

// Set the error probability to e at position pos.
func (q *Phred) SetE(i int, e float64) {
	q.Qual[i-q.Offset] = alphabet.Ephred(e)
}

// Encode the quality at position pos to a letter based on the sequence Encode setting.
func (q *Phred) QEncode(i int) byte {
	return q.Qual[i-q.Offset].Encode(q.Encode)
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
