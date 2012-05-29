// Copyright ©2011-2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

// Use morass when you don't want your data to be a quagmire.
//
// Sort data larger than can fit in memory.
//
//  morass məˈras/
//  1. An area of muddy or boggy ground.
//  2. A complicated or confused situation.
package morass

import (
	"container/heap"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"runtime"
	"sort"
	"sync"
)

var (
	m          = &sync.Mutex{}
	registered = make(map[reflect.Type]struct{})
	nextID     = 0
)

func register(e interface{}, t reflect.Type) {
	m.Lock()
	defer m.Unlock()
	defer func() {
		recover()                  // The only panic that we can get is from trying to register a base type.
		registered[t] = struct{}{} // Remember for next time.
	}()

	if _, exists := registered[t]; !exists {
		registered[t] = struct{}{}
		gob.RegisterName(fmt.Sprintf("ℳ%d", nextID), e)
		nextID++
	}
}

// Is the receiver less than the parameterised interface
type LessInterface interface {
	Less(i interface{}) bool
}

type sortable []LessInterface

func (self sortable) Len() int { return len(self) }

func (self sortable) Less(i, j int) bool { return self[i].Less(self[j]) }

func (self *sortable) Swap(i, j int) { (*self)[i], (*self)[j] = (*self)[j], (*self)[i] }

type file struct {
	head    LessInterface
	file    *os.File
	encoder *gob.Encoder
	decoder *gob.Decoder
}

type files []*file

func (self files) Len() int { return len(self) }

func (self files) Less(i, j int) bool { return self[i].head.Less(self[j].head) }

func (self *files) Swap(i, j int) { (*self)[i], (*self)[j] = (*self)[j], (*self)[i] }

func (self *files) Pop() (i interface{}) {
	i = (*self)[len(*self)-1]
	*self = (*self)[:len(*self)-1]
	return
}

func (self *files) Push(x interface{}) { *self = append(*self, x.(*file)) }

// Type to manage sorting very large data sets.
// Setting AutoClean to true causes the Morass to delete temporary sort files
// when they are depleted.
type Morass struct {
	mutex       sync.Mutex
	t           reflect.Type
	pos, length int64
	chunk       sortable
	chunkSize   int
	pool        chan sortable
	writable    chan sortable
	err         error
	prefix      string
	dir         string
	files       files
	fast        bool
	AutoClear   bool
	AutoClean   bool
}

// Create a new Morass. prefix and dir are passed to ioutil.TempDir. chunkSize specifies
// the amount of sorting to be done in memory, concurrent specifies that temporary file
// writing occurs concurrently with sorting.
// An error is returned if no temporary directory can be created.
// Note that the type is registered with the underlying gob encoder using the name ℳn, where
// n is a sequentially assigned integer string, when the type registered. This is done to avoid using
// too much space and will cause problems when using gob itself on this type. If you intend
// use gob itself with this the type, preregister with gob and morass will use the existing
// registration.
func New(e interface{}, prefix, dir string, chunkSize int, concurrent bool) (*Morass, error) {
	d, err := ioutil.TempDir(dir, prefix)
	if err != nil {
		return nil, err
	}

	m := &Morass{
		chunkSize: chunkSize,
		prefix:    prefix,
		dir:       d,
		pool:      make(chan sortable, 2),
		writable:  make(chan sortable, 1),
		files:     files{},
	}

	m.t = reflect.TypeOf(e)
	register(e, m.t)

	m.chunk = make(sortable, 0, chunkSize)
	if concurrent {
		m.pool <- nil
	}

	f := func(self *Morass) {
		if self.AutoClean {
			self.CleanUp()
		}
	}
	runtime.SetFinalizer(m, f)

	return m, nil
}

// Push a value on to the Morass. Returns any error that occurs.
func (self *Morass) Push(e LessInterface) (err error) {
	if t := reflect.TypeOf(e); t != self.t {
		return errors.New(fmt.Sprintf("Type mismatch: %s != %s", t, self.t))
	}
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if self.err != nil {
		return self.err
	}

	if self.chunk == nil {
		return errors.New("Push on finalised morass")
	}

	if len(self.chunk) == self.chunkSize {
		self.writable <- self.chunk
		go self.write()
		self.chunk = <-self.pool
		if self.err != nil {
			return self.err
		}
		if cap(self.chunk) == 0 {
			self.chunk = make(sortable, 0, self.chunkSize)
		}
	}

	self.chunk = append(self.chunk, e)
	self.pos++
	self.length++

	return
}

