// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package matrix

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"unsafe"
)

// Dense represent a dense matrix type.
type Dense struct {
	Margin     int
	rows, cols int
	matrix     denseRow
}

func MustDense(d *Dense, err error) *Dense {
	if err != nil {
		panic(err)
	}
	return d
}

// Return a sparse matrix based on a slice of float64 slices
func NewDense(a [][]float64) (*Dense, error) {
	if len(a) == 0 || len(a[0]) == 0 {
		return nil, ErrZeroLength
	}

	m := Dense{
		rows: len(a),
		cols: len(a[0]),
	}
	for _, row := range a {
		if len(row) != m.cols {
			return nil, ErrRowLength
		}
	}
	m.matrix = make(denseRow, len(a)*len(a[0]))

	for i, row := range a {
		copy(m.matrix[i*m.cols:(i+1)*m.cols], row)
	}

	return &m, nil
}

// Return the O matrix
func ZeroDense(r, c int) (*Dense, error) {
	if r < 1 || c < 1 {
		return nil, ErrZeroLength
	}

	return &Dense{
		rows:   r,
		cols:   c,
		matrix: make(denseRow, r*c),
	}, nil
}

// Return the I matrix
func IdentityDense(size int) (*Dense, error) {
	if size < 1 {
		return nil, ErrZeroLength
	}

	m := &Dense{
		rows:   size,
		cols:   size,
		matrix: make(denseRow, size*size),
	}

	for i := 0; i < size; i++ {
		m.matrix[i*size+i] = 1
	}

	return m, nil
}

func FuncDense(r, c int, density float64, fn FloatFunc) (*Dense, error) {
	if r < 1 || c < 1 {
		return nil, ErrZeroLength
	}

	m := &Dense{
		rows:   r,
		cols:   c,
		matrix: make(denseRow, r*c),
	}

	for i := range m.matrix {
		if rand.Float64() < density {
			m.matrix[i] = fn()
		}
	}

	return m, nil
}

// Return of the elements of a set of matrices in row major order as a row vector.
func ElementsDense(mats ...Matrix) *Dense {
	var length int
	for _, m := range mats {
		switch m := m.(type) {
		case *Dense:
			length += len(m.matrix)
		}
	}

	t := make(denseRow, 0, length)
	for _, m := range mats {
		switch m := m.(type) {
		case *Dense:
			t = append(t, m.matrix...)
		case Matrix:
			rows, cols := m.Dims()
			for r := 0; r < rows; r++ {
				for c := 0; c < cols; c++ {
					t = append(t, m.At(r, c))
				}
			}
		}
	}

	e := &Dense{
		rows:   1,
		cols:   length,
		matrix: t,
	}

	return e
}

// Return of the elements of the matrix in column major order as a Slice.
func (d *Dense) ElementsVector() []float64 {
	return append([]float64(nil), *(*[]float64)(unsafe.Pointer(&d.matrix))...)
}

func (d *Dense) Clone() Matrix { return d.CloneDense() }

// Return a copy of a matrix
func (d *Dense) CloneDense() *Dense {
	return &Dense{
		rows:   d.rows,
		cols:   d.cols,
		matrix: append(denseRow(nil), d.matrix...),
	}
}

// Return the dimensions of a matrix
func (d *Dense) Dims() (r, c int) {
	return d.rows, d.cols
}

// Calculate the determinant of a matrix
func (d *Dense) Det() float64 {
	panic("not implemented")
}

// Return the minimum non-zero of a matrix
func (d *Dense) Min() float64 {
	return d.matrix.min()
}

// Return the maximum non-zero of a matrix
func (d *Dense) Max() float64 {
	return d.matrix.max()
}

// Set the value at (r, c) to v
func (d *Dense) Set(r, c int, v float64) {
	if r >= d.rows || c >= d.cols || r < 0 || c < 0 {
		panic(ErrIndexOutOfBounds)
	}

	d.set(r, c, v)
}

func (d *Dense) set(r, c int, v float64) {
	d.matrix[r*d.cols+c] = v
}

// Return the value at (r, c)
func (d *Dense) At(r, c int) (v float64) {
	if r >= d.rows || c >= d.cols || c < 0 || r < 0 {
		panic(ErrIndexOutOfBounds)
	}
	return d.at(r, c)
}

