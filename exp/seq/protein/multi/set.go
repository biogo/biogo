package multi

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

import (
	"fmt"
	"github.com/kortschak/biogo/bio"
	"github.com/kortschak/biogo/exp/alphabet"
	"github.com/kortschak/biogo/exp/seq"
	"github.com/kortschak/biogo/exp/seq/protein"
	"github.com/kortschak/biogo/util"
)

type Set []protein.Sequence

// Interface guarantees:
var (
	_ protein.Getter         = &Multi{}
	_ protein.GetterAppender = &Multi{}
)

// Append each []byte in a to the appropriate sequence in the reciever.
func (self Set) AppendEach(a [][]alphabet.QLetter) (err error) {
	if len(a) != self.Count() {
		return bio.NewError(fmt.Sprintf("Number of sequences does not match Count(): %d != %d.", len(a), self.Count()), 0, a)
	}
	var i int
	for _, s := range self {
		if m, ok := s.(protein.GetterAppender); ok {
			count := m.Count()
			if m.AppendEach(a[i:i+count]) != nil {
				panic("internal size mismatch")
			}
			i += count
		} else {
			if ap, ok := s.(seq.Appender); ok {
				ap.AppendQLetters(a[i]...)
			} else {
				panic("Non-Multiple Sequence type without Append")
			}
			i++
		}
	}

	return
}

func (self Set) Get(i int) protein.Sequence {
	var count int
	for _, s := range self {
		if m, ok := s.(protein.Getter); ok {
			count = m.Count()
			if i < count {
				return m.Get(i)
			}
		} else {
			count = 1
			if i == 0 {
				return s
			}
		}
		i -= count
	}

	panic("index out of range")
}

func (self Set) Len() int {
	max := util.MinInt

	for _, s := range self {
		if l := s.Len(); l > max {
			max = l
		}
	}

	return max
}

func (self Set) Count() (c int) {
	for _, s := range self {
		if m, ok := s.(protein.Getter); ok {
			c += m.Count()
		} else {
			c++
		}
	}

	return
}

func (self Set) Reverse() {
	for _, s := range self {
		s.Reverse()
	}
}
