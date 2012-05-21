// Package implementing functions required for PALS sequence alignment
package pals

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
	"github.com/kortschak/biogo/align/pals/dp"
	"github.com/kortschak/biogo/align/pals/filter"
	"github.com/kortschak/biogo/bio"
	"github.com/kortschak/biogo/index/kmerindex"
	"github.com/kortschak/biogo/morass"
	"github.com/kortschak/biogo/seq"
	"github.com/kortschak/biogo/util"
	"io"
	"os"
	"unsafe"
)

// Default values for filter and alignment.
var (
	MaxIGap    = 5
	DiffCost   = 3
	SameCost   = 1
	MatchCost  = DiffCost + SameCost
	BlockCost  = DiffCost * MaxIGap
	RMatchCost = float64(DiffCost) + 1
)

// Default thresholds for filter and alignment.
var (
	DefaultLength      = 400
	DefaultMinIdentity = 0.94
	MaxAvgIndexListLen = 15.0
	TubeOffsetDelta    = 32
)

// Default word characteristics.
var (
	MinWordLength = 4  // For minimum word length, choose k=4 arbitrarily.
	MaxKmerLen    = 15 // Currently limited to 15 due to 32 bit int limit for indexing slices
)

// PALS is a type that can perform pairwise alignments of large sequences based on the papers:
//  PILER: identification and classification of genomic repeats.
//   Robert C. Edgar and Eugene W. Myers. Bioinformatics Suppl. 1:i152-i158 (2005)
//  Efficient q-gram filters for finding all 𝛜-matches over a given length.
//   Kim R. Rasmussen, Jens Stoye, and Eugene W. Myers. J. of Computational Biology 13:296–308 (2006).
type PALS struct {
	target, query *seq.Seq
	selfCompare   bool
	index         *kmerindex.Index
	FilterParams  *filter.Params
	DPParams      *dp.Params
	MaxIGap       int
	DiffCost      int
	SameCost      int
	MatchCost     int
	BlockCost     int
	RMatchCost    float64

	log        Logger
	timer      *util.Timer
	tubeOffset int
	maxMem     *uintptr
	hitFilter  *filter.Filter
	morass     *morass.Morass
	err        error
	threads    int
}

// Return a new PALS aligner. Requires
func New(target, query *seq.Seq, selfComp bool, m *morass.Morass, threads, tubeOffset int, mem *uintptr, log Logger) *PALS {
	return &PALS{
		target:      target,
		query:       query,
		selfCompare: selfComp,
		log:         log,
		tubeOffset:  tubeOffset,
		MaxIGap:     MaxIGap,
		DiffCost:    DiffCost,
		SameCost:    SameCost,
		MatchCost:   MatchCost,
		BlockCost:   BlockCost,
		RMatchCost:  RMatchCost,
		maxMem:      mem,
		morass:      m,
		threads:     threads,
	}
}

// Optimise the PALS parameters for given memory, kmer length, hit length and sequence identity.
// An error is returned if no satisfactory parameters can be found.
func (self *PALS) Optimise(minHitLen int, minId float64) (err error) {
	if minId < 0 || minId > 1.0 {
		return bio.NewError("bad minId", 0, minId)
	}
	if minHitLen <= MinWordLength {
		return bio.NewError("bad minHitLength", 0, minHitLen)
	}

	if self.log != nil {
		self.log.Print("Optimising filter parameters")
	}

	filterParams := &filter.Params{}

	// Lower bound on word length k by requiring manageable index.
	// Given kmer occurs once every 4^k positions.
	// Hence average number of index entries is i = N/(4^k) for random
	// string of length N.
	// Require i <= I, then k > log_4(N/i).
	minWordSize := int(util.Log4(float64(self.target.Len())) - util.Log4(MaxAvgIndexListLen) + 0.5)

	// First choice is that filter criteria are same as DP criteria,
	// but this may not be possible.
	seedLength := minHitLen
	seedDiffs := int(float64(minHitLen) * (1 - minId))

	// Find filter valid filter parameters, starting from preferred case.
	for {
		minWords := -1
		if MaxKmerLen < minWordSize {
			if self.log != nil {
				self.log.Printf("Word size too small: %d < %d\n", MaxKmerLen, minWordSize)
			}
		}
		for wordSize := MaxKmerLen; wordSize >= minWordSize; wordSize-- {
			filterParams.WordSize = wordSize
			filterParams.MinMatch = seedLength
			filterParams.MaxError = seedDiffs
			if self.tubeOffset > 0 {
				filterParams.TubeOffset = self.tubeOffset
			} else {
				filterParams.TubeOffset = filterParams.MaxError + TubeOffsetDelta
			}

			mem := self.MemRequired(filterParams)
			if self.maxMem != nil && mem > *self.maxMem {
				if self.log != nil {
					self.log.Printf("Parameters n=%d k=%d e=%d, mem=%d MB > maxmem=%d MB\n",
						filterParams.MinMatch,
						filterParams.WordSize,
						filterParams.MaxError,
						mem/1e6,
						*self.maxMem/1e6)
				}
				minWords = -1
				continue
			}

			minWords = filter.MinWordsPerFilterHit(seedLength, wordSize, seedDiffs)
			if minWords <= 0 {
				if self.log != nil {
					self.log.Printf("Parameters n=%d k=%d e=%d, B=%d\n",
						filterParams.MinMatch,
						filterParams.WordSize,
						filterParams.MaxError,
						minWords)
				}
				minWords = -1
				continue
			}

			length := self.AvgIndexListLength(filterParams)
			if length > MaxAvgIndexListLen {
				if self.log != nil {
					self.log.Printf("Parameters n=%d k=%d e=%d, B=%d avgixlen=%.2f > max = %.2f\n",
						filterParams.MinMatch,
						filterParams.WordSize,
						filterParams.MaxError,
						minWords,
						length,
						MaxAvgIndexListLen)
				}
				minWords = -1
				continue
			}
			break
		}
		if minWords > 0 {
			break
		}

		// Failed to find filter parameters, try
		// fewer errors and shorter seed.
		if seedLength >= minHitLen/4 {
			seedLength /= 2
			continue
		}
		if seedDiffs > 0 {
			seedDiffs--
			continue
		}

		return bio.NewError("failed to find filter parameters", 0)
	}

	self.FilterParams = filterParams

	self.DPParams = &dp.Params{
		MinHitLength: minHitLen,
		MinId:        minId,
	}

	return
}

