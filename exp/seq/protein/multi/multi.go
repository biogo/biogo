// Copyright ©2011-2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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

// Package multi handles collections of sequences as alignments or sets.
package multi

import (
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/feat"
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/exp/seq/protein"
	"code.google.com/p/biogo/exp/seq/sequtils"
	"code.google.com/p/biogo/util"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"sync"
)

func init() {
	joinerRegistryLock = &sync.RWMutex{}
	joinerRegistry = make(map[reflect.Type]JoinFunc)
}

var (
	joinerRegistryLock *sync.RWMutex
	joinerRegistry     map[reflect.Type]JoinFunc
)

type rowCounter interface {
	Rows() int
}

func rows(s protein.Sequence) int {
	row := 1
	if m, ok := s.(rowCounter); ok {
		row = m.Rows()
	}
	return row
}

type Multi struct {
	protein.Annotation
	Seq        []protein.Sequence
	Consensify seq.ConsenseFunc
	Encode     alphabet.Encoding
}

// Create a new Multi sequence.
func NewMulti(id string, n []protein.Sequence, cons seq.ConsenseFunc) (*Multi, error) {
	var alpha alphabet.Peptide
	for _, s := range n {
		if alpha != nil && s.Alphabet() != alpha {
			return nil, errors.New("multi: inconsistent alphabets")
		} else if alpha == nil {
			alpha = s.Alphabet().(alphabet.Peptide)
		}
	}
	return &Multi{
		Annotation: protein.Annotation{
			ID:    id,
			Alpha: alpha,
		},
		Seq:        n,
		Consensify: cons,
	}, nil
}

// At returns the letter at position pos.
func (m *Multi) At(pos seq.Position) alphabet.QLetter {
	for _, r := range m.Seq {
		row := rows(r)
		if pos.Row < row {
			return r.At(pos)
		}
		pos.Row -= row
	}

	panic("multi: index out of range")
}

// Set sets the letter at position pos to l.
func (m *Multi) Set(pos seq.Position, l alphabet.QLetter) {
	for _, r := range m.Seq {
		row := rows(r)
		if pos.Row < row {
			r.Set(pos, l)
			return
		}
		pos.Row -= row
	}

	panic("multi: index out of range")
}

// SetE sets the quality at position pos to e to reflect the given p(Error).
func (m *Multi) SetE(pos seq.Position, q float64) {
	for _, r := range m.Seq {
		row := rows(r)
		if pos.Row < row {
			if qs, ok := r.(seq.Quality); ok {
				qs.SetE(pos, q)
				return
			}
		}
		pos.Row -= row
	}

	panic("multi: index out of range")
}

// QEncode encodes the quality at position pos to a letter based on the sequence encoding setting.
func (m *Multi) QEncode(pos seq.Position) byte {
	return m.At(pos).Q.Encode(m.Encode)
}

// Encoding returns the quality encoding scheme.
func (m *Multi) Encoding() alphabet.Encoding { return m.Encode }

// SetEncoding sets the quality encoding scheme to e.
func (m *Multi) SetEncoding(e alphabet.Encoding) {
	for _, r := range m.Seq {
		if enc, ok := r.(seq.Scorer); ok {
			enc.SetEncoding(e)
		}
	}
	m.Encode = e
}

// EAt returns the probability of a sequence error at position pos.
func (m *Multi) EAt(pos seq.Position) float64 {
	for _, r := range m.Seq {
		row := rows(r)
		if pos.Row < row {
			if qs, ok := r.(seq.Quality); ok {
				return qs.EAt(pos)
			} else {
				return protein.DefaultQphred.ProbE()
			}
		}
		pos.Row -= row
	}

	panic("multi: index out of range")
}

// Len returns the length of the alignment.
func (m *Multi) Len() int {
	var (
		min = util.MaxInt
		max = util.MinInt
	)

	for _, r := range m.Seq {
		if start := r.Start(); start < min {
			min = start
		}
		if end := r.End(); end > max {
			max = end
		}
	}

	return max - min
}

// Rows returns the number of rows in the alignment.
func (m *Multi) Rows() int {
	var c int
	for _, r := range m.Seq {
		c += rows(r)
	}

	return c
}

