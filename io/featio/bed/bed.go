// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package bed provides types to read and write BED format files according to
// the UCSC specification.
//
// The specification can be found at http://genome.ucsc.edu/FAQ/FAQformat.html#format1.
package bed

import (
	"code.google.com/p/biogo/exp/feat"
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/io/featio"

	"bufio"
	"bytes"
	"errors"
	"fmt"
	"image/color"
	"io"
	"reflect"
	"strconv"
	"unsafe"
)

var (
	ErrBadBedType         = errors.New("bed: bad bed type")
	ErrBadStrandField     = errors.New("bed: bad strand field")
	ErrBadStrand          = errors.New("bed: invalid strand")
	ErrBadColorField      = errors.New("bed: bad color field")
	ErrMissingBlockValues = errors.New("bed: missing block values")
	ErrNoChromField       = errors.New("bed: no chrom field available")
	ErrClosed             = errors.New("bed: writer closed")
)

const (
	chromField = iota
	startField
	endField
	nameField
	scoreField
	strandField
	thickStartField
	thickEndField
	rgbField
	blockCountField
	blockSizesField
	blockStartsField
)

var (
	_ featio.Reader = &Reader{}
	_ featio.Writer = &Writer{}

	_ feat.Feature = &Bed3{}
	_ feat.Feature = &Bed4{}
	_ feat.Feature = &Bed5{}
	_ feat.Feature = &Bed6{}
	_ feat.Feature = &Bed12{}

	_ Bed = &Bed3{}
	_ Bed = &Bed4{}
	_ Bed = &Bed5{}
	_ Bed = &Bed6{}
	_ Bed = &Bed12{}

	_ feat.Orienter = &Bed6{}
	_ feat.Orienter = &Bed12{}
)

type Bed interface {
	feat.Feature
	canBed(int) bool
}

func handlePanic(f feat.Feature, err *error) {
	r := recover()
	if r != nil {
		e, ok := r.(error)
		if !ok {
			panic(r)
		}
		*err = e
		if f != nil {
			f = nil
		}
	}
}

// This function cannot be used to create strings that are expected to persist.
func unsafeString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func mustAtoi(f []byte) int {
	i, err := strconv.ParseInt(unsafeString(f), 0, 0)
	if err != nil {
		panic(err)
	}
	return int(i)
}

func mustAtob(f []byte) byte {
	b, err := strconv.ParseUint(unsafeString(f), 0, 8)
	if err != nil {
		panic(err)
	}
	return byte(b)
}

var charToStrand = func() [256]seq.Strand {
	var t [256]seq.Strand
	for i := range t {
		t[i] = 0x7f
	}
	t['+'] = seq.Plus
	t['.'] = seq.None
	t['-'] = seq.Minus
	return t
}()

func mustAtos(f []byte) seq.Strand {
	if len(f) != 1 {
		panic(ErrBadStrandField)
	}
	s := charToStrand[f[0]]
	if s == 0x7f {
		panic(ErrBadStrand)
	}
	return s
}

func mustAtoRgb(f []byte) color.RGBA {
	c := bytes.SplitN(f, []byte{','}, 4)
	l := len(c)
	if l == 0 || (l == 1 && mustAtoi(c[0]) == 0) {
		return color.RGBA{}
	}
	if l < 3 {
		panic(ErrBadColorField)
	}
	return color.RGBA{
		R: mustAtob(c[0]),
		G: mustAtob(c[1]),
		B: mustAtob(c[2]),
		A: 0xff,
	}
}

func mustAtoa(f []byte) []int {
	c := bytes.Split(f, []byte{','})
	a := make([]int, len(c))
	for i, f := range c {
		if len(f) == 0 {
			return a[:i]
		}
		a[i] = mustAtoi(f)
	}
	return a
}

type Chrom string

func (c Chrom) Start() int             { return 0 }
func (c Chrom) End() int               { return 0 }
func (c Chrom) Len() int               { return 0 }
func (c Chrom) Name() string           { return string(c) }
func (c Chrom) Description() string    { return "bed chrom" }
func (c Chrom) Location() feat.Feature { return nil }

type Bed3 struct {
	Chrom      string
	ChromStart int
	ChromEnd   int
}

