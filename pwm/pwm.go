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

// Package pwm implements a position weight matrix search based on an
// algorithm by Deborah Toledo Flores.
package pwm

import (
	"code.google.com/p/biogo/bio"
	"code.google.com/p/biogo/exp/feat"
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/exp/seq/sequtils"
	ofeat "code.google.com/p/biogo/feat"
	"fmt"
	"math"
	"sort"
)

type probs struct {
	score      float64
	freqs      []int
	occurrence int
}

type probTable []*probs

func (m probTable) Len() int           { return len(m) }
func (m probTable) Less(i, j int) bool { return m[i].score > m[j].score }
func (m probTable) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

type PWM struct {
	matrix    [][]float64
	lookAhead []float64
	table     probTable
	minScore  float64
	Format    string // Format for probability values in attributes.
}

func New(matrix [][]float64) (m *PWM) {
	m = &PWM{
		matrix:    matrix,
		lookAhead: make([]float64, len(matrix)),
		minScore:  math.MaxFloat64,
		Format:    "%e",
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

func (m *PWM) genTable(minScore, score float64, pos int, motif []byte) {
	for i, s := range m.matrix[pos] {
		motif[pos] = byte(i)
		if pos < len(m.matrix)-1 {
			if minScore-(score+s) > m.lookAhead[pos+1] { // will not be able to achieve minScore
				continue
			}
			m.genTable(minScore, score+s, pos+1, motif)
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
			for j := len(m.table) - 1; j >= 0; j-- {
				table := m.table[j]
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
				m.table = append(m.table, &probs{score: score + s, freqs: freqs, occurrence: 1})
			}
		}
	}

}

type Sequence interface {
	seq.Sequence
	Moltype() bio.Moltype
	Orientation() feat.Orientation
}

func (m *PWM) Search(s Sequence, start, end int, minScore float64) []*ofeat.Feature {
	if minScore < m.minScore {
		m.table = make(probTable, 0)
		m.genTable(minScore, 0, 0, make([]byte, len(m.matrix)))
		sort.Sort(m.table)
	}

	var (
		index  = s.Alphabet().LetterIndex()
		length = len(m.matrix)

		freqs = make([]float64, 4)
		zeros = make([]float64, 4)

		diff = 1 / float64(length)

		format = "%-s " + m.Format
		mot    = s.New()
		f      []*ofeat.Feature
	)

LOOP:
	for pos := start; pos+length < end; pos++ {
		// Determine the score for this position.
		var score = 0.
		for i := 0; i < length; i++ {
			base := index[s.At(pos+i).L]
			if base < 0 || minScore-score > m.lookAhead[i] { // not valid base or will not be able to achieve minScore
				continue LOOP
			} else {
				score += m.matrix[i][base]
			}
		}

		if score < minScore {
			continue
		}

		// Calculate base frequencies for window.
		copy(freqs, zeros)
		for i := pos; i < pos+length; i++ {
			base := index[s.At(i).L]
			if base >= 0 {
				freqs[base] += diff
			} else { // Probability for this pos will be meaningless; if N is tolerated, include N in valid alphabet - make special case?
				continue LOOP
			}
		}

		// Descend probability function summing probabilities.
		var (
			prob = 0.
			sp   = 0.
		)
		for _, e := range m.table {
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

		sequtils.Truncate(mot, s, pos, pos+length) // FIXME Should use Slice for non-allocating path.
		f = append(f, &ofeat.Feature{
			Location:   fmt.Sprintf("%s %s", s.Name(), s.Description()),
			Start:      pos,
			End:        pos + length,
			Score:      &score,
			Attributes: fmt.Sprintf(format, mot, prob),
			Strand:     int8(s.Orientation()),
			Moltype:    s.Moltype(),
			Frame:      -1,
		})
	}

	return f
}
