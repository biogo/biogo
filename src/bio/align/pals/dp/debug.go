package dp
// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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
//
import (
	"fmt"
	"path"
	"runtime"
)

type debugging bool

func (d debugging) Println(args ...interface{}) {
	if d {
		fmt.Println(args...)
	}
}

func (d debugging) Printf(format string, args ...interface{}) {
	if d {
		fmt.Printf(format, args...)
	}
}

func (d debugging) PrintlnL(args ...interface{}) {
	if d {
		_, file, line, _ := runtime.Caller(1)
		_, file = path.Split(file)
		fmt.Printf("%s:%d ", file, line)
		fmt.Println(args...)
	}
}

func (d debugging) PrintfL(format string, args ...interface{}) {
	if d {
		_, file, line, _ := runtime.Caller(1)
		_, file = path.Split(file)
		fmt.Printf("%s:%d ", file, line)
		fmt.Printf(format, args...)
	}
}
