package filter

// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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

// Type to store a successfully filtered w × e Parallelogram
type Trapezoid struct {
	Next        *Trapezoid // Organized in a list linked on this field
	Top, Bottom int        // B (query) coords of top and bottom of trapzoidal zone
	Left, Right int        // Left and right diagonals of trapzoidal zone
}

// Move the receiver from the head of the current list to follow element.
// Returns the subsequent Trapezoid of the current list.
func (self *Trapezoid) shunt(element *Trapezoid) (head *Trapezoid) {
	head = self.Next
	*self = *element
	element.join(self)
	return
}

// Joing list to the receiver, returning the reciever.
func (self *Trapezoid) join(list *Trapezoid) *Trapezoid {
	self.Next = list
	return self
}

// Return the receiver and the subsequent Trapezoid in the list.
func (self *Trapezoid) decapitate() (*Trapezoid, *Trapezoid) {
	return self, self.Next
}

// Trapezoid timming method used during merge.
func (self *Trapezoid) clip(lagPosition, lagClip int) {
	if bottom := lagClip + self.Left; self.Bottom < bottom {
		self.Bottom = bottom
	}
	if top := lagPosition + self.Right; self.Top > top {
		self.Top = top
	}
	midPosition := (self.Bottom + self.Top) / 2
	if left := midPosition - lagPosition; self.Left < left {
		self.Left = left
	}
	if right := midPosition - lagClip; self.Right > right {
		self.Right = right
	}
}

// Trapezoid slice type used for sorting Trapezoids during merge.
type Trapezoids []*Trapezoid

// Return the sum of all Trapezoids in the slice.
func (self Trapezoids) Sum() (a, b int) {
	for _, t := range self {
		la, lb := t.Top-t.Bottom, t.Right-t.Left
		a, b = a+la, b+lb
	}
	return
}

// Required for sort.Interface
func (self Trapezoids) Len() int {
	return len(self)
}

// Required for sort.Interface
func (self Trapezoids) Less(i, j int) bool {
	return self[i].Bottom < self[j].Bottom
}

// Required for sort.Interface
func (self Trapezoids) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}
