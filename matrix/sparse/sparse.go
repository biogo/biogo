// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Sparse matrix operations
//
// N.B. There is fairly extensive use of unrecovered panics in sparse to avoid unreadable matrix formulae.
package sparse

import (
	"bytes"
	"code.google.com/p/biogo/matrix"
	"errors"
	"fmt"
	"math"
	"math/rand"
)

var (
	Margin      = 3 // Number of columns/rows visible returned by String
	workbuffers chan sparsecol
	BufferLen   = 100
	Buffers     = 10 // Number of allocated work buffers.
)

func init() {
	Init()
}

// Initialise goroutine and memory handling
func Init() {
	workbuffers = make(chan sparsecol, Buffers)
	for i := 0; i < Buffers; i++ {
		buffer := make(sparsecol, BufferLen)
		workbuffers <- buffer
	}
}

// Sparse matrix type
type Sparse struct {
	r, c   int
	matrix []sparsecol
}

func Must(s *Sparse, err error) *Sparse {
	if err != nil {
		panic(err)
	}
	return s
}

// Return a sparse matrix based on a slice of float64 slices
func Matrix(a [][]float64) *Sparse {
	if len(a) == 0 {
		panic(errors.New("zero dimension in matrix definition"))
	}

	maxRowLen := 0
	for _, r := range a {
		if len(r) > maxRowLen {
			maxRowLen = len(r)
		}
	}

	m := &Sparse{
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

	return m
}

// Return the O matrix
func Zero(r, c int) *Sparse {
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
func Identity(size int) *Sparse {
	if size < 1 {
		panic(errors.New("zero dimension in matrix definition"))
	}

	m := &Sparse{
		r:      size,
		c:      size,
		matrix: make([]sparsecol, size),
	}

	for i := 0; i < size; i++ {
		m.matrix[i] = append(m.matrix[i], elem{r: i, value: 1})
	}

	return m
}

type RandFunc func() float64

func Random(r, c int, density float64, fn RandFunc) *Sparse {
	if r < 1 || c < 1 {
		panic(errors.New("zero dimension in matrix definition"))
	}

	m := &Sparse{
		r:      r,
		c:      c,
		matrix: make([]sparsecol, c),
	}

	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			if rand.Float64() < density {
				m.Set(i, j, fn())
			}
		}
	}

	return m
}

// Return of the non-zero elements of a set of matrices in column major order as a column vector.
func Elements(mats ...*Sparse) *Sparse {
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

	e := &Sparse{
		r:      length,
		c:      1,
		matrix: []sparsecol{t},
	}

	return e
}

// Return of the non-zero elements of the matrix in column major order as an array.
func (s *Sparse) Elements() []float64 {
	elements := Elements(s).matrix[0]
	a := make([]float64, 0, len(elements))
	for _, e := range Elements(s).matrix[0] {
		a = append(a, e.value)
	}

	return a
}

// Return a copy of a matrix
func (s *Sparse) Clone() *Sparse {
	m := &Sparse{
		r:      s.r,
		c:      s.c,
		matrix: make([]sparsecol, len(s.matrix)),
	}

	for j, col := range s.matrix {
		m.matrix[j] = make(sparsecol, len(col))
		copy(m.matrix[j], col)
	}

	return m
}

// Return the dimensions of a matrix
func (s *Sparse) Dims() (r, c int) {
	return s.r, s.c
}

// Calculate the determinant of a matrix
func (s *Sparse) Det() float64 {
	panic("not implemented")
}

// Return the minimum non-zero of a matrix
func (s *Sparse) Min() float64 {
	m := math.MaxFloat64
	for _, col := range s.matrix {
		m = math.Min(col.min(), m)
	}

	return m
}

// Return the maximum non-zero of a matrix
func (s *Sparse) Max() float64 {
	m := -math.MaxFloat64
	for _, col := range s.matrix {
		m = math.Max(col.max(), m)
	}

	return m
}

// Set the value at (r, c) to v
func (s *Sparse) Set(r, c int, v float64) {
	if r >= s.r || c >= s.c || r < 0 || c < 0 {
		panic("sparse: index out of bounds")
	}

	s.set(r, c, v)
}

