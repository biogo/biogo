// Sparse matrix operations
package sparse

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

// N.B. There is fairly extensive use of unrecovered panics in sparse to avoid unreadable matrix formulae.

import (
	"errors"
	"fmt"
	"github.com/kortschak/biogo/matrix"
	"math"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
)

var (
	MaxProcs    int     // Number of goroutines to run concurrently when multiplyin matrices
	Margin      int = 3 // Number of columns/rows visible returned by String
	workbuffers chan sparsecol
	BufferLen   int = 100
)

func init() {
	Init()
}

// Initialise goroutine and memory handling
func Init() {
	MaxProcs = runtime.GOMAXPROCS(0)
	workbuffers = make(chan sparsecol, MaxProcs)
	for i := 0; i < MaxProcs; i++ {
		buffer := make(sparsecol, BufferLen)
		workbuffers <- buffer
	}
}

// Sparse matrix type
type Sparse struct {
	r, c   int
	matrix []sparsecol
}

// Return a sparse matrix based on a slice of float64 slices
func Matrix(a [][]float64) (m *Sparse) {
	if len(a) == 0 {
		panic(errors.New("zero dimension in matrix definition"))
	}

	maxRowLen := 0
	for _, r := range a {
		if len(r) > maxRowLen {
			maxRowLen = len(r)
		}
	}

	m = &Sparse{
		r:      len(a),
		matrix: make([]sparsecol, maxRowLen),
	}

	for i, r := range a {
		for j, v := range r {
			if v != 0 {
				if j > len(m.matrix) {
					t := make([]sparsecol, j+1)
					copy(t, m.matrix)
					m.matrix = t
				}
				m.matrix[j] = append(m.matrix[j], elem{r: i, value: v})
			}
		}
	}

	if m.c = len(m.matrix); m.c == 0 {
		panic(errors.New("zero dimension in matrix definition"))
	}

	return
}

// Return the O matrix
func Zero(r, c int) (m *Sparse) {
	if r < 1 || c < 1 {
		panic(errors.New("zero dimension in matrix definition"))
	}

	return &Sparse{
		r:      r,
		c:      c,
		matrix: make([]sparsecol, c),
	}
}

// Return the I matrix
func Identity(s int) (m *Sparse) {
	if s < 1 {
		panic(errors.New("zero dimension in matrix definition"))
	}

	m = &Sparse{
		r:      s,
		c:      s,
		matrix: make([]sparsecol, s),
	}

	for i := 0; i < s; i++ {
		m.matrix[i] = append(m.matrix[i], elem{r: i, value: 1})
	}

	return
}

func Random(r, c int, density float64) (m *Sparse) {
	m = &Sparse{
		r:      r,
		c:      c,
		matrix: make([]sparsecol, c),
	}

	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			if rand.Float64() < density {
				m.Set(i, j, rand.Float64())
			}
		}
	}

	return
}

// Return of the non-zero elements of a set of matrices in column major order as a column vector.
func Elements(mats ...*Sparse) (e *Sparse) {
	var length int
	for _, m := range mats {
		for _, col := range m.matrix {
			length += len(col)
		}
	}

	t := make(sparsecol, 0, length)
	for _, m := range mats {
		for _, col := range m.matrix {
			for _, e := range col {
				if e.value != 0 {
					t = append(t, elem{r: len(t), value: e.value})
				}
			}
		}
	}

	e = &Sparse{
		r:      length,
		c:      1,
		matrix: []sparsecol{t},
	}

	return
}

// Return of the non-zero elements of the matrix in column major order as an array.
func (self *Sparse) Elements() (a []float64) {
	elements := Elements(self).matrix[0]
	a = make([]float64, 0, len(elements))
	for _, e := range Elements(self).matrix[0] {
		a = append(a, e.value)
	}

	return
}

// Return a copy of a matrix
func (self *Sparse) Copy() (m *Sparse) {
	m = &Sparse{
		r:      self.r,
		c:      self.c,
		matrix: make([]sparsecol, len(self.matrix)),
	}

	for j, col := range self.matrix {
		m.matrix[j] = make(sparsecol, len(col))
		copy(m.matrix[j], col)
	}

	return
}

// Return the dimensions of a matrix
func (self *Sparse) Dims() (r, c int) {
	return self.r, self.c
}

// Calculate the determinant of a matrix
func (self *Sparse) Det() (d float64) {
	panic("not implemented")
}

// Return the minimum non-zero of a matrix
func (self *Sparse) Min() (m float64) {
	m = math.MaxFloat64
	for _, col := range self.matrix {
		m = math.Min(col.min(), m)
	}

	return
}

