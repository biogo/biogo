// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	check "launchpad.net/gocheck"
)

// Tests
type tdb struct {
	k, n   byte
	obtain []byte
}

var T []tdb = []tdb{
	{k: 2, n: 2, obtain: []byte{0, 0, 1, 1}},
	{k: 2, n: 3, obtain: []byte{0, 0, 0, 1, 0, 1, 1, 1}},
	{k: 2, n: 4, obtain: []byte{0, 0, 0, 0, 1, 0, 0, 1, 1, 0, 1, 0, 1, 1, 1, 1}},
	{k: 3, n: 4, obtain: []byte{0, 0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 1, 1, 0, 0, 1, 2, 0, 0, 2, 1, 0, 0, 2, 2, 0, 1, 0, 1, 0, 2, 0, 1, 1, 1, 0, 1, 1, 2, 0, 1, 2, 1, 0, 1, 2, 2, 0, 2, 0, 2, 1, 1, 0, 2, 1, 2, 0, 2, 2, 1, 0, 2, 2, 2, 1, 1, 1, 1, 2, 1, 1, 2, 2, 1, 2, 1, 2, 2, 2, 2}},
	{k: 4, n: 2, obtain: []byte{0, 0, 1, 0, 2, 0, 3, 1, 1, 2, 1, 3, 2, 2, 3, 3}},
	{k: 4, n: 3, obtain: []byte{0, 0, 0, 1, 0, 0, 2, 0, 0, 3, 0, 1, 1, 0, 1, 2, 0, 1, 3, 0, 2, 1, 0, 2, 2, 0, 2, 3, 0, 3, 1, 0, 3, 2, 0, 3, 3, 1, 1, 1, 2, 1, 1, 3, 1, 2, 2, 1, 2, 3, 1, 3, 2, 1, 3, 3, 2, 2, 2, 3, 2, 3, 3, 3}},
}

func (s *S) TestDeBruijn(c *check.C) {
	for i := 0; i < 256; i++ {
		e := make([]byte, i)
		for j := range e {
			e[j] = byte(j)
		}
		c.Check(DeBruijn(byte(i), 1), check.DeepEquals, e)
	}
	for _, t := range T {
		c.Check(DeBruijn(t.k, t.n), check.DeepEquals, t.obtain)
	}
}
