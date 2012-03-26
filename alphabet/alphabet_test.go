package alphabet

// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

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
		comp  Complementable
	)

	for _, a := range []interface{}{DNA, RNA, Protein} {
		c.Check(a, check.Implements, &alpha)
	}

	for _, a := range []interface{}{DNA, RNA} {
		c.Check(a, check.Implements, &comp)
	}

	c.Check(Protein, check.Not(check.Implements), &comp)
}

type testAlphabets struct {
	letters  string
	alphabet Alphabet
}

func (s *S) TestIsValid(c *check.C) {
	for _, t := range []testAlphabets{{N, DNA}, {R, RNA}, {P, Protein}} {
		for i := 0; i < 256; i++ {
			c.Check(t.alphabet.IsValid(byte(i)), check.Equals, strings.ContainsRune(t.letters, unicode.ToUpper(rune(i))) || strings.ContainsRune(t.letters, unicode.ToLower(rune(i))))
		}
	}
}

func (s *S) TestLetter(c *check.C) {
	for _, t := range []testAlphabets{{N, DNA}, {R, RNA}, {P, Protein}} {
		for i := 0; i < t.alphabet.Len(); i++ {
			c.Check(t.alphabet.IndexOf(t.alphabet.Letter(i)), check.Equals, i)
		}
	}
}

func (s *S) TestComplementOf(c *check.C) {
	for _, t := range []testAlphabets{{N, DNA}, {R, RNA}} {
		for i := 0; i < 256; i++ {
			if sc, ok := t.alphabet.(Complementable).ComplementOf(byte(i)); ok {
				dc, ok := t.alphabet.(Complementable).ComplementOf(sc)
				c.Check(ok, check.Equals, true)
				c.Check(dc, check.Equals, byte(i))
			}
		}
	}
}

func (s *S) TestComplementDirect(c *check.C) {
	for _, t := range []testAlphabets{{N, DNA}, {R, RNA}} {
		complement := t.alphabet.(Complementable).ComplementTable()
		for i := 0; i < 256; i++ {
			if sc := complement[i]; sc <= unicode.MaxASCII {
				dc := complement[sc]
				c.Check(dc <= unicode.MaxASCII, check.Equals, true)
				c.Check(dc, check.Equals, byte(i))
			} else {
				c.Check(sc&unicode.MaxASCII, check.Equals, byte(i&unicode.MaxASCII))
			}
		}
	}
}

func (s *S) TestString(c *check.C) {
	e := [...]string{"acgtACGT", "acguACGU", "*-abcdefghijklmnpqrstvxyz*-ABCDEFGHIJKLMNPQRSTVXYZ"}
	for i, t := range []testAlphabets{{N, DNA}, {R, RNA}, {P, Protein}} {
		c.Check(t.alphabet.String(), check.Equals, e[i])
	}
}

func (s *S) TestRangeCheck(c *check.C) {
	var err error
	_, err = NewGeneric(string([]rune{256}), 0, 0, 0, !CaseSensitive)
	c.Check(err, check.Not(check.IsNil))
	_, err = NewGeneric(string([]rune{0}), 0, 0, 0, !CaseSensitive)
	c.Check(err, check.IsNil)
	_, err = NewGeneric(string([]rune{127}), 0, 0, 0, !CaseSensitive)
	c.Check(err, check.IsNil)
	_, err = NewGeneric(string([]rune{-1}), 0, 0, 0, !CaseSensitive)
	c.Check(err, check.Not(check.IsNil))
}

func BenchmarkIsValidGeneric(b *testing.B) {
	b.StopTimer()
	g, _ := NewGeneric(P, 0, 0, 0, !CaseSensitive)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		g.IsValid(byte(i))
	}
}

func BenchmarkIsValidProtein(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Protein.IsValid(byte(i))
	}
}

func BenchmarkIsValidDNA(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DNA.IsValid(byte(i))
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
		DNA.IndexOf(byte(i))
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
		DNA.ComplementOf(byte(i))
	}
}

func BenchmarkComplementDNADirect(b *testing.B) {
	complement := DNA.ComplementTable()
	var c byte
	for i := 0; i < b.N; i++ {
		if c = complement[byte(i)]; c != 0x80 {
		}
	}
}
