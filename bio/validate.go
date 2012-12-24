// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bio

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
func (self *Validator) String() string {
	valid := make([]byte, 0, 256)
	for i, v := range self.valid {
		if v {
			valid = append(valid, byte(i))
		}
	}

	return string(valid)
}
