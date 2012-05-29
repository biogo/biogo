// Copyright ©2011-2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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
