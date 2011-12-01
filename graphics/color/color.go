// Hue Saturation Value Alpha color package
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
	"math"
)

const maxUintF = float64(^uint32(0))

type HSVAColor struct {
	H, S, V, A float32
}

func HSVA(c color.Color) HSVAColor {
	r, g, b, a := c.RGBA()
	red := float64(r)
	blue := float64(b)
	green := float64(g)

	max := math.Max(red, green)
	max = math.Max(max, blue)
	min := math.Min(red, green)
	min = math.Min(min, blue)
	chroma := max - min

	var hue float64
	switch {
	case chroma == 0:
		hue = math.NaN()
	case max == red:
		hue = math.Mod((green-blue)/chroma, 6)
	case max == green:
		hue = (blue-red)/chroma + 2
	case max == blue:
		hue = (red-green)/chroma + 4
	}

	hue *= 60

	return HSVAColor{H: float32(hue), S: float32(chroma / max), V: float32(max), A: float32(float64(a) / maxUintF)}
}

func (self *HSVAColor) RGBA() (r, g, b, a uint32) {
	var red, green, blue float64
	H := float64(self.H)
	S := float64(self.S)
	V := float64(self.V)
	A := float64(self.A)

	a = uint32(maxUintF * A)

	if V == 0 || H == math.NaN() {
		return
	}

	if S == 0 {
		r, g, b = uint32(maxUintF*V), uint32(maxUintF*V), uint32(maxUintF*V)
		return
	}

	chroma := V * S
	hue := math.Mod(H, 360) / 60
	x := chroma * (1 - math.Abs(math.Mod(hue, 2)-1))

	switch math.Floor(hue) {
	case 0:
		red, green = chroma, x
	case 1:
		red, green = x, chroma
	case 2:
		green, blue = chroma, x
	case 3:
		green, blue = x, chroma
	case 4:
		red, blue = x, chroma
	case 5:
		red, blue = chroma, x
	}

	m := V - chroma

	r, g, b = uint32(maxUintF*(red+m)), uint32(maxUintF*(green+m)), uint32(maxUintF*(blue+m))

	return
}
