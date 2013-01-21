// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package providing PALS dynamic programming alignment routines.
package dp

import (
	"code.google.com/p/biogo/align/pals/filter"
	"code.google.com/p/biogo/seq/linear"

	"errors"
	"sort"
)

// A Params holds dynamic programming alignment parameters.
type Params struct {
	MinHitLength int
	MinId        float64
}

// A Costs specifies dynamic programming behaviour.
type Costs struct {
	MaxIGap    int
	DiffCost   int
	SameCost   int
	MatchCost  int
	BlockCost  int
	RMatchCost float64
}

// An Aligner provides allows local alignment of subsections of long sequences.
type Aligner struct {
	target, query *linear.Seq
	k             int
	minHitLength  int
	minId         float64
	segs          DPHits
	Costs         *Costs
}

// Create a new Aligner based on target and query sequences. 
func NewAligner(target, query *linear.Seq, k, minLength int, minId float64) *Aligner {
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
		target:      a.target,
		query:       a.query,
		valueToCode: a.target.Alpha.LetterIndex(),
		trapezoids:  trapezoids,
		covered:     covered,
		minLen:      a.minHitLength,
		maxDiff:     1 - a.minId,

		Costs: *a.Costs,

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
			return 0, 0, errors.New("dp: negative trapezoid area")
		}
		a, b = a+la, b+lb
	}
	return
}
