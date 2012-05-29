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

// Package to read and write BED file formats
package bed

import (
	"bufio"
	"fmt"
	"github.com/kortschak/biogo/bio"
	"github.com/kortschak/biogo/feat"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	chromField = iota
	startField
	endField
	nameField
	scoreField
	strandField
	thickStartField
	thickEndField
	rgbField
	blockCountField
	blockSizesField
	blockStartsField
	lastfield
)

var StrandToChar map[int8]string = map[int8]string{1: "+", 0: "", -1: "-"}
var CharToStrand map[string]int8 = map[string]int8{"+": 1, "": 0, "-": -1}

// BED format reader type.
type Reader struct {
	f       io.ReadCloser
	r       *bufio.Reader
	BedType int
	line    int
}

// Returns a new BED format reader using f.
func NewReader(f io.ReadCloser, b int) *Reader {
	return &Reader{
		f:       f,
		r:       bufio.NewReader(f),
		BedType: b,
	}
}

// Returns a new BED reader using a filename.
func NewReaderName(name string, b int) (r *Reader, err error) {
	f, err := os.Open(name)
	if err != nil {
		return
	}
	return NewReader(f, b), nil
}

// Read a single feature and return it or an error.
func (self *Reader) Read() (f *feat.Feature, err error) {
	var (
		line  string
		elems []string
		se    error
		ok    bool
	)

	line, err = self.r.ReadString('\n')
	if err == nil {
		self.line++
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		line = strings.TrimSpace(line)
		elems = strings.SplitN(line, "\t", self.BedType+1)
		if len(elems) < self.BedType {
			return nil, bio.NewError(fmt.Sprintf("Bad bedtype on line %d", self.line), 0, line)
		}
	} else {
		return
	}

	f = &feat.Feature{Moltype: bio.DNA}

	for i := range elems {
		switch i {
		case chromField:
			f.Location = elems[i]
			if self.BedType <= nameField {
				f.ID = elems[chromField] + ":" + elems[startField] + ".." + elems[endField]
			}
		case startField:
			f.Start, se = strconv.Atoi(elems[i])
			if se != nil {
				f.Start = 0
			}
		case endField:
			f.End, se = strconv.Atoi(elems[i])
			if se != nil {
				f.End = 0
			}
		case nameField:
			f.ID = elems[i]
		case scoreField:
			if f.Score == nil {
				f.Score = new(float64)
			}
			if *f.Score, se = strconv.ParseFloat(elems[i], 64); se != nil {
				*f.Score = 0
			}
		case strandField:
			if f.Strand, ok = CharToStrand[elems[i]]; !ok {
				f.Strand = 0
			}

			// The following fields are unsupported at this stage
		case thickStartField:
		case thickEndField:
		case rgbField:
		case blockCountField:
		case blockSizesField:
		case blockStartsField:
		}
	}

	return
}

// Return the current line number
func (self *Reader) Line() int { return self.line }

// Rewind the reader.
func (self *Reader) Rewind() (err error) {
	if s, ok := self.f.(io.Seeker); ok {
		_, err = s.Seek(0, 0)
		if err == nil {
			self.line = 0
		}
	} else {
		err = bio.NewError("Not a Seeker", 0, self)
	}

	return
}

// Close the reader.
func (self *Reader) Close() (err error) {
	return self.f.Close()
}

// BED format writer type.
type Writer struct {
	f           io.WriteCloser
	w           *bufio.Writer
	BedType     int
	FloatFormat byte
	Precision   int
}

// Returns a new BED format writer using f.
func NewWriter(f io.WriteCloser, b int) *Writer {
	return &Writer{
		f:           f,
		w:           bufio.NewWriter(f),
		BedType:     b,
		FloatFormat: bio.FloatFormat,
		Precision:   bio.Precision,
	}
}

// Returns a new BED format writer using a filename, truncating any existing file.
// If appending is required use NewWriter and os.OpenFile.
func NewWriterName(name string, b int) (w *Writer, err error) {
	f, err := os.Create(name)
	if err != nil {
		return
	}
	return NewWriter(f, b), nil
}

// Write a single feature and return the number of bytes written and any error.
func (self *Writer) Write(f *feat.Feature) (n int, err error) {
	return self.w.WriteString(self.Stringify(f) + "\n")
}

// Convert a feature to a string.
func (self *Writer) Stringify(f *feat.Feature) string {
	fields := make([]string, self.BedType)
	copy(fields, []string{
		string(f.Location),
		strconv.Itoa(f.Start),
		strconv.Itoa(f.End),
	})
	switch self.BedType {
	case 7, 8, 9, 10, 11, 12: // >BED6 not specifically supported
		fallthrough
	case 6:
		fields[strandField] = StrandToChar[f.Strand]
		fallthrough
	case 5:
		fields[scoreField] = strconv.FormatFloat(*f.Score, self.FloatFormat, self.Precision, 64)
		fallthrough
	case 4:
		fields[nameField] = f.ID
	}

	return strings.Join(fields, "\t")
}

// Close the writer, flushing any unwritten data.
func (self *Writer) Close() (err error) {
	err = self.w.Flush()
	if err != nil {
		return
	}
	return self.f.Close()
}
