// Package to find intersections between intervals or sort intervals.
package interval

// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
// Derived from quicksect.py of bx-python ©James Taylor bitbucket.org/james_taylor/bx-python
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
	"fmt"
	"github.com/kortschak/BioGo/bio"
	"github.com/kortschak/BioGo/util"
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
func (self Tree) Merge(i *Interval, overlap int) (inserted, replaced []*Interval) {
	if root, ok := self[i.chromosome]; ok {
		inserted, replaced = root.merge(i, overlap)
		removed := [][]*Interval{replaced}
		self.replace(inserted, removed)
	} else {
		self[i.chromosome] = i
	}

	return
}

// Remove an interval, returning the removed interval with all pointers set to nil.
func (self Tree) Remove(i *Interval) (removed *Interval) {
	if root, ok := self[i.chromosome]; ok {
		var newRoot *Interval
		if newRoot, removed = i.Remove(); i == root {
			if newRoot != nil {
				newRoot.parent = nil
				self[i.chromosome] = newRoot
			} else {
				delete(self, i.chromosome)
			}
		}
	}

	return
}

// Remove an interval, returning the removed interval. Does not adjust ranges within tree.
func (self Tree) FastRemove(i *Interval) (removed *Interval) {
	if root, ok := self[i.chromosome]; ok {
		var newRoot *Interval
		if newRoot, removed = i.fastRemove(); i == root {
			if newRoot != nil {
				newRoot.parent = nil
				self[i.chromosome] = newRoot
			} else {
				delete(self, i.chromosome)
			}
		}
	}

	return
}

func (self Tree) AdjustRange(chromosome string) {
	self[chromosome].adjustRangeRecursive()
}

// Find all intervals in Tree that overlap query. Return a channel that will convey results.
func (self Tree) Intersect(i *Interval, overlap int) (result chan *Interval) {
	result = make(chan *Interval)
	if root, ok := self[i.chromosome]; ok {
		go root.Intersect(i, overlap, result)
	} else {
		close(result)
	}

	return
}

// Find all intervals in Tree that entirely contain query. Return a channel that will convey results.
func (self Tree) Contain(i *Interval, slop int) (result chan *Interval) {
	result = make(chan *Interval)
	if root, ok := self[i.chromosome]; ok {
		go root.Contain(i, slop, result)
	} else {
		close(result)
	}

	return
}

// Find all intervals in Tree that are entirely contained by query. Return a channel that will convey results.
func (self Tree) Within(i *Interval, slop int) (result chan *Interval) {
	result = make(chan *Interval)
	if root, ok := self[i.chromosome]; ok {
		go root.Within(i, slop, result)
	} else {
		close(result)
	}

	return
}

