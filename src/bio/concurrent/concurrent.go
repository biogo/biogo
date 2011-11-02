// Package to apply a function over an array or stream of data.
//
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
package concurrent

// The Concurrent interface represents a processor that allows adding jobs and retrieving results
type Concurrent interface {
	Process(...interface{})
	Result() (interface{}, error)
}

// Function type that operates on data
type Eval func(...interface{}) (interface{}, error)

type Result struct {
	Value interface{}
	Error error
}
