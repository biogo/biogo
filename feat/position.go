// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package feat

// Convert from 1-based to 0-based indexing
func OneToZero(pos int) int {
	if pos == 0 {
		panic("feat: 1-based index == 0")
	}
	if pos > 0 {
		pos--
	}

	return pos
}

// Convert from 0-based to 1-based indexing
func ZeroToOne(pos int) int {
	if pos >= 0 {
		pos++
	}

	return pos
}
