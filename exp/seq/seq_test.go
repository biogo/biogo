// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package seq_test

import (
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/exp/seq/alignment"
	"code.google.com/p/biogo/exp/seq/linear"
	"code.google.com/p/biogo/exp/seq/multi"
	"testing"
)

func TestSeq(t *testing.T) {
	var (
		_ seq.Sequence = &linear.Seq{}
		_ seq.Sequence = &linear.QSeq{}
		_ seq.Sequence = &alignment.Row{}
		_ seq.Sequence = &alignment.QRow{}

		_ seq.Scorer = &linear.QSeq{}
		_ seq.Scorer = &alignment.QRow{}

		_ seq.Rower = &alignment.Seq{}
		_ seq.Rower = &alignment.QSeq{}
		_ seq.Rower = &multi.Multi{}
		_ seq.Rower = &multi.Set{}
	)
}
