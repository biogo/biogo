// Package to read and write FASTA format files
package fasta

// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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

import (
	"bufio"
	"bytes"
	"github.com/kortschak/biogo/bio"
	"github.com/kortschak/biogo/seq"
	"github.com/kortschak/biogo/util"
	"io"
	"os"
)

const (
	IDPrefix  = ">" // default delimiters
	SeqPrefix = ""  // default delimiters
)

// Fasta sequence format reader type.
type Reader struct {
	f         io.ReadCloser
	r         *bufio.Reader
	IDPrefix  []byte
	SeqPrefix []byte
	last      []byte
	line      int
}

// Returns a new fasta format reader using f.
func NewReader(f io.ReadCloser) *Reader {
	return &Reader{
		f:         f,
		r:         bufio.NewReader(f),
		IDPrefix:  []byte(IDPrefix),
		SeqPrefix: []byte(SeqPrefix),
		last:      nil,
	}
}

// Returns a new fasta format reader using a filename.
func NewReaderName(name string) (r *Reader, err error) {
	f, err := os.Open(name)
	if err != nil {
		return
	}
	return NewReader(f), nil
}

// Read a single sequence and return it or an error.
func (self *Reader) Read() (sequence *seq.Seq, err error) {
	var line, label, body []byte
	label = self.last

READ:
	for {
		line, err = self.r.ReadBytes('\n')
		if err == nil {
			self.line++
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			switch {
			case bytes.HasPrefix(line, self.IDPrefix):
				if self.last == nil {
					self.last = line[len(self.IDPrefix):]
				} else {
					label = self.last
					self.last = line[len(self.IDPrefix):] // entering a new sequence so exit read loop
					break READ
				}
			case bytes.HasPrefix(line, self.SeqPrefix):
				line = bytes.Join(bytes.Fields(line[len(self.SeqPrefix):]), nil)
				body = append(body, line...)
			}
		} else {
			if self.last != nil {
				self.last = nil
				err = nil
				break
			} else {
				return nil, io.EOF
			}
		}
	}

	if len(label) > 0 || len(body) > 0 {
		sequence = seq.New(string(label), body, nil)
	} else {
		err = bio.NewError("fasta: empty sequence", 0, self.line)
	}

	return
}

// Rewind the reader.
func (self *Reader) Rewind() (err error) {
	if s, ok := self.f.(io.Seeker); ok {
		self.last = nil
		_, err = s.Seek(0, 0)
		self.r = bufio.NewReader(self.f)
	} else {
		err = bio.NewError("Not a Seeker", 0, self)
	}
	return
}

// Close the reader.
func (self *Reader) Close() (err error) {
	return self.f.Close()
}

// Fasta sequence format writer type.
type Writer struct {
	f         io.WriteCloser
	w         *bufio.Writer
	IDPrefix  []byte
	SeqPrefix []byte
	Width     int
}

// Returns a new fasta format writer using f.
func NewWriter(f io.WriteCloser, width int) *Writer {
	return &Writer{
		f:         f,
		w:         bufio.NewWriter(f),
		IDPrefix:  []byte(IDPrefix),
		SeqPrefix: []byte(SeqPrefix),
		Width:     width,
	}
}

// Returns a new fasta format writer using a filename, truncating any existing file.
// If appending is required use NewWriter and os.OpenFile.
func NewWriterName(name string, width int) (w *Writer, err error) {
	f, err := os.Create(name)
	if err != nil {
		return
	}
	return NewWriter(f, width), nil
}

// Write a single sequence and return the number of bytes written and any error.
func (self *Writer) Write(s *seq.Seq) (n int, err error) {
	var ln int
	n, err = self.w.WriteString(string(self.IDPrefix) + s.ID + "\n")
	if err == nil {
		for i := 0; i*self.Width <= s.Len(); i++ {
			endLinePos := util.Min(self.Width*(i+1), s.Len())
			for _, elem := range [][]byte{self.SeqPrefix, s.Seq[self.Width*i : endLinePos], {'\n'}} {
				ln, err = self.w.Write(elem)
				if n += ln; err != nil {
					return
				}
			}
		}
	}

	return
}

// Flush the writer.
func (self *Writer) Flush() error {
	return self.w.Flush()
}

// Close the writer, flushing any unwritten sequence.
func (self *Writer) Close() (err error) {
	err = self.w.Flush()
	if err != nil {
		return
	}
	return self.f.Close()
}
