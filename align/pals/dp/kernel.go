// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dp

import (
	"github.com/biogo/biogo/align/pals/filter"
	"github.com/biogo/biogo/alphabet"
	"github.com/biogo/biogo/seq/linear"
)

// A kernel handles the actual dp alignment process.
type kernel struct {
	target, query *linear.Seq

	minLen  int
	maxDiff float64

	valueToCode alphabet.Index

	Costs

	lowEnd     DPHit
	highEnd    DPHit
	vectors    [2][]int
	trapezoids []*filter.Trapezoid
	covered    []bool
	slot       int
	result     chan DPHit
}

// An offset slice seems to be the easiest way to implement the C idiom used in PALS to implement
// an offset (by o) view (v) on an array (a):
//
//  int *v, o;
//  int [n]a;
//  v = a - o;
//
// now v[i] is a view on a[i-o]
type offsetSlice struct {
	offset int
	slice  []int
}

func (o *offsetSlice) at(i int) int {
	i -= o.offset
	if i == -1 || i == len(o.slice) {
		return 0
	}
	return o.slice[i]
}
func (o *offsetSlice) set(i, v int) { o.slice[i-o.offset] = v }

var vecBuffering int = 100000

// Handle the recursive search for alignable segments.
func (k *kernel) alignRecursion(t *filter.Trapezoid) {
	mid := (t.Bottom + t.Top) / 2

	k.traceForward(mid, mid-t.Right, mid-t.Left)

	for x := 1; x == 1 || k.highEnd.Bbpos > mid+x*k.MaxIGap && k.highEnd.Score < k.lowEnd.Score; x++ {
		k.traceReverse(k.lowEnd.Bepos, k.lowEnd.Aepos, k.lowEnd.Aepos, mid+k.MaxIGap, k.BlockCost+2*x*k.DiffCost)
	}

	k.highEnd.Aepos, k.highEnd.Bepos = k.lowEnd.Aepos, k.lowEnd.Bepos

	lowTrap, highTrap := *t, *t
	lowTrap.Top = k.highEnd.Bbpos - k.MaxIGap
	highTrap.Bottom = k.highEnd.Bepos + k.MaxIGap

	if k.highEnd.Bepos-k.highEnd.Bbpos >= k.minLen && k.highEnd.Aepos-k.highEnd.Abpos >= k.minLen {
		indel := (k.highEnd.Abpos - k.highEnd.Bbpos) - (k.highEnd.Aepos - k.highEnd.Bepos)
		if indel < 0 {
			if indel == -indel {
				panic("dp: weird number overflow")
			}
			indel = -indel
		}
		identity := ((1 / k.RMatchCost) - float64(k.highEnd.Score-indel)/(k.RMatchCost*float64(k.highEnd.Bepos-k.highEnd.Bbpos)))

		if identity <= k.maxDiff {
			k.highEnd.Error = identity

			for i, trap := range k.trapezoids[k.slot+1:] {
				var trapAProjection, trapBProjection, coverageA, coverageB int

				if trap.Bottom >= k.highEnd.Bepos {
					break
				}

				trapBProjection = trap.Top - trap.Bottom + 1
				trapAProjection = trap.Right - trap.Left + 1
				if trap.Left < k.highEnd.LowDiagonal {
					coverageA = k.highEnd.LowDiagonal
				} else {
					coverageA = trap.Left
				}
				if trap.Right > k.highEnd.HighDiagonal {
					coverageB = k.highEnd.HighDiagonal
				} else {
					coverageB = trap.Right
				}

				if coverageA > coverageB {
					continue
				}

				coverageA = coverageB - coverageA + 1
				if trap.Top > k.highEnd.Bepos {
					coverageB = k.highEnd.Bepos - trap.Bottom + 1
				} else {
					coverageB = trapBProjection
				}

				if (float64(coverageA)/float64(trapAProjection))*(float64(coverageB)/float64(trapBProjection)) > 0.99 {
					k.covered[i] = true
				}
			}

			// Diagonals to this point are query-target, not target-query.
			k.highEnd.LowDiagonal, k.highEnd.HighDiagonal = -k.highEnd.HighDiagonal, -k.highEnd.LowDiagonal

			k.result <- k.highEnd
		}
	}

	if lowTrap.Top-lowTrap.Bottom > k.minLen && lowTrap.Top < t.Top-k.MaxIGap {
		k.alignRecursion(&lowTrap)
	}
	if highTrap.Top-highTrap.Bottom > k.minLen {
		k.alignRecursion(&highTrap)
	}
}

