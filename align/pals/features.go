// Copyright ©2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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

package pals

import (
	"code.google.com/p/biogo.graph"
	"code.google.com/p/biogo/exp/feat"
	"fmt"
)

// A Contig is a base feature type. Other features in the pals package point to this type in
// their location fields.
type Contig string

// Name returns the value of the receiver as a string.
func (c Contig) Name() string { return string(c) }

// Description returns the string "contig".
func (c Contig) Description() string { return "contig" }

// Start returns the value 0.
func (c Contig) Start() int { return 0 }

// End returns the value 0.
func (c Contig) End() int { return 0 }

// Len returns the value 0.
func (c Contig) Len() int { return 0 }

// Location returns a nil feat.Feature.
func (c Contig) Location() feat.Feature { return nil }

// String returns the value of the receiver as a string.
func (c Contig) String() string { return string(c) }

// A Feature is a description of a pals feature interval.
type Feature struct {
	ID   string
	From int
	To   int
	Loc  feat.Feature
}

func (f *Feature) Name() string { return f.ID }

// Description returns the string "pals feature".
func (f *Feature) Description() string    { return "pals feature" }
func (f *Feature) Start() int             { return f.From }
func (f *Feature) End() int               { return f.To }
func (f *Feature) Len() int               { return f.To - f.From }
func (f *Feature) Location() feat.Feature { return f.Loc }

func (f *Feature) String() string {
	return fmt.Sprintf("%s[%d,%d)", f.Loc.Name(), f.From, f.To)
}

// A Pile is a collection of features covering a maximal (potentially contiguous, depending on
// the value of overlap used for creation of the Piler) region of copy count > 0.
type Pile struct {
	From   int
	To     int
	Strand int8
	Loc    feat.Feature
	Images []*Pair
	graph.Node
}

func (p *Pile) Name() string {
	return fmt.Sprintf("%s[%d,%d)", p.Loc.Name(), p.From, p.To)
}

// Description returns the string "pile".
func (p *Pile) Description() string    { return "pile" }
func (p *Pile) Start() int             { return p.From }
func (p *Pile) End() int               { return p.To }
func (p *Pile) Len() int               { return p.To - p.From }
func (p *Pile) Location() feat.Feature { return p.Loc }

func (p *Pile) String() string {
	return fmt.Sprintf("{%s[%d,%d): %v}", p.Loc.Name(), p.From, p.To, p.Images)
}