// Return the maximum non-zero of a matrix
func (self *Sparse) Max() (m float64) {
	m = -math.MaxFloat64
	for _, col := range self.matrix {
		m = math.Max(col.max(), m)
	}

	return
}

// Set the value at (r, c) to v
func (self *Sparse) Set(r, c int, v float64) {
	if r >= self.r || c >= self.c || r < 0 || c < 0 {
		panic(errors.New("Out of bound"))
	}

	self.set(r, c, v)
}

func (self *Sparse) set(r, c int, v float64) {
	col := self.matrix[c]
	lo := 0
	hi := len(col)
	for {
		switch curpos := (lo + hi) / 2; {
		case lo > hi:
			col.insert(lo, elem{r, v})
			self.matrix[c] = col
			return
		case col == nil, r > col[len(col)-1].r:
			col = append(col, elem{r, v})
			self.matrix[c] = col
			return
		case col[curpos].r == r:
			col[curpos].value = v
			return
		case r < col[curpos].r:
			hi = curpos - 1
		case r > col[curpos].r:
			lo = curpos + 1
		}
	}
}

// Return the value at (r, c)
func (self *Sparse) At(r, c int) (v float64) {
	if r >= self.r || c >= self.c || c < 0 || r < 0 {
		panic(errors.New("Out of bound"))
	}
	return self.at(r, c)
}

func (self *Sparse) at(r, c int) float64 {
	col := self.matrix[c]

	lo := 0
	hi := len(col)
	for {
		switch curpos := (lo + hi) / 2; {
		case len(col) == 0, r > col[len(col)-1].r, lo > hi:
			return 0
		case col[curpos].r == r:
			return col[curpos].value
		case r < col[curpos].r:
			hi = curpos - 1
		case r > col[curpos].r:
			lo = curpos + 1
		}
	}

	panic("cannot reach")
}

// Determin a variety of norms
func (self *Sparse) Norm(ord int) (n float64) {
	if ord == 0 {
		for _, c := range self.matrix {
			for _, e := range c {
				n += e.value * e.value
			}
		}
		return math.Sqrt(n)
	}
	switch ord {
	case 2, -2:
		panic("not implemented - feel free to port an svd function to sparse")
	case 1:
		sum := self.SumAxis(matrix.Cols)
		for _, e := range sum.matrix[0] {
			n = math.Max(math.Abs(e.value), n)
		}
		return
	case matrix.Inf:
		sum := self.SumAxis(matrix.Rows)
		for _, e := range sum.matrix[0] {
			n = math.Max(math.Abs(e.value), n)
		}
		return
	case -1:
		n = math.MaxFloat64
		sum := self.SumAxis(matrix.Cols)
		for _, e := range sum.matrix[0] {
			n = math.Min(math.Abs(e.value), n)
		}
		return
	case -matrix.Inf:
		n = math.MaxFloat64
		sum := self.SumAxis(matrix.Rows)
		for _, e := range sum.matrix[0] {
			n = math.Min(math.Abs(e.value), n)
		}
		return
	case matrix.Fro:
		for _, c := range self.matrix {
			for _, e := range c {
				n += e.value * e.value
			}
		}
		return math.Sqrt(n)
	default:
		panic(errors.New("Invalid norm order for matrix"))
	}

	panic("cannot reach")
}

// Return a column or row vector holding the sums of rows or columns
func (self *Sparse) SumAxis(cols bool) (m *Sparse) {
	m = &Sparse{}
	if cols {
		wg := &sync.WaitGroup{}
		m.r, m.c, m.matrix = 1, self.c, make([]sparsecol, self.c)
		for i, c := range self.matrix {
			wg.Add(1)
			go func(i int, c sparsecol) {
				defer func() {
					wg.Done()
				}()
				m.matrix[i] = sparsecol{elem{r: 0, value: c.sum()}}
			}(i, c)
		}
		wg.Wait()
	} else {
		m.r, m.c, m.matrix = self.r, 1, make([]sparsecol, 1)
		data := make([]elem, 0, self.r)
		for i := 0; i < self.r; i++ {
			n := float64(0)
			for j := 0; j < self.c; j++ {
				n += self.at(i, j)
			}
			data = append(data, elem{i, n})
		}
		m.matrix[0] = make([]elem, len(data))
		copy(m.matrix[0], data)
	}

	return
}

