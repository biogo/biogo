// Package providing PALS dynamic programming alignment routines.
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
	"github.com/kortschak/BioGo/bio"
	"github.com/kortschak/BioGo/seq"
	"sort"
)

type Aligner struct {
	target, query *seq.Seq
	k             int
	minHitLength  int
	minId         float64
	threads       int
	segments      []DPHit
}

func NewAligner(target, query *seq.Seq, k, minLength int, minId float64, threads int) *Aligner {
	return &Aligner{
		target:       target,
		query:        query,
		k:            k,
		minHitLength: minLength,
		minId:        minId,
		threads:      threads,
	}
}

func (self *Aligner) AlignTraps(trapezoids []*filter.Trapezoid) (segments []DPHit) {
	covered := make([]bool, len(trapezoids))

	dp := &kernel{
		target:     self.target,
		query:      self.query,
		trapezoids: trapezoids,
		covered:    covered,
		minLen:     self.minHitLength,
		maxDiff:    1 - self.minId,
	}
	for i, t := range trapezoids {
		if !dp.covered[i] && t.Top-t.Bottom >= self.k {
			dp.slot = i
			dp.alignRecursion(t)
		}
	}
	segments = make([]DPHit, len(dp.segments))
	copy(segments, dp.segments)

	/* Remove lower scoring segments that begin or end at
	   the same point as a higher scoring segment.       */

	if len(segments) > 0 {
		var i, j int

		sort.Sort(starts(segments))
		for i = 0; i < len(segments); i = j {
			for j = i + 1; j < len(segments); j++ {
				if segments[j].Abpos != segments[i].Abpos {
					break
				}
				if segments[j].Bbpos != segments[i].Bbpos {
					break
				}
				if segments[j].Score > segments[i].Score {
					segments[i].Score = -1
					i = j
				} else {
					segments[j].Score = -1
				}
			}
		}

		sort.Sort(ends(segments))
		for i = 0; i < len(segments); i = j {
			for j = i + 1; j < len(segments); j++ {
				if segments[j].Aepos != segments[i].Aepos {
					break
				}
				if segments[j].Bepos != segments[i].Bepos {
					break
				}
				if segments[j].Score > segments[i].Score {
					segments[i].Score = -1
					i = j
				} else {
					segments[j].Score = -1
				}
			}
		}

		found := 0
		for i = 0; i < len(segments); i++ {
			if segments[i].Score >= 0 {
				segments[found] = segments[i]
				found++
			}
		}
		segments = segments[:found]
	}

	return
}

func SumDPLengths(hits []DPHit) (sum int, e error) {
	for _, hit := range hits {
		length := hit.Aepos - hit.Abpos
		if length < 0 {
			return 0, bio.NewError("Length < 0", 0, hit)
		}
		sum += length
	}
	return sum, nil
}
