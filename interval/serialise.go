package interval
// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
// Derived from quicksect.py of bx-python ©James Taylor bitbucket.org/james_taylor/bx-python
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

import (
	"bytes"
	"encoding/gob"
	"github.com/kortschak/BioGo/bio"
)

const Version int = 1

func (self *Interval) GobEncode() (b []byte, err error) {
	var branch []byte

	buffer := bytes.NewBuffer(b)
	encoder := gob.NewEncoder(buffer)

	encoder.Encode(Version)

	/*
		for i := 0; i < reflect.TypeOf(self).Elem().NumField(); i++ {
			switch reflect.TypeOf(self).Elem().Field(i).Name {
			case "left", "right":
			default:
				binary.Write(buffer, ByteOrder, reflect.ValueOf(self).Elem().Field(i).Interface())
			}
		}
	*/

	encoder.Encode(self.chromosome)
	encoder.Encode(self.start)
	encoder.Encode(self.end)
	encoder.Encode(self.line)
	encoder.Encode(self.priority)
	encoder.Encode(self.maxEnd)
	encoder.Encode(self.Meta)

	if self.left != nil {
		branch, _ = self.left.GobEncode()
		err = encoder.Encode(branch)
	}

	if self.right != nil {
		branch, _ = self.right.GobEncode()
		err = encoder.Encode(branch)
	}

	b = buffer.Bytes()

	return
}
func (self *Interval) GobDecode(b []byte) (err error) {
	var (
		version int
		branch  []byte
	)

	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)

	if err = decoder.Decode(&version); err == nil {
		if version != Version {
			return bio.NewError("Encoding mismatch", 0, []int{version, Version})
		}
	} else {
		return
	}

	/*		
		for i := 0; i < reflect.TypeOf(self).Elem().NumField(); i++ {
			switch reflect.TypeOf(self).Elem().Field(i).Name {
			case "left", "right", "Meta":
			default:
				v := reflect.ValueOf(reflect.TypeOf(self).Elem().Field(i))
				binary.Read(buffer, ByteOrder, v)
				reflect.ValueOf(self).Elem().Field(i).Set(v)
			}
		}
	*/

	decoder.Decode(self.chromosome)
	decoder.Decode(self.start)
	decoder.Decode(self.end)
	decoder.Decode(self.line)
	decoder.Decode(self.priority)
	decoder.Decode(self.maxEnd)
	decoder.Decode(self.Meta)

	decoder.Decode(branch)
	self.left = &Interval{}
	self.left.GobDecode(branch)

	decoder.Decode(branch)
	self.right = &Interval{}
	self.right.GobDecode(branch)

	return
}
