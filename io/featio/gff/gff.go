// Package to read and write GFF format files
package gff

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
	"fmt"
	"github.com/kortschak/BioGo/bio"
	"github.com/kortschak/BioGo/feat"
	"github.com/kortschak/BioGo/io/seqio/fasta"
	"github.com/kortschak/BioGo/seq"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	nameField = iota
	sourceField
	featureField
	startField
	endField
	scoreField
	strandField
	frameField
	attributeField
	commentField
)

var (
	DefaultVersion                    = 2
	DefaultToOneBased                 = true
	strandToChar      map[int8]string = map[int8]string{1: "+", 0: ".", -1: "-"}
	charToStrand      map[string]int8 = map[string]int8{"+": 1, ".": 0, "-": -1}
)

// GFF format reader type.
type Reader struct {
	f             io.ReadCloser
	r             *bufio.Reader
	Version       int
	OneBased      bool
	SourceVersion string
	Date          time.Time
	TimeFormat    string // Required for parsing date fields
	Type          bio.Moltype
}

// Returns a new GFF format reader using f.
func NewReader(f io.ReadCloser) *Reader {
	return &Reader{
		f:        f,
		r:        bufio.NewReader(f),
		OneBased: DefaultToOneBased,
	}
}

// Returns a new GFF reader using a filename.
func NewReaderName(name string) (r *Reader, err error) {
	var f *os.File
	if f, err = os.Open(name); err != nil {
		return
	}
	return NewReader(f), nil
}

func (self *Reader) commentMetaline(line string) (f *feat.Feature, err error) {
	// Load these into a slice in a MetaField of the Feature
	fields := strings.Split(string(line), " ")
	switch fields[0] {
	case "gff-version":
		if self.Version, err = strconv.Atoi(fields[1]); err != nil {
			self.Version = DefaultVersion
		}
		return self.Read()
	case "source-version":
		if len(fields) > 1 {
			self.SourceVersion = strings.Join(fields[1:], " ")
			return self.Read()
		} else {
			return nil, bio.NewError("Incomplete source-version metaline", 0, fields)
		}
	case "date":
		if len(fields) > 1 {
			self.Date, err = time.Parse(self.TimeFormat, strings.Join(fields[1:], " "))
			return self.Read()
		} else {
			return nil, bio.NewError("Incomplete date metaline", 0, fields)
		}
	case "Type":
		if len(fields) > 1 {
			ok := false
			if self.Type, ok = bio.ParseMoltype[fields[1]]; !ok {
				self.Type = bio.Undefined
			}
			return self.Read()
		} else {
			return nil, bio.NewError("Incomplete Type metaline", 0, fields)
		}
	case "sequence-region":
		if len(fields) > 3 {
			var start, end int
			if start, err = strconv.Atoi(fields[2]); err != nil {
				return nil, err
			} else {
				if self.OneBased {
					start = bio.OneToZero(start)
				}
			}
			if end, err = strconv.Atoi(fields[3]); err != nil {
				return nil, err
			}
			f = &feat.Feature{
				Meta: &feat.Feature{
					ID:    fields[1],
					Start: start,
					End:   end,
				},
			}
		} else {
			return nil, bio.NewError("Incomplete sequence-region metaline", 0, fields)
		}
	case "DNA", "RNA", "Protein":
		if len(fields) > 1 {
			var s *seq.Seq
			if s, err = self.metaSequence(fields[0], fields[1]); err != nil {
				return
			} else {
				f = &feat.Feature{Meta: s}
			}
		} else {
			return nil, bio.NewError("Incomplete sequence metaline", 0, fields)
		}
	default:
		f = &feat.Feature{Meta: line}
	}

	return
}

func (self *Reader) metaSequence(moltype, id string) (sequence *seq.Seq, err error) {
	var line, body []byte

	for {
		if line, err = self.r.ReadBytes('\n'); err == nil {
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			if len(line) == 0 {
				continue
			}
			if len(line) < 2 || !bytes.HasPrefix(line, []byte("##")) {
				return nil, bio.NewError("Corrupt metasequence", 0, line)
			}
			line = bytes.TrimSpace(line[2:])
			if string(line) == "end-"+moltype {
				break
			} else {
				line = bytes.Join(bytes.Fields(line), nil)
				body = append(body, line...)
			}
		} else {
			return nil, err
		}
	}

	sequence = seq.New(id, body, nil)
	sequence.Moltype = bio.ParseMoltype[moltype]

	return
}

