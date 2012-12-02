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
func (self *Reader) Read() (seq.Sequence, error) {
	var (
		buff, line []byte
		isPrefix   bool
		s          seq.Sequence
	)

	for {
		var err error
		if buff, isPrefix, err = self.r.ReadLine(); err != nil {
			if err != io.EOF || self.working == nil {
				return nil, err
			}
			s = self.working
			self.working = nil
			return s, nil
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

		if bytes.HasPrefix(line, self.IDPrefix) {
			if self.working == nil {
				self.working = self.header(line)
				line = nil
			} else {
				s = self.working
				self.working = self.header(line)
				return s, nil
			}
		} else if bytes.HasPrefix(line, self.SeqPrefix) {
			line = bytes.Join(bytes.Fields(line[len(self.SeqPrefix):]), nil)
			self.working.AppendLetters(alphabet.BytesToLetters(line)...)
			line = nil
		} else {
			return nil, fmt.Errorf("fasta: badly formed line %d", line)
		}
	}

	panic("cannot reach")
}

func (self *Reader) header(line []byte) seqio.SequenceAppender {
	s := self.t.Copy().(seqio.SequenceAppender)
	fieldMark := bytes.IndexAny(line, " \t")
	if fieldMark < 0 {
		s.SetName(string(line[len(self.IDPrefix):]))
	} else {
		s.SetName(string(line[len(self.IDPrefix):fieldMark]))
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
func (self *Writer) Write(s seq.Sequence) (n int, err error) {
	var (
		ln     int
		prefix = append([]byte{'\n'}, self.SeqPrefix...)
	)
	id, desc := s.Name(), s.Description()
	header := make([]byte, 0, len(self.IDPrefix)+len(id)+len(desc)+1)
	header = append(header, self.IDPrefix...)
	header = append(header, id...)
	if len(desc) > 0 {
		header = append(header, ' ')
		header = append(header, desc...)
	}

	n, err = self.w.Write(header)
	if err == nil {
		for i := 0; i < s.Len(); i++ {
			if i%self.Width == 0 {
				ln, err = self.w.Write(prefix)
				if n += ln; err != nil {
					return
				}
			}
			ln, err = self.w.Write([]byte{byte(s.At(i).L)})
			if n += ln; err != nil {
				return
			}
		}
		ln, err = self.w.Write([]byte{'\n'})
		if n += ln; err != nil {
			return
		}
	}

	return
}