// Return a column or row vector holding the max of rows or columns
func (self *Sparse) MaxAxis(cols bool) (m *Sparse) {
	m = &Sparse{}
	if cols {
		wg := &sync.WaitGroup{}
		m.r, m.c, m.matrix = 1, self.c, make([]sparsecol, self.c)
		for i, c := range self.matrix {
			wg.Add(1)
			go func(i int, c sparsecol) {
				defer func() {
					wg.Done()
				}()
				m.matrix[i] = sparsecol{elem{r: 0, value: c.max()}}
			}(i, c)
		}
		wg.Wait()
	} else {
		m.r, m.c, m.matrix = self.r, 1, make([]sparsecol, 1)
		data := make([]elem, 0, self.r)
		for i := 0; i < self.r; i++ {
			n := -math.MaxFloat64
			for j := 0; j < self.c; j++ {
				if v := self.at(i, j); v > n {
					n = v
				}
			}
			fmt.Println(i, n)
			data = append(data, elem{i, n})
		}
		m.matrix[0] = make([]elem, len(data))
		copy(m.matrix[0], data)
	}

	return
}

// Return a column or row vector holding the min of rows or columns
func (self *Sparse) MinAxis(cols bool) (m *Sparse) {
	m = &Sparse{}
	if cols {
		wg := &sync.WaitGroup{}
		m.r, m.c, m.matrix = 1, self.c, make([]sparsecol, self.c)
		for i, c := range self.matrix {
			wg.Add(1)
			go func(i int, c sparsecol) {
				defer func() {
					wg.Done()
				}()
				m.matrix[i] = sparsecol{elem{r: 0, value: c.min()}}
			}(i, c)
		}
		wg.Wait()
	} else {
		m.r, m.c, m.matrix = self.r, 1, make([]sparsecol, 1)
		data := make([]elem, 0, self.r)
		for i := 0; i < self.r; i++ {
			n := math.MaxFloat64
			for j := 0; j < self.c; j++ {
				if v := self.at(i, j); v < n {
					n = v
				}
			}
			data = append(data, elem{i, n})
		}
		m.matrix[0] = make([]elem, len(data))
		copy(m.matrix[0], data)
	}

	return
}

// Return the transpose of a matrix
func (self *Sparse) T() (m *Sparse) {
	if self.r == 0 || self.c == 0 { // this is a vector
		m = self.Copy()
		m.r, m.c = m.c, m.r
		return
	}

	m = &Sparse{
		r:      self.c,
		c:      self.r,
		matrix: make([]sparsecol, self.r),
	}
	for j, _ := range m.matrix {
		m.matrix[j] = make(sparsecol, 0, m.r)
	}
	for j, col := range self.matrix {
		for _, e := range col {
			m.matrix[e.r] = append(m.matrix[e.r], elem{r: j, value: e.value})
		}
	}

	for j, _ := range m.matrix {
		t := make(sparsecol, len(m.matrix[j]))
		copy(t, m.matrix[j])
		m.matrix[j] = t
	}

	return
}

// Add one matrix to another
func (self *Sparse) Add(b *Sparse) (m *Sparse) {
	m = &Sparse{
		r:      self.r,
		c:      self.c,
		matrix: make([]sparsecol, self.c),
	}

	wg := &sync.WaitGroup{}
	for j, col := range self.matrix {
		wg.Add(1)
		go func(j int, col sparsecol) {
			defer func() {
				wg.Done()
			}()
			m.matrix[j] = col.foldadd(b.matrix[j])
		}(j, col)
	}
	wg.Wait()

	return
}

// Subtract one matrix from another
func (self *Sparse) Sub(b *Sparse) (m *Sparse) {
	m = &Sparse{
		r:      self.r,
		c:      self.c,
		matrix: make([]sparsecol, self.c),
	}

	wg := &sync.WaitGroup{}
	for j, col := range self.matrix {
		wg.Add(1)
		go func(j int, col sparsecol) {
			defer func() {
				wg.Done()
			}()
			m.matrix[j] = col.foldsub(b.matrix[j])
		}(j, col)
	}
	wg.Wait()

	return
}

// Multiply two matrices element by element
func (self *Sparse) MulElem(b *Sparse) (m *Sparse) {
	if self.r != b.r || self.c != b.c {
		panic(errors.New("Dimension mismatch"))
	}

	m = &Sparse{
		r:      self.r,
		c:      self.c,
		matrix: make([]sparsecol, self.c),
	}

	wg := &sync.WaitGroup{}
	for j, col := range self.matrix {
		wg.Add(1)
		go func(j int, col sparsecol) {
			defer func() {
				wg.Done()
			}()
			m.matrix[j] = col.foldmul(b.matrix[j])
		}(j, col)
	}
	wg.Wait()

	return
}

