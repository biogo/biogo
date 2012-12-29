// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package matrix

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
)

var (
	workbuffers chan sparseRow
	BufferLen   = 100
	Buffers     = 10 // Number of allocated work buffers.
)

func init() {
	Init()
}

// Initialise sparse memory handling
func Init() {
	workbuffers = make(chan sparseRow, Buffers)
	for i := 0; i < Buffers; i++ {
		buffer := make(sparseRow, 0, BufferLen)
		workbuffers <- buffer
	}
}

// Sparse matrix type
type Sparse struct {
	Margin     int
	rows, cols int
	matrix     []sparseRow
}

func MustSparse(s *Sparse, err error) *Sparse {
	if err != nil {
		panic(err)
	}
	return s
}

// Return a sparse matrix based on a slice of float64 slices
func NewSparse(a [][]float64) (*Sparse, error) {
	if len(a) == 0 || len(a[0]) == 0 {
		return nil, ErrZeroLength
	}

	m := &Sparse{
		rows:   len(a),
		cols:   len(a[0]),
		matrix: make([]sparseRow, len(a)),
	}

	for i, row := range a {
		if len(row) != m.cols {
			return nil, ErrRowLength
		}
		for j, v := range row {
			if v != 0 {
				m.matrix[i] = append(m.matrix[i], sparseElem{index: j, value: v})
			}
		}
	}

	return m, nil
}

// Return the O matrix
func ZeroSparse(r, c int) (*Sparse, error) {
	if r < 1 || c < 1 {
		return nil, ErrZeroLength
	}

	return &Sparse{
		rows:   r,
		cols:   c,
		matrix: make([]sparseRow, r),
	}, nil
}

// Return the I matrix
func IdentitySparse(size int) (*Sparse, error) {
	if size < 1 {
		return nil, ErrZeroLength
	}

	m := &Sparse{
		rows:   size,
		cols:   size,
		matrix: make([]sparseRow, size),
	}

	for i := 0; i < size; i++ {
		m.matrix[i] = sparseRow{sparseElem{index: i, value: 1}}
	}

	return m, nil
}

type FloatFunc func() float64

func FuncSparse(r, c int, density float64, fn FloatFunc) (*Sparse, error) {
	if r < 1 || c < 1 {
		return nil, ErrZeroLength
	}

	m := &Sparse{
		rows:   r,
		cols:   c,
		matrix: make([]sparseRow, r),
	}

	for i := range m.matrix {
		for j := 0; j < c; j++ {
			if rand.Float64() < density {
				m.Set(i, j, fn())
			}
		}
	}

	return m, nil
}

func ElementsVector(mats ...Matrix) []float64 {
	var length int
	for _, m := range mats {
		switch m := m.(type) {
		case *Dense:
			length += len(m.matrix)
		case *Sparse:
			for _, row := range m.matrix {
				length += len(row)
			}
		}
	}

	v := make([]float64, 0, length)
	for _, m := range mats {
		switch m := m.(type) {
		case *Dense:
			v = append(v, m.matrix...)
		case *Sparse:
			for _, row := range m.matrix {
				for _, e := range row {
					if e.value != 0 {
						v = append(v, e.value)
					}
				}
			}
		case Matrix:
			rows, cols := m.Dims()
			for r := 0; r < rows; r++ {
				for c := 0; c < cols; c++ {
					v = append(v, m.At(r, c))
				}
			}
		}
	}

	return v
}

// Return of the non-zero elements of a set of matrices in row major order as a row vector.
func ElementsSparse(mats ...Matrix) *Sparse {
	var length int
	for _, m := range mats {
		switch m := m.(type) {
		case *Sparse:
			for _, row := range m.matrix {
				length += len(row)
			}
		}
	}

	t := make(sparseRow, 0, length)
	for _, m := range mats {
		switch m := m.(type) {
		case *Sparse:
			for _, row := range m.matrix {
				for _, e := range row {
					if e.value != 0 {
						t = append(t, sparseElem{index: len(t), value: e.value})
					}
				}
			}
		case Matrix:
			rows, cols := m.Dims()
			for r := 0; r < rows; r++ {
				for c := 0; c < cols; c++ {
					if val := m.At(r, c); val != 0 {
						t = append(t, sparseElem{index: len(t), value: val})
					}
				}
			}
		}
	}

	e := &Sparse{
		rows:   1,
		cols:   length,
		matrix: []sparseRow{t},
	}

	return e
}

