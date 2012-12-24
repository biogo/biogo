// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Packages for reading and writing sequence files
package seqio

import "code.google.com/p/biogo/exp/seq"

// A SequenceAppender is a generic sequence type that can append elements.
type SequenceAppender interface {
	SetName(string)
	SetDescription(string)
	seq.Appender
	seq.Sequence
}

type Reader interface {
	Read() (seq.Sequence, error)
}

type Writer interface {
	Write(seq.Sequence) (int, error)
}
