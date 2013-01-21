// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package feat

import (
	"strings"
)

const (
	Undefined Moltype = iota - 1
	DNA
	RNA
	Protein
)

var (
	moltypeToString = [...]string{
		"DNA", "RNA", "Protein",
	}
	stringToMoltype map[string]Moltype = map[string]Moltype{}
)

func init() {
	for m, s := range moltypeToString {
		stringToMoltype[strings.ToLower(s)] = Moltype(m)
		stringToMoltype[s] = Moltype(m)
	}
}

// Moltype represents the molecule type of a source of sequence data.
type Moltype int8

// Return a string representation of a Moltype.
func (m Moltype) String() string {
	if m == Undefined {
		return "Undefined"
	}
	return moltypeToString[m]
}

// ParseMoltype allows conversion from a string to a Moltype.
func ParseMoltype(s string) Moltype {
	if m, ok := stringToMoltype[s]; ok {
		return m
	}

	return Undefined
}
