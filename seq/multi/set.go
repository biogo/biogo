// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package multi

import (
	"fmt"

	"github.com/biogo/biogo/alphabet"
	"github.com/biogo/biogo/errors"
	"github.com/biogo/biogo/seq"
	"github.com/biogo/biogo/util"
)

type Set []seq.Sequence

// Interface guarantees
var (
	_ seq.Rower       = (*Set)(nil)
	_ seq.RowAppender = (*Set)(nil)
)

// Append each []byte in a to the appropriate sequence in the receiver.
func (s Set) AppendEach(a [][]alphabet.QLetter) (err error) {
	if len(a) != s.Rows() {
		return errors.ArgErr{}.Make(fmt.Sprintf("multi: number of sequences does not match row count: %d != %d.", len(a), s.Rows()))
	}
	for i, r := range s {
		r.(seq.Appender).AppendQLetters(a[i]...)
	}
	return nil
}

func (s Set) Row(i int) seq.Sequence {
	return s[i]
}

func (s Set) Len() int {
	max := util.MinInt

	for _, r := range s {
		if l := r.Len(); l > max {
			max = l
		}
	}

	return max
}

func (s Set) Rows() (c int) {
	return len(s)
}

func (s Set) Reverse() {
	for _, r := range s {
		r.Reverse()
	}
}

func (s Set) RevComp() {
	for _, r := range s {
		r.RevComp()
	}
}
