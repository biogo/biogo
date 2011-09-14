package bio
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

import ()

// Provide default Validators.
var (
	N = []byte("ACGTacgt")
	R = []byte("ACGUacgu")
	P = []byte("ABCDEFGHIJKLMNPQRSTVXYZabcdefghijklmnpqrstvxyz*")
)

var ValidN, ValidR, ValidP *Validator

func init() {
	ValidN = NewValidator(N)
	ValidR = NewValidator(R)
	ValidP = NewValidator(P)
}

// Validator type checks that a sequence conforms to a specified alphabet.
type Validator struct {
	valid [256]bool
}

// Make a new Validator with valid defining the allowable values for the alphabet.
func NewValidator(valid []byte) (v *Validator) {
	v = &Validator{}
	for _, i := range valid {
		v.valid[i] = true
	}

	return
}

// Check that a slice of bytes conforms to an alphabet, returning false
// and the position of the first invalid byte if invalid and true and a negative
// int if valid.
func (self *Validator) Check(n []byte) (valid bool, pos int) {
	for i, v := range n {
		if !self.valid[v] {
			return false, i
		}
	}

	return true, -1
}

// Return a string indicating characters accepted as valid by the Validator.
func (self *Validator) Valid() string {
	valid := make([]byte, 0, 256)
	for i, v := range self.valid {
		if v {
			valid = append(valid, byte(i))
		}
	}

	return string(valid)
}