func (d *Dense) at(r, c int) float64 {
	return d.matrix[r*d.cols+c]
}

func (d *Dense) Trace() float64 {
	if d.rows != d.cols {
		panic(ErrSquare)
	}
	var t float64
	for i := 0; i < len(d.matrix); i += d.cols + 1 {
		t += d.matrix[i]
	}
	return t
}

// Determin a variety of norms
func (d *Dense) Norm(ord int) float64 {
	var n float64
	if ord == 0 {
		for _, e := range d.matrix {
			n += e * e
		}
		return math.Sqrt(n)
	}
	switch ord {
	case 2, -2:
		panic("not implemented - feel free to port an svd function to matrix")
	case 1:
		sum := d.SumAxis(Cols)
		for _, e := range sum.matrix {
			n = math.Max(math.Abs(e), n)
		}
	case Inf:
		sum := d.SumAxis(Rows)
		for _, e := range sum.matrix {
			n = math.Max(math.Abs(e), n)
		}
	case -1:
		n = math.MaxFloat64
		sum := d.SumAxis(Cols)
		for _, e := range sum.matrix {
			n = math.Min(math.Abs(e), n)
		}
	case -Inf:
		n = math.MaxFloat64
		sum := d.SumAxis(Rows)
		for _, e := range sum.matrix {
			n = math.Min(math.Abs(e), n)
		}
	case Fro:
		for _, e := range d.matrix {
			n += e * e
		}
		return math.Sqrt(n)
	default:
		panic(ErrNormOrder)
	}

	return n
}

// Return a column or row vector holding the sums of rows or columns
func (d *Dense) SumAxis(cols bool) *Dense {
	m := &Dense{}
	if !cols {
		m.rows, m.cols, m.matrix = d.rows, 1, make(denseRow, d.rows)
		for i := 0; i < d.rows; i++ {
			row := d.matrix[i*d.cols : (i+1)*d.cols]
			m.matrix[i] = row.sum()
		}
	} else {
		m.rows, m.cols, m.matrix = 1, d.cols, make(denseRow, d.cols)
		for i := 0; i < d.cols; i++ {
			var n float64
			for j := 0; j < d.rows; j++ {
				n += d.at(j, i)
			}
			m.matrix[i] = n
		}
	}

	return m
}

// Return a column or row vector holding the max of rows or columns
func (d *Dense) MaxAxis(cols bool) *Dense {
	m := &Dense{}
	if !cols {
		m.rows, m.cols, m.matrix = d.rows, 1, make(denseRow, d.rows)
		for i := 0; i < d.rows; i++ {
			row := d.matrix[i*d.cols : (i+1)*d.cols]
			m.matrix[i] = row.max()
		}
	} else {
		m.rows, m.cols, m.matrix = 1, d.cols, make(denseRow, d.cols)
		for i := 0; i < d.cols; i++ {
			var n float64
			for j := 0; j < d.rows; j++ {
				n = math.Max(d.at(j, i), n)
			}
			m.matrix[i] = n
		}
	}

	return m
}

// Return a column or row vector holding the min of rows or columns
func (d *Dense) MinAxis(cols bool) *Dense {
	m := &Dense{}
	if !cols {
		m.rows, m.cols, m.matrix = d.rows, 1, make(denseRow, d.rows)
		for i := 0; i < d.rows; i++ {
			row := d.matrix[i*d.cols : (i+1)*d.cols]
			m.matrix[i] = row.min()
		}
	} else {
		m.rows, m.cols, m.matrix = 1, d.cols, make(denseRow, d.cols)
		for i := 0; i < d.cols; i++ {
			var n = math.MaxFloat64
			for j := 0; j < d.rows; j++ {
				n = math.Min(d.at(j, i), n)
			}
			m.matrix[i] = n
		}
	}

	return m
}

func (d *Dense) U() Matrix { return d.UDense() }

