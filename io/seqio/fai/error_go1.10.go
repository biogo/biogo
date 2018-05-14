// Copyright ©2018 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build go1.10

package fai

import "encoding/csv"

func parseError(line, column int, err error) *csv.ParseError {
	return &csv.ParseError{
		StartLine: line,
		Line:      line,
		Column:    column,
		Err:       err,
	}
}
