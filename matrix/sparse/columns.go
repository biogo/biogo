
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

// N.B. There is fairly extensive use of unrecovered panics in sparse to avoid unreadable matrix formulae.

package sparse

import (
	"math"
)

type sparsecol []elem

type elem struct {
	r     int
	value float64
}

var GrowFraction int = 10

func (c *sparsecol) insert(pos int, val elem) {
	switch length := len(*c); {
	case pos >= length:
		t := make(sparsecol, pos, pos+pos/GrowFraction+1)
		copy(t, (*c))
		t = append(t, val)
		(*c) = t
	case length == cap(*c):
		t := make(sparsecol, length+1, length+length/GrowFraction+1)
		copy(t, (*c)[:pos])
		copy(t[pos+1:], (*c)[pos:])
		*c = t
		(*c)[pos] = val
	case pos < length:
		*c = append((*c)[:pos+1], (*c)[pos:]...)
		(*c)[pos] = val
	}
}

func (c sparsecol) sum() (s float64) {
	for _, e := range c {
		s += e.value
	}

	return
}

func (c sparsecol) min() (n float64) {
	if len(c) == 0 {
		return 0
	}
	n = math.MaxFloat64
	for _, e := range c {
		if e.value < n {
			n = e.value
		}
	}

	return
}

func (c sparsecol) max() (n float64) {
	if len(c) == 0 {
		return 0
	}
	n = -math.MaxFloat64
	for _, e := range c {
		if e.value > n {
			n = e.value
		}
	}

	return
}

func (c sparsecol) scale(beta float64) (b sparsecol) {
	b = make(sparsecol, 0, len(c))
	for _, e := range c {
		b = append(b, elem{r: e.r, value: e.value * beta})
	}

	return
}

func (c sparsecol) foldadd(a sparsecol) (b sparsecol) {
	if len(c) > 0 && len(a) > 0 {
		// safely determine no merging to do
		switch {
		case c[len(c)-1].r < a[0].r:
			t := append(c, a...)
			return t
		case a[len(a)-1].r < c[0].r:
			t := append(a, c...)
			return t
		}
	}

	t := <-workbuffers          // get a buffer from the queue
	if cap(t) < len(c)+len(a) { // if it's not long enough make a new one
		t = make(sparsecol, 0, len(c)+len(a))
	}

	// merge overlapping regions
	var i, j int
	for i < len(c) && j < len(a) {
		switch {
		case c[i].r < a[j].r:
			t = append(t, c[i])
			i++
		case c[i].r == a[j].r:
			t = append(t, elem{r: c[i].r, value: c[i].value + a[j].value})
			i++
			j++
		case c[i].r > a[j].r:
			t = append(t, a[j])
			j++
		}
	}

	// finish up
	switch {
	case i < len(c):
		t = append(t, c[i:]...)
	case j < len(a):
		t = append(t, a[j:]...)
	}

	b = make(sparsecol, len(t))
	copy(b, t)

	workbuffers <- t[:0] // clear the buffer and send for next user

	return
}

func (c sparsecol) foldsub(a sparsecol) (b sparsecol) {
	if len(c) > 0 && len(a) > 0 {
		// safely determine no merging to do
		switch {
		case c[len(c)-1].r < a[0].r:
			t := make(sparsecol, len(c), len(c)+len(a))
			copy(t, c)
			for _, e := range a {
				t = append(t, elem{r: e.r, value: -e.value})
			}
			return t
		case a[len(a)-1].r < c[0].r:
			t := make(sparsecol, 0, len(c)+len(a))
			for _, e := range a {
				t = append(t, elem{r: e.r, value: -e.value})
			}
			t = append(t, c...)
			return t
		}
	}

	t := <-workbuffers          // get a buffer from the queue
	if cap(t) < len(c)+len(a) { // if it's not long enough make a new one
		t = make(sparsecol, 0, len(c)+len(a))
	}

	// merge overlapping regions
	var i, j int
	for i < len(c) && j < len(a) {
		switch {
		case c[i].r < a[j].r:
			t = append(t, c[i])
			i++
		case c[i].r == a[j].r:
			t = append(t, elem{r: c[i].r, value: c[i].value - a[j].value})
			i++
			j++
		case c[i].r > a[j].r:
			t = append(t, a[j])
			j++
		}
	}

	// finish up
	switch {
	case i < len(c):
		t = append(t, c[i:]...)
	case j < len(a):
		for _, e := range a[j:] {
			t = append(t, elem{r: e.r, value: -e.value})
		}
	}

	b = make(sparsecol, len(t))
	copy(b, t)

	workbuffers <- t[:0] // clear the buffer and send for next user

	return
}

