package matrix

import "github.com/biogo/biogo/alphabet"

// Match generates a penalty matrix for a.
// Perfect matches have penalty match.
// Gaps have penalty gap.
// Everything else has penalty mismatch.
// For example, Match(alphabet.DNA, 0, 1, -1) generates the original Needleman-Wunsch penalty matrix.
func Match(a alphabet.Alphabet, gap, match, mismatch int) [][]int {
	l := a.Len()
	arr := make([]int, l*l)
	g := a.IndexOf(a.Gap())
	for i := 0; i < l; i++ {
		for j := 0; j < l; j++ {
			score := mismatch
			switch {
			case i == g, j == g:
				score = gap
			case i == j:
				score = match
			}
			arr[i*l+j] = score
		}
	}
	x := make([][]int, l)
	for i := 0; i < l; i++ {
		x[i] = arr[i*l : (i+1)*l]
	}
	return x
}
