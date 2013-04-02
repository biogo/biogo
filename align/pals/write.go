// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pals

import (
	"code.google.com/p/biogo/feat"
	"code.google.com/p/biogo/io/featio/gff"

	"fmt"
	"io"
)

// Writer is a type that writes PALS pair feature in GFFv2 format.
type Writer struct {
	w *gff.Writer
	t *gff.Feature
}

// NewWriter returns a new PALS writer that write PALS alignment features to the io.Writer w.
func NewWriter(w io.Writer, prec, width int, header bool) *Writer {
	gw := gff.NewWriter(w, width, header)
	gw.Precision = prec
	return &Writer{
		w: gw,
		t: &gff.Feature{Source: "pals", Feature: "hit"},
	}
}

// Write writes a single feature and return the number of bytes written and any error.
func (w *Writer) Write(pair *Pair) (n int, err error) {
	t := w.t
	t.SeqName = pair.B.Location().Name()
	t.FeatStart = pair.B.Start()
	t.FeatEnd = pair.B.End()
	t.FeatScore = floatPtr(float64(pair.Score))
	t.FeatStrand = pair.Strand
	t.FeatFrame = gff.NoFrame
	t.FeatAttributes = append(t.FeatAttributes[:0],
		gff.Attribute{
			Tag:   "Target",
			Value: fmt.Sprintf("%s %d %d", pair.A.Location().Name(), feat.ZeroToOne(pair.A.Start()), pair.A.End()),
		},
		gff.Attribute{
			Tag:   "maxe",
			Value: fmt.Sprintf("%.2g", pair.Error),
		},
	)

	return w.w.Write(t)
}

func floatPtr(f float64) *float64 { return &f }
