// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package morass implements file system-backed sorting.
//
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
	registerLock = &sync.Mutex{}
	registered   = make(map[reflect.Type]struct{})
	nextID       = 0
)

func register(e interface{}, t reflect.Type) {
	registerLock.Lock()
	defer registerLock.Unlock()
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

// LessInterface wraps the Less method.
type LessInterface interface {
	// Is the receiver less than the parameterised interface
	Less(i interface{}) bool
}

type sorter []LessInterface

func (s sorter) Len() int { return len(s) }

func (s sorter) Less(i, j int) bool { return s[i].Less(s[j]) }

func (s sorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type file struct {
	head    LessInterface
	file    *os.File
	encoder *gob.Encoder
	decoder *gob.Decoder
}

type files []*file

func (f files) Len() int { return len(f) }

func (f files) Less(i, j int) bool { return f[i].head.Less(f[j].head) }

func (f files) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

func (f *files) Pop() (i interface{}) {
	i = (*f)[len(*f)-1]
	*f = (*f)[:len(*f)-1]
	return
}

func (f *files) Push(x interface{}) { *f = append(*f, x.(*file)) }

// Morass implements sorting of very large data sets.
type Morass struct {
	typ reflect.Type

	pos, length int64

	// dir and prefix specify the location
	// of temporary file.
	dir    string
	prefix string

	// AutoClear specifies that the Morass
	// should call Clear when emptied by
	// a call to Pull.
	AutoClear bool

	// AutoClean specifies that the Morass
	// should call delete temporary sort
	// files when it has been emptied by
	// a call to Pull.
	AutoClean bool

	// fast indicates sorting was performed
	// entirely in memory.
	fast bool

	chunk     sorter
	chunkSize int
	pool      chan sorter
	writable  chan sorter

	filesLock sync.Mutex
	files     files

	errLock sync.Mutex
	_err    error
}

// New creates a new Morass. prefix and dir are passed to ioutil.TempDir. chunkSize specifies
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
		pool:      make(chan sorter, 2),
		writable:  make(chan sorter, 1),
		files:     files{},
	}

	m.typ = reflect.TypeOf(e)
	register(e, m.typ)

	m.chunk = make(sorter, 0, chunkSize)
	if concurrent {
		m.pool <- nil
	}

	runtime.SetFinalizer(m, func(x *Morass) {
		if x.AutoClean {
			x.CleanUp()
		}
	})

	return m, nil
}

// Push a value on to the Morass. Returns any error that occurs.
func (m *Morass) Push(e LessInterface) error {
	if typ := reflect.TypeOf(e); typ != m.typ {
		return fmt.Errorf("morass: type mismatch: %s != %s", typ, m.typ)
	}

	if err := m.err(); err != nil {
		return err
	}

	if m.chunk == nil {
		return errors.New("morass: push on finalised morass")
	}

	if len(m.chunk) == m.chunkSize {
		m.writable <- m.chunk
		go m.write()
		m.chunk = <-m.pool
		if err := m.err(); err != nil {
			return err
		}
		if cap(m.chunk) == 0 {
			m.chunk = make(sorter, 0, m.chunkSize)
		}
	}

	m.chunk = append(m.chunk, e)
	m.pos++
	m.length++

	return nil
}

func (m *Morass) write() {
	writing := <-m.writable
	defer func() {
		m.pool <- writing[:0]
	}()

	sort.Sort(writing)

	tf, err := ioutil.TempFile(m.dir, m.prefix)
	if err != nil {
		m.setErr(err)
		return
	}

	enc := gob.NewEncoder(tf)
	dec := gob.NewDecoder(tf)
	f := &file{head: nil, file: tf, encoder: enc, decoder: dec}

	m.filesLock.Lock()
	m.files = append(m.files, f)
	m.filesLock.Unlock()

	for _, e := range writing {
		if err := enc.Encode(&e); err != nil {
			m.setErr(err)
			return
		}
	}

	m.setErr(tf.Sync())
}

func (m *Morass) setErr(err error) {
	m.errLock.Lock()
	m._err = err
	m.errLock.Unlock()
}

func (m *Morass) err() error {
	m.errLock.Lock()
	defer m.errLock.Unlock()
	return m._err
}

// Pos returns the current position of the cursor in the Morass.
func (m *Morass) Pos() int64 { return m.pos }

// Len returns the current length of the Morass.
func (m *Morass) Len() int64 { return m.length }

// Finalise is called to indicate that the last element has been pushed on to the Morass
// and write out final data.
func (m *Morass) Finalise() error {
	if err := m.err(); err != nil {
		return err
	}

	if m.chunk != nil {
		if m.pos < int64(cap(m.chunk)) {
			m.fast = true
			sort.Sort(m.chunk)
		} else {
			if len(m.chunk) > 0 {
				m.writable <- m.chunk
				m.chunk = nil
				m.write()
				if err := m.err(); err != nil {
					return err
				}
			}
		}
		m.pos = 0
	} else {
		return nil
	}

	if !m.fast {
		for _, f := range m.files {
			_, err := f.file.Seek(0, 0)
			if err != nil {
				return err
			}
			err = f.decoder.Decode(&f.head)
			if err != nil && err != io.EOF {
				return err
			}
		}

		heap.Init(&m.files)
	}

	return nil
}

// Clear resets the Morass to an empty state.
func (m *Morass) Clear() error {
	var err error
	for _, f := range m.files {
		err = f.file.Close()
		if err != nil {
			return err
		}
		err = os.Remove(f.file.Name())
		if err != nil {
			return err
		}
	}
	m._err = nil
	m.files = m.files[:0]
	m.pos = 0
	m.length = 0
	select {
	case m.chunk = <-m.pool:
		if m.chunk == nil {
			m.chunk = make(sorter, 0, m.chunkSize)
		}
	default:
	}

	return nil
}

// CleanUp deletes the file system components of the Morass. After this call
// the Morass is not usable.
func (m *Morass) CleanUp() error {
	return os.RemoveAll(m.dir)
}

// Pull sets the settable value e to the lowest value in the Morass.
// If io.EOF is returned the Morass is empty. Any other error results
// in no value being set on e.
func (m *Morass) Pull(e LessInterface) error {
	var err error
	v := reflect.ValueOf(e)
	if !reflect.Indirect(v).CanSet() {
		return errors.New("morass: Cannot set e")
	}

	if m.fast {
		switch {
		case m.chunk != nil && m.pos < int64(len(m.chunk)):
			e = m.chunk[m.pos].(LessInterface)
			m.pos++
		case m.chunk != nil:
			m.pool <- m.chunk[:0]
			m.chunk = nil
			fallthrough
		default:
			if m.AutoClear {
				m.Clear()
			}
			err = io.EOF
		}
	} else {
		if m.files.Len() > 0 {
			low := heap.Pop(&m.files).(*file)
			e = low.head
			m.pos++
			switch err = low.decoder.Decode(&low.head); err {
			case nil:
				heap.Push(&m.files, low)
			case io.EOF:
				err = nil
				fallthrough
			default:
				low.file.Close()
				if m.AutoClear {
					os.Remove(low.file.Name())
				}
			}
		} else {
			if m.AutoClear {
				m.Clear()
			}
			if m.AutoClean {
				os.RemoveAll(m.dir)
			}
			err = io.EOF
		}
	}

	if err != io.EOF {
		reflect.Indirect(v).Set(reflect.ValueOf(e))
	}

	return err
}
