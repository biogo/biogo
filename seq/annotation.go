// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package seq

import (
	"code.google.com/p/biogo/alphabet"
	"code.google.com/p/biogo/feat"
)

// Strand stores linear sequence strand information.
type Strand int8

const (
	Minus Strand = iota - 1
	None
	Plus
)

func (s Strand) String() string {
	switch s {
	case Plus:
		return "+"
	case None:
		return "."
	case Minus:
		return "-"
	}
	return "undefined"
}

// An Annotation is a basic linear sequence annotation type.
type Annotation struct {
	ID      string
	Desc    string
	Loc     feat.Feature
	Strand  Strand
	Conform feat.Conformation
	Alpha   alphabet.Alphabet
	Offset  int
}

// Name returns the ID string of the sequence.
func (a *Annotation) Name() string { return a.ID }

// SetName sets the ID string of the sequence.
func (a *Annotation) SetName(id string) { a.ID = id }

// Description returns the Desc string of the sequence.
func (a *Annotation) Description() string { return a.Desc }

// SetDescription sets the Desc string of the sequence.
func (a *Annotation) SetDescription(d string) { a.Desc = d }

// Conformation returns the sequence conformation.
func (a *Annotation) Conformation() feat.Conformation { return a.Conform }

// SetConformation sets the sequence conformation.
func (a *Annotation) SetConformation(c feat.Conformation) { a.Conform = c }

// Orientation returns the sequence'a strand as a feat.Orientation.
func (a *Annotation) Orientation() feat.Orientation { return feat.Orientation(a.Strand) }

// SetOrientation sets the sequence'a strand from a feat.Orientation.
func (a *Annotation) SetOrientation(o feat.Orientation) { a.Strand = Strand(o) }

// Location returns the Loc field of the sequence.
func (a *Annotation) Location() feat.Feature { return a.Loc }

// SetLocation sets the Loc field of the sequence.
func (a *Annotation) SetLocation(f feat.Feature) { a.Loc = f }

// Alphabet return the alphabet.Alphabet used by the sequence.
func (a *Annotation) Alphabet() alphabet.Alphabet { return a.Alpha }

// SetAlphabet the sets the alphabet.Alphabet used by the sequence.
func (a *Annotation) SetAlphabet(n alphabet.Alphabet) { a.Alpha = n }

// SetOffset sets the global offset of the sequence to o.
func (a *Annotation) SetOffset(o int) { a.Offset = o }

// Moltype returns the molecule type of the sequence.
func (a *Annotation) Moltype() feat.Moltype { return a.Alpha.Moltype() }

// CloneAnnotation returns a pointer to a copy of the receiver.
func (a *Annotation) CloneAnnotation() *Annotation { ca := *a; return &ca }
