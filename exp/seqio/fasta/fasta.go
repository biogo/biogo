// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package to read and write FASTA format files
package fasta

import (
	"bufio"
	"bytes"
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/exp/seqio"
	"fmt"
	"io"
)

// Default delimiters.
const (
	DefaultIDPrefix  = ">"
	DefaultSeqPrefix = ""
)

// Fasta sequence format reader type.
type Reader struct {
	r         *bufio.Reader
	t         seqio.SequenceAppender
	IDPrefix  []byte
	SeqPrefix []byte
	working   seqio.SequenceAppender
}

// Returns a new fasta format reader using f. Sequences returned by the Reader are copied
// from the provided template.
func NewReader(f io.Reader, template seqio.SequenceAppender) *Reader {
	return &Reader{
		r:         bufio.NewReader(f),
		t:         template,
		IDPrefix:  []byte(DefaultIDPrefix),
		SeqPrefix: []byte(DefaultSeqPrefix),
		working:   nil,
	}
}

// Read a single sequence and return it or an error.
func (r *Reader) Read() (seq.Sequence, error) {
	var (
		buff, line []byte
		isPrefix   bool
		s          seq.Sequence
	)

	for {
		var err error
		if buff, isPrefix, err = r.r.ReadLine(); err != nil {
			if err != io.EOF || r.working == nil {
				return nil, err
			}
			s = r.working
			r.working = nil
			return s, nil
		}
		line = append(line, buff...)
		if isPrefix {
			continue
		}
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		if bytes.HasPrefix(line, r.IDPrefix) {
			if r.working == nil {
				r.working = r.header(line)
				line = nil
			} else {
				s = r.working
				r.working = r.header(line)
				return s, nil
			}
		} else if bytes.HasPrefix(line, r.SeqPrefix) {
			if r.working == nil {
				return nil, fmt.Errorf("fasta: badly formed line %q", line)
			}
			line = bytes.Join(bytes.Fields(line[len(r.SeqPrefix):]), nil)
			r.working.AppendLetters(alphabet.BytesToLetters(line)...)
			line = nil
		} else {
			return nil, fmt.Errorf("fasta: badly formed line %q", line)
		}
	}

	panic("cannot reach")
}

func (r *Reader) header(line []byte) seqio.SequenceAppender {
	s := r.t.Clone().(seqio.SequenceAppender)
	fieldMark := bytes.IndexAny(line, " \t")
	if fieldMark < 0 {
		s.SetName(string(line[len(r.IDPrefix):]))
	} else {
		s.SetName(string(line[len(r.IDPrefix):fieldMark]))
		s.SetDescription(string(line[fieldMark+1:]))
	}

	return s
}

// Fasta sequence format writer type.
type Writer struct {
	w         io.Writer
	IDPrefix  []byte
	SeqPrefix []byte
	Width     int
}

// Returns a new fasta format writer using f.
func NewWriter(w io.Writer, width int) *Writer {
	return &Writer{
		w:         w,
		IDPrefix:  []byte(DefaultIDPrefix),
		SeqPrefix: []byte(DefaultSeqPrefix),
		Width:     width,
	}
}

// Write a single sequence and return the number of bytes written and any error.
func (w *Writer) Write(s seq.Sequence) (n int, err error) {
	var (
		ln     int
		prefix = append([]byte{'\n'}, w.SeqPrefix...)
	)
	id, desc := s.Name(), s.Description()
	header := make([]byte, 0, len(w.IDPrefix)+len(id)+len(desc)+1)
	header = append(header, w.IDPrefix...)
	header = append(header, id...)
	if len(desc) > 0 {
		header = append(header, ' ')
		header = append(header, desc...)
	}

	n, err = w.w.Write(header)
	if err == nil {
		for i := 0; i < s.Len(); i++ {
			if i%w.Width == 0 {
				ln, err = w.w.Write(prefix)
				if n += ln; err != nil {
					return
				}
			}
			ln, err = w.w.Write([]byte{byte(s.At(i).L)})
			if n += ln; err != nil {
				return
			}
		}
		ln, err = w.w.Write([]byte{'\n'})
		if n += ln; err != nil {
			return
		}
	}

	return
}
