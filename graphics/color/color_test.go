package color

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
	"image/color"
	check "launchpad.net/gocheck"
	"testing"
)

var e uint32 = 1

// Checkers
type float32WithinRange struct {
	*check.CheckerInfo
}

var floatWithinRange check.Checker = &float32WithinRange{
	&check.CheckerInfo{Name: "WithinRange", Params: []string{"obtained", "min", "max"}},
}

func (checker *float32WithinRange) Check(params []interface{}, names []string) (result bool, error string) {
	return params[0].(float64) >= params[1].(float64) && params[0].(float64) <= params[2].(float64), ""
}

type uint32WithinRange struct {
	*check.CheckerInfo
}

var uintWithinRange check.Checker = &uint32WithinRange{
	&check.CheckerInfo{Name: "WithinRange", Params: []string{"obtained", "min", "max"}},
}

func (checker *uint32WithinRange) Check(params []interface{}, names []string) (result bool, error string) {
	return params[0].(uint32) >= params[1].(uint32) && params[0].(uint32) <= params[2].(uint32), ""
}

type uint32EpsilonChecker struct {
	*check.CheckerInfo
}

var withinEpsilon check.Checker = &uint32EpsilonChecker{
	&check.CheckerInfo{Name: "EpsilonLessThan", Params: []string{"obtained", "expected", "error"}},
}

func (checker *uint32EpsilonChecker) Check(params []interface{}, names []string) (result bool, error string) {
	d := int64(params[0].(uint32)) - int64(params[1].(uint32))
	if d < 0 {
		if d == -d {
			panic("color: weird number overflow")
		}
		d = -d
	}
	return uint32(d) <= params[2].(uint32), ""
}

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

func (s *S) TestColor(c *check.C) {
	for r := 0; r < 256; r += 5 {
		for g := 0; g < 256; g += 5 {
			for b := 0; b < 256; b += 5 {
				col := color.RGBA{uint8(r), uint8(g), uint8(b), 0}
				cDirectR, cDirectG, cDirectB, cDirectA := col.RGBA()
				hsva := RGBAtoHSVA(col.RGBA())
				c.Check(hsva.H, floatWithinRange, float64(0), float64(360))
				c.Check(hsva.S, floatWithinRange, float64(0), float64(1))
				c.Check(hsva.V, floatWithinRange, float64(0), float64(1))
				cFromHSVR, cFromHSVG, cFromHSVB, cFromHSVA := hsva.RGBA()
				c.Check(cFromHSVR, uintWithinRange, uint32(0), uint32(0xFFFF))
				c.Check(cFromHSVG, uintWithinRange, uint32(0), uint32(0xFFFF))
				c.Check(cFromHSVB, uintWithinRange, uint32(0), uint32(0xFFFF))
				back := RGBAtoHSVA(color.RGBA{uint8(cFromHSVR >> 8), uint8(cFromHSVG >> 8), uint8(cFromHSVB >> 8), uint8(cFromHSVA)}.RGBA())
				c.Check(hsva, check.Equals, back)
				c.Check(cFromHSVR, withinEpsilon, cDirectR, e)
				c.Check(cFromHSVG, withinEpsilon, cDirectG, e)
				c.Check(cFromHSVB, withinEpsilon, cDirectB, e)
				c.Check(cFromHSVA, check.Equals, cDirectA)
				if c.Failed() {
					return
				}
			}
		}
	}
}
