// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dp

// Sort DPHits on start position.
type starts DPHits

func (s starts) Len() int {
	return len(s)
}

func (s starts) Less(i, j int) bool {
	return s[i].Abpos < s[j].Abpos
}

func (s starts) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Sort DPHits on end position.
type ends DPHits

func (e ends) Len() int {
	return len(e)
}

func (e ends) Less(i, j int) bool {
	return e[i].Aepos < e[j].Aepos
}

func (e ends) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
