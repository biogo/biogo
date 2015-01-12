// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fastq

import (
	"code.google.com/p/biogo/alphabet"
	"code.google.com/p/biogo/seq/linear"

	"bytes"
	"gopkg.in/check.v1"
	"io"
	"testing"
)

// Helpers
func constructQL(l [][]alphabet.Letter, q [][]alphabet.Qphred) (ql []alphabet.QLetters) {
	if len(l) != len(q) {
		panic("test data length mismatch")
	}
	ql = make([]alphabet.QLetters, len(l))
	for i := range ql {
		if len(l[i]) != len(q[i]) {
			panic("test data length mismatch")
		}
		if len(l[i]) == 0 {
			continue
		}
		ql[i] = make(alphabet.QLetters, len(l[i]))
		for j := range ql[i] {
			ql[i][j] = alphabet.QLetter{L: l[i][j], Q: q[i][j]}
		}
	}

	return
}

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

var (
	expectedIds = []string{
		"FC12044_91407_8_200_981_857",
		"FC12044_91407_8_200_8_865",
		"FC12044_91407_8_200_292_484",
		"FC12044_91407_8_200_675_16",
		"FC12044_91407_8_200_285_136",
	}

	expectedQLetters = constructQL(
		[][]alphabet.Letter{
			[]alphabet.Letter("AACGAGGGGCGCGACTTGACCTTGG"),
			[]alphabet.Letter("TTTCCCACCCCAGGAAGCCTTGGAC"),
			[]alphabet.Letter("TCAGCCTCCGTGCCCAGCCCACTCC"),
			[]alphabet.Letter("CTCGGGAGGCTGAGGCAGGGGGGTT"),
			[]alphabet.Letter("CCAAATCTTGAATTGTAGCTCCCCT"),
		},
		[][]alphabet.Qphred{
			{49, 55, 44, 50, 50, 55, 55, 55, 55, 50, 55, 48, 55, 48, 55, 37, 50, 55, 48, 37, 48, 42, 44, 55, 50},
			{55, 55, 55, 37, 42, 46, 49, 46, 44, 42, 46, 46, 49, 44, 40, 44, 49, 40, 40, 42, 42, 46, 49, 37, 37},
			{55, 48, 55, 46, 50, 55, 55, 55, 55, 55, 52, 55, 55, 55, 55, 40, 55, 55, 55, 55, 48, 51, 46, 55, 37},
			{46, 55, 51, 55, 55, 55, 50, 55, 55, 48, 55, 55, 46, 55, 55, 42, 44, 55, 55, 44, 55, 46, 42, 48, 37},
			{46, 50, 55, 46, 48, 55, 55, 55, 55, 55, 50, 55, 55, 52, 55, 55, 51, 55, 55, 55, 55, 51, 49, 44, 50},
		},
	)

	plusStart = constructQL(
		[][]alphabet.Letter{
			[]alphabet.Letter("AACGAGGGGCGCGACTTGACCTTGG"),
		},
		[][]alphabet.Qphred{
			{10, 55, 44, 50, 50, 55, 55, 55, 55, 50, 55, 48, 55, 48, 55, 37, 50, 55, 48, 37, 48, 42, 44, 55, 50},
		},
	)
	atStart = constructQL(
		[][]alphabet.Letter{
			[]alphabet.Letter("AACGAGGGGCGCGACTTGACCTTGG"),
		},
		[][]alphabet.Qphred{
			{31, 55, 44, 50, 50, 55, 55, 55, 55, 50, 55, 48, 55, 48, 55, 37, 50, 55, 48, 37, 48, 42, 44, 55, 50},
		},
	)
)