func parseBed3(line []byte) (b *Bed3, err error) {
	const n = 3
	defer handlePanic(b, &err)
	f := bytes.SplitN(line, []byte{'\t'}, n+1)
	if len(f) < n {
		return nil, ErrBadBedType
	}
	b = &Bed3{
		Chrom:      string(f[chromField]),
		ChromStart: mustAtoi(f[startField]),
		ChromEnd:   mustAtoi(f[endField]),
	}
	return
}

func (b *Bed3) Start() int                  { return b.ChromStart }
func (b *Bed3) End() int                    { return b.ChromEnd }
func (b *Bed3) Len() int                    { return b.ChromEnd - b.ChromStart }
func (b *Bed3) Name() string                { return fmt.Sprintf("%s:[%d,%d)", b.Chrom, b.ChromStart, b.ChromEnd) }
func (b *Bed3) Description() string         { return "bed3 feature" }
func (b *Bed3) Location() feat.Feature      { return Chrom(b.Chrom) }
func (b *Bed3) canBed(i int) bool           { return i <= 3 }
func (b *Bed3) Format(fs fmt.State, c rune) { format(b, fs, c) }

type Bed4 struct {
	Chrom      string
	ChromStart int
	ChromEnd   int
	FeatName   string
}

func parseBed4(line []byte) (b *Bed4, err error) {
	const n = 4
	defer handlePanic(b, &err)
	f := bytes.SplitN(line, []byte{'\t'}, n+1)
	if len(f) < n {
		return nil, ErrBadBedType
	}
	b = &Bed4{
		Chrom:      string(f[chromField]),
		ChromStart: mustAtoi(f[startField]),
		ChromEnd:   mustAtoi(f[endField]),
		FeatName:   string(f[nameField]),
	}
	return
}

func (b *Bed4) Start() int                  { return b.ChromStart }
func (b *Bed4) End() int                    { return b.ChromEnd }
func (b *Bed4) Len() int                    { return b.ChromEnd - b.ChromStart }
func (b *Bed4) Name() string                { return b.FeatName }
func (b *Bed4) Description() string         { return "bed4 feature" }
func (b *Bed4) Location() feat.Feature      { return Chrom(b.Chrom) }
func (b *Bed4) canBed(i int) bool           { return i <= 4 }
func (b *Bed4) Format(fs fmt.State, c rune) { format(b, fs, c) }

type Bed5 struct {
	Chrom      string
	ChromStart int
	ChromEnd   int
	FeatName   string
	FeatScore  int
}

func parseBed5(line []byte) (b *Bed5, err error) {
	const n = 5
	defer handlePanic(b, &err)
	f := bytes.SplitN(line, []byte{'\t'}, n+1)
	if len(f) < n {
		return nil, ErrBadBedType
	}
	b = &Bed5{
		Chrom:      string(f[chromField]),
		ChromStart: mustAtoi(f[startField]),
		ChromEnd:   mustAtoi(f[endField]),
		FeatName:   string(f[nameField]),
		FeatScore:  mustAtoi(f[scoreField]),
	}
	return
}

func (b *Bed5) Start() int                  { return b.ChromStart }
func (b *Bed5) End() int                    { return b.ChromEnd }
func (b *Bed5) Len() int                    { return b.ChromEnd - b.ChromStart }
func (b *Bed5) Name() string                { return b.FeatName }
func (b *Bed5) Description() string         { return "bed5 feature" }
func (b *Bed5) Location() feat.Feature      { return Chrom(b.Chrom) }
func (b *Bed5) canBed(i int) bool           { return i <= 5 }
func (b *Bed5) Format(fs fmt.State, c rune) { format(b, fs, c) }

type Bed6 struct {
	Chrom      string
	ChromStart int
	ChromEnd   int
	FeatName   string
	FeatScore  int
	FeatStrand seq.Strand
}

func parseBed6(line []byte) (b *Bed6, err error) {
	const n = 6
	defer handlePanic(b, &err)
	f := bytes.SplitN(line, []byte{'\t'}, n+1)
	if len(f) < n {
		return nil, ErrBadBedType
	}
	b = &Bed6{
		Chrom:      string(f[chromField]),
		ChromStart: mustAtoi(f[startField]),
		ChromEnd:   mustAtoi(f[endField]),
		FeatName:   string(f[nameField]),
		FeatScore:  mustAtoi(f[scoreField]),
		FeatStrand: mustAtos(f[strandField]),
	}
	return
}

