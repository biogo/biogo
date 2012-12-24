// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