// SetOffset sets the global offset of the sequence to o.
func (m *Multi) SetOffset(o int) {
	for _, r := range m.Seq {
		r.SetOffset(r.Start() - m.Offset + o)
	}
	m.Offset = o
}

// Start returns the start position of the sequence in global coordinates.
func (m *Multi) Start() int {
	start := util.MaxInt
	for _, r := range m.Seq {
		if lt := r.Start(); lt < start {
			start = lt
		}
	}

	return start
}

// End returns the end position of the sequence in global coordinates.
func (m *Multi) End() int {
	end := util.MinInt
	for _, m := range m.Seq {
		if rt := m.End(); rt > end {
			end = rt
		}
	}

	return end
}

// Copy returns a copy of the sequence.
func (m *Multi) Copy() *Multi {
	c := &Multi{}
	*c = *m
	c.Seq = make([]protein.Sequence, len(m.Seq))
	for i, r := range m.Seq {
		c.Seq[i] = r.Copy().(protein.Sequence)
	}

	return c
}

// Reverse reverses the order of letters in the the sequence without complementing them.
func (m *Multi) Reverse() {
	end := m.End()
	for _, r := range m.Seq {
		r.Reverse()
		r.SetOffset(end - m.End())
	}
}

// Add adds sequences n to the multiple sequence.
func (m *Multi) Add(n ...protein.Sequence) error {
	for _, r := range n {
		if r.Alphabet() != m.Alpha {
			return errors.New("multi: inconsistent alphabets")
		}
	}
	m.Seq = append(m.Seq, n...)

	return nil
}

// TODO
func (m *Multi) Delete(i int) {}

// Get returns the sequence corresponding to the ith row of the Seq.
func (m *Multi) Get(i int) protein.Sequence {
	var row int
	for _, r := range m.Seq {
		if m, ok := r.(protein.Getter); ok {
			row = m.Rows()
			if i < row {
				return m.Get(i)
			}
		} else {
			row = 1
			if i == 0 {
				return r
			}
		}
		i -= row
	}

	panic("multi: index out of range")
}

// Append appends a to the ith sequence in the receiver.
func (m *Multi) Append(i int, a ...alphabet.QLetter) (err error) {
	return m.Get(i).(seq.Appender).AppendQLetters(a...)
}

// Append each byte of each a to the appropriate sequence in the reciever.
func (m *Multi) AppendColumns(a ...[]alphabet.QLetter) (err error) {
	for i, c := range a {
		if len(c) != m.Rows() {
			return fmt.Errorf("multi: column %d does not match Rows(): %d != %d.", i, len(c), m.Rows())
		}
	}
	for i, b := 0, make([]alphabet.QLetter, 0, len(a)); i < m.Rows(); i, b = i+1, b[:0] {
		for _, r := range a {
			b = append(b, r[i])
		}
		m.Append(i, b...)
	}

	return
}

// AppendEach appends each []alphabet.QLetter in a to the appropriate sequence in the receiver.
func (m *Multi) AppendEach(a [][]alphabet.QLetter) (err error) {
	if len(a) != m.Rows() {
		return fmt.Errorf("multi: number of sequences does not match Rows(): %d != %d.", len(a), m.Rows())
	}
	var i int
	for _, r := range m.Seq {
		if al, ok := r.(protein.AlignedAppender); ok {
			row := al.Rows()
			if al.AppendEach(a[i:i+row]) != nil {
				panic("internal size mismatch")
			}
			i += row
		} else {
			r.(seq.Appender).AppendQLetters(a[i]...)
			i++
		}
	}

	return
}

// Column returns a slice of letters reflecting the column at pos.
func (m *Multi) Column(pos int, fill bool) []alphabet.Letter {
	if pos < m.Start() || pos >= m.End() {
		panic("multi: index out of range")
	}

	var c []alphabet.Letter
	if fill {
		c = make([]alphabet.Letter, 0, m.Rows())
	} else {
		c = []alphabet.Letter{}
	}

	for _, r := range m.Seq {
		if a, ok := r.(seq.Aligned); ok {
			if a.Start() <= pos && pos < a.End() {
				c = append(c, a.Column(pos, fill)...)
			} else if fill {
				c = append(c, m.Alpha.Gap().Repeat(a.Rows())...)
			}
		} else {
			if r.Start() <= pos && pos < r.End() {
				c = append(c, r.At(seq.Position{Col: pos}).L)
			} else if fill {
				c = append(c, m.Alpha.Gap())
			}
		}
	}

	return c
}

