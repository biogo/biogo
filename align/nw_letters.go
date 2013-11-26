// This file is automatically generated. Do not edit - make changes to relevant got file.

// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package align

import (
	"code.google.com/p/biogo/alphabet"
	"code.google.com/p/biogo/feat"

	"fmt"
	"os"
	"text/tabwriter"
)

func drawNWTableLetters(rSeq, qSeq alphabet.Letters, table [][]int) {
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 0, ' ', tabwriter.AlignRight|tabwriter.Debug)
	fmt.Printf("rSeq: %s\n", rSeq)
	fmt.Printf("qSeq: %s\n", qSeq)
	fmt.Fprint(tw, "\tqSeq\t")
	for _, l := range qSeq {
		fmt.Fprintf(tw, "%c\t", l)
	}
	fmt.Fprintln(tw)

	for i, row := range table {
		if i == 0 {
			fmt.Fprint(tw, "rSeq\t")
		} else {
			fmt.Fprintf(tw, "%c\t", rSeq[i-1])
		}

		for _, e := range row {
			fmt.Fprintf(tw, "%2v\t", e)
		}
		fmt.Fprintln(tw)
	}
	tw.Flush()
}

func (a NW) alignLetters(rSeq, qSeq alphabet.Letters, alpha alphabet.Alphabet) ([]feat.Pair, error) {
	gap := len(a) - 1
	for _, row := range a {
		if len(row) != gap+1 {
			return nil, ErrMatrixNotSquare
		}
	}

	index := alpha.LetterIndex()
	r, c := rSeq.Len()+1, qSeq.Len()+1
	table := make([][]int, r)
	for i := range table {
		row := make([]int, c)
		if i == 0 {
			for j := range row[1:] {
				row[j+1] = row[j] + a[gap][index[qSeq[j]]]
			}
		} else {
			row[0] = table[i-1][0] + a[index[rSeq[i-1]]][gap]
		}
		table[i] = row
	}

	var scores [3]int
	for i := 1; i < r; i++ {
		for j := 1; j < c; j++ {
			var (
				rVal = index[rSeq[i-1]]
				qVal = index[qSeq[j-1]]
			)
			if rVal < 0 || qVal < 0 {
				continue
			} else {
				scores[diag] = table[i-1][j-1] + a[rVal][qVal]
				scores[up] = table[i-1][j] + a[rVal][gap]
				scores[left] = table[i][j-1] + a[gap][qVal]
				table[i][j] = max(&scores)
				if debugNeedle {
					drawNWTableLetters(rSeq, qSeq, table)
				}
			}
		}
	}

	var aln []feat.Pair
	score, last := 0, diag
	i, j := r-1, c-1
	maxI, maxJ := i, j
	for i > 0 && j > 0 {
		var (
			rVal = index[rSeq[i-1]]
			qVal = index[qSeq[j-1]]
		)
		if rVal < 0 || qVal < 0 {
			continue
		} else {
			switch table[i][j] {
			case table[i-1][j-1] + a[rVal][qVal]:
				if last != diag {
					aln = append(aln, &featPair{
						a:     feature{start: i, end: maxI},
						b:     feature{start: j, end: maxJ},
						score: score,
					})
					maxI, maxJ = i, j
					score = 0
				}
				score += table[i][j] - table[i-1][j-1]
				i--
				j--
				if i == 0 || j == 0 {
					aln = append(aln, &featPair{
						a:     feature{start: i, end: maxI},
						b:     feature{start: j, end: maxJ},
						score: score,
					})
					score = 0
				}
				last = diag
			case table[i-1][j] + a[rVal][gap]:
				if last != up {
					aln = append(aln, &featPair{
						a:     feature{start: i, end: maxI},
						b:     feature{start: j, end: maxJ},
						score: score,
					})
					maxI, maxJ = i, j
					score = 0
				}
				score += table[i][j] - table[i-1][j]
				i--
				last = up
			case table[i][j-1] + a[gap][qVal]:
				if last != left {
					aln = append(aln, &featPair{
						a:     feature{start: i, end: maxI},
						b:     feature{start: j, end: maxJ},
						score: score,
					})
					maxI, maxJ = i, j
					score = 0
				}
				score += table[i][j] - table[i][j-1]
				j--
				last = left
			}
		}
	}

	if i != j {
		aln = append(aln, &featPair{
			a:     feature{start: 0, end: i},
			b:     feature{start: 0, end: j},
			score: table[i][j],
		})
	}

	for i, j := 0, len(aln)-1; i < j; i, j = i+1, j-1 {
		aln[i], aln[j] = aln[j], aln[i]
	}

	return aln, nil
}
