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
)

const (
	dPadding = 2
)

const debug debugging = false

type Merger struct { // these are static so should be part of object Merger
	target, query       *seq.Seq
	filterParams        *Params
	lPadding, bPadding  int
	binWidth            int
	selfComparison      bool
	freeTraps, trapList *Trapezoid
	trapOrder, tailEnd  *Trapezoid
	eoTerm              *Trapezoid
	trapCount           int
	//report sizes
	trapArea int
}

func NewMerger(index *kmerindex.Index, query *seq.Seq, filterParams *Params) (m *Merger) {
	tubeWidth := filterParams.TubeOffset + filterParams.MaxError
	binWidth := tubeWidth - 1
	lPadding := dPadding + binWidth

	eoTerm := &Trapezoid{
		Lft:  query.Len() + 1 + lPadding,
		Rgt:  query.Len() + 1,
		Bot:  -1,
		Top:  query.Len() + 1,
		Next: nil,
	}

	m = &Merger{
		target:       index.Seq,
		filterParams: filterParams,
		query:        query,
		bPadding:     index.GetK() + 2,
		lPadding:     lPadding,
		binWidth:     binWidth,
		freeTraps:    nil,
		eoTerm:       eoTerm,
		trapOrder:    eoTerm,
		trapList:     nil,
	}

	if index.Seq == query {
		m.selfComparison = true
	}

	return m
}

func (self *Merger) MergeFilterHit(hit *FilterHit) {
	defer func() {
		// test trap
		debug.Printf("  Blist:")
		for b := self.trapOrder; b != nil; b = b.Next {
			debug.Printf(" [%d,%d]x[%d,%d]", b.Bot, b.Top, b.Lft, b.Rgt)
		}
		debug.Println()
	}()

	nd := -hit.DiagIndex
	if self.selfComparison && nd <= self.filterParams.MaxError {
		return
	}

	top := hit.QFrom - self.bPadding

	// test trap
	debug.Printf("  Diag %d [%d,%d]\n", nd, hit.QFrom, hit.QTo)

	var t, f *Trapezoid
	// for b in traporder
	for b := self.trapOrder; ; b = t {
		t = b.Next
		switch {
		case b.Top < top:
			self.trapCount++

			// report area
			self.trapArea += (b.Top - b.Bot + 1) * (b.Rgt - b.Lft + 1)

			if f == nil {
				self.trapOrder = t
			} else {
				f.Next = t
			}
			b.Next = self.trapList
			self.trapList = b
		case nd > b.Rgt+dPadding:
			f = b
		case nd >= b.Lft-self.lPadding:
			if nd+self.binWidth > b.Rgt {
				b.Rgt = nd + self.binWidth
			}
			if nd < b.Lft {
				b.Lft = nd
			}
			if hit.QTo > b.Top {
				b.Top = hit.QTo
			}
			switch {
			case f != nil && f.Rgt+dPadding >= b.Lft:
				f.Rgt = b.Rgt
				if f.Bot > b.Bot {
					f.Bot = b.Bot
				}
				if f.Top < b.Top {
					f.Top = b.Top
				}
				f.Next = t
				b.Next = self.freeTraps
				self.freeTraps = b
			case t != nil && t.Lft-dPadding <= b.Rgt:
				b.Rgt = t.Rgt
				if b.Bot > t.Bot {
					b.Bot = t.Bot
				}
				if b.Top < t.Top {
					b.Top = t.Top
				}
				b.Next = t.Next
				t.Next = self.freeTraps
				self.freeTraps = t
				t = b.Next
				// f = b
			default:
				// f = b
			}
			return
		default: // Add to free_trap list
			if self.freeTraps == nil {
				self.freeTraps = &Trapezoid{}
			}
			if f == nil {
				self.trapOrder = self.freeTraps
				f = self.freeTraps
			} else {
				f.Next = self.freeTraps
				f = self.freeTraps
			}
			self.freeTraps = f.Next
			f.Next = b
			f.Top = hit.QTo
			f.Bot = hit.QFrom
			f.Lft = nd
			f.Rgt = f.Lft + self.binWidth
			// f = b
			return
		}
	}
}

