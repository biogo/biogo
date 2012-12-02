// Copyright ©2011-2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Package to read and write FASTQ format files
package fastq

import (
	"bufio"
	"bytes"
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/exp/seqio"
	"errors"
	"io"
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

// Read a single sequence and return it or an error.
// TODO: Does not read multi-line fastq.
func (self *Reader) Read() (seq.Sequence, error) {
	var (
		buff, line, label []byte
		isPrefix          bool
		seqBuff           []alphabet.QLetter
		t                 seqio.SequenceAppender
	)

	inQual := false

	for {
		var err error
		if buff, isPrefix, err = self.r.ReadLine(); err != nil {
			return nil, err
		}
		if isPrefix {
			line = append(line, buff...)
			continue
		} else {
			line = buff
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		switch {
		case !inQual && line[0] == '@':
			t = self.readHeader(line)
			label, line = line, nil
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
				seqBuff[i].Q = self.enc.DecodeToQphred(line[i])
			}
			t.AppendQLetters(seqBuff...)

			return t, nil
		}
	}

	panic("cannot reach")
}

func (self *Reader) readHeader(line []byte) seqio.SequenceAppender {
	s := self.t.Copy().(seqio.SequenceAppender)
	fieldMark := bytes.IndexAny(line, " \t")
	if fieldMark < 0 {
		s.SetName(string(line[1:]))
	} else {
		s.SetName(string(line[1:fieldMark]))
		s.SetDescription(string(line[fieldMark+1:]))
	}

	return s
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
func (self *Writer) Write(s seq.Sequence) (n int, err error) {
	var (
		ln  int
		enc alphabet.Encoding
	)
	if e, ok := s.(Encoder); ok {
		enc = e.Encoding()
	} else {
		enc = alphabet.Sanger
	}

	n, err = self.writeHeader('@', s)
	if err != nil {
		return
	}
	for i := 0; i < s.Len(); i++ {
		ln, err = self.w.Write([]byte{byte(s.At(i).L)})
		if n += ln; err != nil {
			return
		}
	}
	ln, err = self.w.Write([]byte{'\n'})
	if n += ln; err != nil {
		return
	}
	if self.QID {
		ln, err = self.writeHeader('+', s)
		if n += ln; err != nil {
			return
		}
	} else {
		ln, err = self.w.Write([]byte("+\n"))
		if n += ln; err != nil {
			return
		}
	}
	for i := 0; i < s.Len(); i++ {
		ln, err = self.w.Write([]byte{s.At(i).Q.Encode(enc)})
		if n += ln; err != nil {
			return
		}
	}
	ln, err = self.w.Write([]byte{'\n'})
	if n += ln; err != nil {
		return
	}

	return
}

func (self *Writer) writeHeader(prefix byte, s seq.Sequence) (n int, err error) {
	var ln int
	n, err = self.w.Write([]byte{prefix})
	if err != nil {
		return
	}
	ln, err = io.WriteString(self.w, s.Name())
	if n += ln; err != nil {
		return
	}
	if desc := s.Description(); len(desc) != 0 {
		ln, err = self.w.Write([]byte{' '})
		if n += ln; err != nil {
			return
		}
		ln, err = io.WriteString(self.w, desc)
		if n += ln; err != nil {
			return
		}
	}
	ln, err = self.w.Write([]byte("\n"))
	n += ln
	return
}