// Return of the non-zero elements of the matrix in column major order as a Slice.
func (s *Sparse) ElementsVector() []float64 {
	var length int
	for _, row := range s.matrix {
		length += len(row)
	}

	v := make([]float64, 0, length)
	for _, row := range s.matrix {
		for _, e := range row {
			if e.value != 0 {
				v = append(v, e.value)
			}
		}
	}

	return v
}

func (s *Sparse) Clone() Matrix { return s.CloneSparse() }

// Return a copy of a matrix
func (s *Sparse) CloneSparse() *Sparse {
	m := &Sparse{
		rows:   s.rows,
		cols:   s.cols,
		matrix: make([]sparseRow, len(s.matrix)),
	}

	for j, row := range s.matrix {
		m.matrix[j] = make(sparseRow, len(row))
		copy(m.matrix[j], row)
	}

	return m
}

// Return the dimensions of a matrix
func (s *Sparse) Dims() (r, c int) {
	return s.rows, s.cols
}

// Calculate the determinant of a matrix
func (s *Sparse) Det() float64 {
	panic("not implemented")
}

// Return the minimum of a matrix
func (s *Sparse) Min() float64 {
	m := math.MaxFloat64
	for _, row := range s.matrix {
		m = math.Min(row.min(), m)
		if len(row) < s.cols {
			m = math.Min(m, 0)
		}
	}

	return m
}

// Return the maximum of a matrix
func (s *Sparse) Max() float64 {
	m := -math.MaxFloat64
	for _, row := range s.matrix {
		m = math.Max(row.max(), m)
		if len(row) < s.cols {
			m = math.Max(m, 0)
		}
	}

	return m
}

// Return the minimum non-zero of a matrix
func (s *Sparse) MinNonZero() float64 {
	m := math.MaxFloat64
	var ok bool
	for _, row := range s.matrix {
		r, nz := row.minNonZero()
		if nz {
			m = math.Min(r, m)
			ok = true
		}
	}
	if !ok {
		return 0
	}
	return m
}

// Return the maximum non-zero of a matrix
func (s *Sparse) MaxNonZero() float64 {
	m := -math.MaxFloat64
	var ok bool
	for _, row := range s.matrix {
		r, nz := row.maxNonZero()
		if nz {
			m = math.Max(r, m)
			ok = true
		}
	}
	if !ok {
		return 0
	}
	return m
}

// Set the value at (r, c) to v
func (s *Sparse) Set(r, c int, v float64) {
	if r >= s.rows || c >= s.cols || r < 0 || c < 0 {
		panic(ErrIndexOutOfBounds)
	}

	s.set(r, c, v)
}

func (s *Sparse) set(r, c int, v float64) {
	row := s.matrix[r]
	lo := 0
	hi := len(row)
	for {
		switch curpos := (lo + hi) / 2; {
		case lo > hi:
			row.insert(lo, sparseElem{index: c, value: v})
			s.matrix[r] = row
			return
		case row == nil, c > row[len(row)-1].index:
			row = append(row, sparseElem{index: c, value: v})
			s.matrix[r] = row
			return
		case row[curpos].index == c:
			row[curpos].value = v
			return
		case c < row[curpos].index:
			hi = curpos - 1
		case c > row[curpos].index:
			lo = curpos + 1
		}
	}
}

// Return the value at (r, c)
func (s *Sparse) At(r, c int) (v float64) {
	if r >= s.rows || c >= s.cols || c < 0 || r < 0 {
		panic(ErrIndexOutOfBounds)
	}
	return s.matrix[r].at(c)
}

