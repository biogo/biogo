// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sequtils

import (
	"code.google.com/p/biogo/alphabet"
	"code.google.com/p/biogo/feat"
	"code.google.com/p/biogo/seq"
	"fmt"
	"gopkg.in/check.v1"
	"reflect"
	"testing"
)

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

var (
	lorem = stringToSlice("Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.")
)

type slice []byte

func stringToSlice(s string) slice {
	return []byte(s)
}
func (s slice) Make(len, cap int) alphabet.Slice       { return make(slice, len, cap) }
func (s slice) Len() int                               { return len(s) }
func (s slice) Cap() int                               { return cap(s) }
func (s slice) Slice(start, end int) alphabet.Slice    { return s[start:end] }
func (s slice) Append(a alphabet.Slice) alphabet.Slice { return append(s, a.(slice)...) }
func (s slice) Copy(a alphabet.Slice) int              { return copy(s, a.(slice)) }

type offSlice struct {
	slice
	offset int
	conf   feat.Conformation
}

func stringToOffSlice(s string) *offSlice {
	return &offSlice{slice: stringToSlice(s)}
}
func (os *offSlice) Conformation() feat.Conformation { return os.conf }
func (os *offSlice) SetOffset(o int) error           { os.offset = o; return nil }
func (os *offSlice) Slice() alphabet.Slice           { return os.slice }
func (os *offSlice) SetSlice(sl alphabet.Slice)      { os.slice = sl.(slice) }
func (os *offSlice) String() string                  { return fmt.Sprintf("%s %d", os.slice, os.offset) }

type conformRangeOffSlice struct {
	slice
	offset int
	conf   feat.Conformation
}

func stringToConformRangeOffSlice(s string) *conformRangeOffSlice {
	return &conformRangeOffSlice{slice: stringToSlice(s)}
}
func (os *conformRangeOffSlice) Conformation() feat.Conformation     { return os.conf }
func (os *conformRangeOffSlice) SetConformation(c feat.Conformation) { os.conf = c }
func (os *conformRangeOffSlice) SetOffset(o int) error               { os.offset = o; return nil }
func (os *conformRangeOffSlice) Start() int                          { return os.offset }
func (os *conformRangeOffSlice) End() int                            { return os.offset + len(os.slice) }
func (os *conformRangeOffSlice) Slice() alphabet.Slice               { return os.slice }
func (os *conformRangeOffSlice) SetSlice(sl alphabet.Slice)          { os.slice = sl.(slice) }
func (os *conformRangeOffSlice) String() string                      { return fmt.Sprintf("%s", os.slice) }

func (s *S) TestTruncate(c *check.C) {
	type test struct {
		in         *conformRangeOffSlice
		start, end int
		expect     interface{}
	}

	var T []test = []test{
		{
			in:     &conformRangeOffSlice{slice: lorem, conf: feat.Linear},
			start:  28,
			end:    39,
			expect: stringToSlice("consectetur"),
		},
		{
			in:     &conformRangeOffSlice{slice: lorem, conf: feat.Linear, offset: 1},
			start:  29,
			end:    40,
			expect: stringToSlice("consectetur"),
		},
		{
			in:     &conformRangeOffSlice{slice: lorem, conf: feat.Circular},
			start:  28,
			end:    39,
			expect: stringToSlice("consectetur"),
		},
		{
			in:     &conformRangeOffSlice{slice: lorem, conf: feat.Linear},
			start:  117,
			end:    5,
			expect: "sequtils: start position greater than end position for linear sequence",
		},
		{
			in:     &conformRangeOffSlice{slice: lorem, conf: feat.Circular},
			start:  117,
			end:    5,
			expect: stringToSlice("aliqua.Lorem"),
		},
		{
			in:     &conformRangeOffSlice{slice: lorem, conf: feat.Circular, offset: 1},
			start:  118,
			end:    6,
			expect: stringToSlice("aliqua.Lorem"),
		},
		{
			in:     &conformRangeOffSlice{slice: lorem, conf: feat.Linear},
			start:  5,
			end:    1170,
			expect: "sequtils: index out of range",
		},
		{
			in:     &conformRangeOffSlice{slice: lorem, conf: feat.Circular},
			start:  1170,
			end:    5,
			expect: "sequtils: index out of range",
		},
	}

	for _, t := range T {
		r := &conformRangeOffSlice{}
		if err := Truncate(r, t.in, t.start, t.end); err != nil {
			c.Check(err, check.ErrorMatches, t.expect)
		} else {
			c.Check(r.slice, check.DeepEquals, t.expect)
			c.Check(r.conf, check.Equals, feat.Linear)
			if t.in.conf != feat.Circular || (t.in.conf == feat.Circular && t.end >= t.start) {
				c.Check(t.end-t.start, check.Equals, reflect.ValueOf(t.expect).Len())
			} else {
				c.Check(t.end+reflect.ValueOf(t.in.slice).Len()-t.start, check.Equals, reflect.ValueOf(t.expect).Len())
			}
		}
	}
}

