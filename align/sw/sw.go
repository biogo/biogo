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

// Smith-Waterman sequence alignment package
package sw

import (
	"github.com/kortschak/biogo/bio"
	"github.com/kortschak/biogo/seq"
	"github.com/kortschak/biogo/util"
)

// Default character table lookups.
var LookUpN, LookUpR, LookUpP util.CTL

func init() {
	m := make(map[int]int)

	for i, v := range bio.N {
		m[int(v)] = i % (len(bio.N) / 2)
	}

	LookUpN = *util.NewCTL(m)

	m = make(map[int]int)

	for i, v := range bio.R {
		m[int(v)] = i % (len(bio.R) / 2)
	}

	LookUpR = *util.NewCTL(m)

	m = make(map[int]int)

	for i, v := range bio.P {
		m[int(v)] = i % (len(bio.P) / 2)
	}

	LookUpP = *util.NewCTL(m)
}

const (
	diag = iota
	up
	left
)

func maxIndex(a []int) (d int) {
	max := util.MinInt
	for i, v := range a {
		if v > max {
			max = v
			d = i
		}
	}
	return
}

// Smith-Waterman aligner type.
// Matrix is a square scoring matrix with the last column and last row specifying gap penalties.
// GapChar is the character used to fill gaps. LookUp is used to translate sequance values into
// positions in the scoring matrix.
// Currently gap opening is not considered.
type Aligner struct {
	Matrix  [][]int
	GapChar byte
	LookUp  util.CTL
}

// Method to align two sequences using the Smith-Waterman algorithm. Returns an alignment or an error
// if the scoring matrix is not square.
func (self *Aligner) Align(reference, query *seq.Seq) (aln seq.Alignment, err error) {
	gap := len(self.Matrix) - 1
	for _, row := range self.Matrix {
		if len(row) != gap+1 {
			return nil, bio.NewError("Scoring matrix is not square.", 0, self.Matrix)
		}
	}
	r, c := reference.Len()+1, query.Len()+1
	table := make([][]int, r)
	for i := range table {
		table[i] = make([]int, c)
	}

	max, maxI, maxJ := 0, 0, 0
	var (
		score  int
		scores [3]int
	)

	for i := 1; i < r; i++ {
		for j := 1; j < c; j++ {
			if rVal, qVal := self.LookUp.ValueToCode[reference.Seq[i-1]], self.LookUp.ValueToCode[query.Seq[j-1]]; rVal < 0 || qVal < 0 {
				continue
			} else {
				scores[diag] = table[i-1][j-1] + self.Matrix[rVal][qVal]
				scores[up] = table[i-1][j] + self.Matrix[rVal][gap]
				scores[left] = table[i][j-1] + self.Matrix[gap][qVal]
				score = util.Max(scores[:]...)
				if score < 0 {
					score = 0
				}
				if score >= max { // greedy so make farthest down and right
					max, maxI, maxJ = score, i, j
				}
				table[i][j] = score
			}
		}
	}

	refAln := &seq.Seq{ID: reference.ID, Seq: make([]byte, 0, reference.Len())}
	queryAln := &seq.Seq{ID: query.ID, Seq: make([]byte, 0, query.Len())}

	for i, j := maxI, maxJ; table[i][j] != 0 && i > 0 && j > 0; {
		if rVal, qVal := self.LookUp.ValueToCode[reference.Seq[i-1]], self.LookUp.ValueToCode[query.Seq[j-1]]; rVal < 0 || qVal < 0 {
			continue
		} else {
			scores[diag] = table[i-1][j-1] + self.Matrix[rVal][qVal]
			scores[up] = table[i-1][j] + self.Matrix[gap][qVal]
			scores[left] = table[i][j-1] + self.Matrix[rVal][gap]
			switch d := maxIndex(scores[:]); d {
			case diag:
				i--
				j--
				refAln.Seq = append(refAln.Seq, reference.Seq[i])
				queryAln.Seq = append(queryAln.Seq, query.Seq[j])
			case up:
				i--
				refAln.Seq = append(refAln.Seq, reference.Seq[i])
				queryAln.Seq = append(queryAln.Seq, self.GapChar)
			case left:
				j--
				refAln.Seq = append(refAln.Seq, self.GapChar)
				queryAln.Seq = append(queryAln.Seq, query.Seq[j])
			}
		}
	}

	for i, j := 0, len(refAln.Seq)-1; i < j; i, j = i+1, j-1 {
		refAln.Seq[i], refAln.Seq[j] = refAln.Seq[j], refAln.Seq[i]
	}
	for i, j := 0, len(queryAln.Seq)-1; i < j; i, j = i+1, j-1 {
		queryAln.Seq[i], queryAln.Seq[j] = queryAln.Seq[j], queryAln.Seq[i]
	}

	aln = seq.Alignment{refAln, queryAln}

	return
}
