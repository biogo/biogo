// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build go1.7

package complexity

import check "gopkg.in/check.v1"

func (s *S) TestZ(c *check.C) {
	for i, t := range []struct {
		s string
		c float64
	}{
		{"", 0},
		{"aaaaaaaaaaaaaaaaaaaa", 0.15},
		{"acacacacacacacacacac", 0.2},
		{"acgtacgtacgtacgtacgt", 0.3},
		{"acgacagacagacaagatacgctcacatgctacagcagcactgatgcggactcttagctatgcagctagcatcgacatgcagcgatcagcgagc", 0.48936170212765956},
		{"cctccctaactcattttatgaggccagcatcattctgataccaaagccgggcagagacacaaccaaaaaagagaattttagaccaatatccttgatgaacattgatgcaaaaatcctcaataaaatactggcaaaccgaatccagcagcacatcaaaaagcttatccaccatgatcaagtgggcttcatccctgggatgcaaggctggttcaatatacgcaaatcaataaatgtaatccagcatataaacagagccaaagacaaaaaccacatgattatctcaatagatgcagaaaaaccctttgacaaaattcaacaacccttcatgctaaaaactctcaataaattaggtattgatgggacgtatttcaaaataataagagctatctatgacaaacccacagccaatatcatactgaatgggcaaaaactggaagcattccctttgaaaactggcacaagacagggatgccctctctcaccgctcctattcaacatag", 0.342},
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 0.01},
		{"cctccctaactcattttatgaggccagcatcattctgataccaaagcc---cagagacacaaccaaaaaagagaattttagaccaatatccttgatgaacattgatgcaaaaatcctcaataaaatactggcaaaccgaatccagcagcacatcaaaaagcttatccaccatgatcaagtgggcttcatccctgggatgcaaggctggttcaatatacgcaaatcaataaatgtaatccagcatataaacagagccaaagacaaaaaccacatgattatctcaatagatgcagaaaaaccctttgacaaaattcaacaacccttcatgctaaaaactctcaataaattaggtattgatgggacgtatttcaaaataataagagctatctatgacaaacccacagccaatatcatactgaatgggcaaaaactggaagcattccctttgaaaactggcacaagacagggatgccctctctcaccgctcctattcaacatag", 0.34004024144869216},
	} {
		zc, err := Z(stringToSeq(t.s), 0, len(t.s))
		c.Check(err, check.Equals, nil, check.Commentf("Test: %d", i))
		c.Check(zc, check.Equals, t.c, check.Commentf("Test: %d", i))
	}
}
