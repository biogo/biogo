// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Packages for reading and writing features
package featio

import "code.google.com/p/biogo/feat"

type Reader interface {
	Read() (*feat.Feature, error)
}

type Writer interface {
	Write(*feat.Feature) error
}
