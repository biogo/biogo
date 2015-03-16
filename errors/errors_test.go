// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errors

import (
	"testing"

	"gopkg.in/check.v1"
)

func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

// Helpers
func f(i int) error {
	if i == 0 {
		return Make("message", 0, 10)
	}

	i--
	return f(i)
}

var traceRE = `Trace: message:

 github.com/biogo/biogo/errors.f:
	(?:[A-Z]:)?/.*/github.com/biogo/biogo/errors/errors_test.go#L=22
	(?:[A-Z]:)?/.*/github.com/biogo/biogo/errors/errors_test.go#L=26
	(?:[A-Z]:)?/.*/github.com/biogo/biogo/errors/errors_test.go#L=26
	(?:[A-Z]:)?/.*/github.com/biogo/biogo/errors/errors_test.go#L=26
	(?:[A-Z]:)?/.*/github.com/biogo/biogo/errors/errors_test.go#L=26
	(?:[A-Z]:)?/.*/github.com/biogo/biogo/errors/errors_test.go#L=26

 github.com/biogo/biogo/errors.\(\*S\).TestCaller:
	(?:[A-Z]:)?/.*/github.com/biogo/biogo/errors/errors_test.go#L=52
`

// Tests
func (s *S) TestCaller(c *check.C) {
	err := Make("message", 0, 10, "item")
	c.Check(err.Error(), check.Equals, "message")
	fn, ln := err.FileLine()
	c.Check(fn, check.Matches, "(?:[A-Z]:)?/.*/biogo/errors/errors_test.go")
	c.Check(ln, check.Equals, 45)
	c.Check(err.Package(), check.Equals, "github.com/biogo/biogo/errors.(*S)")
	c.Check(err.Function(), check.Equals, "TestCaller")
	err = f(5).(Error)
	c.Check(err.Tracef(7), check.Matches, traceRE)
}

func (s *S) TestMakeFail(c *check.C) {
	c.Check(func() { Make("message", 0, 0) }, check.Panics, "errors: zero trace depth")
}
