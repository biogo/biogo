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

// Package providing PALS dynamic programming alignment routines.
package dp

import (
	"code.google.com/p/biogo/align/pals/filter"
	"code.google.com/p/biogo/bio"
	"code.google.com/p/biogo/seq"
	"sort"
)

// Holds dp alignment parameters.
type Params struct {
	MinHitLength int
	MinId        float64
}

// An AlignConfig specifies dynamic programming behaviour.
//
// Sensible default parameters for alignment:
//	MaxIGap    = 5
//	DiffCost   = 3
//	SameCost   = 1
//	MatchCost  = DiffCost + SameCost
//	BlockCost  = DiffCost * MaxIGap
//	RMatchCost = DiffCost + 1
type AlignConfig struct {
	MaxIGap    int
	DiffCost   int
	SameCost   int
	MatchCost  int
	BlockCost  int
	RMatchCost float64
}

// An Aligner provides allows local alignment of subsections of long sequences.
type Aligner struct {
	target, query *seq.Seq
	k             int
	minHitLength  int
	minId         float64
	segs          DPHits
	Config        *AlignConfig
}

// Create a new Aligner based on target and query sequences. 
func NewAligner(target, query *seq.Seq, k, minLength int, minId float64) *Aligner {
	return &Aligner{
		target:       target,
		query:        query,
		k:            k,
		minHitLength: minLength,
		minId:        minId,
	}
}

// Align pairs of sequence segments defined by trapezoids.
// Returns aligning segment pairs satisfying length and identity requirements.
func (a *Aligner) AlignTraps(trapezoids filter.Trapezoids) DPHits {
	covered := make([]bool, len(trapezoids))

	dp := &kernel{
		target:     a.target,
		query:      a.query,
		trapezoids: trapezoids,
		covered:    covered,
		minLen:     a.minHitLength,
		maxDiff:    1 - a.minId,

		maxIGap:    a.Config.MaxIGap,
		diffCost:   a.Config.DiffCost,
		sameCost:   a.Config.SameCost,
		matchCost:  a.Config.MatchCost,
		blockCost:  a.Config.BlockCost,
		rMatchCost: a.Config.RMatchCost,

		result: make(chan DPHit),
	}
	w := make(chan struct{})
	var segs DPHits
	go func() {
		defer close(w)
		for h := range dp.result {
			segs = append(segs, h)
		}
	}()
	for i, t := range trapezoids {
		if !dp.covered[i] && t.Top-t.Bottom >= a.k {
			dp.slot = i
			dp.alignRecursion(t)
		}
	}
	close(dp.result)
	<-w

	/* Remove lower scoring segments that begin or end at
	   the same point as a higher scoring segment.       */

	if len(segs) > 0 {
		var i, j int

		sort.Sort(starts(segs))
		for i = 0; i < len(segs); i = j {
			for j = i + 1; j < len(segs); j++ {
				if segs[j].Abpos != segs[i].Abpos {
					break
				}
				if segs[j].Bbpos != segs[i].Bbpos {
					break
				}
				if segs[j].Score > segs[i].Score {
					segs[i].Score = -1
					i = j
				} else {
					segs[j].Score = -1
				}
			}
		}

		sort.Sort(ends(segs))
		for i = 0; i < len(segs); i = j {
			for j = i + 1; j < len(segs); j++ {
				if segs[j].Aepos != segs[i].Aepos {
					break
				}
				if segs[j].Bepos != segs[i].Bepos {
					break
				}
				if segs[j].Score > segs[i].Score {
					segs[i].Score = -1
					i = j
				} else {
					segs[j].Score = -1
				}
			}
		}

		found := 0
		for i = 0; i < len(segs); i++ {
			if segs[i].Score >= 0 {
				segs[found] = segs[i]
				found++
			}
		}
		segs = segs[:found]
	}

	return segs
}

// DPHit holds details of alignment result. 
type DPHit struct {
	Abpos, Bbpos              int     // Start coordinate of local alignment
	Aepos, Bepos              int     // End coordinate of local alignment
	LowDiagonal, HighDiagonal int     // Alignment is between (anti)diagonals LowDiagonal & HighDiagonal
	Score                     int     // Score of alignment where match = SameCost, difference = -DiffCost
	Error                     float64 // Lower bound on error rate of match
}

// DPHits is a collection of alignment results.
type DPHits []DPHit

// Returns the sums of alignment lengths.
func (h DPHits) Sum() (a, b int, err error) {
	for _, hit := range h {
		la, lb := hit.Aepos-hit.Abpos, hit.Bepos-hit.Bbpos
		if la < 0 || lb < 0 {
			return 0, 0, bio.NewError("Area < 0", 0, hit)
		}
		a, b = a+la, b+lb
	}
	return
}
