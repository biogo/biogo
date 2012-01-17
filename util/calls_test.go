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

// Helpers
func f1(c *check.C) {
	c.Check(Name(0), check.Equals, &Caller{Package: "github.com/kortschak/BioGo/util", Function: "f1"})
	c.Check(Name(1), check.Equals, &Caller{Package: "github.com/kortschak/BioGo/util.(*S)", Function: "TestCaller"})
	f2(c)
}

func f2(c *check.C) {
	c.Check(Name(0), check.Equals, &Caller{Package: "github.com/kortschak/BioGo/util", Function: "f2"})
	c.Check(Name(1), check.Equals, &Caller{Package: "github.com/kortschak/BioGo/util", Function: "f1"})
	c.Check(Name(2), check.Equals, &Caller{Package: "github.com/kortschak/BioGo/util.(*S)", Function: "TestCaller"})
}

// Tests
func (s *S) TestCaller(c *check.C) {
	c.Check(Name(0), check.Equals, &Caller{Package: "github.com/kortschak/BioGo/util.(*S)", Function: "TestCaller"})
	f1(c)
}
