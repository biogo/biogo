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
	"io"
	"bio/seq"
	"bio/io/seqio"
	"bio/alignment"
)

type Reader struct {
	seqio.Reader
}

func NewReader(r seqio.Reader) *Reader {
	return &Reader{r}
}

func (self *Reader) Read() (a *alignment.Alignment, err error) {
	var s *seq.Seq
	a = &alignment.Alignment{}
	for {
		if s, err = self.Reader.Read(); err == nil {
			a.Add(s)
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

type Writer struct {
	seqio.Writer
}

func NewWriter(w seqio.Writer) *Writer {
	return &Writer{w}
}

func (self *Writer) Write(a *alignment.Alignment) (n int, err error) {
	var c int
	for _, s := range *a {
		c, err = self.Writer.Write(s)
		n += c
		if err != nil {
			return
		}
	}

	panic("cannot reach")
}
