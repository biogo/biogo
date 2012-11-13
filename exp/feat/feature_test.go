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
	"code.google.com/p/biogo/exp/seq/nucleic"
	na "code.google.com/p/biogo/exp/seq/nucleic/alignment"
	nm "code.google.com/p/biogo/exp/seq/nucleic/multi"
	"code.google.com/p/biogo/exp/seq/protein"
	pa "code.google.com/p/biogo/exp/seq/protein/alignment"
	pm "code.google.com/p/biogo/exp/seq/protein/multi"
	"testing"
)

func TestFeat(t *testing.T) {
	var (
		_ feat.Feature = &nucleic.Seq{}
		_ feat.Feature = &protein.Seq{}
		_ feat.Feature = &nucleic.QSeq{}
		_ feat.Feature = &protein.QSeq{}
		_ feat.Feature = &na.Seq{}
		_ feat.Feature = &pa.Seq{}
		_ feat.Feature = &na.QSeq{}
		_ feat.Feature = &pa.QSeq{}
		_ feat.Feature = &nm.Multi{}
		_ feat.Feature = &pm.Multi{}

		_ feat.Offsetter = &nucleic.Seq{}
		_ feat.Offsetter = &protein.Seq{}
		_ feat.Offsetter = &nucleic.QSeq{}
		_ feat.Offsetter = &protein.QSeq{}
		_ feat.Offsetter = &na.Seq{}
		_ feat.Offsetter = &pa.Seq{}
		_ feat.Offsetter = &na.QSeq{}
		_ feat.Offsetter = &pa.QSeq{}
		_ feat.Offsetter = &nm.Multi{}
		_ feat.Offsetter = &pm.Multi{}

		_ feat.LocationSetter = &nucleic.Seq{}
		_ feat.LocationSetter = &protein.Seq{}
		_ feat.LocationSetter = &nucleic.QSeq{}
		_ feat.LocationSetter = &protein.QSeq{}
		_ feat.LocationSetter = &na.Seq{}
		_ feat.LocationSetter = &pa.Seq{}
		_ feat.LocationSetter = &na.QSeq{}
		_ feat.LocationSetter = &pa.QSeq{}
		_ feat.LocationSetter = &nm.Multi{}
		_ feat.LocationSetter = &pm.Multi{}
	)
}