func (k *kernel) allocateVectors(required int) {
	vecMax := required + required>>2 + vecBuffering
	k.vectors[0] = make([]int, vecMax)
	k.vectors[1] = make([]int, vecMax)
}

// Forward and Reverse D.P. Extension Routines
// Called at the mid-point of trapezoid -- mid X [low,high], the extension
// is computed to an end point and the lowest and highest diagonals
// are recorded. These are returned in a partially filled DPHit
// record, that will be merged with that returned for extension in the
// opposite direction.
func (k *kernel) traceForward(mid, low, high int) {
	odd := false
	var (
		maxScore          int
		maxLeft, maxRight int
		maxI, maxJ        int
		i, j              int
	)

	// Set basis from (mid,low) .. (mid,high).
	if low < 0 {
		low = 0
	}
	if high > k.target.Len() {
		high = k.target.Len()
	}

	if required := (high - low) + k.MaxIGap; required >= len(k.vectors[0]) {
		k.allocateVectors(required)
	}

	thisVector := &offsetSlice{
		slice:  k.vectors[0],
		offset: low,
	}

	for j = low; j <= high; j++ {
		thisVector.set(j, 0)
	}

	high += k.MaxIGap
	if high > k.target.Len() {
		high = k.target.Len()
	}

	for ; j <= high; j++ {
		thisVector.set(j, thisVector.at(j-1)-k.DiffCost)
	}

	maxScore = 0
	maxRight = mid - low
	maxLeft = mid - high
	maxI = mid
	maxJ = low

	// Advance to next row.
	thatVector := &offsetSlice{}
	for i = mid; low <= high && i < k.query.Len(); i++ {
		var cost, score int

		*thatVector = *thisVector
		if !odd {
			thisVector.slice = k.vectors[1]
		} else {
			thisVector.slice = k.vectors[0]
		}
		thisVector.offset = low
		odd = !odd

		score = thatVector.at(low)
		thisVector.set(low, score-k.DiffCost)
		cost = thisVector.at(low)

		for j = low + 1; j <= high; j++ {
			var ratchet, temp int

			temp = cost
			cost = score
			score = thatVector.at(j)
			if k.query.Seq[i] == k.target.Seq[j-1] && k.valueToCode[k.query.Seq[i]] >= 0 {
				cost += k.MatchCost
			}

			ratchet = cost
			if score > ratchet {
				ratchet = score
			}
			if temp > ratchet {
				ratchet = temp
			}

			cost = ratchet - k.DiffCost
			thisVector.set(j, cost)
			if cost >= maxScore {
				maxScore = cost
				maxI = i + 1
				maxJ = j
			}
		}

		if j <= k.target.Len() {
			var ratchet int

			if k.query.Seq[i] == k.target.Seq[j-1] && k.valueToCode[k.query.Seq[i]] >= 0 {
				score += k.MatchCost
			}

			ratchet = score
			if cost > ratchet {
				ratchet = cost
			}

			score = ratchet - k.DiffCost
			thisVector.set(j, score)
			if score > maxScore {
				maxScore = score
				maxI = i + 1
				maxJ = j
			}

			for j++; j <= k.target.Len(); j++ {
				score -= k.DiffCost
				if score < maxScore-k.BlockCost {
					break
				}
				thisVector.set(j, score)
			}
		}

		high = j - 1

		for low <= high && thisVector.at(low) < maxScore-k.BlockCost {
			low++
		}
		for low <= high && thisVector.at(high) < maxScore-k.BlockCost {
			high--
		}

		if required := (high - low) + 2; required > len(k.vectors[0]) {
			k.allocateVectors(required)
		}

		if (i+1)-low > maxRight {
			maxRight = (i + 1) - low
		}
		if (i+1)-high < maxLeft {
			maxLeft = (i + 1) - high
		}
	}

	k.lowEnd.Aepos = maxJ
	k.lowEnd.Bepos = maxI
	k.lowEnd.LowDiagonal = maxLeft
	k.lowEnd.HighDiagonal = maxRight
	k.lowEnd.Score = maxScore
}

