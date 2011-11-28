// Package to find intersections between intervals or sort intervals.
//
// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
// Derived from quicksect.py of bx-python ©James Taylor bitbucket.org/james_taylor/bx-python
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
package interval

import (
	"bio"
	"bio/util"
	"math"
	"math/rand"
)

const negRecLogHalf = 1.4426950408889634073599246810018921374266459541529859341354 // -1/Log(0.5)

// Tree type will store a collection of intervals. While this is not necessary for interval tree
// searching, a Tree stores intervals in a hash based on the chromosome name, reducing search time and easing coding.
type Tree map[string]*Interval

// Create a new interval Tree.
func NewTree() Tree {
	return Tree(make(map[string]*Interval))
}

// Insert an Interval into the Tree.
func (self Tree) Insert(i *Interval) {
	if root, ok := self[i.chromosome]; ok {
		self[i.chromosome] = root.Insert(i)
	} else {
		self[i.chromosome] = i
	}
}

// Merge an Interval into the Tree.
func (self Tree) Merge(i *Interval, overlap int) (replaced []*Interval) {
	if root, ok := self[i.chromosome]; ok {
		var inserted []*Interval
		inserted, replaced = root.merge(i, overlap)
		removed := [][]*Interval{replaced}
		self.replace(inserted, removed)
	} else {
		self[i.chromosome] = i
	}

	return
}

// Remove an interval, returning the removed interval.
func (self Tree) Remove(i *Interval) (removed *Interval) {
	if root, ok := self[i.chromosome]; ok {
		var newRoot *Interval
		if newRoot, removed = i.Remove(); i == root {
			newRoot.parent = nil
			self[i.chromosome] = newRoot
		}
	}

	return
}

func (self Tree) fastRemove(i *Interval) (removed *Interval) {
	// Remove an interval, returning the removed interval. Does not adjust ranges within tree.
	if root, ok := self[i.chromosome]; ok {
		var newRoot *Interval
		if newRoot, removed = i.fastRemove(); i == root {
			newRoot.parent = nil
			self[i.chromosome] = newRoot
		}
	}

	return
}

// Find all intervals in Tree that overlap query. Return a channel that will convey results.
func (self Tree) Intersect(i *Interval, overlap int) (result chan *Interval) {
	if root, ok := self[i.chromosome]; ok {
		result = make(chan *Interval)
		go root.Intersect(i, overlap, result)
	}

	return
}

// Find all intervals in Tree that entirely contain query. Return a channel that will convey results.
func (self Tree) Contain(i *Interval, slop int) (result chan *Interval) {
	if root, ok := self[i.chromosome]; ok {
		result = make(chan *Interval)
		go root.Contain(i, slop, result)
	}

	return
}

// Find all intervals in Tree that are entirely contained by query. Return a channel that will convey results.
func (self Tree) Within(i *Interval, slop int) (result chan *Interval) {
	if root, ok := self[i.chromosome]; ok {
		result = make(chan *Interval)
		go root.Within(i, slop, result)
	}

	return
}

// Traverse all intervals for a chromosome in Tree in order. Return a channel that will convey results.
func (self Tree) Traverse(chromosome string) (result chan *Interval) {
	result = make(chan *Interval)
	go func() {
		self[chromosome].Traverse(result)
	}()

	return result
}

// Traverse all intervals in Tree in order (chromosomes in hash order). Return a channel that will convey results.
func (self Tree) TraverseAll() (result chan *Interval) {
	result = make(chan *Interval)
	go func() {
		for _, c := range self {
			c.Traverse(result)
		}
	}()

	return result
}

// Return the range of the tree's span
func (self Tree) Range(chromosome string) (min, max int) {
	if root, ok := self[chromosome]; ok {
		min, max = root.minStart, root.maxEnd
	}

	return
}

// Flatten a range of intervals intersecting i so that only one interval covers any given location. Intervals
// less than tolerance positions apart are merged into a single new flattened interval.
// Flatting is done by replacement. Return flattened intervals and all intervals originally in intersected region.
// No metadata is transfered to flattened intervals.
func (self Tree) Flatten(i *Interval, overlap, tolerance int) (inserted []*Interval, removed [][]*Interval) {
	if root, ok := self[i.chromosome]; ok {
		inserted, removed = root.flatten(i, overlap, tolerance)
		self.replace(inserted, removed)
	}

	return
}

