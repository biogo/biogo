// Copyright ©2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pals

import (
	"fmt"

	"github.com/biogo/biogo/feat"
	"github.com/biogo/store/interval"
)

var duplicatePair = fmt.Errorf("pals: attempt to add duplicate feature pair to pile")

// Note Location must be comparable according to http://golang.org/ref/spec#Comparison_operators.
type pileInterval struct {
	id         uintptr
	start, end int
	pile       *Pile
	location   feat.Feature
	images     []*Feature
	overlap    int
}

func (i *pileInterval) Overlap(b interval.IntRange) bool {
	return i.end-i.overlap >= b.Start && i.start <= b.End-i.overlap
}
func (i *pileInterval) ID() uintptr { return i.id }
func (i *pileInterval) Range() interval.IntRange {
	return interval.IntRange{Start: i.start + i.overlap, End: i.end - i.overlap}
}

// A Piler performs the aggregation of feature pairs according to the description in section 2.3
// of Edgar and Myers (2005) using an interval tree, giving O(nlogn) time but better space complexity
// and flexibility with feature overlap.
type Piler struct {
	intervals map[feat.Feature]*interval.IntTree
	seen      map[[2]sf]struct{}
	overlap   int
	piled     bool

	// next provides the next ID
	// for merged intervals. IDs
	// are unique across all intervals.
	next uintptr
}

type sf struct {
	loc  feat.Feature
	s, e int
}

// NewPiler creates a Piler object ready for piling feature pairs.
func NewPiler(overlap int) *Piler {
	return &Piler{
		intervals: make(map[feat.Feature]*interval.IntTree),
		seen:      make(map[[2]sf]struct{}),
		overlap:   overlap,
	}
}

// Add adds a feature pair to the piler incorporating the features into piles where appropriate.
func (p *Piler) Add(fp *Pair) error {
	a := sf{loc: fp.A.Location(), s: fp.A.Start(), e: fp.A.End()}
	b := sf{loc: fp.B.Location(), s: fp.B.Start(), e: fp.B.End()}
	ab, ba := [2]sf{a, b}, [2]sf{b, a}

	if _, ok := p.seen[ab]; ok {
		return duplicatePair
	}
	if _, ok := p.seen[ba]; ok {
		return duplicatePair
	}

	p.merge(&pileInterval{id: p.nextID(), start: fp.A.Start(), end: fp.A.End(), location: fp.A.Location(), images: []*Feature{fp.A}, overlap: p.overlap})
	p.merge(&pileInterval{id: p.nextID(), start: fp.B.Start(), end: fp.B.End(), location: fp.B.Location(), images: []*Feature{fp.B}, overlap: p.overlap})
	p.seen[ab] = struct{}{}

	return nil
}

func (p *Piler) nextID() uintptr {
	id := p.next
	p.next++
	return id
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// merge merges an interval into the tree moving location meta data from the replaced intervals
// into the new interval.
func (p *Piler) merge(pi *pileInterval) {
	var (
		f  = true
		r  []interval.IntInterface
		qi = &pileInterval{start: pi.start, end: pi.end}
	)
	t, ok := p.intervals[pi.location]
	if !ok {
		t = &interval.IntTree{}
		p.intervals[pi.location] = t
	}
	t.DoMatching(
		func(e interval.IntInterface) (done bool) {
			r = append(r, e)
			iv := e.(*pileInterval)
			pi.images = append(pi.images, iv.images...)
			if f {
				pi.start = min(iv.start, pi.start)
				f = false
			}
			pi.end = max(iv.end, pi.end)
			return
		},
		qi,
	)
	for _, d := range r {
		t.Delete(d, false)
	}
	t.Insert(pi, false)
}

// A PairFilter is used to determine whether a Pair's images are included in a Pile.
type PairFilter func(*Pair) bool

// Piles returns a slice of piles determined by application of the filter function f to
// the feature pairs that have been added to the piler. Piles may be called more than once,
// but the piles returned in earlier invocations will be altered by subsequent calls.
func (p *Piler) Piles(f PairFilter) []*Pile {
	if !p.piled {
		for _, t := range p.intervals {
			t.Do(func(e interval.IntInterface) (done bool) {
				pa := e.(*pileInterval)
				if pa.pile == nil {
					pa.pile = &Pile{Loc: pa.location, From: pa.start, To: pa.end}
				}
				for _, im := range pa.images {
					if checkSanity {
						assertPileSanity(t, im, pa.pile)
					}
					im.Loc = pa.pile
				}
				return
			})
		}
		p.piled = true
	}

	var piles []*Pile
	for _, t := range p.intervals {
		t.Do(func(e interval.IntInterface) (done bool) {
			pa := e.(*pileInterval)
			pa.pile.Images = pa.pile.Images[:0]
			for _, im := range pa.images {
				if f != nil && !f(im.Pair) {
					continue
				}
				if checkSanity {
					assertPairSanity(im)
				}
				pa.pile.Images = append(pa.pile.Images, im)
			}
			piles = append(piles, pa.pile)
			return
		})
	}

	return piles
}

const checkSanity = false

func assertPileSanity(t *interval.IntTree, im *Feature, pi *Pile) {
	if im.Start() < pi.Start() || im.End() > pi.End() {
		panic(fmt.Sprintf("image extends beyond pile: %#v", im))
	}
	if foundPiles := t.Get(&pileInterval{start: im.Start(), end: im.End()}); len(foundPiles) > 1 {
		var containing int
		for _, pile := range foundPiles {
			r := pile.Range()
			if (r.Start <= im.Start() && r.End > im.End()) || (r.Start < im.Start() && r.End >= im.End()) {
				containing++
			}
		}
		if containing > 1 {
			panic(fmt.Sprintf("found too many piles for %#v", im))
		}
	}
}

func assertPairSanity(im *Feature) {
	if _, ok := im.Loc.(*Pile); !ok {
		panic(fmt.Sprintf("image not allocated to pile %#v", im))
	}
	if _, ok := im.Mate().Loc.(*Pile); !ok {
		panic(fmt.Sprintf("image mate not allocated to pile %#v", im.Mate()))
	}
}
