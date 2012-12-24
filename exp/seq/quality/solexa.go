// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package quality

import (
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/seq"
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
	seq.Annotation
	Qual   Qsolexas
	Encode alphabet.Encoding
}

// Create a new scoring type.
func NewSolexa(id string, q []alphabet.Qsolexa, encode alphabet.Encoding) *Solexa {
	return &Solexa{
		Annotation: seq.Annotation{ID: id},
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
func (q *Solexa) At(i int) alphabet.Qsolexa { return q.Qual[i-q.Offset] }

// Return the error probability at position pos.
func (q *Solexa) EAt(i int) float64 { return q.Qual[i-q.Offset].ProbE() }

// Set the raw score at position pos to qual.
func (q *Solexa) Set(i int, qual alphabet.Qsolexa) { q.Qual[i-q.Offset] = qual }

// Set the error probability to e at position pos.
func (q *Solexa) SetE(i int, e float64) {
	q.Qual[i-q.Offset] = alphabet.Esolexa(e)
}

// Encode the quality at position pos to a letter based on the sequence Encode setting.
func (q *Solexa) QEncode(i int) byte {
	return q.Qual[i-q.Offset].Encode(q.Encode)
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
