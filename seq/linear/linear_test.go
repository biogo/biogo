// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linear

import (
	"testing"

	"github.com/biogo/biogo/alphabet"
	"gopkg.in/check.v1"
)

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

func (s *S) TestWarning(c *check.C) { c.Log("\nFIXME: Tests only in example tests.\n") }

func BenchmarkRevComp(b *testing.B) {
	in := []alphabet.Letter("ATGCtGACTTGGTGCACGTACGTACGTACGTACGTACGTACGTACGTACGTACGTATGCtGACTTGGTGCACGTACGTACGTACGTACGTACGTACGTACGTACGTACGTATGCtGACTTGGTGCACGTACGTACGTACGTACGTACGTACGTACGTACGTACGTATGCtGACTTGGTGCACGTACGTACGTACGTACGTACGTACGTACGTACGTACGTATGCtGACTTGGTGCACGTACGTACGTACGTACGTACGTACGTACGTACGTACGT")
	for i := 0; i < b.N; i++ {
		s := NewSeq("example DNA", in, alphabet.DNA)
		s.RevComp()
	}
}
