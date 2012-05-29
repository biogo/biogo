// Copyright ©2011-2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package morass

import (
	"io"
	check "launchpad.net/gocheck"
	"math/rand"
	"runtime"
	"testing"
	"unsafe"
)

const minInt = -int(^uint(0)>>1) - 1

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

type structLesser struct {
	A int
	B int
}

func (i structLesser) Less(j interface{}) bool { return i.A < j.(structLesser).A }

func (s *S) TestMorass(c *check.C) {
	for _, concurrent := range []bool{false, true} {
		if m, err := New(intLesser(0), "", "", chunk, concurrent); err != nil {
			m.CleanUp()
			c.Fatalf("New Morass failed: %v", err)
		} else {
			var i int
			for i = 0; i < testLen; i++ {
				c.Check(int64(i), check.Equals, m.Pos())
				if err = m.Push(intLesser(rand.Int())); err != nil {
					m.CleanUp()
					c.Fatalf("Push %d failed: %v", i, err)
				}
			}
			if err = m.Finalise(); err != nil {
				m.CleanUp()
				c.Fatalf("Finalise failed: %v", err)
			}
			c.Logf("Pushed %d values", i)
			c.Check(m.Len(), check.Equals, int64(testLen))
		L:
			for i = 0; i <= testLen; i++ {
				var v intLesser
				lv := intLesser(minInt)
				c.Check(int64(i), check.Equals, m.Pos())
				switch err = m.Pull(&v); err {
				case nil:
					c.Check(v.Less(lv), check.Equals, false)
				case io.EOF:
					c.Logf("Pulled %d values", i)
					c.Check(i, check.Equals, testLen)
					break L
				default:
					m.CleanUp()
					c.Fatalf("Pull failed: %v", err)
				}
			}
			if err = m.CleanUp(); err != nil {
				c.Fatalf("CleanUp failed: %v", err)
			}
		}
	}
}

func (s *S) TestFast1(c *check.C) {
	if m, err := New(intLesser(0), "", "", chunk, false); err != nil {
		m.CleanUp()
		c.Fatalf("New Morass failed: %v", err)
	} else {
		for r := 1; r <= 2; r++ {
			var i int
			for i = 0; i < chunk/2; i++ {
				c.Check(int64(i), check.Equals, m.Pos())
				if err = m.Push(intLesser(rand.Int())); err != nil {
					m.CleanUp()
					c.Fatalf("Push %d failed on use %d : %v", i, r, err)
				}
			}
			if err = m.Finalise(); err != nil {
				m.CleanUp()
				c.Fatalf("Finalise failed on use %d: %v", r, err)
			}
			c.Logf("Pushed %d values", i)
			c.Check(m.Len(), check.Equals, int64(chunk/2))
		L:
			for i = 0; i <= testLen; i++ {
				var v intLesser
				lv := intLesser(minInt)
				c.Check(int64(i), check.Equals, m.Pos())
				switch err = m.Pull(&v); err {
				case nil:
					c.Check(v.Less(lv), check.Equals, false)
				case io.EOF:
					c.Logf("Pulled %d values", i)
					c.Check(i, check.Equals, chunk/2)
					break L
				default:
					m.CleanUp()
					c.Fatalf("Pull failed on use %d: %v", r, err)
				}
			}
			if err = m.Clear(); err != nil {
				m.CleanUp()
				c.Fatalf("Clear failed on use %d: %v", r, err)
			}
		}
		if err = m.CleanUp(); err != nil {
			c.Fatalf("CleanUp failed: %v", err)
		}
	}
}

func (s *S) TestFast2(c *check.C) {
	if m, err := New(structLesser{}, "", "", chunk, false); err != nil {
		m.CleanUp()
		c.Fatalf("New Morass failed: %v", err)
	} else {
		for r := 1; r <= 2; r++ {
			var i int
			for i = 0; i < chunk/2; i++ {
				c.Check(int64(i), check.Equals, m.Pos())
				if err = m.Push(structLesser{rand.Int(), r}); err != nil {
					m.CleanUp()
					c.Fatalf("Push %d failed on use %d : %v", i, r, err)
				}
			}
			if err = m.Finalise(); err != nil {
				m.CleanUp()
				c.Fatalf("Finalise failed on use %d: %v", r, err)
			}
			c.Logf("Pushed %d values", i)
			c.Check(m.Len(), check.Equals, int64(chunk/2))
		L:
			for i = 0; i <= testLen; i++ {
				var v structLesser
				lv := structLesser{minInt, 0}
				c.Check(int64(i), check.Equals, m.Pos())
				switch err = m.Pull(&v); err {
				case nil:
					c.Check(v.Less(lv), check.Equals, false)
					c.Check(v.B, check.Equals, r)
				case io.EOF:
					c.Logf("Pulled %d values", i)
					c.Check(i, check.Equals, chunk/2)
					break L
				default:
					m.CleanUp()
					c.Fatalf("Pull failed on use %d: %v", r, err)
				}
			}
			if err = m.Clear(); err != nil {
				m.CleanUp()
				c.Fatalf("Clear failed on use %d: %v", r, err)
			}
		}
		if err = m.CleanUp(); err != nil {
			c.Fatalf("CleanUp failed: %v", err)
		}
	}
}

