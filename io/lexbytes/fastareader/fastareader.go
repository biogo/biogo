// Example use of lexbytes
package fastareader

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
	"fmt"
	"github.com/kortschak/biogo/bio"
	lex "github.com/kortschak/biogo/io/lexbytes"
	"github.com/kortschak/biogo/seq"
	"io"
	"os"
)

// Fasta sequence format reader type.
type Reader struct {
	f io.ReadCloser
	r *bufio.Reader
	l *lex.Lexer
}

// Returns a new fasta format reader using f.
func NewReader(f io.ReadCloser) *Reader {
	b := bufio.NewReader(f)
	return &Reader{
		f: f,
		r: b,
		l: lex.Lex(b, scanForID),
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
	sequence = &seq.Seq{}
	expectedType := lex.ItemID
	for {
		switch item := self.l.NextItem(); {
		case item.Type == lex.ItemError:
			return nil, bio.NewError(item.String(), 0)
		case item.Type == lex.ItemID && expectedType == item.Type:
			sequence.ID = string(item.Val)
			expectedType = lex.ItemSeq
		case item.Type == lex.ItemSeq && expectedType == item.Type:
			if len(sequence.Seq) == 0 {
				sequence.Seq = item.Val
			} else {
				sequence.Seq = append(sequence.Seq, item.Val...)
			}
		case item.Type == lex.ItemEOF:
			if len(sequence.ID) == 0 && len(sequence.Seq) == 0 {
				err = io.EOF
			}
			fallthrough
		case item.Type == lex.ItemEnd:
			return
		default:
			return nil, bio.NewError(fmt.Sprintf("Unexpected item type at line %d", self.l.LineNumber()), 0, item)
		}
	}

	panic("cannot reach")
}

func scanForID(l *lex.Lexer) (lex.StateFn, lex.Item) {
	switch err := l.AcceptRunNot([]byte(">")); {
	case err != nil:
		return l.Errorf("%s", err)
	default:
		_, err = l.Next()
		if err != nil {
			return l.Errorf("%s", err)
		}
		l.Ignore()
		return inID, lex.Item{}
	}

	panic("cannot reach")
}

func inID(l *lex.Lexer) (lex.StateFn, lex.Item) {
	switch err := l.AcceptRunNot([]byte("\n")); {
	case err != nil:
		return l.Errorf("%s", err)
	default:
		return inSeq, l.Emit(lex.ItemID)
	}

	panic("cannot reach")
}

func inSeq(l *lex.Lexer) (lex.StateFn, lex.Item) {
	switch char, err := l.Next(); {
	case err == io.EOF:
		l.Buffer = l.Buffer[:len(l.Buffer)]
		if len(l.Buffer) > 0 {
			return inSeq, l.Emit(lex.ItemSeq)
		}
		return nil, l.Emit(lex.ItemEOF)
	case err != nil:
		return l.Errorf("%s", err)
	case char == '>':
		l.Buffer = l.Buffer[:len(l.Buffer)-1]
		if len(l.Buffer) > 0 {
			return inID, l.Emit(lex.ItemSeq)
		}
		return inID, l.Emit(lex.ItemEnd)
	case char == '\n':
		l.Buffer = l.Buffer[:len(l.Buffer)-1]
		if len(l.Buffer) > 0 {
			return inSeq, l.Emit(lex.ItemSeq)
		}
		return inSeq, lex.Item{}
	case lex.IsSpace(char):
		l.Buffer = l.Buffer[:len(l.Buffer)-1]
		fallthrough
	default:
		return inSeq, lex.Item{}
	}

	panic("cannot reach")
}

// Rewind the reader.
func (self *Reader) Rewind() (err error) {
	if s, ok := self.f.(io.Seeker); ok {
		_, err = s.Seek(0, 0)
		self.l.Rewind(scanForID)
	} else {
		err = bio.NewError("Not a Seeker", 0, self)
	}
	return
}

// Close the reader.
func (self *Reader) Close() (err error) {
	return self.f.Close()
}
