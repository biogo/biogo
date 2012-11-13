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

package multi

import (
	"code.google.com/p/biogo/bio"
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/exp/seq/protein"
	"code.google.com/p/biogo/util"
	"fmt"
)

type Set []protein.Sequence

// Interface guarantees
var (
	_ protein.Getter         = &Set{}
	_ protein.GetterAppender = &Set{}
)

// Append each []QLetter in a to the appropriate sequence in the receiver.
func (s Set) AppendEach(a [][]alphabet.QLetter) (err error) {
	if len(a) != s.Rows() {
		return bio.NewError(fmt.Sprintf("Number of sequences does not match Rows(): %d != %d.", len(a), s.Rows()), 0, a)
	}
	var i int
	for _, r := range s {
		if m, ok := r.(protein.GetterAppender); ok {
			row := m.Rows()
			if m.AppendEach(a[i:i+row]) != nil {
				panic("internal size mismatch")
			}
			i += row
		} else {
			if ap, ok := r.(seq.Appender); ok {
				ap.AppendQLetters(a[i]...)
			} else {
				panic("Non-Multiple Sequence type without Append")
			}
			i++
		}
	}

	return
}

func (s Set) Get(i int) protein.Sequence {
	var row int
	for _, r := range s {
		if m, ok := r.(protein.Getter); ok {
			row = m.Rows()
			if i < row {
				return m.Get(i)
			}
		} else {
			row = 1
			if i == 0 {
				return r
			}
		}
		i -= row
	}

	panic("index out of range")
}

func (s Set) Len() int {
	max := util.MinInt

	for _, r := range s {
		if l := r.Len(); l > max {
			max = l
		}
	}

	return max
}

func (s Set) Rows() (c int) {
	for _, r := range s {
		if m, ok := r.(rowCounter); ok {
			c += m.Rows()
		} else {
			c++
		}
	}

	return
}

func (s Set) Reverse() {
	for _, r := range s {
		r.Reverse()
	}
}
