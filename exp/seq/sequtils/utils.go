// Package sequtils provides generic functions for manipulation of slices used in the seq types.
package sequtils

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
	"github.com/kortschak/biogo/bio"
	"github.com/kortschak/biogo/exp/seq"
	"github.com/kortschak/biogo/feat"
	"github.com/kortschak/biogo/interval"
	"github.com/kortschak/biogo/util"
	"reflect"
)

var emptyString = ""

// Truncate provides a function that may be used by polymer types to implement Truncator.
// It makes use of reflection and so may be slower than type-specific implementations.
// This is the reference implementation and should be used to compare type-specific
// implementation against in testing. 
func Truncate(pol interface{}, start, end int, circular bool) (p interface{}, err error) {
	pv := reflect.ValueOf(pol)
	if l := pv.Len(); start < 0 || end < 0 || start > l || end > l {
		return nil, bio.NewError("Out of range.", 0, nil)
	}
	if start <= end {
		p = pv.Slice(start, end).Interface()
	} else if circular {
		tv := reflect.MakeSlice(pv.Type(), pv.Len()-start, pv.Len()+end-start)
		reflect.Copy(tv, pv.Slice(start, pv.Len()))
		p = reflect.AppendSlice(tv, pv.Slice(0, end)).Interface()
	} else {
		return nil, bio.NewError("Start position greater than end position for non-circular sequence.", 0, pol)
	}

	return
}

// Reverse provides a function that may be used by polymer types to implement Polymer.
// It makes use of reflection and so may be slower than type-specific implementations.
// This is the reference implementation and should be used to compare type-specific
// implementation against in testing. 
func Reverse(pol interface{}) interface{} {
	pv := reflect.ValueOf(pol)
	pLen := pv.Len()

	i, j := 0, pLen-1
	var tv interface{}
	for ; i < j; i, j = i+1, j-1 {
		tv = pv.Index(i).Interface()
		pv.Index(i).Set(pv.Index(j))
		pv.Index(j).Set(reflect.ValueOf(tv))
	}

	return pv.Interface()
}

// Join provides a function that may be used by polymer types.
// It makes use of reflection and so may be slower than type-specific implementations.
// This is the reference implementation and should be used to compare type-specific
// implementation against in testing. 
func Join(pol1, pol2 interface{}, where int) (j interface{}, offset int) {
	if where == seq.End {
		pol2, pol1 = pol1, pol2
	}
	pol2v := reflect.ValueOf(pol2)
	pol1v := reflect.ValueOf(pol1)
	pol2Len := pol2v.Len()
	if where == seq.Start {
		offset = -pol2v.Len()
	}

	tv := reflect.MakeSlice(pol1v.Type(), pol2Len, pol2Len+pol1v.Len())
	reflect.Copy(tv, pol2v)
	j = reflect.AppendSlice(tv, pol1v).Interface()

	return
}

// Stitch provides a function that may be used by polymer types to implement Stitcher.
// It makes use of reflection and so may be slower than type-specific implementations.
// This is the reference implementation and should be used to compare type-specific
// implementation against in testing. 
func Stitch(pol interface{}, offset int, f feat.FeatureSet) (s interface{}, err error) {
	t := interval.NewTree()
	var i *interval.Interval

	for _, feature := range f {
		i, err = interval.New(emptyString, feature.Start, feature.End, 0, nil)
		if err != nil {
			return
		} else {
			t.Insert(i)
		}
	}

	pv := reflect.ValueOf(pol)
	pLen := pv.Len()
	end := pLen + offset
	span, err := interval.New(emptyString, offset, end, 0, nil)
	if err != nil {
		panic("Sequence.End() < Sequence.Start()")
	}
	fs, _ := t.Flatten(span, 0, 0)
	l := 0

	for _, seg := range fs {
		l += util.Min(seg.End(), end) - util.Max(seg.Start(), offset)
	}
	tv := reflect.MakeSlice(pv.Type(), 0, l)

	for _, seg := range fs {
		tv = reflect.AppendSlice(tv, pv.Slice(util.Max(seg.Start()-offset, 0), util.Min(seg.End()-offset, pLen)))
	}

	return tv.Interface(), nil
}

// Compose provides a function that may be used by polymer types to implement Composer.
// It makes use of reflection and so may be slower than type-specific implementations.
// This is the reference implementation and should be used to compare type-specific
// implementation against in testing. 
func Compose(pol interface{}, offset int, f feat.FeatureSet) (s []interface{}, err error) {
	pv := reflect.ValueOf(pol)
	pLen := pv.Len()
	end := pLen + offset

	tv := make([]reflect.Value, len(f))
	for i, seg := range f {
		if seg.End < seg.Start {
			return nil, bio.NewError("Feature End < Start", 0, f)
		}
		l := util.Min(seg.End, end) - util.Max(seg.Start, offset)
		tv[i] = reflect.MakeSlice(pv.Type(), l, l)
		reflect.Copy(tv[i], pv.Slice(util.Max(seg.Start-offset, 0), util.Min(seg.End-offset, pLen)))
	}

	s = make([]interface{}, len(tv))
	for i := range tv {
		s[i] = tv[i].Interface()
	}

	return
}