func (s *Sparse) Trace() float64 {
	if s.rows != s.cols {
		panic(ErrSquare)
	}
	var t float64
	for i, row := range s.matrix {
		t += row.at(i)
	}
	return t
}

// Determin a variety of norms
func (s *Sparse) Norm(ord int) float64 {
	var n float64
	if ord == 0 {
		for _, row := range s.matrix {
			for _, e := range row {
				n += e.value * e.value
			}
		}
		return math.Sqrt(n)
	}
	switch ord {
	case 2, -2:
		panic("not implemented - feel free to port an svd function to matrix")
	case 1:
		sum := s.SumAxis(Cols)
		for _, e := range sum.matrix[0] {
			n = math.Max(math.Abs(e.value), n)
		}
	case Inf:
		sum := s.SumAxis(Rows)
		for _, e := range sum.matrix[0] {
			n = math.Max(math.Abs(e.value), n)
		}
	case -1:
		n = math.MaxFloat64
		sum := s.SumAxis(Cols)
		for _, e := range sum.matrix[0] {
			n = math.Min(math.Abs(e.value), n)
		}
	case -Inf:
		n = math.MaxFloat64
		sum := s.SumAxis(Rows)
		for _, e := range sum.matrix[0] {
			n = math.Min(math.Abs(e.value), n)
		}
	case Fro:
		for _, row := range s.matrix {
			for _, e := range row {
				n += e.value * e.value
			}
		}
		return math.Sqrt(n)
	default:
		panic(ErrNormOrder)
	}

	return n
}

// Return a column or row vector holding the sums of rows or columns
func (s *Sparse) SumAxis(cols bool) *Sparse {
	m := &Sparse{}
	if !cols {
		m.rows, m.cols, m.matrix = s.rows, 1, make([]sparseRow, s.rows)
		for i, row := range s.matrix {
			m.matrix[i] = sparseRow{sparseElem{index: 0, value: row.sum()}}
		}
	} else {
		m.rows, m.cols, m.matrix = 1, s.cols, make([]sparseRow, 1)
		data := make([]sparseElem, 0, s.cols)
		for i := 0; i < s.cols; i++ {
			var n float64
			for _, row := range s.matrix {
				n += row.at(i)
			}
			data = append(data, sparseElem{index: i, value: n})
		}
		m.matrix[0] = make([]sparseElem, len(data))
		copy(m.matrix[0], data)
	}

	return m
}

// Return a column or row vector holding the max of rows or columns
func (s *Sparse) MaxAxis(cols bool) *Sparse {
	m := &Sparse{}
	if !cols {
		m.rows, m.cols, m.matrix = s.rows, 1, make([]sparseRow, s.rows)
		for i, row := range s.matrix {
			m.matrix[i] = sparseRow{sparseElem{index: 0, value: row.max()}}
		}
	} else {
		m.rows, m.cols, m.matrix = 1, s.cols, make([]sparseRow, 1)
		data := make([]sparseElem, 0, s.cols)
		for i := 0; i < s.cols; i++ {
			n := -math.MaxFloat64
			for _, row := range s.matrix {
				n = math.Max(row.at(i), n)
			}
			data = append(data, sparseElem{index: i, value: n})
		}
		m.matrix[0] = make([]sparseElem, len(data))
		copy(m.matrix[0], data)
	}

	return m
}

// Return a column or row vector holding the min of rows or columns
func (s *Sparse) MinAxis(cols bool) *Sparse {
	m := &Sparse{}
	if !cols {
		m.rows, m.cols, m.matrix = s.rows, 1, make([]sparseRow, s.rows)
		for i, row := range s.matrix {
			m.matrix[i] = sparseRow{sparseElem{index: 0, value: row.min()}}
		}
	} else {
		m.rows, m.cols, m.matrix = 1, s.cols, make([]sparseRow, 1)
		data := make([]sparseElem, 0, s.cols)
		for i := 0; i < s.cols; i++ {
			n := math.MaxFloat64
			for _, row := range s.matrix {
				n = math.Min(row.at(i), n)
			}
			data = append(data, sparseElem{index: i, value: n})
		}
		m.matrix[0] = make([]sparseElem, len(data))
		copy(m.matrix[0], data)
	}

	return m
}