// Flatten a range of intervals within i so that only one interval covers any given location.
// Flatting is done by replacement. Return flattened intervals and all intervals originally in intersected region.
// No metadata is transfered to flattened intervals.
func (self Tree) FlattenWithin(i *Interval, slop, tolerance int) (inserted []*Interval, removed [][]*Interval) {
	if root, ok := self[i.chromosome]; ok {
		inserted, removed = root.flattenWithin(i, slop, tolerance)
		self.replace(inserted, removed)
	}

	return
}

func (self Tree) replace(inserted []*Interval, removed [][]*Interval) {
	// Helper function for Flatten* and Merge methods. Unsafe for use when replacement intervals do not cover removed intervals.
	// TODO: Check this: Possibly always unsafe - in which case use root.adjustRangeRecursive() between loops.
	for _, section := range removed {
		for _, target := range section {
			self.fastRemove(target)
		}
	}
	for _, replacement := range inserted {
		self.Insert(replacement)
	}
}

// Interval type stores start and end of interval and meta data in line and Meta (meta may be used to link to a feat.Feature).
type Interval struct {
	chromosome          string
	start, end, line    int
	minStart, maxEnd    int
	Meta                interface{}
	priority            int
	left, right, parent *Interval
}

// Create a new Interval.
func New(chrom string, start, end, line int, meta interface{}) (*Interval, error) {
	if end < start {
		return nil, bio.NewError("Interval end < start", 0, []int{start, end})
	}
	return &Interval{
		chromosome: chrom,
		start:      start,
		end:        end,
		line:       line,
		minStart:   start,
		maxEnd:     end,
		priority:   int(negRecLogHalf*math.Log(-1/(rand.Float64()-1))) + 1,
		Meta:       meta,
	}, nil
}

func (self *Interval) Chromosome() string { return self.chromosome }

func (self *Interval) Start() int { return self.start }

func (self *Interval) End() int { return self.end }

func (self *Interval) Line() int { return self.line }

func (self *Interval) adjustRange() {
	if self.left != nil && self.right != nil {
		self.minStart = util.Min(self.start, self.left.minStart, self.right.minStart)
		self.maxEnd = util.Max(self.end, self.left.maxEnd, self.right.maxEnd)
	} else if self.left != nil {
		self.minStart = util.Min(self.start, self.left.minStart)
		self.maxEnd = util.Max(self.end, self.left.maxEnd)
	} else if self.right != nil {
		self.minStart = util.Min(self.start, self.right.minStart)
		self.maxEnd = util.Max(self.end, self.right.maxEnd)
	}
}

func (self *Interval) adjustRangeRecursive() {
	if self.left != nil {
		self.left.adjustRangeRecursive()
	}
	if self.right != nil {
		self.right.adjustRangeRecursive()
	}
	if self.left == nil && self.right == nil {
		self.minStart, self.maxEnd = self.start, self.end
	}
	self.adjustRange()

	return
}

func (self *Interval) adjustRangeParental() {
	if self.parent != nil {
		self.parent.adjustRange()
		self.parent.adjustRangeParental()
	}
}

// Insert an Interval into a tree returning the new root. Receiver should be the root of the tree.
func (self *Interval) Insert(i *Interval) (root *Interval) {
	root = self

	if i.start >= self.start {
		// insert to right tree
		if self.right != nil {
			self.right = self.right.Insert(i)
		} else {
			self.right = i
			i.parent = self
		}
		// rebalance tree
		if self.priority < self.right.priority {
			root = self.rotateLeft()
		}
	} else {
		// insert to left tree
		if self.left != nil {
			self.left = self.left.Insert(i)
		} else {
			self.left = i
			i.parent = self
		}
		// rebalance tree
		if self.priority < self.left.priority {
			root = self.rotateRight()
		}
	}

	root.adjustRange()

	return
}

