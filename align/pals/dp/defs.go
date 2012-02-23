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

import ()

var (
	MaxIGap    int     = 5
	DiffCost   int     = 3
	SameCost   int     = 1
	MatchCost  int     = DiffCost + SameCost
	BlockCost  int     = DiffCost * MaxIGap
	RMatchCost float64 = float64(DiffCost) + 1
)

// Holds dp alignment parameters.
type Params struct {
	MinHitLength int
	MinId        float64
}

// DPHit holds details of alignment result. 
type DPHit struct {
	Abpos, Bbpos              int     // Start coordinate of local alignment
	Aepos, Bepos              int     // End coordinate of local alignment
	LowDiagonal, HighDiagonal int     // Alignment is between (anti)diagonals LowDiagonal & HighDiagonal
	Score                     int     // Score of alignment where match = SameCost, difference = -DiffCost
	Error                     float64 // Lower bound on error rate of match
}
