package bio
// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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

const (
	Undefined Moltype = iota - 1
	DNA
	RNA
	Protein
)

var (
	Precision        = 4
	FloatFormat byte = 'f'
)

type Moltype int8

var moltypesToString = [...]string{
	"DNA", "RNA", "Protein",
}

func (self Moltype) String() string {
	if self < 0 {
		return "Undefined"
	}
	return moltypesToString[self]
}

var ParseMoltype map[string]Moltype = map[string]Moltype{"DNA": DNA, "RNA": RNA, "Protein": Protein}

// Convert from 1-based to 0-based indexing
func OneToZero(pos int) int {
	if pos == 0 {
		panic("1-based index == 0")
	}
	if pos > 0 {
		pos--
	}

	return pos
}

// Convert from 0-based to 1-based indexing
func ZeroToOne(pos int) int {
	if pos >= 0 {
		pos++
	}

	return pos
}