func (self *Interval) rotateRight() (root *Interval) {
	root = self.left
	root.parent = self.parent
	self.left, self.left.right = self.left.right, self
	self.parent = root
	if self.left != nil {
		self.left.parent = self
	}
	self.adjustRange()

	return
}

func (self *Interval) rotateLeft() (root *Interval) {
	root = self.right
	root.parent = self.parent
	self.right, self.right.left = self.right.left, self
	self.parent = root
	if self.right != nil {
		self.right.parent = self
	}
	self.adjustRange()

	return
}

func (self *Interval) merge(i *Interval, overlap int) (inserted []*Interval, removed []*Interval) {
	r := make(chan *Interval)
	removed = []*Interval{}

	go func() {
		min, max := util.MaxInt, util.MinInt
		for old := range r {
			min, max = util.Min(min, old.start), util.Max(max, old.end)
			removed = append(removed, old)
		}
		n, _ := New("", util.Min(i.start, min), util.Max(i.end, max), 0, nil)
		inserted = []*Interval{n}
		// Do something sensible when only one interval is found and the only action is to extend or ignore
	}()
	self.intersect(i, overlap, r)
	close(r)

	return
}

// Find Intervals that intersect with the query (search is recursive inorder), and pass results on provided channel.
// The overlap parameter determines how much overlap is required:
//     overlap < 0 intervals can be up to overlap away
//     overlap = 0 intervals abut
//     overlap > 0 intervals must overlap by overlap
func (self *Interval) Intersect(i *Interval, overlap int, r chan<- *Interval) {
	self.intersect(i, overlap, r)
	close(r)
}

func (self *Interval) intersect(i *Interval, overlap int, r chan<- *Interval) {
	// Short circuit search for known subtree failure 
	if i.end-overlap < self.minStart || i.start+overlap > self.maxEnd {
		return
	}
	if self.left != nil && i.start < self.left.maxEnd-overlap {
		self.left.intersect(i, overlap, r)
	}
	if i.start < self.end-overlap && i.end > self.start+overlap {
		r <- self
	}
	if self.right != nil && i.end > self.start+overlap {
		self.right.intersect(i, overlap, r)
	}
}

// Find Intervals completely containing the query (search is recursive inorder), and pass results on provided channel.
// The slop parameter determines how much slop is allowed:
//     slop < 0 query must be within interval by slop
//     slop = 0 intervals may completely coincide
//     slop > 0 query may extend beyond interval by slop
func (self *Interval) Contain(i *Interval, slop int, r chan<- *Interval) {
	self.contain(i, slop, r)
	close(r)
}

func (self *Interval) contain(i *Interval, slop int, r chan<- *Interval) {
	// Short circuit search for known subtree failure 
	if i.start+slop < self.minStart || i.end-slop > self.maxEnd {
		return
	}
	if self.left != nil && i.start < self.left.maxEnd-slop {
		self.left.contain(i, slop, r)
	}
	if self.start <= i.start+slop && self.end >= i.end-slop {
		r <- self
	}
	if self.right != nil && i.end > self.start+slop {
		self.right.contain(i, slop, r)
	}
}

// Find Intervals completely within with the query (search is recursive inorder), and pass results on provided channel.
// The slop parameter determines how much slop is allowed:
//     slop < 0 intervals must be within query by slop
//     slop = 0 intervals may completely coincide
//     slop > 0 intervals may extend beyond query by slop
func (self *Interval) Within(i *Interval, slop int, r chan<- *Interval) {
	self.within(i, slop, r)
	close(r)
}

func (self *Interval) within(i *Interval, slop int, r chan<- *Interval) {
	// Short circuit search for known subtree failure 
	if i.start-slop > self.minStart || i.end+slop < self.maxEnd {
		return
	}
	if self.left != nil && i.start < self.left.maxEnd-slop {
		self.left.within(i, slop, r)
	}
	if i.start <= self.start+slop && i.end >= self.end-slop {
		r <- self
	}
	if self.right != nil && i.end > self.start+slop {
		self.right.within(i, slop, r)
	}
}

