package tree
// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
// Derived from PyCogent tree package Copyright ©2007-2011, The Cogent Project, under GPL2 or greater
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
	"container/vector"
)

const (
	PreOrder = iota
	PostOrder
	PrePostOrder
	LevelOrder
)

type Node struct {
	Name       string
	support    float32
	length     float32
	parameters map[string]float32
	children   NodeList
	parent     *Node
	tree       *Tree
}

func NewNode(name string, support, length float32) *Node {
	return &Node{
		Name:    name,
		support: support,
		length:  length,
	}
}

func (self *Node) AddNode(n *Node) {
	self.children.Push(n)
	n.parent = self
	n.tree = self.tree
	self.tree.nodes.Push(n)
	for _, c := range n.children.nodeList {
		self.tree.nodes.Push(c.(*Node))
	}
}

func (self *Node) NodeIterator(order byte, includeSelf bool) (c chan *Node, quit chan bool) {
	c = make(chan *Node)
	quit = make(chan bool)
	switch order {
	case PreOrder:
		self.PreOrder(includeSelf, c, quit)
	case PostOrder:
		self.PostOrder(includeSelf, c, quit)
	case PrePostOrder:
		self.PrePostOrder(includeSelf, c, quit)
	case LevelOrder:
		self.LevelOrder(includeSelf, c, quit)
	}
	return
}

func (self *Node) PreOrder(includeSelf bool, c chan *Node, quit chan bool) {
	go func() {
		defer func() {
			close(c)
		}()

		var (
			i    int
			this *Node
		)

		stack := vector.Vector{}
		stack.Insert(0, self)
		for len(stack) > 0 {
			i = len(stack) - 1
			this = stack[i].(*Node)
			stack = stack[:i]
			if this == self || includeSelf {
				select {
				case c <- this:
				case <-quit:
					return
				}
			}
			if this.children.Len() > 0 {
				stack.InsertVector(len(stack), &this.children.nodeList)
			}
		}
	}()
}

func (self *Node) PostOrder(includeSelf bool, c chan *Node, quit chan bool) {
	go func() {
		defer func() {
			close(c)
		}()

		var (
			index       int
			this, child *Node
		)

		childIndex := vector.IntVector{}
		childIndex.Push(0)
		this = self
		for {
			index = childIndex[len(childIndex)-1]
			if index < len(this.children.nodeList) {
				child = this.children.nodeList[index].(*Node)
				if len(child.children.nodeList) > 0 {
					childIndex.Push(0)
					this = child
					index = 0
				} else {
					select {
					case c <- child:
					case <-quit:
						return
					}
					childIndex[len(childIndex)-1]++
				}
			} else {
				if includeSelf || this != self {
					select {
					case c <- this:
					case <-quit:
						return
					}
				}
				if this == self {
					break
				}
				this = this.parent
				childIndex = childIndex[:len(childIndex)-1]
				childIndex[len(childIndex)-1]++
			}
		}
	}()
}

func (self *Node) PrePostOrder(includeSelf bool, c chan *Node, quit chan bool) {
	go func() {
		defer func() {
			close(c)
		}()

		if self.children.Len() < 1 {
			if includeSelf {
				select {
				case c <- self:
				case <-quit:
					return
				}
			}
		} else {
			var (
				this, child *Node
				i, index    int
			)

			childIndex := vector.IntVector{}
			childIndex.Push(0)
			this = self
			for {
				index = childIndex.Last()
				if index < 1 {
					if this != self || includeSelf {
						select {
						case c <- self:
						case <-quit:
							return
						}
					}
				}
				if index < this.children.nodeList.Len() {
					child = this.children.nodeList[index].(*Node)
					if len(child.children.nodeList) > 0 {
						childIndex.Push(0)
						this = child
						index = 0
					} else {
						select {
						case c <- child:
						case <-quit:
							return
						}
						childIndex[len(childIndex)-1]++
					}
				} else {
					if includeSelf || this != self {
						select {
						case c <- this:
						case <-quit:
							return
						}
					}
					if this == self {
						break
					}
					this = this.parent
					i = len(childIndex) - 1
					childIndex = childIndex[:i]
					childIndex[i]++
				}
			}
		}
	}()
}

func (self *Node) LevelOrder(includeSelf bool, c chan *Node, quit chan bool) {
	go func() {
		defer func() {
			close(c)
		}()

		var (
			queue vector.Vector
			this/*, child*/ *Node
			child interface{} // until container vector removed this is necessary
		)

		queue.Push(self)

		for queue.Len() > 0 {
			this = queue[0].(*Node)
			queue.Delete(0)
			if this != self || includeSelf {
				select {
				case c <- this:
				case <-quit:
					return
				}
			}
			if this.children.nodeList.Len() > 0 {
				for _, child = range this.children.nodeList {
					queue.Push(child)
				}
			}
		}
	}()
}

func (self *Node) Nodes(order byte, includeSelf bool) (nodes NodeList) {
	iterator, _ := self.NodeIterator(order, includeSelf)
	for n := range iterator {
		nodes.Push(n)
	}
	return
}

func (self *Node) InternalNodeIterator(includeSelf bool) (c chan *Node, quit chan bool) {
	c = make(chan *Node)
	quit = make(chan bool)
	go func() {
		defer func() {
			close(c)
		}()

		iterator, q := self.NodeIterator(PreOrder, includeSelf)
		for n := range iterator {
			if n.children.Len() > 0 {
				select {
				case c <- n:
				case <-quit:
					close(q)
					return
				}
			}
		}
	}()
	return
}

func (self *Node) InternalNodes(includeSelf bool) (internals NodeList) {
	iterator, _ := self.LeafIterator(includeSelf)
	for n := range iterator {
		internals.Push(n)
	}
	return
}

func (self *Node) LeafIterator(includeSelf bool) (c chan *Node, quit chan bool) {
	c = make(chan *Node)
	quit = make(chan bool)
	go func() {
		defer func() {
			close(c)
		}()

		if self.children.Len() < 1 {
			if includeSelf {
				select {
				case c <- self:
				case <-quit:
					return
				}
			}
		} else {
			var (
				i    int
				this *Node
			)

			stack := vector.Vector{}
			stack.Push(self)
			for len(stack) > 0 {
				i = len(stack) - 1
				this = stack[i].(*Node)
				stack = stack[:i]
				if len(this.children.nodeList) > 0 {
					stack.InsertVector(len(stack), &this.children.nodeList)
				} else {
					select {
					case c <- this:
					case <-quit:
						return
					}
				}
			}
		}
	}()
	return
}

func (self *Node) Leaves(includeSelf bool) (leaves NodeList) {
	iterator, _ := self.LeafIterator(includeSelf)
	for n := range iterator {
		leaves.Push(n)
	}
	return
}
