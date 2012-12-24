// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Packages for reading and writing sequence files
package seqio

import "code.google.com/p/biogo/seq"

type Reader interface {
	Read() (*seq.Seq, error)
	Rewind() error
	Close() error
}

type Writer interface {
	Write(*seq.Seq) (int, error)
}
