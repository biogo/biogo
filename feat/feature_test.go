// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package feat_test

import (
	"code.google.com/p/biogo/feat"
	"code.google.com/p/biogo/seq/alignment"
	"code.google.com/p/biogo/seq/linear"
	"code.google.com/p/biogo/seq/multi"
	"testing"
)

func TestFeat(t *testing.T) {
	var (
		_ feat.Feature = (*linear.Seq)(nil)
		_ feat.Feature = (*linear.QSeq)(nil)
		_ feat.Feature = (*alignment.Seq)(nil)
		_ feat.Feature = (*alignment.QSeq)(nil)
		_ feat.Feature = (*multi.Multi)(nil)

		_ feat.Offsetter = (*linear.Seq)(nil)
		_ feat.Offsetter = (*linear.QSeq)(nil)
		_ feat.Offsetter = (*alignment.Seq)(nil)
		_ feat.Offsetter = (*alignment.QSeq)(nil)
		_ feat.Offsetter = (*multi.Multi)(nil)

		_ feat.LocationSetter = (*linear.Seq)(nil)
		_ feat.LocationSetter = (*linear.QSeq)(nil)
		_ feat.LocationSetter = (*alignment.Seq)(nil)
		_ feat.LocationSetter = (*alignment.QSeq)(nil)
		_ feat.LocationSetter = (*multi.Multi)(nil)
	)
}
