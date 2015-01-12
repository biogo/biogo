// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bed

import (
	"code.google.com/p/biogo/feat"
	"code.google.com/p/biogo/seq"

	"bytes"
	"fmt"
	"gopkg.in/check.v1"
	"image/color"
	"strings"
	"testing"
)

var (
	validBeds = []int{3, 4, 5, 6, 12}
	bedTests  = []struct {
		fields    int
		line      string
		skipWrite bool // We do not trail with commas on lists, but we do handle them on read.
		beds      []Bed
	}{
		{
			3, "chr1	11873	14409\n", false,
			[]Bed{
				&Bed3{"chr1", 11873, 14409},
			},
		},
		{
			4, "chr1	11873	14409	uc001aaa.3\n", false,
			[]Bed{
				&Bed3{"chr1", 11873, 14409},
				&Bed4{"chr1", 11873, 14409, "uc001aaa.3"},
			},
		},
		{
			5, "chr1	11873	14409	uc001aaa.3	3\n", false,
			[]Bed{
				&Bed3{"chr1", 11873, 14409},
				&Bed4{"chr1", 11873, 14409, "uc001aaa.3"},
				&Bed5{"chr1", 11873, 14409, "uc001aaa.3", 3},
			},
		},
		{
			6, "chr1	11873	14409	uc001aaa.3	3	+\n", false,
			[]Bed{
				&Bed3{"chr1", 11873, 14409},
				&Bed4{"chr1", 11873, 14409, "uc001aaa.3"},
				&Bed5{"chr1", 11873, 14409, "uc001aaa.3", 3},
				&Bed6{"chr1", 11873, 14409, "uc001aaa.3", 3, seq.Plus},
			},
		},
		{
			6, "chr1	11873	14409	uc001aaa.3	3	-\n", false,
			[]Bed{
				&Bed3{"chr1", 11873, 14409},
				&Bed4{"chr1", 11873, 14409, "uc001aaa.3"},
				&Bed5{"chr1", 11873, 14409, "uc001aaa.3", 3},
				&Bed6{"chr1", 11873, 14409, "uc001aaa.3", 3, seq.Minus},
			},
		},
		{
			6, "chr1	11873	14409	uc001aaa.3	3	.\n", false,
			[]Bed{
				&Bed3{"chr1", 11873, 14409},
				&Bed4{"chr1", 11873, 14409, "uc001aaa.3"},
				&Bed5{"chr1", 11873, 14409, "uc001aaa.3", 3},
				&Bed6{"chr1", 11873, 14409, "uc001aaa.3", 3, seq.None},
			},
		},
		{
			12, "chr1	11873	14409	uc001aaa.3	3	+	11873	11873	0	3	354,109,1189,	0,739,1347,\n", true,
			[]Bed{
				&Bed3{"chr1", 11873, 14409},
				&Bed4{"chr1", 11873, 14409, "uc001aaa.3"},
				&Bed5{"chr1", 11873, 14409, "uc001aaa.3", 3},
				&Bed6{"chr1", 11873, 14409, "uc001aaa.3", 3, seq.Plus},
				&Bed12{"chr1", 11873, 14409, "uc001aaa.3", 3, seq.Plus, 11873, 11873, color.RGBA{}, 3, []int{354, 109, 1189}, []int{0, 739, 1347}},
			},
		},
		{
			12, "chr1	11873	14409	uc001aaa.3	3	+	11873	11873	255,128,0	3	354,109,1189,	0,739,1347,\n", true,
			[]Bed{
				&Bed3{"chr1", 11873, 14409},
				&Bed4{"chr1", 11873, 14409, "uc001aaa.3"},
				&Bed5{"chr1", 11873, 14409, "uc001aaa.3", 3},
				&Bed6{"chr1", 11873, 14409, "uc001aaa.3", 3, seq.Plus},
				&Bed12{"chr1", 11873, 14409, "uc001aaa.3", 3, seq.Plus, 11873, 11873, color.RGBA{255, 128, 0, 255}, 3, []int{354, 109, 1189}, []int{0, 739, 1347}},
			},
		},
		{
			12, "chr1	11873	14409	uc001aaa.3	3	+	11873	11873	0	3	354,109,1189	0,739,1347\n", false,
			[]Bed{
				&Bed3{"chr1", 11873, 14409},
				&Bed4{"chr1", 11873, 14409, "uc001aaa.3"},
				&Bed5{"chr1", 11873, 14409, "uc001aaa.3", 3},
				&Bed6{"chr1", 11873, 14409, "uc001aaa.3", 3, seq.Plus},
				&Bed12{"chr1", 11873, 14409, "uc001aaa.3", 3, seq.Plus, 11873, 11873, color.RGBA{}, 3, []int{354, 109, 1189}, []int{0, 739, 1347}},
			},
		},
	}
)

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

