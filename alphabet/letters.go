// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package alphabet

import (
	"math"
	"unsafe"
)

// The Slice interface reflects the built-in slice type behavior.
type Slice interface {
	// Make makes a Slice with the same concrete type as the receiver. Make will
	// panic if len or cap are less than zero or cap is less than len.
	Make(len, cap int) Slice

	// Len returns the length of the Slice.
	Len() int

	// Cap returns the capacity of the Slice.
	Cap() int

	// Slice returns a slice of the Slice. The returned slice may be backed by
	// the same array as the receiver.
	Slice(start, end int) Slice

	// Append appends src... to the receiver and returns the resulting slice. If the append
	// results in a grow slice the receiver will not reflect the appended slice, so the
	// returned Slice should always be stored. Append should panic if src and the receiver
	// are not the same concrete type.
	Append(src Slice) Slice

	// Copy copies elements from src into the receiver, returning the number of elements
	// copied. Copy should panic if src and the receiver are not the same concrete type.
	Copy(src Slice) int
}

// An Encoding represents a quality score encoding scheme.
//                                                                                             Q-range
//
//  Sanger         !"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHI···                                 0 - 40
//  Solexa                                 ··;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_`abcdefgh··· -5 - 40
//  Illumina 1.3+                                 @ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_`abcdefgh···  0 - 40
//  Illumina 1.5+                                 xxḆCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_`abcdefgh···  3 - 40
//  Illumina 1.8+  !"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJ···                                0 - 40
//
//                 !"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_`abcdefgh··· ···{|}~
//                 |                         |    |        |                              |          |
//                33                        59   64       73                            104        126
//
//  Q-range for typical raw reads
type Encoding int8

const (
	None        Encoding = iota - 1 // All letters are decoded as scores with p(Error) = NaN.
	Sanger                          // Phred+33
	Solexa                          // Solexa+64
	Illumina1_3                     // Phred+64
	Illumina1_5                     // Phred+64 0,1=unused, 2=Read Segment Quality Control Indicator (Ḇ)
	Illumina1_8                     // Phred+33
	Illumina1_9                     // Phred+33
)

// DecodeToPhred interprets the byte q as an e encoded quality and returns the corresponding Phred score.
func (e Encoding) DecodeToQphred(q byte) Qphred {
	switch e {
	case Sanger, Illumina1_8, Illumina1_9:
		return Qphred(q) - 33
	case Illumina1_3, Illumina1_5:
		return Qphred(q) - 64
	case Solexa:
		return (Qsolexa(q) - 64).Qphred()
	case None:
		return 0xff
	default:
		panic("alphabet: illegal encoding")
	}
}

// DecodeToPhred interprets the byte q as an e encoded quality and returns the corresponding Solexa score.
func (e Encoding) DecodeToQsolexa(q byte) Qsolexa {
	switch e {
	case Sanger, Illumina1_8, Illumina1_9:
		return (Qphred(q) - 33).Qsolexa()
	case Illumina1_3, Illumina1_5:
		return (Qphred(q) - 64).Qsolexa()
	case Solexa:
		return Qsolexa(q) - 64
	case None:
		return -128
	default:
		panic("alphabet: illegal encoding")
	}
}

// A Letter represents a sequence letter.
type Letter byte

const logThreshL = 2e2 // Approximate count where range loop becomes slower than copy

// Repeat a Letter count times.
func (l Letter) Repeat(count int) []Letter {
	r := make([]Letter, count)
	switch {
	case count == 0:
	case count < logThreshL:
		for i := range r {
			r[i] = l
		}
	default:
		r[0] = l
		for i := 1; i < len(r); {
			i += copy(r[i:], r[:i])
		}
	}

	return r
}

// BytesToLetters converts a []byte to a []Letter.
func BytesToLetters(b []byte) []Letter { return *(*[]Letter)(unsafe.Pointer(&b)) }

// LettersToBytes converts a []Letter to a []byte.
func LettersToBytes(l []Letter) []byte { return *(*[]byte)(unsafe.Pointer(&l)) }

// A Letters is a slice of Letter that satisfies the Slice interface.
type Letters []Letter

func (l Letters) Make(len, cap int) Slice    { return make(Letters, len, cap) }
func (l Letters) Len() int                   { return len(l) }
func (l Letters) Cap() int                   { return cap(l) }
func (l Letters) Slice(start, end int) Slice { return l[start:end] }
func (l Letters) Append(src Slice) Slice     { return append(l, src.(Letters)...) }
func (l Letters) Copy(src Slice) int         { return copy(l, src.(Letters)) }
func (l Letters) String() string             { return string(LettersToBytes(l)) }

// A Columns is a slice of []Letter that satisfies the alphabet.Slice interface.
type Columns [][]Letter

