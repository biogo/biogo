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

package gff

import (
	"github.com/kortschak/biogo/bio"
	"github.com/kortschak/biogo/feat"
	"github.com/kortschak/biogo/seq"
	"io"
	"io/ioutil"
	check "launchpad.net/gocheck"
	"os"
	"testing"
)

var (
	G = []string{
		"../../testdata/test.gff",
		"../../testdata/metaline.gff",
	}
)

// Helpers
func floatPtr(f float64) *float64 { return &f }

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

var (
	expect []feat.Feature = []feat.Feature{
		{ID: "SEQ1:102..105", Source: "EMBL", Location: "SEQ1", Start: 102, End: 105, Feature: "atg", Score: nil, Probability: nil, Attributes: "", Comments: "", Frame: 0, Strand: 1, Moltype: 0x0, Meta: interface{}(nil)},
		{ID: "SEQ1:102..172", Source: "EMBL", Location: "SEQ1", Start: 102, End: 172, Feature: "exon", Score: nil, Probability: nil, Attributes: "", Comments: "", Frame: 0, Strand: 1, Moltype: 0x0, Meta: interface{}(nil)},
		{ID: "SEQ1:171..173", Source: "EMBL", Location: "SEQ1", Start: 171, End: 173, Feature: "splice5", Score: nil, Probability: nil, Attributes: "", Comments: "", Frame: -1, Strand: 1, Moltype: 0x0, Meta: interface{}(nil)},
		{ID: "SEQ1:171..173", Source: "netgene", Location: "SEQ1", Start: 171, End: 173, Feature: "splice5", Score: floatPtr(0.94), Probability: nil, Attributes: "", Comments: "", Frame: -1, Strand: 1, Moltype: 0x0, Meta: interface{}(nil)},
		{ID: "SEQ1:162..182", Source: "genie", Location: "SEQ1", Start: 162, End: 182, Feature: "sp5-20", Score: floatPtr(2.3), Probability: nil, Attributes: "", Comments: "", Frame: -1, Strand: 1, Moltype: 0x0, Meta: interface{}(nil)},
		{ID: "SEQ1:167..177", Source: "genie", Location: "SEQ1", Start: 167, End: 177, Feature: "sp5-10", Score: floatPtr(2.1), Probability: nil, Attributes: "", Comments: "", Frame: -1, Strand: 1, Moltype: 0x0, Meta: interface{}(nil)},
		{ID: "SEQ2:16..19", Source: "grail", Location: "SEQ2", Start: 16, End: 19, Feature: "ATG", Score: floatPtr(2.1), Probability: nil, Attributes: "", Comments: "", Frame: 0, Strand: -1, Moltype: 0x0, Meta: interface{}(nil)},
	}
	expectMeta []interface{} = []interface{}{
		&seq.Seq{ID: "<seqname>", Seq: []byte("acggctcggattggcgctggatgatagatcagacgac..."), Offset: 0, Strand: 1, Circular: false, Moltype: 0x0, Quality: (*seq.Quality)(nil), Inplace: false, Meta: interface{}(nil)},
		&seq.Seq{ID: "<seqname>", Seq: []byte("acggcucggauuggcgcuggaugauagaucagacgac..."), Offset: 0, Strand: 1, Circular: false, Moltype: 0x1, Quality: (*seq.Quality)(nil), Inplace: false, Meta: interface{}(nil)},
		&seq.Seq{ID: "<seqname>", Seq: []byte("MVLSPADKTNVKAAWGKVGAHAGEYGAEALERMFLSF..."), Offset: 0, Strand: 1, Circular: false, Moltype: 0x2, Quality: (*seq.Quality)(nil), Inplace: false, Meta: interface{}(nil)},
		&feat.Feature{ID: "<seqname>", Source: "", Location: "", Start: 0, End: 5, Feature: "", Score: nil, Probability: nil, Attributes: "", Comments: "", Frame: 0, Strand: 0, Moltype: 0x0, Meta: interface{}(nil)},
	}
	writeMeta []interface{} = []interface{}{
		"gff-version 2",
		"source-version <source> <version-text>",
		"date Mon Jan 2 15:04:05 MST 2006",
		"Type <type> <seqname>",
		&seq.Seq{ID: "<seqname>", Seq: []byte("acggctcggattggcgctggatgatagatcagacgac..."), Offset: 0, Strand: 1, Circular: false, Moltype: 0x0, Quality: (*seq.Quality)(nil), Inplace: false, Meta: interface{}(nil)},
		&seq.Seq{ID: "<seqname>", Seq: []byte("acggcucggauuggcgcuggaugauagaucagacgac..."), Offset: 0, Strand: 1, Circular: false, Moltype: 0x1, Quality: (*seq.Quality)(nil), Inplace: false, Meta: interface{}(nil)},
		&seq.Seq{ID: "<seqname>", Seq: []byte("MVLSPADKTNVKAAWGKVGAHAGEYGAEALERMFLSF..."), Offset: 0, Strand: 1, Circular: false, Moltype: 0x2, Quality: (*seq.Quality)(nil), Inplace: false, Meta: interface{}(nil)},
		&feat.Feature{ID: "<seqname>", Source: "", Location: "", Start: 0, End: 5, Feature: "", Score: floatPtr(0), Probability: floatPtr(0), Attributes: "", Comments: "", Frame: 0, Strand: 0, Moltype: 0x0, Meta: interface{}(nil)},
	}
)