func (b *Bed6) Start() int                    { return b.ChromStart }
func (b *Bed6) End() int                      { return b.ChromEnd }
func (b *Bed6) Len() int                      { return b.ChromEnd - b.ChromStart }
func (b *Bed6) Name() string                  { return b.FeatName }
func (b *Bed6) Description() string           { return "bed6 feature" }
func (b *Bed6) Location() feat.Feature        { return Chrom(b.Chrom) }
func (b *Bed6) Orientation() feat.Orientation { return feat.Orientation(b.FeatStrand) }
func (b *Bed6) canBed(i int) bool             { return i <= 6 }
func (b *Bed6) Format(fs fmt.State, c rune)   { format(b, fs, c) }

type Bed12 struct {
	Chrom       string
	ChromStart  int
	ChromEnd    int
	FeatName    string
	FeatScore   int
	FeatStrand  seq.Strand
	ThickStart  int
	ThickEnd    int
	Rgb         color.RGBA
	BlockCount  int
	BlockSizes  []int
	BlockStarts []int
}

func parseBed12(line []byte) (b *Bed12, err error) {
	const n = 12
	defer handlePanic(b, &err)
	f := bytes.SplitN(line, []byte{'\t'}, n+1)
	if len(f) < n {
		return nil, ErrBadBedType
	}
	b = &Bed12{
		Chrom:       string(f[chromField]),
		ChromStart:  mustAtoi(f[startField]),
		ChromEnd:    mustAtoi(f[endField]),
		FeatName:    string(f[nameField]),
		FeatScore:   mustAtoi(f[scoreField]),
		FeatStrand:  mustAtos(f[strandField]),
		ThickStart:  mustAtoi(f[thickStartField]),
		ThickEnd:    mustAtoi(f[thickEndField]),
		Rgb:         mustAtoRgb(f[rgbField]),
		BlockCount:  mustAtoi(f[blockCountField]),
		BlockSizes:  mustAtoa(f[blockSizesField]),
		BlockStarts: mustAtoa(f[blockStartsField]),
	}
	if b.BlockCount != len(b.BlockSizes) || b.BlockCount != len(b.BlockStarts) {
		return nil, ErrMissingBlockValues
	}
	return
}

func (b *Bed12) Start() int                    { return b.ChromStart }
func (b *Bed12) End() int                      { return b.ChromEnd }
func (b *Bed12) Len() int                      { return b.ChromEnd - b.ChromStart }
func (b *Bed12) Name() string                  { return b.FeatName }
func (b *Bed12) Description() string           { return "bed12 feature" }
func (b *Bed12) Location() feat.Feature        { return Chrom(b.Chrom) }
func (b *Bed12) Orientation() feat.Orientation { return feat.Orientation(b.FeatStrand) }
func (b *Bed12) canBed(i int) bool             { return i <= 12 }
func (b *Bed12) Format(fs fmt.State, c rune)   { format(b, fs, c) }

// BED format reader type.
type Reader struct {
	r       *bufio.Reader
	BedType int
	line    int
}

// Returns a new BED format reader using r.
func NewReader(r io.Reader, b int) *Reader {
	return &Reader{
		r:       bufio.NewReader(r),
		BedType: b,
	}
}

// Read a single feature and return it or an error.
func (r *Reader) Read() (f feat.Feature, err error) {
	line, err := r.r.ReadBytes('\n')
	if err != nil {
		return
	}
	r.line++
	line = bytes.TrimSpace(line)

	switch r.BedType {
	case 3:
		f, err = parseBed3(line)
	case 4:
		f, err = parseBed4(line)
	case 5:
		f, err = parseBed5(line)
	case 6:
		f, err = parseBed6(line)
	case 12:
		f, err = parseBed12(line)
	}
	if err != nil {
		return nil, fmt.Errorf("%v at line %d", err, r.line)
	}

	return
}

// Return the current line number
func (r *Reader) Line() int { return r.line }

