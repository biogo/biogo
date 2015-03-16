// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errors

import (
	"fmt"
	"io"

	"gopkg.in/check.v1"
)

func (s *S) TestChain(c *check.C) {
	err := io.EOF
	err = Link(err, fmt.Errorf("failed: %v", err))
	c.Check(err.Error(), check.Equals, "failed: EOF")
	c.Check(Cause(err), check.Equals, io.EOF)
	c.Check(Errors(err), check.DeepEquals, []error{io.EOF, fmt.Errorf("failed: EOF")})
}

// userChain is the basic implementation without an UnwrapAll method.
type userChain []error

func (c userChain) Error() string {
	if len(c) > 0 {
		return c[len(c)-1].Error()
	}
	return ""
}
func (c userChain) Cause() error {
	if len(c) > 0 {
		return c[0]
	}
	return nil
}
func (c userChain) Link(err error) Chain { return append(c, err) }
func (c userChain) Last() (Chain, error) {
	switch len(c) {
	case 0:
		return nil, nil
	case 1:
		return nil, c[0]
	default:
		return c[:len(c)-1], c[len(c)-1]
	}
}

func (s *S) TestUserChain(c *check.C) {
	var err error = userChain{io.EOF}
	err = Link(err, fmt.Errorf("failed: %v", err))
	c.Check(err.Error(), check.Equals, "failed: EOF")
	c.Check(Cause(err), check.Equals, io.EOF)
	unwrapped := Errors(err)
	c.Check(Cause(err), check.Equals, unwrapped[0])
	for i, e := range unwrapped {
		c.Check(e.Error(), check.Equals, []error{io.EOF, fmt.Errorf("failed: EOF")}[i].Error())
	}
}
