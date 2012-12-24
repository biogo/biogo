// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package providing PALS sequence hit filter routines based on 
// 'Efficient q-gram filters for finding all 𝛜-matches over a given length.'
//   Kim R. Rasmussen, Jens Stoye, and Eugene W. Myers. J. of Computational Biology 13:296–308 (2006).
package filter

import (
	"code.google.com/p/biogo/bio"
	"code.google.com/p/biogo/exp/seq/linear"
	"code.google.com/p/biogo/index/kmerindex"
	"code.google.com/p/biogo/morass"
)

// Ukonnen's Lemma: U(n, q, 𝛜) := (n + 1) - q(⌊𝛜n⌋ + 1)
func MinWordsPerFilterHit(hitLength, wordLength, maxErrors int) int {
	return hitLength + 1 - wordLength*(maxErrors+1)
}

// Type for passing filter parameters.
type Params struct {
	WordSize   int
	MinMatch   int
	MaxError   int
	TubeOffset int
}

// Filter implements a q-gram filter similar to that described in Rassmussen 2005.
// This implementation is a translation of the C++ code written by Edgar and Myers.
type Filter struct {
	target         *linear.Seq
	index          *kmerindex.Index
	tubes          []tubeState
	morass         *morass.Morass
	k              int
	minMatch       int
	maxError       int
	maxKmerDist    int
	minKmersPerHit int
	tubeOffset     int
	selfAlign      bool
	complement     bool
}

// Return a new Filter using index as the target and filter parameters in params.
func New(index *kmerindex.Index, params *Params) (f *Filter) {
	f = &Filter{
		index:      index,
		target:     index.Seq,
		k:          index.GetK(),
		minMatch:   params.MinMatch,
		maxError:   params.MaxError,
		tubeOffset: params.TubeOffset,
	}

	return
}

// Filter a query sequence against the stored index. If query and the target are the same sequence,
// selfAlign can be used to avoid double seaching - behavior is undefined if the the sequences are not the same.
// A morass is used to store and sort individual filter hits.
func (f *Filter) Filter(query *linear.Seq, selfAlign, complement bool, morass *morass.Morass) error {
	f.selfAlign = selfAlign
	f.complement = complement
	f.morass = morass
	f.k = f.index.GetK()

	// Ukonnen's Lemma
	f.minKmersPerHit = MinWordsPerFilterHit(f.minMatch, f.k, f.maxError)

	// Maximum distance between SeqQ positions of two k-mers in a match
	// (More stringent bounds may be possible, but not a big problem
	// if two adjacent matches get merged).
	f.maxKmerDist = f.minMatch - f.k

	tubeWidth := f.tubeOffset + f.maxError

	if f.tubeOffset < f.maxError {
		return bio.NewError("TubeOffset < MaxError", 0, []int{f.tubeOffset, f.maxError})
	}

	maxActiveTubes := (f.target.Len()+tubeWidth-1)/f.tubeOffset + 1
	f.tubes = make([]tubeState, maxActiveTubes)

	// Ticker tracks cycling of circular list of active tubes.
	ticker := tubeWidth

	var err error
	err = f.index.ForEachKmerOf(query, 0, query.Len(), func(index *kmerindex.Index, position, kmer int) {
		from := 0
		if kmer > 0 {
			from = index.FingerAt(kmer - 1)
		}
		to := index.FingerAt(kmer)
		for i := from; i < to; i++ {
			f.commonKmer(index.PosAt(i), position)
		}

		if ticker--; ticker == 0 {
			if e := f.tubeEnd(position); e != nil {
				panic(e) // Caught by fastkmerindex.ForEachKmerOf and returned
			}
			ticker = f.tubeOffset
		}
	})
	if err != nil {
		return err
	}

	err = f.tubeEnd(query.Len() - 1)
	if err != nil {
		return err
	}

	diagFrom := f.diagIndex(f.target.Len()-1, query.Len()-1) - tubeWidth
	diagTo := f.diagIndex(0, query.Len()-1) + tubeWidth

	tubeFrom := f.tubeIndex(diagFrom)
	if tubeFrom < 0 {
		tubeFrom = 0
	}

	tubeTo := f.tubeIndex(diagTo)

	for tubeIndex := tubeFrom; tubeIndex <= tubeTo; tubeIndex++ {
		err = f.tubeFlush(tubeIndex)
		if err != nil {
			return err
		}
	}

	f.tubes = nil

	return f.morass.Finalise()
}

// A tubeState stores active filter bin states.
// tubeState is repeated in the pals package to allow memory calculation without exporting tubeState from filter package.
type tubeState struct {
	QLo   int
	QHi   int
	Count int
}

// Called when q=Qlen - 1 to flush any hits in each tube.
func (f *Filter) tubeFlush(tubeIndex int) error {
	tube := &f.tubes[tubeIndex%cap(f.tubes)]

	if tube.Count < f.minKmersPerHit {
		return nil
	}

	err := f.addHit(tubeIndex, tube.QLo, tube.QHi)
	if err != nil {
		return err
	}
	tube.Count = 0

	return nil
}

func (f *Filter) diagIndex(t, q int) int {
	return f.target.Len() - t + q
}

func (f *Filter) tubeIndex(d int) int {
	return d / f.tubeOffset
}

// Found a common k-mer SeqT[t] and SeqQ[q].
func (f *Filter) commonKmer(t, q int) error {
	if f.selfAlign && ((f.complement && (q < f.target.Len()-t)) || (!f.complement && (q <= t))) {
		return nil
	}

	diagIndex := f.diagIndex(t, q)
	tubeIndex := f.tubeIndex(diagIndex)

	err := f.hitTube(tubeIndex, q)
	if err != nil {
		return err
	}

	// Hit in overlapping tube preceding this one?
	if diagIndex%f.tubeOffset < f.maxError {
		if tubeIndex == 0 {
			tubeIndex = cap(f.tubes) - 1
		} else {
			tubeIndex--
		}
		err = f.hitTube(tubeIndex, q)
	}

	return err
}

func (f *Filter) hitTube(tubeIndex, q int) error {
	tube := &f.tubes[tubeIndex%cap(f.tubes)]

	if tube.Count == 0 {
		tube.Count = 1
		tube.QLo = q
		tube.QHi = q
		return nil
	}

	if q-tube.QHi > f.maxKmerDist {
		if tube.Count >= f.minKmersPerHit {
			err := f.addHit(tubeIndex, tube.QLo, tube.QHi)
			if err != nil {
				return err
			}
		}

		tube.Count = 1
		tube.QLo = q
		tube.QHi = q
		return nil
	}

	tube.Count++
	tube.QHi = q

	return nil
}

// Called when end of a tube is reached
// A point in the tube -- the point with maximal q -- is (Tlen-1,q-1).
func (f *Filter) tubeEnd(q int) error {
	diagIndex := f.diagIndex(f.target.Len()-1, q-1)
	tubeIndex := f.tubeIndex(diagIndex)
	tube := &f.tubes[tubeIndex%cap(f.tubes)]

	if tube.Count >= f.minKmersPerHit {
		err := f.addHit(tubeIndex, tube.QLo, tube.QHi)
		if err != nil {
			return err
		}
	}

	tube.Count = 0

	return nil
}

func (f *Filter) addHit(tubeIndex, QLo, QHi int) error {
	fh := FilterHit{
		QFrom:     QLo,
		QTo:       QHi + f.k,
		DiagIndex: f.target.Len() - tubeIndex*f.tubeOffset,
	}

	return f.morass.Push(fh)
}
