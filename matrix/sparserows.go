// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package matrix

import (
	"math"
)

type sparseRow []sparseElem

type sparseElem struct {
	index int
	value float64
}

var GrowFraction int = 10

func (r sparseRow) at(col int) float64 {
	lo := 0
	hi := len(r)
	for {
		switch curpos := (lo + hi) / 2; {
		case len(r) == 0, col > r[len(r)-1].index, lo > hi:
			return 0
		case r[curpos].index == col:
			return r[curpos].value
		case col < r[curpos].index:
			hi = curpos - 1
		case col > r[curpos].index:
			lo = curpos + 1
		}
	}

	panic("cannot reach")
}

func (r *sparseRow) insert(pos int, val sparseElem) {
	switch length := len(*r); {
	case pos >= length:
		t := make(sparseRow, pos, pos+pos/GrowFraction+1)
		copy(t, (*r))
		t = append(t, val)
		(*r) = t
	case length == cap(*r):
		t := make(sparseRow, length+1, length+length/GrowFraction+1)
		copy(t, (*r)[:pos])
		copy(t[pos+1:], (*r)[pos:])
		*r = t
		(*r)[pos] = val
	case pos < length:
		*r = append((*r)[:pos+1], (*r)[pos:]...)
		(*r)[pos] = val
	}
}

func (r sparseRow) sum() float64 {
	var s float64
	for _, e := range r {
		s += e.value
	}

	return s
}

func (r sparseRow) min() float64 {
	if len(r) == 0 {
		return 0
	}
	n := math.MaxFloat64
	for _, e := range r {
		n = math.Min(n, e.value)
	}

	return n
}

func (r sparseRow) max() float64 {
	if len(r) == 0 {
		return 0
	}
	n := -math.MaxFloat64
	for _, e := range r {
		n = math.Max(n, e.value)
	}

	return n
}

func (r sparseRow) minNonZero() (float64, bool) {
	if len(r) == 0 {
		return 0, false
	}
	n := math.MaxFloat64
	var ok bool
	for _, e := range r {
		if e.value != 0 {
			n = math.Min(n, e.value)
			ok = true
		}
	}

	return n, ok
}

func (r sparseRow) maxNonZero() (float64, bool) {
	if len(r) == 0 {
		return 0, false
	}
	n := -math.MaxFloat64
	var ok bool
	for _, e := range r {
		if e.value != 0 {
			n = math.Max(n, e.value)
			ok = true
		}
	}

	return n, ok
}

func (r sparseRow) scale(beta float64) sparseRow {
	b := make(sparseRow, 0, len(r))
	for _, e := range r {
		b = append(b, sparseElem{index: e.index, value: e.value * beta})
	}

	return b
}

func (r sparseRow) foldAdd(a sparseRow) sparseRow {
	if len(r) > 0 && len(a) > 0 {
		// safely determine no merging to do
		switch {
		case r[len(r)-1].index < a[0].index:
			t := append(r, a...)
			return t
		case a[len(a)-1].index < r[0].index:
			t := append(a, r...)
			return t
		}
	}

	t := <-workbuffers          // get a buffer from the queue
	if cap(t) < len(r)+len(a) { // if it's not long enough make a new one
		t = make(sparseRow, 0, len(r)+len(a))
	}

	// merge overlapping regions
	var i, j int
	for i < len(r) && j < len(a) {
		switch {
		case r[i].index < a[j].index:
			t = append(t, r[i])
			i++
		case r[i].index == a[j].index:
			t = append(t, sparseElem{index: r[i].index, value: r[i].value + a[j].value})
			i++
			j++
		case r[i].index > a[j].index:
			t = append(t, a[j])
			j++
		}
	}

	// finish up
	switch {
	case i < len(r):
		t = append(t, r[i:]...)
	case j < len(a):
		t = append(t, a[j:]...)
	}

	b := make(sparseRow, len(t))
	copy(b, t)

	workbuffers <- t[:0] // clear the buffer and send for next user

	return b
}

