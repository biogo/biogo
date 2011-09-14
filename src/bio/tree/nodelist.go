package tree
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
	"container/vector"
	"os"
)

type NodeList struct {
	nodeList vector.Vector
	nodeMap  map[string]bool // use number of children? instead of bool
}

func (self NodeList) Len() int {
	return self.nodeList.Len()
}

func (self NodeList) Less(i, j int) bool {
	return self.nodeList[i].(*Node).Name < self.nodeList[j].(*Node).Name
}

func (self NodeList) Swap(i, j int) {
	ni := self.nodeList.At(i).(*Node).Name
	nj := self.nodeList.At(j).(*Node).Name
	self.nodeMap[ni], self.nodeMap[nj] = self.nodeMap[nj], self.nodeMap[ni]
	self.nodeList.Swap(i, j)
}

func (self *NodeList) Pop() (n *Node) {
	n = self.nodeList.Pop().(*Node)
	self.nodeMap[n.Name] = false, false
	return
}

func (self *NodeList) Push(n *Node) (err os.Error) {
	if _, present := self.nodeMap[n.Name]; !present {
		self.nodeList.Push(n)
		self.nodeMap[n.Name] = true
	} else {
		err = os.NewError("Cannot push non-unique nodes onto NodeList")
	}

	return
}

func (self *NodeList) At(i int) (n *Node) {
	n = self.nodeList.At(i).(*Node)
	return
}

// Extend this so that can call vector.Vector provided function 
