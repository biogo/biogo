// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sequtils provides generic functions for manipulation of biogo/seq/... types.
package sequtils

import (
	"github.com/biogo/biogo/alphabet"
	"github.com/biogo/biogo/feat"
	"github.com/biogo/biogo/seq"

	"sort"

	"github.com/biogo/biogo/errors"
)

// A Joinable can be joined to another of the same concrete type using the Join function.
type Joinable interface {
	SetOffset(int) error
	Slice() alphabet.Slice
	SetSlice(alphabet.Slice)
}

// Join joins a and be with the target location of b specified by where. The offset of dst
// will be updated if src is prepended. Join will panic if dst and src do not hold the same
// concrete Slice type. Circular sequences cannot be joined.
func Join(dst, src Joinable, where int) error {
	dstC, dstOk := dst.(seq.Conformationer)
	srcC, srcOk := src.(seq.Conformationer)
	switch {
	case dstOk && dstC.Conformation() > feat.Linear, srcOk && srcC.Conformation() > feat.Linear:
		return errors.ArgErr{}.Make("sequtils: cannot join circular sequence")
	}

	o := dst
	if where == seq.End {
		src, dst = dst, src
	}
	dstSl, srcSl := dst.Slice(), src.Slice()
	srcLen := srcSl.Len()
	if where == seq.Start {
		dst.SetOffset(-srcLen)
	}
	t := dst.Slice().Make(srcLen, srcLen+dstSl.Len())
	t.Copy(srcSl)
	o.SetSlice(t.Append(dstSl))
	return nil
}

// A Sliceable can be truncated, stitched and composed.
type Sliceable interface {
	Start() int
	End() int
	SetOffset(int) error
	Slice() alphabet.Slice
	SetSlice(alphabet.Slice)
}

// Truncate performs a truncation on src from start to end and places the result in dst.
// The conformation of dst is set to linear and the offset is set to start. If dst and src
// are not equal, a copy of the truncation is allocated. Only circular sequences can be
// truncated with start > end.
func Truncate(dst, src Sliceable, start, end int) error {
	var (
		sl     = src.Slice()
		offset = src.Start()
	)
	if start < offset || end > src.End() {
		return errors.ArgErr{}.Make("sequtils: index out of range")
	}
	if start <= end {
		if dst == src {
			dst.SetSlice(sl.Slice(start-offset, end-offset))
		} else {
			dst.SetSlice(sl.Make(0, end-start).Append(sl.Slice(start-offset, end-offset)))
		}
		dst.SetOffset(start)
		if dst, ok := dst.(seq.ConformationSetter); ok {
			dst.SetConformation(feat.Linear)
		}
		return nil
	}

	if src, ok := src.(seq.Conformationer); !ok || src.Conformation() == feat.Linear {
		return errors.ArgErr{}.Make("sequtils: start position greater than end position for linear sequence")
	}
	if end < offset || start > src.End() {
		return errors.ArgErr{}.Make("sequtils: index out of range")
	}
	t := sl.Make(sl.Len()-start+offset, sl.Len()+end-start)
	t.Copy(sl.Slice(start-offset, sl.Len()))
	dst.SetSlice(t.Append(sl.Slice(0, end-offset)))
	dst.SetOffset(start)
	if dst, ok := dst.(seq.ConformationSetter); ok {
		dst.SetConformation(feat.Linear)
	}

	return nil
}

type feats []feat.Feature

func (f feats) Len() int           { return len(f) }
func (f feats) Less(i, j int) bool { return f[i].Start() < f[j].Start() }
func (f feats) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

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

