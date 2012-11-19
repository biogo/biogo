// Copyright ©2011-2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

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