// Make makes a QColumns with the cap and len for each column set to the number of rows of the
// receiver.
func (lc Columns) Make(len, cap int) Slice {
	r := lc.Rows()
	return make(Columns, len, cap).MakeRows(r, r)
}

// MakeRows makes a column with len and cap for each column of the receiver and returns the receiver.
func (lc Columns) MakeRows(len, cap int) Slice {
	for i := range lc {
		lc[i] = make([]Letter, len, cap)
	}
	return lc
}

// Rows returns the number of positions in each column.
func (lc Columns) Rows() int { return len(lc[0]) }

func (lc Columns) Len() int                   { return len(lc) }
func (lc Columns) Cap() int                   { return cap(lc) }
func (lc Columns) Slice(start, end int) Slice { return lc[start:end] }
func (lc Columns) Append(a Slice) Slice {
	// TODO deep copy the columns.
	return append(lc, a.(Columns)...)
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func (lc Columns) Copy(a Slice) int {
	ac := a.(Columns)
	var n int
	for i, src := range ac[:min(len(lc), len(ac))] {
		n += copy(lc[i], src)
	}
	return n
}

// A QLetter represents a sequence letter with an associated quality score.
type QLetter struct {
	L Letter
	Q Qphred
}

const logThreshQL = 1e2 // Approximate count where range loop becomes slower than copy

// Repeat a QLetter count times.
func (ql QLetter) Repeat(count int) []QLetter {
	r := make([]QLetter, count)
	switch {
	case count == 0:
	case count < logThreshQL:
		for i := range r {
			r[i] = ql
		}
	default:
		r[0] = ql
		for i := 1; i < len(r); {
			i += copy(r[i:], r[:i])
		}
	}

	return r
}

func (ql QLetter) String() string { return string(ql.L) }

// A QLetters is a slice of QLetter that satisfies the Slice interface.
type QLetters []QLetter

func (ql QLetters) Make(len, cap int) Slice    { return make(QLetters, len, cap) }
func (ql QLetters) Len() int                   { return len(ql) }
func (ql QLetters) Cap() int                   { return cap(ql) }
func (ql QLetters) Slice(start, end int) Slice { return ql[start:end] }
func (ql QLetters) Append(src Slice) Slice     { return append(ql, src.(QLetters)...) }
func (ql QLetters) Copy(src Slice) int         { return copy(ql, src.(QLetters)) }

// A QColumns is a slice of []QLetter that satisfies the Slice interface.
type QColumns [][]QLetter

// Make makes a QColumns with the cap and len for each column set to the number of rows of the
// receiver.
func (qc QColumns) Make(len, cap int) Slice {
	r := qc.Rows()
	return make(QColumns, len, cap).MakeRows(r, r)
}

// MakeRows makes a column with len and cap for each column of the receiver and returns the receiver.
func (qc QColumns) MakeRows(len, cap int) Slice {
	for i := range qc {
		qc[i] = make([]QLetter, len, cap)
	}
	return qc
}

// Rows returns the number of positions in each column.
func (qc QColumns) Rows() int { return len(qc[0]) }

func (qc QColumns) Len() int                   { return len(qc) }
func (qc QColumns) Cap() int                   { return cap(qc) }
func (qc QColumns) Slice(start, end int) Slice { return qc[start:end] }
func (qc QColumns) Append(a Slice) Slice {
	// TODO deep copy the columns.
	return append(qc, a.(QColumns)...)
}

func (qc QColumns) Copy(a Slice) int {
	ac := a.(QColumns)
	var n int
	for i, src := range ac[:min(len(qc), len(ac))] {
		n += copy(qc[i], src)
	}
	return n
}

// A Qscore represents a quality score.
type Qscore interface {
	ProbE() float64
	Encode(Encoding) byte
	String() string
}

var nan = math.NaN()

// A Qphred represents a Phred quality score.
type Qphred byte

// Ephred returns the Qphred for a error probability p.
func Ephred(p float64) Qphred {
	if p == 0 {
		return 254
	}
	if math.IsNaN(p) {
		return 255
	}
	Q := -10 * math.Log10(p)
	Q += 0.5
	if Q > 254 {
		Q = 254
	}
	return Qphred(Q)
}

// ProbE returns the error probability for the receiver's Phred value.
func (qp Qphred) ProbE() float64 {
	return phredETable[qp]
}

// phredETable holds a lookup for phred E values.
var phredETable = func() [256]float64 {
	t := [256]float64{254: 0, 255: nan}
	for q := range t[:254] {
		t[q] = math.Pow(10, -(float64(q) / 10))
	}
	return t
}()

// Qsolexa converts the quality value from Phred to Solexa. This conversion is lossy and
// should be avoided; the epsilon on the E value associated with a converted Qsolexa is
// bounded approximately by math.Pow(10, 1e-4-float64(qp)/10) over the range 0 < qp < 127.
func (qp Qphred) Qsolexa() Qsolexa { return phredSolexaTable[qp] }

// phredSolexaTable holds a lookup for the near equivalent solexa score of a phred score.
var phredSolexaTable = func() [256]Qsolexa {
	t := [256]Qsolexa{254: 127, 255: -128}
	for q := range t[:254] {
		Q := 10 * math.Log10(math.Pow(10, float64(q)/10)-1)
		if Q > 0 {
			Q += 0.5
		} else {
			Q -= 0.5
		}
		if Q > 127 {
			Q = 127
		}
		t[q] = Qsolexa(Q)
	}
	return t
}()

// Encode encodes the receiver's Phred score to a byte based on the specified encoding.
func (qp Qphred) Encode(e Encoding) (q byte) {
	if qp == 254 {
		return '~'
	}
	if qp == 255 {
		return ' '
	}
	switch e {
	case Sanger, Illumina1_8, Illumina1_9:
		q = byte(qp)
		if q <= 93 {
			q += 33
		}
	case Illumina1_3:
		q = byte(qp)
		if q <= 62 {
			q += 64
		}
	case Illumina1_5:
		q = byte(qp)
		if q <= 62 {
			q += 64
		}
		if q < 'B' {
			q = 'B'
		}
		return q
	case Solexa:
		q = byte(qp.Qsolexa())
		if q <= 62 {
			q += 64
		}
	case None:
		return ' '
	}

	return
}

func (qp Qphred) String() string {
	if qp < 254 {
		return string([]byte{byte(qp)})
	} else if qp == 255 {
		return " "
	}
	return "\u221e"
}

// A Qsolexa represents a Solexa quality score.
type Qsolexa int8

// Esolexa returns the Qsolexa for a error probability p.
func Esolexa(p float64) Qsolexa {
	if p == 0 {
		return 127
	}
	if math.IsNaN(p) {
		return -128
	}
	Q := -10 * math.Log10(p/(1-p))
	if Q > 0 {
		Q += 0.5
	} else {
		Q -= 0.5
	}
	return Qsolexa(Q)
}

// ProbE returns the error probability for the receiver's Phred value.
func (qs Qsolexa) ProbE() float64 { return solexaETable[int(qs)+128] }

// solexaETable holds a translated lookup table for solexa E values. Since solexa
// scores can extend into negative territory, the table is shifted 128 into the
// positive.
var solexaETable = func() [256]float64 {
	t := [256]float64{0: nan, 255: 0}
	for q := range t[1:255] {
		pq := math.Pow(10, -(float64(q-127) / 10))
		t[q+1] = pq / (1 + pq)
	}
	return t
}()

// Qphred converts the quality value from Solexa to Phred. This conversion is lossy and
// should be avoided; the epsilon on the E value associated with a converted Qphred is
// bounded approximately by math.Pow(10, 1e-4-float64(qs)/10) over the range 0 < qs < 127.
func (qs Qsolexa) Qphred() Qphred { return solexaPhredTable[int(qs)+128] }

// solexaPhredTable holds a lookup for the near equivalent phred score of a solexa
// score. Since solexa scores can extend into negative territory, the table is
// shifted 128 into the positive.
var solexaPhredTable = func() [256]Qphred {
	t := [256]Qphred{0: 255, 255: 0}
	for q := range t[1:255] {
		qs := q - 127
		Q := Qphred(10*math.Log10(math.Pow(10, float64(qs)/10)) + 0.5)
		if Q > 254 {
			Q = 254
		}
		t[q+1] = Q
	}
	return t
}()

// Encode encodes the receiver's Solexa score to a byte based on the specified encoding.
func (qs Qsolexa) Encode(e Encoding) (q byte) {
	if qs == 127 {
		return '~'
	}
	if qs == -128 {
		return ' '
	}
	switch e {
	case Sanger, Illumina1_8:
		q = byte(qs.Qphred())
		if q <= 93 {
			q += 33
		}
	case Illumina1_3:
		q = byte(qs.Qphred())
		if q <= 62 {
			q += 64
		}
	case Illumina1_5:
		q = byte(qs.Qphred())
		if q <= 62 {
			q += 64
		}
		if q < 'B' {
			q = 'B'
		}
	case Solexa:
		q = byte(qs)
		if q <= 62 {
			q += 64
		}
	case None:
		return ' '
	}

	return
}

func (qs Qsolexa) String() string {
	if qs < 127 && qs != -128 {
		return string([]byte{byte(qs)})
	} else if qs == -128 {
		return " "
	}
	return "\u221e"
}
