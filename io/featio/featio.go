// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package featio provides interfaces for feature I/O functions. 
package featio

import "code.google.com/p/biogo/exp/feat"

// Reader is the common feat.Feature reader interface.
type Reader interface {
	// Read reads a feat.Feature, returning any error that occurs during the read.
	Read() (feat.Feature, error)
}

// Writer is the common feat.Feature writer interface.
type Writer interface {
	// Write write a feat.Feature, returning the number of bytes written and any
	// error that occurs during the write.
	Write(feat.Feature) (n int, err error)
}