func (s *S) TestUnsafeString(c *check.C) {
	for _, t := range []string{
		"I",
		"Lorem",
		"Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
	} {
		b := []byte(t)
		c.Check(unsafeString(b), check.Equals, t)
	}
}

func (s *S) TestReadBed(c *check.C) {
	for i, b := range bedTests {
		for j, typ := range validBeds {
			buf := strings.NewReader(b.line)
			r, err := NewReader(buf, typ)
			c.Assert(err, check.Equals, nil)
			f, err := r.Read()
			if typ <= b.fields {
				c.Check(f, check.DeepEquals, b.beds[j], check.Commentf("Test: %d type: Bed%d", i, typ))
				c.Check(err, check.Equals, nil)
			} else {
				c.Check(f, check.Equals, nil)
				c.Check(err, check.ErrorMatches, fmt.Sprintf("%s.*", ErrBadBedType), check.Commentf("Test: %d type: Bed%d", i, typ))
			}
		}
	}
}

func (s *S) TestWriteBed(c *check.C) {
	for i, b := range bedTests {
		for _, typ := range validBeds {
			buf := &bytes.Buffer{}
			w, err := NewWriter(buf, typ)
			c.Assert(err, check.Equals, nil)
			n, err := w.Write(b.beds[len(b.beds)-1])
			c.Check(n, check.Equals, buf.Len())
			if typ <= b.fields {
				trunc := strings.Join(strings.Split(b.line, "\t")[:typ], "\t")
				if trunc[len(trunc)-1] != '\n' {
					trunc += "\n"
				}
				c.Check(err, check.Equals, nil)
				if !b.skipWrite {
					c.Check(buf.String(), check.Equals, trunc, check.Commentf("Test: %d type: Bed%d", i, typ))
				}
			} else {
				c.Check(buf.String(), check.Equals, "")
				c.Check(err, check.ErrorMatches, fmt.Sprintf("%s.*", ErrBadBedType), check.Commentf("Test: %d type: Bed%d", i, typ))
			}
		}
	}
}

type tf struct {
	chrom      feat.Feature
	start, end int
	name       string
}

func (f *tf) Start() int { return f.start }
func (f *tf) End() int   { return f.end }
func (f *tf) Len() int   { return f.end - f.start }
func (f *tf) Name() string {
	if f.name == "" {
		return fmt.Sprintf("%s", f.chrom)
	}
	return f.name
}
func (f *tf) Description() string    { return "test feat" }
func (f *tf) Location() feat.Feature { return f.chrom }

type sctf struct {
	tf
	score int
}

func (f *sctf) Score() int { return f.score }

type sttf struct {
	tf
	strand seq.Strand
}

func (f *sttf) Orientation() feat.Orientation { return feat.Orientation(f.strand) }

type ctf struct {
	tf
	score  int
	strand seq.Strand
}

func (f *ctf) Score() int                    { return f.score }
func (f *ctf) Orientation() feat.Orientation { return feat.Orientation(f.strand) }

