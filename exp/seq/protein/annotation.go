// Copyright ©2011-2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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

package protein

import (
	"code.google.com/p/biogo/bio"
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/feat"
)

// An Annotation is a basic protein sequence annotation type.
type Annotation struct {
	ID      string
	Desc    string
	Loc     feat.Feature
	Conform feat.Conformation
	Alpha   alphabet.Peptide
	Offset  int
}

// Protein is a no-op method required to satisfy protein.Sequence interface.
func (a *Annotation) Protein() {}

// Name returns the ID string of the sequence.
func (a *Annotation) Name() string { return a.ID }

// SetName sets the ID string of the sequence.
func (a *Annotation) SetName(id string) { a.ID = id }

// Description returns the Desc string of the sequence.
func (a *Annotation) Description() string { return a.Desc }

// SetDescription sets the Desc string of the sequence.
func (a *Annotation) SetDescription(d string) { a.Desc = d }

// Location returns the Loc field of the sequence.
func (a *Annotation) Location() feat.Feature { return a.Loc }

// SetLocation sets the Loc field of the sequence.
func (a *Annotation) SetLocation(f feat.Feature) { a.Loc = f }

// Alphabet return the alphabet.Alphabet used by the sequence.
func (a *Annotation) Alphabet() alphabet.Alphabet { return a.Alpha }

// SetAlphabet the sets the alphabet.Protein used by the sequence.
func (a *Annotation) SetAlphabet(n alphabet.Peptide) { a.Alpha = n }

// SetOffset sets the global offset of the sequence to o.
func (a *Annotation) SetOffset(o int) { a.Offset = o }

// Moltype returns the molecule type of the sequence.
func (a *Annotation) Moltype() bio.Moltype { return a.Alpha.Moltype() }
