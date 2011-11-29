package filter
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

// Ukonnen's Lemma
func MinWordsPerFilterHit(hitLength, wordLength, maxErrors int) int {
	return hitLength + 1 - wordLength*(maxErrors+1)
}

var (
	MaxIGap    int     = 5
	DiffCost   int     = 3
	SameCost   int     = 1
	MatchCost  int     = DiffCost + SameCost
	BlockCost  int     = DiffCost * MaxIGap
	RMatchCost float64 = float64(DiffCost) + 1
)

type Params struct {
	WordSize   int
	MinMatch   int
	MaxError   int
	TubeOffset int
}

type TubeState struct {
	QLo   int
	QHi   int
	Count int
}
