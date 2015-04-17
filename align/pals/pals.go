// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package pals implements functions and methods required for PALS sequence alignment.
package pals

import (
	"github.com/biogo/biogo/align/pals/dp"
	"github.com/biogo/biogo/align/pals/filter"
	"github.com/biogo/biogo/index/kmerindex"
	"github.com/biogo/biogo/morass"
	"github.com/biogo/biogo/seq/linear"
	"github.com/biogo/biogo/util"

	"errors"
	"io"
	"unsafe"
)

// Default thresholds for filter and alignment.
var (
	DefaultLength      = 400
	DefaultMinIdentity = 0.94
	MaxAvgIndexListLen = 15.0
	TubeOffsetDelta    = 32
)

// Default filter and dynamic programming Cost values.
const (
	MaxIGap    = 5
	DiffCost   = 3
	SameCost   = 1
	MatchCost  = DiffCost + SameCost
	BlockCost  = DiffCost * MaxIGap
	RMatchCost = float64(DiffCost) + 1
)

// A dp.Costs based on the default cost values.
var defaultCosts = dp.Costs{
	MaxIGap:    MaxIGap,
	DiffCost:   DiffCost,
	SameCost:   SameCost,
	MatchCost:  MatchCost,
	BlockCost:  BlockCost,
	RMatchCost: RMatchCost,
}

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
	target, query *linear.Seq
	selfCompare   bool
	index         *kmerindex.Index
	FilterParams  *filter.Params
	DPParams      *dp.Params
	dp.Costs

	log        Logger
	timer      *util.Timer
	tubeOffset int
	maxMem     *uintptr
	hitFilter  *filter.Filter
	morass     *morass.Morass
	trapezoids filter.Trapezoids
	err        error
	threads    int
}

// Return a new PALS aligner. Requires
func New(target, query *linear.Seq, selfComp bool, m *morass.Morass, threads, tubeOffset int, mem *uintptr, log Logger) *PALS {
	return &PALS{
		target:      target,
		query:       query,
		selfCompare: selfComp,
		log:         log,
		tubeOffset:  tubeOffset,
		Costs:       defaultCosts,
		maxMem:      mem,
		morass:      m,
		threads:     threads,
	}
}

