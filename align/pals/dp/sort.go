package dp

// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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

// Sort DPHits on start position.
type starts DPHits

func (self starts) Len() int {
	return len(self)
}

func (self starts) Less(i, j int) bool {
	return self[i].Abpos < self[j].Abpos
}

func (self starts) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

// Sort DPHits on end position.
type ends DPHits

func (self ends) Len() int {
	return len(self)
}

func (self ends) Less(i, j int) bool {
	return self[i].Aepos < self[j].Aepos
}

func (self ends) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}
