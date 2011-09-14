// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
//   This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU General Public License for more details.
//
//   You should have received a copy of the GNU General Public License
//   along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
package dp

var vecBuffering int = 100000

/*** FORWARD AND REVERSE D.P. EXTENSION ROUTINES ***/
/*      Called at the mid-point of trapezoid -- mid X [lo,hi], the extension
        is computed to an end point and the lowest and highest diagonals
        are recorded.  These are returned in a partially filled DPHit
        record, that will be merged with that returned for extension in the
        opposite direction.
*/

func (self *DP) traceForwardPath(mid, lo, hi int) {
	var (
		mxv, mxl, mxr, mxi, mxj int
		i, j                    int
		w                       []int
	)
	debug.Printf("F %d x [%d,%d] %v\n", mid, hi, lo, !self.comp)

	/* Set basis from (mid,lo) .. (mid,hi) */
	if lo < 0 {
		lo = 0
	}

	if hi > self.b.Len() {
		hi = self.b.Len()
	}

	if hi-lo+MaxIGap >= cap(self.vec)/2 {
		vecMax := 3*(hi-lo+MaxIGap+1)/2 + vecBuffering
		self.vec = make([]int, 2*vecMax)
	}

	vec1 := self.vec[:len(self.vec)/2]
	vec2 := self.vec[len(self.vec)/2:]
	vec := vec1
	base := lo
	odd := true

	for j = lo; j <= hi; j++ {
		vec[j-base] = 0
	}

	hi += MaxIGap
	if hi > self.b.Len() {
		hi = self.b.Len()
	}

	for ; j <= hi; j++ {
		vec[j-base] = vec[j-1-base] - DiffCost
	}

	mxv = 0
	mxr = mid - lo
	mxl = mid - hi
	mxi = mid
	mxj = lo

	/* Advance to next row */
	for i = mid; lo <= hi && i < self.a.Len(); i++ {
		w = vec
		if odd {
			vec = vec2
		} else {
			vec = vec1
		}

		odd = !odd

		v := w[lo-base]
		c := v - DiffCost
		vec[lo-base] = c

		for j = lo + 1; j <= hi; j++ {
			t := c
			c = v
			v = w[j-base]
			if self.a.Seq[i] == self.b.Seq[j-1] && lookUp.ValueToCode[self.a.Seq[i]] >= 0 {
				c += MatchCost
			}

			r := c

			if v > r {
				r = v
			}

			if t > r {
				r = t
			}

			c = r - DiffCost
			vec[j-base] = c

			if c >= mxv {
				mxv = c
				mxi = i + 1
				mxj = j
			}
		}

		if j <= self.b.Len() {
			if self.a.Seq[i] == self.b.Seq[j-1] && lookUp.ValueToCode[self.a.Seq[i]] >= 0 {
				v += MatchCost
			}

			r := v

			if c > r {
				r = c
			}

			v = r - DiffCost
			vec[j-base] = v

			if v > mxv {
				mxv = v
				mxi = i + 1
				mxj = j
			}

			for j++; j <= self.b.Len(); j++ {
				v -= DiffCost
				if v < mxv-BlockCost {
					break
				}
				vec[j-base] = v
			}
		}

		hi = j - 1

		for lo <= hi && self.vec[lo-base] < mxv-BlockCost {
			lo += 1
		}

		for lo <= hi && self.vec[hi-base] < mxv-BlockCost {
			hi -= 1
		}

		if hi-lo+2 > cap(self.vec)/2 {
			vecMax := 3*(hi-lo+2)/2 + vecBuffering
			self.vec = make([]int, 2*vecMax)
		}

		if i+1-lo > mxr {
			mxr = i + 1 - lo
		}

		if i+1-hi < mxl {
			mxl = i + 1 - hi
		}
	}

	self.dp[forward].Aepos = mxj
	self.dp[forward].Bepos = mxi
	self.dp[forward].Ldiag = mxl
	self.dp[forward].Hdiag = mxr
	self.dp[forward].Score = mxv
}

func (self *DP) traceReversePath(top, lo, hi, bot, xfactor int) {
	var (
		mxv, mxl, mxr, mxi, mxj int
		i, j                    int
		w                       []int
	)
	debug.Printf("R [%d,%d]x[%d,%d] %d %v\n", top, bot, hi, lo, xfactor, !self.comp)

	/* Set basis from (top,lo) .. (top,hi) */
	if lo < 0 {
		lo = 0
	}

	if hi > self.b.Len() {
		hi = self.b.Len()
	}

	if hi-lo+MaxIGap >= cap(self.vec)/2 {
		vecMax := 3*(hi-lo+MaxIGap+1)/2 + vecBuffering
		self.vec = make([]int, 2*vecMax)
	}
	vec1 := self.vec[:len(self.vec)/2]
	vec2 := self.vec[len(self.vec)/2:]
	vec := vec1
	base := hi - cap(vec)/2 + 1
	odd := true

	for j = hi; j >= lo; j-- {
		vec[j-base] = 0
	}

	lo -= MaxIGap

	if lo < 0 {
		lo = 0
	}

	for ; j >= lo; j-- {
		vec[j-base] = vec[j+1-base] - DiffCost
	}

	mxv = 0
	mxr = top - lo
	mxl = top - hi
	mxi = top
	mxj = lo

	/* Advance to next row */
	if top-1 <= bot {
		xfactor = BlockCost
	}

	for i = top - 1; lo <= hi && i >= 0; i-- {
		w = vec
		if odd {
			vec = vec2
		} else {
			vec = vec1
		}

		odd = !odd

		v := w[hi-base]
		c := v - DiffCost
		vec[hi-base] = c

		for j = hi - 1; j >= lo; j-- {
			t := c
			c = v
			v = w[j-base]

			if self.a.Seq[i] == self.b.Seq[j] && lookUp.ValueToCode[self.a.Seq[i]] >= 0 {
				c += MatchCost
			}
			r := c

			if v > r {
				r = v
			}

			if t > r {
				r = t
			}

			c = r - DiffCost
			vec[j-base] = c

			if c >= mxv {
				mxv = c
				mxi = i
				mxj = j
			}
		}

		if j >= 0 {
			if self.a.Seq[i] == self.b.Seq[j] && lookUp.ValueToCode[self.a.Seq[i]] >= 0 {
				v += MatchCost
			}
			r := v

			if c > r {
				r = c
			}

			v = r - DiffCost
			vec[j-base] = v

			if v > mxv {
				mxv = v
				mxi = i
				mxj = j
			}

			for j--; j >= 0; j-- {
				v -= DiffCost
				if v < mxv-xfactor {
					break
				}
				vec[j-base] = v
			}
		}

		lo = j + 1

		for lo <= hi && vec[lo-base] < mxv-xfactor {
			lo += 1
		}

		for lo <= hi && vec[hi-base] < mxv-xfactor {
			hi -= 1
		}

		if i == bot {
			xfactor = BlockCost
		}

		if hi-lo+2 > cap(self.vec)/2 {
			vecMax := 3*(hi-lo+2)/2 + vecBuffering
			self.vec = make([]int, 2*vecMax)
		}

		if i-lo > mxr {
			mxr = i - lo
		}

		if i-hi < mxl {
			mxl = i - hi
		}
	}

	self.dp[reverse].Abpos = mxj
	self.dp[reverse].Bbpos = mxi
	self.dp[reverse].Ldiag = mxl
	self.dp[reverse].Hdiag = mxr
	self.dp[reverse].Score = mxv
}
