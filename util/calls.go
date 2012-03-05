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
