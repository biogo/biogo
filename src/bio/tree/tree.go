// Copyright ©2011 Dan Kortschak
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
package tree

import (
	"errors"
	"math"
	"sort"
)

type Tree struct { // struct for saving associated metadata to avoid computation
	Name          string
	root          *Node // nominal root - basically just a handle for the tree data
	nodes         NodeList
	matrix        *Matrix
	treeAltered   bool
	matrixAltered bool
}

// Tree functions

func New(name string) *Tree {
	n := NewNode("root", 1.0, 0.0)
	return &Tree{
		Name:          name,
		root:          n,
		matrix:        nil,
		treeAltered:   true,
		matrixAltered: false,
	}
}

func (self *Tree) Root(root *Node) *Node {
	if root != nil {
		self.root = root
		self.treeAltered = true
		self.matrixAltered = false // we accept that if you are doing this you are prepared to clobber the matrix
	}
	return self.root
}

func (self *Tree) Matrix(matrix *Matrix) (m *Matrix, e error) {
	if matrix != nil {
		self.matrix = matrix
		self.matrixAltered = true
		self.treeAltered = false // we accept that if you are doing this you are prepared to clobber the tree
	}
	if self.treeAltered { // only get up-to-date matrix
		e = self.Reconcile()
	}
	return self.matrix, e
}

func (self *Tree) Reconcile() error {
	switch {
	case self.treeAltered && self.matrixAltered: // can't know what to do - this is bad
		return errors.New("Both tree and matrix altered: cannot reconcile")
	case !self.treeAltered && self.matrixAltered: // matrix has been altered so update tree
		self.matrixAltered = false
		return self.treeFromMatrix()
	case self.treeAltered && !self.matrixAltered: // tree has been altered so update matrix
		self.treeAltered = false
		self.matrixFromTree()
		return nil
	default: // why are we even here? nothing has changed
		return nil
	}
	panic("cannot reach")
}

func (self *Tree) treeFromMatrix() (err error) {
	// generate a tree from matrix
	return
}

func (self *Tree) matrixFromTree() {
	// assume the nodelist is OK - it isn't (ignore for the moment)

	var nodes, leaf, internal NodeList
	copy(nodes.nodeList, self.nodes.nodeList)
	sort.Sort(nodes)
	for _, v := range self.nodes.nodeList {
		if v.children.Len() > 0 {
			internal.nodeList = append(internal.nodeList, v)
		} else {
			leaf.nodeList = append(leaf.nodeList, v)
		}
	}

	result := make([][]float32, nodes.Len())
	for i, _ := range result {
		result[i] = make([]float32, internal.Len())
	}

	for i, node := range internal.nodeList {
		decendents := node.Leaves(false)
		for j, d := range leaf.nodeList {
			if decendents.nodeMap[d.Name] {
				result[i][j] = d.length
			} else {
				result[i][j] = float32(math.NaN())
			}
		}
	}

	self.matrix.edgeLength = result
	self.matrix.nodes = leaf
	self.matrix.internal = internal
}
