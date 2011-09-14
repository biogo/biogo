// Basic Feature package
//
// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
//   This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU General Public License for more details.
//
//   You should have received a copy of the GNU General Public License
//   along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
package feat

import (
	"strconv"
)

var (
	moltypesToString        = [...]string{"dna", "rna", "protein"}
	defaultID        string = "nil"
)

// Feature type
type Feature struct {
	ID          []byte
	Source      []byte
	Location    []byte
	Start       int
	End         int
	Feature     []byte
	Score       float64
	Probability float64
	Attributes  []byte
	Comments    []byte
	Frame       int8
	Strand      int8
	Moltype     byte
	Meta        interface{}
}

// Return a new Feature
func New(ID []byte) *Feature {
	return &Feature{ID: ID}
}

// Return the length of a Feature
func (self *Feature) Len() int {
	return self.End - self.Start
}

// Return the molecule type of the Feature
func (self *Feature) MoltypeAsString() string {
	return moltypesToString[self.Moltype]
}

var defaultStringFunc = func(f *Feature) string {
	var id, comments string
	if string(f.ID) == "" {
		id = defaultID
	} else {
		id = string(f.ID)
	}
	if string(f.Comments) != "" {
		comments = " " + string(f.Comments)
	}
	return id + ":" +
		string(f.Location) + ":" +
		strconv.Itoa(f.Start) + ".." +
		strconv.Itoa(f.End) +
		":(" +
		string(f.Feature) + " " + string(f.Source) +
		"):" +
		string(f.Attributes) + comments
}

// Return the canonical string conversion of the Feature - for particulat formats, use bio/io/featio/ packages.
var StringFunc = defaultStringFunc

func (self *Feature) String() string {
	return StringFunc(self)
}
