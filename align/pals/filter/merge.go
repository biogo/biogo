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

package filter

import (
	"code.google.com/p/biogo/index/kmerindex"
	"code.google.com/p/biogo/seq"
	"sort"
)

const (
	diagonalPadding = 2
)

// A Merger aggregates and clips an ordered set of trapezoids.
type Merger struct {
	target, query              *seq.Seq
	filterParams               *Params
	maxIGap                    int
	leftPadding, bottomPadding int
	binWidth                   int
	selfComparison             bool
	freeTraps, trapList        *Trapezoid
	trapOrder, tail            *Trapezoid
	eoTerm                     *Trapezoid
	trapCount                  int
}

// Create a new Merger using the provided kmerindex, query sequence, filter parameters and maximum inter-segment gap length.
// If selfCompare is true only the upper diagonal of the comparison matrix is examined.
func NewMerger(index *kmerindex.Index, query *seq.Seq, filterParams *Params, maxIGap int, selfCompare bool) (m *Merger) {
	tubeWidth := filterParams.TubeOffset + filterParams.MaxError
	binWidth := tubeWidth - 1
	leftPadding := diagonalPadding + binWidth

	eoTerm := &Trapezoid{
		Left:   query.Len() + 1 + leftPadding,
		Right:  query.Len() + 1,
		Bottom: -1,
		Top:    query.Len() + 1,
		Next:   nil,
	}

	m = &Merger{
		target:         index.Seq,
		filterParams:   filterParams,
		maxIGap:        maxIGap,
		query:          query,
		selfComparison: selfCompare,
		bottomPadding:  index.GetK() + 2,
		leftPadding:    leftPadding,
		binWidth:       binWidth,
		eoTerm:         eoTerm,
		trapOrder:      eoTerm,
	}

	return m
}

// Merge a filter hit into the collection.
func (self *Merger) MergeFilterHit(hit *FilterHit) {
	Left := -hit.DiagIndex
	if self.selfComparison && Left <= self.filterParams.MaxError {
		return
	}
	Top := hit.QTo
	Bottom := hit.QFrom

	var temp, free *Trapezoid
	for base := self.trapOrder; ; base = temp {
		temp = base.Next
		switch {
		case Bottom-self.bottomPadding > base.Top:
			if free == nil {
				self.trapOrder = temp
			} else {
				free.join(temp)
			}
			self.trapList = base.join(self.trapList)
			self.trapCount++
		case Left-diagonalPadding > base.Right:
			free = base
		case Left+self.leftPadding >= base.Left:
			if Left+self.binWidth > base.Right {
				base.Right = Left + self.binWidth
			}
			if Left < base.Left {
				base.Left = Left
			}
			if Top > base.Top {
				base.Top = Top
			}

			if free != nil && free.Right+diagonalPadding >= base.Left {
				free.Right = base.Right
				if free.Bottom > base.Bottom {
					free.Bottom = base.Bottom
				}
				if free.Top < base.Top {
					free.Top = base.Top
				}

				free.join(temp)
				self.freeTraps = base.join(self.freeTraps)
			} else if temp != nil && temp.Left-diagonalPadding <= base.Right {
				base.Right = temp.Right
				if base.Bottom > temp.Bottom {
					base.Bottom = temp.Bottom
				}
				if base.Top < temp.Top {
					base.Top = temp.Top
				}
				base.join(temp.Next)
				self.freeTraps = temp.join(self.freeTraps)
				temp = base.Next
			}

			return
		default:
			if self.freeTraps == nil {
				self.freeTraps = &Trapezoid{}
			}
			if free == nil {
				self.trapOrder = self.freeTraps
			} else {
				free.join(self.freeTraps)
			}

			free, self.freeTraps = self.freeTraps.decapitate()
			free.join(base)

			free.Top = Top
			free.Bottom = Bottom
			free.Left = Left
			free.Right = Left + self.binWidth

			return
		}
	}
}

func (self *Merger) clipVertical() {
	for base := self.trapList; base != nil; base = base.Next {
		lagPosition := base.Bottom - self.maxIGap + 1
		if lagPosition < 0 {
			lagPosition = 0
		}
		lastPosition := base.Top + self.maxIGap
		if lastPosition > self.query.Len() {
			lastPosition = self.query.Len()
		}

		i := 0
		for i = lagPosition; i < lastPosition; i++ {
			if lookUp.ValueToCode[self.query.Seq[i]] >= 0 {
				if i-lagPosition >= self.maxIGap {
					if lagPosition-base.Bottom > 0 {
						if self.freeTraps == nil {
							self.freeTraps = &Trapezoid{}
						}

						self.freeTraps = self.freeTraps.shunt(base)

						base.Top = lagPosition
						base = base.Next
						base.Bottom = i
						self.trapCount++
					} else {
						base.Bottom = i
					}
				}
				lagPosition = i + 1
			}
		}
		if i-lagPosition >= self.maxIGap {
			base.Top = lagPosition
		}
	}
}

func (self *Merger) clipTrapezoids() {
	for base := self.trapList; base != nil; base = base.Next {
		if base.Top-base.Bottom < self.bottomPadding-2 {
			continue
		}

		aBottom := base.Bottom - base.Right
		aTop := base.Top - base.Left

		lagPosition := aBottom - self.maxIGap + 1
		if lagPosition < 0 {
			lagPosition = 0
		}
		lastPosition := aTop + self.maxIGap
		if lastPosition > self.target.Len() {
			lastPosition = self.target.Len()
		}

		lagClip := aBottom
		i := 0
		for i = lagPosition; i < lastPosition; i++ {
			if lookUp.ValueToCode[self.target.Seq[i]] >= 0 {
				if i-lagPosition >= self.maxIGap {
					if lagPosition > lagClip {
						if self.freeTraps == nil {
							self.freeTraps = &Trapezoid{}
						}

						self.freeTraps = self.freeTraps.shunt(base)

						base.clip(lagPosition, lagClip)

						base = base.Next
						self.trapCount++
					}
					lagClip = i
				}
				lagPosition = i + 1
			}
		}

		if i-lagPosition < self.maxIGap {
			lagPosition = aTop
		}

		base.clip(lagPosition, lagClip)

		self.tail = base
	}
}

// Finalise the merged collection and return a sorted slice of Trapezoids.
func (self *Merger) FinaliseMerge() (trapezoids Trapezoids) {
	var next *Trapezoid
	for base := self.trapOrder; base != self.eoTerm; base = next {
		next = base.Next
		self.trapList = base.join(self.trapList)
		self.trapCount++
	}

	self.clipVertical()
	self.clipTrapezoids()

	if self.tail != nil {
		self.freeTraps = self.tail.join(self.freeTraps)
	}

	trapezoids = make(Trapezoids, self.trapCount)
	for i, z := 0, self.trapList; i < self.trapCount; i++ {
		trapezoids[i] = z
		z = z.Next
		trapezoids[i].Next = nil
	}

	sort.Sort(Trapezoids(trapezoids))

	return
}
