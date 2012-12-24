// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pals

import (
	"code.google.com/p/biogo/feat"
	"code.google.com/p/biogo/io/featio/gff"
	"fmt"
	"io"
	"os"
)

var t *feat.Feature = &feat.Feature{Source: "pals", Feature: "hit"}

// PALS pair writer type.
type Writer struct {
	w *gff.Writer
}

// Returns a new PALS writer using f.
func NewWriter(f io.WriteCloser, v, width int, header bool) (w *Writer) {
	return &Writer{gff.NewWriter(f, v, width, header)}
}

// Returns a new PALS writer using a filename, truncating any existing file.
// If appending is required use NewWriter and os.OpenFile.
func NewWriterName(name string, v, width int, header bool) (*Writer, error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	return NewWriter(f, v, width, header), nil
}

// Write a single feature and return the number of bytes written and any error.
func (w *Writer) Write(pair *Pair) (n int, err error) {
	t.Location = pair.B.Name()
	t.Start = pair.B.Start()
	t.End = pair.B.End()
	t.Score = floatPtr(float64(pair.Score))
	t.Strand = pair.Strand
	t.Frame = -1
	t.Attributes = fmt.Sprintf("Target %s %d %d; maxe %.2g", pair.A.Name(), pair.A.Start()+1, pair.A.End(), pair.Error) // +1 is kludge for absence of gffwriter
	return w.w.Write(t)
}

func floatPtr(f float64) *float64 { return &f }

// Close the writer, flushing any unwritten data.
func (w *Writer) Close() (err error) {
	return w.w.Close()
}
