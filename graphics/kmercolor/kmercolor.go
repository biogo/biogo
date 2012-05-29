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

// Package to represent k-mer sequences as a color
package kmercolor

import (
	"github.com/kortschak/biogo/graphics/color"
	"github.com/kortschak/biogo/index/kmerindex"
	"github.com/kortschak/biogo/util"
	"image"
	imagecolor "image/color"
)

const (
	H = 1 << iota
	S
	V
	A
)

// A KmerColor represents a kmerindex.Kmer as an HSVA, mapping the numberical value of the Kmer to a hue.
type KmerColor struct {
	kmer, kmask kmerindex.Kmer
	color.HSVA
	low, high float64
}

// Make a new KmerColor initialising the Kmer length to k.
func New(k int) (c *KmerColor) {
	return &KmerColor{
		kmask: kmerindex.Kmer(util.Pow4(k) - 1),
	}
}

// Set the Kmer to kmer and return the receiver.
func (self *KmerColor) Kmer(kmer kmerindex.Kmer) *KmerColor {
	self.kmer = kmer
	return self
}

// Define the range of hues used by the KmerColor. imagecolor.Color is an alias to the core library image/color package to avoid a name conflict.
func (self *KmerColor) ColorRange(low, high imagecolor.Color) {
	self.low = color.RGBAtoHSVA(low.RGBA()).H
	self.high = color.RGBAtoHSVA(high.RGBA()).H
}

// Set the S value of the underlying HSVA and return the reciever.
func (self *KmerColor) S(s float64) *KmerColor {
	self.HSVA.S = s
	return self
}

// Set the V value of the underlying HSVA and return the reciever.
func (self *KmerColor) V(v float64) *KmerColor {
	self.HSVA.V = v
	return self
}

// Set the A value of the underlying HSVA and return the reciever.
func (self *KmerColor) A(a float64) *KmerColor {
	self.HSVA.A = a
	return self
}

// Satisfy the color.Color interface. RGBA maps the kmer value to a color with the hue between the low and high color values.
func (self *KmerColor) RGBA() (r, g, b, a uint32) {
	if self.high >= self.low {
		self.HSVA.H = ((self.high - self.low) * float64(self.kmer) / float64(self.kmask)) + self.low
	} else {
		self.HSVA.H = self.high - ((self.high - self.low) * float64(self.kmer) / float64(self.kmask))
	}
	return self.HSVA.RGBA()
}

// A KmerRainbow produces an image reflecting the kmer distribution of a sequence.
type KmerRainbow struct {
	*image.RGBA
	Index      *kmerindex.Index
	Max        int
	BackGround color.HSVA
}

// Create a new KmerRainbow defined by the rectangle r, kmerindex index and background color.
func NewKmerRainbow(r image.Rectangle, index *kmerindex.Index, background color.HSVA) *KmerRainbow { // should generalise the BG color
	h := r.Dy()
	kmers := make([]int, h)
	kmask := util.Pow4(index.GetK())
	kmaskf := float64(kmask)
	f := func(index *kmerindex.Index, _, kmer int) {
		kmers[int(float64(kmer)*float64(h)/kmaskf)]++
	}
	index.ForEachKmerOf(index.Seq, 0, index.Seq.Len(), f)
	max := util.Max(kmers...)

	return &KmerRainbow{
		RGBA:       image.NewRGBA(r),
		Index:      index,
		Max:        max,
		BackGround: background,
	}
}

// SubImage returns an image representing the portion of the original image visible through r. The returned value shares pixels with the original image.
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

// Render the rainbow based on block of sequence in the index with the given size. Left and right define the extent of the rendering.
// Vary specifies which color values change in response to kmer frequency.
func (self *KmerRainbow) Paint(vary int, block, size, left, right int) (i *image.RGBA, err error) {
	right = util.Min(right, self.Rect.Dx())
	kmers := make([]uint32, self.RGBA.Rect.Dy())
	kmask := util.Pow4(self.Index.GetK())
	kmaskf := float64(kmask)
	f := func(index *kmerindex.Index, _, kmer int) {
		kmers[int(float64(kmer)*float64(self.RGBA.Rect.Dy())/kmaskf)]++
	}
	self.Index.ForEachKmerOf(self.Index.Seq, block*size, (block+1)*size-1, f)
	c := color.HSVA{}
	lf := float64(len(kmers)) / 360
	var val float64
	scale := 1 / float64(self.Max)
	for y, v := range kmers {
		val = float64(v) / scale
		c.H = float64(y) / lf
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

// A CGR produces a Chaos Game Representation of a sequence. Deschavanne (1999).
type CGR KmerRainbow

// Create a new CGR defined by the kmerindex index and background color.
func NewCGR(index *kmerindex.Index, background color.HSVA) *CGR { // should generalise the BG color
	max := 0
	f := func(index *kmerindex.Index, _, kmer int) {
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
		BackGround: background,
	}
}

// SubImage returns an image representing the portion of the original image visible through r. The returned value shares pixels with the original image.
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

// Render the rainbow based on block of sequence in the index with the given size.
// Vary specifies which color values change in response to kmer frequency. Setting desch to true 
// specifies using the ordering described in Deschavanne (1999).
func (self *CGR) Paint(vary int, desch bool, block, size int) (i *image.RGBA, err error) {
	k := self.Index.GetK()

	kmask := util.Pow4(k)
	kmers := make([]uint, kmask)
	f := func(index *kmerindex.Index, _, kmer int) {
		kmers[kmer]++
	}
	self.Index.ForEachKmerOf(self.Index.Seq, block*size, (block+1)*size-1, f)

	c := &color.HSVA{}
	max := util.UMax(kmers...)
	scale := 1 / float64(max)

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
		val := float64(v) * scale
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
