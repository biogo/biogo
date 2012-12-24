// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Commonly used functions
package util

// A CTL hold a code to letter lookup table.
// TODO: Replace this with the functionality provided by alphabet.
type CTL struct {
	ValueToCode [256]int
}

// Inititialise and return a CTL based on a map m.
func NewCTL(m map[int]int) (t *CTL) {
	t = &CTL{}
	for i := 0; i < 256; i++ {
		if code, present := m[i]; present {
			t.ValueToCode[i] = code
		} else {
			t.ValueToCode[i] = -1
		}
	}

	return
}
