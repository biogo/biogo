// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	check "launchpad.net/gocheck"
	"os"
	"os/exec"
)

// Tests
func (s *S) TestHash(c *check.C) {
	// FIXME: This will not work with MacOS.
	_, err := exec.LookPath("md5sum")
	if err != nil {
		c.Skip(err.Error())
	}
	md5sum := exec.Command("md5sum", "./files_test.go")
	b := &bytes.Buffer{}
	md5sum.Stdout = b
	err = md5sum.Run()
	if err != nil {
		c.Fatal(err)
	}
	f, err := os.Open("./files_test.go")
	if err != nil {
		c.Fatalf("%v %s", md5sum, err)
	}
	x, err := ioutil.ReadAll(f)
	if err != nil {
		c.Fatal(err)
	}
	f.Seek(0, 0)

	md5hash, err := Hash(md5.New(), f)
	if err != nil {
		c.Fatal(err)
	}
	md5string := fmt.Sprintf("%x .*\n", md5hash)

	c.Check(string(b.Bytes()), check.Matches, md5string)

	y, err := ioutil.ReadAll(f)
	if err != nil {
		c.Fatal(err)
	}
	c.Check(x, check.DeepEquals, y)
}