func (self *Morass) write() {
	writing := <-self.writable
	defer func() {
		self.pool <- writing[:0]
	}()

	sort.Sort(&writing)

	var tf *os.File
	if tf, self.err = ioutil.TempFile(self.dir, self.prefix); self.err != nil {
		return
	}

	enc := gob.NewEncoder(tf)
	dec := gob.NewDecoder(tf)
	f := &file{head: nil, file: tf, encoder: enc, decoder: dec}
	self.files = append(self.files, f)

	for _, e := range writing {
		if self.err = enc.Encode(&e); self.err != nil {
			return
		}
	}

	self.err = tf.Sync()
}

// Return the corrent position of the cursor in the Morass.
func (self *Morass) Pos() int64 { return self.pos }

// Return the corrent length of the Morass.
func (self *Morass) Len() int64 { return self.length }

// Indicate that the last element has been pushed on to the Morass and write out final data.
// Returns any error that occurs.
func (self *Morass) Finalise() (err error) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if self.err != nil {
		return self.err
	}

	if self.chunk != nil {
		if self.pos < int64(cap(self.chunk)) {
			self.fast = true
			sort.Sort(&self.chunk)
		} else {
			if len(self.chunk) > 0 {
				self.writable <- self.chunk
				self.chunk = nil
				self.write()
				if self.err != nil {
					return self.err
				}
			}
		}
		self.pos = 0
	} else {
		return nil
	}

	if !self.fast {
		for _, f := range self.files {
			_, err = f.file.Seek(0, 0)
			if err != nil {
				return
			}
			err = f.decoder.Decode(&f.head)
			if err != nil && err != io.EOF {
				return
			}
		}

		heap.Init(&self.files)
	}

	return nil
}

// Reset the Morass to an empty state.
// Returns any error that occurs.
func (self *Morass) Clear() (err error) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	return self.clear()
}

func (self *Morass) clear() (err error) {
	for _, f := range self.files {
		err = f.file.Close()
		if err != nil {
			return
		}
		err = os.Remove(f.file.Name())
		if err != nil {
			return
		}
	}
	self.err = nil
	self.files = self.files[:0]
	self.pos = 0
	self.length = 0
	select {
	case self.chunk = <-self.pool:
		if self.chunk == nil {
			self.chunk = make(sortable, 0, self.chunkSize)
		}
	default:
	}

	return
}

// Delete the file system components of the Morass. After this call the Morass is not usable.
// Returns any error that occurs.
func (self *Morass) CleanUp() (err error) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	return os.RemoveAll(self.dir)
}

// Set the settable value e to the lowest value in the Morass.
// io.EOF indicate the Morass is empty. Any other error results in no value being set on e.
func (self *Morass) Pull(e LessInterface) (err error) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	v := reflect.ValueOf(e)
	if !reflect.Indirect(v).CanSet() {
		return errors.New("morass: Cannot set e")
	}

	if self.fast {
		switch {
		case self.chunk != nil && self.pos < int64(len(self.chunk)):
			e = self.chunk[self.pos].(LessInterface)
			self.pos++
		case self.chunk != nil:
			self.pool <- self.chunk[:0]
			self.chunk = nil
			fallthrough
		default:
			if self.AutoClear {
				self.clear()
			}
			err = io.EOF
		}
	} else {
		if self.files.Len() > 0 {
			low := heap.Pop(&self.files).(*file)
			e = low.head
			self.pos++
			switch err = low.decoder.Decode(&low.head); err {
			case nil:
				heap.Push(&self.files, low)
			case io.EOF:
				err = nil
				fallthrough
			default:
				low.file.Close()
				if self.AutoClear {
					os.Remove(low.file.Name())
				}
			}
		} else {
			if self.AutoClear {
				self.clear()
			}
			if self.AutoClean {
				os.RemoveAll(self.dir)
			}
			err = io.EOF
		}
	}

	if err != io.EOF {
		reflect.Indirect(v).Set(reflect.ValueOf(e))
	}

	return
}
