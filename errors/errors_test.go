// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errors

import (
	check "launchpad.net/gocheck"
	"testing"
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

 code.google.com/p/biogo/errors.f:
	.*/code.google.com/p/biogo/errors/errors_test.go#L=21
	.*/code.google.com/p/biogo/errors/errors_test.go#L=25
	.*/code.google.com/p/biogo/errors/errors_test.go#L=25
	.*/code.google.com/p/biogo/errors/errors_test.go#L=25
	.*/code.google.com/p/biogo/errors/errors_test.go#L=25
	.*/code.google.com/p/biogo/errors/errors_test.go#L=25

 code.google.com/p/biogo/errors.\(\*S\).TestCaller:
	.*/code.google.com/p/biogo/errors/errors_test.go#L=60

 reflect.Value.call:
	.*/go/src/pkg/reflect/value.go#L=[0-9]+

 reflect.Value.Call:
	.*/go/src/pkg/reflect/value.go#L=[0-9]+

 launchpad.net/gocheck._func_006:
	.*/launchpad.net/gocheck/gocheck.go#L=[0-9]+
`

// Tests
func (s *S) TestCaller(c *check.C) {
	err := Make("message", 0, 10, "item")
	c.Check(err.Error(), check.Equals, "message")
	fn, ln := err.FileLine()
	c.Check(fn, check.Matches, "/.*/biogo/errors/errors_test.go")
	c.Check(ln, check.Equals, 53)
	c.Check(err.Package(), check.Equals, "code.google.com/p/biogo/errors.(*S)")
	c.Check(err.Function(), check.Equals, "TestCaller")
	err = f(5).(Error)
	c.Check(err.Tracef(10), check.Matches, traceRE)
}

func (s *S) TestMakeFail(c *check.C) {
	c.Check(func() { Make("message", 0, 0) }, check.Panics, "errors: zero trace depth")
}
