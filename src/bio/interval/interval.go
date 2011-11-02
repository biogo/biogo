// Package to find intersections between intervals or sort intervals.
//
// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
// Ported from quicksect.py of bx-python ©James Taylor bitbucket.org/james_taylor/bx-python
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
	"math"
	"rand"
	"bio"
	"bio/util"
)

const negRecLogHalf = 1.4426950408889634073599246810018921374266459541529859341354 // -1/Log(0.5)

// Tree type will store a collection of intervals. While this is not necessary for interval tree
// searching, a Tree stores intervals in a hash based on the chromosome name, reducing search time and easing coding.
type Tree map[string][]*Interval

// Create a new interval Tree.
func NewTree() *Tree {
	t := Tree(make(map[string][]*Interval))
	return &t
}

// Insert an Interval into the Tree.
func (self *Tree) Insert(i *Interval) {
	if c, ok := (*self)[i.chromosome]; ok {
		c = append(c, i)
		r := c[0].Insert(i)
		r, c[0] = c[0], r // bring root to front of slice
		(*self)[i.chromosome] = c
	} else {
		c = make([]*Interval, 0)
		c = append(c, i)
		(*self)[i.chromosome] = c
	}
}

// Find all intervals in Tree that overlap query. Return a channel that will convey results.
func (self *Tree) Intersect(i *Interval, overlap int) (result chan *Interval) {
	if c, ok := (*self)[i.chromosome]; ok {
		result = make(chan *Interval)
		go c[0].Intersect(i, overlap, result)
	}

	return
}

// Traverse all intervals in Tree in order (chromosomes in hash order). Return a channel that will convey results.
func (self *Tree) Traverse() (result chan *Interval) {
	result = make(chan *Interval)
	go func() {
		for _, c := range *self {
			(*c[0]).Traverse(result)
		}
	}()

	return result
}

// Interval type stores start and end of interval and meta data in line and Meta (meta may be used to link to a feat.Feature).
type Interval struct {
	chromosome       string
	start, end, line int
	Meta             interface{}
	priority         float32
	maxEnd, minEnd   int
	left, right      *Interval
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
		maxEnd:     end,
		minEnd:     end,
		priority:   float32(math.Ceil(negRecLogHalf * math.Log(-1/(rand.Float64()-1)))),
		Meta:       meta,
	}, nil
}

func (self *Interval) Chromosome() string { return self.chromosome }

func (self *Interval) Start() int { return self.start }

func (self *Interval) End() int { return self.end }

func (self *Interval) Line() int { return self.line }

// Insert an Interval into a tree returning the new root. Receiver should be the root of the tree.
func (self *Interval) Insert(i *Interval) (root *Interval) {
	root = self

	if i.start >= self.start {
		// insert to right tree
		if self.right != nil {
			self.right = self.right.Insert(i)
		} else {
			self.right = i
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
		}
		// rebalance tree
		if self.priority < self.left.priority {
			root = self.rotateRight()
		}
	}

	if root.right != nil && root.left != nil {
		root.maxEnd = util.Max(root.end, root.right.maxEnd, root.left.maxEnd)
		root.minEnd = util.Min(root.end, root.right.minEnd, root.left.minEnd)
	} else if root.right != nil {
		root.maxEnd = util.Max(root.end, root.right.maxEnd)
		root.minEnd = util.Min(root.end, root.right.minEnd)
	} else if root.left != nil {
		root.maxEnd = util.Max(root.end, root.left.maxEnd)
		root.minEnd = util.Min(root.end, root.left.minEnd)
	}

	return
}

func (self *Interval) rotateRight() (root *Interval) {
	root = self.left
	self.left = self.left.right
	root.right = self
	if self.right != nil && self.left != nil {
		self.maxEnd = util.Max(self.end, self.right.maxEnd, self.left.maxEnd)
		self.minEnd = util.Min(self.end, self.right.minEnd, self.left.minEnd)
	} else if self.right != nil {
		self.maxEnd = util.Max(self.end, self.right.maxEnd)
		self.minEnd = util.Min(self.end, self.right.minEnd)
	} else if self.left != nil {
		self.maxEnd = util.Max(self.end, self.left.maxEnd)
		self.minEnd = util.Max(self.end, self.left.minEnd)
	}

	return
}

func (self *Interval) rotateLeft() (root *Interval) {
	root = self.right
	self.right = self.right.left
	root.left = self
	if self.right != nil && self.left != nil {
		self.maxEnd = util.Max(self.end, self.right.maxEnd, self.left.maxEnd)
		self.minEnd = util.Min(self.end, self.right.minEnd, self.left.minEnd)
	} else if self.right != nil {
		self.maxEnd = util.Max(self.end, self.right.maxEnd)
		self.minEnd = util.Min(self.end, self.right.minEnd)
	} else if self.left != nil {
		self.maxEnd = util.Max(self.end, self.left.maxEnd)
		self.minEnd = util.Max(self.end, self.left.minEnd)
	}

	return
}

// Find Intervals that intersect with the query (search is recursive inorder), and pass results on provided channel.
// The overlap parameter determines how much overlap is required:
//     overlap = 0 intervals abut
//     overlap > 0 intervals must overlap by overlap
//     overlap < 0 intervals can be up to overlap away
func (self *Interval) Intersect(i *Interval, overlap int, r chan<- *Interval) {
	self.intersect(i, overlap, r)
	close(r)
}

func (self *Interval) intersect(i *Interval, overlap int, r chan<- *Interval) {
	if self.left != nil && i.start <= self.left.maxEnd-overlap-1 {
		self.left.intersect(i, overlap, r)
	}
	if i.start <= self.end-overlap-1 && i.end >= self.start+overlap-1 {
		r <- self
	}
	if self.right != nil && i.end >= self.start+overlap-1 {
		self.right.intersect(i, overlap, r)
	}
}

// Traverse all intervals accesible from the current Interval in tree order and pass results on provided channel.
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