func (s *S) TestReuse1(c *check.C) {
	if m, err := New(intLesser(0), "", "", chunk, false); err != nil {
		m.CleanUp()
		c.Fatalf("New Morass failed: %v", err)
	} else {
		for r := 1; r <= 2; r++ {
			var i int
			for i = 0; i < testLen; i++ {
				c.Check(int64(i), check.Equals, m.Pos())
				if err = m.Push(intLesser(rand.Int())); err != nil {
					m.CleanUp()
					c.Fatalf("Push %d failed on use %d : %v", i, r, err)
				}
			}
			if err = m.Finalise(); err != nil {
				m.CleanUp()
				c.Fatalf("Finalise failed on use %d: %v", r, err)
			}
			c.Logf("Pushed %d values", i)
			c.Check(m.Len(), check.Equals, int64(testLen))
		L:
			for i = 0; i <= testLen; i++ {
				var v intLesser
				lv := intLesser(minInt)
				c.Check(int64(i), check.Equals, m.Pos())
				switch err = m.Pull(&v); err {
				case nil:
					c.Check(v.Less(lv), check.Equals, false)
				case io.EOF:
					c.Logf("Pulled %d values", i)
					c.Check(i, check.Equals, testLen)
					break L
				default:
					m.CleanUp()
					c.Fatalf("Pull failed on use %d: %v", r, err)
				}
			}
			if err = m.Clear(); err != nil {
				m.CleanUp()
				c.Fatalf("Clear failed on use %d: %v", r, err)
			}
		}
		if err = m.CleanUp(); err != nil {
			c.Fatalf("CleanUp failed: %v", err)
		}
	}
}

func (s *S) TestReuse2(c *check.C) {
	if m, err := New(structLesser{}, "", "", chunk, false); err != nil {
		m.CleanUp()
		c.Fatalf("New Morass failed: %v", err)
	} else {
		for r := 1; r <= 2; r++ {
			var i int
			for i = 0; i < testLen; i++ {
				c.Check(int64(i), check.Equals, m.Pos())
				if err = m.Push(structLesser{rand.Int(), r}); err != nil {
					m.CleanUp()
					c.Fatalf("Push %d failed on use %d : %v", i, r, err)
				}
			}
			if err = m.Finalise(); err != nil {
				m.CleanUp()
				c.Fatalf("Finalise failed on use %d: %v", r, err)
			}
			c.Logf("Pushed %d values", i)
			c.Check(m.Len(), check.Equals, int64(testLen))
		L:
			for i = 0; i <= testLen; i++ {
				var v structLesser
				lv := structLesser{minInt, 0}
				c.Check(int64(i), check.Equals, m.Pos())
				switch err = m.Pull(&v); err {
				case nil:
					c.Check(v.Less(lv), check.Equals, false)
					c.Check(v.B, check.Equals, r)
				case io.EOF:
					c.Logf("Pulled %d values", i)
					c.Check(i, check.Equals, testLen)
					break L
				default:
					m.CleanUp()
					c.Fatalf("Pull failed on use %d: %v", r, err)
				}
			}
			if err = m.Clear(); err != nil {
				m.CleanUp()
				c.Fatalf("Clear failed on use %d: %v", r, err)
			}
		}
		if err = m.CleanUp(); err != nil {
			c.Fatalf("CleanUp failed: %v", err)
		}
	}
}