func (s *S) TestWriteFeature(c *check.C) {
	for i, f := range []struct {
		feat feat.Feature
		typ  int
		line string
		err  error
	}{
		// Vanilla feature.
		{
			&tf{chrom: Chrom("test chrom"), start: 1, end: 99}, 3,
			"test chrom\t1\t99\n", nil,
		},
		{
			&tf{chrom: Chrom("test chrom"), start: 1, end: 99, name: "test feat"}, 3,
			"test chrom\t1\t99\n", nil,
		},
		{
			&tf{chrom: Chrom("test chrom"), start: 1, end: 99, name: "test feat"}, 4,
			"test chrom\t1\t99\ttest feat\n", nil,
		},
		{
			&tf{chrom: Chrom("test chrom"), start: 1, end: 99, name: "test feat"}, 5,
			"test chrom\t1\t99\ttest feat\t0\n", nil,
		},
		{
			&tf{chrom: Chrom("test chrom"), start: 1, end: 99, name: "test feat"}, 6,
			"test chrom\t1\t99\ttest feat\t0\t.\n", nil,
		},
		{
			&tf{chrom: Chrom("test chrom"), start: 1, end: 99, name: "test feat"}, 12,
			"test chrom\t1\t99\ttest feat\t0\t.\n", ErrBadBedType,
		},

		// Scorer.
		{
			&sctf{tf: tf{chrom: Chrom("test chrom"), start: 1, end: 99}, score: 100}, 3,
			"test chrom\t1\t99\n", nil,
		},
		{
			&sctf{tf: tf{chrom: Chrom("test chrom"), start: 1, end: 99, name: "test feat"}, score: 100}, 3,
			"test chrom\t1\t99\n", nil,
		},
		{
			&sctf{tf: tf{chrom: Chrom("test chrom"), start: 1, end: 99, name: "test feat"}, score: 100}, 4,
			"test chrom\t1\t99\ttest feat\n", nil,
		},
		{
			&sctf{tf: tf{chrom: Chrom("test chrom"), start: 1, end: 99, name: "test feat"}, score: 100}, 5,
			"test chrom\t1\t99\ttest feat\t100\n", nil,
		},
		{
			&sctf{tf: tf{chrom: Chrom("test chrom"), start: 1, end: 99, name: "test feat"}, score: 100}, 6,
			"test chrom\t1\t99\ttest feat\t100\t.\n", nil,
		},
		{
			&sctf{tf: tf{chrom: Chrom("test chrom"), start: 1, end: 99, name: "test feat"}, score: 100}, 12,
			"test chrom\t1\t99\ttest feat\t100\t.\n", ErrBadBedType,
		},

		// feat.Orientater.
		{
			&sttf{tf: tf{chrom: Chrom("test chrom"), start: 1, end: 99}, strand: +1}, 3,
			"test chrom\t1\t99\n", nil,
		},
		{
			&sttf{tf: tf{chrom: Chrom("test chrom"), start: 1, end: 99, name: "test feat"}, strand: +1}, 3,
			"test chrom\t1\t99\n", nil,
		},
		{
			&sttf{tf: tf{chrom: Chrom("test chrom"), start: 1, end: 99, name: "test feat"}, strand: +1}, 4,
			"test chrom\t1\t99\ttest feat\n", nil,
		},
		{
			&sttf{tf: tf{chrom: Chrom("test chrom"), start: 1, end: 99, name: "test feat"}, strand: +1}, 5,
			"test chrom\t1\t99\ttest feat\t0\n", nil,
		},
		{
			&sttf{tf: tf{chrom: Chrom("test chrom"), start: 1, end: 99, name: "test feat"}, strand: +1}, 6,
			"test chrom\t1\t99\ttest feat\t0\t+\n", nil,
		},
		{
			&sttf{tf: tf{chrom: Chrom("test chrom"), start: 1, end: 99, name: "test feat"}, strand: +1}, 12,
			"test chrom\t1\t99\ttest feat\t0\t+\n", ErrBadBedType,
		},

		// Complete.
		{
			&ctf{tf: tf{chrom: Chrom("test chrom"), start: 1, end: 99}, score: 100, strand: +1}, 3,
			"test chrom\t1\t99\n", nil,
		},
		{
			&ctf{tf: tf{chrom: Chrom("test chrom"), start: 1, end: 99, name: "test feat"}, score: 100, strand: +1}, 3,
			"test chrom\t1\t99\n", nil,
		},
		{
			&ctf{tf: tf{chrom: Chrom("test chrom"), start: 1, end: 99, name: "test feat"}, score: 100, strand: +1}, 4,
			"test chrom\t1\t99\ttest feat\n", nil,
		},
		{
			&ctf{tf: tf{chrom: Chrom("test chrom"), start: 1, end: 99, name: "test feat"}, score: 100, strand: +1}, 5,
			"test chrom\t1\t99\ttest feat\t100\n", nil,
		},
		{
			&ctf{tf: tf{chrom: Chrom("test chrom"), start: 1, end: 99, name: "test feat"}, score: 100, strand: +1}, 6,
			"test chrom\t1\t99\ttest feat\t100\t+\n", nil,
		},
		{
			&ctf{tf: tf{chrom: Chrom("test chrom"), start: 1, end: 99, name: "test feat"}, score: 100, strand: +1}, 12,
			"test chrom\t1\t99\ttest feat\t100\t+\n", ErrBadBedType,
		},
	} {
		buf := &bytes.Buffer{}
		w, err := NewWriter(buf, f.typ)
		c.Assert(err, check.Equals, nil)
		n, err := w.Write(f.feat)
		c.Check(n, check.Equals, buf.Len())
		c.Check(err, check.Equals, f.err)
		c.Check(buf.String(), check.Equals, f.line, check.Commentf("Test: %d type: Bed%d", i, f.typ))
	}
}