func (s *Sparse) set(r, c int, v float64) {
	col := s.matrix[c]
	lo := 0
	hi := len(col)
	for {
		switch curpos := (lo + hi) / 2; {
		case lo > hi:
			col.insert(lo, elem{r, v})
			s.matrix[c] = col
			return
		case col == nil, r > col[len(col)-1].r:
			col = append(col, elem{r, v})
			s.matrix[c] = col
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
func (s *Sparse) At(r, c int) (v float64) {
	if r >= s.r || c >= s.c || c < 0 || r < 0 {
		panic("sparse: index out of bound")
	}
	return s.at(r, c)
}

func (s *Sparse) at(r, c int) float64 {
	col := s.matrix[c]

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
func (s *Sparse) Norm(ord int) float64 {
	var n float64
	if ord == 0 {
		for _, c := range s.matrix {
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
		sum := s.SumAxis(matrix.Cols)
		for _, e := range sum.matrix[0] {
			n = math.Max(math.Abs(e.value), n)
		}
	case matrix.Inf:
		sum := s.SumAxis(matrix.Rows)
		for _, e := range sum.matrix[0] {
			n = math.Max(math.Abs(e.value), n)
		}
	case -1:
		n = math.MaxFloat64
		sum := s.SumAxis(matrix.Cols)
		for _, e := range sum.matrix[0] {
			n = math.Min(math.Abs(e.value), n)
		}
	case -matrix.Inf:
		n = math.MaxFloat64
		sum := s.SumAxis(matrix.Rows)
		for _, e := range sum.matrix[0] {
			n = math.Min(math.Abs(e.value), n)
		}
	case matrix.Fro:
		for _, c := range s.matrix {
			for _, e := range c {
				n += e.value * e.value
			}
		}
		return math.Sqrt(n)
	default:
		panic("sparse: invalid norm order for matrix")
	}

	return n
}

// Return a column or row vector holding the sums of rows or columns
func (s *Sparse) SumAxis(cols bool) *Sparse {
	m := &Sparse{}
	if cols {
		m.r, m.c, m.matrix = 1, s.c, make([]sparsecol, s.c)
		for i, c := range s.matrix {
			m.matrix[i] = sparsecol{elem{r: 0, value: c.sum()}}
		}
	} else {
		m.r, m.c, m.matrix = s.r, 1, make([]sparsecol, 1)
		data := make([]elem, 0, s.r)
		for i := 0; i < s.r; i++ {
			n := float64(0)
			for j := 0; j < s.c; j++ {
				n += s.at(i, j)
			}
			data = append(data, elem{i, n})
		}
		m.matrix[0] = make([]elem, len(data))
		copy(m.matrix[0], data)
	}

	return m
}

// Return a column or row vector holding the max of rows or columns
func (s *Sparse) MaxAxis(cols bool) *Sparse {
	m := &Sparse{}
	if cols {
		m.r, m.c, m.matrix = 1, s.c, make([]sparsecol, s.c)
		for i, c := range s.matrix {
			m.matrix[i] = sparsecol{elem{r: 0, value: c.max()}}
		}
	} else {
		m.r, m.c, m.matrix = s.r, 1, make([]sparsecol, 1)
		data := make([]elem, 0, s.r)
		for i := 0; i < s.r; i++ {
			n := -math.MaxFloat64
			for j := 0; j < s.c; j++ {
				if v := s.at(i, j); v > n {
					n = v
				}
			}
			data = append(data, elem{i, n})
		}
		m.matrix[0] = make([]elem, len(data))
		copy(m.matrix[0], data)
	}

	return m
}

// Return a column or row vector holding the min of rows or columns
func (s *Sparse) MinAxis(cols bool) *Sparse {
	m := &Sparse{}
	if cols {
		m.r, m.c, m.matrix = 1, s.c, make([]sparsecol, s.c)
		for i, c := range s.matrix {
			m.matrix[i] = sparsecol{elem{r: 0, value: c.min()}}
		}
	} else {
		m.r, m.c, m.matrix = s.r, 1, make([]sparsecol, 1)
		data := make([]elem, 0, s.r)
		for i := 0; i < s.r; i++ {
			n := math.MaxFloat64
			for j := 0; j < s.c; j++ {
				if v := s.at(i, j); v < n {
					n = v
				}
			}
			data = append(data, elem{i, n})
		}
		m.matrix[0] = make([]elem, len(data))
		copy(m.matrix[0], data)
	}

	return m
}

// Return the transpose of a matrix
func (s *Sparse) T() *Sparse {
	var m *Sparse
	if s.r == 0 || s.c == 0 { // this is a vector
		m = s.Clone()
		m.r, m.c = m.c, m.r
		return m
	}

	m = &Sparse{
		r:      s.c,
		c:      s.r,
		matrix: make([]sparsecol, s.r),
	}
	for j, _ := range m.matrix {
		m.matrix[j] = make(sparsecol, 0, m.r)
	}
	for j, col := range s.matrix {
		for _, e := range col {
			m.matrix[e.r] = append(m.matrix[e.r], elem{r: j, value: e.value})
		}
	}

	for j, _ := range m.matrix {
		t := make(sparsecol, len(m.matrix[j]))
		copy(t, m.matrix[j])
		m.matrix[j] = t
	}

	return m
}

// Add one matrix to another
func (s *Sparse) Add(b *Sparse) *Sparse {
	m := &Sparse{
		r:      s.r,
		c:      s.c,
		matrix: make([]sparsecol, s.c),
	}

	for j, col := range s.matrix {
		m.matrix[j] = col.foldadd(b.matrix[j])
	}

	return m
}

// Subtract one matrix from another
func (s *Sparse) Sub(b *Sparse) *Sparse {
	m := &Sparse{
		r:      s.r,
		c:      s.c,
		matrix: make([]sparsecol, s.c),
	}

	for j, col := range s.matrix {
		m.matrix[j] = col.foldsub(b.matrix[j])
	}

	return m
}

// Multiply two matrices element by element
func (s *Sparse) MulElem(b *Sparse) *Sparse {
	if s.r != b.r || s.c != b.c {
		panic("sparse: dimension mismatch")
	}

	m := &Sparse{
		r:      s.r,
		c:      s.c,
		matrix: make([]sparsecol, s.c),
	}

	for j, col := range s.matrix {
		m.matrix[j] = col.foldmul(b.matrix[j])
	}

	return m
}

// Test for equality of two matrices
func (s *Sparse) Equals(b *Sparse) bool {
	for j, col := range s.matrix {
		if !col.foldequal(b.matrix[j]) {
			return false
		}
	}

	return true
}

// Test for approximate equality of two matrices, tolerance for equality given by error
func (s *Sparse) EqualsApprox(b *Sparse, error float64) bool {
	for j, col := range s.matrix {
		if !col.foldapprox(b.matrix[j], error) {
			return false
		}
	}

	return true
}

// Scale a matrix by a factor
func (s *Sparse) Scalar(f float64) *Sparse {
	m := &Sparse{
		r:      s.r,
		c:      s.c,
		matrix: make([]sparsecol, s.c),
	}

	for j, col := range s.matrix {
		m.matrix[j] = col.scale(f)
	}

	return m
}

// Calculate the sum of a matrix
func (s *Sparse) Sum() float64 {
	var sum float64
	for _, col := range s.matrix {
		sum += col.sum()
	}

	return sum
}

// Calculate the inner product of two matrices
func (s *Sparse) Inner(b *Sparse) float64 {
	var p float64
	if s.r != b.r || s.c != b.c {
		panic("sparse: dimension mismatch")
	}

	for j, col := range s.matrix {
		p += col.foldmul(b.matrix[j]).sum()
	}

	return p
}

// Multiply two matrices returning the product.
func (s *Sparse) Dot(b *Sparse) *Sparse {
	switch {
	case s.c != b.r:
		panic("sparse: dimension mismatch")
	}

	p := &Sparse{
		r:      s.r,
		c:      b.c,
		matrix: make([]sparsecol, b.c),
	}

	t := s.T()

	for j, col := range b.matrix {
		for i, row := range t.matrix {
			if v := col.foldmul(row).sum(); v != 0 {
				p.matrix[j] = append(p.matrix[j], elem{r: i, value: v})
			}
		}
	}

	return p
}

// Join a matrix below s returning the new matrix
func (s *Sparse) Stack(b *Sparse) (*Sparse, error) {
	if s.c != b.c {
		return nil, errors.New("sparse: dimension mismatch")
	}

	m := &Sparse{
		r:      s.r + b.r,
		c:      s.c,
		matrix: make([]sparsecol, s.c),
	}

	for j, col := range b.matrix {
		m.matrix[j] = make(sparsecol, len(s.matrix[j]), len(s.matrix[j])+len(col))
		copy(m.matrix[j], s.matrix[j])
		for _, e := range col {
			m.matrix[j] = append(m.matrix[j], elem{r: e.r + s.r, value: e.value})
		}
	}

	return m, nil
}

// Join a matrix to the right of s returning the new matrix
func (s *Sparse) Augment(b *Sparse) (*Sparse, error) {
	if s.r != b.r {
		return nil, errors.New("sparse: dimension mismatch")
	}

	m := s.Clone()
	m.matrix = append(m.matrix, b.Clone().matrix...)
	m.c = s.c + b.c

	return m, nil
}

// Return a matrix with all elements at (r, c) set to zero where FilterFunc(r, c) returns false
func (s *Sparse) Filter(f matrix.FilterFunc) *Sparse {
	m := &Sparse{
		r:      s.r,
		c:      s.c,
		matrix: make([]sparsecol, len(s.matrix)),
	}

	t := make(sparsecol, 0, len(s.matrix[0]))
	for j, col := range s.matrix {
		for i, e := range col {
			if f(i, j, e.value) {
				t = append(t, e)
			}
		}
		m.matrix[j] = make(sparsecol, len(t))
		copy(m.matrix[j], t)
		t = t[:0]
	}

	return m
}

// Apply a function to non-zero elements of the matrix
func (s *Sparse) Apply(f matrix.ApplyFunc) *Sparse {
	m := s.Clone()
	for j, col := range m.matrix {
		for i, e := range col {
			if v := f(i, j, e.value); v != e.value {
				m.matrix[j][i] = elem{r: e.r, value: v}
			}
		}
	}

	return m
}

// Apply a function to all elements of the matrix
func (s *Sparse) ApplyAll(f matrix.ApplyFunc) *Sparse {
	m := s.Clone()
	for j := 0; j < m.c; j++ {
		for i := 0; i < m.r; i++ {
			old := m.at(i, j)
			v := f(i, j, old)
			if v != old {
				m.set(i, j, v)
			}
		}
	}

	return m
}

// Clean zero elements from a matrix
func (s *Sparse) Clean() *Sparse {
	m := &Sparse{
		r:      s.r,
		c:      s.c,
		matrix: make([]sparsecol, len(s.matrix)),
	}

	t := make(sparsecol, 0, len(s.matrix[0]))
	for j, col := range s.matrix {
		for _, e := range col {
			if e.value != 0 {
				t = append(t, e)
			}
		}
		m.matrix[j] = make(sparsecol, len(t))
		copy(m.matrix[j], t)
		t = t[:0]
	}

	return m
}

// Clean elements within epsilon of zero from a matrix
func (s *Sparse) CleanError(epsilon float64) *Sparse {
	m := &Sparse{
		r:      s.r,
		c:      s.c,
		matrix: make([]sparsecol, len(s.matrix)),
	}

	t := make(sparsecol, 0, len(s.matrix[0]))
	for j, col := range s.matrix {
		for _, e := range col {
			if math.Abs(e.value) > epsilon {
				t = append(t, e)
			}
		}
		m.matrix[j] = make(sparsecol, len(t))
		copy(m.matrix[j], t)
		t = t[:0]
	}

	return m
}

func (s *Sparse) String() string {
	pc := Margin
	if Margin < 0 {
		var c int
		pc, c = s.Dims()
		if c > pc {
			pc = c
		}
	}
	b := &bytes.Buffer{}
	if s.r > 2*pc || s.c > 2*pc {
		r, c := s.Dims()
		fmt.Fprintf(b, "Dims(%v, %v)\n", r, c)
	}
	format := fmt.Sprintf("%% -*.*%c", matrix.Format)
	for i := 0; i < s.r; i++ {
		switch {
		case s.r == 1:
			fmt.Fprint(b, "[")
		case i == 0:
			fmt.Fprint(b, "⎡")
		case i < s.r-1:
			fmt.Fprint(b, "⎢")
		default:
			fmt.Fprint(b, "⎣")
		}

		for j := 0; j < s.c; j++ {
			if j >= pc && j < s.c-pc {
				j = s.c - pc - 1
				if i == 0 || i == s.r-1 {
					fmt.Fprint(b, "...  ...  ")
				} else {
					fmt.Fprint(b, "          ")
				}
				continue
			}

			fmt.Fprintf(b, format, matrix.Precision+matrix.Pad[matrix.Format], matrix.Precision, s.at(i, j))
			if j < s.c-1 {
				fmt.Fprint(b, "  ")
			}
		}

		switch {
		case s.r == 1:
			fmt.Fprintln(b, "]")
		case i == 0:
			fmt.Fprintln(b, "⎤")
		case i < s.r-1:
			fmt.Fprintln(b, "⎥")
		default:
			fmt.Fprintln(b, "⎦")
		}

		if i >= pc-1 && i < s.r-pc && 2*pc < s.r {
			i = s.r - pc - 1
			fmt.Fprint(b, " .\n .\n .\n")
			continue
		}
	}

	return b.String()
}
