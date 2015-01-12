// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kmerindex

import (
	"code.google.com/p/biogo/alphabet"
	"code.google.com/p/biogo/seq/linear"
	"code.google.com/p/biogo/util"

	"gopkg.in/check.v1"
	"math/rand"
	"strings"
	"testing"
)

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct {
	*linear.Seq
}

var _ = check.Suite(&S{})

var testLen = 1000

func (s *S) SetUpSuite(c *check.C) {
	MaxKmerLen = 14
	s.Seq = linear.NewSeq("", nil, alphabet.DNA)
	s.Seq.Seq = make(alphabet.Letters, testLen)
	for i := range s.Seq.Seq {
		s.Seq.Seq[i] = [...]alphabet.Letter{'A', 'C', 'G', 'T', 'a', 'c', 'g', 't'}[rand.Int()%8]
	}
}

func (s *S) TestKmerIndexCheck(c *check.C) {
	for k := MinKmerLen; k <= MaxKmerLen; k++ {
		if i, err := New(k, s.Seq); err != nil {
			c.Fatalf("New KmerIndex failed: %v", err)
		} else {
			ok, _ := i.Check()
			c.Check(ok, check.Equals, false)
			i.Build()
			ok, f := i.Check()
			c.Check(f, check.Equals, s.Seq.Len()-k+1)
			c.Check(ok, check.Equals, true)
		}
	}
}

func (s *S) TestKmerFrequencies(c *check.C) {
	for k := MinKmerLen; k <= MaxKmerLen; k++ {
		if i, err := New(k, s.Seq); err != nil {
			c.Fatalf("New KmerIndex failed: %v", err)
		} else {
			freqs, ok := i.KmerFrequencies()
			c.Check(ok, check.Equals, true)
			hashFreqs := make(map[string]int)
			for i := 0; i+k <= s.Seq.Len(); i++ {
				hashFreqs[strings.ToLower(string(alphabet.LettersToBytes(s.Seq.Seq[i:i+k])))]++
			}
			for key := range freqs {
				c.Check(freqs[key], check.Equals, hashFreqs[i.Format(key)],
					check.Commentf("key %x, string of %q\n", key, i.Format(key)))
			}
			for key := range hashFreqs {
				if keyKmer, err := i.KmerOf(key); err != nil {
					c.Fatal(err)
				} else {
					c.Check(freqs[keyKmer], check.Equals, hashFreqs[key],
						check.Commentf("keyKmer %x, string of %q, key %q\n", keyKmer, i.Format(keyKmer), key))
				}
			}
		}
	}
}

func (s *S) TestKmerPositions(c *check.C) {
	for k := MinKmerLen; k < MaxKmerLen; k++ { // don't test full range to time's sake
		if i, err := New(k, s.Seq); err != nil {
			c.Fatalf("New KmerIndex failed: %v", err)
		} else {
			i.Build()
			hashPos := make(map[string][]int)
			for i := 0; i+k <= s.Seq.Len(); i++ {
				p := strings.ToLower(string(alphabet.LettersToBytes(s.Seq.Seq[i : i+k])))
				hashPos[p] = append(hashPos[p], i)
			}
			pos, ok := i.KmerIndex()
			c.Check(ok, check.Equals, true)
			for p := range pos {
				c.Check(pos[p], check.DeepEquals, hashPos[i.Format(p)])
			}
		}
	}
}

func (s *S) TestKmerPositionsString(c *check.C) {
	for k := MinKmerLen; k < MaxKmerLen; k++ { // don't test full range to time's sake
		if i, err := New(k, s.Seq); err != nil {
			c.Fatalf("New KmerIndex failed: %v", err)
		} else {
			i.Build()
			hashPos := make(map[string][]int)
			for i := 0; i+k <= s.Seq.Len(); i++ {
				p := strings.ToLower(string(alphabet.LettersToBytes(s.Seq.Seq[i : i+k])))
				hashPos[p] = append(hashPos[p], i)
			}
			pos, ok := i.StringKmerIndex()
			c.Check(ok, check.Equals, true)
			for p := range pos {
				c.Check(pos[p], check.DeepEquals, hashPos[p])
			}
		}
	}
}

func (s *S) TestKmerKmerUtilities(c *check.C) {
	for k := MinKmerLen; k <= 8; k++ { // again not testing all exhaustively
		for kmer := Kmer(0); uint(kmer) <= util.Pow4(k)-1; kmer++ {
			// Interconversion between string and Kmer
			s, err := Format(kmer, k, alphabet.DNA)
			c.Assert(err, check.Equals, nil)
			rk, err := KmerOf(k, alphabet.DNA.LetterIndex(), s)
			c.Assert(err, check.Equals, nil)
			c.Check(rk, check.Equals, kmer)

			// Complementation
			dc := ComplementOf(k, ComplementOf(k, kmer))
			skmer, _ := Format(kmer, k, alphabet.DNA)
			sdc, _ := Format(dc, k, alphabet.DNA)
			c.Check(dc, check.Equals, kmer, check.Commentf("kmer: %s\ndouble complement: %s\n", skmer, sdc))

			// GC content
			ks, _ := Format(kmer, k, alphabet.DNA)
			gc := 0
			for _, b := range ks {
				if b == 'g' || b == 'c' {
					gc++
				}
			}
			c.Check(GCof(k, kmer), check.Equals, float64(gc)/float64(k))
		}
	}
}
