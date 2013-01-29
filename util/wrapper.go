// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"io"
)

// A Wrapper provides hard line wrapping and output limits to an io.Writer.
type Wrapper struct {
	w     io.Writer
	n     int
	width int
	limit int
}

// NewWrapper returns a Wrapper that causes wraps lines at width bytes and
// limits the number of bytes written to the provided limit.
func NewWrapper(w io.Writer, width, limit int) *Wrapper {
	return &Wrapper{
		w:     w,
		width: width,
		limit: limit,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Write writes the lesser of len(p) or the Writer's limit bytes from p to
// the underlying data stream. It returns the number of bytes written from
// p (0 <= n <= len(p)) and any error encountered that caused the write to
// stop early, except the Writer's own limit.
func (w *Wrapper) Write(p []byte) (n int, err error) {
	if w.limit >= 0 {
		if w.n >= w.limit {
			return 0, nil
		}
		p = p[:min(w.limit-w.n, len(p))]
	}
	if w.width <= 0 {
		return w.w.Write(p)
	}
	var _n int
	for len(p) > 0 {
		if w.n != 0 && w.n%w.width == 0 {
			_n, err = w.w.Write([]byte{'\n'})
			n += _n
			if err != nil {
				return
			}
		}
		_n, err = w.w.Write(p[:min(w.width-w.n%w.width, len(p))])
		n += _n
		w.n += _n
		if err != nil {
			return
		}
		p = p[_n:]
	}
	return n, err
}
