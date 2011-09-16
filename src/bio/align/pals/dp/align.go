package dp
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
	"os"
	"bio/seq"
	"sort"
	"sync"
	"bio"
	"bio/align/pals/filter"
)

const debug debugging = false

type Aligner struct {
	a, b         *seq.Seq
	k            int
	minHitLength int
	minId        float64
	comp         bool
	threads      int
}

func NewAligner(a, b *seq.Seq, k, minLength int, minId float64, comp bool, threads int) *Aligner {
	return &Aligner{
		a:            a,
		b:            b,
		k:            k,
		minHitLength: minLength,
		minId:        minId,
		comp:         comp,
		threads:      threads,
	}
}

func (self *Aligner) AlignTraps(trapezoids []*filter.Trapezoid) (segSols []DPHit) {

	numSegs := 0
	fseg := 0

	minLen := self.minHitLength
	maxDiff := 1 - self.minId

	covered := make([]bool, len(trapezoids)) // all false

	sort.Sort(&traps{trapezoids})

	dp := make(chan *DP, self.threads)
	result := make(chan DPHit)

	for i := 0; i < self.threads; i++ {
		dp <- &DP{
			a:       self.a,
			b:       self.b,
			minLen:  minLen,
			maxDiff: maxDiff,
			comp:    self.comp,
			covered: covered,
			result:  result,
		}
	}

	go func() {
		for {
			segSols = append(segSols, <-result)
		}
	}()

	wg := &sync.WaitGroup{}
	for i, z := range trapezoids {
		if !covered[i] {
			if z.Top-z.Bot < self.k {
				continue
			}

			wg.Add(1)
			go func(p *DP, i int) {
				defer func() {
					dp <- p
					wg.Done()
				}()
				p.trapezoids = trapezoids[i+1:]
				p.slot = i

				p.alignRecursion(trapezoids[i])
			}(<-dp, i)
		}
	}

	//wait for alignments to finish
	wg.Wait()
	numSegs = len(segSols)

	/* Remove lower scoring segments that begin or end at
	   the same point as a higher scoring segment.       */
	sort.Sort(&starts{segSols[fseg:]})

	var j int

	for i := fseg; i < numSegs; i = j {
		for j = i + 1; j < numSegs; j++ {
			if segSols[j].Abpos != segSols[i].Abpos {
				break
			}
			if segSols[j].Bbpos != segSols[i].Bbpos {
				break
			}
			if segSols[j].Score > segSols[i].Score {
				segSols[i].Score = -1
				i = j
			} else {
				segSols[j].Score = -1
			}
		}
	}

	sort.Sort(&ends{segSols[fseg:]})

	for i := fseg; i < numSegs; i = j {
		for j = i + 1; j < numSegs; j++ {
			if segSols[j].Aepos != segSols[i].Aepos {
				break
			}
			if segSols[j].Bepos != segSols[i].Bepos {
				break
			}
			if segSols[j].Score > segSols[i].Score {
				segSols[i].Score = -1
				i = j
			} else {
				segSols[j].Score = -1
			}
		}
	}
	for i := fseg; i < numSegs; i++ {
		if segSols[i].Score >= 0 {
			segSols[fseg] = segSols[i]
			fseg++
		}
	}
	segSols = segSols[:fseg]
	trapezoids = nil
	return segSols
}

func SumDPLengths(hits []DPHit) (sum int, e os.Error) {
	for _, hit := range hits {
		length := hit.Aepos - hit.Abpos
		if length < 0 {
			return 0, bio.NewError("Length < 0", 0, hit)
		}
		sum += length
	}
	return sum, nil
}
