// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package seqio_test

import (
	"code.google.com/p/biogo/io/seqio"
	"code.google.com/p/biogo/io/seqio/fasta"
	"code.google.com/p/biogo/io/seqio/fastq"

	"testing"
)

func TestSeqio(t *testing.T) {
	var (
		_ seqio.Reader = (*fasta.Reader)(nil)
		_ seqio.Reader = (*fastq.Reader)(nil)
		_ seqio.Writer = (*fasta.Writer)(nil)
		_ seqio.Writer = (*fastq.Writer)(nil)
	)
}
