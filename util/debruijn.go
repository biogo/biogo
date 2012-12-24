// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

// Return a deBruijn sequence for a n-words with k letters.
func DeBruijn(k, n byte) (s []byte) {
	switch k {
	case 0:
		return []byte{}
	case 1:
		return make([]byte, n)
	}

	a := make([]byte, k*n)
	s = make([]byte, 0, Pow(int(k), n))

	var db func(byte, byte)
	db = func(t, p byte) {
		if t > n {
			if n%p == 0 {
				for j := byte(1); j <= p; j++ {
					s = append(s, a[j])
				}
			}
		} else {
			a[t] = a[t-p]
			db(t+1, p)
			for j := a[t-p] + 1; j < k; j++ {
				a[t] = j
				db(t+1, t)
			}

		}
	}
	db(1, 1)

	return
}
