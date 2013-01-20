#!/usr/bin/bash

# Copyright ©2012 The bíogo Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

cat < phred.go | \
	gofmt -r 'Phred -> Solexa' | \
	gofmt -r 'NewPhred -> NewSolexa' | \
	gofmt -r 'Qphred -> Qsolexa' | \
	gofmt -r 'Qphreds -> Qsolexas' | \
	gofmt -r 'DecodeToQphred -> DecodeToQsolexa' | \
	gofmt -r 'Ephred -> Esolexa' \
> solexa.go
