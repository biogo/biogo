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

package seq_test

import (
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/exp/seq/alignment"
	"code.google.com/p/biogo/exp/seq/linear"
	"testing"
)

func TestSeq(t *testing.T) {
	var (
		_ seq.Sequence = &linear.Seq{}
		_ seq.Sequence = &linear.QSeq{}
		_ seq.Sequence = &alignment.Seq{}
		_ seq.Sequence = &alignment.QSeq{}

		_ seq.Scorer = &linear.QSeq{}
		_ seq.Scorer = &alignment.QSeq{}
	)
}
