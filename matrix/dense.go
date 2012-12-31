// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package matrix

import (
	"fmt"
	"math"
	"math/rand"
	"unsafe"
)

// Type Dense represents a dense matrix.
type Dense struct {
	Margin     int // The number of cells in from the edge of the matrix to format.
	rows, cols int
	matrix     denseRow
}

// A DensePanicker is a function that returns a dense matrix and may panic.
type DensePanicker func() *Dense

// MaybeDense will recover a panic with a type matrix.Error from fn, and return this error.
// Any other error is re-panicked.
func MaybeDense(fn DensePanicker) (d *Dense, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			if err, ok = r.(Error); ok {
				return
			}
			panic(r)
		}
	}()
	return fn(), nil
}

// MustDense can be used to wrap a function returning a dense matrix and an error.
// If the returned error is not nil, MustDense will panic.
func MustDense(d *Dense, err error) *Dense {
	if err != nil {
		panic(err)
	}
	return d
}

// NewDense returns a dense matrix based on a slice of float64 slices. An error is returned
// if either dimension is zero or rows are not of equal length.
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

// ZeroDense returns an r row by c column O matrix. An error is returned if either dimension
// is zero.
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

// IdentityDense returns the a size by size I matrix. An error is returned if size is zero.
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

// FuncDense returns a dense matrix filled with the returned values of fn with a matrix density of rho.
// An error is returned if either dimension is zero.
func FuncDense(r, c int, rho float64, fn FloatFunc) (*Dense, error) {
	if r < 1 || c < 1 {
		return nil, ErrZeroLength
	}

	m := &Dense{
		rows:   r,
		cols:   c,
		matrix: make(denseRow, r*c),
	}

	for i := range m.matrix {
		if rand.Float64() < rho {
			m.matrix[i] = fn()
		}
	}

	return m, nil
}

// ElementsDense returns the elements of mats concatenated, row-wise, into a row vector.
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

// ElementsVector returns the matrix's elements concatenated, row-wise, into a float slice.
func (d *Dense) ElementsVector() []float64 {
	return append([]float64(nil), *(*[]float64)(unsafe.Pointer(&d.matrix))...)
}

// Clone returns a copy of the matrix.
func (d *Dense) Clone() Matrix { return d.CloneDense() }

// Clone returns a copy of the matrix, retaining its concrete type.
func (d *Dense) CloneDense() *Dense {
	return &Dense{
		rows:   d.rows,
		cols:   d.cols,
		matrix: append(denseRow(nil), d.matrix...),
	}
}

// Dense returns the matrix as a Dense. The returned matrix is not a copy.
func (d *Dense) Dense() *Dense { return d }

// Sparse returns a copy of the matrix represented as a Sparse.
func (d *Dense) Sparse() *Sparse {
	s := &Sparse{
		rows:   d.rows,
		cols:   d.cols,
		matrix: make([]sparseRow, d.rows),
	}

	for r := 0; r < d.rows; r++ {
		for c := 0; c < d.cols; c++ {
			if v := d.at(r, c); v != 0 {
				s.matrix[r] = append(s.matrix[r], sparseElem{index: c, value: v})
			}
		}
	}

	return s
}

// Dims return the dimensions of the matrix.
func (d *Dense) Dims() (r, c int) {
	return d.rows, d.cols
}

// Reshape, returns a shallow copy of with the dimensions set to r and c. Reshape will
// panic with ErrShape if r x c does not equal the number of elements in the matrix.
func (d *Dense) Reshape(r, c int) Matrix { return d.ReshapeDense(r, c) }

// ReshapeDense, returns a shallow copy of with the dimensions set to r and c, retaining the concrete
// type of the matrix. ReshapeDense will panic with ErrShape if r x c does not equal the number of
// elements in the matrix.
func (d *Dense) ReshapeDense(r, c int) *Dense {
	if r*c != d.rows*d.cols {
		panic(ErrShape)
	}
	return &Dense{
		rows:   r,
		cols:   c,
		matrix: d.matrix,
	}
}

// Det returns the determinant of the matrix.
// TODO: implement
func (d *Dense) Det() float64 {
	panic("not implemented")
}

// Min returns the value of the minimum element value of the matrix.
func (d *Dense) Min() float64 {
	return d.matrix.min()
}

// Max returns the value of the maximum element value of the matrix.
func (d *Dense) Max() float64 {
	return d.matrix.max()
}

// Set sets the value of the element at (r, c) to v. Set will panic with ErrIndexOutOfRange
// if r or c are not legal indices.
func (d *Dense) Set(r, c int, v float64) {
	if r >= d.rows || c >= d.cols || r < 0 || c < 0 {
		panic(ErrIndexOutOfRange)
	}

	d.set(r, c, v)
}

func (d *Dense) set(r, c int, v float64) {
	d.matrix[r*d.cols+c] = v
}

