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

package bed

import (
	"github.com/kortschak/biogo/bio"
	"github.com/kortschak/biogo/feat"
	"io"
	"io/ioutil"
	check "launchpad.net/gocheck"
	"os"
	"strings"
	"testing"
)

type bedTest struct {
	bName string
	bType int
}

var (
	B = []bedTest{
		{"../../testdata/test3.bed", 3},
		{"../../testdata/test4.bed", 4},
		{"../../testdata/test5.bed", 5},
		{"../../testdata/test6.bed", 6},
		{"../../testdata/test12.bed", 12},
	}
)

// Helpers
func floatPtr(f float64) *float64 { return &f }

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

var (
	expect [][]feat.Feature = [][]feat.Feature{
		{
			{ID: "chr1:11873..14409", Source: "", Location: "chr1", Start: 11873, End: 14409, Feature: "", Score: nil, Probability: nil, Attributes: "", Comments: "", Frame: 0, Strand: 0, Moltype: 0, Meta: interface{}(nil)},
		},
		{
			{ID: "uc001aaa.3", Source: "", Location: "chr1", Start: 11873, End: 14409, Feature: "", Score: nil, Probability: nil, Attributes: "", Comments: "", Frame: 0, Strand: 0, Moltype: 0, Meta: interface{}(nil)},
		},
		{
			{ID: "uc001aaa.3", Source: "", Location: "chr1", Start: 11873, End: 14409, Feature: "", Score: floatPtr(3), Probability: nil, Attributes: "", Comments: "", Frame: 0, Strand: 0, Moltype: 0, Meta: interface{}(nil)},
		},
		{
			{ID: "uc001aaa.3", Source: "", Location: "chr1", Start: 11873, End: 14409, Feature: "", Score: floatPtr(3), Probability: nil, Attributes: "", Comments: "", Frame: 0, Strand: 1, Moltype: 0, Meta: interface{}(nil)},
		},
		{
			{ID: "uc001aaa.3", Source: "", Location: "chr1", Start: 11873, End: 14409, Feature: "", Score: floatPtr(3), Probability: nil, Attributes: "", Comments: "", Frame: 0, Strand: 1, Moltype: 0, Meta: interface{}(nil)},
		},
	}
)

func (s *S) TestReadBed(c *check.C) {
	obtain := []*feat.Feature{}
	for k, b := range B {
		if r, err := NewReaderName(b.bName, b.bType); err != nil {
			c.Fatalf("Failed to open %q: %s", b.bName, err)
		} else {
			for i := 0; i < 3; i++ {
				for {
					if f, err := r.Read(); err != nil {
						if err == io.EOF {
							break
						} else {
							c.Fatalf("Failed to read %q: %s", b.bName, err)
						}
					} else {
						obtain = append(obtain, f)
					}
				}
				if c.Failed() {
					break
				}
				if len(obtain) == len(expect[k]) {
					for j := range obtain {
						c.Check(*obtain[j], check.DeepEquals, expect[k][j])
					}
				} else {
					c.Log(k, b)
					c.Check(len(obtain), check.Equals, len(expect[k]))
				}
				if err = r.Rewind(); err != nil {
					c.Fatalf("Failed to Rewind: %s", err)
				}
				obtain = nil
			}
			r.Close()
		}
	}
}

func (s *S) TestWriteBed(c *check.C) {
	bio.Precision = 0
	o := c.MkDir()
	for k, b := range B {
		if w, err := NewWriterName(o+"/b", b.bType); err != nil {
			c.Fatalf("Failed to open %q for write: %s", o+"/b", err)
		} else {
			for i := range expect[k] {
				if _, err = w.Write(&expect[k][i]); err != nil {
					c.Fatalf("Failed to write %q: %s", o+"/b", err)
				}
			}

			if err = w.Close(); err != nil {
				c.Fatalf("Failed to Close %q: %s", o+"/b", err)
			}

			var (
				of, gf *os.File
				ob, gb []byte
			)
			if of, err = os.Open(b.bName); err != nil {
				c.Fatalf("Failed to Open %q: %s", b.bName, err)
			}
			if gf, err = os.Open(o + "/b"); err != nil {
				c.Fatalf("Failed to Open %q: %s", o+"/b", err)
			}
			if ob, err = ioutil.ReadAll(of); err != nil {
				c.Fatalf("Failed to read %q: %s", b.bName, err)
			}
			if gb, err = ioutil.ReadAll(gf); err != nil {
				c.Fatalf("Failed to read %q: %s", o+"/b", err)
			}
			if b.bType < 12 {
				c.Check(string(gb), check.Equals, string(ob))
			} else {
				c.Check(strings.Split(string(gb), "\t")[:6], check.DeepEquals, strings.Split(string(ob), "\t")[:6])
			}
		}
	}
}
