// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errors

// Chain is an error and layered error annotations.
type Chain interface {
	// The error behavior of a Chain is based on the last annotation applied.
	error

	// Cause returns the initial error in the Chain.
	Cause() error

	// Link adds an annotation layer to the Chain.
	Link(error) Chain

	// Last returns the Chain, or nil if the Chain is empty, and the most recent annotation.
	Last() (Chain, error)
}

// Links is an optional interface used by the Errors function.
type Links interface {
	Errors() []error // Errors returns a flat list of errors in temporal order of annotation.
}

// NewChain returns a new Chain based on the provided error. If the error is a Chain it
// is returned unaltered.
func NewChain(err error) Chain {
	if c, ok := err.(Chain); ok {
		return c
	}
	return chain{err}
}

// Cause returns the initially identified cause of an error if the error is a Chain, or the error
// itself if it is not.
func Cause(err error) error {
	if c, ok := err.(Chain); ok {
		return c.Cause()
	}
	return err
}

// Link adds an annotation to an error, returning a Chain.
func Link(err, annotation error) Chain { return NewChain(err).Link(annotation) }

// Last returns the most recent annotation of an error and the remaining chain
// after the annotation is removed or nil if no further errors remain. Last returns
// a nil Chain if the error is not a Chain.
func Last(err error) (Chain, error) {
	if c, ok := err.(Chain); ok {
		return c.Last()
	}
	return nil, err
}

// Errors returns a flat list of errors in temporal order of annotation. If the provided
// error is not a Chain a single element slice of error is returned containing the error.
// If the error implements Links, its Errors method is called and the result returned.
func Errors(err error) []error {
	if err == nil {
		return nil
	}
	switch c := err.(type) {
	case Links:
		return c.Errors()
	case Chain:
		var errs []error
		for c != nil {
			c, err = c.Last()
			errs = append(errs, err)
		}
		return reverse(errs)
	default:
		return []error{err}
	}
}

func reverse(err []error) []error {
	for i, j := 0, len(err)-1; i < j; i, j = i+1, j-1 {
		err[i], err[j] = err[j], err[i]
	}
	return err
}

// chain is the basic implementation.
type chain []error

func (c chain) Error() string {
	if len(c) > 0 {
		return c[len(c)-1].Error()
	}
	return ""
}
func (c chain) Cause() error {
	if len(c) > 0 {
		return c[0]
	}
	return nil
}
func (c chain) Link(err error) Chain { return append(c, err) }
func (c chain) Last() (Chain, error) {
	switch len(c) {
	case 0:
		return nil, nil
	case 1:
		return nil, c[0]
	default:
		return c[:len(c)-1], c[len(c)-1]
	}
}
func (c chain) Errors() []error { return c }
