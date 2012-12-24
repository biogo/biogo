// Copyright ©2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pals

import (
	"code.google.com/p/biogo.interval"
	"code.google.com/p/biogo/exp/feat"
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
func (i *PileInterval) ID() uintptr              { return uintptr(unsafe.Pointer(i)) }
func (i *PileInterval) Range() interval.IntRange { return interval.IntRange{i.Start, i.End} }

type containQuery struct {
	start, end int
	slop       int
	location   feat.Feature
}

func (q containQuery) Overlap(b interval.IntRange) bool {
	return b.Start <= q.start+q.slop && b.End >= q.end-q.slop
}
func (q containQuery) ID() uintptr              { return 0 }
func (q containQuery) Range() interval.IntRange { return interval.IntRange{q.start, q.end} }

// A Piler performs the aggregation of feature pairs according to the description in section 2.3
// of Edgar and Myers (2005) using an interval tree, giving O(nlogn) time but better space complexity
// and flexibility with feature overlap.
type Piler struct {
	intervals map[feat.Feature]*interval.IntTree
	seen      map[sp]struct{}
	overlap   int
}

type (
	sf struct {
		loc  feat.Feature
		s, e int
	}

	sp struct {
		a, b sf
	}
)

// NewPiler creates a Piler object ready for piling feature pairs.
func NewPiler(overlap int) *Piler {
	return &Piler{
		intervals: make(map[feat.Feature]*interval.IntTree),
		seen:      make(map[sp]struct{}),
		overlap:   overlap,
	}
}

// Add adds a feature pair to the piler incorporating the features into piles where appropriate.
func (p *Piler) Add(fp *Pair) error {
	a := sf{fp.A.Location(), fp.A.Start(), fp.A.End()}
	b := sf{fp.B.Location(), fp.B.Start(), fp.B.End()}
	ab, ba := sp{a, b}, sp{b, a}

	if _, ok := p.seen[ab]; ok {
		return duplicatePair
	}
	if _, ok := p.seen[ba]; ok {
		return duplicatePair
	}

	p.merge(&PileInterval{fp.A.Start(), fp.A.End(), fp.A.Location(), []*Pair{fp}, p.overlap})
	p.merge(&PileInterval{fp.B.Start(), fp.B.End(), fp.B.Location(), nil, p.overlap})
	p.seen[ab] = struct{}{}
	p.seen[ba] = struct{}{}

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

// merge merges an interval into the tree mioving location meta data from the replaced intervals
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
type PileFilter func(a, b feat.Feature, pa, pb *PileInterval) bool

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
						pp.A.(*Feature).Loc = tp
						pm[pa] = tp
					} else {
						pp.A.(*Feature).Loc = wp
						wp.Images = append(wp.Images, pp)
					}
					if wp, ok := pm[pb]; !ok {
						tp := &Pile{
							Loc: pb.Location, From: pb.Start, To: pb.End,
							Images: []*Pair{pp.Invert()},
						}
						pp.B.(*Feature).Loc = tp
						pm[pb] = tp
					} else {
						pp.B.(*Feature).Loc = wp
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

// Pile returns a Pile representation of the pile containing i.
// An error is returned if more than one pile would be returned.
func (p *Piler) Pile(q feat.Feature) (*Pile, error) {
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

func (p *Piler) pile(q feat.Feature) (*PileInterval, error) {
	var (
		qi = containQuery{
			start:    q.Start(),
			end:      q.End(),
			location: q.Location(),
			slop:     p.overlap,
		}
		t  = p.intervals[qi.location]
		c  = 0
		pt interval.IntInterface
	)

	t.DoMatching(
		func(e interval.IntInterface) (done bool) {
			c++
			pt = e
			return
		},
		qi,
	)

	// Sanity check: no pile should overlap any other pile within overlap constraints
	// TODO: Should this be a panic?
	if c > 1 {
		return nil, fmt.Errorf("pals: internal inconsistency - too many results:", c)
	}

	return pt.(*PileInterval), nil
}