// Optimise the PALS parameters for given memory, kmer length, hit length and sequence identity.
// An error is returned if no satisfactory parameters can be found.
func (p *PALS) Optimise(minHitLen int, minId float64) error {
	if minId < 0 || minId > 1.0 {
		return errors.New("pals: minimum identity out of range")
	}
	if minHitLen <= MinWordLength {
		return errors.New("pals: minimum hit length too short")
	}

	if p.log != nil {
		p.log.Print("Optimising filter parameters")
	}

	filterParams := &filter.Params{}

	// Lower bound on word length k by requiring manageable index.
	// Given kmer occurs once every 4^k positions.
	// Hence average number of index entries is i = N/(4^k) for random
	// string of length N.
	// Require i <= I, then k > log_4(N/i).
	minWordSize := int(util.Log4(float64(p.target.Len())) - util.Log4(MaxAvgIndexListLen) + 0.5)

	// First choice is that filter criteria are same as DP criteria,
	// but this may not be possible.
	seedLength := minHitLen
	seedDiffs := int(float64(minHitLen) * (1 - minId))

	// Find filter valid filter parameters, starting from preferred case.
	for {
		minWords := -1
		if MaxKmerLen < minWordSize {
			if p.log != nil {
				p.log.Printf("Word size too small: %d < %d\n", MaxKmerLen, minWordSize)
			}
		}
		for wordSize := MaxKmerLen; wordSize >= minWordSize; wordSize-- {
			filterParams.WordSize = wordSize
			filterParams.MinMatch = seedLength
			filterParams.MaxError = seedDiffs
			if p.tubeOffset > 0 {
				filterParams.TubeOffset = p.tubeOffset
			} else {
				filterParams.TubeOffset = filterParams.MaxError + TubeOffsetDelta
			}

			mem := p.MemRequired(filterParams)
			if p.maxMem != nil && mem > *p.maxMem {
				if p.log != nil {
					p.log.Printf("Parameters n=%d k=%d e=%d, mem=%d MB > maxmem=%d MB\n",
						filterParams.MinMatch,
						filterParams.WordSize,
						filterParams.MaxError,
						mem/1e6,
						*p.maxMem/1e6)
				}
				minWords = -1
				continue
			}

			minWords = filter.MinWordsPerFilterHit(seedLength, wordSize, seedDiffs)
			if minWords <= 0 {
				if p.log != nil {
					p.log.Printf("Parameters n=%d k=%d e=%d, B=%d\n",
						filterParams.MinMatch,
						filterParams.WordSize,
						filterParams.MaxError,
						minWords)
				}
				minWords = -1
				continue
			}

			length := p.AvgIndexListLength(filterParams)
			if length > MaxAvgIndexListLen {
				if p.log != nil {
					p.log.Printf("Parameters n=%d k=%d e=%d, B=%d avgixlen=%.2f > max = %.2f\n",
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

		return errors.New("pals: failed to find filter parameters")
	}

	p.FilterParams = filterParams

	p.DPParams = &dp.Params{
		MinHitLength: minHitLen,
		MinId:        minId,
	}

	return nil
}

// Return an estimate of the average number of hits for any given kmer.
func (p *PALS) AvgIndexListLength(filterParams *filter.Params) float64 {
	return float64(p.target.Len()) / float64(int(1)<<(uint(filterParams.WordSize)*2))
}

// Return an estimate of the amount of memory required for the filter.
func (p *PALS) filterMemRequired(filterParams *filter.Params) uintptr {
	words := util.Pow4(filterParams.WordSize)
	tubeWidth := filterParams.TubeOffset + filterParams.MaxError
	maxActiveTubes := (p.target.Len()+tubeWidth-1)/filterParams.TubeOffset + 1
	tubes := uintptr(maxActiveTubes) * unsafe.Sizeof(tubeState{})
	finger := unsafe.Sizeof(uint32(0)) * uintptr(words)
	pos := unsafe.Sizeof(0) * uintptr(p.target.Len())

	return finger + pos + tubes
}

// filter.tubeState is repeated here to allow memory calculation without
// exporting tubeState from filter package.
type tubeState struct {
	QLo   int
	QHi   int
	Count int
}

// Return an estimate of the total amount of memory required.
func (p *PALS) MemRequired(filterParams *filter.Params) uintptr {
	filter := p.filterMemRequired(filterParams)
	sequence := uintptr(p.target.Len()) + unsafe.Sizeof(p.target)
	if p.target != p.query {
		sequence += uintptr(p.query.Len()) + unsafe.Sizeof(p.query)
	}

	return filter + sequence
}

// Build the kmerindex for filtering.
func (p *PALS) BuildIndex() error {
	p.notify("Indexing")
	ki, err := kmerindex.New(p.FilterParams.WordSize, p.target)
	if err != nil {
		return err
	} else {
		ki.Build()
		p.notify("Indexed")
	}
	p.index = ki
	p.hitFilter = filter.New(p.index, p.FilterParams)

	return nil
}

// Share allows the receiver to use the index and parameters of m.
func (p *PALS) Share(m *PALS) {
	p.index = m.index
	p.FilterParams = m.FilterParams
	p.DPParams = m.DPParams
	p.hitFilter = filter.New(p.index, p.FilterParams)
}

// Align performs filtering and alignment for one strand of query.
func (p *PALS) Align(complement bool) (dp.Hits, error) {
	if p.err != nil {
		return nil, p.err
	}
	var (
		working *linear.Seq
		err     error
	)
	if complement {
		p.notify("Complementing query")
		working = p.query.Clone().(*linear.Seq)
		working.RevComp()
		p.notify("Complemented query")
	} else {
		working = p.query
	}

	p.notify("Filtering")
	err = p.hitFilter.Filter(working, p.selfCompare, complement, p.morass)
	if err != nil {
		return nil, err
	}
	p.notifyf("Identified %d filter hits", p.morass.Len())

	p.notify("Merging")
	merger := filter.NewMerger(p.index, working, p.FilterParams, p.MaxIGap, p.selfCompare)
	var h filter.Hit
	for {
		if err = p.morass.Pull(&h); err != nil {
			break
		}
		merger.MergeFilterHit(&h)
	}
	if err != nil && err != io.EOF {
		return nil, err
	}
	p.err = p.morass.Clear()
	p.trapezoids = merger.FinaliseMerge()
	lt, lq := p.trapezoids.Sum()
	p.notifyf("Merged %d trapezoids covering %d x %d", len(p.trapezoids), lt, lq)

	p.notify("Aligning")
	aligner := dp.NewAligner(
		p.target, working,
		p.FilterParams.WordSize, p.DPParams.MinHitLength, p.DPParams.MinId,
	)
	aligner.Costs = &p.Costs
	hits := aligner.AlignTraps(p.trapezoids)
	hitCoverageA, hitCoverageB, err := hits.Sum()
	if err != nil {
		return nil, err
	}
	p.notifyf("Aligned %d hits covering %d x %d", len(hits), hitCoverageA, hitCoverageB)

	return hits, nil
}

// Trapezoids returns the filter trapezoids identified during a call to Align.
func (p *PALS) Trapezoids() filter.Trapezoids { return p.trapezoids }

// AlignFrom performs filtering and alignment for one strand of query using the
// provided filter trapezoids as seeds.
func (p *PALS) AlignFrom(traps filter.Trapezoids, complement bool) (dp.Hits, error) {
	if p.err != nil {
		return nil, p.err
	}
	var (
		working *linear.Seq
		err     error
	)
	if complement {
		p.notify("Complementing query")
		working = p.query.Clone().(*linear.Seq)
		working.RevComp()
		p.notify("Complemented query")
	} else {
		working = p.query
	}

	p.notify("Aligning")
	aligner := dp.NewAligner(
		p.target, working,
		p.FilterParams.WordSize, p.DPParams.MinHitLength, p.DPParams.MinId,
	)
	aligner.Costs = &p.Costs
	hits := aligner.AlignTraps(traps)
	hitCoverageA, hitCoverageB, err := hits.Sum()
	if err != nil {
		return nil, err
	}
	p.notifyf("Aligned %d hits covering %d x %d", len(hits), hitCoverageA, hitCoverageB)

	return hits, nil
}

// Remove file system components of filter. This should be called after
// the last use of the aligner.
func (p *PALS) CleanUp() error { return p.morass.CleanUp() }

// Interface for logger used by PALS.
type Logger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
}

func (p *PALS) notify(n string) {
	if p.log != nil {
		p.log.Print(n)
	}
}

func (p *PALS) notifyf(f string, n ...interface{}) {
	if p.log != nil {
		p.log.Printf(f, n...)
	}
}
