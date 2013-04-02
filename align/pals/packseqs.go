// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pals

import (
	"code.google.com/p/biogo/alphabet"
	"code.google.com/p/biogo/feat"
	"code.google.com/p/biogo/seq"
	"code.google.com/p/biogo/seq/linear"
	"code.google.com/p/biogo/util"
	"errors"
	"fmt"
)

var (
	binSize    int = 1 << 10
	minPadding int = 50
)

// A Packer collects a set of sequence into a Packed sequence.
type Packer struct {
	packed  *Packed
	lastPad int
	length  int
}

type Packed struct {
	*linear.Seq
	seqMap
}

// Convert coordinates in a packed sequence into a feat.Feature.
func (pa *Packed) feature(from, to int, comp bool) (feat.Feature, error) {
	if comp {
		from, to = pa.Len()-to, pa.Len()-from
	}
	if from >= to {
		return nil, errors.New("pals: from > to")
	}

	// DPHit coordinates sometimes over/underflow.
	// This is a lazy hack to work around it, should really figure
	// out what is going on.
	if from < 0 {
		from = 0
	}
	if to > pa.Len() {
		to = pa.Len()
	}

	// Take midpoint of segment -- lazy hack again, endpoints
	// sometimes under / overflow
	bin := (from + to) / (2 * binSize)
	binCount := (pa.Len() + binSize - 1) / binSize

	if bin < 0 || bin >= binCount {
		return nil, fmt.Errorf("pals: bin %d out of range 0..%d", bin, binCount-1)
	}

	contigIndex := pa.seqMap.binMap[bin]

	if contigIndex < 0 || contigIndex >= len(pa.seqMap.contigs) {
		return nil, fmt.Errorf("pals: contig %s index %d out of range 0..%d", pa.ID, contigIndex, len(pa.seqMap.contigs))
	}

	length := to - from

	if length < 0 {
		return nil, errors.New("pals: length < 0")
	}

	contig := pa.seqMap.contigs[contigIndex]
	contigFrom := from - contig.from
	contigTo := contigFrom + length

	if contigFrom < 0 {
		contigFrom = 0
	}

	if contigTo > contig.Len() {
		contigTo = contig.Len()
	}

	return &Feature{
		ID:   contig.ID,
		From: contigFrom,
		To:   contigTo,
		Loc:  Contig(contig.ID),
	}, nil
}

// Create a new Packer.
func NewPacker(id string) *Packer {
	return &Packer{
		packed: &Packed{
			Seq:    &linear.Seq{Annotation: seq.Annotation{ID: id}},
			seqMap: seqMap{},
		},
	}
}

// Pack a sequence into the Packed sequence. Returns a string giving diagnostic information.
func (pa *Packer) Pack(seq *linear.Seq) (string, error) {
	if pa.packed.Alpha == nil {
		pa.packed.Alpha = seq.Alpha
	} else if pa.packed.Alpha != seq.Alpha {
		return "", errors.New("pals: alphabet mismatch")
	}

	c := contig{Seq: seq}

	padding := binSize - seq.Len()%binSize
	if padding < minPadding {
		padding += binSize
	}

	pa.length += pa.lastPad
	c.from = pa.length
	pa.length += seq.Len()
	pa.lastPad = padding

	m := &pa.packed.seqMap
	bins := make([]int, (padding+seq.Len())/binSize)
	for i := 0; i < len(bins); i++ {
		bins[i] = len(m.contigs)
	}
	m.binMap = append(m.binMap, bins...)
	m.contigs = append(m.contigs, c)

	return fmt.Sprintf("%20s\t%10d\t%7d-%-d", seq.ID[:util.Min(20, len(seq.ID))], seq.Len(), len(m.binMap)-len(bins), len(m.binMap)-1), nil
}

// Finalise the sequence packing.
func (pa *Packer) FinalisePack() *Packed {
	lastPad := 0
	seq := make(alphabet.Letters, 0, pa.length)
	for _, c := range pa.packed.seqMap.contigs {
		padding := binSize - c.Len()%binSize
		if padding < minPadding {
			padding += binSize
		}
		seq = append(seq, alphabet.Letter('N').Repeat(lastPad)...)
		seq = append(seq, c.Seq.Seq...)
		lastPad = padding
	}
	pa.packed.Seq.Seq = seq

	return pa.packed
}

// A contig holds a sequence within a SeqMap.
type contig struct {
	*linear.Seq
	from int
}

// A seqMap is a collection of sequences mapped to a Packed sequence.
type seqMap struct {
	contigs []contig
	binMap  []int
}
