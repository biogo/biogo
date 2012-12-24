// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package feat

import ()

type FeatureSet []*Feature

func (self FeatureSet) Add(f *Feature) (g FeatureSet) {
	self = append(self, f)
	return self
}