// Stitch produces a subsequence of src defined by fs and places the the result in dst.
// The subsequences are guaranteed to be in order and non-overlapping even if not provided as such.
// Stitching a circular sequence returns a linear sequence.
func Stitch(dst, src Sliceable, fs feat.Set) error {
	var (
		sl     = src.Slice()
		offset = src.Start()
		ff     = feats(fs.Features())
	)
	for _, f := range ff {
		if f.End() < f.Start() {
			return errors.ArgErr{}.Make("sequtils: feature end < feature start")
		}
	}
	ff = append(feats(nil), ff...)
	sort.Sort(ff)
	// FIXME Does not correctly deal with circular sequences and feature sets.
	// Range over ff if src is circular and  and trunc at start and end, do checks to
	// see if feature splits on origin and rearrange tail to front.

	pLen := sl.Len()
	end := pLen + offset

	type fi struct{ s, e int }
	var (
		fsp = make([]*fi, 0, len(ff))
		csp *fi
	)
	for i, f := range ff {
		if s := f.Start(); i == 0 || s > csp.e {
			csp = &fi{s: s, e: f.End()}
			fsp = append(fsp, csp)
		} else {
			csp.e = max(csp.e, f.End())
		}
	}

	var l int
	for _, f := range fsp {
		l += max(0, min(f.e, end)-max(f.s, offset))
	}
	t := sl.Make(0, l)

	for _, f := range fsp {
		fs, fe := max(f.s-offset, 0), min(f.e-offset, pLen)
		if fs >= fe {
			continue
		}
		t = t.Append(sl.Slice(fs, fe))
	}

	dst.SetSlice(t)
	if dst, ok := dst.(seq.ConformationSetter); ok {
		dst.SetConformation(feat.Linear)
	}
	dst.SetOffset(0)

	return nil
}

type SliceReverser interface {
	Sliceable
	New() seq.Sequence
	Alphabet() alphabet.Alphabet
	SetAlphabet(alphabet.Alphabet) error
	RevComp()
	Reverse()
}

// Compose produces a composition of src defined by the features in fs. The subparts of
// the composition may be out of order and if features in fs specify orientation may be
// reversed or reverse complemented depending on the src - if src is a SliceReverser and
// its alphabet is a Complementor the segment will be reverse complemented, if the alphabte
// is not a Complementor these segments will only be reversed. If src is not a SliceREverser
// and a reverse segment is specified an error is returned.
// Composing a circular sequence returns a linear sequence.
func Compose(dst, src Sliceable, fs feat.Set) error {
	var (
		sl     = src.Slice()
		offset = src.Start()
		ff     = feats(fs.Features())
	)

	pLen := sl.Len()
	end := pLen + offset

	t := make([]alphabet.Slice, len(ff))
	var tl int
	for i, f := range ff {
		if f.End() < f.Start() {
			return errors.ArgErr{}.Make("sequtils: feature end < feature start")
		}
		l := min(f.End(), end) - max(f.Start(), offset)
		tl += l
		t[i] = sl.Make(l, l)
		t[i].Copy(sl.Slice(max(f.Start()-offset, 0), min(f.End()-offset, pLen)))
	}

	c := sl.Make(0, tl)
	var r SliceReverser
	for i, ts := range t {
		if f, ok := ff[i].(feat.Orienter); ok && f.Orientation() == feat.Reverse {
			switch src := src.(type) {
			case SliceReverser:
				if r == nil {
					r = src.New().(SliceReverser)
					if _, ok := src.Alphabet().(alphabet.Complementor); ok {
						r.SetAlphabet(src.Alphabet())
						r.SetSlice(ts)
						r.RevComp()
					} else {
						r.SetSlice(ts)
						r.Reverse()
					}
				}
			default:
				return errors.ArgErr{}.Make("sequtils: unable to reverse segment during compose")
			}
			c = c.Append(r.Slice())
		} else {
			c = c.Append(ts)
		}
	}

	dst.SetSlice(c)
	if dst, ok := dst.(seq.ConformationSetter); ok {
		dst.SetConformation(feat.Linear)
	}
	dst.SetOffset(0)

	return nil
}

// A QualityFeature describes a segment of sequence quality information. EAt() called with
// column values within Start() and End() is expected to return valid error probabilities for
// the zero'th row position.
type QualityFeature interface {
	feat.Feature
	EAt(int) float64
}

// Trim uses the modified-Mott trimming function to determine the start and end positions
// of good sequence. http://www.phrap.org/phredphrap/phred.html
func Trim(q QualityFeature, limit float64) (start, end int) {
	var sum, max float64
	for i := q.Start(); i < q.End(); i++ {
		sum += limit - q.EAt(i)
		if sum < 0 {
			sum, start = 0, i+1
		}
		if sum >= max {
			max, end = sum, i+1
		}
	}
	return
}
