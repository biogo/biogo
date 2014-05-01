// Copyright ©2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pals

import (
	"code.google.com/p/biogo.store/interval"
	"code.google.com/p/biogo/feat"

	"fmt"
	"unsafe"
)

var duplicatePair = fmt.Errorf("pals: attempt to add duplicate feature pair to pile")

// Note Location must be comparable according to http://golang.org/ref/spec#Comparison_operators.
type PileInterval struct {
	Start, End int
	Location   feat.Feature
	Pairs      []*Pair
	overlap    int
}

func (i *PileInterval) Overlap(b interval.IntRange) bool {
	return i.End-i.overlap >= b.Start && i.Start <= b.End-i.overlap
}
func (i *PileInterval) ID() uintptr { return uintptr(unsafe.Pointer(i)) }
func (i *PileInterval) Range() interval.IntRange {
	return interval.IntRange{Start: i.Start + i.overlap, End: i.End - i.overlap}
}

// A Piler performs the aggregation of feature pairs according to the description in section 2.3
// of Edgar and Myers (2005) using an interval tree, giving O(nlogn) time but better space complexity
// and flexibility with feature overlap.
type Piler struct {
	intervals map[feat.Feature]*interval.IntTree
	seen      map[[2]sf]struct{}
	overlap   int
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

	p.merge(&PileInterval{Start: fp.A.Start(), End: fp.A.End(), Location: fp.A.Location(), Pairs: []*Pair{fp}, overlap: p.overlap})
	p.merge(&PileInterval{Start: fp.B.Start(), End: fp.B.End(), Location: fp.B.Location(), Pairs: nil, overlap: p.overlap})
	p.seen[ab] = struct{}{}

	return nil
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
func (p *Piler) merge(pi *PileInterval) {
	var (
		f  = true
		r  []interval.IntInterface
		qi = &PileInterval{Start: pi.Start, End: pi.End}
	)
	t, ok := p.intervals[pi.Location]
	if !ok {
		t = &interval.IntTree{}
		p.intervals[pi.Location] = t
	}
	t.DoMatching(
		func(e interval.IntInterface) (done bool) {
			iv := e.(*PileInterval)
			r = append(r, e)
			pi.Pairs = append(pi.Pairs, iv.Pairs...)
			if f {
				pi.Start = min(iv.Start, pi.Start)
				f = false
			}
			pi.End = max(iv.End, pi.End)
			return
		},
		qi,
	)
	for _, d := range r {
		t.Delete(d, false)
	}
	t.Insert(pi, false)
}

// A PileFilter is used to determine whether a Pair is included in a Pile
type PileFilter func(a, b *Feature, pa, pb *PileInterval) bool

// Piles returns a slice of piles determined by application of the filter function f to
// the feature pairs that have been added to the piler.
func (p *Piler) Piles(f PileFilter) ([]*Pile, error) {
	var (
		pm  = make(map[*PileInterval]*Pile)
		err error
	)
	for _, t := range p.intervals {
		t.Do(
			func(e interval.IntInterface) (done bool) {
				var (
					pa = e.(*PileInterval)
					pb *PileInterval
				)
				for _, pp := range pa.Pairs {
					pb, err = p.pile(pp.B)
					if err != nil {
						return true // Terminate Do() and allow Piles() to return err.
					}

					if f != nil && !f(pp.A, pp.B, pa, pb) {
						continue
					}
					if wp, ok := pm[pa]; !ok {
						tp := &Pile{
							Loc: pa.Location, From: pa.Start, To: pa.End,
							Images: []*Pair{pp},
						}
						pp.A.Loc = tp
						pm[pa] = tp
					} else {
						pp.A.Loc = wp
						wp.Images = append(wp.Images, pp)
					}
					if wp, ok := pm[pb]; !ok {
						tp := &Pile{
							Loc: pb.Location, From: pb.Start, To: pb.End,
							Images: []*Pair{pp.Invert()},
						}
						pp.B.Loc = tp
						pm[pb] = tp
					} else {
						pp.B.Loc = wp
						wp.Images = append(wp.Images, pp.Invert())
					}
				}
				return
			},
		)
	}
	if err != nil {
		return nil, err
	}

	piles := make([]*Pile, 0, len(pm))
	for _, pile := range pm {
		piles = append(piles, pile)
	}

	return piles, nil
}

// Pile returns a Pile representation of the pile containing q.
// An error is returned if more than one pile would be returned.
func (p *Piler) Pile(q *Feature) (*Pile, error) {
	pi, err := p.pile(q)
	if err != nil {
		return nil, err
	}
	return &Pile{
		From:   pi.Start,
		To:     pi.End,
		Loc:    pi.Location,
		Images: pi.Pairs,
	}, nil
}

func (p *Piler) pile(q *Feature) (*PileInterval, error) {
	var (
		qi = query{
			start: q.Start() + p.overlap,
			end:   q.End() - p.overlap,
		}
		t  = p.intervals[q.Location()]
		c  = 0
		pt interval.IntInterface
	)

	findContainers(
		t.Root,
		func(e interval.IntInterface) {
			c++
			pt = e
		},
		qi,
	)

	// Sanity check: no pile should overlap any other pile within overlap constraints
	// TODO: Should this be a panic?
	if c > 1 {
		return nil, fmt.Errorf("pals: internal inconsistency - too many results: %d", c)
	}

	return pt.(*PileInterval), nil
}

type query struct {
	start, end int
}

func (q query) Range() interval.IntRange { return interval.IntRange{Start: q.start, End: q.end} }

type simple struct{ query }

func (q simple) Overlap(b interval.IntRange) bool {
	return q.end > b.Start && q.start < b.End
}

type contained struct{ query }

func (q contained) Overlap(b interval.IntRange) bool {
	return q.start >= b.Start && q.end <= b.End
}

func findContainers(n *interval.IntNode, fn func(interval.IntInterface), q query) {
	if n.Left != nil && (simple{q}.Overlap(n.Left.Range)) {
		findContainers(n.Left, fn, q)
	}
	if (contained{q}.Overlap(n.Interval)) {
		fn(n.Elem)
	}
	if n.Right != nil && (simple{q}.Overlap(n.Right.Range)) {
		findContainers(n.Right, fn, q)
	}
}
