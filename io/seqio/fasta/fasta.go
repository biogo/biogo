// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package fasta provides types to read and write FASTA format files.
package fasta

import (
	"github.com/biogo/biogo/alphabet"
	"github.com/biogo/biogo/io/seqio"
	"github.com/biogo/biogo/seq"

	"bufio"
	"bytes"
	"fmt"
	"io"
)

var (
	_ seqio.Reader = (*Reader)(nil)
	_ seqio.Writer = (*Writer)(nil)
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
	err       error
}

// Returns a new fasta format reader using f. Sequences returned by the Reader are copied
// from the provided template.
func NewReader(f io.Reader, template seqio.SequenceAppender) *Reader {
	return &Reader{
		r:         bufio.NewReader(f),
		t:         template,
		IDPrefix:  []byte(DefaultIDPrefix),
		SeqPrefix: []byte(DefaultSeqPrefix),
	}
}

// Read a single sequence and return it and potentially an error. Note that
// a non-nil returned error may be associated with a valid sequence, so it is
// the responsibility of the caller to examine the error to determine whether
// the read was successful.
// Note that if the Reader's template type returns different non-nil error
// values from calls to SetName and SetDescription, a new error string will be
// returned on each call to Read. So to allow direct error comparison these
// methods should return the same error.
func (r *Reader) Read() (seq.Sequence, error) {
	var (
		buff, line []byte
		isPrefix   bool
		s          seq.Sequence
	)
	defer func() {
		if r.working == nil {
			r.err = nil
		}
	}()

	for {
		var err error
		if buff, isPrefix, err = r.r.ReadLine(); err != nil {
			if err != io.EOF || r.working == nil {
				return nil, err
			}
			s, err = r.working, r.err
			r.working = nil
			return s, err
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
				r.working, r.err = r.header(line)
				line = nil
			} else {
				s, err = r.working, r.err
				r.working, r.err = r.header(line)
				return s, err
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
}

func (r *Reader) header(line []byte) (seqio.SequenceAppender, error) {
	s := r.t.Clone().(seqio.SequenceAppender)
	fieldMark := bytes.IndexAny(line, " \t")
	var err error
	if fieldMark < 0 {
		err = s.SetName(string(line[len(r.IDPrefix):]))
		return s, err
	} else {
		err = s.SetName(string(line[len(r.IDPrefix):fieldMark]))
		_err := s.SetDescription(string(line[fieldMark+1:]))
		if err != nil || _err != nil {
			switch {
			case err == _err:
				return s, err
			case err != nil && _err != nil:
				return s, fmt.Errorf("fasta: multiple errors: name: %s, desc:%s", err, _err)
			case err != nil:
				return s, err
			case _err != nil:
				return s, _err
			}
		}
	}

	return s, nil
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
		_n     int
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
	if err != nil {
		return n, err
	}
	for i := 0; i < s.Len(); i++ {
		if i%w.Width == 0 {
			_n, err = w.w.Write(prefix)
			if n += _n; err != nil {
				return n, err
			}
		}
		_n, err = w.w.Write([]byte{byte(s.At(i).L)})
		if n += _n; err != nil {
			return n, err
		}
	}
	_n, err = w.w.Write([]byte{'\n'})
	if n += _n; err != nil {
		return n, err
	}

	return n, nil
}
