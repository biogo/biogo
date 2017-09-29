// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package align

import (
	"fmt"
	"strings"
	"testing"

	"github.com/biogo/biogo/alphabet"
	"github.com/biogo/biogo/io/seqio/fasta"
	"github.com/biogo/biogo/seq/linear"
	"gopkg.in/check.v1"
)

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

func (s *S) TestWarning(c *check.C) { c.Log("\nFIXME: Tests only in example tests.\n") }

// https://github.com/biogo/biogo/issues/58
func (s *S) TestIssue58(c *check.C) {
	a := "GCTCACTAAAAACACAATCTACAACAGACGTTGCACTAACACTGTAATTGCCTTTAGTCC"
	b := "ACTGCGTA"

	nwsa := &linear.Seq{Seq: alphabet.BytesToLetters([]byte(a))}
	nwsa.Alpha = alphabet.DNAgapped
	nwsb := &linear.Seq{Seq: alphabet.BytesToLetters([]byte(b))}
	nwsb.Alpha = alphabet.DNAgapped

	needle := NWAffine{
		Matrix: Linear{
			{0, -1, -1, -1, -1},
			{-1, 1, -1, -1, -1},
			{-1, -1, 1, -1, -1},
			{-1, -1, -1, 1, -1},
			{-1, -1, -1, -1, 1},
		},
		GapOpen: -1,
	}

	aln, err := needle.Align(nwsa, nwsb)
	c.Check(err, check.Equals, nil)
	c.Check(fmt.Sprint(aln), check.Equals, "[[0,4)/-=-5 [4,7)/[0,3)=3 [7,32)/-=-26 [32,34)/[3,5)=2 [34,43)/-=-10 [43,46)/[5,8)=3 [46,60)/-=-15]")
}

func BenchmarkSWAlign(b *testing.B) {
	t := &linear.Seq{}
	t.Alpha = alphabet.DNAgapped
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
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		smith.Align(swsa, swsb)
	}
}

func BenchmarkNWAlign(b *testing.B) {
	t := &linear.Seq{}
	t.Alpha = alphabet.DNAgapped
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		needle.Align(nwsa, nwsb)
	}
}

func BenchmarkSWAffineAlign(b *testing.B) {
	t := &linear.Seq{}
	t.Alpha = alphabet.DNAgapped
	r := fasta.NewReader(strings.NewReader(crspFa), t)
	swsa, _ := r.Read()
	swsb, _ := r.Read()

	smith := SWAffine{
		Matrix: Linear{
			{2, -1, -1, -1, -1},
			{-1, 2, -1, -1, -1},
			{-1, -1, 2, -1, -1},
			{-1, -1, -1, 2, -1},
			{-1, -1, -1, -1, 0},
		},
		GapOpen: -5,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		smith.Align(swsa, swsb)
	}
}

func BenchmarkNWAffineAlign(b *testing.B) {
	t := &linear.Seq{}
	t.Alpha = alphabet.DNAgapped
	r := fasta.NewReader(strings.NewReader(crspFa), t)
	nwsa, _ := r.Read()
	nwsb, _ := r.Read()

	needle := NWAffine{
		Matrix: Linear{
			{10, -3, -1, -4, -5},
			{-3, 9, -5, 0, -5},
			{-1, -5, 7, -3, -5},
			{-4, 0, -3, 8, -5},
			{-4, -4, -4, -4, 0},
		},
		GapOpen: -10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		needle.Align(nwsa, nwsb)
	}
}
