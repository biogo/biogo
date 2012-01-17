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
	"math"
)

// Tests
func (s *S) TestMin(c *check.C) {
	c.Check(Min(MaxInt, MinInt), check.Equals, MinInt)
	c.Check(Min(0, MinInt), check.Equals, MinInt)
	c.Check(Min(0, MaxInt), check.Equals, 0)
}

func (s *S) TestUMin(c *check.C) {
	c.Check(UMin(MaxUint, MinUint), check.Equals, MinUint)
	c.Check(UMin(0, MinUint), check.Equals, MinUint)
	c.Check(UMin(0, MaxUint), check.Equals, uint(0))
}

func (s *S) TestMax(c *check.C) {
	c.Check(Max(MaxInt, MinInt), check.Equals, MaxInt)
	c.Check(Max(0, MinInt), check.Equals, 0)
	c.Check(Max(0, MaxInt), check.Equals, MaxInt)
}

func (s *S) TestUMax(c *check.C) {
	c.Check(UMax(MaxUint, MinUint), check.Equals, MaxUint)
	c.Check(UMax(0, MinUint), check.Equals, uint(0))
	c.Check(UMax(0, MaxUint), check.Equals, MaxUint)
}

func (s *S) TestPowLog(c *check.C) {
	for i := 0; i <15; i++ {
		c.Check(Pow(4, byte(i)), check.Equals, int(Pow4(i)))
		c.Check(Pow(4, byte(i)), check.Equals, int(math.Pow(4, float64(i))))
		c.Check(int(Log4(float64(Pow4(i)))), check.Equals, i)
	}
}
