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