// Return the transpose of a matrix
func (d *Dense) UDense() *Dense {
	if d.rows != d.cols {
		panic(ErrSquare)
	}
	m := &Dense{
		rows:   d.rows,
		cols:   d.cols,
		matrix: make(denseRow, len(d.matrix)),
	}
	for i := 0; i < d.rows; i++ {
		copy(m.matrix[i*d.cols+i:(i+1)*d.cols], d.matrix[i*d.cols+i:(i+1)*d.cols])
	}
	return m
}

func (d *Dense) L() Matrix { return d.LDense() }

// Return the lower triangular matrix
func (d *Dense) LDense() *Dense {
	if d.rows != d.cols {
		panic(ErrSquare)
	}
	m := &Dense{
		rows:   d.rows,
		cols:   d.cols,
		matrix: make(denseRow, len(d.matrix)),
	}
	for i := 0; i < d.rows; i++ {
		copy(m.matrix[i*d.cols:i*d.cols+i+1], d.matrix[i*d.cols:i*d.cols+i+1])
	}
	return m
}

func (d *Dense) T() Matrix { return d.TDense() }

// Return the upper triangular matrix
func (d *Dense) TDense() *Dense {
	var m *Dense
	if d.rows == 0 || d.cols == 0 { // this is a vector
		m = d.CloneDense()
		m.rows, m.cols = m.cols, m.rows
		return m
	}

	m = &Dense{
		rows:   d.cols,
		cols:   d.rows,
		matrix: make(denseRow, len(d.matrix)),
	}
	for i := 0; i < d.cols; i++ {
		for j := 0; j < d.rows; j++ {
			m.set(i, j, d.at(j, i))
		}
	}

	return m
}