// Test for equality of two matrices
func (self *Sparse) Equals(b *Sparse) bool {
	equal := make(chan bool)

	for j, col := range self.matrix {
		go func(j int, col sparsecol) {
			defer func() {
				if r := recover(); r != nil {
					if e, ok := r.(runtime.Error); ok {
						if e.Error() == "runtime error: send on closed channel" {
							return
						}
					}
					panic(r)
				}
			}()
			if col.foldequal(b.matrix[j]) {
				equal <- true
			} else {
				equal <- false
			}
		}(j, col)
	}

	for i := 0; i < len(self.matrix); i++ {
		if !<-equal {
			close(equal)
			return false
		}
	}

	return true
}

// Test for approximate equality of two matrices, tolerance for equality given by error
func (self *Sparse) EqualsApprox(b *Sparse, error float64) bool {
	equal := make(chan bool)

	for j, col := range self.matrix {
		go func(j int, col sparsecol) {
			defer func() {
				if r := recover(); r != nil {
					if e, ok := r.(runtime.Error); ok {
						if e.Error() == "runtime error: send on closed channel" {
							return
						}
					}
					panic(r)
				}
			}()
			if col.foldapprox(b.matrix[j], error) {
				equal <- true
			} else {
				equal <- false
			}
		}(j, col)
	}

	for i := 0; i < len(self.matrix); i++ {
		if !<-equal {
			close(equal)
			return false
		}
	}

	return true
}

// Scale a matrix by a factor
func (self *Sparse) Scalar(f float64) (m *Sparse) {
	m = &Sparse{
		r:      self.r,
		c:      self.c,
		matrix: make([]sparsecol, self.c),
	}

	wg := &sync.WaitGroup{}
	for j, col := range self.matrix {
		wg.Add(1)
		go func(j int, col sparsecol) {
			defer func() {
				wg.Done()
			}()
			m.matrix[j] = col.scale(f)
		}(j, col)
	}
	wg.Wait()

	return
}

// Calculate the sum of a matrix
func (self *Sparse) Sum() (s float64) {
	for _, col := range self.matrix {
		s += col.sum()
	}

	return
}

// Calculate the inner product of two matrices
func (self *Sparse) Inner(b *Sparse) (p float64) {
	if self.r != b.r || self.c != b.c {
		panic(errors.New("Dimension mismatch"))
	}

	for j, col := range self.matrix {
		p += col.foldmul(b.matrix[j]).sum()
	}

	return
}

// Multiply two matrices returning the product - columns calculated concurrently
func (self *Sparse) Dot(b *Sparse) (p *Sparse) {
	switch {
	case self.c != b.r:
		panic(errors.New("Dimension mismatch"))
	}

	p = &Sparse{
		r:      self.r,
		c:      b.c,
		matrix: make([]sparsecol, b.c),
	}

	t := self.T()

	wg := &sync.WaitGroup{}

	for j, col := range b.matrix {
		wg.Add(1)
		go func(j int, col sparsecol) {
			defer func() {
				wg.Done()
			}()
			for i, row := range t.matrix {
				if v := col.foldmul(row).sum(); v != 0 {
					p.matrix[j] = append(p.matrix[j], elem{r: i, value: v})
				}
			}
		}(j, col)
	}

	wg.Wait()
	return
}

// Join a matrix below self returning the new matrix
func (self *Sparse) Stack(b *Sparse) (m *Sparse) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); !ok {
				panic(fmt.Errorf("pkg: %v", r))
			} else {
				panic(err)
			}
		}
	}()

	if self.c != b.c {
		panic(errors.New("Dimension mismatch"))
	}

	m = &Sparse{
		r:      self.r + b.r,
		c:      self.c,
		matrix: make([]sparsecol, self.c),
	}

	wg := &sync.WaitGroup{}
	for j, col := range b.matrix {
		wg.Add(1)
		go func(j int, col sparsecol) {
			defer func() {
				wg.Done()
			}()
			m.matrix[j] = make(sparsecol, len(self.matrix[j]), len(self.matrix[j])+len(col))
			copy(m.matrix[j], self.matrix[j])
			for _, e := range col {
				m.matrix[j] = append(m.matrix[j], elem{r: e.r + self.r, value: e.value})
			}
		}(j, col)
	}
	wg.Wait()

	return
}

