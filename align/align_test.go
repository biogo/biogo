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

package align

import (
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/seq/linear"
	"code.google.com/p/biogo/exp/seqio/fasta"
	check "launchpad.net/gocheck"
	"strings"
	"testing"
)

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

func (s *S) TestXXX(c *check.C) {
}

func BenchmarkSWAlign(b *testing.B) {
	b.StopTimer()
	t := &linear.Seq{}
	t.Alpha = alphabet.DNA
	r := fasta.NewReader(strings.NewReader(crspFa), t)
	swsa, _ := r.Read()
	swsb, _ := r.Read()

	smith := SW{
		{2, -1, -1, -1, -1},
		{-1, 2, -1, -1, -1},
		{-1, -1, 2, -1, -1},
		{-1, -1, -1, 2, -1},
		{-1, -1, -1, -1, 0},
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		smith.Align(swsa, swsb)
	}
}

func BenchmarkNWAlign(b *testing.B) {
	b.StopTimer()
	t := &linear.Seq{}
	t.Alpha = alphabet.DNA
	r := fasta.NewReader(strings.NewReader(crspFa), t)
	nwsa, _ := r.Read()
	nwsb, _ := r.Read()

	needle := NW{
		{10, -3, -1, -4, -5},
		{-3, 9, -5, 0, -5},
		{-1, -5, 7, -3, -5},
		{-4, 0, -3, 8, -5},
		{-4, -4, -4, -4, 0},
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		needle.Align(nwsa, nwsb)
	}
}
