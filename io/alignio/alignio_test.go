// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package alignio

import (
	"code.google.com/p/biogo/io/seqio/fasta"
	"code.google.com/p/biogo/io/seqio/fastq"
	check "launchpad.net/gocheck"
	"testing"
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
		c.Fatalf("Failed to open %q: %s", fa, err)
	} else {
		if a, err := NewReader(r).Read(); err != nil {
			c.Fatalf("Failed to read %q: %s", fa, err)
		} else {
			c.Check(len(a), check.Equals, 11)
		}
	}
}

func (s *S) TestReadFastq(c *check.C) {
	if r, err := fastq.NewReaderName(fq); err != nil {
		c.Fatalf("Failed to open %q: %s", fq, err)
	} else {
		if a, err := NewReader(r).Read(); err != nil {
			c.Fatalf("Failed to read %q: %s", fq, err)
		} else {
			c.Check(len(a), check.Equals, 25)
		}
	}
}