func (d *Dense) Add(b Matrix) Matrix {
	switch b := b.(type) {
	case *Dense:
		return d.AddDense(b)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// Add one matrix to another
func (d *Dense) AddDense(b *Dense) *Dense {
	if d.rows != b.rows || d.cols != b.cols {
		panic(ErrShape)
	}

	return &Dense{
		rows:   d.rows,
		cols:   d.cols,
		matrix: d.matrix.foldAdd(b.matrix),
	}
}

func (d *Dense) Sub(b Matrix) Matrix {
	switch b := b.(type) {
	case *Dense:
		return d.SubDense(b)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// Subtract one matrix from another
func (d *Dense) SubDense(b *Dense) *Dense {
	if d.rows != b.rows || d.cols != b.cols {
		panic(ErrShape)
	}

	return &Dense{
		rows:   d.rows,
		cols:   d.cols,
		matrix: d.matrix.foldSub(b.matrix),
	}
}

func (d *Dense) MulElem(b Matrix) Matrix {
	switch b := b.(type) {
	case *Dense:
		return d.MulElemDense(b)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// Multiply two matrices element by element
func (d *Dense) MulElemDense(b *Dense) *Dense {
	if d.rows != b.rows || d.cols != b.cols {
		panic(ErrShape)
	}

	return &Dense{
		rows:   d.rows,
		cols:   d.cols,
		matrix: d.matrix.foldMul(b.matrix),
	}
}

func (d *Dense) Equals(b Matrix) bool {
	switch b := b.(type) {
	case *Dense:
		return d.EqualsDense(b)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// Test for equality of two matrices
func (d *Dense) EqualsDense(b *Dense) bool {
	if d.rows != b.rows || d.cols != b.cols {
		return false
	}
	return d.matrix.foldEqual(b.matrix)
}

func (d *Dense) EqualsApprox(b Matrix, epsilon float64) bool {
	switch b := b.(type) {
	case *Dense:
		return d.EqualsApproxDense(b, epsilon)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// Test for approximate equality of two matrices, tolerance for equality given by error
func (d *Dense) EqualsApproxDense(b *Dense, epsilon float64) bool {
	if d.rows != b.rows || d.cols != b.cols {
		return false
	}
	return d.matrix.foldApprox(b.matrix, epsilon)
}

func (d *Dense) Scalar(f float64) Matrix { return d.ScalarSparse(f) }

// Scale a matrix by a factor
func (d *Dense) ScalarSparse(f float64) *Dense {
	return &Dense{
		rows:   d.rows,
		cols:   d.cols,
		matrix: d.matrix.scale(f),
	}
}

// Calculate the sum of a matrix
func (d *Dense) Sum() float64 {
	return d.matrix.sum()
}

func (d *Dense) Inner(b Matrix) float64 {
	switch b := b.(type) {
	case *Dense:
		return d.InnerDense(b)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// Calculate the inner product of two matrices
func (d *Dense) InnerDense(b *Dense) float64 {
	if d.rows != b.rows || d.cols != b.cols {
		panic(ErrShape)
	}
	return d.matrix.foldMulSum(b.matrix)
}

func (d *Dense) Dot(b Matrix) Matrix {
	switch b := b.(type) {
	case *Dense:
		return d.DotDense(b)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// Multiply two matrices returning the product.
func (d *Dense) DotDense(b *Dense) *Dense {
	if d.cols != b.rows {
		panic(ErrShape)
	}

	p := &Dense{
		rows:   d.rows,
		cols:   b.cols,
		matrix: make(denseRow, d.rows*b.cols),
	}

	var t []float64
	for i := 0; i < b.cols; i++ {
		for j := 0; j < b.rows; j++ {
			t = append(t, b.at(j, i))
		}
		for j := 0; j < d.rows; j++ {
			row := d.matrix[j*d.cols : (j+1)*d.cols]
			p.set(j, i, row.foldMulSum(t))
		}
		t = t[:0]
	}

	return p
}

func (d *Dense) Augment(b Matrix) Matrix {
	switch b := b.(type) {
	case *Dense:
		return d.AugmentDense(b)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// Join a matrix to the right of d returning the new matrix
func (d *Dense) AugmentDense(b *Dense) *Dense {
	if d.rows != b.rows {
		panic(ErrColLength)
	}

	m := &Dense{
		rows:   d.cols,
		cols:   d.cols + b.cols,
		matrix: make(denseRow, len(d.matrix)+len(b.matrix)),
	}

	for i := 0; i < m.rows; i++ {
		copy(m.matrix[i*m.cols:i*m.cols+d.cols], d.matrix[i*d.cols:(i+1)*d.cols])
		copy(m.matrix[i*m.cols+d.cols:(i+1)*m.cols], b.matrix[i*b.cols:(i+1)*b.cols])
	}

	return m
}

func (d *Dense) Stack(b Matrix) Matrix {
	switch b := b.(type) {
	case *Dense:
		return d.StackDense(b)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// Join a matrix below d returning the new matrix
func (d *Dense) StackDense(b *Dense) *Dense {
	if d.cols != b.cols {
		panic(ErrRowLength)
	}

	m := &Dense{
		rows:   d.rows + b.rows,
		cols:   d.cols,
		matrix: make(denseRow, len(d.matrix)+len(b.matrix)),
	}
	copy(m.matrix, d.matrix)
	copy(m.matrix[len(d.matrix):], b.matrix)

	return m
}

func (d *Dense) Filter(f FilterFunc) Matrix { return d.FilterDense(f) }

// Return a matrix with all elements at (r, c) set to zero where FilterFunc(r, c) returns false
func (d *Dense) FilterDense(f FilterFunc) *Dense {
	m := &Dense{
		rows:   d.rows,
		cols:   d.cols,
		matrix: make(denseRow, len(d.matrix)),
	}

	for i, e := range d.matrix {
		if f(i/d.cols, i%d.cols, e) {
			m.matrix[i] = e
		}
	}

	return m
}

func (d *Dense) Apply(f ApplyFunc) Matrix { return d.ApplyDense(f) }

// Apply a function to non-zero elements of the matrix
func (d *Dense) ApplyDense(f ApplyFunc) *Dense {
	m := d.CloneDense()
	for i, e := range m.matrix {
		if v := f(i/d.cols, i%d.cols, e); v != e {
			m.matrix[i] = v
		}
	}

	return m
}

func (d *Dense) ApplyAll(f ApplyFunc) Matrix { return d.ApplyDense(f) }

func (d *Dense) Format(fs fmt.State, c rune) {
	if c == 'v' && fs.Flag('#') {
		fmt.Fprintf(fs, "&%#v", *d)
		return
	}
	Format(d, d.Margin, fs, c)
}

func (d *Dense) String() string {
	b := &bytes.Buffer{}
	fmt.Fprintf(b, "%6.4e", d)
	return b.String()
}
