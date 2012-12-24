// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"runtime"
	"strings"
)

type Caller struct {
	Package  string
	Function string
	File     string
	Line     int
}

func GetCaller(skip int) *Caller {
	if pc, _, _, ok := runtime.Caller(skip + 1); ok {
		function := runtime.FuncForPC(pc)
		caller := strings.Split(function.Name(), ".")
		file, line := function.FileLine(pc)
		return &Caller{
			Package:  strings.Join(caller[0:len(caller)-1], "."),
			Function: caller[len(caller)-1],
			File:     file,
			Line:     line,
		}
	}
	return nil
}
