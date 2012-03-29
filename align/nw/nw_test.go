package nw

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
	"github.com/kortschak/biogo/io/seqio/fasta"
	check "launchpad.net/gocheck"
	"testing"
)

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

func (s *S) TestXXX(c *check.C) {
}

func BenchmarkAlign(b *testing.B) {
	b.StopTimer()
	if r, err := fasta.NewReaderName("../testdata/crsp.fa"); err != nil {
		return
	} else {
		nwsa, _ := r.Read()
		nwsb, _ := r.Read()

		nwm := [][]int{
			{10, -3, -1, -4, -5},
			{-3, 9, -5, 0, -5},
			{-1, -5, 7, -3, -5},
			{-4, 0, -3, 8, -5},
			{-4, -4, -4, -4, 0},
		}

		needle := &Aligner{Matrix: nwm, LookUp: LookUpN, GapChar: '-'}
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			needle.Align(nwsa, nwsb)
		}
	}
}
