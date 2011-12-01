// Package to represent k-mer sequences as a color
package kmercolor
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
	"github.com/kortschak/BioGo/graphics/color"
	"github.com/kortschak/BioGo/index/kmerindex"
	"github.com/kortschak/BioGo/util"
	"image"
)

const (
	H = 1 << iota
	S
	V
	A
)

type KmerColor struct {
	kmer, kmask kmerindex.Kmer
	*color.HSVAColor
}

func New(k int) (c *KmerColor) {
	return &KmerColor{
		kmask:     kmerindex.Kmer(util.Pow4(k) - 1),
		HSVAColor: &color.HSVAColor{},
	}
}

func (self *KmerColor) Kmer(kmer kmerindex.Kmer) *KmerColor {
	self.kmer = kmer
	return self
}

func (self *KmerColor) S(s float32) *KmerColor {
	self.HSVAColor.S = s
	return self
}

func (self *KmerColor) V(v float32) *KmerColor {
	self.HSVAColor.V = v
	return self
}

func (self *KmerColor) A(a float32) *KmerColor {
	self.HSVAColor.A = a
	return self
}

func (self *KmerColor) RGBA() (r, g, b, a uint32) {
	self.HSVAColor.H = float32(self.kmer) / float32(self.kmask)
	return self.HSVAColor.RGBA()
}

type KmerRainbow struct {
	*image.RGBA
	Index      *kmerindex.Index
	Max        int
	BackGround *color.HSVAColor
}

func NewKmerRainbow(r image.Rectangle, index *kmerindex.Index, color *color.HSVAColor) *KmerRainbow { // should generalise the BG color
	h := r.Dy()
	kmers := make([]int, h)
	kmask := util.Pow4(index.GetK())
	kmaskf := float32(kmask)
	f := func(index *kmerindex.Index, j, kmer int) {
		kmers[int(float32(kmer)*float32(h)/kmaskf)]++
	}
	index.ForEachKmerOf(index.Seq, 0, index.Seq.Len(), f)
	max := util.Max(kmers...)

	return &KmerRainbow{
		RGBA:       image.NewRGBA(r),
		Index:      index,
		Max:        max,
		BackGround: color,
	}
}

func (self *KmerRainbow) SubImage(r image.Rectangle) image.Image {
	return &KmerRainbow{
		RGBA: &image.RGBA{
			Pix:    self.RGBA.Pix,
			Stride: self.RGBA.Stride,
			Rect:   self.Rect.Intersect(r),
		},
		Max:        self.Max,
		Index:      self.Index,
		BackGround: self.BackGround,
	}
}

func (self *KmerRainbow) Paint(vary int, start, size, left, right int) (i *image.RGBA, err error) {
	right = util.Min(right, self.Rect.Dx())
	kmers := make([]uint32, self.RGBA.Rect.Dy())
	kmask := util.Pow4(self.Index.GetK())
	kmaskf := float32(kmask)
	f := func(index *kmerindex.Index, j, kmer int) {
		kmers[int(float32(kmer)*float32(self.RGBA.Rect.Dy())/kmaskf)]++
	}
	self.Index.ForEachKmerOf(self.Index.Seq, start*size, (start+1)*size-1, f)
	c := &color.HSVAColor{}
	lf := float32(len(kmers)) / 360
	var val float32
	scale := 1 / float32(self.Max)
	for y, v := range kmers {
		val = float32(v) / scale
		c.H = float32(y) / lf
		if vary&S != 0 {
			c.S = val
		} else {
			c.S = self.BackGround.S
		}
		if vary&V != 0 {
			c.V = val
		} else {
			c.V = self.BackGround.V
		}
		if vary&A != 0 {
			c.A = val
		} else {
			c.A = self.BackGround.A
		}
		if left >= 0 && right > left {
			for x := left; x < right; x++ {
				self.Set(x, y, c)
			}
		} else {
			println(left, right)
			for x := 0; x < self.Rect.Dx(); x++ {
				self.Set(x, y, c)
			}
		}
	}

	return self.RGBA, nil
}

type CGR KmerRainbow

func NewCGR(index *kmerindex.Index, color *color.HSVAColor) *CGR { // should generalise the BG color
	max := 0
	f := func(index *kmerindex.Index, j, kmer int) {
		if freq := index.FingerAt(kmer); freq > max {
			max = freq
		}
	}
	index.ForEachKmerOf(index.Seq, 0, index.Seq.Len(), f)

	k := uint(index.GetK())

	return &CGR{
		RGBA:       image.NewRGBA(image.Rect(0, 0, 1<<k, 1<<k)),
		Index:      index,
		Max:        max,
		BackGround: color,
	}
}

func (self *CGR) SubImage(r image.Rectangle) image.Image {
	return &CGR{
		RGBA: &image.RGBA{
			Pix:    self.RGBA.Pix,
			Stride: self.RGBA.Stride,
			Rect:   self.Rect.Intersect(r),
		},
		Max:        self.Max,
		Index:      self.Index,
		BackGround: self.BackGround,
	}
}

func (self *CGR) Paint(vary int, desch bool, start, size int) (i *image.RGBA, err error) {
	k := self.Index.GetK()

	kmask := util.Pow4(k)
	kmers := make([]uint, kmask)
	f := func(index *kmerindex.Index, j, kmer int) {
		kmers[kmer]++
	}
	self.Index.ForEachKmerOf(self.Index.Seq, start*size, (start+1)*size-1, f)

	c := &color.HSVAColor{}
	max := util.UMax(kmers...)
	scale := 1 / float32(max)

	for kmer, v := range kmers {
		x, y := 0, 0
		if desch {
			xdiff := 0
			for i, km := k-1, kmer; i >= 0; i, km = i-1, km>>2 {
				xdiff = ((km & 2) >> 1)
				x += xdiff << uint(i)
				y += ((km & 1) ^ (xdiff ^ 1)) << uint(i)
			}
		} else {
			for i, km := k-1, kmer; i >= 0; i, km = i-1, km>>2 {
				x += (km & 1) << uint(i)
				y += (((km & 1) ^ ((km & 2) >> 1)) ^ 1) << uint(i)
			}
		}
		val := float32(v) * scale
		if vary&H != 0 {
			c.H = val * 240
		} else {
			c.H = self.BackGround.H
		}
		if vary&S != 0 {
			c.S = val
		} else {
			c.S = self.BackGround.S
		}
		if vary&V != 0 {
			c.V = val
		} else {
			c.V = self.BackGround.V
		}
		if vary&A != 0 {
			c.A = val
		} else {
			c.A = self.BackGround.A
		}
		self.Set(x, y, c)
	}

	return self.RGBA, nil
}
