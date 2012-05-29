// Copyright ©2011-2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

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
