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
	"fmt"
	"github.com/kortschak/BioGo/bio"
	"github.com/kortschak/BioGo/bio/align/pals/dp"
	"github.com/kortschak/BioGo/bio/align/pals/filter"
	"github.com/kortschak/BioGo/bio/seq"
	"github.com/kortschak/BioGo/bio/util"
	"reflect"
	"unsafe"
)

/***
Generate default aligner parameters given:
	minimum hit length.
	minimum fractional identity.
	sequence lengths.
	maximum memory.
	self alignment?
***/

const (
	DefaultLength      = 400
	DefaultMinIdentity = 0.94
	MinWordLength      = 4
	MaxAvgIndexListLen = 15
	TubeOffsetDelta    = 32
)

var MaxKmerLen = 15
var Debug bool

/***
For minimum word length, choose k=4 arbitrarily.
For max, k=16 definitely won't work with 32-bit size_t
because 4^16 = 2^32 = 4G >> 4GB when considering words
are stored in ints
k=15 might be OK, but would have to look carefully at
boundary cases, which I haven't done.
k=14 is definitely safe, so set this as upper bound.
***/

func FilterMemRequired(target *seq.Seq, filterParams filter.Params) uint64 {
	words := util.Pow4(filterParams.WordSize)
	tubeWidth := filterParams.TubeOffset + filterParams.MaxError
	maxActiveTubes := (target.Len()+tubeWidth-1)/filterParams.TubeOffset + 1
	tubes := uint64(maxActiveTubes) * uint64(unsafe.Sizeof(reflect.TypeOf(filter.TubeState{})))
	finger := uint64(unsafe.Sizeof(reflect.TypeOf(uint32(0)))) * uint64(words)
	pos := uint64(unsafe.Sizeof(reflect.TypeOf(int(0)))) * uint64(target.Len())
	return finger + pos + tubes
}

func AvgIndexListLength(target *seq.Seq, filterParams filter.Params) uint64 {
	return uint64(target.Len()) / (1 << (uint64(filterParams.WordSize) * 2))
}

func TotalMemRequired(target, query *seq.Seq, filterParams filter.Params) uint64 {
	filter := FilterMemRequired(target, filterParams)
	sequence := target.Len()
	if target != query {
		sequence += query.Len()
	}
	return filter + uint64(sequence)
}

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
		if MaxKmerLen < minWordSize && Debug {
			fmt.Printf("Word size too small: %d < %d\n", MaxKmerLen, minWordSize)
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
				if Debug {
					fmt.Printf("Parameters n=%d k=%d e=%d, mem=%d Mb > maxmem=%d Mb\n",
						filterParams.MinMatch,
						filterParams.WordSize,
						filterParams.MaxError,
						mem/1e6,
						maxMem/1e6)
				}
				minWords = -1
				continue
			}

			minWords = filter.MinWordsPerFilterHit(seedLength, wordSize, seedDiffs)
			if minWords <= 0 {
				if Debug {
					fmt.Printf("Parameters n=%d k=%d e=%d, B=%d\n",
						filterParams.MinMatch,
						filterParams.WordSize,
						filterParams.MaxError,
						minWords)
				}
				minWords = -1
				continue
			}

			length := AvgIndexListLength(target, *filterParams)
			if length > MaxAvgIndexListLen {
				if Debug {
					fmt.Printf("Parameters n=%d k=%d e=%d, B=%d avgixlen=%d > max = %d\n",
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

		return nil, nil, bio.NewError("failed to find filter parameters", 0, nil)
	}

	dpParams.MinHitLength = minHitLen
	dpParams.MinId = minId

	return filterParams, dpParams, nil
}

// What follows is fluff that goes into the application itself

/***
Alignment parameters can be specified in three ways:

(1) All defaults.
(2) By -length and -pctid only.
(3) All parameters specified:
	-wordsize -seedlength -seeddiffs -length -pctid

Optional parameters:
	-maxmem			(Default = 80% of RAM, 0 = no maximum)
	-tube			(Tube offset, Default = 32)
***
void GetParams(int SeqLengthT, int SeqLengthQ, bool Self,
               filter.FilterParams *ptrFP, DPParams *ptrDP)
{
	const char *strLength = ValueOpt("length");
	const char *strPctId = ValueOpt("pctid");
	const char *strWordSize = ValueOpt("wordsize");
	const char *strSeedLength = ValueOpt("seedlength");
	const char *strSeedErrors = ValueOpt("seeddiffs");

	const char *strMaxMem = ValueOpt("maxmem");
	const char *strTubeOffset = ValueOpt("tubeoffset");

	const size_t MaxMem = (0 == strMaxMem) ? GetRAMSize()*RAM_FRACT : strtoul(strMaxMem, NULL, 0);
	const int TubeOffset = (0 == strTubeOffset) ? -1 : atoi(strTubeOffset);

// All parameters specified
	if (0 != strWordSize || 0 != strSeedLength || 0 != strSeedErrors) {
		if (0 == strWordSize || 0 == strSeedLength || 0 == strSeedErrors ||
		        0 == strLength || 0 == strPctId) {
			Quit("Missing one or more of: -wordsize, -seedlength, -seeddiffs, -length, -pctid");
		}

		ptrFP->MaxError = atoi(strSeedErrors);
		ptrFP->MinMatch = atoi(strSeedLength);
		ptrFP->WordSize = atoi(strWordSize);
		ptrFP->TubeOffset = TubeOffset > 0 ? TubeOffset : ptrFP->MaxError + TUBE_OFFSET_DELTA;

		ptrDP->MinHitLength = atoi(strLength);
		ptrDP->MinId = atoi(strPctId)/100.0;
	}

// -length and -pctid
	else if (0 != strLength || 0 != strPctId) {
		if (0 == strLength || 0 == strPctId) {
			Quit("Missing option -length or -pctid");
		}

		int Length = atoi(strLength);
		double MinId = atof(strPctId)/100.0;

		DefParams(Length, MinId, SeqLengthT, SeqLengthQ, Self, MaxMem, ptrFP, ptrDP);
	}

// All defaults
	else {
		DefParams(DEFAULT_LENGTH, DEFAULT_MIN_ID, SeqLengthT, SeqLengthQ,
		          Self, MaxMem, ptrFP, ptrDP);
	}
}
*/