func (s *S) TestAutoClear(c *check.C) {
	if m, err := New(intLesser(0), "", "", chunk, false); err != nil {
		m.CleanUp()
		c.Fatalf("New Morass failed: %v", err)
	} else {
		m.AutoClear = true
		for r := 1; r <= 2; r++ {
			var i int
			for i = 0; i < testLen; i++ {
				c.Check(int64(i), check.Equals, m.Pos())
				if err = m.Push(intLesser(rand.Int())); err != nil {
					c.Fatalf("Push %d failed on use %d : %v", i, r, err)
				}
			}
			if err = m.Finalise(); err != nil {
				m.CleanUp()
				c.Fatalf("Finalise failed on use %d: %v", r, err)
			}
			c.Logf("Pushed %d values", i)
			c.Check(m.Len(), check.Equals, int64(testLen))
		L:
			for i = 0; i <= testLen; i++ {
				var v intLesser
				lv := intLesser(minInt)
				c.Check(int64(i), check.Equals, m.Pos())
				switch err = m.Pull(&v); err {
				case nil:
					c.Check(v.Less(lv), check.Equals, false)
				case io.EOF:
					c.Logf("Pulled %d values", i)
					c.Check(i, check.Equals, testLen)
					break L
				default:
					m.CleanUp()
					c.Fatalf("Pull failed on repeat %d: %v", r, err)
				}
			}
		}
		if err = m.CleanUp(); err != nil {
			c.Fatalf("CleanUp failed: %v", err)
		}
	}
}

func (s *S) TestAutoClearSafety(c *check.C) {
	if m, err := New(intLesser(0), "", "", chunk, false); err != nil {
		m.CleanUp()
		c.Fatalf("New Morass failed: %v", err)
	} else {
		m.AutoClear = true
		for r := 1; r <= 2; r++ {
			var i int
			for i = 0; i < testLen; i++ {
				c.Check(int64(i), check.Equals, m.Pos())
				if err = m.Push(intLesser(rand.Int())); err != nil {
					c.Fatalf("Push %d failed on use %d : %v", i, r, err)
				}
			}
			if err = m.Finalise(); err != nil {
				m.CleanUp()
				c.Fatalf("Finalise failed on use %d: %v", r, err)
			}
			c.Logf("Pushed %d values", i)
			c.Check(m.Len(), check.Equals, int64(testLen))
		L:
			for i = 0; i <= testLen; i++ {
				var v intLesser
				lv := intLesser(minInt)
				c.Check(int64(i), check.Equals, m.Pos())
				switch err = m.Pull(&v); err {
				case nil:
					c.Check(v.Less(lv), check.Equals, false)
				case io.EOF:
					c.Logf("Pulled %d values", i)
					c.Check(i, check.Equals, testLen)
					break L
				default:
					m.CleanUp()
					c.Fatalf("Pull failed on repeat %d: %v", r, err)
				}
			}
			if err = m.Clear(); err != nil {
				m.CleanUp()
				c.Fatalf("Clear failed on repeat %d: %v", r, err)
			}
		}
		if err = m.CleanUp(); err != nil {
			c.Fatalf("CleanUp failed: %v", err)
		}
	}
}

func BenchmarkFast(b *testing.B) {
	benchmark(b, chunk, chunk/2, true)
}

func BenchmarkConcurrent(b *testing.B) {
	benchmark(b, chunk, testLen, true)
}

func BenchmarkSequential(b *testing.B) {
	benchmark(b, chunk, testLen, false)
}

func benchmark(b *testing.B, chunk, count int, concurrent bool) {
	b.StopTimer()
	runtime.GC()
	b.StartTimer()
	b.SetBytes(int64(unsafe.Sizeof(intLesser(0))) * int64(count))
	if m, err := New(intLesser(0), "", "", chunk, concurrent); err == nil {
		m.AutoClear = true
		for i := 0; i < b.N; i++ {
			for j := 0; j < count; j++ {
				if err = m.Push(intLesser(rand.Int())); err != nil {
					m.CleanUp()
					b.Fatalf("Push %d failed: %v", j, err)
				}
			}
			if err = m.Finalise(); err != nil {
				m.CleanUp()
				b.Fatalf("Finalise failed: %v", err)
			}
		L:
			for j := 0; j <= count; j++ {
				var v intLesser
				switch err = m.Pull(&v); err {
				case nil:
				case io.EOF:
					m.Clear()
					break L
				default:
					m.CleanUp()
					b.Fatalf("Pull %d failed: %v", j, err)
				}
			}
		}
		if err = m.CleanUp(); err != nil {
			b.Fatalf("Finalise failed: %v", err)
		}
	}
}
