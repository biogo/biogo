// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package alphabet

import (
	"fmt"
	check "launchpad.net/gocheck"
	"math"
)

type approxChecker struct {
	*check.CheckerInfo
}

var approx check.Checker = &approxChecker{
	&check.CheckerInfo{Name: "Approx", Params: []string{"obtained", "expected", "epsilon"}},
}

func (checker *approxChecker) Check(params []interface{}, names []string) (result bool, error string) {
	defer func() {
		if v := recover(); v != nil {
			result = false
			error = fmt.Sprint(v)
		}
	}()
	return math.Abs(params[0].(float64)-params[1].(float64)) <= params[2].(float64)*params[1].(float64), ""
}

// Tests
func (s *S) TestPhred(c *check.C) {
	// Confirm landmarks.
	for _, t := range []struct {
		E float64
		Q Qphred
	}{
		{E: 1e-1, Q: 10},
		{E: 1e-2, Q: 20},
		{E: 1e-3, Q: 30},
		{E: 1e-4, Q: 40},
		{E: 1e-5, Q: 50},
	} {
		c.Check(Ephred(t.E), check.Equals, t.Q)
		c.Check(t.Q.ProbE(), check.Equals, t.E)
	}
	for q := Qphred(0); q < 254; q++ {
		c.Check(q.ProbE(), check.Equals, math.Pow(10, -(float64(q)/10)))
		c.Check(Ephred(q.ProbE()), check.Equals, q)
	}
	c.Check(Qphred(254).ProbE(), check.Equals, 0.)
	c.Check(math.IsNaN(Qphred(255).ProbE()), check.Equals, true)
}

func (s *S) TestSolexa(c *check.C) {
	// Confirm landmarks.
	for _, t := range []struct {
		E float64
		Q Qsolexa
	}{
		{E: 1e-1 / (1 + 1e-1), Q: 10},
		{E: 1e-2 / (1 + 1e-2), Q: 20},
		{E: 1e-3 / (1 + 1e-3), Q: 30},
		{E: 1e-4 / (1 + 1e-4), Q: 40},
		{E: 1e-5 / (1 + 1e-5), Q: 50},
	} {
		c.Check(Esolexa(t.E), check.Equals, t.Q)
		c.Check(t.Q.ProbE(), approx, t.E, 1e-15)
	}
	c.Check(math.IsNaN(Qsolexa(-128).ProbE()), check.Equals, true)
	for q := -127; q < 127; q++ {
		pq := math.Pow(10, -(float64(q) / 10))
		pq /= (1 + pq)
		c.Check(Qsolexa(q).ProbE(), check.Equals, pq)
		c.Check(Esolexa(Qsolexa(q).ProbE()), check.Equals, Qsolexa(q))
	}
	c.Check(Qsolexa(127).ProbE(), check.Equals, 0.)
}

func (s *S) TestInterconversion(c *check.C) {
	for q := 0; q < 127; q++ {
		if 10 <= q && q < 127 {
			c.Check(Qphred(q).Qsolexa().Qphred(), check.Equals, Qphred(q))
			c.Check(Qsolexa(q).Qphred().Qsolexa(), check.Equals, Qsolexa(q))
		}
		c.Check(Qphred(q).Qsolexa().ProbE(), approx, Qphred(q).ProbE(), math.Pow(10, 1e-4-float64(q)/10),
			check.Commentf("Test %d at E = %e", q, Qphred(q).ProbE()))
		c.Check(Qsolexa(q).Qphred().ProbE(), approx, Qsolexa(q).ProbE(), math.Pow(10, 1e-4-float64(q)/10),
			check.Commentf("Test %d at E = %e", q, Qsolexa(q).ProbE()))
	}
}
