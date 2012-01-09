package alignio

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
	check "launchpad.net/gocheck"
	"testing"
	"github.com/kortschak/BioGo/io/seqio/fasta"
	"github.com/kortschak/BioGo/io/seqio/fastq"
)

var (
	fa = "../testdata/testaln.fasta"
	fq = "../testdata/testaln.fastq"
)

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

func (s *S) TestReadFasta(c *check.C) {
	if r, err := fasta.NewReaderName(fa); err != nil {
		c.Fatalf("Failed to open %q: %s", fa , err)
	} else {
		if a, err := NewReader(r).Read(); err != nil {
			c.Fatalf("Failed to read %q: %s", fa , err)
		} else {
			c.Check(len(a), check.Equals, 11)
		}
	}
}

func (s *S) TestReadFastq(c *check.C) {
	if r, err := fastq.NewReaderName(fq); err != nil {
		c.Fatalf("Failed to open %q: %s", fq , err)
	} else {
		if a, err := NewReader(r).Read(); err != nil {
			c.Fatalf("Failed to read %q: %s", fq , err)
		} else {
			c.Check(len(a), check.Equals, 25)
		}
	}
}
