// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package seqio_test

import (
	"code.google.com/p/biogo/exp/seqio"
	"code.google.com/p/biogo/exp/seqio/fasta"
	"code.google.com/p/biogo/exp/seqio/fastq"
	"testing"
)

func TestSeqio(t *testing.T) {
	var (
		_ seqio.Reader = &fasta.Reader{}
		_ seqio.Reader = &fastq.Reader{}
		_ seqio.Writer = &fasta.Writer{}
		_ seqio.Writer = &fastq.Writer{}
	)
}