func (s *S) TestReadGFF(c *check.C) {
	obtain := []*feat.Feature{}
	if r, err := NewReaderName(G[0]); err != nil {
		c.Fatalf("Failed to open %q: %s", G[0], err)
	} else {
		for i := 0; i < 3; i++ {
			for {
				if f, err := r.Read(); err != nil {
					if err == io.EOF {
						break
					} else {
						c.Fatalf("Failed to read %q: %s", G[0], err)
					}
				} else {
					obtain = append(obtain, f)
				}
			}
			if c.Failed() {
				break
			}
			if len(obtain) == len(expect) {
				for j := range obtain {
					c.Check(*obtain[j], check.DeepEquals, expect[j])
				}
			} else {
				c.Check(len(obtain), check.Equals, len(expect))
			}
		}
		c.Check(r.Type, check.Equals, bio.Moltype(0))
		r.Close()
	}
}

func (s *S) TestReadMetaline(c *check.C) {
	obtain := []interface{}{}
	if r, err := NewReaderName(G[1]); err != nil {
		c.Fatalf("Failed to open %q: %s", G[1], err)
	} else {
		r.TimeFormat = "Mon Jan _2 15:04:05 MST 2006"
		for i := 0; i < 3; i++ {
			for {
				if f, err := r.Read(); err != nil {
					if err == io.EOF {
						break
					} else {
						c.Fatalf("Failed to read %q: %s", G[1], err)
					}
				} else {
					obtain = append(obtain, f.Meta)
				}
			}
			if c.Failed() {
				break
			}
			if len(obtain) == len(expectMeta) {
				for j := range obtain {
					c.Check(obtain[j], check.DeepEquals, expectMeta[j])
				}
			} else {
				c.Check(len(obtain), check.Equals, len(expectMeta))
			}
			obtain = nil
			if err = r.Rewind(); err != nil {
				c.Fatalf("Failed to rewind %s", err)
			}
		}
		c.Check(r.SourceVersion, check.Equals, "<source> <version-text>")
		c.Check(r.Date.Format(r.TimeFormat), check.Equals, "Mon Jan  2 15:04:05 MST 2006")
		c.Check(r.Type, check.Equals, bio.Undefined)
		r.Close()
	}
}

func (s *S) TestWriteGFF(c *check.C) {
	bio.Precision = -1
	g := G[0]
	o := c.MkDir()
	expectSize := 224
	var total int
	if w, err := NewWriterName(o+"/g", 2, 60, false); err != nil {
		c.Fatalf("Failed to open %q for write: %s", o+"/g", err)
	} else {
		for i := range expect {
			if n, err := w.Write(&expect[i]); err != nil {
				c.Fatalf("Failed to write %q: %s", o+"/g", err)
			} else {
				total += n
			}
		}

		if err = w.Close(); err != nil {
			c.Fatalf("Failed to Close %q: %s", o+"/g", err)
		}
		c.Check(total, check.Equals, expectSize)
		total = 0

		var (
			of, gf *os.File
			ob, gb []byte
		)
		if of, err = os.Open(g); err != nil {
			c.Fatalf("Failed to Open %q: %s", g, err)
		}
		if gf, err = os.Open(o + "/g"); err != nil {
			c.Fatalf("Failed to Open %q: %s", o+"/g", err)
		}
		if ob, err = ioutil.ReadAll(of); err != nil {
			c.Fatalf("Failed to read %q: %s", g, err)
		}
		if gb, err = ioutil.ReadAll(gf); err != nil {
			c.Fatalf("Failed to read %q: %s", o+"/g", err)
		}

		c.Check(string(gb), check.Equals, string(ob))
	}
}

func (s *S) TestWriteMetaline(c *check.C) {
	bio.Precision = -1
	g := G[1]
	o := c.MkDir()
	expectSize := 372
	var total int
	if w, err := NewWriterName(o+"/g", 2, 37, false); err != nil { // 37 magic number enforces linebreaks on sequence to match examples from http://www.sanger.ac.uk/resources/software/gff/spec.html
		c.Fatalf("Failed to open %q for write: %s", o+"/g", err)
	} else {
		for i := range writeMeta {
			if n, err := w.WriteMetaData(writeMeta[i]); err != nil {
				c.Fatalf("Failed to write %q: %s", o+"/g", err)
			} else {
				total += n
			}
		}

		if err = w.Close(); err != nil {
			c.Fatalf("Failed to Close %q: %s", o+"/g", err)
		}
		c.Check(total, check.Equals, expectSize)
		total = 0

		var (
			of, gf *os.File
			ob, gb []byte
		)
		if of, err = os.Open(g); err != nil {
			c.Fatalf("Failed to Open %q: %s", g, err)
		}
		if gf, err = os.Open(o + "/g"); err != nil {
			c.Fatalf("Failed to Open %q: %s", o+"/g", err)
		}
		if ob, err = ioutil.ReadAll(of); err != nil {
			c.Fatalf("Failed to read %q: %s", g, err)
		}
		if gb, err = ioutil.ReadAll(gf); err != nil {
			c.Fatalf("Failed to read %q: %s", o+"/g", err)
		}

		c.Check(string(gb), check.Equals, string(ob))
	}
}
