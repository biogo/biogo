package filter
// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
//   This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU General Public License for more details.
//
//   You should have received a copy of the GNU General Public License
//   along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
import (
	"bio/seq"
	"bio/index/kmerindex"
	"sort"
)

const (
	diagonalPadding = 2
)

type Merger struct { // these are static so should be part of object Merger
	target, query              *seq.Seq
	filterParams               *Params
	leftPadding, bottomPadding int
	binWidth                   int
	selfComparison             bool
	freeTraps, trapList        *Trapezoid
	trapOrder, tail            *Trapezoid
	eoTerm                     *Trapezoid
	trapCount                  int
}

func NewMerger(index *kmerindex.Index, query *seq.Seq, filterParams *Params, selfCompare bool) (m *Merger) {
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
		target:        index.Seq,
		filterParams:  filterParams,
		query:         query,
		selfComparison: selfCompare,
		bottomPadding: index.GetK() + 2,
		leftPadding:   leftPadding,
		binWidth:      binWidth,
		eoTerm:        eoTerm,
		trapOrder:     eoTerm,
	}

	return m
}

func (self *Merger) MergeFilterHit(hit *FilterHit) {
	Left := -hit.DiagIndex
	if self.selfComparison && Left <= self.filterParams.MaxError {
		return
	}
	Top := hit.QTo
	Bottom := hit.QFrom

	var temp, free *Trapezoid
	for base := self.trapOrder; true; base = temp {
		temp = base.Next
		if Bottom-self.bottomPadding > base.Top {
			if free == nil {
				self.trapOrder = temp
			} else {
				free.Join(temp)
			}
			self.trapList = base.Join(self.trapList)
			self.trapCount++
		} else if Left-diagonalPadding > base.Right {
			free = base
		} else if Left+self.leftPadding >= base.Left {
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

				free.Join(temp)
				self.freeTraps = base.Join(self.freeTraps)
			} else if temp != nil && temp.Left-diagonalPadding <= base.Right {
				base.Right = temp.Right
				if base.Bottom > temp.Bottom {
					base.Bottom = temp.Bottom
				}
				if base.Top < temp.Top {
					base.Top = temp.Top
				}
				base.Join(temp.Next)
				self.freeTraps = temp.Join(self.freeTraps)
				temp = base.Next
			}

			return
		} else {
			if self.freeTraps == nil {
				self.freeTraps = &Trapezoid{}
			}
			if free == nil {
				self.trapOrder = self.freeTraps
			} else {
				free.Join(self.freeTraps)
			}

			free, self.freeTraps = self.freeTraps.Decapitate()
			free.Join(base)

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
		lagPosition := base.Bottom - MaxIGap + 1
		if lagPosition < 0 {
			lagPosition = 0
		}
		lastPosition := base.Top + MaxIGap
		if lastPosition > self.query.Len() {
			lastPosition = self.query.Len()
		}

		i := 0
		for i = lagPosition; i < lastPosition; i++ {
			if lookUp.ValueToCode[self.query.Seq[i]] >= 0 {
				if i-lagPosition >= MaxIGap {
					if lagPosition-base.Bottom > 0 {
						if self.freeTraps == nil {
							self.freeTraps = &Trapezoid{}
						}

						self.freeTraps = self.freeTraps.Shunt(base)

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
		if i-lagPosition >= MaxIGap {
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

		lagPosition := aBottom - MaxIGap + 1
		if lagPosition < 0 {
			lagPosition = 0
		}
		lastPosition := aTop + MaxIGap
		if lastPosition > self.target.Len() {
			lastPosition = self.target.Len()
		}

		lagClip := aBottom
		i := 0
		for i = lagPosition; i < lastPosition; i++ {
			if lookUp.ValueToCode[self.target.Seq[i]] >= 0 {
				if i-lagPosition >= MaxIGap {
					if lagPosition > lagClip {
						if self.freeTraps == nil {
							self.freeTraps = &Trapezoid{}
						}

						self.freeTraps = self.freeTraps.Shunt(base)

						base.Clip(lagPosition, lagClip)

						base = base.Next
						self.trapCount++
					}
					lagClip = i
				}
				lagPosition = i + 1
			}
		}

		if i-lagPosition < MaxIGap {
			lagPosition = aTop
		}

		base.Clip(lagPosition, lagClip)

		self.tail = base
	}
}

func (self *Merger) FinaliseMerge() (trapezoids Trapezoids) {
	var next *Trapezoid
	for base := self.trapOrder; base != self.eoTerm; base = next {
		next = base.Next
		self.trapList = base.Join(self.trapList)
		self.trapCount++
	}

	self.clipVertical()
	self.clipTrapezoids()

	if self.tail != nil {
		self.freeTraps = self.tail.Join(self.freeTraps)
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

func SumTrapLengths(trapezoids Trapezoids) (sum int) {
	for _, temp := range trapezoids {
		length := temp.Top - temp.Bottom
		sum += length
	}
	return sum
}
