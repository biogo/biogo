package filter
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
type FilterHit struct {
	QFrom     int
	QTo       int
	DiagIndex int
}

// This is a direct translation of the qsort compar function used by PALS.
// However it results in a different sort order (with respect to the non-key
// fields) for FilterHits because of differences in the underlying sort
// algorithms and their respective sort stability.
// This appears to have some impact on FilterHit merging.
func (self FilterHit) Less(y interface{}) bool {
	return self.QFrom < y.(FilterHit).QFrom
}
