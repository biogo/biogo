// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package fai implements FAI fasta sequence file index handling.
//
// This package is deprecated. Please use the API-compatible version at github.com/biogo/hts/fai.
package fai

import (
	"io"

	"github.com/biogo/hts/fai"
)

var ErrNonUnique = fai.ErrNonUnique

// Index is an FAI index.
type Index = fai.Index

// Record is a single FAI index record.
type Record = fai.Record

// ReadFrom returns an Index from the stream provided by an io.Reader or an error. If the input
// contains non-unique records the error is a csv.ParseError identifying the second non-unique
// record.
func ReadFrom(r io.Reader) (idx Index, err error) {
	return fai.ReadFrom(r)
}
