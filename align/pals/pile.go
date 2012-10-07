// Copyright ©2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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

package pals

import (
	"code.google.com/p/biogo.interval"
	"code.google.com/p/biogo/feat"
	"fmt"
	"unsafe"
)

var duplicatePair = fmt.Errorf("pals: attempt to add duplicate feature pair to pile")

// A Piler performs the aggregation of feature pairs according to the description in section 2.3
// of Edgar and Myers (2005) using an interval tree, giving O(nlogn) time but better space complexity
// and flexibility with feature overlap.
type Piler struct {
	intervals map[string]*interval.IntTree
	seen      map[sp]struct{}
	overlap   int
}

type (
	sf struct {
		loc  string
		s, e int
	}

	sp struct {
		a, b sf
	}
)

type PileInterval struct {
	Start, End int
	Location   string
	Pairs      []*FeaturePair
	overlap    int
}

func (i *PileInterval) Overlap(b interval.IntRange) bool {
	return i.End-i.overlap >= b.Start && i.Start <= b.End-i.overlap
}
func (i *PileInterval) ID() uintptr              { return uintptr(unsafe.Pointer(i)) }
func (i *PileInterval) Range() interval.IntRange { return interval.IntRange{i.Start, i.End} }

type ContainQuery struct {
	Start, End int
	Slop       int
	Location   string
}

func (q *ContainQuery) Overlap(b interval.IntRange) bool {
	return b.Start <= q.Start+q.Slop && b.End >= q.End-q.Slop
}
func (q *ContainQuery) ID() uintptr              { return 0 }
func (q *ContainQuery) Range() interval.IntRange { return interval.IntRange{q.Start, q.End} }

// NewPiler creates a Piler object ready for piling feature pairs.
func NewPiler(overlap int) *Piler {
	return &Piler{
		intervals: make(map[string]*interval.IntTree),
		seen:      make(map[sp]struct{}),
		overlap:   overlap,
	}
}

// Add adds a feature pair to the piler incorporating the features into piles where appropriate.
func (p *Piler) Add(fp *FeaturePair) error {
	a := sf{fp.A.Location, fp.A.Start, fp.A.End}
	b := sf{fp.B.Location, fp.B.Start, fp.B.End}
	ab, ba := sp{a, b}, sp{b, a}

	if _, ok := p.seen[ab]; ok {
		return duplicatePair
	}
	if _, ok := p.seen[ba]; ok {
		return duplicatePair
	}

	p.merge(&PileInterval{fp.A.Start, fp.A.End, string(fp.A.Location), []*FeaturePair{fp}, p.overlap})
	p.merge(&PileInterval{fp.B.Start, fp.B.End, string(fp.B.Location), nil, p.overlap})
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

// merge an interval into the tree adding meta data from the replaced intervals into the new interval
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

// A Pile is a collection of features covering a maximal (potentially contiguous, depending on
// the value of overlap used for creation of the Piler) region of copy count > 0.
type Pile struct {
	Pile   *feat.Feature
	Images []*FeaturePair
}

// A PileFilter is used to determine whether a FeaturePair is included in a Pile
type PileFilter func(a, b *feat.Feature, pa, pb *PileInterval) bool

// We use the Features' Meta field to point back to the containing Pile, so Meta cannot be used for other things here.
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
					pb, err = p.Pile(
						&ContainQuery{
							Start:    pp.B.Start,
							End:      pp.B.End,
							Location: pp.B.Location,
							Slop:     p.overlap,
						},
					)
					if err != nil {
						return true // Terminate Do() and allow Piles() to return err.
					}

					if f != nil && !f(pp.A, pp.B, pa, pb) {
						continue
					}
					if wp, ok := pm[pa]; !ok {
						tp := &Pile{
							Pile:   &feat.Feature{Location: pa.Location, Start: pa.Start, End: pa.End},
							Images: []*FeaturePair{pp},
						}
						pp.A.Meta = tp
						pm[pa] = tp
					} else {
						pp.A.Meta = wp
						wp.Images = append(wp.Images, pp)
					}
					if wp, ok := pm[pb]; !ok {
						tp := &Pile{
							Pile:   &feat.Feature{Location: pb.Location, Start: pb.Start, End: pb.End},
							Images: []*FeaturePair{pp.Invert()},
						}
						pp.B.Meta = tp
						pm[pb] = tp
					} else {
						pp.B.Meta = wp
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

// Pile returns the interval representation of the pile containing i.
// An error is returned if more than one pile would be returned.
func (p *Piler) Pile(q *ContainQuery) (*PileInterval, error) {
	var (
		t  = p.intervals[q.Location]
		c  = 0
		pt interval.IntInterface
	)

	t.DoMatching(
		func(e interval.IntInterface) (done bool) {
			c++
			pt = e
			return
		},
		q,
	)

	// Sanity check: no pile should overlap any other pile within overlap constraints
	// TODO: Should this be a panic?
	if c > 1 {
		return nil, fmt.Errorf("pals: internal inconsistency - too many results:", c)
	}

	return pt.(*PileInterval), nil
}
