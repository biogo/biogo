// Copyright ©2011-2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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

package dp

import (
	"github.com/kortschak/biogo/align/pals/filter"
	"github.com/kortschak/biogo/seq"
)

const (
	low = iota
	high
)

// A kernel handles the actual dp alignment process.
type kernel struct {
	target, query *seq.Seq

	minLen  int
	maxDiff float64

	maxIGap    int
	diffCost   int
	sameCost   int
	matchCost  int
	blockCost  int
	rMatchCost float64

	lowEnd     DPHit
	highEnd    DPHit
	vectors    [2][]int
	trapezoids []*filter.Trapezoid
	covered    []bool
	slot       int
	result     chan DPHit
}

// An offset slice seems to be the easiest way to implement the C idiom used in PALS to implement
// an offset (by o)  view (v) on an array (a):
//  int *v, o;
//  int [n]a;
//  v = a - o;
//  // now v[i] is a view on a[i-o]
type offsetSlice struct {
	offset int
	slice  []int
}

func (o *offsetSlice) at(i int) (v int) { return o.slice[i-o.offset] } // v return name due to go issue 3315 TODO: Remove when issue is resolved.
func (o *offsetSlice) set(i, v int)     { o.slice[i-o.offset] = v }

var vecBuffering int = 100000

// Handle the recusive search for alignable segments.
func (self *kernel) alignRecursion(workingTrap *filter.Trapezoid) {
	mid := (workingTrap.Bottom + workingTrap.Top) / 2

	self.traceForward(mid, mid-workingTrap.Right, mid-workingTrap.Left)

	for x := 1; x == 1 || self.highEnd.Bbpos > mid+x*self.maxIGap && self.highEnd.Score < self.lowEnd.Score; x++ {
		self.traceReverse(self.lowEnd.Bepos, self.lowEnd.Aepos, self.lowEnd.Aepos, mid+self.maxIGap, self.blockCost+2*x*self.diffCost)
	}

	self.highEnd.Aepos, self.highEnd.Bepos = self.lowEnd.Aepos, self.lowEnd.Bepos

	lowTrap, highTrap := *workingTrap, *workingTrap
	lowTrap.Top = self.highEnd.Bbpos - self.maxIGap
	highTrap.Bottom = self.highEnd.Bepos + self.maxIGap

	if self.highEnd.Bepos-self.highEnd.Bbpos >= self.minLen && self.highEnd.Aepos-self.highEnd.Abpos >= self.minLen {
		indel := (self.highEnd.Abpos - self.highEnd.Bbpos) - (self.highEnd.Aepos - self.highEnd.Bepos)
		if indel < 0 {
			if indel == -indel {
				panic("dp: weird number overflow")
			}
			indel = -indel
		}
		identity := ((1 / self.rMatchCost) - float64(self.highEnd.Score-indel)/(self.rMatchCost*float64(self.highEnd.Bepos-self.highEnd.Bbpos)))

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

			self.result <- self.highEnd
		}
	}

	if lowTrap.Top-lowTrap.Bottom > self.minLen && lowTrap.Top < workingTrap.Top-self.maxIGap {
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
	odd := false
	var (
		maxScore          int
		maxLeft, maxRight int
		maxI, maxJ        int
		i, j              int
	)

	/* Set basis from (mid,low) .. (mid,high) */
	if low < 0 {
		low = 0
	}
	if high > self.target.Len() {
		high = self.target.Len()
	}

	if required := (high - low) + self.maxIGap; required >= len(self.vectors[0]) {
		self.allocateVectors(required)
	}

	thisVector := &offsetSlice{
		slice:  self.vectors[0],
		offset: low,
	}

	for j = low; j <= high; j++ {
		thisVector.set(j, 0)
	}

	high += self.maxIGap
	if high > self.target.Len() {
		high = self.target.Len()
	}

	for ; j <= high; j++ {
		thisVector.set(j, thisVector.at(j-1)-self.diffCost)
	}

	maxScore = 0
	maxRight = mid - low
	maxLeft = mid - high
	maxI = mid
	maxJ = low

	/* Advance to next row */
	thatVector := &offsetSlice{}
	for i = mid; low <= high && i < self.query.Len(); i++ {
		var cost, score int

		*thatVector = *thisVector
		if !odd {
			thisVector.slice = self.vectors[1]
		} else {
			thisVector.slice = self.vectors[0]
		}
		thisVector.offset = low
		odd = !odd

		score = thatVector.at(low)
		thisVector.set(low, score-self.diffCost)
		cost = thisVector.at(low)

		for j = low + 1; j <= high; j++ {
			var ratchet, temp int

			temp = cost
			cost = score
			score = thatVector.at(j)
			if self.query.Seq[i] == self.target.Seq[j-1] && lookUp.ValueToCode[self.query.Seq[i]] >= 0 {
				cost += self.matchCost
			}

			ratchet = cost
			if score > ratchet {
				ratchet = score
			}
			if temp > ratchet {
				ratchet = temp
			}

			cost = ratchet - self.diffCost
			thisVector.set(j, cost)
			if cost >= maxScore {
				maxScore = cost
				maxI = i + 1
				maxJ = j
			}
		}

		if j <= self.target.Len() {
			var ratchet int

			if self.query.Seq[i] == self.target.Seq[j-1] && lookUp.ValueToCode[self.query.Seq[i]] >= 0 {
				score += self.matchCost
			}

			ratchet = score
			if cost > ratchet {
				ratchet = cost
			}

			score = ratchet - self.diffCost
			thisVector.set(j, score)
			if score > maxScore {
				maxScore = score
				maxI = i + 1
				maxJ = j
			}

			for j++; j <= self.target.Len(); j++ {
				score -= self.diffCost
				if score < maxScore-self.blockCost {
					break
				}
				thisVector.set(j, score)
			}
		}

		high = j - 1

		for low <= high && thisVector.at(low) < maxScore-self.blockCost {
			low++
		}
		for low <= high && thisVector.at(high) < maxScore-self.blockCost {
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
	odd := false
	var (
		maxScore          int
		maxLeft, maxRight int
		maxI, maxJ        int
		i, j              int
	)

	/* Set basis from (top,low) .. (top,high) */
	if low < 0 {
		low = 0
	}
	if high > self.target.Len() {
		high = self.target.Len()
	}

	if required := (high - low) + self.maxIGap; required >= len(self.vectors[0]) {
		self.allocateVectors(required)
	}

	thisVector := &offsetSlice{
		slice:  self.vectors[0],
		offset: high - (len(self.vectors[0]) - 1),
	}
	for j = high; j >= low; j-- {
		thisVector.set(j, 0)
	}

	low -= self.maxIGap
	if low < 0 {
		low = 0
	}

	for ; j >= low; j-- {
		thisVector.set(j, thisVector.at(j+1)-self.diffCost)
	}

	maxScore = 0
	maxRight = top - low
	maxLeft = top - high
	maxI = top
	maxJ = low

	/* Advance to next row */
	if top-1 <= bottom {
		xfactor = self.blockCost
	}

	thatVector := &offsetSlice{}
	for i = top - 1; low <= high && i >= 0; i-- {
		var cost, score int

		*thatVector = *thisVector
		if !odd {
			thisVector.slice = self.vectors[1]
		} else {
			thisVector.slice = self.vectors[0]
		}
		thisVector.offset = high - (len(self.vectors[0]) - 1)
		odd = !odd

		score = thatVector.at(high)
		thisVector.set(high, score-self.diffCost)
		cost = thisVector.at(high)

		for j = high - 1; j >= low; j-- {
			var ratchet, temp int

			temp = cost
			cost = score
			score = thatVector.at(j)
			if self.query.Seq[i] == self.target.Seq[j] && lookUp.ValueToCode[self.query.Seq[i]] >= 0 {
				cost += self.matchCost
			}

			ratchet = cost
			if score > ratchet {
				ratchet = score
			}
			if temp > ratchet {
				ratchet = temp
			}

			cost = ratchet - self.diffCost
			thisVector.set(j, cost)
			if cost >= maxScore {
				maxScore = cost
				maxI = i
				maxJ = j
			}
		}

		if j >= 0 {
			var ratchet int

			if self.query.Seq[i] == self.target.Seq[j] && lookUp.ValueToCode[self.query.Seq[i]] >= 0 {
				score += self.matchCost
			}

			ratchet = score
			if cost > ratchet {
				ratchet = cost
			}

			score = ratchet - self.diffCost
			thisVector.set(j, score)
			if score > maxScore {
				maxScore = score
				maxI = i
				maxJ = j
			}

			for j--; j >= 0; j-- {
				score -= self.diffCost
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
			xfactor = self.blockCost
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
