// Package for reading and writing multiple sequence alignment files
package alignio

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
	"github.com/kortschak/BioGo/io/seqio"
	"github.com/kortschak/BioGo/seq"
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
		if s, err = self.Reader.Read(); err == nil {
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
