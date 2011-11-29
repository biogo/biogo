//Use morass when you don't want your data to be a quagmire.
//
//Sort data larger than can fit in memory.
//
//  morass məˈras/
//    1. An area of muddy or boggy ground.
//    2. A complicated or confused situation.
package morass
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
	"bio"
	"container/heap"
	"encoding/gob"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"runtime"
	"sort"
	"sync"
)

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

type Morass struct {
	mutex       sync.Mutex
	pos, length int64
	chunk       sortable
	done        chan sortable
	err         chan error
	prefix      string
	dir         string
	files       files
	finalised   bool
	fast        bool
	AutoClean   bool
}

func New(prefix, dir string, chunkSize int, concurrent bool) (*Morass, error) {
	d, err := ioutil.TempDir(dir, prefix)
	if err != nil {
		return nil, err
	}

	m := &Morass{
		prefix: prefix,
		dir:    d,
		done:   make(chan sortable, 1),
		files:  make(files, 0),
		err:    make(chan error, 1),
	}

	m.chunk = make(sortable, 0, chunkSize)
	if concurrent {
		m.done <- make(sortable, 0)
	}

	f := func(self *Morass) {
		if self.AutoClean {
			self.CleanUp()
		}
	}
	runtime.SetFinalizer(m, f)

	return m, nil
}

func (self *Morass) Push(e LessInterface) (err error) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	select {
	case err = <-self.err:
		if err != nil {
			return
		}
	default:
	}

	if self.finalised {
		return bio.NewError("Push on finalised morass", 0, nil)
	}

	if c := cap(self.chunk); len(self.chunk) == c {
		go self.write(self.chunk)
		self.chunk = <-self.done
		if cap(self.chunk) == 0 {
			self.chunk = make(sortable, 0, c)
		}
	}

	gob.Register(e)
	self.chunk = append(self.chunk, e)
	self.pos++
	self.length++

	return
}

func (self *Morass) write(writing sortable) (err error) {
	defer func() {
		self.err <- err
		self.done <- writing[:0]
	}()

	select {
	case <-self.err:
	default:
	}

	sort.Sort(&writing)

	var tf *os.File
	if tf, err = ioutil.TempFile(self.dir, self.prefix); err != nil {
		return
	}

	enc := gob.NewEncoder(tf)
	dec := gob.NewDecoder(tf)
	f := &file{head: nil, file: tf, encoder: enc, decoder: dec}
	self.files = append(self.files, f)

	for _, e := range writing {
		if err = enc.Encode(&e); err != nil {
			return
		}
	}

	err = tf.Sync()

	return
}

func (self *Morass) Pos() int64 { return self.pos }

func (self *Morass) Len() int64 { return self.length }

func (self *Morass) Finalise() (err error) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	select {
	case err = <-self.err:
		if err != nil {
			return
		}
	default:
	}

	if !self.finalised {
		if self.pos < int64(cap(self.chunk)) {
			self.fast = true
			sort.Sort(&self.chunk)
		} else {
			if len(self.chunk) > 0 {
				go self.write(self.chunk)
				err = <-self.err
			}
		}
		self.pos = 0
		self.finalised = true
	} else {
		return nil
	}

	if !self.fast {
		for _, f := range self.files {
			if _, err = f.file.Seek(0, 0); err != nil {
				return
			}
			if err = f.decoder.Decode(&f.head); err != nil && err != io.EOF {
				return
			}
		}

		heap.Init(&self.files)
	}

	return nil
}

func (self *Morass) Clear() (err error) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	for _, f := range self.files {
		if err = f.file.Close(); err != nil {
			return
		}
		if err = os.Remove(f.file.Name()); err != nil {
			return
		}
		f = nil
	}
	self.pos = 0
	self.length = 0
	self.finalised = false
	self.chunk = self.chunk[:0]

	return
}

func (self *Morass) CleanUp() (err error) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	return os.RemoveAll(self.dir)
}

func (self *Morass) Pull(e LessInterface) (err error) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if len(self.chunk) == 0 {
		return io.EOF
	}

	v := reflect.ValueOf(e)
	if !reflect.Indirect(v).CanSet() {
		return errors.New("morass: Cannot set e")
	}

	if self.fast {
		if self.pos < int64(len(self.chunk)) {
			e = self.chunk[self.pos].(LessInterface)
			self.pos++
			err = nil
		} else {
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
				if self.AutoClean {
					os.Remove(low.file.Name())
				}
			}
		} else {
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
