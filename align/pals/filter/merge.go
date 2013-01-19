// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package filter

import (
	"code.google.com/p/biogo/exp/seq/linear"
	"code.google.com/p/biogo/index/kmerindex"

	"sort"
)

const (
	diagonalPadding = 2
)

// A Merger aggregates and clips an ordered set of trapezoids.
type Merger struct {
	target, query              *linear.Seq
	filterParams               *Params
	maxIGap                    int
	leftPadding, bottomPadding int
	binWidth                   int
	selfComparison             bool
	freeTraps, trapList        *Trapezoid
	trapOrder, tail            *Trapezoid
	eoTerm                     *Trapezoid
	trapCount                  int
	valueToCode                []int
}

// Create a new Merger using the provided kmerindex, query sequence, filter parameters and maximum inter-segment gap length.
// If selfCompare is true only the upper diagonal of the comparison matrix is examined.
func NewMerger(ki *kmerindex.Index, query *linear.Seq, filterParams *Params, maxIGap int, selfCompare bool) *Merger {
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

	return &Merger{
		target:         ki.Seq(),
		filterParams:   filterParams,
		maxIGap:        maxIGap,
		query:          query,
		selfComparison: selfCompare,
		bottomPadding:  ki.K() + 2,
		leftPadding:    leftPadding,
		binWidth:       binWidth,
		eoTerm:         eoTerm,
		trapOrder:      eoTerm,
		valueToCode:    ki.Seq().Alpha.LetterIndex(),
	}
}

// Merge a filter hit into the collection.
func (m *Merger) MergeFilterHit(hit *FilterHit) {
	Left := -hit.DiagIndex
	if m.selfComparison && Left <= m.filterParams.MaxError {
		return
	}
	Top := hit.QTo
	Bottom := hit.QFrom

	var temp, free *Trapezoid
	for base := m.trapOrder; ; base = temp {
		temp = base.Next
		switch {
		case Bottom-m.bottomPadding > base.Top:
			if free == nil {
				m.trapOrder = temp
			} else {
				free.join(temp)
			}
			m.trapList = base.join(m.trapList)
			m.trapCount++
		case Left-diagonalPadding > base.Right:
			free = base
		case Left+m.leftPadding >= base.Left:
			if Left+m.binWidth > base.Right {
				base.Right = Left + m.binWidth
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
				m.freeTraps = base.join(m.freeTraps)
			} else if temp != nil && temp.Left-diagonalPadding <= base.Right {
				base.Right = temp.Right
				if base.Bottom > temp.Bottom {
					base.Bottom = temp.Bottom
				}
				if base.Top < temp.Top {
					base.Top = temp.Top
				}
				base.join(temp.Next)
				m.freeTraps = temp.join(m.freeTraps)
				temp = base.Next
			}

			return
		default:
			if m.freeTraps == nil {
				m.freeTraps = &Trapezoid{}
			}
			if free == nil {
				m.trapOrder = m.freeTraps
			} else {
				free.join(m.freeTraps)
			}

			free, m.freeTraps = m.freeTraps.decapitate()
			free.join(base)

			free.Top = Top
			free.Bottom = Bottom
			free.Left = Left
			free.Right = Left + m.binWidth

			return
		}
	}
}

func (m *Merger) clipVertical() {
	for base := m.trapList; base != nil; base = base.Next {
		lagPosition := base.Bottom - m.maxIGap + 1
		if lagPosition < 0 {
			lagPosition = 0
		}
		lastPosition := base.Top + m.maxIGap
		if lastPosition > m.query.Len() {
			lastPosition = m.query.Len()
		}

		i := 0
		for i = lagPosition; i < lastPosition; i++ {
			if m.valueToCode[m.query.Seq[i]] >= 0 {
				if i-lagPosition >= m.maxIGap {
					if lagPosition-base.Bottom > 0 {
						if m.freeTraps == nil {
							m.freeTraps = &Trapezoid{}
						}

						m.freeTraps = m.freeTraps.shunt(base)

						base.Top = lagPosition
						base = base.Next
						base.Bottom = i
						m.trapCount++
					} else {
						base.Bottom = i
					}
				}
				lagPosition = i + 1
			}
		}
		if i-lagPosition >= m.maxIGap {
			base.Top = lagPosition
		}
	}
}

func (m *Merger) clipTrapezoids() {
	for base := m.trapList; base != nil; base = base.Next {
		if base.Top-base.Bottom < m.bottomPadding-2 {
			continue
		}

		aBottom := base.Bottom - base.Right
		aTop := base.Top - base.Left

		lagPosition := aBottom - m.maxIGap + 1
		if lagPosition < 0 {
			lagPosition = 0
		}
		lastPosition := aTop + m.maxIGap
		if lastPosition > m.target.Len() {
			lastPosition = m.target.Len()
		}

		lagClip := aBottom
		i := 0
		for i = lagPosition; i < lastPosition; i++ {
			if m.valueToCode[m.target.Seq[i]] >= 0 {
				if i-lagPosition >= m.maxIGap {
					if lagPosition > lagClip {
						if m.freeTraps == nil {
							m.freeTraps = &Trapezoid{}
						}

						m.freeTraps = m.freeTraps.shunt(base)

						base.clip(lagPosition, lagClip)

						base = base.Next
						m.trapCount++
					}
					lagClip = i
				}
				lagPosition = i + 1
			}
		}

		if i-lagPosition < m.maxIGap {
			lagPosition = aTop
		}

		base.clip(lagPosition, lagClip)

		m.tail = base
	}
}

// Finalise the merged collection and return a sorted slice of Trapezoids.
func (m *Merger) FinaliseMerge() (trapezoids Trapezoids) {
	var next *Trapezoid
	for base := m.trapOrder; base != m.eoTerm; base = next {
		next = base.Next
		m.trapList = base.join(m.trapList)
		m.trapCount++
	}

	m.clipVertical()
	m.clipTrapezoids()

	if m.tail != nil {
		m.freeTraps = m.tail.join(m.freeTraps)
	}

	trapezoids = make(Trapezoids, m.trapCount)
	for i, z := 0, m.trapList; i < m.trapCount; i++ {
		trapezoids[i] = z
		z = z.Next
		trapezoids[i].Next = nil
	}

	sort.Sort(Trapezoids(trapezoids))

	return
}
