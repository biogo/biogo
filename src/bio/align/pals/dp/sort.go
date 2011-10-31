package dp
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
import "bio/align/pals/filter"

//
type traps struct{ traps []*filter.Trapezoid }

func (self traps) Len() int {
	return len(self.traps)
}

func (self traps) Less(i, j int) bool {
	return (*self.traps[i]).Bottom < (*self.traps[j]).Bottom
}

func (self traps) Swap(i, j int) {
	self.traps[i], self.traps[j] = self.traps[j], self.traps[i]
}

//
type starts struct{ hits []DPHit }

func (self starts) Len() int {
	return len(self.hits)
}

func (self starts) Less(i, j int) bool {
	return self.hits[i].Abpos < self.hits[j].Abpos
}

func (self starts) Swap(i, j int) {
	self.hits[i], self.hits[j] = self.hits[j], self.hits[i]
}

//
type ends struct{ hits []DPHit }

func (self ends) Len() int {
	return len(self.hits)
}

func (self ends) Less(i, j int) bool {
	return self.hits[i].Aepos < self.hits[j].Aepos
}

func (self ends) Swap(i, j int) {
	self.hits[i], self.hits[j] = self.hits[j], self.hits[i]
}
