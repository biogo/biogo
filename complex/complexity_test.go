// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package complexity

import (
	"code.google.com/p/biogo/alphabet"
	"code.google.com/p/biogo/seq/linear"
	"gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

// Helper
func stringToSeq(s string) *linear.Seq {
	return linear.NewSeq("", alphabet.BytesToLetters([]byte(s)), alphabet.DNA)
}

// Tests
func (s *S) TestEntropic(c *check.C) {
	for i, t := range []struct {
		s string
		c float64
	}{
		{"", 0},
		{"aaaaaaaaaaaaaaaaaaaa", 0},
		{"acacacacacacacacacac", 0.5},
		{"acgtacgtacgtacgtacgt", 1},
		{"acgacagacagacaagatacgctcacatgctacagcagcactgatgcggactcttagctatgcagctagcatcgacatgcagcgatcagcgagc", 0.9799077484954553},
		{"cctccctaactcattttatgaggccagcatcattctgataccaaagccgggcagagacacaaccaaaaaagagaattttagaccaatatccttgatgaacattgatgcaaaaatcctcaataaaatactggcaaaccgaatccagcagcacatcaaaaagcttatccaccatgatcaagtgggcttcatccctgggatgcaaggctggttcaatatacgcaaatcaataaatgtaatccagcatataaacagagccaaagacaaaaaccacatgattatctcaatagatgcagaaaaaccctttgacaaaattcaacaacccttcatgctaaaaactctcaataaattaggtattgatgggacgtatttcaaaataataagagctatctatgacaaacccacagccaatatcatactgaatgggcaaaaactggaagcattccctttgaaaactggcacaagacagggatgccctctctcaccgctcctattcaacatag", 0.9610297459211902},
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 0},
		{"cctccctaactcattttatgaggccagcatcattctgataccaaagcc---cagagacacaaccaaaaaagagaattttagaccaatatccttgatgaacattgatgcaaaaatcctcaataaaatactggcaaaccgaatccagcagcacatcaaaaagcttatccaccatgatcaagtgggcttcatccctgggatgcaaggctggttcaatatacgcaaatcaataaatgtaatccagcatataaacagagccaaagacaaaaaccacatgattatctcaatagatgcagaaaaaccctttgacaaaattcaacaacccttcatgctaaaaactctcaataaattaggtattgatgggacgtatttcaaaataataagagctatctatgacaaacccacagccaatatcatactgaatgggcaaaaactggaagcattccctttgaaaactggcacaagacagggatgccctctctcaccgctcctattcaacatag", 0.958612004852684},
	} {
		ec, err := Entropic(stringToSeq(t.s), 0, len(t.s))
		c.Check(err, check.Equals, nil, check.Commentf("Test: %d", i))
		c.Check(ec, check.Equals, t.c, check.Commentf("Test: %d", i))
	}
}