func (self *Merger) FinaliseMerge() []*Trapezoid {
	var (
		i int
		t *Trapezoid
	)

	for b := self.trapOrder; b != self.eoTerm; b = t {
		t = b.Next
		self.trapCount++

		// report sizes
		self.trapArea += (b.Top - b.Bot + 1) * (b.Rgt - b.Lft + 1)

		b.Next = self.trapList
		self.trapList = b
	}

	// report sizes
	debug.Printf("\n  %9d trapezoids of area %d (%f%% of matrix)\n",
		self.trapCount, self.trapArea,
		(100*float32(self.trapCount)/float32(self.target.Len()))/float32(self.query.Len()))
	// trim trap
	debug.Println("SeqQ trimming:")

	for b := self.trapList; b != nil; b = b.Next {
		lag := b.Bot - MaxIGap + 1
		if lag < 0 {
			lag = 0
		}
		lst := b.Top + MaxIGap
		if lst > self.query.Len() {
			lst = self.query.Len()
		}

		// trim trap
		debug.Printf("   [%d,%d]x[%d,%d] = %d\n", b.Bot, b.Top, b.Lft, b.Rgt, b.Top-b.Bot+1)
		for i = lag; i < lst; i++ {
			if lookUp.ValueToCode[self.query.Seq[i]] >= 0 {
				if i-lag >= MaxIGap {
					if lag-b.Bot > 0 {
						if self.freeTraps == nil {
							self.freeTraps = &Trapezoid{}
						}
						t = self.freeTraps.Next
						*self.freeTraps = *b
						b.Next = self.freeTraps
						self.freeTraps = t
						b.Top = lag
						b = b.Next
						b.Bot = i
						self.trapCount++
					} else {
						b.Bot = i
					}

					// trim trap
					debug.Printf("  Cut trap SeqQ[%d,%d]\n", lag, i)
				}
				lag = i + 1
			}
		}
		if i-lag >= MaxIGap {
			b.Top = lag
		}
	}

	// trim trap
	debug.Println("SeqT trimming:")

	self.tailEnd = nil
	for b := self.trapList; b != nil; b = b.Next {
		if b.Top-b.Bot < self.bPadding-2 {
			continue
		}
		aBot := b.Bot - b.Rgt
		aTop := b.Top - b.Lft

		// trim trap
		debug.Printf("   [%d,%d]x[%d,%d] = %d\n",
			b.Bot, b.Top, b.Lft, b.Rgt, b.Top-b.Bot+1)
		lag := aBot - MaxIGap + 1
		if lag < 0 {
			lag = 0
		}
		lst := aTop + MaxIGap
		if lst > self.target.Len() {
			lst = self.target.Len()
		}

		lclip := aBot
		for i = lag; i < lst; i++ {
			if lookUp.ValueToCode[self.target.Seq[i]] >= 0 {
				if i-lag >= MaxIGap {
					if lag > lclip {
						if self.freeTraps == nil {
							self.freeTraps = &Trapezoid{}
						}
						t = self.freeTraps.Next
						*self.freeTraps = *b
						b.Next = self.freeTraps
						self.freeTraps = t

						// trim trap
						debug.Printf("     Clip to %d,%d\n", lclip, lag)

						{
							x := lclip + b.Lft
							if b.Bot < x {
								b.Bot = x
							}
							x = lag + b.Rgt
							if b.Top > x {
								b.Top = x
							}
							m := (b.Bot + b.Top) / 2
							x = m - lag
							if b.Lft < x {
								b.Lft = x
							}
							x = m - lclip
							if b.Rgt > x {
								b.Rgt = x
							}

							// trim trap
							debug.Printf("        [%d,%d]x[%d,%d] = %d\n",
								b.Bot, b.Top, b.Lft, b.Rgt, b.Top-b.Bot+1)
						}
						b = b.Next
						self.trapCount++
					}
					lclip = i
				}
				lag = i + 1
			}
		}
		if i-lag < MaxIGap {
			lag = aTop
		}

		// trim trap
		debug.Printf("     Clip to %d,%d\n", lclip, lag)

		{
			x := lclip + b.Lft
			if b.Bot < x {
				b.Bot = x
			}
			x = lag + b.Rgt
			if b.Top > x {
				b.Top = x
			}
			m := (b.Bot + b.Top) / 2
			x = m - lag
			if b.Lft < x {
				b.Lft = x
			}
			x = m - lclip
			if b.Rgt > x {
				b.Rgt = x
			}

			// trim trap
			debug.Printf("        [%d,%d]x[%d,%d] = %d\n",
				b.Bot, b.Top, b.Lft, b.Rgt, b.Top-b.Bot+1)
		}
		self.tailEnd = b
	}

	if self.tailEnd != nil {
		self.tailEnd.Next = self.freeTraps
		self.freeTraps = self.trapList
	}

	// report sizes
	debug.Printf("  %9d trimmed trap.s of area %d (%f%% of matrix)\n",
		self.trapCount, self.trapArea,
		(100*float64(self.trapCount)/float64(self.target.Len()))/float64(self.query.Len()))

	// Excerpt of code from dp.align.go
	trapezoids := make([]*Trapezoid, self.trapCount)

	for i, z := 0, self.trapList; i < len(trapezoids); i++ {
		trapezoids[i] = z
		z = z.Next
		trapezoids[i].Next = nil
	}

	return trapezoids
}

func SumTrapLengths(trapezoids []*Trapezoid) (sum int) {
	for _, t := range trapezoids {
		length := t.Top - t.Bot
		sum += length
	}
	return sum
}
