package dp

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
	"github.com/kortschak/BioGo/align/pals/filter"
	"github.com/kortschak/BioGo/seq"
)

const (
	low = iota
	high
)

// A kernel handles the actual dp alignment process.
type kernel struct {
	target, query *seq.Seq
	minLen        int
	maxDiff       float64
	lowEnd        DPHit
	highEnd       DPHit
	vectors       [2][]int
	trapezoids    []*filter.Trapezoid
	covered       []bool
	segments      DPHits
	slot          int
	result        chan DPHit
}

// An offset slice seems to be the easiest way to implement the C idiom used in PALS to implement
// an offset (by o)  view (v) on an array (a):
//  int *v, o;
//  int [n]a;
//  v = a - o;
//  // now v[i] is a view on a[i-o]
// No method is provided (perhaps when inlining is further implemented).
type offsetSlice struct {
	offset int
	slice  []int
}

var vecBuffering int = 100000

// Handle the recusive search for alignable segments.
func (self *kernel) alignRecursion(workingTrap *filter.Trapezoid) {
	mid := (workingTrap.Bottom + workingTrap.Top) / 2

	self.traceForward(mid, mid-workingTrap.Right, mid-workingTrap.Left)

	for x := 1; x == 1 || self.highEnd.Bbpos > mid+x*MaxIGap && self.highEnd.Score < self.lowEnd.Score; x++ {
		self.traceReverse(self.lowEnd.Bepos, self.lowEnd.Aepos, self.lowEnd.Aepos, mid+MaxIGap, BlockCost+2*x*DiffCost)
	}

	self.highEnd.Aepos, self.highEnd.Bepos = self.lowEnd.Aepos, self.lowEnd.Bepos

	lowTrap, highTrap := *workingTrap, *workingTrap
	lowTrap.Top = self.highEnd.Bbpos - MaxIGap
	highTrap.Bottom = self.highEnd.Bepos + MaxIGap

	if self.highEnd.Bepos-self.highEnd.Bbpos >= self.minLen && self.highEnd.Aepos-self.highEnd.Abpos >= self.minLen {
		indel := (self.highEnd.Abpos - self.highEnd.Bbpos) - (self.highEnd.Aepos - self.highEnd.Bepos)
		if indel < 0 {
			indel = -indel
		}
		identity := ((1 / RMatchCost) - float64(self.highEnd.Score-indel)/(RMatchCost*float64(self.highEnd.Bepos-self.highEnd.Bbpos)))

		if identity <= self.maxDiff {
			self.highEnd.Error = identity

			for i, trap := range self.trapezoids[self.slot+1:] {
				var trapAProjection, trapBProjection, coverageA, coverageB int

				if trap.Bottom >= self.highEnd.Bepos {
					break
				}

				trapBProjection = trap.Top - trap.Bottom + 1
				trapAProjection = trap.Right - trap.Left + 1
				if trap.Left < self.highEnd.LowDiagonal {
					coverageA = self.highEnd.LowDiagonal
				} else {
					coverageA = trap.Left
				}
				if trap.Right > self.highEnd.HighDiagonal {
					coverageB = self.highEnd.HighDiagonal
				} else {
					coverageB = trap.Right
				}

				if coverageA > coverageB {
					continue
				}

				coverageA = coverageB - coverageA + 1
				if trap.Top > self.highEnd.Bepos {
					coverageB = self.highEnd.Bepos - trap.Bottom + 1
				} else {
					coverageB = trapBProjection
				}

				if (float64(coverageA)/float64(trapAProjection))*(float64(coverageB)/float64(trapBProjection)) > 0.99 {
					self.covered[i] = true
				}
			}

			// diagonals to this point are query-target, not target-query.
			self.highEnd.LowDiagonal, self.highEnd.HighDiagonal = -self.highEnd.HighDiagonal, -self.highEnd.LowDiagonal

			self.segments = append(self.segments, self.highEnd)
		}
	}

	if lowTrap.Top-lowTrap.Bottom > self.minLen && lowTrap.Top < workingTrap.Top-MaxIGap {
		self.alignRecursion(&lowTrap)
	}
	if highTrap.Top-highTrap.Bottom > self.minLen {
		self.alignRecursion(&highTrap)
	}
}