func (s *Sparse) U() Matrix { return s.USparse() }

// Return the transpose of a matrix
func (s *Sparse) USparse() *Sparse {
	if s.rows != s.cols {
		panic(ErrSquare)
	}
	m := &Sparse{
		rows:   s.rows,
		cols:   s.cols,
		matrix: make([]sparseRow, s.rows),
	}
	for i, row := range s.matrix {
		for j, e := range row {
			if e.index >= i {
				m.matrix[i] = append(m.matrix[i], row[j:]...)
				break
			}
		}
	}
	return m
}

func (s *Sparse) L() Matrix { return s.LSparse() }

// Return the lower triangular matrix
func (s *Sparse) LSparse() *Sparse {
	if s.rows != s.cols {
		panic(ErrSquare)
	}
	m := &Sparse{
		rows:   s.rows,
		cols:   s.cols,
		matrix: make([]sparseRow, s.rows),
	}
	for i, row := range s.matrix {
		for j := len(row) - 1; j >= 0; j-- {
			if row[j].index <= i {
				m.matrix[i] = append(m.matrix[i], row[:j+1]...)
				break
			}
		}
	}
	return m
}

func (s *Sparse) T() Matrix { return s.TSparse() }

// Return the upper triangular matrix
func (s *Sparse) TSparse() *Sparse {
	var m *Sparse
	if s.rows == 0 || s.cols == 0 { // this is a vector
		m = s.CloneSparse()
		m.rows, m.cols = m.cols, m.rows
		return m
	}

	m = &Sparse{
		rows:   s.cols,
		cols:   s.rows,
		matrix: make([]sparseRow, s.cols),
	}
	for j, _ := range m.matrix {
		m.matrix[j] = make(sparseRow, 0, m.rows)
	}
	for j, row := range s.matrix {
		for _, e := range row {
			m.matrix[e.index] = append(m.matrix[e.index], sparseElem{index: j, value: e.value})
		}
	}

	for j, _ := range m.matrix {
		t := make(sparseRow, len(m.matrix[j]))
		copy(t, m.matrix[j])
		m.matrix[j] = t
	}

	return m
}

