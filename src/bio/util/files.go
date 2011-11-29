package util
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

import (
	"crypto/md5"
	"github.com/kortschak/BioGo/bio"
	"io"
	"os"
)

const (
	bufferLen = 1 << 15
)

var buffer = make([]byte, bufferLen)

type ReadSeekStater interface {
	io.ReadSeeker
	Stat() (*os.FileInfo, error)
}

func Hash(file ReadSeekStater) (sum []byte, err error) {
	var fi *os.FileInfo
	if fi, err = file.Stat(); err != nil || fi.IsDirectory() {
		return nil, bio.NewError("Is a directory", 0, file)
	}

	file.Seek(0, 0)

	var n int
	h := md5.New()

	for err == nil || err == io.ErrUnexpectedEOF {
		n, err = io.ReadAtLeast(file, buffer, bufferLen)
		h.Write(buffer[:n])
	}

	if err == io.EOF || err == io.ErrUnexpectedEOF {
		err = nil
	}

	file.Seek(0, 0)
	sum = h.Sum()

	return
}
