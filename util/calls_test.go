// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	check "launchpad.net/gocheck"
)

// Helpers
func f1(c *check.C) {
	c.Check(GetCaller(0).Package, check.Equals, "code.google.com/p/biogo/util")
	c.Check(GetCaller(0).Function, check.Equals, "f1")
	c.Check(GetCaller(1).Package, check.Equals, "code.google.com/p/biogo/util.(*S)")
	c.Check(GetCaller(1).Function, check.Equals, "TestCaller")
	f2(c)
}

func f2(c *check.C) {
	c.Check(GetCaller(0).Package, check.Equals, "code.google.com/p/biogo/util")
	c.Check(GetCaller(0).Function, check.Equals, "f2")
	c.Check(GetCaller(1).Package, check.Equals, "code.google.com/p/biogo/util")
	c.Check(GetCaller(1).Function, check.Equals, "f1")
	c.Check(GetCaller(2).Package, check.Equals, "code.google.com/p/biogo/util.(*S)")
	c.Check(GetCaller(2).Function, check.Equals, "TestCaller")
}

// Tests
func (s *S) TestCaller(c *check.C) {
	c.Check(GetCaller(0).Package, check.Equals, "code.google.com/p/biogo/util.(*S)")
	c.Check(GetCaller(0).Function, check.Equals, "TestCaller")
	f1(c)
}
