// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"gopkg.in/check.v1"
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
	for i := 0; i < 15; i++ {
		c.Check(Pow(4, byte(i)), check.Equals, int(Pow4(i)))
		c.Check(Pow(4, byte(i)), check.Equals, int(math.Pow(4, float64(i))))
		c.Check(int(Log4(float64(Pow4(i)))), check.Equals, i)
	}
}