func (self *kernel) allocateVectors(required int) {
	vecMax := required + required>>2 + vecBuffering
	self.vectors[0] = make([]int, vecMax)
	self.vectors[1] = make([]int, vecMax)
}

// Forward and Reverse D.P. Extension Routines
// Called at the mid-point of trapezoid -- mid X [low,high], the extension
// is computed to an end point and the lowest and highest diagonals
// are recorded. These are returned in a partially filled DPHit
// record, that will be merged with that returned for extension in the
// opposite direction.
func (self *kernel) traceForward(mid, low, high int) {
	thisVector := &offsetSlice{}
	odd := false
	var maxScore, maxLeft, maxRight, maxI, maxJ int
	var i, j int

	/* Set basis from (mid,low) .. (mid,high) */
	if low < 0 {
		low = 0
	}
	if high > self.target.Len() {
		high = self.target.Len()
	}

	if required := (high - low) + MaxIGap; required >= len(self.vectors[0]) {
		self.allocateVectors(required)
	}

	thisVector.slice = self.vectors[0]
	thisVector.offset = low

	for j = low; j <= high; j++ {
		thisVector.slice[j-thisVector.offset] = 0
	}

	high += MaxIGap
	if high > self.target.Len() {
		high = self.target.Len()
	}

	for ; j <= high; j++ {
		thisVector.slice[j-thisVector.offset] = thisVector.slice[j-thisVector.offset-1] - DiffCost
	}

	maxScore = 0
	maxRight = mid - low
	maxLeft = mid - high
	maxI = mid
	maxJ = low

	/* Advance to next row */
	for i = mid; low <= high && i < self.query.Len(); i++ {
		var cost, score int

		thatVector := *thisVector
		if !odd {
			thisVector.slice = self.vectors[1]
		} else {
			thisVector.slice = self.vectors[0]
		}
		thisVector.offset = low
		odd = !odd

		score = thatVector.slice[low-thatVector.offset]
		thisVector.slice[low-thisVector.offset] = score - DiffCost
		cost = thisVector.slice[low-thisVector.offset]

		for j = low + 1; j <= high; j++ {
			var ratchet, temp int

			temp = cost
			cost = score
			score = thatVector.slice[j-thatVector.offset]
			if self.query.Seq[i] == self.target.Seq[j-1] && lookUp.ValueToCode[self.query.Seq[i]] >= 0 {
				cost += MatchCost
			}

			ratchet = cost
			if score > ratchet {
				ratchet = score
			}
			if temp > ratchet {
				ratchet = temp
			}

			cost = ratchet - DiffCost
			thisVector.slice[j-thisVector.offset] = cost
			if cost >= maxScore {
				maxScore = cost
				maxI = i + 1
				maxJ = j
			}
		}

		if j <= self.target.Len() {
			var ratchet int

			if self.query.Seq[i] == self.target.Seq[j-1] && lookUp.ValueToCode[self.query.Seq[i]] >= 0 {
				score += MatchCost
			}

			ratchet = score
			if cost > ratchet {
				ratchet = cost
			}

			score = ratchet - DiffCost
			thisVector.slice[j-thisVector.offset] = score
			if score > maxScore {
				maxScore = score
				maxI = i + 1
				maxJ = j
			}

			for j++; j <= self.target.Len(); j++ {
				score -= DiffCost
				if score < maxScore-BlockCost {
					break
				}
				thisVector.slice[j-thisVector.offset] = score
			}
		}

		high = j - 1

		for low <= high && thisVector.slice[low-thisVector.offset] < maxScore-BlockCost {
			low++
		}
		for low <= high && thisVector.slice[high-thisVector.offset] < maxScore-BlockCost {
			high--
		}

		if required := (high - low) + 2; required > len(self.vectors[0]) {
			self.allocateVectors(required)
		}

		if (i+1)-low > maxRight {
			maxRight = (i + 1) - low
		}
		if (i+1)-high < maxLeft {
			maxLeft = (i + 1) - high
		}
	}

	self.lowEnd.Aepos = maxJ
	self.lowEnd.Bepos = maxI
	self.lowEnd.LowDiagonal = maxLeft
	self.lowEnd.HighDiagonal = maxRight
	self.lowEnd.Score = maxScore
}

