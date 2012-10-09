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

package util

import (
	"code.google.com/p/biogo/bio"
	"hash"
	"io"
	"os"
)

const (
	bufferLen = 1 << 15
)

var buffer = make([]byte, bufferLen)

// Hash returns the h hash sum of file f and any error. The hash is not reset on return,
// so if individual files are to be hashed with the same h, it should be reset.
func Hash(h hash.Hash, f *os.File) (sum []byte, err error) {
	fi, err := f.Stat()
	if err != nil || fi.IsDir() {
		return nil, bio.NewError("Is a directory", 0, f)
	}

	s := io.NewSectionReader(f, 0, fi.Size())

	for n, buffer := 0, make([]byte, bufferLen); err == nil; {
		n, err = s.Read(buffer)
		h.Write(buffer[:n])
	}
	if err == io.EOF {
		err = nil
	}

	sum = h.Sum(nil)

	return
}
