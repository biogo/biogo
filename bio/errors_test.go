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
	c.Check(ln, check.Equals, 35)
	c.Check(err.Package(), check.Equals, "code.google.com/p/biogo/bio.(*S)")
	c.Check(err.Function(), check.Equals, "TestCaller")
	err = f(5).(Error)
	fmt.Println(err.Tracef(10))
}
