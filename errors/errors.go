// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package errors supports generic rich error reporting.
//
// This package is deprecated. Since it was written much better
// approaches have been developed.
package errors

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
)

// Type Error is the interface for rich error reporting supported by the
// errors package.
type Error interface {
	// FileLine returns the file name and line number of caller
	// stored at creation of the Error.
	FileLine() (file string, line int)

	// Trace returns a slice continuing the stack trace stored at
	// creation of the Error.
	Trace() (stack []*runtime.Func)

	// Package returns the package name of the stored caller.
	Package() string

	// Function returns the function name of the stored caller.
	Function() string

	// Items returns any items retained by caller.
	Items() []interface{}

	// Tracef returns a formatted stack trace of the error
	// extending depth frames into the stack, 0 indicates no limit.
	Tracef(depth int) string
	error
}

type errorBase struct {
	*runtime.Func
	pc      []uintptr
	message string
	items   []interface{}
}

// Make creates a new Error with message, storing information about the
// caller stack frame skip levels above the caller and any item that may
// be needed for handling the error. The number of frames stored is specified
// by the depth parameter. If depth is zero, Make will panic.
func Make(message string, skip, depth int, items ...interface{}) Error {
	if depth == 0 {
		panic("errors: zero trace depth")
	}
	err := &errorBase{
		pc:      make([]uintptr, depth),
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

// Make creates a new Error with message, storing information about the
// caller stack frame skip levels above the caller and any item that may
// be needed for handling the error. The number of frames stored is specified
// by the depth parameter. If depth is zero, Make will panic.
func MakeErr(err *errorBase, message string, skip, depth int, items ...interface{}) Error {
	if depth == 0 {
		panic("errors: zero trace depth")
	}

	err.pc = make([]uintptr, depth)
	err.message = message
	err.items = items

	var n int
	if n = runtime.Callers(skip+2, err.pc); n > 0 {
		err.Func = runtime.FuncForPC(err.pc[0])
	}
	err.pc = err.pc[:n]

	return err
}

// Return the file name and line number of caller stored at creation of
// the Error.
func (err *errorBase) FileLine() (file string, line int) {
	return err.Func.FileLine(err.pc[0])
}

// Return a slice contining the stack trace stored at creation of the Error.
func (err *errorBase) Trace() (stack []*runtime.Func) {
	stack = make([]*runtime.Func, len(err.pc))
	for i, pc := range err.pc {
		stack[i] = runtime.FuncForPC(pc)
	}

	return
}

// Return the package name of the stored caller.
func (err *errorBase) Package() string {
	caller := strings.Split(err.Func.Name(), ".")
	return strings.Join(caller[0:len(caller)-1], ".")
}

// Return the function name of the stored caller.
func (err *errorBase) Function() string {
	caller := strings.Split(err.Func.Name(), ".")
	return caller[len(caller)-1]
}

// Return any items retained by caller.
func (err *errorBase) Items() []interface{} { return err.items }

// A formatted stack trace of the error extending depth frames into the
// stack, 0 indicates no limit.
func (err *errorBase) Tracef(depth int) string {
	var last, name string
	b := &bytes.Buffer{}
	fmt.Fprintf(b, "Trace: %s:\n", err.message)
	for i, frame := range err.Trace() {
		if depth > 0 && i >= depth {
			break
		}
		file, line := frame.FileLine(err.pc[i])
		if name = frame.Name(); name != last {
			fmt.Fprintf(b, "\n %s:\n", frame.Name())
		}
		last = name
		fmt.Fprintf(b, "\t%s#L=%d\n", file, line)
	}

	return string(b.Bytes())
}

// Satisfy the error interface.
func (err errorBase) Error() string {
	return err.message
}

func (be errorBase) Make(message string) errorBase {
	return be.MakeTrace(message, 0, 5)
}

func (be errorBase) MakeTrace(message string, skip, depth int, items ...interface{}) errorBase {
	if depth == 0 {
		panic("errors: zero trace depth")
	}

	be.pc = make([]uintptr, depth)
	be.message = message
	be.items = items

	var n int
	if n = runtime.Callers(skip+2, be.pc); n > 0 {
		be.Func = runtime.FuncForPC(be.pc[0])
	}
	be.pc = be.pc[:n]

	return be
}

// Argument Errors indicate that the function has been provided bad data
type ArgErr struct{ errorBase }

type InvalidAlphabetErr struct{ ArgErr }
type InvalidAlphabetPairingErr struct{ ArgErr }
type MismatchedTypesErr struct{ ArgErr }
type MismatchedAlphabetsErr struct{ ArgErr }
type NoAlphabetErr struct{ ArgErr }
type NotGappedAlphabetErr struct{ ArgErr }
type TypeNotHandledErr struct{ ArgErr }
type MatrixNotSquareErr struct{ ArgErr }

type BadBedTypeErr struct{ ArgErr }
type BadStrandFieldErr struct{ ArgErr }
type BadStrandErr struct{ ArgErr }
type BadColorFieldErr struct{ ArgErr }
type MissingBlockValuesErr struct{ ArgErr }
type MissingChromFieldErr struct{ ArgErr }

type KTooLargeErr struct{ ArgErr }
type KTooSmallErr struct{ ArgErr }
type KSeqTooShortErr struct{ ArgErr }
type BadKmerErr struct{ ArgErr }
type BadKmerTextLenErr struct{ ArgErr }
type IllegalKmerTextErr struct{ ArgErr }

type ExonNotInTranscriptErr struct{ ArgErr }
type NoZeroStartExonErr struct{ ArgErr }
type ExonOverlapErr struct{ ArgErr }
type ExonLocationDiffersErr struct{ ArgErr }
type NewExonLocationDiffersErr struct{ ArgErr }

type TranscriptLocationMismatchesGeneErr struct{ ArgErr }
type NoZeroStartTranscriptErr struct{ ArgErr }

// StateErr types should indicate that, while processing, the function has entered a bad state
type StateErr struct{ errorBase }

type QualitySequenceHeaderMismatchErr struct{ StateErr }
type QualitySequenceLengthMismatchErr struct{ StateErr }

// ConcurrencyErr are thrown when a part of Biogo's concurrency model has failed
type ConcurrencyErr struct{ StateErr }

// MultiErr are used to wrap cases where more than a single error are thrown by a logic block
type MultiError struct {
	errorBase
	Errors []error
}
