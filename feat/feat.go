// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Basic Feature package
package feat

import (
	"code.google.com/p/biogo/bio"
	"strconv"
)

var (
	moltypesToString        = [...]string{"dna", "rna", "protein"}
	defaultID        string = "nil"
)

// Feature type
type Feature struct {
	ID          string
	Source      string
	Location    string
	Start       int
	End         int
	Feature     string
	Score       *float64
	Probability *float64
	Attributes  string
	Comments    string
	Frame       int8
	Strand      int8
	Moltype     bio.Moltype
	Meta        interface{}
}

// Return a new Feature
func New(ID string) *Feature {
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
