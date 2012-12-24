// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
