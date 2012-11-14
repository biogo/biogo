// Copyright ©2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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

// Package feat provides the base for storage and manipulation of biological interval information.
package feat

type Conformationer interface {
	Conformation() Conformation
	SetConformation(Conformation)
}

type Conformation int8

func (c Conformation) String() string {
	switch c {
	case UndefinedConformation:
		return "undefined"
	case Linear:
		return "linear"
	case Circular:
		return "circular"
	}
	panic("feat: illegal conformation")
}

const (
	UndefinedConformation Conformation = iota - 1
	Linear
	Circular
)

type Orienter interface {
	Orientation() Orientation
}

type Orientation int8

func (o Orientation) String() string {
	switch o {
	case Reverse:
		return "reverse"
	case NotOriented:
		return "not oriented"
	case Forward:
		return "forward"
	}
	panic("feat: illegal orientation")
}

const (
	Reverse Orientation = iota - 1
	NotOriented
	Forward
)

type Range interface {
	Start() int
	End() int
	Len() int
}

type Feature interface {
	Range
	Name() string
	Description() string
	Location() Feature
}

type Offsetter interface {
	SetOffset(int)
}

type Mutable interface {
	SetStart(int)
	SetEnd(int)
}

type LocationSetter interface {
	SetLocation(Feature)
}

type Pair interface {
	Features() [2]Feature
}

type Set interface {
	Features() []Feature
}

type Adder interface {
	Set
	Add(...Feature)
}

type Collection interface {
	Set
	Location() Feature
}
