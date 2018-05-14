// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package fai implement FAI fasta sequence file index handling.
package fai

import (
	"encoding/csv"
	"errors"
	"io"
	"strconv"
)

const (
	nameField = iota
	lengthField
	startField
	basesField
	bytesField
)

var ErrNonUnique = errors.New("non-unique record name")

// Index is a FAI index.
type Index map[string]Record

// Record is a single FAI index record.
type Record struct {
	// Name is the name of the sequence.
	Name string
	// Length is the length of the sequence.
	Length int
	// Start is the starting seek offset of
	// the sequence.
	Start int64
	// BasesPerLine is the number of sequences
	// bases per line.
	BasesPerLine int
	// BytesPerLine is the number of bytes
	// used to represent each line.
	BytesPerLine int
}

// Position returns the seek offset of the sequence position p for the
// given Record.
func (r Record) Position(p int) int64 {
	if p < 0 || r.Length <= p {
		panic("fai: index out of range")
	}
	return r.Start + int64(p/r.BasesPerLine*r.BytesPerLine+p%r.BasesPerLine)
}

func mustAtoi(fields []string, index, line int) int {
	i, err := strconv.ParseInt(fields[index], 10, 0)
	if err != nil {
		panic(parseError(line, index, err))
	}
	return int(i)
}

func mustAtoi64(fields []string, index, line int) int64 {
	i, err := strconv.ParseInt(fields[index], 10, 64)
	if err != nil {
		panic(parseError(line, index, err))
	}
	return i
}

// ReadFrom returns an Index from the stream provided by an io.Reader or an error. If the input
// contains non-unique records the error is a csv.ParseError identifying the second non-unique
// record.
func ReadFrom(r io.Reader) (idx Index, err error) {
	tr := csv.NewReader(r)
	tr.Comma = '\t'
	tr.FieldsPerRecord = 5
	defer func() {
		r := recover()
		if r != nil {
			e, ok := r.(error)
			if !ok {
				panic(r)
			}
			if _, ok = r.(*csv.ParseError); !ok {
				panic(r)
			}
			err = e
			idx = nil
		}
	}()
	for line := 1; ; line++ {
		rec, err := tr.Read()
		if err == io.EOF {
			return idx, nil
		}
		if err != nil {
			return nil, err
		}
		if idx == nil {
			idx = make(Index)
		} else if _, exists := idx[rec[nameField]]; exists {
			return nil, parseError(line, 0, ErrNonUnique)
		}
		idx[rec[nameField]] = Record{
			Name:         rec[nameField],
			Length:       mustAtoi(rec, lengthField, line),
			Start:        mustAtoi64(rec, startField, line),
			BasesPerLine: mustAtoi(rec, basesField, line),
			BytesPerLine: mustAtoi(rec, bytesField, line),
		}
	}
}
