// A package that allows arrays greater than MaxInt (currently int is implemented as 32 bits
// wide independent of architecture.
//
// A BigArray is an array slice of array slices with the index being split into high and low
// bits.
// 
// The current implementation is only for BigArray of int32 as using the generic interface{}
// will double the space required for the sake of generality. Other types are trivially
// implementable.
//
// This package SHOULD NOT be necessary, but unfortunately it is.
package bigarray
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
	"fmt"
	"runtime"
	"bio/util"
)

const (
	maxInt32 = int(^uint32(0) >> 1)
)

// Int32 BigArray type
type BigI32Array struct {
	body      [][]int32
	sliceSize int
	loBitMask int64
	loBits    byte
}

// Create a new BigArray.
func NewBigI32Array(size int64, loBits byte) (b *BigI32Array) {
	sliceSize := 1 << loBits
	residualLoBitMask := int64(^uint(0) >> loBits)
	loBitMask := int64(^uint(residualLoBitMask << loBits))

	lo := int(size & loBitMask)
	hi := int(size >> loBits)
	body := make([][]int32, hi+1)
	for i, _ := range body {
		body[i] = make([]int32, sliceSize)
	}
	body[len(body)-1] = body[len(body)-1][:lo]
	b = &BigI32Array{
		body:      body,
		sliceSize: sliceSize,
		loBits:    loBits,
		loBitMask: loBitMask,
	}

	return
}

// Return the lenght of the array.
func (self *BigI32Array) Len() int64 {
	return int64(len(self.body))*int64(self.sliceSize) + int64(len(self.body[len(self.body)-1]))
}

// Return the value at postion i.
func (self *BigI32Array) At(i int64) (v int32) {
	lo := int(i & self.loBitMask)
	hi := int(i >> self.loBits)
	v = self.body[hi][lo]

	return
}

// Return a BigArray slice from position start to position end.
func (self *BigI32Array) Slice(start, end int64) (slice *BigI32Array, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			if err, ok = r.(error); !ok {
				caller := util.Name(1) // is this the right level? defer may screw with things
				err = fmt.Errorf("%v.%v: %v", caller.Package, caller.Function, r)
			}
		}
	}()

	slo := int(start & self.loBitMask)
	shi := int(start >> self.loBits)
	elo := int(end & self.loBitMask)
	ehi := int(end >> self.loBits)
	slice = NewBigI32Array(end-start+1, self.loBits)
	if shi == ehi {
		copy(slice.body[0], self.body[shi][slo:elo])
	} else {
		for i := 0; i < len(slice.body); i++ {
			copy(slice.body[i][:], self.nativeSlice(int64(i*self.sliceSize), int64((i+1)*self.sliceSize-1)))
		}
	}

	return
}

// Append a BigArray to the current BigArray
func (self *BigI32Array) Append(a *BigI32Array, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			if err, ok = r.(error); !ok {
				caller := util.Name(1) // is this the right level? defer may screw with things
				err = fmt.Errorf("%v.%v: %v", caller.Package, caller.Function, r)
			}
		}
	}()

	offset := int64(self.sliceSize - len(self.body[len(self.body)-1]))

	t := NewBigI32Array(self.Len()+a.Len(), self.loBits)

	for i, _ := range self.body {
		copy(t.body[i], self.body[i])
	}

	t.body[len(self.body)-1] = append(t.body[len(self.body)-1], a.body[0][:int(offset)]...)

	var last int64
	for i := int64(0); (i+1)*int64(self.sliceSize)+offset < a.Len(); i++ {
		t.body[int(i)+len(self.body)] = a.nativeSlice(i*int64(self.sliceSize)+offset, (i+1)*int64(self.sliceSize)+offset-1)
		last = i
	}

	t.body[len(t.body)-1] = a.nativeSlice(last*int64(self.sliceSize)+offset, (last+1)*int64(self.sliceSize)+offset-1)

	self = t
	runtime.GC()
}

// Alter the underlying geometry of the BigArray. loBits specifies the width of the addressing
// of the low order-addressed slices. A smaller number decreases the average amount of wasted
// space but will have an impact on slicing and appending speed.
func (self *BigI32Array) Reshape(loBits byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			if err, ok = r.(error); !ok {
				caller := util.Name(1) // is this the right level? defer may screw with things
				err = fmt.Errorf("%v.%v: %v", caller.Package, caller.Function, r)
			}
		}
	}()

	t := NewBigI32Array(self.Len(), loBits)

	for i := int64(0); i < int64(len(t.body))-1; i++ {
		t.body[i] = self.nativeSlice(int64(i)*int64(t.sliceSize), (i+1)*int64(t.sliceSize)-1)
	}

	t.body[len(t.body)] = self.nativeSlice(int64(len(t.body)*t.sliceSize), t.Len()-1)

	self = t
	runtime.GC()
}

// Return a Go native slice from a BigArray - must be no longer than MaxInt
func (self *BigI32Array) NativeSlice(start, end int64) (slice []int32, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			if err, ok = r.(error); !ok {
				caller := util.Name(1) // is this the right level? defer may screw with things
				err = fmt.Errorf("%v.%v: %v", caller.Package, caller.Function, r)
			}
		}
	}()

	slice = self.nativeSlice(start, end)

	return
}

// Internal native slice method panics instead of returning error
func (self *BigI32Array) nativeSlice(start, end int64) (slice []int32) {
	slo := int(start & self.loBitMask)
	shi := int(start >> self.loBits)
	elo := int(end & self.loBitMask)
	ehi := int(end >> self.loBits)
	if end-start > int64(maxInt32) {
		panic("Attempting to take a native slice bigger than Go array capacity") // Caught by package methods
	}
	slice = make([]int32, int(end-start))
	if shi == ehi {
		copy(slice, self.body[shi][slo:elo])
	} else {
		copy(slice, self.body[shi][slo:])
		for i := shi + 1; i <= ehi; i++ {
			slice = append(slice, self.body[ehi][:elo]...)
		}
	}

	return
}

// Set the value at position i to x.
func (self *BigI32Array) Set(i int64, x int32) {
	lo := int(i & self.loBitMask)
	hi := int(i >> self.loBits)
	self.body[hi][lo] = x
}

// DoAt perform a function on position i. This is included to speed up some array operations.
func (self *BigI32Array) DoAt(i int64, f func(int32, interface{}) int32, params interface{}) {
	lo := int(i & self.loBitMask)
	hi := int(i >> self.loBits)
	self.body[hi][lo] = f(self.body[hi][lo], params)
}

// Do calls function f for each element of the vector, in order.
func (self *BigI32Array) Do(f func(int32, interface{}) int32, params interface{}) {
	for i := int64(0); i < self.Len()-1; i++ {
		self.DoAt(i, f, params)
	}
}
