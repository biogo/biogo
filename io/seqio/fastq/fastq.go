// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package fastq provides types to read and write FASTQ format files.
package fastq

import (
	"code.google.com/p/biogo/alphabet"
	"code.google.com/p/biogo/io/seqio"
	"code.google.com/p/biogo/seq"

	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
)

var (
	_ seqio.Reader = (*Reader)(nil)
	_ seqio.Writer = (*Writer)(nil)
)

type Encoder interface {
	Encoding() alphabet.Encoding
}

// Fastq sequence format reader type.
type Reader struct {
	r   *bufio.Reader
	t   seqio.SequenceAppender
	enc alphabet.Encoding
}

// Returns a new fastq format reader using r. Sequences returned by the Reader are copied
// from the provided template.
func NewReader(r io.Reader, template seqio.SequenceAppender) *Reader {
	var enc alphabet.Encoding
	if e, ok := template.(Encoder); ok {
		enc = e.Encoding()
	} else {
		enc = alphabet.None
	}

	return &Reader{
		r:   bufio.NewReader(r),
		t:   template,
		enc: enc,
	}
}

// Read a single sequence and return it  and potentially an error. Note that
// a non-nil returned error may be associated with a valid sequence, so it is
// the responsibility of the caller to examine the error to determine whether
// the read was successful.
// Note that if the Reader's template type returns different non-nil error
// values from calls to SetName and SetDescription, a new error string will be
// returned on each call to Read. So to allow direct error comparison these
// methods should return the same error.
// TODO: Does not read multi-line fastq.
func (r *Reader) Read() (seq.Sequence, error) {
	var (
		buff, line, label []byte
		isPrefix          bool
		seqBuff           []alphabet.QLetter
		t                 seqio.SequenceAppender
	)

	inQual := false

	var err error
	for {
		if buff, isPrefix, err = r.r.ReadLine(); err != nil {
			return nil, err
		}
		line = append(line, buff...)
		if isPrefix {
			continue
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		switch {
		case !inQual && line[0] == '@':
			var _err error
			t, _err = r.readHeader(line)
			if err == nil && _err != nil {
				err = _err
			}
			label = line
		case !inQual && line[0] == '+':
			if len(label) == 0 {
				return nil, errors.New("fastq: no header line parsed before +line in fastq format")
			}
			if len(line) > 1 && bytes.Compare(label[1:], line[1:]) != 0 {
				return nil, errors.New("fastq: quality header does not match sequence header")
			}
			inQual = true
		case !inQual:
			line = bytes.Join(bytes.Fields(line), nil)
			seqBuff = make([]alphabet.QLetter, len(line))
			for i := range line {
				seqBuff[i].L = alphabet.Letter(line[i])
			}
		case inQual:
			line = bytes.Join(bytes.Fields(line), nil)
			if len(line) != len(seqBuff) {
				return nil, errors.New("fastq: sequence/quality length mismatch")
			}
			for i := range line {
				seqBuff[i].Q = r.enc.DecodeToQphred(line[i])
			}
			t.AppendQLetters(seqBuff...)

			return t, err
		}
		line = nil
	}
}

func (r *Reader) readHeader(line []byte) (seqio.SequenceAppender, error) {
	s := r.t.Clone().(seqio.SequenceAppender)
	fieldMark := bytes.IndexAny(line, " \t")
	var err error
	if fieldMark < 0 {
		err = s.SetName(string(line[1:]))
		return s, err
	} else {
		err = s.SetName(string(line[1:fieldMark]))
		_err := s.SetDescription(string(line[fieldMark+1:]))
		if err != nil || _err != nil {
			switch {
			case err == _err:
				return s, err
			case err != nil && _err != nil:
				return s, fmt.Errorf("fastq: multiple errors: name: %s, desc:%s", err, _err)
			case err != nil:
				return s, err
			case _err != nil:
				return s, _err
			}
		}
	}

	return s, nil
}

// Fastq sequence format writer type.
type Writer struct {
	w   io.Writer
	QID bool // Include ID on +lines
}

// Returns a new fastq format writer using w.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w: w,
	}
}

// Write a single sequence and return the number of bytes written and any error.
func (w *Writer) Write(s seq.Sequence) (n int, err error) {
	var (
		_n  int
		enc alphabet.Encoding
	)
	if e, ok := s.(Encoder); ok {
		enc = e.Encoding()
	} else {
		enc = alphabet.Sanger
	}

	n, err = w.writeHeader('@', s)
	if err != nil {
		return
	}
	for i := 0; i < s.Len(); i++ {
		_n, err = w.w.Write([]byte{byte(s.At(i).L)})
		if n += _n; err != nil {
			return
		}
	}
	_n, err = w.w.Write([]byte{'\n'})
	if n += _n; err != nil {
		return
	}
	if w.QID {
		_n, err = w.writeHeader('+', s)
		if n += _n; err != nil {
			return
		}
	} else {
		_n, err = w.w.Write([]byte("+\n"))
		if n += _n; err != nil {
			return
		}
	}
	for i := 0; i < s.Len(); i++ {
		_n, err = w.w.Write([]byte{s.At(i).Q.Encode(enc)})
		if n += _n; err != nil {
			return
		}
	}
	_n, err = w.w.Write([]byte{'\n'})
	if n += _n; err != nil {
		return
	}

	return
}

func (w *Writer) writeHeader(prefix byte, s seq.Sequence) (n int, err error) {
	var _n int
	n, err = w.w.Write([]byte{prefix})
	if err != nil {
		return
	}
	_n, err = io.WriteString(w.w, s.Name())
	if n += _n; err != nil {
		return
	}
	if desc := s.Description(); len(desc) != 0 {
		_n, err = w.w.Write([]byte{' '})
		if n += _n; err != nil {
			return
		}
		_n, err = io.WriteString(w.w, desc)
		if n += _n; err != nil {
			return
		}
	}
	_n, err = w.w.Write([]byte("\n"))
	n += _n
	return
}
