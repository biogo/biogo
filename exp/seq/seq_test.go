package seq_test

// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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

import (
	"github.com/kortschak/biogo/exp/seq"
	"github.com/kortschak/biogo/exp/seq/nucleic"
	na "github.com/kortschak/biogo/exp/seq/nucleic/alignment"
	pm "github.com/kortschak/biogo/exp/seq/nucleic/multi"
	"github.com/kortschak/biogo/exp/seq/nucleic/packed"
	"github.com/kortschak/biogo/exp/seq/protein"
	pa "github.com/kortschak/biogo/exp/seq/protein/alignment"
	nm "github.com/kortschak/biogo/exp/seq/protein/multi"
	"testing"
)

func TestSeq(t *testing.T) {
	var (
		_ seq.Sequence = &nucleic.Seq{}
		_ seq.Sequence = &nucleic.QSeq{}
		_ seq.Sequence = &packed.Seq{}
		_ seq.Sequence = &packed.QSeq{}
		_ seq.Sequence = &na.Seq{}
		_ seq.Sequence = &na.QSeq{}
		_ seq.Sequence = &nm.Multi{}
		_ seq.Sequence = &protein.Seq{}
		_ seq.Sequence = &protein.QSeq{}
		_ seq.Sequence = &pa.Seq{}
		_ seq.Sequence = &pa.QSeq{}
		_ seq.Sequence = &pm.Multi{}

		_ seq.Scorer = &nucleic.QSeq{}
		_ seq.Scorer = &packed.QSeq{}
		_ seq.Scorer = &na.QSeq{}
		_ seq.Scorer = &nm.Multi{}
		_ seq.Scorer = &protein.QSeq{}
		_ seq.Scorer = &pa.QSeq{}
		_ seq.Scorer = &pm.Multi{}
	)
}
