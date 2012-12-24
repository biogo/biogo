// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package feat_test

import (
	"code.google.com/p/biogo/exp/feat"
	"code.google.com/p/biogo/exp/seq/alignment"
	"code.google.com/p/biogo/exp/seq/linear"
	"code.google.com/p/biogo/exp/seq/multi"
	"testing"
)

func TestFeat(t *testing.T) {
	var (
		_ feat.Feature = &linear.Seq{}
		_ feat.Feature = &linear.QSeq{}
		_ feat.Feature = &alignment.Seq{}
		_ feat.Feature = &alignment.QSeq{}
		_ feat.Feature = &multi.Multi{}

		_ feat.Offsetter = &linear.Seq{}
		_ feat.Offsetter = &linear.QSeq{}
		_ feat.Offsetter = &alignment.Seq{}
		_ feat.Offsetter = &alignment.QSeq{}
		_ feat.Offsetter = &multi.Multi{}

		_ feat.LocationSetter = &linear.Seq{}
		_ feat.LocationSetter = &linear.QSeq{}
		_ feat.LocationSetter = &alignment.Seq{}
		_ feat.LocationSetter = &alignment.QSeq{}
		_ feat.LocationSetter = &multi.Multi{}
	)
}
