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
package dp

import (
	"bio/seq"
	"bio/align/pals/filter"
)

const (
	forward = iota
	reverse
)

type DP struct {
	a, b       *seq.Seq
	minLen     int
	maxDiff    float64
	dp         [2]DPHit
	vec        []int
	comp       bool
	trapezoids []*filter.Trapezoid
	covered    []bool
	slot       int
	result     chan DPHit
}

func (self *DP) alignRecursion(z *filter.Trapezoid) {
	var (
		mid, indel int
		percent    float64
	)

	debug.Printf("A [%d,%d]x[%d,%d] %v\n", z.Bot, z.Top, z.Lft, z.Rgt, !self.comp)
	mid = (z.Bot + z.Top) / 2

	self.traceForwardPath(mid, mid-z.Rgt, mid-z.Lft)
	for x := 1; true; x++ {
		self.traceReversePath(self.dp[forward].Bepos, self.dp[forward].Aepos, self.dp[forward].Aepos,
			mid+MaxIGap, BlockCost+2*x*DiffCost)
		if !(self.dp[reverse].Bbpos > mid+x*MaxIGap && self.dp[reverse].Score < self.dp[forward].Score) {
			break
		}
	}

	hit := self.dp[reverse]

	hit.Aepos = self.dp[forward].Aepos
	hit.Bepos = self.dp[forward].Bepos
	ltrp := *z
	htrp := *z
	ltrp.Top = hit.Bbpos - MaxIGap
	htrp.Bot = hit.Bepos + MaxIGap
	debug.Println("-0")
	if hit.Bepos-hit.Bbpos >= self.minLen && hit.Aepos-hit.Abpos >= self.minLen {
		debug.Println("--1")
		indel = (hit.Abpos - hit.Bbpos) - (hit.Aepos - hit.Bepos)
		if indel < 0 {
			indel = -indel
		}
		percent = (1 / RMatchCost) - float64(hit.Score-indel)/(RMatchCost*float64(hit.Bepos-hit.Bbpos))
		if percent <= self.maxDiff {
			debug.Println("---2")
			hit.Error = percent
			var ta, tb, ua, ub int
			debug.Println("length", len(self.trapezoids))
			for j, t := range self.trapezoids {
				debug.Println("start", t.Top, t.Bot, t.Lft, t.Rgt)
				if t.Bot >= hit.Bepos {
					debug.Println("broke", t.Bot, hit.Bepos)
					break
				}

				tb = t.Top - t.Bot + 1
				ta = t.Rgt - t.Lft + 1

				if t.Lft < hit.Ldiag {
					ua = hit.Ldiag
				} else {
					ua = t.Lft
				}

				if t.Rgt > hit.Hdiag {
					ub = hit.Hdiag
				} else {
					ub = t.Rgt
				}

				if ua > ub {
					debug.Println("continued", ua, ub)
					continue
				}

				ua = ub - ua + 1

				if t.Top > hit.Bepos {
					ub = hit.Bepos - t.Bot + 1
				} else {
					ub = tb
				}

				if ((float64(ua))/float64(ta))*((float64(ub))/float64(tb)) > .99 {
					self.covered[j+self.slot] = true
				}
			}

			d := hit.Ldiag // diags to this point are b-a, not a-b
			hit.Ldiag = -(hit.Hdiag)
			hit.Hdiag = -d

			self.result <- hit
		}
	}

	if ltrp.Top-ltrp.Bot > self.minLen && ltrp.Top < z.Top-MaxIGap {
		self.alignRecursion(&ltrp)
	}

	if htrp.Top-htrp.Bot > self.minLen {
		self.alignRecursion(&htrp)
	}
	debug.Printf("  Hit from (%d,%d) to (%d,%d) within [%d,%d] score %d %v\n",
		hit.Abpos, hit.Bbpos, hit.Aepos, hit.Bepos,
		hit.Ldiag, hit.Hdiag, hit.Score, !self.comp)
}