func format(b Bed, fs fmt.State, c rune) {
	bv := reflect.ValueOf(b)
	if bv.IsNil() {
		fmt.Fprint(fs, "<nil>")
		return
	}
	bv = bv.Elem()
	switch c {
	case 'v':
		if fs.Flag('#') {
			fmt.Fprintf(fs, "&%#v", bv.Interface())
			return
		}
		fallthrough
	case 's':
		width, _ := fs.Width()
		if !b.canBed(width) {
			fmt.Fprintf(fs, "%%!(BADWIDTH)%T", width, b)
			return
		}
		if width == 0 {
			width = bv.NumField()
		}
		for i := 0; i < width; i++ {
			f := bv.Field(i).Interface()
			if i >= rgbField {
				switch i {
				case rgbField:
					rv := reflect.ValueOf(f)
					if reflect.DeepEqual(rv.Interface(), color.RGBA{}) {
						fs.Write([]byte{'0'})
					} else {
						fmt.Fprintf(fs, "%d,%d,%d",
							rv.Field(0).Interface(), rv.Field(1).Interface(), rv.Field(2).Interface())
					}
				case blockCountField:
					fmt.Fprint(fs, f)
				case blockSizesField, blockStartsField:
					av := reflect.ValueOf(f)
					l := av.Len()
					for j := 0; j < l; j++ {
						fmt.Fprint(fs, av.Index(j).Interface())
						if j < l-1 {
							fs.Write([]byte{','})
						}
					}
				}
			} else {
				fmt.Fprint(fs, f)
			}
			if i < width-1 {
				fs.Write([]byte{'\t'})
			}
		}
	default:
		fmt.Fprintf(fs, "%%!%c(%T=%3s)", c, b)
	}
}

// BED format writer type.
type Writer struct {
	w       *bufio.Writer
	closed  bool
	BedType int
}

// Returns a new BED format writer using f.
func NewWriter(f io.Writer, b int) *Writer {
	return &Writer{
		w:       bufio.NewWriter(f),
		BedType: b,
	}
}

type Scorer interface {
	Score() int
}

// Write a single feature and return the number of bytes written and any error.
func (w *Writer) Write(f feat.Feature) (n int, err error) {
	if w.closed {
		return 0, ErrClosed
	}
	defer func() {
		if err != nil {
			return
		}
		err = w.w.WriteByte('\n')
		if err != nil {
			return
		}
		n++
	}()

	// Handle Bed types.
	if f, ok := f.(Bed); ok {
		if !f.canBed(w.BedType) {
			return 0, ErrBadBedType
		}
		return fmt.Fprintf(w.w, "%*s", w.BedType, f)
	}

	// Handle other feature types.
	if f.Location() == nil {
		return 0, ErrNoChromField
	}

	// Bed3
	n, err = fmt.Fprintf(w.w, "%s\t%d\t%d", f.Location(), f.Start(), f.End())
	if w.BedType == 3 {
		return n, err
	}

	// Bed4
	in, err := fmt.Fprintf(w.w, "\t%s", f.Name())
	n += in
	if w.BedType == 4 || err != nil {
		return n, err
	}

	// Bed5
	if f, ok := f.(Scorer); ok {
		in, err := fmt.Fprintf(w.w, "\t%d", f.Score())
		n += in
		if err != nil {
			return n, err
		}
	} else {
		in, err := fmt.Fprintf(w.w, "\t0")
		n += in
		if err != nil {
			return n, err
		}
	}
	if w.BedType == 5 {
		return
	}

	// Bed6
	if f, ok := f.(feat.Orienter); ok {
		in, err := fmt.Fprintf(w.w, "\t%s", seq.Strand(f.Orientation()))
		n += in
		if err != nil {
			return n, err
		}
	} else {
		in, err := fmt.Fprintf(w.w, "\t.")
		n += in
		if err != nil {
			return n, err
		}
	}
	if w.BedType == 6 || w.BedType == 0 {
		return
	}

	// Don't handle Bed12.
	w.w.Write([]byte{'\n'})
	return n, ErrBadBedType
}

// Close closes the Writer. The underlying io.Writer is not closed.
func (w *Writer) Close() error {
	if w.closed {
		return nil
	}
	w.closed = true
	return w.w.Flush()
}