var (
	fqTests = []struct {
		fq       string
		verbatim bool
		ids      []string
		seqs     []alphabet.QLetters
	}{
		{
			fq: `@FC12044_91407_8_200_981_857
AACGAGGGGCGCGACTTGACCTTGG
+FC12044_91407_8_200_981_857
RXMSSXXXXSXQXQXFSXQFQKMXS
@FC12044_91407_8_200_8_865
TTTCCCACCCCAGGAAGCCTTGGAC
+FC12044_91407_8_200_8_865
XXXFKOROMKOORMIMRIIKKORFF
@FC12044_91407_8_200_292_484
TCAGCCTCCGTGCCCAGCCCACTCC
+FC12044_91407_8_200_292_484
XQXOSXXXXXUXXXXIXXXXQTOXF
@FC12044_91407_8_200_675_16
CTCGGGAGGCTGAGGCAGGGGGGTT
+FC12044_91407_8_200_675_16
OXTXXXSXXQXXOXXKMXXMXOKQF
@FC12044_91407_8_200_285_136
CCAAATCTTGAATTGTAGCTCCCCT
+FC12044_91407_8_200_285_136
OSXOQXXXXXSXXUXXTXXXXTRMS
`,
			verbatim: true,
			ids:      expectedIds,
			seqs: []alphabet.QLetters{
				expectedQLetters[0],
				expectedQLetters[1],
				expectedQLetters[2],
				expectedQLetters[3],
				expectedQLetters[4],
			},
		},
		{
			fq: `@FC12044_91407_8_200_981_857
AACGAGGGGCGCGACTTGACCTTGG
+FC12044_91407_8_200_981_857
@XMSSXXXXSXQXQXFSXQFQKMXS
@FC12044_91407_8_200_8_865
TTTCCCACCCCAGGAAGCCTTGGAC
+FC12044_91407_8_200_8_865
XXXFKOROMKOORMIMRIIKKORFF
@FC12044_91407_8_200_292_484
TCAGCCTCCGTGCCCAGCCCACTCC
+FC12044_91407_8_200_292_484
XQXOSXXXXXUXXXXIXXXXQTOXF
@FC12044_91407_8_200_675_16
CTCGGGAGGCTGAGGCAGGGGGGTT
+FC12044_91407_8_200_675_16
OXTXXXSXXQXXOXXKMXXMXOKQF
@FC12044_91407_8_200_285_136
CCAAATCTTGAATTGTAGCTCCCCT
+FC12044_91407_8_200_285_136
OSXOQXXXXXSXXUXXTXXXXTRMS
`,
			verbatim: true,
			ids:      expectedIds,
			seqs: []alphabet.QLetters{
				atStart[0],
				expectedQLetters[1],
				expectedQLetters[2],
				expectedQLetters[3],
				expectedQLetters[4],
			},
		},
		{
			fq: `@FC12044_91407_8_200_981_857
AACGAGGGGCGCGACTTGACCTTGG
+FC12044_91407_8_200_981_857
+XMSSXXXXSXQXQXFSXQFQKMXS
@FC12044_91407_8_200_8_865
TTTCCCACCCCAGGAAGCCTTGGAC
+FC12044_91407_8_200_8_865
XXXFKOROMKOORMIMRIIKKORFF
@FC12044_91407_8_200_292_484
TCAGCCTCCGTGCCCAGCCCACTCC
+FC12044_91407_8_200_292_484
XQXOSXXXXXUXXXXIXXXXQTOXF
@FC12044_91407_8_200_675_16
CTCGGGAGGCTGAGGCAGGGGGGTT
+FC12044_91407_8_200_675_16
OXTXXXSXXQXXOXXKMXXMXOKQF
@FC12044_91407_8_200_285_136
CCAAATCTTGAATTGTAGCTCCCCT
+FC12044_91407_8_200_285_136
OSXOQXXXXXSXXUXXTXXXXTRMS
`,
			verbatim: true,
			ids:      expectedIds,
			seqs: []alphabet.QLetters{
				plusStart[0],
				expectedQLetters[1],
				expectedQLetters[2],
				expectedQLetters[3],
				expectedQLetters[4],
			},
		},
		{
			fq: `@FC12044_91407_8_200_981_857
AACGAGGGGCGCGACTTGACCTTGG
+FC12044_91407_8_200_981_857
RXMSSXXXXSXQXQXFSXQFQKMXS
@FC12044_91407_8_200_8_865
TTTCCCACCCCAGGAAGCCTTGGAC
+FC12044_91407_8_200_8_865
XXXFKOROMKOORMIMRIIKKORFF
@FC12044_91407_8_200_292_484
TCAGCCTCCGTGCCCAGCCCACTCC
+FC12044_91407_8_200_292_484
XQXOSXXXXXUXXXXIXXXXQTOXF
@FC12044_91407_8_200_675_16
CTCGGGAGGCTGAGGCAGGGGGGTT
+FC12044_91407_8_200_675_16
OXTXXXSXXQXXOXXKMXXMXOKQF
@FC12044_91407_8_200_285_136

+FC12044_91407_8_200_285_136

`,
			verbatim: true,
			ids:      expectedIds,
			seqs: []alphabet.QLetters{
				expectedQLetters[0],
				expectedQLetters[1],
				expectedQLetters[2],
				expectedQLetters[3],
				nil,
			},
		},
		{
			fq: `@FC12044_91407_8_200_981_857

+FC12044_91407_8_200_981_857

@FC12044_91407_8_200_8_865
TTTCCCACCCCAGGAAGCCTTGGAC
+FC12044_91407_8_200_8_865
XXXFKOROMKOORMIMRIIKKORFF
@FC12044_91407_8_200_292_484
TCAGCCTCCGTGCCCAGCCCACTCC
+FC12044_91407_8_200_292_484
XQXOSXXXXXUXXXXIXXXXQTOXF
@FC12044_91407_8_200_675_16
CTCGGGAGGCTGAGGCAGGGGGGTT
+FC12044_91407_8_200_675_16
OXTXXXSXXQXXOXXKMXXMXOKQF
@FC12044_91407_8_200_285_136
CCAAATCTTGAATTGTAGCTCCCCT
+FC12044_91407_8_200_285_136
OSXOQXXXXXSXXUXXTXXXXTRMS
`,
			verbatim: true,
			ids:      expectedIds,
			seqs: []alphabet.QLetters{
				nil,
				expectedQLetters[1],
				expectedQLetters[2],
				expectedQLetters[3],
				expectedQLetters[4],
			},
		},
		{
			fq: `@FC12044_91407_8_200_981_857
AACGAGGGGCGCGACTTGACCTTGG
+FC12044_91407_8_200_981_857
RXMSSXXXXSXQXQXFSXQFQKMXS
@FC12044_91407_8_200_8_865

+FC12044_91407_8_200_8_865

@FC12044_91407_8_200_292_484
TCAGCCTCCGTGCCCAGCCCACTCC
+FC12044_91407_8_200_292_484
XQXOSXXXXXUXXXXIXXXXQTOXF
@FC12044_91407_8_200_675_16
CTCGGGAGGCTGAGGCAGGGGGGTT
+FC12044_91407_8_200_675_16
OXTXXXSXXQXXOXXKMXXMXOKQF
@FC12044_91407_8_200_285_136
CCAAATCTTGAATTGTAGCTCCCCT
+FC12044_91407_8_200_285_136
OSXOQXXXXXSXXUXXTXXXXTRMS
`,
			verbatim: true,
			ids:      expectedIds,
			seqs: []alphabet.QLetters{
				expectedQLetters[0],
				nil,
				expectedQLetters[2],
				expectedQLetters[3],
				expectedQLetters[4],
			},
		},
		{
			fq: `@FC12044_91407_8_200_981_857
AACGAGGGGCGCGACTTGACCTTGG
+FC12044_91407_8_200_981_857
RXMSSXXXXSXQXQXFSXQFQKMXS

@FC12044_91407_8_200_8_865
TTTCCCACCCCAGGAAGCCTTGGAC
+FC12044_91407_8_200_8_865

XXXFKOROMKOORMIMRIIKKORFF
@FC12044_91407_8_200_292_484

TCAGCCTCCGTGCCCAGCCCACTCC

+FC12044_91407_8_200_292_484
XQXOSXXXXXUXXXXIXXXXQTOXF
@FC12044_91407_8_200_675_16

CTCGGGAGGCTGAGGCAGGGGGGTT
+FC12044_91407_8_200_675_16
OXTXXXSXXQXXOXXKMXXMXOKQF
@FC12044_91407_8_200_285_136
CCAAATCTTGAATTGTAGCTCCCCT
+FC12044_91407_8_200_285_136

OSXOQXXXXXSXXUXXTXXXXTRMS`,
			verbatim: false,
			ids:      expectedIds,
			seqs: []alphabet.QLetters{
				expectedQLetters[0],
				expectedQLetters[1],
				expectedQLetters[2],
				expectedQLetters[3],
				expectedQLetters[4],
			},
		},
		{
			fq: `@FC12044_91407_8_200_981_857
AACGAGGGGCGCGACTTGACCTTGG
+FC12044_91407_8_200_981_857
RXMSSXXXXSXQXQXFSXQFQKMXS

@FC12044_91407_8_200_8_865
TTTCCCACCCCAGGAAGCCTTGGAC
+FC12044_91407_8_200_8_865

XXXFKOROMKOORMIMRIIKKORFF
@FC12044_91407_8_200_292_484

TCAGCCTCCGTGCCCAGCCCACTCC

+FC12044_91407_8_200_292_484
XQXOSXXXXXUXXXXIXXXXQTOXF
@FC12044_91407_8_200_675_16

CTCGGGAGGCTGAGGCAGGGGGGTT
+FC12044_91407_8_200_675_16
OXTXXXSXXQXXOXXKMXXMXOKQF
@FC12044_91407_8_200_285_136

+FC12044_91407_8_200_285_136

`,
			verbatim: false,
			ids:      expectedIds,
			seqs: []alphabet.QLetters{
				expectedQLetters[0],
				expectedQLetters[1],
				expectedQLetters[2],
				expectedQLetters[3],
				nil,
			},
		},
		{
			fq: `@FC12044_91407_8_200_981_857
AACGAGGGGCGCGACTTGACCTTGG
+FC12044_91407_8_200_981_857
RXMSSXXXXSXQXQXFSXQFQKMXS

@FC12044_91407_8_200_8_865
TTTCCCACCCCAGGAAGCCTTGGAC
+FC12044_91407_8_200_8_865

XXXFKOROMKOORMIMRIIKKORFF
@FC12044_91407_8_200_292_484

TCAGCCTCCGTGCCCAGCCCACTCC

+FC12044_91407_8_200_292_484
XQXOSXXXXXUXXXXIXXXXQTOXF
@FC12044_91407_8_200_675_16

CTCGGGAGGCTGAGGCAGGGGGGTT
+FC12044_91407_8_200_675_16
OXTXXXSXXQXXOXXKMXXMXOKQF
@FC12044_91407_8_200_285_136
+FC12044_91407_8_200_285_136`,
			verbatim: false,
			ids:      expectedIds,
			seqs: []alphabet.QLetters{
				expectedQLetters[0],
				expectedQLetters[1],
				expectedQLetters[2],
				expectedQLetters[3],
				nil,
			},
		},
	}
)

