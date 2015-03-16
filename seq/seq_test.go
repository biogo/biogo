// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package seq_test

import (
	"github.com/biogo/biogo/seq"
	"github.com/biogo/biogo/seq/alignment"
	"github.com/biogo/biogo/seq/linear"
	"github.com/biogo/biogo/seq/multi"

	"testing"

	_ "gopkg.in/check.v1" // Necessary to squelch complaints when testing ./biogo/... verbosely.
)

func TestSeq(t *testing.T) {
	var (
		_ seq.Sequence = (*linear.Seq)(nil)
		_ seq.Sequence = (*linear.QSeq)(nil)
		_ seq.Sequence = (*alignment.Row)(nil)
		_ seq.Sequence = (*alignment.QRow)(nil)

		_ seq.Scorer = (*linear.QSeq)(nil)
		_ seq.Scorer = (*alignment.QRow)(nil)

		_ seq.Rower = (*alignment.Seq)(nil)
		_ seq.Rower = (*alignment.QSeq)(nil)
		_ seq.Rower = (*multi.Multi)(nil)
		_ seq.Rower = (*multi.Set)(nil)
	)
}
