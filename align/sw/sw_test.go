package sw
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
	"github.com/kortschak/BioGo/io/seqio/fasta"
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
		swsa, _ := r.Read()
		swsb, _ := r.Read()

		swm := [][]int{
			{2, -1, -1, -1, -1},
			{-1, 2, -1, -1, -1},
			{-1, -1, 2, -1, -1},
			{-1, -1, -1, 2, -1},
			{-1, -1, -1, -1, 0},
		}

		smith := &Aligner{Matrix: swm, LookUp: LookUpN, GapChar: '-'}
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			smith.Align(swsa, swsb)
		}
	}
}
