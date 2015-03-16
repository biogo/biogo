// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package genome defines types useful for representing cytogenetic features.
package genome

import (
	"github.com/biogo/biogo/feat"
	"github.com/biogo/biogo/seq"
)

// Chromosome represent a chromosome and associated features. Elements in the
// Features field should return the Chromosome when their Location() method is
// called.
type Chromosome struct {
	Chr      string
	Desc     string
	Length   int
	Features []feat.Feature
}

func (c *Chromosome) Start() int             { return 0 }
func (c *Chromosome) End() int               { return c.Length }
func (c *Chromosome) Len() int               { return c.Length }
func (c *Chromosome) Name() string           { return c.Chr }
func (c *Chromosome) Description() string    { return c.Desc }
func (c *Chromosome) Location() feat.Feature { return nil }

// Band represents a chromosome band.
type Band struct {
	Band     string
	Desc     string
	Chr      feat.Feature
	StartPos int
	EndPos   int
	Giemsa   string
}

func (b *Band) Start() int             { return b.StartPos }
func (b *Band) End() int               { return b.EndPos }
func (b *Band) Len() int               { return b.End() - b.Start() }
func (b *Band) Name() string           { return b.Band }
func (b *Band) Description() string    { return b.Desc }
func (b *Band) Location() feat.Feature { return b.Chr }

// Fragment represents an assembly fragment.
type Fragment struct {
	Frag      string
	Desc      string
	Chr       feat.Feature
	ChrStart  int
	ChrEnd    int
	FragStart int
	FragEnd   int
	Type      byte
	Strand    seq.Strand
}

func (f *Fragment) Start() int             { return f.ChrStart }
func (f *Fragment) End() int               { return f.ChrEnd }
func (f *Fragment) Len() int               { return f.End() - f.Start() }
func (f *Fragment) Name() string           { return f.Frag }
func (f *Fragment) Description() string    { return f.Desc }
func (f *Fragment) Location() feat.Feature { return f.Chr }
