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

// Position weight matrix search
//
// based on algorithm by Deborah Toledo Flores
package pwm

import (
	"github.com/kortschak/biogo/bio"
	"github.com/kortschak/biogo/feat"
	"github.com/kortschak/biogo/seq"
	"github.com/kortschak/biogo/util"
	"math"
	"sort"
	"strconv"
)

var (
	valid  [256]bool
	LookUp util.CTL
)

func init() {
	m := make(map[int]int)

	for i, v := range bio.N {
		m[int(v)] = i % 4
	}

	LookUp = *util.NewCTL(m)

	Valid(bio.N)
}

func Valid(alphabet []byte) {
	for i := range valid {
		valid[i] = false
	}
	for _, v := range alphabet {
		valid[v] = true
	}
}

type probs struct {
	score      float64
	freqs      []int
	occurrence int
}

type probTable []*probs

func (self probTable) Len() int { return len(self) }

func (self probTable) Less(i, j int) bool { return self[i].score > self[j].score }

func (self *probTable) Swap(i, j int) { (*self)[i], (*self)[j] = (*self)[j], (*self)[i] }

type PWM struct {
	matrix      [][]float64
	lookAhead   []float64
	table       probTable
	minScore    float64
	FloatFormat byte
	Precision   int
}

func New(matrix [][]float64) (m *PWM) { // try this and also matrix []map[byte]float64 for speed comparison
	m = &PWM{
		matrix:      matrix,
		lookAhead:   make([]float64, len(matrix)),
		minScore:    math.MaxFloat64,
		FloatFormat: bio.FloatFormat,
		Precision:   bio.Precision,
	}

	var maxVal, maxScore float64

	for i := len(matrix) - 1; i >= 0; i-- {
		maxVal = 0
		for _, v := range matrix[i] {
			if v > maxVal {
				maxVal = v
			}
		}
		maxScore += maxVal
		m.lookAhead[i] = maxScore
	}

	for i := range matrix {
		for j := range matrix[i] {
			matrix[i][j] /= maxScore
		}
		m.lookAhead[i] /= maxScore
	}

	return
}

func (self *PWM) genTable(minScore, score float64, position int, motif []byte) {
	for i, s := range self.matrix[position] {
		motif[position] = byte(i)
		if position < len(self.matrix)-1 {
			if minScore-(score+s) > self.lookAhead[position+1] { // will not be able to achieve minScore
				continue
			}
			self.genTable(minScore, score+s, position+1, motif)
		} else {
			if score+s < minScore { // will not be able to achieve minScore
				continue
			}
			// count frequencies of states in current motif
			freqs := make([]int, 4)
			for _, j := range motif {
				freqs[j]++
			}
			found := false
			for j := len(self.table) - 1; j >= 0; j-- {
				table := self.table[j]
				if table.score != score+s {
					continue // if using insertion sort, if table.score > score+s then we can found = false and break
				}
				match := true
				for k := range freqs {
					if freqs[k] != table.freqs[k] {
						match = false
						break
					}
				}
				if match {
					table.occurrence++
					found = true
					break
				}
			}

			if !found {
				self.table = append(self.table, &probs{score: score + s, freqs: freqs, occurrence: 1}) // use a insertion sort (based on score) ? will make search quicker
			}
		}
	}

}

func (self *PWM) Search(sequence *seq.Seq, start, end int, minScore float64) (scores []*feat.Feature) { // return this as an array of features
	if minScore < self.minScore {
		self.table = make(probTable, 0)
		self.genTable(minScore, 0, 0, make([]byte, len(self.matrix)))
		sort.Sort(&self.table)
		/*
			for _, e := range self.table {
				fmt.Printf("%f\t", e.score)
				for c, f := range e.freqs {
					if f > 0 {
						fmt.Printf("'%c':%d ", [4]byte{'A','C','G','T'}[c], f)
					}
				}
				fmt.Printf("%d\n", e.occurrence)
			}
		*/
	}

	length := len(self.matrix)

	freqs := make([]float64, 4)
	zeros := make([]float64, 4)

	diff := 1 / float64(length)
LOOP:
	for position := start; position+length < end; position++ {
		// determine the score for this position
		score := float64(0)
		for i := 0; i < length; i++ {
			if base := LookUp.ValueToCode[sequence.Seq[position+i]]; base < 0 || minScore-score > self.lookAhead[i] { // not valid base or will not be able to achieve minScore
				continue LOOP
			} else {
				score += self.matrix[i][base]
			}
		}

		if score < minScore {
			continue
		}

		// calculate base frequencies for window
		copy(freqs, zeros)
		for i := position; i < position+length; i++ {
			if base := LookUp.ValueToCode[sequence.Seq[i]]; base >= 0 {
				freqs[base] += diff
			} else { // probability for this position will be meaningless - if N is tolerated, include N in valid alphabet - make special case?
				continue LOOP
			}
		}

		// descend probability function summing probabilities
		prob := float64(0)
		sp := float64(0)
		for _, e := range self.table {
			sp = 1
			if e.score < score {
				break
			}
			for i, f := range freqs {
				sp *= math.Pow(f, float64(e.freqs[i]))
			}
			sp *= float64(e.occurrence)
			prob += sp
		}

		scores = append(scores, &feat.Feature{
			Location:   sequence.ID,
			Start:      position + 1,
			End:        position + length,
			Score:      &score,
			Attributes: string(sequence.Seq[position:position+length]) + " " + strconv.FormatFloat(prob, self.FloatFormat, self.Precision, 64),
			Strand:     sequence.Strand,
			Moltype:    sequence.Moltype,
			Frame:      -1,
		})
	}

	return
}
