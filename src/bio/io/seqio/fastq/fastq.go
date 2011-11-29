// Package to read and write FASTQ format files
package fastq
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
	"os"
	"io"
	"bufio"
	"bytes"
	"bio"
	"bio/seq"
)

const (
	Sanger = iota
	Solexa
	Illumina
)

// Fastq sequence format reader type.
type Reader struct {
	f        io.ReadCloser
	r        *bufio.Reader
	Encoding int
}

// Returns a new fastq format reader using r.
func NewReader(f io.ReadCloser) *Reader {
	return &Reader{
		f: f,
		r: bufio.NewReader(f),
	}
}

// Returns a new fastq format reader using a filename.
func NewReaderName(name string) (r *Reader, err error) {
	var f *os.File
	if f, err = os.Open(name); err != nil {
		return
	}
	return NewReader(f), nil
}

// Read a single sequence and return it or an error.
func (self *Reader) Read() (sequence *seq.Seq, err error) {
	var line, label, seqBody, qualBody []byte
	sequence = &seq.Seq{}

	inQual := false
READ:
	for {
		if line, err = self.r.ReadBytes('\n'); err == nil {
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			switch {
			case !inQual && line[0] == '@':
				label = line[1:]
			case !inQual && line[0] == '+':
				if len(sequence.ID) == 0 {
					return nil, bio.NewError("No ID line parsed at +line in fastq format", 0, nil)
				}
				if len(line) > 1 && bytes.Compare(sequence.ID, line[1:]) != 0 {
					return nil, bio.NewError("Quality ID does not match sequence ID", 0, nil)
				}
				inQual = true
			case !inQual:
				line = bytes.Join(bytes.Fields(line), nil)
				seqBody = append(seqBody, line...)
			case inQual:
				line = bytes.Join(bytes.Fields(line), nil)
				qualBody = append(qualBody, line...)
				if len(qualBody) >= len(seqBody) {
					break READ
				}
			}
		} else {
			return
		}
	}

	if len(seqBody) != len(qualBody) {
		return nil, bio.NewError("Quality length does not match sequence length", 0, nil)
	}

	sequence = seq.New(label, seqBody, seq.NewQuality(label, self.decodeQuality(qualBody)))

	return
}

func (self *Reader) decodeQuality(q []byte) (qs []int8) {
	qs = make([]int8, 0, len(q))

	switch self.Encoding {
	case Sanger:
		for _, qe := range q {
			qs = append(qs, int8(qe)-33)
		}
	case Solexa:
		for _, qe := range q {
			qs = append(qs, seq.SolexaToSanger(int8(qe)-64))
		}
	case Illumina:
		for _, qe := range q {
			qs = append(qs, int8(qe)-64)
		}
	}

	return qs
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

// Fastq sequence format writer type.
type Writer struct {
	f        io.WriteCloser
	w        *bufio.Writer
	template [][]byte
	Encoding int
}

// Returns a new fastq format writer using w.
func NewWriter(f io.WriteCloser) *Writer {
	return &Writer{
		f: f,
		w: bufio.NewWriter(f),
		template: [][]byte{
			[]byte{'@'},
			[]byte{}, // ID
			[]byte{'\n'},
			[]byte{}, // Seq
			[]byte("\n+\n"),
			[]byte{}, // Quality
			[]byte{'\n'},
		},
	}
}

// Returns a new fastq format writer using a filename, truncating any existing file.
// If appending is required use NewWriter and os.OpenFile.
func NewWriterName(name string) (w *Writer, err error) {
	var f *os.File
	if f, err = os.Create(name); err != nil {
		return
	}
	return NewWriter(f), nil
}

// Write a single sequence and return the number of bytes written and any error.
func (self *Writer) Write(s *seq.Seq) (n int, err error) {
	if s.Quality == nil {
		return 0, bio.NewError("No quality associated with sequence", 0, s)
	}
	if s.Len() == s.Quality.Len() {
		self.template[1] = s.ID
		self.template[3] = s.Seq
		self.template[5] = self.encodeQuality(s.Quality.Qual)
		var tn int
		for _, t := range self.template {
			tn, err = self.w.Write(t)
			n += tn
			if err != nil {
				return
			}
		}
	} else {
		return 0, bio.NewError("Sequence length and quality length do not match", 0, s)
	}

	return
}

// Could do this by lookup - make three tables (or at least one table for Sanger -> Solexa)
func (self *Writer) encodeQuality(q []int8) (qe []byte) {
	qe = make([]byte, 0, len(q))

	switch self.Encoding {
	case Sanger:
		for _, qv := range q {
			if qv <= 93 {
				qv += 33
			}
			qe = append(qe, byte(qv))
		}
	case Solexa:
		for _, qv := range q {
			qv = seq.SangerToSolexa(qv)
			if qv <= 62 {
				qv += 64
			}
			qe = append(qe, byte(qv))
		}
	case Illumina:
		for _, qv := range q {
			if qv <= 62 {
				qv += 64
			}
			qe = append(qe, byte(qv))
		}
	}

	return qe
}

// Close the writer, flushing any unwritten sequence.
func (self *Writer) Close() (err error) {
	if err = self.w.Flush(); err != nil {
		return
	}
	return self.f.Close()
}
