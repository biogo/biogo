// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package complex provides routines for evaluating sequence complexity.
package complexity

import (
	"code.google.com/p/biogo/seq"

	"compress/zlib"
	"fmt"
	"math"
)

const tableLength = 10000

var lnFacTable = genLnFac(tableLength)

func genLnFac(l int) (table []float64) {
	table = make([]float64, l)
	lnfac := 0.

	for i := 1; i < l; i++ {
		lnfac += math.Log(float64(i))
		table[i] = lnfac
	}

	return
}

const ln2pi = 1.8378770664093454835606594728112352797227949472755668

func lnFac(x int) float64 {
	if x < len(lnFacTable) {
		return lnFacTable[x]
	}
	// use Sterling's approximation for queries outside the table:
	return (float64(x)+0.5)*math.Log(float64(x)) - float64(x) + ln2pi/2
}

func logBaseK(logk, x float64) float64 {
	return math.Log(x) / logk
}

// EntropicComplexity returns the entropic complexity of a segment of s defined by
// start and end.
func EntropicComplexity(s seq.Sequence, start, end int) (ce float64, err error) {
	if start < s.Start() || end > s.End() {
		err = fmt.Errorf("complex: index out of range")
		return
	}
	if start == end {
		return 0, nil
	}

	N := float64(end - start)
	k := s.Alphabet().Len()
	logk := math.Log(float64(k))
	n := make([]float64, k)

	// tally classes
	it := s.Alphabet().LetterIndex()
	for i := start; i < end; i++ {
		if ind := it[s.At(i).L]; ind >= 0 {
			n[ind]++
		}
	}

	// -∑i=1..k((n_i/N)*log_k(n_i/N))
	for i := 0; i < k; i++ {
		if n[i] != 0 { // ignore zero counts
			ce += n[i] * logBaseK(logk, n[i]/N)
		}
	}
	ce = -ce / N

	return
}

// WFComplexity returns the Wootton and Federhen complexity of a segment of s defined by
// start and end.
func WFComplexity(s seq.Sequence, start, end int) (cwf float64, err error) {
	if start < s.Start() || end > s.End() {
		err = fmt.Errorf("complex: index out of range")
		return
	}
	if start == end {
		return 0, nil
	}

	N := float64(end - start)
	k := s.Alphabet().Len()
	logk := math.Log(float64(k))
	n := make([]int, k)

	// tally classes
	it := s.Alphabet().LetterIndex()
	for i := start; i < end; i++ {
		if ind := it[s.At(i).L]; ind >= 0 {
			n[ind]++
		}
	}

	// 1/N*log_k(N!/∏i=1..k(n_i!))
	cwf = lnFac(end - start)
	for i := 0; i < k; i++ {
		cwf -= lnFac(n[i])
	}
	cwf /= N * logk

	return
}

type byteCounter int

func (b *byteCounter) Write(p []byte) (n int, err error) {
	*b += byteCounter(len(p))
	return len(p), nil
}

var overhead = calcOverhead()

func calcOverhead() byteCounter {
	b := new(byteCounter)
	z := zlib.NewWriter(b)
	z.Write([]byte{0})
	z.Close()

	return *b - 1
}

// ZComplexity returns the zlib compression estimate of complexity of a segment of s defined by
// start and end.
func ZComplexity(s seq.Sequence, start, end int) (cz float64, err error) {
	if start < s.Start() || end > s.End() {
		err = fmt.Errorf("complex: index out of range")
		return
	}
	if start == end {
		return 0, nil
	}

	bc := new(byteCounter)
	z := zlib.NewWriter(bc)
	defer z.Close()
	it := s.Alphabet().LetterIndex()
	for i := start; i < end; i++ {
		if b := byte(s.At(i).L); it[b] >= 0 {
			z.Write([]byte{b})
		}
	}
	z.Close()

	cz = (float64(*bc - overhead)) / float64(end-start)

	return
}