func (s *S) TestJoin(c *check.C) {
	type test struct {
		a, b   *offSlice
		where  int
		expect *offSlice
	}

	var T []test = []test{
		{
			a:      &offSlice{lorem[28:40], 0, feat.Linear},
			b:      &offSlice{lorem[117:], 0, feat.Linear},
			where:  seq.Start,
			expect: &offSlice{stringToSlice("aliqua.consectetur "), -7, feat.Linear},
		},
		{
			a:      &offSlice{lorem[28:40], 0, feat.Linear},
			b:      &offSlice{lorem[117:], 0, feat.Linear},
			where:  seq.End,
			expect: &offSlice{stringToSlice("consectetur aliqua."), 0, feat.Linear},
		},
	}

	for _, t := range T {
		Join(t.a, t.b, t.where)
		c.Check(t.a, check.DeepEquals, t.expect)
	}
}

type fe struct {
	s, e   int
	orient feat.Orientation
	feat.Feature
}

func (f fe) Start() int                    { return f.s }
func (f fe) End() int                      { return f.e }
func (f fe) Len() int                      { return f.e - f.s }
func (f fe) Orientation() feat.Orientation { return f.orient }

type fs []feat.Feature

func (f fs) Features() []feat.Feature { return []feat.Feature(f) }

func (s *S) TestStitch(c *check.C) {
	type test struct {
		in     *conformRangeOffSlice
		f      feat.Set
		expect interface{}
	}

	var T []test = []test{
		{
			in:     &conformRangeOffSlice{slice: lorem, conf: feat.Linear, offset: 0},
			f:      fs{fe{s: 12, e: 18}, fe{s: 24, e: 26}, fe{s: 103, e: 110}},
			expect: stringToSlice("dolor et dolore"),
		},
		{
			in:     &conformRangeOffSlice{slice: lorem, conf: feat.Linear, offset: -1},
			f:      fs{fe{s: 12, e: 18}, fe{s: 24, e: 26}, fe{s: 103, e: 110}},
			expect: stringToSlice("olor st,dolore "),
		},
		{
			in:     &conformRangeOffSlice{slice: lorem, conf: feat.Linear, offset: 0},
			f:      fs{fe{s: 12, e: 18}, fe{s: 13, e: 17}, fe{s: 24, e: 26}, fe{s: 103, e: 110}},
			expect: stringToSlice("dolor et dolore"),
		},
		{
			in:     &conformRangeOffSlice{slice: lorem, conf: feat.Linear, offset: -1},
			f:      fs{fe{s: 12, e: 18}, fe{s: 13, e: 17}, fe{s: 24, e: 26}, fe{s: 103, e: 110}},
			expect: stringToSlice("olor st,dolore "),
		},
		{
			in:     &conformRangeOffSlice{slice: lorem, conf: feat.Linear, offset: 0},
			f:      fs{fe{s: 19, e: 18}},
			expect: "sequtils: feature end < feature start",
		},
	}

	for _, t := range T {
		r := &conformRangeOffSlice{}
		if err := Stitch(r, t.in, t.f); err != nil {
			c.Check(err, check.ErrorMatches, t.expect)
		} else {
			c.Check(r.slice, check.DeepEquals, t.expect)
		}
	}
}

func (s *S) TestCompose(c *check.C) {
	type test struct {
		in     *conformRangeOffSlice
		f      feat.Set
		expect interface{}
	}

	var T []test = []test{
		{
			in:     &conformRangeOffSlice{slice: lorem, conf: feat.Linear, offset: 0},
			f:      fs{fe{s: 12, e: 18}, fe{s: 24, e: 26}, fe{s: 103, e: 110}},
			expect: stringToSlice("dolor et dolore"),
		},
		{
			in:     &conformRangeOffSlice{slice: lorem, conf: feat.Linear, offset: -1},
			f:      fs{fe{s: 12, e: 18}, fe{s: 24, e: 26}, fe{s: 103, e: 110}},
			expect: stringToSlice("olor st,dolore "),
		},
		{
			in:     &conformRangeOffSlice{slice: lorem, conf: feat.Linear, offset: 0},
			f:      fs{fe{s: 12, e: 18}, fe{s: 13, e: 17}, fe{s: 24, e: 26}, fe{s: 103, e: 110}},
			expect: stringToSlice("dolor oloret dolore"),
		},
		{
			in:     &conformRangeOffSlice{slice: lorem, conf: feat.Linear, offset: -1},
			f:      fs{fe{s: 12, e: 18}, fe{s: 13, e: 17}, fe{s: 24, e: 26}, fe{s: 103, e: 110}},
			expect: stringToSlice("olor slor t,dolore "),
		},
		{
			in:     &conformRangeOffSlice{slice: lorem, conf: feat.Linear, offset: -1},
			f:      fs{fe{s: 12, e: 18}, fe{s: 13, e: 17}, fe{s: 24, e: 26, orient: feat.Reverse}, fe{s: 103, e: 110}},
			expect: "sequtils: unable to reverse segment during compose",
		},
		{
			in:     &conformRangeOffSlice{slice: lorem, conf: feat.Linear, offset: 0},
			f:      fs{fe{s: 19, e: 18}},
			expect: "sequtils: feature end < feature start",
		},
	}

	for _, t := range T {
		r := &conformRangeOffSlice{}
		if err := Compose(r, t.in, t.f); err == nil {
			c.Check(r.slice, check.DeepEquals, t.expect)
		} else {
			c.Check(err, check.ErrorMatches, t.expect)
		}
	}
}