// Traverse all intervals accessible from the current Interval in tree order and pass results on provided channel.
func (self *Interval) Traverse(r chan<- *Interval) {
	self.traverse(r)
	close(r)
}

func (self *Interval) traverse(r chan<- *Interval) {
	if self.left != nil {
		self.left.traverse(r)
	}
	r <- self
	if self.right != nil {
		self.right.traverse(r)
	}
}

func (self *Interval) unlinkFromParent() *Interval {
	if self.parent == nil {
		return self
	}
	if self.parent.left == self {
		self.parent.left = nil
	} else if self.parent.right == self {
		self.parent.right = nil
	}

	return self
}

func (self *Interval) leftMost() *Interval {
	if self.left != nil {
		return self.left.leftMost()
	}

	return self
}

func (self *Interval) rightMost() *Interval {
	if self.right != nil {
		return self.right.rightMost()
	}

	return self
}

func (self *Interval) flatten(i *Interval, overlap, tolerance int) (inserted []*Interval, removed [][]*Interval) {
	r := make(chan *Interval)
	inserted = []*Interval{}
	removed = [][]*Interval{[]*Interval{}}

	go func() {
		min := util.MaxInt
		j := 0
		var last *Interval
		for old := range r {
			if last != nil && old.start-tolerance > last.end {
				n, _ := New("", min, last.end, 0, nil)
				inserted = append(inserted, n)
				min = old.start
				j++
			} else {
				min = util.Min(min, old.start)
			}
			if len(removed) < j {
				removed = append(removed, []*Interval{})
			}
			removed[j] = append(removed[j], old)
			last = old
		}
	}()
	self.intersect(i, overlap, r)
	close(r)

	return
}

func (self *Interval) flattenWithin(i *Interval, slop, tolerance int) (inserted []*Interval, removed [][]*Interval) {
	r := make(chan *Interval)
	inserted = []*Interval{}
	removed = [][]*Interval{[]*Interval{}}

	go func() {
		min := util.MaxInt
		j := 0
		var last *Interval
		for old := range r {
			if last != nil && old.start-tolerance > last.end {
				n, _ := New("", min, last.end, 0, nil)
				inserted = append(inserted, n)
				min = old.start
				j++
			} else {
				min = util.Min(min, old.start)
			}
			if len(removed) < j {
				removed = append(removed, []*Interval{})
			}
			removed[j] = append(removed[j], old)
			last = old
		}
	}()
	self.within(i, slop, r)
	close(r)

	return
}

// Remove an interval. Returns the new root of the subtree and the removed interval.
func (self *Interval) Remove() (root, removed *Interval) {
	root, removed = self.fastRemove()
	root.adjustRangeRecursive()
	root.adjustRangeParental()

	return
}

func (self *Interval) fastRemove() (root, removed *Interval) {
	// Remove an interval. Returns the new root of the subtree and the removed interval.
	// Does not adjust ranges of descendent and parental nodes.
	root, removed = self.remove()
	removed.left, removed.right, removed.parent = nil, nil, nil
	removed.minStart, removed.maxEnd = removed.start, removed.end

	return
}

func (self *Interval) remove() (root, removed *Interval) {
	removed = self.unlinkFromParent()

	if self.left == nil && self.right == nil {
		return nil, removed
	} else if (self.left == nil) != (self.right == nil) {
		if self.left == nil {
			self.parent.Insert(self.right)
		} else if self.right == nil {
			self.parent.Insert(self.left)
		}
	} else {
		var promotable, descendent *Interval
		if rand.Float64() < 0.5 {
			promotable, descendent = self.left.rightMost(), self.right
		} else {
			promotable, descendent = self.right.leftMost(), self.left
		}
		_, promotable = promotable.remove()
		if self.parent != nil {
			promotable = promotable.Insert(descendent)
			root = self.parent.Insert(promotable)
		} else {
			root = promotable
			root.left, root.right = removed.left, removed.right
			root.priority, removed.priority = removed.priority, root.priority
		}
	}

	return
}

// Return the range of the node's subtree span
func (self *Interval) Range() (int, int) {
	return self.minStart, self.maxEnd
}
