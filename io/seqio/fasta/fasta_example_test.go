// Copyright ©2020 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fasta_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/biogo/biogo/alphabet"
	"github.com/biogo/biogo/io/seqio"
	"github.com/biogo/biogo/io/seqio/fasta"
	"github.com/biogo/biogo/seq/linear"
)

func ExampleReader() {
	const multiFasta = `
	>SequenceA dam methylation site
	GATC
	>SequenceB ori motif
	CTAG
	>SequenceC CTCF binding motif
	CCGCGNGGNGGCAG`

	data := strings.NewReader(multiFasta)

	// fasta.Reader requires a known type template to fill
	// with FASTA data. Here we use *linear.Seq.
	template := linear.NewSeq("", nil, alphabet.DNAredundant)
	r := fasta.NewReader(data, template)

	// Make a seqio.Scanner to simplify iterating over a
	// stream of data.
	sc := seqio.NewScanner(r)

	// Iterate through each sequence in a multifasta and examine the
	// ID, description and sequence data.
	for sc.Next() {
		// Get the current sequence and type assert to *linear.Seq.
		// While this is unnecessary here, it can be useful to have
		// the concrete type.
		s := sc.Seq().(*linear.Seq)

		// Print the sequence ID, description and sequence data.
		fmt.Printf("%q %q %s\n", s.ID, s.Desc, s.Seq)
	}
	if err := sc.Error(); err != nil {
		log.Fatal(err)
	}

	// Output:
	// "SequenceA" "dam methylation site" GATC
	// "SequenceB" "ori motif" CTAG
	// "SequenceC" "CTCF binding motif" CCGCGNGGNGGCAG
}
