// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package alignio

import (
	"github.com/biogo/biogo/alphabet"
	"github.com/biogo/biogo/io/seqio/fasta"
	"github.com/biogo/biogo/io/seqio/fastq"
	"github.com/biogo/biogo/seq"
	"github.com/biogo/biogo/seq/linear"
	"github.com/biogo/biogo/seq/multi"

	"strings"
	"testing"

	"gopkg.in/check.v1"
)

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

func (s *S) TestReadFasta(c *check.C) {
	r := fasta.NewReader(strings.NewReader(fa), linear.NewSeq("", nil, alphabet.Protein))
	m, _ := multi.NewMulti("", nil, seq.DefaultConsensus)
	a, err := NewReader(r, m).Read()
	c.Check(err, check.Equals, nil)
	c.Check(a.Rows(), check.Equals, 11)
}

func (s *S) TestReadFastq(c *check.C) {
	r := fastq.NewReader(strings.NewReader(fq), linear.NewQSeq("", nil, alphabet.DNA, alphabet.Sanger))
	m, _ := multi.NewMulti("", nil, seq.DefaultQConsensus)
	a, err := NewReader(r, m).Read()
	c.Check(err, check.Equals, nil)
	c.Check(a.Rows(), check.Equals, 25)
}
