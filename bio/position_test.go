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

package bio

import (
	check "launchpad.net/gocheck"
)

type testPos []struct {
	zb, ob int
}

var T = testPos{
	{0, 1},
	{1, 0},
	{-1, -1},
	{1, 2},
}

// Tests
func (s *S) TestPosition(c *check.C) {
	for _, t := range T {
		if t.ob == 0 {
			c.Check(func() { OneToZero(t.ob) }, check.Panics, "1-based index == 0")
		} else {
			c.Check(OneToZero(t.ob), check.Equals, t.zb)
			c.Check(ZeroToOne(t.zb), check.Equals, t.ob)
		}
	}
}
