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
	"github.com/kortschak/BioGo/align/pals/dp"
	"github.com/kortschak/BioGo/align/pals/filter"
	"github.com/kortschak/BioGo/bio"
	"github.com/kortschak/BioGo/seq"
	"github.com/kortschak/BioGo/util"
	"reflect"
	"unsafe"
)

const (
	MaxKmerLen         = 15 // Currently limited to 15 due to 32 bit int limit for indexing slices
	DefaultLength      = 400
	DefaultMinIdentity = 0.94
	MaxAvgIndexListLen = 15
	TubeOffsetDelta    = 32
)

var MinWordLength = 4 // For minimum word length, choose k=4 arbitrarily.

var Debug debug

// Return an estimate of the amount of memory required for the filter.
func FilterMemRequired(target *seq.Seq, filterParams filter.Params) uint64 {
	words := util.Pow4(filterParams.WordSize)
	tubeWidth := filterParams.TubeOffset + filterParams.MaxError
	maxActiveTubes := (target.Len()+tubeWidth-1)/filterParams.TubeOffset + 1
	tubes := uint64(maxActiveTubes) * uint64(unsafe.Sizeof(reflect.TypeOf(filter.TubeState{})))
	finger := uint64(unsafe.Sizeof(reflect.TypeOf(uint32(0)))) * uint64(words)
	pos := uint64(unsafe.Sizeof(reflect.TypeOf(int(0)))) * uint64(target.Len())

	return finger + pos + tubes
}

func AvgIndexListLength(target *seq.Seq, filterParams filter.Params) float64 {
	return float64(target.Len()) / float64(int(1) << (uint64(filterParams.WordSize) * 2))
}

// Return an estimate of the total amount of memory required.
func TotalMemRequired(target, query *seq.Seq, filterParams filter.Params) uint64 {
	filter := FilterMemRequired(target, filterParams)
	sequence := target.Len()
	if target != query {
		sequence += query.Len()
	}

	return filter + uint64(sequence)
}

// Return aligner and filter parameters given:
//  minimum hit length.
//  minimum fractional identity.
//  sequence lengths.
//  maximum memory.
//  self alignment?
func OptimiseParameters(minHitLen int, minId float64, target, query *seq.Seq, tubeOffset int, maxMem uint64) (filterParams *filter.Params, dpParams *dp.Params, e error) {
	if minId < 0 || minId > 1.0 {
		e = bio.NewError("bad minId", 0, minId)
		return
	}
	if minHitLen <= MinWordLength {
		e = bio.NewError("bad minHitLength", 0, minHitLen)
		return
	}

	filterParams = &filter.Params{}
	dpParams = &dp.Params{}

	// Lower bound on word length k by requiring manageable index.
	// Given kmer occurs once every 4^k positions.
	// Hence average number of index entries is i = N/(4^k) for random
	// string of length N.
	// Require i <= I, then k > log_4(N/i).
	minWordSize := int(util.Log4(float64(target.Len())) - util.Log4(MaxAvgIndexListLen) + 0.5)

	// First choice is that filter criteria are same as DP criteria,
	// but this may not be possible.
	seedLength := minHitLen
	seedDiffs := int(float64(minHitLen) * (1 - minId))

	// Find filter valid filter parameters,
	// starting from preferred case.
	for {
		minWords := -1
		if MaxKmerLen < minWordSize {
			Debug.Printf("Word size too small: %d < %d\n", MaxKmerLen, minWordSize)
		}
		for wordSize := MaxKmerLen; wordSize >= minWordSize; wordSize-- {
			filterParams.WordSize = wordSize
			filterParams.MinMatch = seedLength
			filterParams.MaxError = seedDiffs
			if tubeOffset > 0 {
				filterParams.TubeOffset = tubeOffset
			} else {
				filterParams.TubeOffset = filterParams.MaxError + TubeOffsetDelta
			}

			mem := TotalMemRequired(target, query, *filterParams)
			if maxMem > 0 && mem > maxMem {
				Debug.Printf("Parameters n=%d k=%d e=%d, mem=%d Mb > maxmem=%d Mb\n",
					filterParams.MinMatch,
					filterParams.WordSize,
					filterParams.MaxError,
					mem/1e6,
					maxMem/1e6)
				minWords = -1
				continue
			}

			minWords = filter.MinWordsPerFilterHit(seedLength, wordSize, seedDiffs)
			if minWords <= 0 {
				Debug.Printf("Parameters n=%d k=%d e=%d, B=%d\n",
					filterParams.MinMatch,
					filterParams.WordSize,
					filterParams.MaxError,
					minWords)
				minWords = -1
				continue
			}

			length := AvgIndexListLength(target, *filterParams)
			if length > MaxAvgIndexListLen {
				Debug.Printf("Parameters n=%d k=%d e=%d, B=%d avgixlen=%d > max = %d\n",
					filterParams.MinMatch,
					filterParams.WordSize,
					filterParams.MaxError,
					minWords,
					length,
					MaxAvgIndexListLen)
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

		return nil, nil, bio.NewError("failed to find filter parameters", 0, nil)
	}

	dpParams.MinHitLength = minHitLen
	dpParams.MinId = minId

	return filterParams, dpParams, nil
}