// ColumnQL returns a slice of quality letters reflecting the column at pos.
func (m *Multi) ColumnQL(pos int, fill bool) []alphabet.QLetter {
	if pos < m.Start() || pos >= m.End() {
		panic("multi: index out of range")
	}

	var c []alphabet.QLetter
	if fill {
		c = make([]alphabet.QLetter, 0, m.Rows())
	} else {
		c = []alphabet.QLetter{}
	}

	for _, r := range m.Seq {
		if a, ok := r.(seq.Aligned); ok {
			if a.Start() <= pos && pos < a.End() {
				c = append(c, a.ColumnQL(pos, fill)...)
			} else if fill {
				c = append(c, alphabet.QLetter{L: m.Alpha.Gap()}.Repeat(a.Rows())...)
			}
		} else {
			if r.Start() <= pos && pos < r.End() {
				c = append(c, r.At(seq.Position{Col: pos}))
			} else if fill {
				c = append(c, alphabet.QLetter{L: m.Alpha.Gap()})
			}
		}
	}

	return c
}

// IsFlush returns a boolean indicating whether the end specified by where is flush - that is
// all the contributing sequences start at the same offset.
func (m *Multi) IsFlush(where int) bool {
	if m.Rows() <= 1 {
		return true
	}
	var start, end int
	for i, r := range m.Seq {
		if lt, rt := r.Start(), r.End(); i > 0 &&
			((lt != start && where&seq.Start != 0) ||
				(rt != end && where&seq.End != 0)) {
			return false
		} else if i == 0 {
			start, end = lt, rt
		}
	}
	return true
}

// Flush fills ragged sequences with the receiver's gap letter so that all sequences are flush.
func (m *Multi) Flush(where int, fill alphabet.Letter) {
	if m.IsFlush(where) {
		return
	}

	if where&seq.Start != 0 {
		start := m.Start()
		for _, r := range m.Seq {
			if r.Start()-start < 1 {
				continue
			}
			switch sl := r.Slice().(type) {
			case alphabet.Letters:
				r.SetSlice(alphabet.Letters(append(fill.Repeat(r.Start()-start), sl...)))
			case alphabet.QLetters:
				r.SetSlice(alphabet.QLetters(append(alphabet.QLetter{L: fill}.Repeat(r.Start()-start), sl...)))
			}
			r.SetOffset(start)
		}
	}
	if where&seq.End != 0 {
		end := m.End()
		for i := 0; i < m.Rows(); i++ {
			r := m.Get(i)
			if end-r.End() < 1 {
				continue
			}
			r.(seq.Appender).AppendQLetters(alphabet.QLetter{L: fill}.Repeat(end - r.End())...)
		}
	}
}

// Subseq returns a multiple subsequence slice of the receiver.
func (m *Multi) Subseq(start, end int) (*Multi, error) {
	var ns []protein.Sequence

	for _, r := range m.Seq {
		rs := reflect.New(reflect.TypeOf(r)).Interface().(sequtils.Sliceable)
		err := sequtils.Truncate(rs, r, start, end)
		if err != nil {
			return nil, err
		}
		ns = append(ns, rs.(protein.Sequence))
	}

	ss := &Multi{}
	*ss = *m
	ss.Seq = ns

	return ss, nil
}

// Truncate truncates the the receiver from start to end.
func (m *Multi) Truncate(start, end int) error {
	for _, r := range m.Seq {
		err := sequtils.Truncate(r, r, start, end)
		if err != nil {
			return err
		}
	}

	return nil
}

// Join joins a to the receiver at the end specied by where.
func (m *Multi) Join(a *Multi, where int) error {
	if m.Rows() != a.Rows() {
		return fmt.Errorf("multi: row number mismatch %d != %d", m.Rows(), a.Rows())
	}

	switch where {
	case seq.Start:
		if !a.IsFlush(seq.End) {
			a.Flush(seq.End, m.Alpha.Gap())
		}
		if !m.IsFlush(seq.Start) {
			m.Flush(seq.Start, m.Alpha.Gap())
		}
	case seq.End:
		if !a.IsFlush(seq.Start) {
			a.Flush(seq.Start, m.Alpha.Gap())
		}
		if !m.IsFlush(seq.End) {
			m.Flush(seq.End, m.Alpha.Gap())
		}
	}

	for i := 0; i < m.Rows(); i++ {
		r := m.Get(i)
		as := a.Get(i)
		err := joinOne(r, as, where)
		if err != nil {
			return err
		}
	}

	return nil
}

