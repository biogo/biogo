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
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/util"
	"fmt"
)

type Set []seq.Sequence

// Interface guarantees
var (
	_ seq.Rower       = &Set{}
	_ seq.RowAppender = &Set{}
)

// Append each []byte in a to the appropriate sequence in the reciever.
func (s Set) AppendEach(a [][]alphabet.QLetter) (err error) {
	if len(a) != s.Rows() {
		return fmt.Errorf("multi: number of sequences does not match row count: %d != %d.", len(a), s.Rows())
	}
	for i, r := range s {
		r.(seq.Appender).AppendQLetters(a[i]...)
	}
	return nil
}

func (s Set) Row(i int) seq.Sequence {
	return s[i]
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
	return len(s)
}

func (s Set) Reverse() {
	for _, r := range s {
		r.Reverse()
	}
}

func (s Set) RevComp() {
	for _, r := range s {
		r.RevComp()
	}
}
