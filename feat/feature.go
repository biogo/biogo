// Copyright ©2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package feat provides the base for storage and manipulation of biological interval information.
package feat

type Conformationer interface {
	Conformation() Conformation
	SetConformation(Conformation) error
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

type OrientSetter interface {
	SetOrientation(Orientation) error
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
	SetOffset(int) error
}

type Mutable interface {
	SetStart(int) error
	SetEnd(int) error
}

type LocationSetter interface {
	SetLocation(Feature) error
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

// OrientationOf returns the orientation of the given feature relative to
// the deepest orientable location and the reference feature, which may be
// the feature itself if it is not an Orienter or has a nil Location.
// The returned orientation will always be Forward or Reverse.
// OrientationOf will panic if the feature chain is deeper than 1000 links.
func OrientationOf(f Feature) (ori Orientation, ref Feature) {
	ori = Forward
	for n := 0; n < 1000; n++ {
		o, ok := f.(Orienter)
		if !ok {
			return ori, f
		}
		if o := o.Orientation(); o != NotOriented {
			ori *= o
			f = f.Location()
			continue
		}
		return ori, f
	}
	panic("feat: feature chain too long")
}