// At return the value of the element at (r, c). At will panic with ErrIndexOutOfRange if
// r or c are not legal indices.
func (d *Dense) At(r, c int) (v float64) {
	if r >= d.rows || c >= d.cols || c < 0 || r < 0 {
		panic(ErrIndexOutOfRange)
	}
	return d.at(r, c)
}

func (d *Dense) at(r, c int) float64 {
	return d.matrix[r*d.cols+c]
}

// Trace returns the trace of a square matrix. Trace will panic with ErrSquare if the matrix
// is not square.
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

// Norm returns a variety of norms for the matrix.
//
// Valid ord values are:
//
// 	          1 - max of the sum of the absolute values of columns
// 	         -1 - min of the sum of the absolute values of columns
// 	 matrix.Inf - max of the sum of the absolute values of rows
// 	-matrix.Inf - min of the sum of the absolute values of rows
// 	 matrix.Fro - Frobenius norm (0 is an alias to this)
//
// Norm will panic with ErrNormOrder if an illegal norm order is specified.
func (d *Dense) Norm(ord int) float64 {
	var n float64
	if ord == 0 {
		ord = Fro
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

// SumAxis return a column or row vector holding the sums of rows or columns.
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

// MaxAxis return a column or row vector holding the maximum of rows or columns.
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

// MinAxis return a column or row vector holding the minimum of rows or columns.
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

// U returns the upper triangular matrix of the matrix. U will panic with ErrSquare if the matrix is not
// square.
func (d *Dense) U() Matrix { return d.UDense() }

// UDense returns the upper triangular matrix of the matrix retaining the concrete type of the matrix.
// UDense will panic with ErrSquare if the matrix is not square.
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

// L returns the lower triangular matrix of the matrix. L will panic with ErrSquare if the matrix is not
// square.
func (d *Dense) L() Matrix { return d.LDense() }

// LDense returns the lower triangular matrix of the matrix retaining the concrete type of the matrix.
// LDense will panic with ErrSquare if the matrix is not square.
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

// T returns the transpose of the matrix.
func (d *Dense) T() Matrix { return d.TDense() }

// TDense returns the transpose of the matrix retaining the concrete type of the matrix.
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

// Add returns the sum of the matrix and the parameter. Add will panic with ErrShape if the
// two matrices do not have the same dimensions.
func (d *Dense) Add(b Matrix) Matrix {
	switch b := b.(type) {
	case *Dense:
		return d.AddDense(b)
	case *Sparse:
		panic("not implemented")
	case *Pivot:
		panic("not implemented")
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// AddDense returns a dense matrix which is the sum of the matrix and the parameter. AddDense will
// panic with ErrShape if the two matrices do not have the same dimensions.
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

// Sub returns the result of subtraction of the parameter from the matrix. Sub will panic with ErrShape
// if the two matrices do not have the same dimensions.
func (d *Dense) Sub(b Matrix) Matrix {
	switch b := b.(type) {
	case *Dense:
		return d.SubDense(b)
	case *Sparse:
		panic("not implemented")
	case *Pivot:
		panic("not implemented")
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// SubDense returns the result a dense matrics which is the result of subtraction of the parameter from the matrix.
// SubDense will panic with ErrShape if the two matrices do not have the same dimensions.
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

// MulElem returns the element-wise multiplication of the matrix and the parameter. MulElem will panic with ErrShape
// if the two matrices do not have the same dimensions.
func (d *Dense) MulElem(b Matrix) Matrix {
	switch b := b.(type) {
	case *Dense:
		return d.MulElemDense(b)
	case *Sparse:
		panic("not implemented")
	case *Pivot:
		panic("not implemented")
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// MulElemDense returns a dense matrix which is the result of element-wise multiplication of the matrix and the parameter.
// MulElemDense will panic with ErrShape if the two matrices do not have the same dimensions.
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

// Equals returns the equality of two matrices.
func (d *Dense) Equals(b Matrix) bool {
	switch b := b.(type) {
	case *Dense:
		return d.EqualsDense(b)
	case *Sparse:
		panic("not implemented")
	case *Pivot:
		panic("not implemented")
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// EqualsDense returns the equality of two dense matrices.
func (d *Dense) EqualsDense(b *Dense) bool {
	if d.rows != b.rows || d.cols != b.cols {
		return false
	}
	return d.matrix.foldEqual(b.matrix)
}

// EqualsApprox returns the approximate equality of two matrices, tolerance for elemen-wise equality is
// given by epsilon.
func (d *Dense) EqualsApprox(b Matrix, epsilon float64) bool {
	switch b := b.(type) {
	case *Dense:
		return d.EqualsApproxDense(b, epsilon)
	case *Sparse:
		panic("not implemented")
	case *Pivot:
		panic("not implemented")
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// EqualsApproxDense returns the approximate equality of two dense matrices, tolerance for element-wise
// equality is given by epsilon.
func (d *Dense) EqualsApproxDense(b *Dense, epsilon float64) bool {
	if d.rows != b.rows || d.cols != b.cols {
		return false
	}
	return d.matrix.foldApprox(b.matrix, epsilon)
}

// Scalar returns the scalar product of the matrix and f.
func (d *Dense) Scalar(f float64) Matrix { return d.ScalarDense(f) }

// ScalarDense returns the scalar product of the matrix and f as a Dense.
func (d *Dense) ScalarDense(f float64) *Dense {
	return &Dense{
		rows:   d.rows,
		cols:   d.cols,
		matrix: d.matrix.scale(f),
	}
}

// Sum returns the sum of elements in the matrix.
func (d *Dense) Sum() float64 {
	return d.matrix.sum()
}

// Inner returns the sum of element-wise multiplication of the matrix and the parameter. Inner will
// panic with ErrShape if the two matrices do not have the same dimensions.
func (d *Dense) Inner(b Matrix) float64 {
	switch b := b.(type) {
	case *Dense:
		return d.InnerDense(b)
	case *Sparse:
		panic("not implemented")
	case *Pivot:
		panic("not implemented")
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// InnerDense returns a dense matrix which is the result of element-wise multiplication of the matrix and the parameter.
// InnerDense will panic with ErrShape if the two matrices do not have the same dimensions.
func (d *Dense) InnerDense(b *Dense) float64 {
	if d.rows != b.rows || d.cols != b.cols {
		panic(ErrShape)
	}
	return d.matrix.foldMulSum(b.matrix)
}

// Dot returns the matrix product of the matrix and the parameter. Dot will panic with ErrShape if
// the column dimension of the receiver does not equal the row dimension of the parameter.
func (d *Dense) Dot(b Matrix) Matrix {
	switch b := b.(type) {
	case *Dense:
		return d.DotDense(b)
	case *Sparse:
		panic("not implemented")
	case *Pivot:
		panic("not implemented")
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// DotDense returns the matrix product of the matrix and the parameter as a dense matrix. DotDense will panic
// with ErrShape if the column dimension of the receiver does not equal the row dimension of the parameter.
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
		var nonZero bool
		for j := 0; j < b.rows; j++ {
			v := b.at(j, i)
			if v != 0 {
				nonZero = true
			}
			t = append(t, v)
		}
		if nonZero {
			for j := 0; j < d.rows; j++ {
				row := d.matrix[j*d.cols : (j+1)*d.cols]
				p.set(j, i, row.foldMulSum(t))
			}
		}
		t = t[:0]
	}

	return p
}

// Augment returns the augmentation of the receiver with the parameter. Augment will panic with
// ErrColLength if the column dimensions of the two matrices do not match.
func (d *Dense) Augment(b Matrix) Matrix {
	switch b := b.(type) {
	case *Dense:
		return d.AugmentDense(b)
	case *Sparse:
		panic("not implemented")
	case *Pivot:
		panic("not implemented")
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// AugmentDense returns the augmentation of the receiver with the parameter as a dense matrix.
// AugmentDense will panic with ErrColLength if the column dimensions of the two matrices do not match.
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

// Stack returns the stacking of the receiver with the parameter. Stack will panic with
// ErrRowLength if the column dimensions of the two matrices do not match.
func (d *Dense) Stack(b Matrix) Matrix {
	switch b := b.(type) {
	case *Dense:
		return d.StackDense(b)
	case *Sparse:
		panic("not implemented")
	case *Pivot:
		panic("not implemented")
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// StackDense returns the augmentation of the receiver with the parameter as a dense matrix.
// StackDense will panic with ErrRowLength if the column dimensions of the two matrices do not match.
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

// Filter return a matrix with all elements at (r, c) set to zero where FilterFunc(r, c, v) returns false.
func (d *Dense) Filter(f FilterFunc) Matrix { return d.FilterDense(f) }

// FilterDense return a dense matrix with all elements at (r, c) set to zero where FilterFunc(r, c, v) returns false.
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

// Apply returns a matrix which has had a function applied to all elements of the matrix.
func (d *Dense) Apply(f ApplyFunc) Matrix { return d.ApplyDense(f) }

// ApplyDense returns a dense matrix which has had a function applied to all elements of the matrix.
func (d *Dense) ApplyDense(f ApplyFunc) *Dense {
	m := d.CloneDense()
	for i, e := range m.matrix {
		if v := f(i/d.cols, i%d.cols, e); v != e {
			m.matrix[i] = v
		}
	}

	return m
}

// ApplyAll returns a matrix which has had a function applied to all elements of the matrix.
func (d *Dense) ApplyAll(f ApplyFunc) Matrix { return d.ApplyDense(f) }

// ApplyAllDense returns a matrix which has had a function applied to all elements of the matrix.
func (d *Dense) ApplyAllDense(f ApplyFunc) Matrix { return d.ApplyDense(f) }

// Format satisfies the fmt.Formatter interface.
func (d *Dense) Format(fs fmt.State, c rune) {
	if c == 'v' && fs.Flag('#') {
		fmt.Fprintf(fs, "&%#v", *d)
		return
	}
	Format(d, d.Margin, '.', fs, c)
}