func (self *kernel) traceReverse(top, low, high, bottom, xfactor int) {
	thisVector := &offsetSlice{}
	odd := false
	var maxScore, maxLeft, maxRight, maxI, maxJ int
	var i, j int

	/* Set basis from (top,low) .. (top,high) */
	if low < 0 {
		low = 0
	}
	if high > self.target.Len() {
		high = self.target.Len()
	}

	if required := (high - low) + MaxIGap; required >= len(self.vectors[0]) {
		self.allocateVectors(required)
	}

	thisVector.slice = self.vectors[0]
	thisVector.offset = ((len(self.vectors[0]) - 1) - high)

	for j = high; j >= low; j-- {
		thisVector.slice[j+thisVector.offset] = 0
	}

	low -= MaxIGap
	if low < 0 {
		low = 0
	}

	for ; j >= low; j-- {
		thisVector.slice[j+thisVector.offset] = thisVector.slice[j+thisVector.offset+1] - DiffCost
	}

	maxScore = 0
	maxRight = top - low
	maxLeft = top - high
	maxI = top
	maxJ = low

	/* Advance to next row */
	if top-1 <= bottom {
		xfactor = BlockCost
	}

	for i = top - 1; low <= high && i >= 0; i-- {
		var cost, score int

		thatVector := *thisVector
		if !odd {
			thisVector.slice = self.vectors[1]
		} else {
			thisVector.slice = self.vectors[0]
		}
		thisVector.offset = ((len(self.vectors[0]) - 1) - high)
		odd = !odd

		score = thatVector.slice[high+thatVector.offset]
		thisVector.slice[high+thisVector.offset] = score - DiffCost
		cost = thisVector.slice[high+thisVector.offset]

		for j = high - 1; j >= low; j-- {
			var ratchet, temp int

			temp = cost
			cost = score
			score = thatVector.slice[j+thatVector.offset]
			if self.query.Seq[i] == self.target.Seq[j] && lookUp.ValueToCode[self.query.Seq[i]] >= 0 {
				cost += MatchCost
			}

			ratchet = cost
			if score > ratchet {
				ratchet = score
			}
			if temp > ratchet {
				ratchet = temp
			}

			cost = ratchet - DiffCost
			thisVector.slice[j+thisVector.offset] = cost
			if cost >= maxScore {
				maxScore = cost
				maxI = i
				maxJ = j
			}
		}

		if j >= 0 {
			var ratchet int

			if self.query.Seq[i] == self.target.Seq[j] && lookUp.ValueToCode[self.query.Seq[i]] >= 0 {
				score += MatchCost
			}

			ratchet = score
			if cost > ratchet {
				ratchet = cost
			}

			score = ratchet - DiffCost
			thisVector.slice[j+thisVector.offset] = score
			if score > maxScore {
				maxScore = score
				maxI = i
				maxJ = j
			}

			for j--; j >= 0; j-- {
				score -= DiffCost
				if score < maxScore-xfactor {
					break
				}
				thisVector.slice[j+thisVector.offset] = score
			}
		}

		low = j + 1

		for low <= high && thisVector.slice[low+thisVector.offset] < maxScore-xfactor {
			low++
		}
		for low <= high && thisVector.slice[high+thisVector.offset] < maxScore-xfactor {
			high--
		}

		if i == bottom {
			xfactor = BlockCost
		}

		if required := (high - low) + 2; required > len(self.vectors[0]) {
			self.allocateVectors(required)
		}

		if i-low > maxRight {
			maxRight = i - low
		}
		if i-high < maxLeft {
			maxLeft = i - high
		}
	}

	self.highEnd.Abpos = maxJ
	self.highEnd.Bbpos = maxI
	self.highEnd.LowDiagonal = maxLeft
	self.highEnd.HighDiagonal = maxRight
	self.highEnd.Score = maxScore
}