func (r sparseRow) foldSub(a sparseRow) sparseRow {
	if len(r) > 0 && len(a) > 0 {
		// safely determine no merging to do
		switch {
		case r[len(r)-1].index < a[0].index:
			t := make(sparseRow, len(r), len(r)+len(a))
			copy(t, r)
			for _, e := range a {
				t = append(t, sparseElem{index: e.index, value: -e.value})
			}
			return t
		case a[len(a)-1].index < r[0].index:
			t := make(sparseRow, 0, len(r)+len(a))
			for _, e := range a {
				t = append(t, sparseElem{index: e.index, value: -e.value})
			}
			t = append(t, r...)
			return t
		}
	}

	t := <-workbuffers          // get a buffer from the queue
	if cap(t) < len(r)+len(a) { // if it's not long enough make a new one
		t = make(sparseRow, 0, len(r)+len(a))
	}

	// merge overlapping regions
	var i, j int
	for i < len(r) && j < len(a) {
		switch {
		case r[i].index < a[j].index:
			t = append(t, r[i])
			i++
		case r[i].index == a[j].index:
			t = append(t, sparseElem{index: r[i].index, value: r[i].value - a[j].value})
			i++
			j++
		case r[i].index > a[j].index:
			t = append(t, a[j])
			j++
		}
	}

	// finish up
	switch {
	case i < len(r):
		t = append(t, r[i:]...)
	case j < len(a):
		for _, e := range a[j:] {
			t = append(t, sparseElem{index: e.index, value: -e.value})
		}
	}

	b := make(sparseRow, len(t))
	copy(b, t)

	workbuffers <- t[:0] // clear the buffer and send for next user

	return b
}

func (r sparseRow) foldMul(a sparseRow) sparseRow {
	t := <-workbuffers          // get a buffer from the queue
	if cap(t) < len(r)+len(a) { // if it's not long enough make a new one
		t = make(sparseRow, 0, len(r)+len(a))
	}

	// merge overlapping regions
	var i, j int
	for i < len(r) && j < len(a) {
		switch {
		case r[i].index < a[j].index:
			i++
		case r[i].index == a[j].index:
			t = append(t, sparseElem{index: r[i].index, value: r[i].value * a[j].value})
			i++
			j++
		case r[i].index > a[j].index:
			j++
		}
	}

	b := make(sparseRow, len(t))
	copy(b, t)

	workbuffers <- t[:0] // clear the buffer and send for next user

	return b
}

func (r sparseRow) foldMulSum(a sparseRow) float64 {
	var s float64

	// merge overlapping regions
	var i, j int
	for i < len(r) && j < len(a) {
		switch {
		case r[i].index < a[j].index:
			i++
		case r[i].index == a[j].index:
			s += r[i].value * a[j].value
			i++
			j++
		case r[i].index > a[j].index:
			j++
		}
	}

	return s
}

func (r sparseRow) foldEqual(a sparseRow) bool {
	if len(r) > 0 && len(a) > 0 && (r[len(r)-1].index < a[0].index || a[len(a)-1].index < r[0].index) {
		for _, e := range a {
			if e.value != 0 {
				return false
			}
		}
		for _, e := range r {
			if e.value != 0 {
				return false
			}
		}
	}

	// merge overlapping regions
	var i, j int
	for i < len(r) && j < len(a) {
		switch {
		case r[i].index < a[j].index:
			if r[i].value != 0 {
				return false
			} else {
				i++
			}
		case r[i].index == a[j].index:
			if r[i].value != a[j].value {
				return false
			} else {
				i++
				j++
			}
		case r[i].index > a[j].index:
			if a[j].value != 0 {
				return false
			} else {
				j++
			}
		}
	}

	// finish up
	switch {
	case i < len(r):
		for _, e := range r[i:] {
			if e.value != 0 {
				return false
			}
		}
	case j < len(a):
		for _, e := range a[j:] {
			if e.value != 0 {
				return false
			}
		}
	}

	return true
}

func (r sparseRow) foldApprox(a sparseRow, error float64) bool {
	if len(r) > 0 && len(a) > 0 && (r[len(r)-1].index < a[0].index || a[len(a)-1].index < r[0].index) {
		for _, e := range a {
			if math.Abs(e.value) > error {
				return false
			}
		}
		for _, e := range r {
			if math.Abs(e.value) > error {
				return false
			}
		}
	}

	// merge overlapping regions
	var i, j int
	for i < len(r) && j < len(a) {
		switch {
		case r[i].index < a[j].index:
			if math.Abs(r[i].value) > error {
				return false
			} else {
				i++
			}
		case r[i].index == a[j].index:
			if math.Abs(r[i].value-a[j].value) > error {
				return false
			} else {
				i++
				j++
			}
		case r[i].index > a[j].index:
			if math.Abs(a[j].value) > error {
				return false
			} else {
				j++
			}
		}
	}

	// finish up
	switch {
	case i < len(r):
		for _, e := range r[i:] {
			if math.Abs(e.value) > error {
				return false
			}
		}
	case j < len(a):
		for _, e := range a[j:] {
			if math.Abs(e.value) > error {
				return false
			}
		}
	}

	return true
}
