// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package matrix

import (
	"fmt"
	"strconv"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Format is matrix a formatting function.
func Format(m Matrix, margin int, fs fmt.State, c rune) {
	var (
		mw         int
		b, padding []byte
		pc         int
		rows, cols = m.Dims()
		prec, pOk  = fs.Precision()
		width, _   = fs.Width()
	)
	if margin <= 0 {
		pc = rows
		if cols > pc {
			pc = cols
		}
	}
	if !pOk {
		prec = -1
	}

	switch c {
	case 'v', 'e', 'E', 'f', 'F', 'g', 'G':
		// Note that the '#' flag should have been dealt with by the type.
		// So %v is treated exactly as %g here.
		b, mw = maxCellWidth(m, c, pc, prec)
	default:
		fmt.Fprintf(fs, "%%!%c(%T=Dims(%d, %d))", c, m, rows, cols)
		return
	}
	width = max(width, mw)
	padding = make([]byte, max(width, 2))
	for i := range padding {
		padding[i] = ' '
	}
	skipZero := fs.Flag('#')

	if rows > 2*pc || cols > 2*pc {
		fmt.Fprintf(fs, "Dims(%d, %d)\n", rows, cols)
	}

	for i := 0; i < rows; i++ {
		var el string
		switch {
		case rows == 1:
			fmt.Fprint(fs, "[")
			el = "]"
		case i == 0:
			fmt.Fprint(fs, "⎡")
			el = "⎤\n"
		case i < rows-1:
			fmt.Fprint(fs, "⎢")
			el = "⎥\n"
		default:
			fmt.Fprint(fs, "⎣")
			el = "⎦"
		}

		for j := 0; j < cols; j++ {
			if j >= pc && j < cols-pc {
				j = cols - pc - 1
				if i == 0 || i == rows-1 {
					fmt.Fprint(fs, "...  ...  ")
				} else {
					fmt.Fprint(fs, "          ")
				}
				continue
			}

			v := m.At(i, j)
			if v == 0 && skipZero {
				b = b[:1]
				b[0] = '.'
			} else {
				if c == 'v' {
					b = strconv.AppendFloat(b[:0], v, 'g', prec, 64)
				} else {
					b = strconv.AppendFloat(b[:0], v, byte(c), prec, 64)
				}
			}
			if fs.Flag('-') {
				fs.Write(b)
				fs.Write(padding[:width-len(b)])
			} else {
				fs.Write(padding[:width-len(b)])
				fs.Write(b)
			}

			if j < cols-1 {
				fs.Write(padding[:2])
			}
		}

		fmt.Fprint(fs, el)

		if i >= pc-1 && i < rows-pc && 2*pc < rows {
			i = rows - pc - 1
			fmt.Fprint(fs, " .\n .\n .\n")
			continue
		}
	}
}

func maxCellWidth(m Matrix, c rune, pc, prec int) ([]byte, int) {
	var (
		b          = make([]byte, 0, 64)
		rows, cols = m.Dims()
		max        int
	)
	for i := 0; i < rows; i++ {
		if i >= pc-1 && i < rows-pc && 2*pc < rows {
			i = rows - pc - 1
			continue
		}
		for j := 0; j < cols; j++ {
			if j >= pc && j < cols-pc {
				continue
			}

			b = strconv.AppendFloat(b, m.At(i, j), byte(c), prec, 64)
			if len(b) > max {
				max = len(b)
			}
			b = b[:0]
		}
	}
	return b, max
}
