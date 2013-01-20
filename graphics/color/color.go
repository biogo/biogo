// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Hue Saturation Value Alpha color package
package color

import (
	"image/color"
	"math"
)

const maxChannelValue = float64(0xFFFF)

// HSVAModel converts any color.Color to an HSVA color.
var HSVAModel color.Model = color.ModelFunc(hsvaModel)

func hsvaModel(c color.Color) color.Color {
	if _, ok := c.(HSVA); ok {
		return c
	}
	return RGBAtoHSVA(c.RGBA())
}

// HSVAColor represents a Hue/Saturation/Value/Alpha color.
// H is valid within [0°, 360°]. S, V and A are valid within [0, 1].
type HSVA struct {
	H, S, V, A float64
}

// Convert r, g, b, a to HSVA
func RGBAtoHSVA(r, g, b, a uint32) HSVA {
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
		hue = 0 // should really be math.NaN() since we have a 0 length vector, but 0 seems to be the convention and it may simplify imports in dependent packages
	case max == red:
		hue = math.Mod((green-blue)/chroma, 6)
	case max == green:
		hue = (blue-red)/chroma + 2
	case max == blue:
		hue = (red-green)/chroma + 4
	}

	hue *= 60

	var s float64
	if chroma != 0 {
		s = chroma / max
	}

	return HSVA{
		H: math.Mod(math.Mod(hue, 360)+360, 360),
		S: s,
		V: max / maxChannelValue,
		A: float64(a) / maxChannelValue,
	}
}

// RGBA() allows HSVAColor to satisfy the color.Color interface.
func (c HSVA) RGBA() (r, g, b, a uint32) {
	var red, green, blue float64

	a = uint32(maxChannelValue * c.A)

	if c.V == 0 {
		return
	}

	if c.S == 0 {
		r, g, b = uint32(maxChannelValue*c.V), uint32(maxChannelValue*c.V), uint32(maxChannelValue*c.V)
		return
	}

	chroma := c.V * c.S
	m := c.V - chroma

	if !math.IsNaN(c.H) {
		hue := math.Mod(c.H, 360) / 60
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
	} else {
		red, green, blue = 0, 0, 0
	}

	r, g, b = uint32(maxChannelValue*(red+m)), uint32(maxChannelValue*(green+m)), uint32(maxChannelValue*(blue+m))

	return
}
