package morass

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
	"github.com/kortschak/BioGo/util"
	"io"
	check "launchpad.net/gocheck"
	"math/rand"
	"testing"
)

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

const (
	chunk   = 100
	testLen = 10000
)

var _ = check.Suite(&S{})

type intLesser int

func (i intLesser) Less(j interface{}) bool { return i < j.(intLesser) }

func (s *S) TestMorass(c *check.C) {
	for _, concurrent := range []bool{false, true} {
		if m, err := New("", "", chunk, concurrent); err != nil {
			c.Fatalf("New Morass failed: %v", err)
		} else {
			for i := 0; i < testLen; i++ {
				if err = m.Push(intLesser(rand.Int())); err != nil {
					c.Fatalf("Push %d failed: %v", i, err)
					c.Check(int64(i), check.Equals, m.Pos())
				}
			}
			if err = m.Finalise(); err != nil {
				c.Fatalf("Finalise failed: %v", err)
			}
			c.Check(m.Len(), check.Equals, int64(testLen))
		L:
			for i := 0; i <= testLen; i++ {
				var v intLesser
				lv := intLesser(util.MinInt)
				c.Check(int64(i), check.Equals, m.Pos())
				switch err = m.Pull(&v); err {
				case nil:
					c.Check(v.Less(lv), check.Equals, false)
				case io.EOF:
					break L
				default:
					c.Fatalf("Pull failed: %v", err)
				}
			}
			if err = m.CleanUp(); err != nil {
				c.Fatalf("CleanUp failed: %v", err)
			}
		}
	}
}
