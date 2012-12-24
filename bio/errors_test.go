// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bio

import (
	"fmt"
	check "launchpad.net/gocheck"
)

// Helpers
func f(i int) error {
	if i == 0 {
		return NewError("message", 0, nil)
	}

	i--
	return f(i)
}

// Tests
func (s *S) TestCaller(c *check.C) {
	err := NewError("message", 0, "item")
	c.Check(err.Error(), check.Equals, "message")
	fn, ln := err.FileLine()
	c.Check(fn, check.Matches, "/.*/biogo/bio/errors_test.go")
	c.Check(ln, check.Equals, 24)
	c.Check(err.Package(), check.Equals, "code.google.com/p/biogo/bio.(*S)")
	c.Check(err.Function(), check.Equals, "TestCaller")
	err = f(5).(Error)
	fmt.Println(err.Tracef(10))
}
