package util

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
	check "launchpad.net/gocheck"
)

// Tests
type tdb struct {
	k, n   byte
	obtain []byte
}

var T []tdb = []tdb{
	{k: 2, n: 2, obtain: []byte{0, 0, 1, 1}},
	{k: 2, n: 3, obtain: []byte{0, 0, 0, 1, 0, 1, 1, 1}},
	{k: 2, n: 4, obtain: []byte{0, 0, 0, 0, 1, 0, 0, 1, 1, 0, 1, 0, 1, 1, 1, 1}},
	{k: 3, n: 4, obtain: []byte{0, 0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 1, 1, 0, 0, 1, 2, 0, 0, 2, 1, 0, 0, 2, 2, 0, 1, 0, 1, 0, 2, 0, 1, 1, 1, 0, 1, 1, 2, 0, 1, 2, 1, 0, 1, 2, 2, 0, 2, 0, 2, 1, 1, 0, 2, 1, 2, 0, 2, 2, 1, 0, 2, 2, 2, 1, 1, 1, 1, 2, 1, 1, 2, 2, 1, 2, 1, 2, 2, 2, 2}},
	{k: 4, n: 2, obtain: []byte{0, 0, 1, 0, 2, 0, 3, 1, 1, 2, 1, 3, 2, 2, 3, 3}},
	{k: 4, n: 3, obtain: []byte{0, 0, 0, 1, 0, 0, 2, 0, 0, 3, 0, 1, 1, 0, 1, 2, 0, 1, 3, 0, 2, 1, 0, 2, 2, 0, 2, 3, 0, 3, 1, 0, 3, 2, 0, 3, 3, 1, 1, 1, 2, 1, 1, 3, 1, 2, 2, 1, 2, 3, 1, 3, 2, 1, 3, 3, 2, 2, 2, 3, 2, 3, 3, 3}},
}

func (s *S) TestDeBruijn(c *check.C) {
	for i := 0; i < 256; i++ {
		e := make([]byte, i)
		for j := range e {
			e[j] = byte(j)
		}
		c.Check(DeBruijn(byte(i), 1), check.DeepEquals, e)
	}
	for _, t := range T {
		c.Check(DeBruijn(t.k, t.n), check.DeepEquals, t.obtain)
	}
}