func (k *kernel) traceReverse(top, low, high, bottom, xfactor int) {
	odd := false
	var (
		maxScore          int
		maxLeft, maxRight int
		maxI, maxJ        int
		i, j              int
	)

	// Set basis from (top,low) .. (top,high).
	if low < 0 {
		low = 0
	}
	if high > k.target.Len() {
		high = k.target.Len()
	}

	if required := (high - low) + k.MaxIGap; required >= len(k.vectors[0]) {
		k.allocateVectors(required)
	}

	thisVector := &offsetSlice{
		slice:  k.vectors[0],
		offset: high - (len(k.vectors[0]) - 1),
	}
	for j = high; j >= low; j-- {
		thisVector.set(j, 0)
	}

	low -= k.MaxIGap
	if low < 0 {
		low = 0
	}

	for ; j >= low; j-- {
		thisVector.set(j, thisVector.at(j+1)-k.DiffCost)
	}

	maxScore = 0
	maxRight = top - low
	maxLeft = top - high
	maxI = top
	maxJ = low

	// Advance to next row.
	if top-1 <= bottom {
		xfactor = k.BlockCost
	}

	thatVector := &offsetSlice{}
	for i = top - 1; low <= high && i >= 0; i-- {
		var cost, score int

		*thatVector = *thisVector
		if !odd {
			thisVector.slice = k.vectors[1]
		} else {
			thisVector.slice = k.vectors[0]
		}
		thisVector.offset = high - (len(k.vectors[0]) - 1)
		odd = !odd

		score = thatVector.at(high)
		thisVector.set(high, score-k.DiffCost)
		cost = thisVector.at(high)

		for j = high - 1; j >= low; j-- {
			var ratchet, temp int

			temp = cost
			cost = score
			score = thatVector.at(j)
			if k.query.Seq[i] == k.target.Seq[j] && k.valueToCode[k.query.Seq[i]] >= 0 {
				cost += k.MatchCost
			}

			ratchet = cost
			if score > ratchet {
				ratchet = score
			}
			if temp > ratchet {
				ratchet = temp
			}

			cost = ratchet - k.DiffCost
			thisVector.set(j, cost)
			if cost >= maxScore {
				maxScore = cost
				maxI = i
				maxJ = j
			}
		}

		if j >= 0 {
			var ratchet int

			if k.query.Seq[i] == k.target.Seq[j] && k.valueToCode[k.query.Seq[i]] >= 0 {
				score += k.MatchCost
			}

			ratchet = score
			if cost > ratchet {
				ratchet = cost
			}

			score = ratchet - k.DiffCost
			thisVector.set(j, score)
			if score > maxScore {
				maxScore = score
				maxI = i
				maxJ = j
			}

			for j--; j >= 0; j-- {
				score -= k.DiffCost
				if score < maxScore-xfactor {
					break
				}
				thisVector.set(j, score)
			}
		}

		low = j + 1

		for low <= high && thisVector.at(low) < maxScore-xfactor {
			low++
		}
		for low <= high && thisVector.at(high) < maxScore-xfactor {
			high--
		}

		if i == bottom {
			xfactor = k.BlockCost
		}

		if required := (high - low) + 2; required > len(k.vectors[0]) {
			k.allocateVectors(required)
		}

		if i-low > maxRight {
			maxRight = i - low
		}
		if i-high < maxLeft {
			maxLeft = i - high
		}
	}

	k.highEnd.Abpos = maxJ
	k.highEnd.Bbpos = maxI
	k.highEnd.LowDiagonal = maxLeft
	k.highEnd.HighDiagonal = maxRight
	k.highEnd.Score = maxScore
}
