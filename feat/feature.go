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

// Orienter wraps the Orientation method.
//
// Orientation returns the orientation of the feature relative to its location.
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

// Feature is a Range whose coordinates are defined relative to a feature
// location. Start and End return the coordinates of the feature relative to
// its location which can be nil. In the latter case callers should make no
// assumptions whether coordinates of such features are comparable.
type Feature interface {
	Range
	// Name returns the name of the feature.
	Name() string
	// Description returns the description of the feature.
	Description() string
	// Location returns the reference feature on which the feature is located.
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

// BasePositionOf returns the position in f coordinates converted to
// coordinates relative to the first nil feature location, and a reference
// which is the feature location preceding the nil. The returned reference
// feature should be used by callers of BasePositionOf to verify that
// coordinates are comparable.
// BasePositionOf will panic if the feature chain is deeper than 1000 links.
func BasePositionOf(f Feature, position int) (int, Feature) {
	for n := 0; n < 1000; n++ {
		position += f.Start()
		if f.Location() != nil {
			f = f.Location()
			continue
		}
		return position, f
	}
	panic("feat: feature chain too long")
}

// PositionWithin returns the position in f coordinates converted to
// coordinates relative to the given reference feature and a boolean
// indicating whether f can be located relative to ref.
// PositionWithin will panic if the feature chain is deeper than 1000 links.
func PositionWithin(f, ref Feature, position int) (pos int, ok bool) {
	for n := 0; n < 1000; n++ {
		if f == ref {
			return position, f != nil
		}
		position += f.Start()
		if f.Location() != nil {
			f = f.Location()
			continue
		}
		return 0, false
	}
	panic("feat: feature chain too long")
}

// BaseOrientationOf returns the orientation of f relative to a reference
// which is the first non-nil, non-orientable feature location, and that
// reference feature. The returned reference feature should be used by callers
// of BaseOrientationOf to verify that orientations are comparable. If f is
// not orientable, the returned orientation will be NotOriented and the
// reference will be the first orientable or last non-nil feature.
// BaseOrientationOf will panic if the feature chain is deeper than 1000 links.
func BaseOrientationOf(f Feature) (ori Orientation, ref Feature) {
	o, ok := f.(Orienter)
	if !ok || o.Orientation() == NotOriented {
		for n := 0; n < 1000; n++ {
			if o, ok = f.Location().(Orienter); ok && o.Orientation() != NotOriented {
				return NotOriented, f.Location()
			}
			if f.Location() == nil {
				return NotOriented, f
			}
			f = f.Location()
		}
		panic("feat: feature chain too long")
	}

	ori = Forward
	for n := 0; n < 1000; n++ {
		ori *= o.Orientation()
		if o, ok = f.Location().(Orienter); ok && o.Orientation() != NotOriented {
			f = f.Location()
			continue
		}
		if f.Location() != nil {
			return ori, f.Location()
		}
		return ori, f
	}
	panic("feat: feature chain too long")
}

// OrientationWithin returns the orientation of f relative to the given
// reference feature. The returned orientation will be NotOriented if f is not
// located within the reference or if f is not orientable.
// OrientationWithin will panic if the feature chain is deeper than 1000 links.
func OrientationWithin(f, ref Feature) Orientation {
	if ref == nil {
		return NotOriented
	}
	ori := Forward
	for n := 0; n < 1000; n++ {
		o, ok := f.(Orienter)
		if !ok {
			return NotOriented
		}
		if o := o.Orientation(); o != NotOriented {
			if f == ref {
				return ori
			}
			ori *= o
			f = f.Location()
			if f == ref {
				return ori
			}
			continue
		}
		return NotOriented
	}
	panic("feat: feature chain too long")
}