func (c sparsecol) foldmul(a sparsecol) (b sparsecol) {
	t := <-workbuffers          // get a buffer from the queue
	if cap(t) < len(c)+len(a) { // if it's not long enough make a new one
		t = make(sparsecol, 0, len(c)+len(a))
	}

	// merge overlapping regions
	var i, j int
	for i < len(c) && j < len(a) {
		switch {
		case c[i].r < a[j].r:
			i++
		case c[i].r == a[j].r:
			t = append(t, elem{r: c[i].r, value: c[i].value * a[j].value})
			i++
			j++
		case c[i].r > a[j].r:
			j++
		}
	}

	b = make(sparsecol, len(t))
	copy(b, t)

	workbuffers <- t[:0] // clear the buffer and send for next user

	return
}

func (c sparsecol) foldequal(a sparsecol) (equality bool) {
	if len(c) > 0 && len(a) > 0 && (c[len(c)-1].r < a[0].r || a[len(a)-1].r < c[0].r) {
		for _, e := range a {
			if e.value != 0 {
				return
			}
		}
		for _, e := range c {
			if e.value != 0 {
				return
			}
		}
	}

	// merge overlapping regions
	var i, j int
	for i < len(c) && j < len(a) {
		switch {
		case c[i].r < a[j].r:
			if c[i].value != 0 {
				return
			} else {
				i++
			}
		case c[i].r == a[j].r:
			if c[i].value != a[j].value {
				return
			} else {
				i++
				j++
			}
		case c[i].r > a[j].r:
			if a[j].value != 0 {
				return
			} else {
				j++
			}
		}
	}

	// finish up
	switch {
	case i < len(c):
		for _, e := range c[i:] {
			if e.value != 0 {
				return
			}
		}
	case j < len(a):
		for _, e := range a[j:] {
			if e.value != 0 {
				return
			}
		}
	}

	return true
}

func (c sparsecol) foldapprox(a sparsecol, error float64) (equality bool) {
	if len(c) > 0 && len(a) > 0 && (c[len(c)-1].r < a[0].r || a[len(a)-1].r < c[0].r) {
		for _, e := range a {
			if math.Abs(e.value) > error {
				return
			}
		}
		for _, e := range c {
			if math.Abs(e.value) > error {
				return
			}
		}
	}

	// merge overlapping regions
	var i, j int
	for i < len(c) && j < len(a) {
		switch {
		case c[i].r < a[j].r:
			if math.Abs(c[i].value) > error {
				return
			} else {
				i++
			}
		case c[i].r == a[j].r:
			if math.Abs(c[i].value-a[j].value) > error {
				return
			} else {
				i++
				j++
			}
		case c[i].r > a[j].r:
			if math.Abs(a[j].value) > error {
				return
			} else {
				j++
			}
		}
	}

	// finish up
	switch {
	case i < len(c):
		for _, e := range c[i:] {
			if math.Abs(e.value) > error {
				return
			}
		}
	case j < len(a):
		for _, e := range a[j:] {
			if math.Abs(e.value) > error {
				return
			}
		}
	}

	return true
}
