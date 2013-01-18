// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package kmercolor provide the capacity to represent k-mer sequences as colors.
package kmercolor

import (
	"code.google.com/p/biogo/graphics/color"
	"code.google.com/p/biogo/index/kmerindex"
	"code.google.com/p/biogo/util"
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
func New(k int) *KmerColor {
	return &KmerColor{
		kmask: kmerindex.Kmer(util.Pow4(k) - 1),
	}
}

// Set the Kmer to kmer and return the receiver.
func (c *KmerColor) Kmer(kmer kmerindex.Kmer) *KmerColor {
	c.kmer = kmer
	return c
}

// Define the range of hues used by the KmerColor. imagecolor.Color is an alias to the core library image/color package to avoid a name conflict.
func (c *KmerColor) ColorRange(low, high imagecolor.Color) {
	c.low = color.RGBAtoHSVA(low.RGBA()).H
	c.high = color.RGBAtoHSVA(high.RGBA()).H
}

// Set the S value of the underlying HSVA and return the reciever.
func (c *KmerColor) S(s float64) *KmerColor {
	c.HSVA.S = s
	return c
}

// Set the V value of the underlying HSVA and return the reciever.
func (c *KmerColor) V(v float64) *KmerColor {
	c.HSVA.V = v
	return c
}

// Set the A value of the underlying HSVA and return the reciever.
func (c *KmerColor) A(a float64) *KmerColor {
	c.HSVA.A = a
	return c
}

// Satisfy the color.Color interface. RGBA maps the kmer value to a color with the hue between the low and high color values.
func (c *KmerColor) RGBA() (r, g, b, a uint32) {
	if c.high >= c.low {
		c.HSVA.H = ((c.high - c.low) * float64(c.kmer) / float64(c.kmask)) + c.low
	} else {
		c.HSVA.H = c.high - ((c.high - c.low) * float64(c.kmer) / float64(c.kmask))
	}
	return c.HSVA.RGBA()
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
	s := index.GetSeq()
	index.ForEachKmerOf(s, 0, s.Len(), f)
	max := util.Max(kmers...)

	return &KmerRainbow{
		RGBA:       image.NewRGBA(r),
		Index:      index,
		Max:        max,
		BackGround: background,
	}
}

// SubImage returns an image representing the portion of the original image visible through r. The returned value shares pixels with the original image.
func (kr *KmerRainbow) SubImage(r image.Rectangle) image.Image {
	return &KmerRainbow{
		RGBA: &image.RGBA{
			Pix:    kr.RGBA.Pix,
			Stride: kr.RGBA.Stride,
			Rect:   kr.Rect.Intersect(r),
		},
		Max:        kr.Max,
		Index:      kr.Index,
		BackGround: kr.BackGround,
	}
}

// Render the rainbow based on block of sequence in the index with the given size. Left and right define the extent of the rendering.
// Vary specifies which color values change in response to kmer frequency.
func (kr *KmerRainbow) Paint(vary int, block, size, left, right int) (i *image.RGBA, err error) {
	right = util.Min(right, kr.Rect.Dx())
	kmers := make([]uint32, kr.RGBA.Rect.Dy())
	kmask := util.Pow4(kr.Index.GetK())
	kmaskf := float64(kmask)
	f := func(index *kmerindex.Index, _, kmer int) {
		kmers[int(float64(kmer)*float64(kr.RGBA.Rect.Dy())/kmaskf)]++
	}
	kr.Index.ForEachKmerOf(kr.Index.GetSeq(), block*size, (block+1)*size-1, f)
	c := color.HSVA{}
	lf := float64(len(kmers)) / 360
	var val float64
	scale := 1 / float64(kr.Max)
	for y, v := range kmers {
		val = float64(v) / scale
		c.H = float64(y) / lf
		if vary&S != 0 {
			c.S = val
		} else {
			c.S = kr.BackGround.S
		}
		if vary&V != 0 {
			c.V = val
		} else {
			c.V = kr.BackGround.V
		}
		if vary&A != 0 {
			c.A = val
		} else {
			c.A = kr.BackGround.A
		}
		if left >= 0 && right > left {
			for x := left; x < right; x++ {
				kr.Set(x, y, c)
			}
		} else {
			println(left, right)
			for x := 0; x < kr.Rect.Dx(); x++ {
				kr.Set(x, y, c)
			}
		}
	}

	return kr.RGBA, nil
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
	s := index.GetSeq()
	index.ForEachKmerOf(s, 0, s.Len(), f)

	k := uint(index.GetK())

	return &CGR{
		RGBA:       image.NewRGBA(image.Rect(0, 0, 1<<k, 1<<k)),
		Index:      index,
		Max:        max,
		BackGround: background,
	}
}

// SubImage returns an image representing the portion of the original image visible through r. The returned value shares pixels with the original image.
func (cg *CGR) SubImage(r image.Rectangle) image.Image {
	return &CGR{
		RGBA: &image.RGBA{
			Pix:    cg.RGBA.Pix,
			Stride: cg.RGBA.Stride,
			Rect:   cg.Rect.Intersect(r),
		},
		Max:        cg.Max,
		Index:      cg.Index,
		BackGround: cg.BackGround,
	}
}

// Render the rainbow based on block of sequence in the index with the given size.
// Vary specifies which color values change in response to kmer frequency. Setting desch to true 
// specifies using the ordering described in Deschavanne (1999).
func (cg *CGR) Paint(vary int, desch bool, block, size int) (i *image.RGBA, err error) {
	k := cg.Index.GetK()

	kmask := util.Pow4(k)
	kmers := make([]uint, kmask)
	f := func(index *kmerindex.Index, _, kmer int) {
		kmers[kmer]++
	}
	cg.Index.ForEachKmerOf(cg.Index.GetSeq(), block*size, (block+1)*size-1, f)

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
			c.H = cg.BackGround.H
		}
		if vary&S != 0 {
			c.S = val
		} else {
			c.S = cg.BackGround.S
		}
		if vary&V != 0 {
			c.V = val
		} else {
			c.V = cg.BackGround.V
		}
		if vary&A != 0 {
			c.A = val
		} else {
			c.A = cg.BackGround.A
		}
		cg.Set(x, y, c)
	}

	return cg.RGBA, nil
}
