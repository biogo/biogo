// Needleman-Wunsch sequence alignment package
package nw
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

import (
	"github.com/kortschak/BioGo/bio"
	"github.com/kortschak/BioGo/bio/alignment"
	"github.com/kortschak/BioGo/bio/seq"
	"github.com/kortschak/BioGo/bio/util"
)

var LookUp util.CTL

func init() {
	m := make(map[int]int)

	for i, v := range bio.N {
		m[int(v)] = i % 4
	}

	LookUp = *util.NewCTL(m)
}

type Aligner struct {
	Gap     int
	Matrix  [][]int
	GapChar byte
}

func (self *Aligner) Align(reference, query *seq.Seq) (aln *alignment.Alignment) {
	r, c := reference.Len()+1, query.Len()+1
	table := make([][]int, r)
	for i := range table {
		table[i] = make([]int, c)
	}

	for i := 1; i < r; i++ {
		for j := 1; j < c; j++ {
			if rVal, qVal := LookUp.ValueToCode[reference.Seq[i-1]], LookUp.ValueToCode[query.Seq[j-1]]; rVal < 0 || qVal < 0 {
				continue
			} else {
				match := table[i-1][j-1] + self.Matrix[rVal][qVal]
				delete := table[i-1][j] + self.Gap
				insert := table[i][j-1] + self.Gap
				table[i][j] = util.Max(match, delete, insert)
			}
		}
	}

	refAln := &seq.Seq{ID: append([]byte{}, reference.ID...), Seq: make([]byte, 0, reference.Len())}
	queryAln := &seq.Seq{ID: append([]byte{}, query.ID...), Seq: make([]byte, 0, query.Len())}

	var score, scoreDiag, scoreUp, scoreLeft int

	i, j := r-1, c-1
	for i > 0 && j > 0 {
		score = table[i][j]
		scoreDiag = table[i-1][j-1]
		scoreUp = table[i][j-1]
		scoreLeft = table[i-1][j]
		if rVal, qVal := LookUp.ValueToCode[reference.Seq[i-1]], LookUp.ValueToCode[query.Seq[j-1]]; rVal < 0 || qVal < 0 {
			continue
		} else {
			switch {
			case score == scoreDiag+self.Matrix[rVal][qVal]:
				refAln.Seq = append(refAln.Seq, reference.Seq[i-1])
				queryAln.Seq = append(queryAln.Seq, query.Seq[j-1])
				i--
				j--
			case score == scoreLeft+self.Gap:
				refAln.Seq = append(refAln.Seq, reference.Seq[i-1])
				queryAln.Seq = append(queryAln.Seq, self.GapChar)
				i--
			case score == scoreUp+self.Gap:
				refAln.Seq = append(refAln.Seq, self.GapChar)
				queryAln.Seq = append(queryAln.Seq, query.Seq[j-1])
				j--
			default:
				panic("lost path")
			}
		}
	}

	for ; i > 0; i-- {
		refAln.Seq = append(refAln.Seq, reference.Seq[i-1])
		queryAln.Seq = append(queryAln.Seq, self.GapChar)
	}
	for ; j > 0; j-- {
		refAln.Seq = append(refAln.Seq, self.GapChar)
		queryAln.Seq = append(queryAln.Seq, query.Seq[j-1])
	}

	for i, j := 0, len(refAln.Seq)-1; i < j; i, j = i+1, j-1 {
		refAln.Seq[i], refAln.Seq[j] = refAln.Seq[j], refAln.Seq[i]
	}
	for i, j := 0, len(queryAln.Seq)-1; i < j; i, j = i+1, j-1 {
		queryAln.Seq[i], queryAln.Seq[j] = queryAln.Seq[j], queryAln.Seq[i]
	}

	aln = &alignment.Alignment{refAln, queryAln}

	return
}
