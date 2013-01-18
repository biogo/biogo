// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package alignio for reading and writing multiple sequence alignment files
package alignio

import (
	"code.google.com/p/biogo/exp/seq/multi"
	"code.google.com/p/biogo/exp/seqio"
	"io"
)

// Reader implements multiple sequence reading from a seqio.Reader.
type Reader struct {
	r seqio.Reader
	m *multi.Multi
}

// NewReader return a new Reader that will read sequences from r into m.
func NewReader(r seqio.Reader, m *multi.Multi) *Reader {
	return &Reader{r, m}
}

// Read the contents of the embedded seqio.Reader into a seq.Sequence.
// Returns the Alignment, or nil and an error if an error occurs.
func (r *Reader) Read() (*multi.Multi, error) {
	for {
		s, err := r.r.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		r.m.Add(s)
	}

	m, t := r.m, *r.m
	t.Seq = nil
	r.m = &t

	return m, nil
}

// Writer implements multiple sequence writing to a seqio.Writer.
type Writer struct {
	w seqio.Writer
}

// Return a new Writer.
func NewWriter(w seqio.Writer) *Writer {
	return &Writer{w}
}

// Write a multi.Multi to the embedded seqio.Writer.
// Returns the number of bytes written and any error. 
func (w *Writer) Write(m *multi.Multi) (n int, err error) {
	var c int
	for _, s := range m.Seq {
		c, err = w.w.Write(s)
		n += c
		if err != nil {
			break
		}
	}

	return
}
