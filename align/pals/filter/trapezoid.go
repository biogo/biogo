// Copyright ©2011-2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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

package filter

// Type to store a successfully filtered w × e Parallelogram
type Trapezoid struct {
	Next        *Trapezoid // Organized in a list linked on this field
	Top, Bottom int        // B (query) coords of top and bottom of trapzoidal zone
	Left, Right int        // Left and right diagonals of trapzoidal zone
}

// Move the receiver from the head of the current list to follow element.
// Returns the subsequent Trapezoid of the current list.
func (tr *Trapezoid) shunt(element *Trapezoid) (head *Trapezoid) {
	head = tr.Next
	*tr = *element
	element.join(tr)
	return
}

// Joing list to the receiver, returning the reciever.
func (tr *Trapezoid) join(list *Trapezoid) *Trapezoid {
	tr.Next = list
	return tr
}

// Return the receiver and the subsequent Trapezoid in the list.
func (tr *Trapezoid) decapitate() (*Trapezoid, *Trapezoid) {
	return tr, tr.Next
}

// Trapezoid timming method used during merge.
func (tr *Trapezoid) clip(lagPosition, lagClip int) {
	if bottom := lagClip + tr.Left; tr.Bottom < bottom {
		tr.Bottom = bottom
	}
	if top := lagPosition + tr.Right; tr.Top > top {
		tr.Top = top
	}
	midPosition := (tr.Bottom + tr.Top) / 2
	if left := midPosition - lagPosition; tr.Left < left {
		tr.Left = left
	}
	if right := midPosition - lagClip; tr.Right > right {
		tr.Right = right
	}
}

// Trapezoid slice type used for sorting Trapezoids during merge.
type Trapezoids []*Trapezoid

// Return the sum of all Trapezoids in the slice.
func (tr Trapezoids) Sum() (a, b int) {
	for _, t := range tr {
		la, lb := t.Top-t.Bottom, t.Right-t.Left
		a, b = a+la, b+lb
	}
	return
}

// Required for sort.Interface
func (tr Trapezoids) Len() int {
	return len(tr)
}

// Required for sort.Interface
func (tr Trapezoids) Less(i, j int) bool {
	return tr[i].Bottom < tr[j].Bottom
}

// Required for sort.Interface
func (tr Trapezoids) Swap(i, j int) {
	tr[i], tr[j] = tr[j], tr[i]
}
