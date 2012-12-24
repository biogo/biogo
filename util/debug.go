// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"os"
)

type Debug bool

func (d Debug) Println(args ...interface{}) {
	if d {
		caller := GetCaller(1)
		fmt.Fprintf(os.Stderr, "%s %s#%d:", caller.Package, caller.File, caller.Line)
		fmt.Fprintln(os.Stderr, args...)
	}
}

func (d Debug) Printf(format string, args ...interface{}) {
	if d {
		caller := GetCaller(1)
		fmt.Fprintf(os.Stderr, "%s %s#%d:", caller.Package, caller.File, caller.Line)
		fmt.Fprintf(os.Stderr, format, args...)
	}
}
