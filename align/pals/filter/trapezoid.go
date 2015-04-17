// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package filter

// Type to store a successfully filtered w × e Parallelogram
type Trapezoid struct {
	Next        *Trapezoid // Organized in a list linked on this field
	Top, Bottom int        // B (query) coords of top and bottom of trapzoidal zone
	Left, Right int        // Left and right diagonals of trapzoidal zone
}

// prependFrontTo prepends a copy of front of list onto the front of the list
// taking the allocation from the receiver which acts as a pool. It returns the
// next element of the pool.
func (tr *Trapezoid) prependFrontTo(list *Trapezoid) *Trapezoid {
	next := tr.Next
	*tr = *list
	list.join(tr)
	return next
}

// join joins list to the receiver, returning the receiver.
func (tr *Trapezoid) join(list *Trapezoid) *Trapezoid {
	tr.Next = list
	return tr
}

// Return the receiver and the subsequent Trapezoid in the list.
func (tr *Trapezoid) decapitate() (*Trapezoid, *Trapezoid) {
	return tr, tr.Next
}

// clip is the trapezoid trimming method used during merge.
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

// Trapezoids implements the sort.Sort interface.
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
