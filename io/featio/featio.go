// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package featio provides interfaces for feature I/O functions.
package featio

import (
	"code.google.com/p/biogo/feat"

	"io"
)

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

// Scanner wraps a Reader to provide a convenient loop interface for reading feature data.
// Successive calls to the Scan method will step through the features of the provided
// Reader. Scanning stops unrecoverably at EOF or the first error.
//
// Note that it is possible for a Reader to return a valid feature and a non-nil error. So
// programs that need more control over error handling should use a Reader directly instead.
type Scanner struct {
	r   Reader
	f   feat.Feature
	err error
}

// NewScanner returns a Scanner to read from r.
func NewScanner(r Reader) *Scanner { return &Scanner{r: r} }

// Scan advances the Scanner past the next feature, which will then be available through
// the Feat method. It returns false when the scan stops, either by reaching the end of the
// input or an error. After Scan returns false, the Err method will return any error that
// occurred during scanning, except that if it was io.EOF, Err will return nil.
func (s *Scanner) Scan() bool {
	if s.err != nil {
		return false
	}
	s.f, s.err = s.r.Read()
	return s.err == nil
}

// Err returns the first non-EOF error that was encountered by the Scanner.
func (s *Scanner) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}

// Feat returns the most recent feature read by a call to Scan.
func (s *Scanner) Feat() feat.Feature { return s.f }
