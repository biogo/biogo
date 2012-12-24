// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package for reading and writing multiple sequence alignment files
package alignio

import (
	"code.google.com/p/biogo/io/seqio"
	"code.google.com/p/biogo/seq"
	"io"
)

// Reader implements multiple sequence reading from a seqio.Reader.
type Reader struct {
	seqio.Reader
}

// Return a new Reader.
func NewReader(r seqio.Reader) *Reader {
	return &Reader{r}
}

// Read the contents of the embedded seqio.Reader into a seq.Alignment.
// Returns the Alignment, or nil and an error if an error occurs.
func (self *Reader) Read() (a seq.Alignment, err error) {
	var s *seq.Seq
	a = seq.Alignment{}
	for {
		s, err = self.Reader.Read()
		if err == nil {
			a = append(a, s)
		} else {
			if err == io.EOF {
				return a, nil
			} else {
				return nil, err
			}
		}
	}

	panic("cannot reach")
}

// Writer implements multiple sequence writing to a seqio.Writer.
type Writer struct {
	seqio.Writer
}

// Return a new Writer.
func NewWriter(w seqio.Writer) *Writer {
	return &Writer{w}
}

// Write a seq.Alignment to the embedded seqio.Reader.
// Returns the number of bytes written and any error. 
func (self *Writer) Write(a seq.Alignment) (n int, err error) {
	var c int
	for _, s := range a {
		c, err = self.Writer.Write(s)
		n += c
		if err != nil {
			return
		}
	}

	panic("cannot reach")
}
