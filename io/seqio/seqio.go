// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package seqio provides interfaces for sequence I/O functions.
package seqio

import "code.google.com/p/biogo/seq"

// A SequenceAppender is a generic sequence type that can append elements.
type SequenceAppender interface {
	SetName(string) error
	SetDescription(string) error
	seq.Appender
	seq.Sequence
}

// Reader is the common seq.Sequence reader interface.
type Reader interface {
	// Read reads a seq.Sequence, returning any error that occurs during the read.
	Read() (seq.Sequence, error)
}

// Writer is the common seq.Sequence writer interface.
type Writer interface {
	// Write write a seq.Sequence, returning the number of bytes written and any
	// error that occurs during the write.
	Write(seq.Sequence) (int, error)
}
