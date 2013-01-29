// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"bytes"
	check "launchpad.net/gocheck"
)

const lorem = `Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.`

// Tests
func (s *S) TestWrapper(c *check.C) {
	for i, t := range []struct {
		w   *Wrapper
		in  string
		out string
		err error
	}{
		{
			w:   &Wrapper{limit: 0, width: 0},
			in:  lorem,
			out: ``,
			err: nil,
		},
		{
			w:   &Wrapper{limit: -1, width: 0},
			in:  lorem,
			out: lorem,
			err: nil,
		},
		{
			w:  &Wrapper{limit: -1, width: 20},
			in: lorem,
			out: "" +
				"Lorem ipsum dolor si\n" +
				"t amet, consectetur \n" +
				"adipisicing elit, se\n" +
				"d do eiusmod tempor \n" +
				"incididunt ut labore\n" +
				" et dolore magna ali\n" +
				"qua.\n",
			err: nil,
		},
		{
			w:  &Wrapper{limit: 33, width: 20},
			in: lorem,
			out: "" +
				"Lorem ipsum dolor si\n" +
				"t amet, conse\n",
			err: nil,
		},
		{
			w:  &Wrapper{n: 2, limit: 33, width: 20},
			in: lorem,
			out: "" +
				"Lorem ipsum dolor \n" +
				"sit amet, con\n",
			err: nil,
		},
	} {
		b := &bytes.Buffer{}
		t.w.w = b
		n, err := t.w.Write([]byte(t.in))
		s := b.String()
		c.Check(err, check.Equals, t.err, check.Commentf("Test %d", i))
		c.Check(s, check.Equals, t.out, check.Commentf("Test %d", i))
		if t.w.limit >= 0 {
			c.Check(t.w.n, check.Equals, min(len(t.in), t.w.limit), check.Commentf("Test %d", i))
		}
		c.Check(n, check.Equals, len(s), check.Commentf("Test %d", i))
	}
}