func (s *S) TestReadFastq(c *check.C) {
	for _, t := range fqTests {
		r := NewReader(bytes.NewBufferString(t.fq), linear.NewQSeq("", nil, alphabet.DNA, alphabet.Sanger))
		var n int
		for n = 0; ; n++ {
			if s, err := r.Read(); err != nil {
				if err == io.EOF {
					break
				} else {
					c.Fatalf("Failed to read %s in %q: %s", expectedIds[n], t.fq, err)
				}
			} else {
				l := s.(*linear.QSeq)
				header := l.Name()
				if desc := l.Description(); len(desc) > 0 {
					header += " " + desc
				}
				c.Check(header, check.Equals, t.ids[n])
				c.Check(l.Slice(), check.DeepEquals, t.seqs[n])
			}
		}
		c.Check(n, check.Equals, len(t.ids))
	}
}

func (s *S) TestWriteFastq(c *check.C) {
	for i, t := range fqTests {
		if !t.verbatim {
			continue
		}
		for j := 0; j < 2; j++ {
			var n int
			b := &bytes.Buffer{}
			w := NewWriter(b)
			w.QID = j == 0
			seq := linear.NewQSeq("", nil, alphabet.DNA, alphabet.Sanger)

			for i := range expectedIds {
				seq.ID = t.ids[i]
				seq.Seq = t.seqs[i]
				_n, err := w.Write(seq)
				c.Assert(err, check.Equals, nil, check.Commentf("Failed to write to buffer: %s", err))
				n += _n
			}

			c.Check(n, check.Equals, b.Len())

			if w.QID {
				c.Check(string(b.Bytes()), check.Equals, t.fq, check.Commentf("Write test %d", i))
			}
		}
	}
}