// Traverse all intervals for a chromosome in Tree in order. Return a channel that will convey results.
func (self Tree) Traverse(chromosome string) (result chan *Interval) {
	result = make(chan *Interval)
	if t := self[chromosome]; t != nil {
		go func() {
			t.Traverse(result)
		}()
	} else {
		close(result)
	}

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
// Return flattened intervals and all intervals originally in intersected region.
// No metadata is transfered to flattened intervals.
func (self Tree) Flatten(i *Interval, overlap, tolerance int) (flat []*Interval, rich [][]*Interval) {
	if root, ok := self[i.chromosome]; ok {
		r := make(chan *Interval)
		go root.Intersect(i, overlap, r)
		flat, rich = root.Flatten(r, tolerance)
	}

	return
}

// Flatten a range of intervals containing i so that only one interval covers any given location.
// Return flattened intervals and all intervals originally in containing region.
// No metadata is transfered to flattened intervals.
func (self Tree) FlattenContaining(i *Interval, slop, tolerance int) (flat []*Interval, rich [][]*Interval) {
	if root, ok := self[i.chromosome]; ok {
		r := make(chan *Interval)
		go root.Contain(i, slop, r)
		flat, rich = root.Flatten(r, tolerance)
	}

	return
}

// Flatten a range of intervals within i so that only one interval covers any given location.
// Return flattened intervals and all intervals originally in contained region.
// No metadata is transfered to flattened intervals.
func (self Tree) FlattenWithin(i *Interval, slop, tolerance int) (flat []*Interval, rich [][]*Interval) {
	if root, ok := self[i.chromosome]; ok {
		r := make(chan *Interval)
		go root.Within(i, slop, r)
		flat, rich = root.Flatten(r, tolerance)
	}

	return
}

func (self Tree) replace(inserted []*Interval, removed [][]*Interval) {
	// Helper function for Merge method. Unsafe for use when replacement intervals do not cover removed intervals.
	for _, section := range removed {
		for _, target := range section {
			self.FastRemove(target)
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
		return nil, bio.NewError("Interval end < start", 0, start, end)
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

// Return the chromosome identifier for an Interval node.
func (self *Interval) Chromosome() string { return self.chromosome }

// Return the start position of an Interval node.
func (self *Interval) Start() int { return self.start }

// Return the end position of an Interval node.
func (self *Interval) End() int { return self.end }

// Return the line number of an Interval node - not used except for reference to file.
func (self *Interval) Line() int { return self.line }

// Return a pointer to the parent of an Interval node.
func (self *Interval) Parent() *Interval { return self.parent }

// Return a pointer to the left child of an Interval node.
func (self *Interval) Left() *Interval { return self.left }

// Return a pointer to the right child of an Interval node.
func (self *Interval) Right() *Interval { return self.right }

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
	for n := self.parent; n != nil; n = n.parent {
		n.adjustRange()
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

func (self *Interval) merge(i *Interval, overlap int) (inserted, removed []*Interval) {
	r := make(chan *Interval)
	removed = []*Interval{}
	wait := make(chan struct{})

	go func() {
		defer close(wait)
		min, max := util.MaxInt, util.MinInt
		for old := range r {
			min, max = util.Min(min, old.start), util.Max(max, old.end)
			removed = append(removed, old)
		}
		n, _ := New(i.chromosome, util.Min(i.start, min), util.Max(i.end, max), 0, nil)
		inserted = []*Interval{n}
		// Do something sensible when only one interval is found and the only action is to extend or ignore
	}()
	self.intersect(i, overlap, r)
	close(r)
	<-wait

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
	if self.left != nil && i.start <= self.left.maxEnd-overlap {
		self.left.intersect(i, overlap, r)
	}
	if i.start <= self.end-overlap && i.end >= self.start+overlap {
		r <- self
	}
	if self.right != nil && i.end >= self.start+overlap {
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
	if i.end-slop < self.minStart || i.start+slop > self.maxEnd {
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

// Return the previous interval in tree traverse order.
func (self *Interval) ScanLeft() (n *Interval) {
	if self.left != nil {
		return self.left.RightMost()
	}

	if self.parent == nil {
		return
	}

	for n = self; ; n = n.parent {
		if n.parent == nil {
			n = nil
			break
		} else if n.parent.right == n {
			n = n.parent
			break
		}
	}

	return
}

// Return the next interval in tree traverse order.
func (self *Interval) ScanRight() (n *Interval) {
	if self.right != nil {
		return self.right.LeftMost()
	}

	if self.parent == nil {
		return
	}

	for n = self; ; n = n.parent {
		if n.parent == nil {
			n = nil
			break
		} else if n.parent.left == n {
			n = n.parent
			break
		}
	}

	return
}

func (self *Interval) LeftMost() (n *Interval) {
	for n = self; n.left != nil; n = n.left {
	}

	return
}

func (self *Interval) RightMost() (n *Interval) {
	for n = self; n.right != nil; n = n.right {
	}

	return
}

// Merge a range of intervals provided by r. Returns merged intervals in a slice and
// intervals contributing to merged intervals groups in a slice of slices.
func (self *Interval) Flatten(r chan *Interval, tolerance int) (flat []*Interval, rich [][]*Interval) {
	flat = []*Interval{}
	rich = [][]*Interval{{}}

	min, max := util.MaxInt, util.MinInt
	var last *Interval
	for current := range r {
		if last != nil && current.start-tolerance > max {
			n, _ := New(current.chromosome, min, max, 0, nil)
			flat = append(flat, n)
			min = current.start
			max = current.end
			rich = append(rich, []*Interval{})
		} else {
			min = util.Min(min, current.start)
			max = util.Max(max, current.end)
		}
		rich[len(rich)-1] = append(rich[len(rich)-1], current)
		last = current
	}
	n, _ := New(last.chromosome, min, max, 0, nil)
	flat = append(flat, n)

	return
}

// Remove an interval. Returns the new root of the subtree and the removed interval.
func (self *Interval) Remove() (root, removed *Interval) {
	parent := self.parent
	root, removed = self.fastRemove()
	if root != nil {
		root.adjustRangeRecursive()
		root.adjustRangeParental()
	} else if parent != nil {
		parent.adjustRangeRecursive()
		parent.adjustRangeParental()
	}

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

func (self *Interval) changeLinkFromParentTo(l *Interval) {
	if self.parent != nil {
		if self.parent.left == self {
			self.parent.left = l
		} else if self.parent.right == self {
			self.parent.right = l
		}
	}
}

func (self *Interval) skip() (replacement *Interval) {
	switch {
	case self.left != nil && self.right != nil:
		panic("cannot skip a fully connected node")
	case self.left != nil:
		replacement = self.left
		replacement.parent = self.parent
	case self.right != nil:
		replacement = self.right
		replacement.parent = self.parent
	default:
		replacement = nil
	}
	self.changeLinkFromParentTo(replacement)

	return
}

func (self *Interval) remove() (root, removed *Interval) {
	removed = self

	switch {
	case self.left != nil && self.right != nil:
		if rand.Float64() < 0.5 {
			root = self.left.RightMost()
		} else {
			root = self.right.LeftMost()
		}
		root.skip()
		root.parent = self.parent
		root.left = self.left
		if root.left != nil {
			root.left.parent = root
		}
		root.right = self.right
		if root.right != nil {
			root.right.parent = root
		}
		root.priority, self.priority = self.priority, root.priority
		if self.parent != nil {
			if self.parent.left == self {
				self.parent.left = root
			} else {
				self.parent.right = root
			}
		}
	case self.left == nil && self.right == nil:
		self.changeLinkFromParentTo(nil)
	default:
		root = self.skip()
	}

	return
}

// Return the range of the node's subtree span
func (self *Interval) Range() (int, int) {
	return self.minStart, self.maxEnd
}

// Default function for String method
var StringFunc = defaultStringFunc

func defaultStringFunc(i *Interval) string {
	return fmt.Sprintf("%q:[%d, %d)", i.chromosome, i.start, i.end)
}

// String method
func (self *Interval) String() string {
	return StringFunc(self)
}
