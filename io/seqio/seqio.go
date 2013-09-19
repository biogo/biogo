// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package seqio provides interfaces for sequence I/O functions.
package seqio

import (
	"code.google.com/p/biogo/seq"

	"io"
)

// A SequenceAppender is a generic sequence type that can append elements.
type SequenceAppender interface {
	SetName(string) error
	SetDescription(string) error
	seq.Appender
	seq.Sequence
}

// Reader is the common seq.Sequence reader interface.
type Reader interface {
	// Read reads a seq.Sequence, returning the sequence and any error that
	// occurred during the read.
	Read() (seq.Sequence, error)
}

// Writer is the common seq.Sequence writer interface.
type Writer interface {
	// Write write a seq.Sequence, returning the number of bytes written and any
	// error that occurs during the write.
	Write(seq.Sequence) (int, error)
}

// Scanner wraps a Reader to provide a convenient loop interface for reading sequence data.
// Successive calls to the Scan method will step through the sequences of the provided
// Reader. Scanning stops unrecoverably at EOF or the first error.
//
// Note that it is possible for a Reader to return a valid sequence and a non-nil error. So
// programs that need more control over error handling should use a Reader directly instead.
type Scanner struct {
	r   Reader
	seq seq.Sequence
	err error
}

// NewScanner returns a Scanner to read from r.
func NewScanner(r Reader) *Scanner { return &Scanner{r: r} }

type funcReader func() (seq.Sequence, error)

func (f funcReader) Read() (seq.Sequence, error) { return f() }

// NewScannerFromFunc returns a Scanner to read sequences returned by calls to f.
func NewScannerFromFunc(f func() (seq.Sequence, error)) *Scanner { return &Scanner{r: funcReader(f)} }

// Next advances the Scanner past the next sequence, which will then be available through
// the Seq method. It returns false when the scan stops, either by reaching the end of the
// input or an error. After Next returns false, the Err method will return any error that
// occurred during scanning, except that if it was io.EOF, Err will return nil.
func (s *Scanner) Next() bool {
	if s.err != nil {
		return false
	}
	s.seq, s.err = s.r.Read()
	return s.err == nil
}

// Err returns the first non-EOF error that was encountered by the Scanner.
func (s *Scanner) Error() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}

// Seq returns the most recent sequence read by a call to Next.
func (s *Scanner) Seq() seq.Sequence { return s.seq }
