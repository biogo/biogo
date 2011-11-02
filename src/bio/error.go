package bio
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

import (
	"runtime"
	"strings"
	"fmt"
)

// Base Error handling for bio packages.
type Error struct {
	*runtime.Func
	pc      []uintptr
	message string
	Item    interface{}
}

func NewError(message string, skip int, item interface{}) (err *Error) {
	err = &Error{
		pc:      make([]uintptr, 1),
		message: message,
		Item:    item,
	}

	if n := runtime.Callers(skip+1, err.pc); n > 0 {
		err.Func = runtime.FuncForPC(err.pc[0])
	}

	return
}

func (self *Error) FileLine() (file string, line int) {
	return self.Func.FileLine((*self).pc[0])
}

func (self *Error) Trace() (stack []*runtime.Func) {
	stack = make([]*runtime.Func, len(self.pc))
	for i, pc := range self.pc {
		stack[i] = runtime.FuncForPC(pc)
	}

	return
}

func (self *Error) Package() string {
	return strings.Split(self.Func.Name(), ".")[0]
}

// A formatted stack trace of the error extending depth frames into the stack, 0 indicates no limit. 
func (self *Error) Tracef(depth int) (trace string) {
	trace = "Trace: " + self.message + ":\n"
	for i, frame := range self.Trace() {
		if depth > 0 && i >= depth {
			break
		}
		file, line := frame.FileLine(self.pc[i])
		trace += fmt.Sprintf(" %s%s:%s %d\n", strings.Repeat(" ", i), self.Func.Name(), file, line)
	}

	return
}

func (self *Error) Error() string {
	return self.message
}
