// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
//   This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU General Public License for more details.
//
//   You should have received a copy of the GNU General Public License
//   along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
package bed

import (
	"os"
	"io"
	"bufio"
	"strings"
	"strconv"
	"bio"
	"bio/feat"
)

const (
	chrom = iota
	start
	end
	name
	score
	strand
	thickStart
	thickEnd
	rgb
	blockCount
	blockSizes
	blockStarts
	lastfield
)

var StrandToChar map[int8]string = map[int8]string{1: "+", 0: "", -1: "-"}
var CharToStrand map[string]int8 = map[string]int8{"+": 1, "": 0, "-": -1}

// BED format reader type.
type Reader struct {
	f       io.ReadCloser
	r       *bufio.Reader
	BedType int
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
	var f *os.File
	if f, err = os.Open(name); err != nil {
		return
	}
	return NewReader(f, b), nil
}

// Read a single feature and return it or an error.
func (self *Reader) Read() (f *feat.Feature, err error) {
	var (
		line  string
		elems []string
		s     int8
		ok    bool
	)

	if line, err = self.r.ReadString('\n'); err == nil {
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		line = strings.TrimSpace(line)
		elems = strings.SplitN(line, "\t", self.BedType)
	}

	if s, ok = CharToStrand[elems[strand]]; !ok {
		s = 0
	}

	start, se := strconv.Atoi(elems[start])
	if se != nil {
		start = 0
	}
	end, se := strconv.Atoi(elems[end])
	if se != nil {
		end = 0
	}
	score, se := strconv.Atof64(elems[score])
	if se != nil {
		score = 0
	}

	return &feat.Feature{
		ID:         []byte(elems[name]),
		Location:   []byte(elems[chrom]),
		Start:      start,
		End:        end,
		Feature:    nil,
		Score:      score,
		Attributes: nil,
		Comments:   nil,
		Frame:      0,
		Strand:     s,
		Moltype:    bio.DNA,
	}, err
}

// Rewind the reader.
func (self *Reader) Rewind() (err error) {
	if s, ok := self.f.(io.Seeker); ok {
		_, err = s.Seek(0, 0)
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
	f       io.WriteCloser
	w       *bufio.Writer
	BedType int
}

// Returns a new BED format writer using f.
func NewWriter(f io.WriteCloser, b int) *Writer {
	return &Writer{
		f:       f,
		w:       bufio.NewWriter(f),
		BedType: b,
	}
}

// Returns a new BED format writer using a filename, truncating any existing file.
// If appending is required use NewWriter and os.OpenFile.
func NewWriterName(name string, b int) (w *Writer, err error) {
	var f *os.File
	if f, err = os.Create(name); err != nil {
		return
	}
	return NewWriter(f, b), nil
}

// Write a single feature and return the number of bytes written and any error.
func (self *Writer) Write(f *feat.Feature) (n int, err error) {
	return self.w.WriteString(self.String(f) + "\n")
}

// Convert a feature to a string.
func (self *Writer) String(f *feat.Feature) (line string) {
	line = string(f.Location) + "\t" + string(f.Start) + "\t" + string(f.End)
	if self.BedType > 3 {
		line += string(f.ID) + "\t"
	}
	if self.BedType > 4 {
		line += strconv.Ftoa64(f.Score, 'g', -1) + "\t"
	}
	if self.BedType > 5 {
		line += StrandToChar[f.Strand] + "\t"
	}
	if self.BedType > 6 {
		line += strings.Repeat("\t", 6)
	}

	return
}

// Close the writer, flushing any unwritten data.
func (self *Writer) Close() (err error) {
	if err = self.w.Flush(); err != nil {
		return
	}
	return self.f.Close()
}