func joinOne(m, am protein.Sequence, where int) error {
	switch m.(type) {
	case *protein.Seq:
		_, ok := am.(*protein.Seq)
		if !ok {
			goto MISMATCH
		}
		return sequtils.Join(m, am, where)
	case *protein.QSeq:
		_, ok := am.(*protein.QSeq)
		if !ok {
			goto MISMATCH
		}
		return sequtils.Join(m, am, where)
	default:
		joinerRegistryLock.RLock()
		defer joinerRegistryLock.RUnlock()
		joinerFunc, ok := joinerRegistry[reflect.TypeOf(m)]
		if !ok {
			return fmt.Errorf("multi: sequence type %T not handled.", m)
		}
		if reflect.TypeOf(m) != reflect.TypeOf(am) {
			goto MISMATCH
		}
		return joinerFunc(m, am, where)
	}

MISMATCH:
	return fmt.Errorf("multi: sequence type mismatch: %T != %T.", m, am)
}

type JoinFunc func(a, b protein.Sequence, where int) (err error)

func RegisterJoiner(p protein.Sequence, f JoinFunc) {
	joinerRegistryLock.Lock()
	joinerRegistry[reflect.TypeOf(p)] = f
	joinerRegistryLock.Unlock()
}

type ft struct {
	s, e int
}

func (f *ft) Start() int                    { return f.s }
func (f *ft) End() int                      { return f.e }
func (f *ft) Len() int                      { return f.e - f.s }
func (f *ft) Orientation() feat.Orientation { return feat.Forward }
func (f *ft) Name() string                  { return "" }
func (f *ft) Description() string           { return "" }
func (f *ft) Location() feat.Feature        { return nil }

type fts []feat.Feature

func (f fts) Features() []feat.Feature { return f }
func (f fts) Len() int                 { return len(f) }
func (f fts) Less(i, j int) bool       { return f[i].Start() < f[j].Start() }
func (f fts) Swap(i, j int)            { f[i], f[j] = f[j], f[i] }

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Stitch produces a subsequence of the receiver defined by fs. The result is stored in the receiver
// and all contributing sequences are modified.
func (m *Multi) Stitch(fs feat.Set) error {
	ff := fs.Features()
	for _, f := range ff {
		if f.End() < f.Start() {
			return errors.New("multi: feature end < feature start")
		}
	}
	ff = append(fts(nil), ff...)
	sort.Sort(fts(ff))

	var (
		fsp = make(fts, 0, len(ff))
		csp *ft
	)
	for i, f := range ff {
		if s := f.Start(); i == 0 || s > csp.e {
			csp = &ft{s: s, e: f.End()}
			fsp = append(fsp, csp)
		} else {
			csp.e = max(csp.e, f.End())
		}
	}

	return m.Compose(fsp)
}

// Compose produces a composition of the receiver defined by the features in fs. The result is stored
// in the receiver and all contributing sequences are modified.
func (m *Multi) Compose(fs feat.Set) error {
	m.Flush(seq.Start|seq.End, m.Alpha.Gap())

	for _, r := range m.Seq {
		err := sequtils.Compose(r, r, fs)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Multi) String() string {
	t := m.Consensus(false)
	return t.String()
}

// Consensus returns a quality sequence reflecting the consensus of the receiver determined by the
// Consensify field.
func (m *Multi) Consensus(includeMissing bool) *protein.QSeq {
	cm := make([]alphabet.QLetter, 0, m.Len())
	alpha := m.Alphabet()
	for i := m.Start(); i < m.End(); i++ {
		cm = append(cm, m.Consensify(m, alpha, i, includeMissing))
	}

	c := protein.NewQSeq("Consensus:"+m.ID, cm, m.Alpha, m.Encode)
	c.SetOffset(m.Offset)

	return c
}
