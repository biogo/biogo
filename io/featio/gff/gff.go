// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gff provides types to read and write version 2 General Feature Format
// files according to the Sanger Institute specification.
//
// The specification can be found at http://www.sanger.ac.uk/resources/software/gff/spec.html.
package gff

import (
	"code.google.com/p/biogo/alphabet"
	"code.google.com/p/biogo/feat"
	"code.google.com/p/biogo/io/featio"
	"code.google.com/p/biogo/io/seqio/fasta"
	"code.google.com/p/biogo/seq"
	"code.google.com/p/biogo/seq/linear"

	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"strconv"
	"time"
	"unicode"
	"unsafe"
)

var (
	_ featio.Reader = (*Reader)(nil)
	_ featio.Writer = (*Writer)(nil)
)

// Version is the GFF version that is read and written.
const Version = 2

// "Astronomical" time format is the format specified in the GFF specification
const Astronomical = "2006-01-02"

type Error string

func (e Error) Error() string { return string(e) }

var (
	ErrBadStrandField = Error("gff: bad strand field")
	ErrBadStrand      = Error("gff: invalid strand")
	ErrClosed         = Error("gff: writer closed")
	ErrBadTag         = Error("gff: invalid tag")
	ErrCannotHeader   = Error("gff: cannot write header: data written")
	ErrNotHandled     = Error("gff: type not handled")
	ErrFieldMissing   = Error("gff: missing fields")
	ErrBadMoltype     = Error("gff: invalid moltype")
	ErrEmptyMetaLine  = Error("gff: empty comment metaline")
	ErrBadMetaLine    = Error("gff: incomplete metaline")
	ErrBadSequence    = Error("gff: corrupt metasequence")
)

const (
	nameField = iota
	sourceField
	featureField
	startField
	endField
	scoreField
	strandField
	frameField
	attributeField
	commentField
	lastField
)

// 
type Frame int8

func (f Frame) String() string {
	if f <= NoFrame || f > Frame2 {
		return "."
	}
	return [...]string{"0", "1", "2"}[f]
}

const (
	NoFrame Frame = iota - 1
	Frame0
	Frame1
	Frame2
)

// A Sequence is a feat.Feature
type Sequence struct {
	SeqName string
	Type    feat.Moltype
}

func (s Sequence) Start() int             { return 0 }
func (s Sequence) End() int               { return 0 }
func (s Sequence) Len() int               { return 0 }
func (s Sequence) Name() string           { return string(s.SeqName) }
func (s Sequence) Description() string    { return "GFF sequence" }
func (s Sequence) Location() feat.Feature { return nil }
func (s Sequence) MolType() feat.Moltype  { return s.Type }

// A Region is a feat.Feature
type Region struct {
	Sequence
	RegionStart int
	RegionEnd   int
}

func (r *Region) Start() int             { return r.RegionStart }
func (r *Region) End() int               { return r.RegionEnd }
func (r *Region) Len() int               { return r.RegionEnd - r.RegionStart }
func (r *Region) Description() string    { return "GFF region" }
func (r *Region) Location() feat.Feature { return r.Sequence }

// An Attribute represents a GFF2 attribute field record. Attribute field records
// must have an tag value structure following the syntax used within objects in a
// .ace file, flattened onto one line by semicolon separators.
// Tags must be standard identifiers ([A-Za-z][A-Za-z0-9_]*). Free text values
// must be quoted with double quotes.
//
// Note: all non-printing characters in free text value strings (e.g. newlines,
// tabs, control characters, etc) must be explicitly represented by their C (UNIX)
// style backslash-escaped representation.
type Attribute struct {
	Tag, Value string
}

type Attributes []Attribute

func (a Attributes) Get(tag string) string {
	for _, tv := range a {
		if tv.Tag == tag {
			return tv.Value
		}
	}
	return ""
}