// Read a single feature or part and return it or an error.
func (self *Reader) Read() (f *feat.Feature, err error) {
	var (
		line  string
		elems []string
		s     int8
		ok    bool
	)

	for {
		if line, err = self.r.ReadString('\n'); err == nil {
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			line = strings.TrimSpace(line)
			if len(line) == 0 { // ignore blank lines
				continue
			} else if strings.HasPrefix(line, "##") {
				f, err = self.commentMetaline(line[2:])
				return
			} else if line[0] != '#' { // ignore comments
				elems = strings.SplitN(line, "\t", 10)
				break
			}
		} else {
			return
		}
	}

	if s, ok = charToStrand[elems[strandField]]; !ok {
		s = 0
	}

	startPos, se := strconv.Atoi(elems[startField])
	if se != nil {
		startPos = 0
	} else {
		if self.OneBased {
			startPos = bio.OneToZero(startPos)
		}
	}

	endPos, se := strconv.Atoi(elems[endField])
	if se != nil {
		endPos = 0
	}

	fr, se := strconv.Atoi(elems[frameField])
	if se != nil {
		fr = -1
	}

	score, se := strconv.ParseFloat(elems[scoreField], 64)
	if se != nil {
		score = math.NaN()
	}

	f = &feat.Feature{
		ID:       elems[nameField] + ":" + strconv.Itoa(startPos) + ".." + strconv.Itoa(endPos),
		Location: elems[nameField],
		Source:   elems[sourceField],
		Start:    startPos,
		End:      endPos,
		Feature:  elems[featureField],
		Score:    score,
		Frame:    int8(fr),
		Strand:   s,
		Moltype:  self.Type, // currently we default to bio.DNA
	}

	if len(elems) > attributeField {
		f.Attributes = elems[attributeField]
	}
	if len(elems) > commentField {
		f.Comments = elems[commentField]
	}

	return
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

// GFF format writer type.
type Writer struct {
	f           io.WriteCloser
	w           *bufio.Writer
	Version     int
	OneBased    bool
	FloatFormat byte
	Precision   int
	Width       int
}

// Returns a new GFF format writer using f.
// When header is true, a version header will be written to the GFF.
func NewWriter(f io.WriteCloser, version, width int, header bool) (w *Writer) {
	w = &Writer{
		f:           f,
		w:           bufio.NewWriter(f),
		Version:     version,
		OneBased:    DefaultToOneBased,
		FloatFormat: bio.FloatFormat,
		Precision:   bio.Precision,
		Width:       width,
	}

	if header {
		w.WriteMetaData(fmt.Sprintf("gff-version %d", version))
	}

	return
}

// Returns a new GFF format writer using a filename, truncating any existing file.
// If appending is required use NewWriter and os.OpenFile.
// When header is true, a version header will be written to the GFF.
func NewWriterName(name string, version, width int, header bool) (w *Writer, err error) {
	var f *os.File
	if f, err = os.Create(name); err != nil {
		return
	}
	return NewWriter(f, version, width, header), nil
}

// Write a single feature and return the number of bytes written and any error.
func (self *Writer) Write(f *feat.Feature) (n int, err error) {
	return self.w.WriteString(self.Stringify(f) + "\n")
}

// Convert a feature to a string.
func (self *Writer) Stringify(f *feat.Feature) string {
	fields := make([]string, 8, 10)

	var start int
	if self.OneBased {
		start = bio.ZeroToOne(f.Start)
	}

	copy(fields, []string{
		f.Location,
		f.Source,
		f.Feature,
		strconv.Itoa(start),
		strconv.Itoa(f.End),
	})

	if !math.IsNaN(f.Score) {
		fields[scoreField] = strconv.FormatFloat(f.Score, self.FloatFormat, self.Precision, 64)
	} else {
		fields[scoreField] = "."
	}

	if f.Moltype == bio.DNA {
		fields[strandField] = strandToChar[f.Strand]
	} else {
		fields[strandField] = "."
	}

	if frame := strconv.Itoa(int(f.Frame)); (f.Moltype == bio.DNA || self.Version < 2) && (frame == "0" || frame == "1" || frame == "2") {
		fields[frameField] = frame
	} else {
		fields[frameField] = "."
	}
	if f.Attributes != "" || f.Comments != "" {
		fields = append(fields, f.Attributes)
	}
	if f.Comments != "" {
		fields = append(fields, "#" + f.Comments)
	}

	return strings.Join(fields,"\t")
}

// Write meta data to a GFF file.
func (self *Writer) WriteMetaData(d interface{}) (n int, err error) {
	switch d.(type) {
	case []byte, string:
		n, err = self.w.WriteString("##" + d.(string) + "\n")
	case *seq.Seq:
		sw := fasta.NewWriter(self.f, self.Width)
		sw.IDPrefix = fmt.Sprintf("##%s ", d.(*seq.Seq).Moltype)
		sw.SeqPrefix = "##"
		if n, err = sw.Write(d.(*seq.Seq)); err != nil {
			return
		}
		if err = sw.Flush(); err != nil {
			return
		}
		var m int
		m, err = self.w.WriteString("##end-" + d.(*seq.Seq).Moltype.String() + "\n")
		n += m
		if err != nil {
			return
		}
		err = self.w.Flush()
		return
	case *feat.Feature:
		start := d.(*feat.Feature).Start
		if self.OneBased && start >= 0 {
			start++
		}
		n, err = self.w.WriteString("##sequence-region " + string(d.(*feat.Feature).ID) + " " +
			strconv.Itoa(start) + " " +
			strconv.Itoa(d.(*feat.Feature).End) + "\n")
	default:
		n, err = 0, bio.NewError("Unknown meta data type", 0, d)
	}

	if err == nil {
		err = self.w.Flush()
	}

	return
}

// Write a comment line to a GFF file
func (self *Writer) WriteComment(c string) (n int, err error) {
	n, err = self.w.WriteString("# " + c + "\n")

	return
}

// Flush the writer.
func (self *Writer) Flush() error {
	return self.w.Flush()
}

// Close the writer, flushing any unwritten data.
func (self *Writer) Close() (err error) {
	if err = self.w.Flush(); err != nil {
		return
	}
	return self.f.Close()
}