// Return an estimate of the average number of hits for any given kmer.
func (self *PALS) AvgIndexListLength(filterParams *filter.Params) float64 {
	return float64(self.target.Len()) / float64(int(1)<<(uint(filterParams.WordSize)*2))
}

// Return an estimate of the amount of memory required for the filter.
func (self *PALS) filterMemRequired(filterParams *filter.Params) uintptr {
	words := util.Pow4(filterParams.WordSize)
	tubeWidth := filterParams.TubeOffset + filterParams.MaxError
	maxActiveTubes := (self.target.Len()+tubeWidth-1)/filterParams.TubeOffset + 1
	tubes := uintptr(maxActiveTubes) * unsafe.Sizeof(tubeState{})
	finger := unsafe.Sizeof(uint32(0)) * uintptr(words)
	pos := unsafe.Sizeof(0) * uintptr(self.target.Len())

	return finger + pos + tubes
}

// filter.tubeState is repeated here to allow memory calculation without exporting tubeState from filter package.
type tubeState struct {
	QLo   int
	QHi   int
	Count int
}

// Return an estimate of the total amount of memory required.
func (self *PALS) MemRequired(filterParams *filter.Params) uintptr {
	filter := self.filterMemRequired(filterParams)
	sequence := uintptr(self.target.Len()) + unsafe.Sizeof(self.target)
	if self.target != self.query {
		sequence += uintptr(self.query.Len()) + unsafe.Sizeof(self.query)
	}

	return filter + sequence
}

// Build the kmerindex for filtering.
func (self *PALS) BuildIndex() (err error) {
	self.notify("Indexing")
	index, err := kmerindex.New(self.FilterParams.WordSize, self.target)
	if err != nil {
		return
	} else {
		index.Build()
		self.notify("Indexed")
	}
	self.index = index
	self.hitFilter = filter.New(self.index, self.FilterParams)

	return
}

// Share allows the receiver to use the index and parameters of p.
func (self *PALS) Share(p *PALS) {
	(*self).index = p.index
	(*self).FilterParams = p.FilterParams
	(*self).DPParams = p.DPParams
	self.hitFilter = filter.New(self.index, self.FilterParams)
}

// Perform filtering and alignment for one strand of query.
func (self *PALS) Align(complement bool) (hits dp.DPHits, err error) {
	if self.err != nil {
		return nil, self.err
	}
	var working *seq.Seq
	if complement {
		self.notify("Complementing query")
		working, _ = self.query.RevComp()
		self.notify("Complemented query")
	} else {
		working = self.query
	}

	self.notify("Filtering")
	if err = self.hitFilter.Filter(working, self.selfCompare, complement, self.morass); err != nil {
		return
	}
	self.notifyf("Identified %d filter hits", self.morass.Len())

	self.notify("Merging")
	merger := filter.NewMerger(self.index, working, self.FilterParams, self.MaxIGap, self.selfCompare)
	var hit filter.FilterHit
	for {
		if err = self.morass.Pull(&hit); err != nil {
			break
		}
		merger.MergeFilterHit(&hit)
	}
	if err != nil && err != io.EOF {
		return
	}
	self.err = self.morass.Clear()
	trapezoids := merger.FinaliseMerge()
	lt, lq := trapezoids.Sum()
	self.notifyf("Merged %d trapezoids covering %d x %d", len(trapezoids), lt, lq)

	self.notify("Aligning")
	aligner := dp.NewAligner(self.target, working, self.FilterParams.WordSize, self.DPParams.MinHitLength, self.DPParams.MinId)
	aligner.Config = &dp.AlignConfig{
		MaxIGap:    self.MaxIGap,
		DiffCost:   self.DiffCost,
		SameCost:   self.SameCost,
		MatchCost:  self.MatchCost,
		BlockCost:  self.BlockCost,
		RMatchCost: self.RMatchCost,
	}
	hits = aligner.AlignTraps(trapezoids)
	hitCoverageA, hitCoverageB, err := hits.Sum()
	if err != nil {
		return nil, err
	}
	self.notifyf("Aligned %d hits covering %d x %d", len(hits), hitCoverageA, hitCoverageB)

	return
}

// Remove filesystem components of filter. This should be called after the last use of the aligner.
func (self *PALS) CleanUp() error { return self.morass.CleanUp() }

// Interface for logger used by PALS.
type Logger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})
}

func (self *PALS) notify(n string) {
	if self.log != nil {
		self.log.Print(n)
	}
}

func (self *PALS) notifyf(f string, n ...interface{}) {
	if self.log != nil {
		self.log.Printf(f, n...)
	}
}

func (self *PALS) fatal(n string) {
	if self.log != nil {
		self.log.Fatal(n)
	}
	os.Exit(1)
}

func (self *PALS) fatalf(f string, n ...interface{}) {
	if self.log != nil {
		self.log.Fatalf(f, n...)
	}
	os.Exit(1)
}
