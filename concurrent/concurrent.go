// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package to apply a function over an array or stream of data.
package concurrent

// The Concurrent interface represents a processor that allows adding jobs and retrieving results
type Concurrent interface {
	Process(...interface{})
	Result() (interface{}, error)
}