// Join a matrix to the right of self returning the new matrix
func (self *Sparse) Augment(b *Sparse) (m *Sparse) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); !ok {
				panic(fmt.Errorf("pkg: %v", r))
			} else {
				panic(err)
			}
		}
	}()

	if self.r != b.r {
		panic(errors.New("Dimension mismatch"))
	}

	m = self.Copy()
	m.matrix = append(m.matrix, b.Copy().matrix...)
	m.c = self.c + b.c

	return
}

// Return a matrix with all elements at (r, c) set to zero where FilterFunc(r, c) returns false
func (self *Sparse) Filter(f matrix.FilterFunc) (m *Sparse) {
	m = &Sparse{
		r:      self.r,
		c:      self.c,
		matrix: make([]sparsecol, len(self.matrix)),
	}

	t := make(sparsecol, 0, len(self.matrix[0]))
	for j, col := range self.matrix {
		for i, e := range col {
			if f(i, j, e.value) {
				t = append(t, e)
			}
		}
		m.matrix[j] = make(sparsecol, len(t))
		copy(m.matrix[j], t)
		t = t[:0]
	}

	return
}

// Apply a function to non-zero elements of the matrix
func (self *Sparse) Apply(f matrix.ApplyFunc) (m *Sparse) {
	m = self.Copy()
	for j, col := range m.matrix {
		for i, e := range col {
			if v := f(i, j, e.value); v != e.value {
				m.matrix[j][i] = elem{r: e.r, value: v}
			}
		}
	}

	return
}

// Apply a function to all elements of the matrix
func (self *Sparse) ApplyAll(f matrix.ApplyFunc) (m *Sparse) {
	m = self.Copy()
	for j := 0; j < m.c; j++ {
		for i := 0; i < m.r; i++ {
			old := m.at(i, j)
			v := f(i, j, old)
			if v != old {
				m.set(i, j, v)
			}
		}
	}

	return
}

// Clean zero elements from a matrix
func (self *Sparse) Clean() (m *Sparse) {
	m = &Sparse{
		r:      self.r,
		c:      self.c,
		matrix: make([]sparsecol, len(self.matrix)),
	}

	t := make(sparsecol, 0, len(self.matrix[0]))
	for j, col := range self.matrix {
		for _, e := range col {
			if e.value != 0 {
				t = append(t, e)
			}
		}
		m.matrix[j] = make(sparsecol, len(t))
		copy(m.matrix[j], t)
		t = t[:0]
	}

	return
}

// Clean elements within error of zero from a matrix
func (self *Sparse) CleanError(error float64) (m *Sparse) {
	m = &Sparse{
		r:      self.r,
		c:      self.c,
		matrix: make([]sparsecol, len(self.matrix)),
	}

	t := make(sparsecol, 0, len(self.matrix[0]))
	for j, col := range self.matrix {
		for _, e := range col {
			if math.Abs(e.value) > error {
				t = append(t, e)
			}
		}
		m.matrix[j] = make(sparsecol, len(t))
		copy(m.matrix[j], t)
		t = t[:0]
	}

	return
}

func (self *Sparse) String() (s string) {
	pc := Margin
	if Margin < 0 {
		var c int
		pc, c = self.Dims()
		if c > pc {
			pc = c
		}
	}
	if self.r > 2*pc || self.c > 2*pc {
		r, c := self.Dims()
		s = fmt.Sprintf("Dims(%v, %v)\n", r, c)
	}
	for i := 0; i < self.r; i++ {
		var l string
		for j := 0; j < self.c; j++ {
			if j >= pc && j < self.c-pc {
				j = self.c - pc - 1
				if i == 0 || i == self.r-1 {
					l += "...  ...  "
				} else {
					l += "          "
				}
				continue
			}

			if j < self.c-1 {
				l += fmt.Sprintf("%-*s", matrix.Precision+matrix.Pad[matrix.Format]+2, strconv.FormatFloat(self.at(i, j), matrix.Format, matrix.Precision, 64))
			} else {
				l += fmt.Sprintf("%-*s", matrix.Precision+matrix.Pad[matrix.Format], strconv.FormatFloat(self.at(i, j), matrix.Format, matrix.Precision, 64))
			}
		}
		switch {
		case self.r == 1:
			s += "[" + l + "]\n"
		case i == 0:
			s += "⎡" + l + "⎤\n"
		case i < self.r-1:
			s += "⎢" + l + "⎥\n"
		default:
			s += "⎣" + l + "⎦\n"
		}

		if i >= pc-1 && i < self.r-pc && 2*pc < self.r {
			i = self.r - pc - 1
			s += " .\n .\n .\n"
			continue
		}
	}

	return
}