func (a Attributes) Format(fs fmt.State, c rune) {
	for i, tv := range a {
		fmt.Fprintf(fs, "%s %s", tv.Tag, tv.Value)
		if i < len(a)-1 {
			fs.Write([]byte("; "))
		}
	}
}

// A Feature represents a standard GFF2 feature.
type Feature struct {
	// The name of the sequence. Having an explicit sequence name allows
	// a feature file to be prepared for a data set of multiple sequences.
	// Normally the seqname will be the identifier of the sequence in an
	// accompanying fasta format file. An alternative is that SeqName is
	// the identifier for a sequence in a public database, such as an
	// EMBL/Genbank/DDBJ accession number. Which is the case, and which
	// file or database to use, should be explained in accompanying
	// information.
	SeqName string

	// The source of this feature. This field will normally be used to
	// indicate the program making the prediction, or if it comes from
	// public database annotation, or is experimentally verified, etc.
	Source string

	// The feature type name.
	Feature string

	// FeatStart must be less than FeatEnd and non-negative - GFF indexing
	// is one-base and GFF features cannot have a zero length or a negative
	// position. gff.Feature indexing is, to be consistent with the rest of
	// the library zero-based half open. Translation between zero- and one-
	// based indexing is handled by the gff package.
	FeatStart, FeatEnd int

	// A floating point value representing the score for the feature. A nil
	// value indicates the score is not available.
	FeatScore *float64

	// The strand of the feature - one of seq.Plus, seq.Minus or seq.None.
	// seq.None should be used when strand is not relevant, e.g. for
	// dinucleotide repeats. This field should be set to seq.None for RNA
	// and protein features.
	FeatStrand seq.Strand

	// FeatFrame indicates the frame of the feature. and takes the values
	// Frame0, Frame1, Frame2 or NoFrame. Frame0 indicates that the
	// specified region is in frame. Frame1 indicates that there is one
	// extra base, and Frame2 means that the third base of the region
	// is the first base of a codon. If the FeatStrand is seq.Minus, then
	// the first base of the region is value of FeatEnd, because the
	// corresponding coding region will run from FeatEnd to FeatStart on
	// the reverse strand. As with FeatStrand, if the frame is not relevant
	// then set FeatFrame to NoFram. This field should be set to seq.None
	// for RNA and protein features.
	FeatFrame Frame

	// FeatAttributes represents a collection of GFF2 attributes.
	FeatAttributes Attributes

	// Free comments.
	Comments string
}

func (g *Feature) Start() int { return g.FeatStart }
func (g *Feature) End() int   { return g.FeatEnd }
func (g *Feature) Len() int   { return g.FeatEnd - g.FeatStart }
func (g *Feature) Name() string {
	return fmt.Sprintf("%s/%s:[%d,%d)", g.Feature, g.SeqName, g.FeatStart, g.FeatEnd)
}
func (g *Feature) Description() string    { return fmt.Sprintf("%s/%s", g.Feature, g.Source) }
func (g *Feature) Location() feat.Feature { return Sequence{SeqName: g.SeqName} }

