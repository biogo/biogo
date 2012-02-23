// String lexer package
//
// Derived from template/parse/lex.go Copyright 2011 The Go Authors.
package lexstrings

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
	"strings"
	"unicode"
)

// Item represents a token or text string returned from the scanner.
type Item struct {
	Type ItemType
	Val  string
}

func (self Item) String() string {
	switch {
	case self.Type == ItemEOF:
		return "EOF"
	case self.Type == ItemError:
		return self.Val
	case len(self.Val) > 10:
		return fmt.Sprintf("%.10q...", self.Val)
	}
	return fmt.Sprintf("%q", self.Val)
}

// ItemType identifies the type of Lex Items.
type ItemType int

const (
	Incomplete ItemType = iota // incomplete item
	ItemError                  // error occurred; value is text of error
	ItemChar                   // printable ASCII character; grab bag for comma etc.
	ItemEOF
	ItemEnd
	ItemField
	ItemNumber
	ItemText
	ItemID
	ItemSeq
	ItemQual
	LastBuiltin
)

// Make the types prettyprint.
var ItemName = map[ItemType]string{
	Incomplete: "incomplete",
	ItemError:  "error",
	ItemChar:   "char",
	ItemEOF:    "EOF",
	ItemEnd:    "end",
	ItemField:  "field",
	ItemNumber: "number",
	ItemText:   "text",
	ItemID:     "id",
	ItemSeq:    "seq",
	ItemQual:   "qual",
}

func (self ItemType) String() string {
	s := ItemName[self]
	if s == "" {
		return fmt.Sprintf("Item%d", int(self))
	}
	return s
}

// StateFn represents the state of the scanner as a function that returns the next state.
type StateFn func(*Lexer) (StateFn, Item)

// Lexer holds the state of the scanner.
type Lexer struct {
	r       *bufio.Reader // the Reader being scanned.
	Buffer  []rune        // buffer to store tokens being built.
	state   StateFn       // the next lexing function to enter.
	line    int           // current line of the input.
	pos     int           // current position in line of the input.
	lastPos int           // last position in line of the input.
}

// Lex creates a new scanner for the input string.
func Lex(input *bufio.Reader, initState StateFn) *Lexer {
	self := &Lexer{
		r:     input,
		line:  1,
		pos:   1,
		state: initState,
	}
	return self
}

// NextItem returns the next Item from the input.
func (self *Lexer) NextItem() Item {
	var item Item
	for {
		if self.state == nil {
			return Item{ItemEOF, ""}
		}
		if self.state, item = self.state(self); item.Type > Incomplete {
			return item
		}
	}

	panic("cannot reach")
}

// Rewind changes the Lexer to another state - presumably the initial state.
func (self *Lexer) Rewind(state StateFn) {
	self.state = state
}

// Next returns the next rune in the input.
func (self *Lexer) Next() (char rune, err error) {
	if char, _, err = self.r.ReadRune(); err != nil {
		return
	}

	self.Buffer = append(self.Buffer, char)

	if char == '\n' {
		self.line++
		self.lastPos, self.pos = self.pos, 1
	} else {
		self.lastPos = self.pos
		self.pos++
	}

	return
}

// Backup steps back one rune. Can only be called once per call of next.
func (self *Lexer) Backup() (err error) {
	if err = self.r.UnreadRune(); err != nil {
		return
	}

	if self.Buffer[len(self.Buffer)-1] == '\n' {
		self.line--
	}
	self.pos = self.lastPos

	self.Buffer = self.Buffer[:len(self.Buffer)-1]

	return
}

// Peek returns but does not consume the next rune in the input.
func (self *Lexer) Peek() (char rune, err error) {
	if char, err = self.Next(); err != nil {
		return
	}

	err = self.Backup()

	return
}

// Emit passes an Item back to the client.
func (self *Lexer) Emit(t ItemType) Item {
	item := string(self.Buffer)
	self.Buffer = self.Buffer[:0]
	return Item{t, item}
}

// Ignore skips over the pending input before this point.
func (self *Lexer) Ignore() {
	self.Buffer = self.Buffer[:0]
}

// Accept consumes the next rune if it's from the valid set.
func (self *Lexer) Accept(valid string) (ok bool, err error) {
	if next, err := self.Next(); err == nil && strings.IndexRune(valid, next) >= 0 {
		return true, nil
	} else if err != nil {
		return false, err
	}
	err = self.Backup()

	return
}

// AcceptRun consumes a run of runes from the valid set.
func (self *Lexer) AcceptRun(valid string) (err error) {
	for {
		if next, err := self.Next(); err == nil && strings.IndexRune(valid, next) < 0 {
			break
		} else if err != nil {
			return err
		}
	}
	err = self.Backup()

	return
}

// AcceptNot consumes the next char if it's not from the invalid set.
func (self *Lexer) AcceptNot(invalid string) (ok bool, err error) {
	if next, err := self.Next(); err == nil && strings.IndexRune(invalid, next) < 0 {
		return true, nil
	} else if err != nil {
		return false, err
	}
	err = self.Backup()

	return
}

// AcceptRunNot consumes a run of bytes not from the invalid set.
func (self *Lexer) AcceptRunNot(invalid string) (err error) {
	for {
		if next, err := self.Next(); err == nil && strings.IndexRune(invalid, next) >= 0 {
			break
		} else if err != nil {
			return err
		}
	}
	err = self.Backup()

	return
}

// LineNumber reports which line we're on.
func (self *Lexer) LineNumber() int {
	return self.line
}

// LinePosition reports where we are in the line we're on.
func (self *Lexer) LinePosition() int {
	return self.pos
}

// Error returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating self.run.
func (self *Lexer) Errorf(format string, args ...interface{}) (StateFn, Item) {
	return nil, Item{ItemError, fmt.Sprintf(format, args...)}
}

// IsSpace reports whether char is a space character.
func IsSpace(char rune) bool {
	switch char {
	case ' ', '\t', '\n', '\r':
		return true
	}
	return false
}

// IsAlphaNumeric reports whether char is an alphabetic, digit, or underscore.
func IsAlphaNumeric(char rune) bool {
	return char == '_' || unicode.IsLetter(char) || unicode.IsDigit(char)
}

// ScanNumber scans a number: decimal, octal, hex, or float.  This
// isn't a perfect number scanner - for instance it accepts "." and "0x0.2"
// and "089" - but when it's wrong the input is invalid and the parser (via
// strconv) will notice.
//	func LexNumber(self *Lexer) {
//		if !self.ScanNumber() {
//			return self.Errorf("bad number syntax: %q", self.input[self.start:self.pos])
//		}
//		self.Emit(ItemNumber)
//		return <nextStateFn>
//	}
func (self *Lexer) ScanNumber() (ok bool, err error) {
	// Optional leading sign.
	if _, err = self.Accept("+-"); err != nil {
		return
	}
	// Is it hex?
	digits := "0123456789"
	if _, err = self.Accept("0"); err == nil {
		if _, err = self.Accept("xX"); err == nil {
			digits = "0123456789abcdefABCDEF"
		} else {
			return
		}
	} else {
		return
	}
	if err = self.AcceptRun(digits); err != nil {
		return
	}
	if ok, err = self.Accept("."); ok && err == nil {
		self.AcceptRun(digits)
	} else if err != nil {
		return
	}
	if ok, err = self.Accept("eE"); ok && err == nil {
		self.Accept("+-")
		self.AcceptRun("0123456789")
	} else if err != nil {
		return
	}
	// Next thing mustn't be alphanumeric.
	if rune, err := self.Peek(); IsAlphaNumeric(rune) && err == nil {
		self.Next()
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
