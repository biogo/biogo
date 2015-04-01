// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package filter

// Type to store individual q-gram query filter hits.
type Hit struct {
	From     int
	To       int
	Diagonal int
}

// This is a direct translation of the qsort compar function used by PALS.
// However it results in a different sort order (with respect to the non-key
// fields) for Hits because of differences in the underlying sort algorithms
// and their respective sort stability.
// This appears to have some impact on Hit merging.
func (h Hit) Less(y interface{}) bool {
	return h.From < y.(Hit).From
}
