// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bio

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
)

// Trace depth
var TraceDepth = 10

// Base Error handling for bio packages.
type Error interface {
	FileLine() (file string, line int) // Return the file name and line number of caller stored at creation of the Error.
	Trace() (stack []*runtime.Func)    // Return a slice contining the stack trace stored at creation of the Error.
	Package() string                   // Return the package name of the stored caller.
	Function() string                  // Return the function name of the stored caller.
	Items() []interface{}              // Return any items retained by caller.
	Tracef(depth int) string           // A formatted stack trace of the error extending depth frames into the stack, 0 indicates no limit. 
	error
}

type errorBase struct {
	*runtime.Func
	pc      []uintptr
	message string
	items   []interface{}
}

// Create a new Error with message, storing information about the caller stack frame skip levels above the caller and any item that may be needed for handling the error.
func NewError(message string, skip int, items ...interface{}) Error {
	err := &errorBase{
		pc:      make([]uintptr, TraceDepth),
		message: message,
		items:   items,
	}

	var n int
	if n = runtime.Callers(skip+2, err.pc); n > 0 {
		err.Func = runtime.FuncForPC(err.pc[0])
	}
	err.pc = err.pc[:n]

	return err
}

// Return the file name and line number of caller stored at creation of the Error.
func (self *errorBase) FileLine() (file string, line int) {

	return self.Func.FileLine(self.pc[0])
}

// Return a slice contining the stack trace stored at creation of the Error.
func (self *errorBase) Trace() (stack []*runtime.Func) {
	stack = make([]*runtime.Func, len(self.pc))
	for i, pc := range self.pc {
		stack[i] = runtime.FuncForPC(pc)
	}

	return
}

// Return the package name of the stored caller.
func (self *errorBase) Package() string {
	caller := strings.Split(self.Func.Name(), ".")
	return strings.Join(caller[0:len(caller)-1], ".")
}

// Return the function name of the stored caller.
func (self *errorBase) Function() string {
	caller := strings.Split(self.Func.Name(), ".")
	return caller[len(caller)-1]
}

// Return any items retained by caller.
func (self *errorBase) Items() []interface{} { return self.items }

// A formatted stack trace of the error extending depth frames into the stack, 0 indicates no limit. 
func (self *errorBase) Tracef(depth int) string {
	var last, name string
	b := &bytes.Buffer{}
	fmt.Fprintf(b, "Trace: %s:\n", self.message)
	for i, frame := range self.Trace() {
		if depth > 0 && i >= depth {
			break
		}
		file, line := frame.FileLine(self.pc[i])
		if name = frame.Name(); name != last {
			fmt.Fprintf(b, "\n %s:\n", frame.Name())
		}
		last = name
		fmt.Fprintf(b, "\t%s#L=%d\n", file, line)
	}

	return string(b.Bytes())
}

// Satisfy the error interface.
func (self *errorBase) Error() string {
	return self.message
}
