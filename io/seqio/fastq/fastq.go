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
	"bufio"
	"bytes"
	"github.com/kortschak/BioGo/bio"
	"github.com/kortschak/BioGo/seq"
	"io"
	"os"
)

/*
Encodings



                                                                                            Q-range  Encoding

 Sanger         !"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHI···                                 0 - 40  Phred+33
 Solexa                                 ··;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_`abcdefgh··· -5 - 40  Solexa+64
 Illumina 1.3+                                 @ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_`abcdefgh···  0 - 40  Phred+64
 Illumina 1.5+                                 xxḆCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_`abcdefgh···  3 - 40  Phred+64*
 Illumina 1.8   !"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJ···                                0 - 40  Phred+33

                !"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_`abcdefghijklmnopqrstuvwxyz{|}~
                |                         |    |        |                              |                     |
               33                        59   64       73                            104                   126

 Q-range for typical raw reads

 * 0=unused, 1=unused, 2=Read Segment Quality Control Indicator (Ḇ)
*/

// Fastq sequence format reader type.
type Reader struct {
	f        io.ReadCloser
	r        *bufio.Reader
	Encoding seq.Encoding
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
				if len(line) > 1 && bytes.Compare(label, line[1:]) != 0 {
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

	labelString := string(label)
	sequence = seq.New(labelString, seqBody, seq.NewQuality(labelString, self.decodeQuality(qualBody)))

	return
}

func (self *Reader) decodeQuality(q []byte) (qs []seq.Qsanger) {
	qs = make([]seq.Qsanger, 0, len(q))

	switch self.Encoding {
	case seq.Sanger, seq.Illumina1_8:
		for _, qe := range q {
			qs = append(qs, seq.Qsanger(qe-33))
		}
	case seq.Solexa:
		for _, qe := range q {
			qs = append(qs, seq.Qsolexa(qe-64).ToSanger())
		}
	case seq.Illumina1_3, seq.Illumina1_5:
		for _, qe := range q {
			qs = append(qs, seq.Qsanger(qe-64))
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
	Encoding seq.Encoding
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
		self.template[1] = []byte(s.ID)
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
func (self *Writer) encodeQuality(q []seq.Qsanger) (qe []byte) {
	qe = make([]byte, 0, len(q))

	switch self.Encoding {
	case seq.Solexa:
		for _, qv := range q {
			qe = append(qe, qv.ToSolexa().Encode(seq.Solexa))
		}
	default:
		for _, qv := range q {
			qe = append(qe, qv.Encode(self.Encoding))
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
