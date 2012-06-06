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

package sequtils

import (
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/feat"
	check "launchpad.net/gocheck"
	"reflect"
	"testing"
)

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

var (
	lorem = "Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
)

//func Truncate(pol interface{}, start, end int, circular bool) (p interface{}, err error)
// Truncate pol from start to end, allowing wrap around if circular, and return.
func (s *S) TestTruncate(c *check.C) {
	type test struct {
		in         interface{}
		start, end int
		circular   bool
		expect     interface{}
	}

	var T []test = []test{
		{in: []byte(lorem), start: 28, end: 39, circular: false, expect: []byte("consectetur")},
		{in: []byte(lorem), start: 28, end: 39, circular: true, expect: []byte("consectetur")},
		{in: []byte(lorem), start: 117, end: 5, circular: false, expect: "Start position greater than end position for non-circular sequence."},
		{in: []byte(lorem), start: 117, end: 5, circular: true, expect: []byte("aliqua.Lorem")},
	}

	for _, t := range T {
		if trunc, err := Truncate(t.in, t.start, t.end, t.circular); err == nil {
			c.Check(trunc, check.DeepEquals, t.expect)
			if !t.circular || (t.circular && t.end >= t.start) {
				c.Check(t.end-t.start, check.Equals, reflect.ValueOf(t.expect).Len())
			} else {
				c.Check(t.end+reflect.ValueOf(t.in).Len()-t.start, check.Equals, reflect.ValueOf(t.expect).Len())
			}
		} else {
			c.Check(err, check.ErrorMatches, t.expect)
		}
	}
}

//func Reverse(pol interface{}) interface{}
// Reverse the order of pol inplace and return.
func (s *S) TestReverse(c *check.C) {
	type test struct {
		in, expect interface{}
	}

	var T []test = []test{
		{
			in:     []byte(lorem),
			expect: []byte(".auqila angam erolod te erobal tu tnudidicni ropmet domsuie od des ,tile gnicisipida rutetcesnoc ,tema tis rolod muspi meroL"),
		},
	}

	for _, t := range T {
		c.Check(Reverse(t.in), check.DeepEquals, t.expect)
		c.Check(Reverse(t.in), check.DeepEquals, Reverse(Reverse(t.in)))
	}
}

//func Join(pol interface{}, p interface{}, where int) (j interface{}, offset int)
// Join p to pol inplace according to where specified, and return.
func (s *S) TestJoin(c *check.C) {
	type result struct {
		v interface{}
		i int
	}
	type test struct {
		a, b   interface{}
		where  int
		expect result
	}

	var T []test = []test{
		{a: []byte(lorem[28:40]), b: []byte(lorem[117:]), where: seq.Start, expect: result{v: []byte("aliqua.consectetur "), i: -7}},
		{a: []byte(lorem[28:40]), b: []byte(lorem[117:]), where: seq.End, expect: result{v: []byte("consectetur aliqua."), i: 0}},
	}

	for _, t := range T {
		v, i := Join(t.a, t.b, t.where)
		c.Check(v, check.DeepEquals, t.expect.v)
		c.Check(i, check.Equals, t.expect.i)
	}
}

//func Stitch(pol interface{}, offset int, f feat.FeatureSet) (s interface{}, err error)
// Join together colinear segments of pol described by f with an offset applied, and return.
func (s *S) TestStitch(c *check.C) {
	type result struct {
		v interface{}
		i int
	}
	type test struct {
		in     interface{}
		f      feat.FeatureSet
		offset int
		expect result
	}

	var T []test = []test{
		{
			in: []byte(lorem), f: feat.FeatureSet{{Start: 12, End: 18}, {Start: 24, End: 26}, {Start: 103, End: 110}},
			expect: result{v: []byte("dolor et dolore"), i: 15},
		},
		{
			in: []byte(lorem), f: feat.FeatureSet{{Start: 12, End: 18}, {Start: 24, End: 26}, {Start: 103, End: 110}}, offset: -1,
			expect: result{v: []byte("olor st,dolore "), i: 15},
		},
		{
			in: []byte(lorem), f: feat.FeatureSet{{Start: 12, End: 18}, {Start: 13, End: 17}, {Start: 24, End: 26}, {Start: 103, End: 110}},
			expect: result{v: []byte("dolor et dolore"), i: 15},
		},
		{
			in: []byte(lorem), f: feat.FeatureSet{{Start: 12, End: 18}, {Start: 13, End: 17}, {Start: 24, End: 26}, {Start: 103, End: 110}}, offset: -1,
			expect: result{v: []byte("olor st,dolore "), i: 15},
		},
		{
			in: []byte(lorem), f: feat.FeatureSet{{Start: 19, End: 18}},
			expect: result{v: "Interval end < start"},
		},
	}

	for _, t := range T {
		if stitch, err := Stitch(t.in, t.offset, t.f); err == nil {
			c.Check(stitch, check.DeepEquals, t.expect.v)
			c.Check(reflect.ValueOf(t.expect.v).Len(), check.Equals, t.expect.i)
		} else {
			c.Check(err, check.ErrorMatches, t.expect.v)
		}
	}
}

//func Compose(pol interface{}, offset int, f feat.FeatureSet) (s interface{}, err error)
// Join together sequentially ordered disjunct segments of pol described by f with an offset applied, and return.
func (s *S) TestCompose(c *check.C) {
	type test struct {
		in     interface{}
		f      feat.FeatureSet
		offset int
		expect interface{}
	}

	var T []test = []test{
		{in: []byte(lorem), f: feat.FeatureSet{{Start: 12, End: 18}, {Start: 24, End: 26}, {Start: 103, End: 110}}, expect: []byte("dolor et dolore")},
		{in: []byte(lorem), f: feat.FeatureSet{{Start: 12, End: 18}, {Start: 24, End: 26}, {Start: 103, End: 110}}, offset: -1, expect: []byte("olor st,dolore ")},
		{in: []byte(lorem), f: feat.FeatureSet{{Start: 12, End: 18}, {Start: 13, End: 17}, {Start: 24, End: 26}, {Start: 103, End: 110}}, expect: []byte("dolor oloret dolore")},
		{in: []byte(lorem), f: feat.FeatureSet{{Start: 12, End: 18}, {Start: 13, End: 17}, {Start: 24, End: 26}, {Start: 103, End: 110}}, offset: -1, expect: []byte("olor slor t,dolore ")},
		{in: []byte(lorem), f: feat.FeatureSet{{Start: 19, End: 18}}, expect: "Feature End < Start"},
	}

	for _, t := range T {
		if compose, err := Compose(t.in, t.offset, t.f); err == nil {
			composed := []byte{}
			for _, seg := range compose {
				composed = append(composed, seg.([]byte)...)
			}
			c.Check(composed, check.DeepEquals, t.expect)
			l := 0
			for _, f := range t.f {
				l += f.Len()
			}
			c.Check(l, check.Equals, reflect.ValueOf(t.expect).Len())
		} else {
			c.Check(err, check.ErrorMatches, t.expect)
		}
	}
}
