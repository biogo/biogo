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

package bio

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
		stringToMoltype[s] = Moltype(m)
	}
}

// Moltype represents the molecule type of a source of sequence data.
type Moltype int8

// Return a string representation of a Moltype.
func (self Moltype) String() string {
	if self == Undefined {
		return "Undefined"
	}
	return moltypeToString[self]
}

// ParseMoltype allows conversion from a string to a Moltype.
func ParseMoltype(s string) Moltype {
	if m, ok := stringToMoltype[s]; ok {
		return m
	}

	return Undefined
}
