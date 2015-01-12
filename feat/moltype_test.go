// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package feat

import (
	"gopkg.in/check.v1"
)

// Tests
func (s *S) TestMoltype(c *check.C) {
	for i, s := range moltypeToString {
		c.Check(Moltype(i).String(), check.Equals, s)
		c.Check(ParseMoltype(s), check.Equals, Moltype(i))
		c.Check(ParseMoltype(s+"salt"), check.Equals, Undefined)
	}
}
