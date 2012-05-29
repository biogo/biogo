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
	"github.com/kortschak/biogo/bio"
	"github.com/kortschak/biogo/exp/alphabet"
	"github.com/kortschak/biogo/exp/seq"
	"github.com/kortschak/biogo/exp/seqio"
	"io"
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
	r   *bufio.Reader
	t   seqio.SequenceAppender
	enc alphabet.Encoding
}

// Returns a new fastq format reader using r.
func NewReader(r io.Reader, t seqio.SequenceAppender) *Reader {
	var enc alphabet.Encoding
	if e, ok := t.(seq.Encoder); ok {
		enc = e.Encoding()
	} else {
		enc = alphabet.None
	}

	return &Reader{
		r:   bufio.NewReader(r),
		t:   t,
		enc: enc,
	}
}

// Read a single sequence and return it or an error.
// TODO: Does not read multi-line fastq.
func (self *Reader) Read() (s seq.Sequence, err error) {
	var (
		buff, line, label []byte
		isPrefix          bool
		seqBuff           []alphabet.QLetter
		t                 seqio.SequenceAppender
	)

	inQual := false

	for {
		if buff, isPrefix, err = self.r.ReadLine(); err == nil {
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
					return nil, bio.NewError("fastq: no header line parsed before +line in fastq format", 0)
				}
				if len(line) > 1 && bytes.Compare(label[1:], line[1:]) != 0 {
					return nil, bio.NewError("fastq: quality header does not match sequence header", 0)
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
					return nil, bio.NewError("fastq: sequence/quality length mismatch", 0)
				}
				for i := range line {
					seqBuff[i].Q = alphabet.DecodeToQphred(line[i], self.enc)
				}
				t.AppendQLetters(seqBuff...)

				return t, nil
			}
		} else {
			return
		}
	}

	panic("cannot reach")
}

func (self *Reader) readHeader(line []byte) (s seqio.SequenceAppender) {
	s = self.t.Copy().(seqio.SequenceAppender)
	fieldMark := bytes.IndexAny(line, " \t")
	if fieldMark < 0 {
		*s.Name() = string(line[1:])
		*s.Description() = ""
	} else {
		*s.Name() = string(line[1:fieldMark])
		*s.Description() = string(line[fieldMark+1:])
	}

	return
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
func (self *Writer) Write(s seqio.Sequence) (n int, err error) {
	var (
		ln, c int
		enc   alphabet.Encoding
	)
	if e, ok := s.(seq.Encoder); ok {
		enc = e.Encoding()
	} else {
		enc = alphabet.Sanger
	}

	n, err = self.writeHeader('@', s)
	if err != nil {
		return
	}
	for i := 0; i < s.Len(); i++ {
		ln, err = self.w.Write([]byte{byte(s.At(seq.Position{Pos: i, Ind: c}).L)})
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
		ln, err = self.w.Write([]byte{s.At(seq.Position{Pos: i, Ind: c}).Q.Encode(enc)})
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

func (self *Writer) writeHeader(prefix byte, s seqio.Sequence) (n int, err error) {
	var ln int
	n, err = self.w.Write([]byte{prefix})
	if err != nil {
		return
	}
	ln, err = io.WriteString(self.w, *s.Name())
	if n += ln; err != nil {
		return
	}
	if desc := *s.Description(); len(desc) != 0 {
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
