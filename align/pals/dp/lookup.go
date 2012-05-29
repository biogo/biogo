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

import (
	"github.com/kortschak/biogo/bio"
	"github.com/kortschak/biogo/util"
)

var lookUp util.CTL

func init() {
	m := make(map[int]int)

	for i, v := range bio.N {
		m[int(v)] = i
		m[int(v+32)] = i
	}

	lookUp = *util.NewCTL(m)
}