func handlePanic(f feat.Feature, err *error) {
	r := recover()
	if r != nil {
		e, ok := r.(Error)
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

func mustAtofPtr(f []byte) *float64 {
	if len(f) == 1 && f[0] == '.' {
		return nil
	}
	i, err := strconv.ParseFloat(unsafeString(f), 64)
	if err != nil {
		panic(err)
	}
	return &i
}

func mustAtoFr(f []byte) Frame {
	if len(f) == 1 && f[0] == '.' {
		return NoFrame
	}
	b, err := strconv.ParseInt(unsafeString(f), 0, 8)
	if err != nil {
		panic(err)
	}
	return Frame(b)
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

var alphaNum = func() [256]bool {
	var t [256]bool
	for i := 'a'; i <= 'z'; i++ {
		t[i] = true
	}
	for i := 'A'; i <= 'Z'; i++ {
		t[i] = true
	}
	t['_'] = true
	return t
}()

func splitAnnot(f []byte) (tag, value []byte) {
	var (
		i     int
		b     byte
		split bool
	)
	for i, b = range f {
		space := unicode.IsSpace(rune(b))
		if !split {
			if !space && !alphaNum[b] {
				panic(ErrBadTag)
			}
			if space {
				split = true
				tag = f[:i]
			}
		} else if !space {
			break
		}
	}
	if !split {
		return f, nil
	}
	return tag, f[i:]
}

func mustAtoa(f []byte) []Attribute {
	c := bytes.Split(f, []byte{';'})
	a := make([]Attribute, 0, len(c))
	for _, f := range c {
		f = bytes.TrimSpace(f)
		if len(f) == 0 {
			continue
		}
		tag, value := splitAnnot(f)
		if len(tag) == 0 {
			panic(ErrBadTag)
		} else {
			a = append(a, Attribute{Tag: string(tag), Value: string(value)})
		}
	}
	return a
}

type Metadata struct {
	Name          string
	Date          time.Time
	Version       int
	SourceVersion string
	Type          feat.Moltype
}

// A Reader can parse GFFv2 formatted io.Reader and return feat.Features.
type Reader struct {
	r          *bufio.Reader
	TimeFormat string // Required for parsing date fields. Defaults to astronomical format.

	Metadata
}

// NewReader returns a new GFFv2 format reader that reads from r.
func NewReader(r io.Reader) *Reader {
	return &Reader{
		r:          bufio.NewReader(r),
		TimeFormat: Astronomical,
		Metadata:   Metadata{Type: feat.Undefined},
	}
}

func (r *Reader) commentMetaline(line []byte) (f feat.Feature, err error) {
	fields := bytes.Split(line, []byte{' '})
	if len(fields) < 1 {
		return nil, ErrEmptyMetaLine
	}
	switch unsafeString(fields[0]) {
	case "gff-version":
		v := mustAtoi(fields[1])
		if v > Version {
			return nil, ErrNotHandled
		}
		r.Version = Version
		return r.Read()
	case "source-version":
		if len(fields) <= 1 {
			return nil, ErrBadMetaLine
		}
		r.SourceVersion = string(bytes.Join(fields[1:], []byte{' '}))
		return r.Read()
	case "date":
		if len(fields) <= 1 {
			return nil, ErrBadMetaLine
		}
		if len(r.TimeFormat) > 0 {
			r.Date, err = time.Parse(r.TimeFormat, unsafeString(bytes.Join(fields[1:], []byte{' '})))
			if err != nil {
				return nil, err
			}
		}
		return r.Read()
	case "Type", "type":
		if len(fields) <= 1 {
			return nil, ErrBadMetaLine
		}
		r.Type = feat.ParseMoltype(unsafeString(fields[1]))
		if len(fields) > 2 {
			r.Name = string(fields[2])
		}
		return r.Read()
	case "sequence-region":
		if len(fields) <= 3 {
			return nil, ErrBadMetaLine
		}
		return &Region{
			Sequence:    Sequence{SeqName: string(fields[1]), Type: r.Type},
			RegionStart: feat.OneToZero(mustAtoi(fields[2])),
			RegionEnd:   mustAtoi(fields[3]),
		}, nil
	case "DNA", "RNA", "Protein", "dna", "rna", "protein":
		if len(fields) <= 1 {
			return nil, ErrBadMetaLine
		}
		return r.metaSeq(fields[0], fields[1])
	default:
		return nil, ErrNotHandled
	}

	return
}

func (r *Reader) metaSeq(moltype, id []byte) (seq.Sequence, error) {
	var line, body []byte

	var err error
	for {
		line, err = r.r.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if len(line) < 2 || !bytes.HasPrefix(line, []byte("##")) {
			return nil, ErrBadSequence
		}
		line = bytes.TrimSpace(line[2:])
		if unsafeString(line) == "end-"+unsafeString(moltype) {
			break
		} else {
			line = bytes.Join(bytes.Fields(line), nil)
			body = append(body, line...)
		}
	}

	var alpha alphabet.Alphabet
	switch feat.ParseMoltype(unsafeString(moltype)) {
	case feat.DNA:
		alpha = alphabet.DNA
	case feat.RNA:
		alpha = alphabet.RNA
	case feat.Protein:
		alpha = alphabet.Protein
	default:
		return nil, ErrBadMoltype
	}
	s := linear.NewSeq(string(id), alphabet.BytesToLetters(body), alpha)

	return s, err
}

// Read reads a single feature or part and return it or an error. A call to read may
// have side effects on the Reader's Metadata field.
func (r *Reader) Read() (f feat.Feature, err error) {
	defer handlePanic(f, &err)

	var line []byte
	for {
		line, err = r.r.ReadBytes('\n')
		if err != nil {
			return
		}
		line = bytes.TrimSpace(line)
		if len(line) == 0 { // ignore blank lines
			continue
		} else if bytes.HasPrefix(line, []byte("##")) {
			f, err = r.commentMetaline(line[2:])
			return
		} else if line[0] != '#' { // ignore comments
			break
		}
	}

	fields := bytes.SplitN(line, []byte{'\t'}, lastField)
	if len(fields) < frameField {
		return nil, ErrFieldMissing
	}

	gff := &Feature{
		SeqName:    string(fields[nameField]),
		Source:     string(fields[sourceField]),
		Feature:    string(fields[featureField]),
		FeatStart:  feat.OneToZero(mustAtoi(fields[startField])),
		FeatEnd:    mustAtoi(fields[endField]),
		FeatScore:  mustAtofPtr(fields[scoreField]),
		FeatStrand: mustAtos(fields[strandField]),
		FeatFrame:  mustAtoFr(fields[frameField]),
	}

	if len(fields) <= attributeField {
		return gff, nil
	}
	gff.FeatAttributes = mustAtoa(fields[attributeField])
	if len(fields) <= commentField {
		return gff, nil
	}
	gff.Comments = string(fields[commentField])

	return gff, nil
}

// A Writer outputs features and sequences into GFFv2 format.
type Writer struct {
	w          *bufio.Writer
	TimeFormat string
	Precision  int
	Width      int
	header     bool
	closed     bool
}

// Returns a new GFF format writer using w. When header is true,
// a version header will be written to the GFF.
func NewWriter(w io.Writer, width int, header bool) *Writer {
	gw := &Writer{
		w:          bufio.NewWriter(w),
		Width:      width,
		TimeFormat: Astronomical,
		Precision:  -1,
	}

	if header {
		gw.WriteMetaData(Version)
	}

	return gw
}

// Write writes a single feature and return the number of bytes written and any error.
// gff.Features are written as a canonical GFF line, seq.Sequences are written as inline
// sequence in GFF format (note that only sequences of feat.Moltype DNA, RNA and Protein
// are supported). gff.Sequences are not handled as they have a zero length. All other
// feat.Feature are written as sequence region metadata lines.
func (w *Writer) Write(f feat.Feature) (n int, err error) {
	w.header = true
	switch f := f.(type) {
	case *Feature:
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
		n, err = fmt.Fprintf(w.w, "%s\t%s\t%s\t%d\t%d\t",
			f.SeqName,
			f.Source,
			f.Feature,
			feat.ZeroToOne(f.FeatStart),
			f.FeatEnd,
		)
		if err != nil {
			return n, err
		}
		var in int
		if f.FeatScore != nil && !math.IsNaN(*f.FeatScore) {
			in, err = fmt.Fprintf(w.w, "%.*f", w.Precision, *f.FeatScore)
			if err != nil {
				return n, err
			}
			n += in
		} else {
			err = w.w.WriteByte('.')
			if err != nil {
				return n, err
			}
			n++
		}
		in, err = fmt.Fprintf(w.w, "\t%s\t%s",
			f.FeatStrand,
			f.FeatFrame,
		)
		if err != nil {
			return n, err
		}
		n += in
		if f.FeatAttributes != nil {
			in, err = fmt.Fprintf(w.w, "\t%v", f.FeatAttributes)
			n += in
			if err != nil {
				return n, err
			}
		} else if f.Comments != "" {
			err = w.w.WriteByte('\t')
			if err != nil {
				return
			}
			n++
		}
		if f.Comments != "" {
			in, err = fmt.Fprintf(w.w, "\t%s", f.Comments)
		}
		return n + in, err
	case seq.Sequence:
		sw := fasta.NewWriter(w.w, w.Width)
		moltype := f.Alphabet().Moltype()
		if moltype < feat.DNA || moltype > feat.Protein {
			return 0, ErrNotHandled
		}
		sw.IDPrefix = [...][]byte{
			feat.DNA:     []byte("##DNA "),
			feat.RNA:     []byte("##RNA "),
			feat.Protein: []byte("##Protein "),
		}[moltype]
		sw.SeqPrefix = []byte("##")
		n, err = sw.Write(f)
		if err != nil {
			return n, err
		}
		var in int
		in, err = w.w.WriteString([...]string{
			feat.DNA:     "##end-DNA\n",
			feat.RNA:     "##end-RNA\n",
			feat.Protein: "##end-Protein\n",
		}[moltype])
		if err != nil {
			return n, err
		}
		return n + in, err
	case Sequence:
		return 0, ErrNotHandled
	case *Region:
		return fmt.Fprintf(w.w, "##sequence-region %s %d %d\n", f.SeqName, feat.ZeroToOne(f.RegionStart), f.RegionEnd)
	default:
		return fmt.Fprintf(w.w, "##sequence-region %s %d %d\n", f.Name(), feat.ZeroToOne(f.Start()), f.End())
	}

	panic("cannot reach")
}

// WriteMetaData writes a meta data line to a GFF file. The type of metadata line
// depends on the type of d: strings and byte slices are written verbatim, an int is
// interpreted as a version number and can only be written before any other data,
// feat.Moltype and gff.Sequence types are written as sequence type lines, gff.Features
// and gff.Regions are written as sequence regions, sequences are written in GFF
// format and time.Time values are written as date line. All other type return an
// ErrNotHandled.
func (w *Writer) WriteMetaData(d interface{}) (n int, err error) {
	defer func() { w.header = true }()
	switch d := d.(type) {
	case string:
		return fmt.Fprintf(w.w, "##%s\n", d)
	case []byte:
		return fmt.Fprintf(w.w, "##%s\n", d)
	case int:
		if w.header {
			return 0, ErrCannotHeader
		}
		return fmt.Fprintf(w.w, "##gff-version %d\n", d)
	case feat.Moltype:
		return fmt.Fprintf(w.w, "##Type %s\n", d)
	case Sequence:
		return fmt.Fprintf(w.w, "##Type %s %s\n", d.Type, d.SeqName)
	case *Feature:
		return fmt.Fprintf(w.w, "##sequence-region %s %d %d\n", d.SeqName, feat.ZeroToOne(d.FeatStart), d.FeatEnd)
	case feat.Feature:
		return w.Write(d)
	case time.Time:
		return fmt.Fprintf(w.w, "##date %s\n", d.Format(w.TimeFormat))
	}
	return 0, ErrNotHandled
}

// WriteComment writes a comment line to a GFF file.
func (w *Writer) WriteComment(c string) (n int, err error) {
	return fmt.Fprintf(w.w, "# %s\n", c)
}

// Close closes the Writer. The underlying io.Writer is not closed.
func (w *Writer) Close() error {
	if w.closed {
		return nil
	}
	w.closed = true
	return w.w.Flush()
}