func (s *Sparse) Add(b Matrix) Matrix {
	switch b := b.(type) {
	case *Sparse:
		return s.AddSparse(b)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// Add one matrix to another
func (s *Sparse) AddSparse(b *Sparse) *Sparse {
	if s.rows != b.rows || s.cols != b.cols {
		panic(ErrShape)
	}

	m := &Sparse{
		rows:   s.rows,
		cols:   s.cols,
		matrix: make([]sparseRow, s.rows),
	}

	for j, row := range s.matrix {
		m.matrix[j] = row.foldAdd(b.matrix[j])
	}

	return m
}

func (s *Sparse) Sub(b Matrix) Matrix {
	switch b := b.(type) {
	case *Sparse:
		return s.SubSparse(b)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// Subtract one matrix from another
func (s *Sparse) SubSparse(b *Sparse) *Sparse {
	if s.rows != b.rows || s.cols != b.cols {
		panic(ErrShape)
	}

	m := &Sparse{
		rows:   s.rows,
		cols:   s.cols,
		matrix: make([]sparseRow, s.rows),
	}

	for j, row := range s.matrix {
		m.matrix[j] = row.foldSub(b.matrix[j])
	}

	return m
}

func (s *Sparse) MulElem(b Matrix) Matrix {
	switch b := b.(type) {
	case *Sparse:
		return s.MulElemSparse(b)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// Multiply two matrices element by element
func (s *Sparse) MulElemSparse(b *Sparse) *Sparse {
	if s.rows != b.rows || s.cols != b.cols {
		panic(ErrShape)
	}

	m := &Sparse{
		rows:   s.rows,
		cols:   s.cols,
		matrix: make([]sparseRow, s.rows),
	}

	for j, row := range s.matrix {
		m.matrix[j] = row.foldMul(b.matrix[j])
	}

	return m
}

func (s *Sparse) Equals(b Matrix) bool {
	switch b := b.(type) {
	case *Sparse:
		return s.EqualsSparse(b)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// Test for equality of two matrices
func (s *Sparse) EqualsSparse(b *Sparse) bool {
	if s.rows != b.rows || s.cols != b.cols {
		return false
	}

	for j, row := range s.matrix {
		if !row.foldEqual(b.matrix[j]) {
			return false
		}
	}

	return true
}

func (s *Sparse) EqualsApprox(b Matrix, epsilon float64) bool {
	switch b := b.(type) {
	case *Sparse:
		return s.EqualsApproxSparse(b, epsilon)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// Test for approximate equality of two matrices, tolerance for equality given by error
func (s *Sparse) EqualsApproxSparse(b *Sparse, epsilon float64) bool {
	if s.rows != b.rows || s.cols != b.cols {
		return false
	}

	for j, row := range s.matrix {
		if !row.foldApprox(b.matrix[j], epsilon) {
			return false
		}
	}

	return true
}

func (s *Sparse) Scalar(f float64) Matrix { return s.ScalarSparse(f) }

// Scale a matrix by a factor
func (s *Sparse) ScalarSparse(f float64) *Sparse {
	m := &Sparse{
		rows:   s.rows,
		cols:   s.cols,
		matrix: make([]sparseRow, s.rows),
	}

	for j, row := range s.matrix {
		m.matrix[j] = row.scale(f)
	}

	return m
}

// Calculate the sum of a matrix
func (s *Sparse) Sum() float64 {
	var sum float64
	for _, row := range s.matrix {
		sum += row.sum()
	}

	return sum
}

func (s *Sparse) Inner(b Matrix) float64 {
	switch b := b.(type) {
	case *Sparse:
		return s.InnerSparse(b)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// Calculate the inner product of two matrices
func (s *Sparse) InnerSparse(b *Sparse) float64 {
	var p float64
	if s.rows != b.rows || s.cols != b.cols {
		panic(ErrShape)
	}

	for j, row := range s.matrix {
		p += row.foldMulSum(b.matrix[j])
	}

	return p
}

func (s *Sparse) Dot(b Matrix) Matrix {
	switch b := b.(type) {
	case *Sparse:
		return s.DotSparse(b)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// Multiply two matrices returning the product.
func (s *Sparse) DotSparse(b *Sparse) *Sparse {
	if s.cols != b.rows {
		panic(ErrShape)
	}

	p := &Sparse{
		rows:   s.rows,
		cols:   b.cols,
		matrix: make([]sparseRow, s.rows),
	}

	t := <-workbuffers
	for i := 0; i < b.cols; i++ {
		for j := 0; j < b.rows; j++ {
			if v := b.matrix[j].at(i); v != 0 {
				t = append(t, sparseElem{index: j, value: v})
			}
		}
		for j, row := range s.matrix {
			if v := row.foldMulSum(t); v != 0 {
				p.matrix[j] = append(p.matrix[j], sparseElem{index: i, value: v})
			}
		}
		t = t[:0]
	}
	workbuffers <- t

	return p
}

func (s *Sparse) Augment(b Matrix) Matrix {
	switch b := b.(type) {
	case *Sparse:
		return s.AugmentSparse(b)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// Join a matrix to the right of s returning the new matrix
func (s *Sparse) AugmentSparse(b *Sparse) *Sparse {
	if s.rows != b.rows {
		panic(ErrColLength)
	}

	m := &Sparse{
		rows:   s.rows,
		cols:   s.cols + b.cols,
		matrix: make([]sparseRow, s.rows),
	}

	for j, row := range b.matrix {
		m.matrix[j] = make(sparseRow, len(s.matrix[j]), len(s.matrix[j])+len(row))
		copy(m.matrix[j], s.matrix[j])
		for _, e := range row {
			m.matrix[j] = append(m.matrix[j], sparseElem{index: e.index + s.cols, value: e.value})
		}
	}

	return m
}

func (s *Sparse) Stack(b Matrix) Matrix {
	switch b := b.(type) {
	case *Sparse:
		return s.StackSparse(b)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// Join a matrix below s returning the new matrix
func (s *Sparse) StackSparse(b *Sparse) *Sparse {
	if s.cols != b.cols {
		panic(ErrRowLength)
	}

	m := &Sparse{
		rows:   s.rows + b.rows,
		cols:   s.cols,
		matrix: make([]sparseRow, len(s.matrix)+len(b.matrix)),
	}
	copy(m.matrix, s.CloneSparse().matrix)
	copy(m.matrix[len(s.matrix):], b.CloneSparse().matrix)

	return m
}

func (s *Sparse) Filter(f FilterFunc) Matrix { return s.FilterSparse(f) }

// Return a matrix with all elements at (r, c) set to zero where FilterFunc(r, c) returns false
func (s *Sparse) FilterSparse(f FilterFunc) *Sparse {
	m := &Sparse{
		rows:   s.rows,
		cols:   s.cols,
		matrix: make([]sparseRow, len(s.matrix)),
	}

	t := make(sparseRow, 0, len(s.matrix[0]))
	for j, row := range s.matrix {
		for i, e := range row {
			if f(i, j, e.value) {
				t = append(t, e)
			}
		}
		m.matrix[j] = make(sparseRow, len(t))
		copy(m.matrix[j], t)
		t = t[:0]
	}

	return m
}

func (s *Sparse) Apply(f ApplyFunc) Matrix { return s.ApplySparse(f) }

// Apply a function to non-zero elements of the matrix
func (s *Sparse) ApplySparse(f ApplyFunc) *Sparse {
	m := s.CloneSparse()
	for j, row := range m.matrix {
		for i, e := range row {
			if v := f(i, j, e.value); v != e.value {
				m.matrix[j][i] = sparseElem{index: e.index, value: v}
			}
		}
	}

	return m
}

func (s *Sparse) ApplyAll(f ApplyFunc) Matrix { return s.ApplyAllSparse(f) }

// Apply a function to all elements of the matrix
func (s *Sparse) ApplyAllSparse(f ApplyFunc) *Sparse {
	m := s.CloneSparse()
	for i, row := range s.matrix {
		for j := 0; j < m.cols; j++ {
			old := row.at(j)
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
		rows:   s.rows,
		cols:   s.cols,
		matrix: make([]sparseRow, len(s.matrix)),
	}

	t := make(sparseRow, 0, len(s.matrix[0]))
	for j, row := range s.matrix {
		for _, e := range row {
			if e.value != 0 {
				t = append(t, e)
			}
		}
		m.matrix[j] = make(sparseRow, len(t))
		copy(m.matrix[j], t)
		t = t[:0]
	}

	return m
}

// Clean elements within epsilon of zero from a matrix
func (s *Sparse) CleanError(epsilon float64) *Sparse {
	m := &Sparse{
		rows:   s.rows,
		cols:   s.cols,
		matrix: make([]sparseRow, len(s.matrix)),
	}

	t := make(sparseRow, 0, len(s.matrix[0]))
	for j, row := range s.matrix {
		for _, e := range row {
			if math.Abs(e.value) > epsilon {
				t = append(t, e)
			}
		}
		m.matrix[j] = make(sparseRow, len(t))
		copy(m.matrix[j], t)
		t = t[:0]
	}

	return m
}

func (s *Sparse) Format(fs fmt.State, c rune) {
	if c == 'v' && fs.Flag('#') {
		fmt.Fprintf(fs, "&%#v", *s)
		return
	}
	Format(s, s.Margin, '.', fs, c)
}

func (s *Sparse) String() string {
	b := &bytes.Buffer{}
	fmt.Fprintf(b, "%6.4e", s)
	return b.String()
}
