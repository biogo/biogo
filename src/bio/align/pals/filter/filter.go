// Package providing PALS sequence hit filter routines
package filter
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
	"github.com/kortschak/BioGo/bio"
	"github.com/kortschak/BioGo/bio/index/kmerindex"
	"github.com/kortschak/BioGo/bio/morass"
	"github.com/kortschak/BioGo/bio/seq"
)

type Filter struct {
	target         *seq.Seq
	index          *kmerindex.Index
	tubes          []TubeState
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

func New(index *kmerindex.Index, params Params) (f *Filter) {
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

func (self *Filter) Filter(query *seq.Seq, selfAlign, complement bool, morass *morass.Morass) (err error) {
	self.selfAlign = selfAlign
	self.complement = complement
	self.morass = morass
	self.k = self.index.GetK()

	// Ukonnen's Lemma
	self.minKmersPerHit = MinWordsPerFilterHit(self.minMatch, self.k, self.maxError)

	// Maximum distance between SeqQ positions of two k-mers in a match
	// (More stringent bounds may be possible, but not a big problem
	// if two adjacent matches get merged).
	self.maxKmerDist = self.minMatch - self.k

	tubeWidth := self.tubeOffset + self.maxError

	if self.tubeOffset < self.maxError {
		return bio.NewError("TubeOffset < MaxError", 0, []int{self.tubeOffset, self.maxError})
	}

	maxActiveTubes := (self.target.Len()+tubeWidth-1)/self.tubeOffset + 1
	self.tubes = make([]TubeState, maxActiveTubes)

	// Ticker tracks cycling of circular list of active tubes.
	ticker := tubeWidth

	f := func(index *kmerindex.Index, position, kmer int) {
		from := 0
		if kmer > 0 {
			from = index.FingerAt(kmer - 1)
		}
		to := index.FingerAt(kmer)
		for i := from; i < to; i++ {
			self.commonKmer(index.PosAt(i), position)
		}

		if ticker--; ticker == 0 {
			if e := self.tubeEnd(position); e != nil {
				panic(e) // Caught by fastkmerindex.ForEachKmerOf and returned
			}
			ticker = self.tubeOffset
		}
	}

	if err = self.index.ForEachKmerOf(query, 0, query.Len(), f); err != nil {
		return
	}

	if err = self.tubeEnd(query.Len() - 1); err != nil {
		return
	}

	diagFrom := self.diagIndex(self.target.Len()-1, query.Len()-1) - tubeWidth
	diagTo := self.diagIndex(0, query.Len()-1) + tubeWidth

	tubeFrom := self.tubeIndex(diagFrom)
	if tubeFrom < 0 {
		tubeFrom = 0
	}

	tubeTo := self.tubeIndex(diagTo)

	for tubeIndex := tubeFrom; tubeIndex <= tubeTo; tubeIndex++ {
		if err = self.tubeFlush(tubeIndex); err != nil {
			return
		}
	}

	self.tubes = nil

	return self.morass.Finalise()
}

// Called when q=Qlen - 1 to flush any hits in each tube.
func (self *Filter) tubeFlush(tubeIndex int) (e error) {
	tube := &self.tubes[tubeIndex%cap(self.tubes)]

	if tube.Count < self.minKmersPerHit {
		return
	}

	if e = self.addHit(tubeIndex, tube.QLo, tube.QHi); e != nil {
		return
	}
	tube.Count = 0

	return
}

func (self *Filter) diagIndex(t, q int) int {
	return self.target.Len() - t + q
}

func (self *Filter) tubeIndex(d int) int {
	return d / self.tubeOffset
}

// Found a common k-mer SeqT[t] and SeqQ[q].
func (self *Filter) commonKmer(t, q int) (e error) {
	if self.selfAlign && ((self.complement && (q < self.target.Len()-t)) || (!self.complement && (q <= t))) {
		return
	}

	diagIndex := self.diagIndex(t, q)
	tubeIndex := self.tubeIndex(diagIndex)

	if e = self.hitTube(tubeIndex, q); e != nil {
		return
	}

	// Hit in overlapping tube preceding this one?
	if diagIndex%self.tubeOffset < self.maxError {
		if tubeIndex == 0 {
			tubeIndex = cap(self.tubes) - 1
		} else {
			tubeIndex--
		}
		e = self.hitTube(tubeIndex, q)
	}

	return
}

func (self *Filter) hitTube(tubeIndex, q int) (e error) {
	tube := &self.tubes[tubeIndex%cap(self.tubes)]

	if tube.Count == 0 {
		tube.Count = 1
		tube.QLo = q
		tube.QHi = q
		return
	}

	if q-tube.QHi > self.maxKmerDist {
		if tube.Count >= self.minKmersPerHit {
			if e = self.addHit(tubeIndex, tube.QLo, tube.QHi); e != nil {
				return
			}
		}

		tube.Count = 1
		tube.QLo = q
		tube.QHi = q
		return
	}

	tube.Count++
	tube.QHi = q

	return
}

// Called when end of a tube is reached
// A point in the tube -- the point with maximal q -- is (Tlen-1,q-1).
func (self *Filter) tubeEnd(q int) (e error) {
	diagIndex := self.diagIndex(self.target.Len()-1, q-1)
	tubeIndex := self.tubeIndex(diagIndex)
	tube := &self.tubes[tubeIndex%cap(self.tubes)]

	if tube.Count >= self.minKmersPerHit {
		if e = self.addHit(tubeIndex, tube.QLo, tube.QHi); e != nil {
			return
		}
	}

	tube.Count = 0

	return
}

func (self *Filter) addHit(tubeIndex, QLo, QHi int) (e error) {
	fh := FilterHit{
		QFrom:     QLo,
		QTo:       QHi + self.k,
		DiagIndex: self.target.Len() - tubeIndex*self.tubeOffset,
	}

	return self.morass.Push(fh)
}
