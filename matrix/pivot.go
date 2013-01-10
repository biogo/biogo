// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package matrix

import (
	"code.google.com/p/biogo.blas"
	"fmt"
	"math"
)

// Type Pivot represents a permutation matrix.
type Pivot struct {
	Margin int
	sign   float64
	matrix []int
	xirtam []int
}

// A PivotPanicker is a function that returns a permutation matrix and may panic.
type PivotPanicker func() *Pivot

// MaybePivot will recover a panic with a type matrix.Error from fn, and return this error.
// Any other error is re-panicked.
func MaybePivot(fn PivotPanicker) (p *Pivot, err error) {
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

// MustPivot can be used to wrap a function returning a permutation matrix and an error.
// If the returned error is not nil, MustPivot will panic.
func MustPivot(p *Pivot, err error) *Pivot {
	if err != nil {
		panic(err)
	}
	return p
}

// NewPivot returns a permutation matrix based on a slice of ints representing column indices of non-zero
// elements. An error is returned if the slice length is zero or column indices do not appear exactly once
// or column indices are out of range.
func NewPivot(p []int) (*Pivot, error) {
	if len(p) == 0 {
		return nil, ErrZeroLength
	}

	cv := make([]byte, len(p))
	x := make([]int, len(p))
	for c, r := range p {
		if r < 0 || r >= len(cv) {
			return nil, ErrIndexOutOfRange
		}
		x[r] = c
		cv[r]++
	}
	for _, v := range cv {
		if v != 1 {
			return nil, ErrPivot
		}
	}

	return &Pivot{
		sign:   sign(p),
		matrix: p,
		xirtam: x,
	}, nil
}

func sign(a []int) float64 {
	b := make([]bool, len(a))
	sign := 1.
	for col, row := range a {
		for row != col && !b[col] {
			b[col] = true
			sign = -sign
			col = a[col]
		}
		b[col] = true
	}
	return sign
}

// IdentityPivot returns the a size by size I matrix. An error is returned if size is zero.
func IdentityPivot(size int) (*Pivot, error) {
	if size < 1 {
		return nil, ErrZeroLength
	}

	m := &Pivot{
		sign:   1,
		matrix: make([]int, size),
	}

	for i := range m.matrix {
		m.matrix[i] = i
	}

	return m, nil
}

// Clone returns a copy of the matrix.
func (p *Pivot) Clone(_ Matrix) Matrix { return p.ClonePivot() }

// Clone returns a copy of the matrix, retaining its concrete type.
func (p *Pivot) ClonePivot() *Pivot {
	return &Pivot{
		sign:   p.sign,
		matrix: append([]int(nil), p.matrix...),
	}
}

// Sparse returns a copy of the matrix represented as a Sparse.
func (p *Pivot) Sparse(s *Sparse) *Sparse {
	s = s.reallocate(len(p.matrix), len(p.matrix))

	for r, c := range p.xirtam {
		s.matrix[r] = sparseRow{sparseElem{index: c, value: 1}}
	}

	return s
}

// Dense returns a copy of the matrix represented as a Dense.
func (p *Pivot) Dense(d *Dense) *Dense {
	d = d.reallocate(len(p.matrix), len(p.matrix))

	for r, c := range p.xirtam {
		d.set(r, c, 1)
	}

	return d
}

// Dims return the dimensions of the matrix.
func (p *Pivot) Dims() (r, c int) {
	return len(p.matrix), len(p.matrix)
}

// Reshape, returns a shallow copy of with the dimensions set to r and c. Reshape will
// panic with ErrShape if r x c does not equal the number of elements in the matrix.
func (p *Pivot) Reshape(r, c int) Matrix { return p.ReshapeSparse(r, c) }

// ReshapeSparse, returns a shallow copy of with the dimensions set to r and c, represented
// as a sparse matrix. ReshapeSparse will panic with ErrShape if r x c does not equal the number of
// elements in the matrix.
func (p *Pivot) ReshapeSparse(r, c int) *Sparse {
	if r*c != len(p.matrix)*len(p.matrix) {
		panic(ErrShape)
	}
	panic("not implemented")
}

// Det returns the determinant of the matrix.
func (p *Pivot) Det() float64 {
	return p.sign
}

// Min returns the value of the minimum element value of the matrix, 0.
func (p *Pivot) Min() float64 {
	return 0
}

// Max returns the value of the maximum element value of the matrix, 1.
func (p *Pivot) Max() float64 {
	return 1
}

// At return the value of the element at (r, c). At will panic with ErrIndexOutOfRange if
// r or c are not legal indices.
func (p *Pivot) At(r, c int) (v float64) {
	if r >= len(p.matrix) || c >= len(p.matrix) || c < 0 || r < 0 {
		panic(ErrIndexOutOfRange)
	}
	return p.at(r, c)
}

func (p *Pivot) at(r, c int) float64 {
	if p.matrix[c] == r {
		return 1
	}
	return 0
}

// Trace returns the trace of the matrix.
func (p *Pivot) Trace() float64 {
	var t float64
	for i := 0; i < len(p.matrix); i++ {
		if p.matrix[i] == i {
			t++
		}
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
func (p *Pivot) Norm(ord int) float64 {
	switch ord {
	case 2, -2:
		return 1
	case 1:
		return 1
	case Inf:
		return 1
	case -1:
		return 1
	case -Inf:
		return 1
	case 0, Fro:
		return math.Sqrt(float64(len(p.matrix)))
	default:
		panic(ErrNormOrder)
	}

	panic("cannot reach")
}

// SumAxis return a column or row vector holding the sums of rows or columns.
func (p *Pivot) SumAxis(cols bool) *Dense {
	m := make([]float64, len(p.matrix))
	for i := range m {
		m[i] = 1
	}
	var r, c int
	if cols {
		r, c = 1, len(m)
	} else {
		r, c = len(m), 1
	}
	return &Dense{
		matrix: m,
		rows:   r,
		cols:   c,
	}
}

// MaxAxis return a column or row vector holding the maximum of rows or columns.
func (p *Pivot) MaxAxis(cols bool) *Dense {
	m := make([]float64, len(p.matrix))
	for i := range m {
		m[i] = 1
	}
	var r, c int
	if cols {
		r, c = 1, len(m)
	} else {
		r, c = len(m), 1
	}
	return &Dense{
		matrix: m,
		rows:   r,
		cols:   c,
	}
}

// MinAxis return a column or row vector holding the minimum of rows or columns.
func (p *Pivot) MinAxis(cols bool) *Dense {
	m := make([]float64, len(p.matrix))
	var r, c int
	if cols {
		r, c = 1, len(m)
	} else {
		r, c = len(m), 1
	}
	return &Dense{
		matrix: m,
		rows:   r,
		cols:   c,
	}
}

// U returns the upper triangular matrix of the matrix.
func (p *Pivot) U(c Matrix) Matrix {
	cc, _ := c.(*Sparse)
	return p.USparse(cc)
}

// USparse returns the upper triangular matrix of the matrix represented as a sparse matrix.
func (p *Pivot) USparse(c *Sparse) *Sparse {
	c = c.reallocate(len(p.matrix), len(p.matrix))
	for row, col := range p.xirtam {
		if row >= col {
			c.matrix[row] = append(c.matrix[row], sparseElem{index: col, value: 1})
		}
	}
	return c
}

// L returns the lower triangular matrix of the matrix.
func (p *Pivot) L(c Matrix) Matrix {
	cc, _ := c.(*Sparse)
	return p.LSparse(cc)
}

// LDense returns the lower triangular matrix of the matrix represented as a sparse matrix.
func (p *Pivot) LSparse(c *Sparse) *Sparse {
	c = c.reallocate(len(p.matrix), len(p.matrix))
	for row, col := range p.xirtam {
		if row <= col {
			c.matrix[row] = append(c.matrix[row], sparseElem{index: col, value: 1})
		}
	}
	return c
}

// T returns the transpose of the matrix.
func (p *Pivot) T(_ Matrix) Matrix { return p.TPivot() }

// TPivot returns the transpose of the matrix retaining the concrete type of the matrix.
func (p *Pivot) TPivot() *Pivot {
	c := &Pivot{
		sign:   p.sign,
		matrix: p.xirtam,
		xirtam: p.matrix,
	}
	return c
}

// Add returns the sum of the matrix and the parameter. Add will panic with ErrShape if the
// two matrices do not have the same dimensions.
func (p *Pivot) Add(b, c Matrix) Matrix {
	switch b := b.(type) {
	case *Pivot:
		cc, _ := c.(*Sparse)
		return p.AddPivot(b, cc)
	case *Sparse:
		cc, _ := c.(*Sparse)
		return b.addPivot(p, cc)
	case *Dense:
		cc, _ := c.(*Dense)
		return b.addPivot(p, cc)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// AddPivot returns a sparse matrix which is the sum of the matrix and the parameter. AddPivot will
// panic with ErrShape if the two matrices do not have the same dimensions.
func (p *Pivot) AddPivot(b *Pivot, c *Sparse) *Sparse {
	if len(p.matrix) != len(b.matrix) {
		panic(ErrShape)
	}

	c = c.reallocate(p.Dims())

	return nil
}

// Sub returns the result of subtraction of the parameter from the matrix. Sub will panic with ErrShape
// if the two matrices do not have the same dimensions.
func (p *Pivot) Sub(b, c Matrix) Matrix {
	switch b := b.(type) {
	case *Pivot:
		cc, _ := c.(*Sparse)
		return p.SubPivot(b, cc)
	case *Sparse:
		cc, _ := c.(*Sparse)
		return p.subSparse(b, cc)
	case *Dense:
		cc, _ := c.(*Dense)
		return p.subDense(b, cc)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// SubPivot returns the result a dense matrics which is the result of subtraction of the parameter from the matrix.
// SubPivot will panic with ErrShape if the two matrices do not have the same dimensions.
func (p *Pivot) SubPivot(b *Pivot, c *Sparse) *Sparse {
	if len(p.matrix) != len(b.matrix) {
		panic(ErrShape)
	}

	c = c.reallocate(p.Dims())

	return nil
}

func (p *Pivot) subDense(b, c *Dense) *Dense {
	if len(p.matrix) != b.rows || len(p.matrix) != b.cols {
		panic(ErrShape)
	}

	if c != b {
		c = c.reallocate(p.Dims())
		copy(c.matrix, b.matrix)
		blas.Dscal(len(c.matrix), -1, c.matrix, 1)
	}
	for row, col := range p.xirtam {
		c.matrix[row*c.cols+col]++
	}

	return c
}

func (p *Pivot) subSparse(b, c *Sparse) *Sparse {
	if b.rows != len(p.matrix) || b.cols != len(p.matrix) {
		panic(ErrShape)
	}

	if c != b {
		c = b.CloneSparse(c)
	}
	for row, col := range p.xirtam {
		_, i := c.matrix[row].atInd(col)
		if i < 0 {
			c.set(row, col, 1)
			continue
		}
		c.matrix[row][i].value = 1 - c.matrix[row][i].value
	}

	return c
}

// MulElem returns the element-wise multiplication of the matrix and the parameter. MulElem will panic with ErrShape
// if the two matrices do not have the same dimensions.
func (p *Pivot) MulElem(b, c Matrix) Matrix {
	switch b := b.(type) {
	case *Pivot:
		cc, _ := c.(*Sparse)
		return p.MulElemPivot(b, cc)
	case *Sparse:
		cc, _ := c.(*Sparse)
		return b.Filter(func(row, col int, _ float64) bool { return p.xirtam[row] == col }, cc)
	case *Dense:
		cc, _ := c.(*Dense)
		return b.mulElemPivot(p, cc)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// MulElemPivot returns a dense matrix which is the result of element-wise multiplication of the matrix and the parameter.
// MulElemPivot will panic with ErrShape if the two matrices do not have the same dimensions.
func (p *Pivot) MulElemPivot(b *Pivot, c *Sparse) *Sparse {
	if len(p.matrix) != len(b.matrix) {
		panic(ErrShape)
	}

	c = c.reallocate(p.Dims())

	return nil
}

// Equals returns the equality of two matrices.
func (p *Pivot) Equals(b Matrix) bool {
	switch b := b.(type) {
	case *Pivot:
		return p.EqualsPivot(b)
	case *Sparse:
		return b.equalsPivot(p)
	case *Dense:
		return b.equalsPivot(p)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// EqualsPivot returns the equality of two dense matrices.
func (p *Pivot) EqualsPivot(b *Pivot) bool {
	if p.sign != b.sign || len(p.matrix) != len(b.matrix) {
		return false
	}
	for i, v := range p.matrix {
		if b.matrix[i] != v {
			return false
		}
	}
	return true
}

// EqualsApprox returns the approximate equality of two matrices, tolerance for element-wise equality is
// given by epsilon.
func (p *Pivot) EqualsApprox(b Matrix, epsilon float64) bool {
	switch b := b.(type) {
	case *Pivot:
		return p.EqualsApproxPivot(b, epsilon)
	case *Sparse:
		return b.equalsApproxPivot(p, epsilon)
	case *Dense:
		return b.equalsApproxPivot(p, epsilon)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// EqualsApproxPivot returns the approximate equality of two dense matrices, tolerance for element-wise
func (p *Pivot) EqualsApproxPivot(b *Pivot, epsilon float64) bool {
	if len(p.matrix) != len(b.matrix) {
		return false
	}
	for i, v := range p.matrix {
		if b.matrix[i] != v && 1 > epsilon {
			return false
		}
	}
	return true
}

// Scalar returns the scalar product of the matrix and f.
func (p *Pivot) Scalar(f float64, c Matrix) Matrix {
	cc, _ := c.(*Sparse)
	return p.ScalarSparse(f, cc)
}

// ScalarSparse returns the scalar product of the matrix and f as a Dense.
func (p *Pivot) ScalarSparse(f float64, c *Sparse) *Sparse {
	c = c.reallocate(p.Dims())
	for row, col := range p.xirtam {
		c.matrix[row] = append(c.matrix[row], sparseElem{index: col, value: 1 * f})
	}
	return c
}

// Sum returns the sum of elements in the matrix.
func (p *Pivot) Sum() float64 {
	return float64(len(p.matrix))
}

// Inner returns the sum of element-wise multiplication of the matrix and the parameter. Inner will
// panic with ErrShape if the two matrices do not have the same dimensions.
func (p *Pivot) Inner(b Matrix) float64 {
	switch b := b.(type) {
	case *Pivot:
		return p.InnerPivot(b)
	case *Sparse:
		return b.innerPivot(p)
	case *Dense:
		return b.innerPivot(p)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// InnerPivot returns a dense matrix which is the result of element-wise multiplication of the matrix and the parameter.
// InnerPivot will panic with ErrShape if the two matrices do not have the same dimensions.
func (p *Pivot) InnerPivot(b *Pivot) float64 {
	if len(p.matrix) != len(b.matrix) {
		panic(ErrShape)
	}
	var ip float64
	for i, v := range p.matrix {
		if b.matrix[i] == v {
			ip += p.sign * b.sign
		}
	}
	return ip
}

// Dot returns the matrix product of the matrix and the parameter. Dot will panic with ErrShape if
// the column dimension of the receiver does not equal the row dimension of the parameter.
func (p *Pivot) Dot(b, c Matrix) Matrix {
	switch b := b.(type) {
	case *Pivot:
		cc, _ := c.(*Pivot)
		return p.DotPivot(b, cc)
	case *Sparse:
		cc, _ := c.(*Sparse)
		return p.DotSparse(b, cc)
	case *Dense:
		cc, _ := c.(*Dense)
		return p.DotDense(b, cc)
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// DotPivot returns the matrix product of the matrix and the parameter as a dense matrix. DotDense will panic
// with ErrShape if the column dimension of the receiver does not equal the row dimension of the parameter.
func (p *Pivot) DotPivot(b, c *Pivot) *Pivot {
	if len(p.matrix) != len(b.matrix) {
		panic(ErrShape)
	}

	np := &Pivot{
		sign:   p.sign * b.sign,
		matrix: make([]int, len(p.matrix)),
		xirtam: make([]int, len(p.matrix)),
	}
	for r := 0; r < len(b.matrix); r++ {
		np.xirtam[r] = p.xirtam[b.xirtam[r]]
	}
	for r, c := range np.xirtam {
		np.matrix[c] = r
	}

	return np
}

// DotSparse returns the matrix product of the matrix and the parameter as a dense matrix. DotDense will panic
// with ErrShape if the column dimension of the receiver does not equal the row dimension of the parameter.
func (p *Pivot) DotSparse(b, c *Sparse) *Sparse {
	if len(p.matrix) != b.rows {
		panic(ErrShape)
	}

	if c != b {
		c = c.reallocate(b.rows, b.cols)
		for to, from := range p.matrix {
			c.matrix[to] = append(c.matrix[to], b.matrix[from]...)
		}
		return c
	}

	visit := make([]bool, len(p.matrix))
	for to, from := range p.matrix {
		for to != from && !visit[from] {
			visit[from] = true
			b.matrix[from], b.matrix[to] = c.matrix[to], c.matrix[from]
			from = p.matrix[from]
		}
		visit[from] = true
	}

	return c
}

// swap rows of a dense matrix
func (p *Pivot) DotDense(b *Dense, c *Dense) *Dense {
	if len(p.matrix) != b.rows {
		panic(ErrShape)
	}

	if c != b {
		c = c.reallocate(b.rows, b.cols)
		for to, from := range p.matrix {
			blas.Dcopy(b.cols, b.matrix[from*b.cols:], 1, c.matrix[to*c.cols:], 1)
		}
		return c
	}

	visit := make([]bool, len(p.matrix))
	for to, from := range p.matrix {
		for to != from && !visit[from] {
			visit[from] = true
			blas.Dswap(b.cols, b.matrix[from*b.cols:], 1, c.matrix[to*c.cols:], 1)
			from = p.matrix[from]
		}
		visit[from] = true
	}

	return c
}

// Augment returns the augmentation of the receiver with the parameter. Augment will panic with
// ErrColLength if the column dimensions of the two matrices do not match.
func (p *Pivot) Augment(b, c Matrix) Matrix {
	switch b := b.(type) {
	case *Pivot:
		cc, _ := c.(*Sparse)
		return p.AugmentPivot(b, cc)
	case *Sparse:
		panic("not implemented")
	case *Dense:
		panic("not implemented")
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// AugmentPivot returns the augmentation of the receiver with the parameter as a sparse matrix.
// AugmentPivot will panic with ErrColLength if the column dimensions of the two matrices do not match.
func (p *Pivot) AugmentPivot(b *Pivot, c *Sparse) *Sparse {
	if len(p.matrix) != len(b.matrix) {
		panic(ErrColLength)
	}
	c = c.reallocate(len(p.matrix), len(p.matrix)*2)
	for row, col := range p.xirtam {
		c.matrix[row] = append(c.matrix[row], sparseElem{index: col, value: 1}, sparseElem{index: b.xirtam[row] + len(p.matrix), value: 1})
	}
	return c
}

// Stack returns the stacking of the receiver with the parameter. Stack will panic with
// ErrRowLength if the column dimensions of the two matrices do not match.
func (p *Pivot) Stack(b, c Matrix) Matrix {
	switch b := b.(type) {
	case *Pivot:
		cc, _ := c.(*Sparse)
		return p.StackPivot(b, cc)
	case *Sparse:
		cc, _ := c.(*Sparse)
		return p.StackSparse(b, cc)
	case *Dense:
		panic("not implemented")
	default:
		panic("not implemented")
	}

	panic("cannot reach")
}

// StackSparse returns the augmentation of the receiver with the parameter as a sparse matrix.
// StackSparse will panic with ErrRowLength if the column dimensions of the two matrices do not match.
func (p *Pivot) StackPivot(b *Pivot, c *Sparse) *Sparse {
	if len(p.matrix) != len(b.matrix) {
		panic(ErrRowLength)
	}
	c = c.reallocate(len(p.matrix)*2, len(p.matrix))
	for row, col := range p.xirtam {
		c.matrix[row] = append(c.matrix[row], sparseElem{index: col, value: 1})
	}
	for row, col := range b.xirtam {
		c.matrix[row+len(p.matrix)] = append(c.matrix[row+len(p.matrix)], sparseElem{index: col, value: 1})
	}
	return c
}

// StackSparse returns the augmentation of the receiver with the parameter as a sparse matrix.
// StackSparse will panic with ErrRowLength if the column dimensions of the two matrices do not match.
func (p *Pivot) StackSparse(b *Sparse, c *Sparse) *Sparse {
	if len(p.matrix) != b.cols {
		panic(ErrRowLength)
	}
	c = c.reallocate(len(p.matrix)*2, len(p.matrix))
	for row, col := range p.xirtam {
		c.matrix[row] = append(c.matrix[row], sparseElem{index: col, value: 1})
	}
	for row, col := range b.matrix {
		c.matrix[row+len(p.matrix)] = append(c.matrix[row+len(p.matrix)], col...)
	}
	return c
}

// Filter return a matrix with all elements at (r, c) set to zero where FilterFunc(r, c, v) returns false.
func (p *Pivot) Filter(f FilterFunc, c Matrix) Matrix {
	cc, _ := c.(*Sparse)
	return p.FilterSparse(f, cc)
}

// FilterSparse return a sparse matrix with all elements at (r, c) set to zero where FilterFunc(r, c, v) returns false.
func (p *Pivot) FilterSparse(f FilterFunc, c *Sparse) *Sparse {
	c = c.reallocate(p.Dims())
	for row, col := range p.xirtam {
		if f(row, col, 1) {
			c.matrix[row] = sparseRow{sparseElem{index: col, value: 1}}
		}
	}
	return c
}

// Apply returns a matrix which has had a function applied to all non-zero elements of the matrix.
func (p *Pivot) Apply(f ApplyFunc, c Matrix) Matrix {
	cc, _ := c.(*Sparse)
	return p.ApplySparse(f, cc)
}

// ApplySparse returns a dense matrix which has had a function applied to all non-zero elements of the matrix.
func (p *Pivot) ApplySparse(f ApplyFunc, c *Sparse) *Sparse {
	c = c.reallocate(p.Dims())
	for row, col := range p.xirtam {
		if v := f(row, col, 1); v != 0 {
			c.matrix[row] = append(c.matrix[row], sparseElem{index: col, value: v})
		}
	}
	return c
}

// ApplyAll returns a matrix which has had a function applied to all elements of the matrix.
func (p *Pivot) ApplyAll(f ApplyFunc, c Matrix) Matrix {
	cc, _ := c.(*Sparse)
	return p.ApplyAllSparse(f, cc)
}

// ApplyAllSparse returns a matrix which has had a function applied to all elements of the matrix.
func (p *Pivot) ApplyAllSparse(f ApplyFunc, c *Sparse) *Sparse {
	c = c.reallocate(p.Dims())
	for i := 0; i < len(p.matrix); i++ {
		for j := 0; j < len(p.matrix); j++ {
			old := p.at(i, j)
			v := f(i, j, old)
			if v != old {
				c.set(i, j, v)
			}
		}
	}

	return c
}

// Format satisfies the fmt.Formatter interface.
func (p *Pivot) Format(fs fmt.State, c rune) {
	if c == 'v' && fs.Flag('#') {
		fmt.Fprintf(fs, "&%#v", *p)
		return
	}
	Format(p, p.Margin, '.', fs, c)
}
