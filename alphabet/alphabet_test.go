// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package alphabet

import (
	check "launchpad.net/gocheck"
	"strings"
	"testing"
	"unicode"
)

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

func (s *S) TestInterfaces(c *check.C) {
	var (
		alpha Alphabet
		comp  Complementor
	)

	for _, a := range []interface{}{DNA, RNA, Protein} {
		c.Check(a, check.Implements, &alpha)
	}

	for _, a := range []interface{}{DNA, DNAredundant, RNA, RNAredundant} {
		c.Check(a, check.Implements, &comp)
	}

	c.Check(Protein, check.Not(check.Implements), &comp)
}

type testAlphabets struct {
	alphabet Alphabet
	letters  string
}

func (s *S) TestIsValid(c *check.C) {
	for _, t := range []testAlphabets{
		{DNA, "acgt"},
		{RNA, "acgu"},
		{Protein, "abcdefghijklmnpqrstvwxyz*"},
	} {
		for i := 0; i < 256; i++ {
			c.Check(t.alphabet.IsValid(Letter(i)), check.Equals, strings.ContainsRune(t.letters, unicode.ToUpper(rune(i))) || strings.ContainsRune(t.letters, unicode.ToLower(rune(i))))
		}
	}
}

func uc(l Letter) Letter {
	return Letter(unicode.ToUpper(rune(l)))
}

func (s *S) TestLetter(c *check.C) {
	for _, t := range []Alphabet{
		DNA,
		RNA,
		Protein,
	} {
		for i := 0; i < t.Len(); i++ {
			c.Check(t.IndexOf(t.Letter(i)), check.Equals, i,
				check.Commentf("Index %d: %c == %d", i, t.Letter(i), t.IndexOf(t.Letter(i))))
			c.Check(t.IndexOf(uc(t.Letter(i))), check.Equals, i,
				check.Commentf("Index %d: %c == %d", i, uc(t.Letter(i)), t.IndexOf(uc(t.Letter(i)))))
		}
	}
}

func (s *S) TestComplement(c *check.C) {
	for _, t := range []Complementor{
		DNA,
		RNA,
	} {
		for i := 0; i < 256; i++ {
			if sc, ok := t.Complement(Letter(i)); ok {
				dc, ok := t.Complement(sc)
				c.Check(ok, check.Equals, true)
				c.Check(dc, check.Equals, Letter(i))
			}
		}
	}
}

func (s *S) TestComplementDirect(c *check.C) {
	for _, t := range []Complementor{
		DNA,
		RNA,
	} {
		complement := t.ComplementTable()
		for i := 0; i < 256; i++ {
			if sc := complement[i]; sc <= unicode.MaxASCII {
				dc := complement[sc]
				c.Check(dc <= unicode.MaxASCII, check.Equals, true)
				c.Check(dc, check.Equals, Letter(i))
			} else {
				c.Check(sc&unicode.MaxASCII, check.Equals, Letter(i&unicode.MaxASCII))
			}
		}
	}
}

func (s *S) TestLetters(c *check.C) {
	for _, t := range []testAlphabets{
		{DNA, "acgtACGT"},
		{RNA, "acguACGU"},
		{Protein, "abcdefghijklmnpqrstvwxyz*ABCDEFGHIJKLMNPQRSTVWXYZ*"},
	} {
		c.Check(t.alphabet.Letters(), check.Equals, t.letters)
	}
}

func (s *S) TestRangeCheck(c *check.C) {
	var err error
	_, err = newAlphabet(string([]rune{256}), 0, 0, 0, !CaseSensitive)
	c.Check(err, check.Not(check.IsNil))
	_, err = newAlphabet(string([]rune{0}), 0, 0, 0, !CaseSensitive)
	c.Check(err, check.IsNil)
	_, err = newAlphabet(string([]rune{127}), 0, 0, 0, !CaseSensitive)
	c.Check(err, check.IsNil)
	_, err = newAlphabet(string([]rune{-1}), 0, 0, 0, !CaseSensitive)
	c.Check(err, check.Not(check.IsNil))
}

func BenchmarkIsValid(b *testing.B) {
	b.StopTimer()
	g, _ := newAlphabet("abcdefghijklmnpqrstvwxyz*", 0, 0, 0, !CaseSensitive)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		g.IsValid(Letter(i))
	}
}

func BenchmarkIsValidProtein(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Protein.IsValid(Letter(i))
	}
}

func BenchmarkIsValidDNA(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DNA.IsValid(Letter(i))
	}
}

func BenchmarkIsValidDNADirect(b *testing.B) {
	valid := DNA.ValidLetters()
	for i := 0; i < b.N; i++ {
		_ = valid[byte(i)]
	}
}

func BenchmarkIndexDNA(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DNA.IndexOf(Letter(i))
	}
}

func BenchmarkIndexDNADirect(b *testing.B) {
	index := DNA.LetterIndex()
	for i := 0; i < b.N; i++ {
		_ = index[byte(i)]
	}
}

func BenchmarkComplementDNA(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DNA.Complement(Letter(i))
	}
}

func BenchmarkComplementDNADirect(b *testing.B) {
	comp := DNA.ComplementTable()
	var c Letter
	for i := 0; i < b.N; i++ {
		if c = comp[Letter(i)]; c != 0x80 {
		}
	}
}