func (s *S) TestWF(c *check.C) {
	for i, t := range []struct {
		s string
		c float64
	}{
		{"", 0},
		{"aaaaaaaaaaaaaaaaaaaa", 0},
		{"acacacacacacacacacac", 0.4373815422868077},
		{"acgtacgtacgtacgtacgt", 0.8362455384618038},
		{"acgacagacagacaagatacgctcacatgctacagcagcactgatgcggactcttagctatgcagctagcatcgacatgcagcgatcagcgagc", 0.9280876953247041},
		{"cctccctaactcattttatgaggccagcatcattctgataccaaagccgggcagagacacaaccaaaaaagagaattttagaccaatatccttgatgaacattgatgcaaaaatcctcaataaaatactggcaaaccgaatccagcagcacatcaaaaagcttatccaccatgatcaagtgggcttcatccctgggatgcaaggctggttcaatatacgcaaatcaataaatgtaatccagcatataaacagagccaaagacaaaaaccacatgattatctcaatagatgcagaaaaaccctttgacaaaattcaacaacccttcatgctaaaaactctcaataaattaggtattgatgggacgtatttcaaaataataagagctatctatgacaaacccacagccaatatcatactgaatgggcaaaaactggaagcattccctttgaaaactggcacaagacagggatgccctctctcaccgctcctattcaacatag", 0.9477548340774316},
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 0},
		{"cctccctaactcattttatgaggccagcatcattctgataccaaagcc---cagagacacaaccaaaaaagagaattttagaccaatatccttgatgaacattgatgcaaaaatcctcaataaaatactggcaaaccgaatccagcagcacatcaaaaagcttatccaccatgatcaagtgggcttcatccctgggatgcaaggctggttcaatatacgcaaatcaataaatgtaatccagcatataaacagagccaaagacaaaaaccacatgattatctcaatagatgcagaaaaaccctttgacaaaattcaacaacccttcatgctaaaaactctcaataaattaggtattgatgggacgtatttcaaaataataagagctatctatgacaaacccacagccaatatcatactgaatgggcaaaaactggaagcattccctttgaaaactggcacaagacagggatgccctctctcaccgctcctattcaacatag", 0.9452813728370209},
	} {
		wfc, err := WF(stringToSeq(t.s), 0, len(t.s))
		c.Check(err, check.Equals, nil, check.Commentf("Test: %d", i))
		c.Check(wfc, check.Equals, t.c, check.Commentf("Test: %d", i))
	}
}

func (s *S) TestZ(c *check.C) {
	for i, t := range []struct {
		s string
		c float64
	}{
		{"", 0},
		{"aaaaaaaaaaaaaaaaaaaa", 0.15},
		{"acacacacacacacacacac", 0.2},
		{"acgtacgtacgtacgtacgt", 0.3},
		{"acgacagacagacaagatacgctcacatgctacagcagcactgatgcggactcttagctatgcagctagcatcgacatgcagcgatcagcgagc", 0.5},
		{"cctccctaactcattttatgaggccagcatcattctgataccaaagccgggcagagacacaaccaaaaaagagaattttagaccaatatccttgatgaacattgatgcaaaaatcctcaataaaatactggcaaaccgaatccagcagcacatcaaaaagcttatccaccatgatcaagtgggcttcatccctgggatgcaaggctggttcaatatacgcaaatcaataaatgtaatccagcatataaacagagccaaagacaaaaaccacatgattatctcaatagatgcagaaaaaccctttgacaaaattcaacaacccttcatgctaaaaactctcaataaattaggtattgatgggacgtatttcaaaataataagagctatctatgacaaacccacagccaatatcatactgaatgggcaaaaactggaagcattccctttgaaaactggcacaagacagggatgccctctctcaccgctcctattcaacatag", 0.358},
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 0.01},
		{"cctccctaactcattttatgaggccagcatcattctgataccaaagcc---cagagacacaaccaaaaaagagaattttagaccaatatccttgatgaacattgatgcaaaaatcctcaataaaatactggcaaaccgaatccagcagcacatcaaaaagcttatccaccatgatcaagtgggcttcatccctgggatgcaaggctggttcaatatacgcaaatcaataaatgtaatccagcatataaacagagccaaagacaaaaaccacatgattatctcaatagatgcagaaaaaccctttgacaaaattcaacaacccttcatgctaaaaactctcaataaattaggtattgatgggacgtatttcaaaataataagagctatctatgacaaacccacagccaatatcatactgaatgggcaaaaactggaagcattccctttgaaaactggcacaagacagggatgccctctctcaccgctcctattcaacatag", 0.35412474849094566},
	} {
		zc, err := Z(stringToSeq(t.s), 0, len(t.s))
		c.Check(err, check.Equals, nil, check.Commentf("Test: %d", i))
		c.Check(zc, check.Equals, t.c, check.Commentf("Test: %d", i))
	}
}

func (s *S) TestLnFac(c *check.C) {
	const tolerance = 1e-9
	table := genLnFac(tableLength * 100)
	for x, exact := range table {
		if exact == 0 {
			c.Check(lnFac(x), check.Equals, exact)
		} else {
			c.Check(exact/lnFac(x)-1 < tolerance, check.Equals, true)
		}
	}
}
