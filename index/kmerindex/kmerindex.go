// Copyright ©2011-2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package kmerindex performs Kmer indexing package based on Bob Edgar and
// Gene Meyers' approach used in PALS.
package kmerindex

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/biogo/biogo/errors"

	"github.com/biogo/biogo/alphabet"
	"github.com/biogo/biogo/seq/linear"
	"github.com/biogo/biogo/util"
)

var (
	ErrKTooLarge      = errors.KTooLargeErr{}.Make("kmerindex: k too large")
	ErrKTooSmall      = errors.KTooSmallErr{}.Make("kmerindex: k too small")
	ErrShortSeq       = errors.KSeqTooShortErr{}.Make("kmerindex: sequence to short for k")
	ErrBadAlphabet    = errors.InvalidAlphabetErr{}.Make("kmerindex: alphabet size != 4")
	ErrBadKmer        = errors.BadKmerErr{}.Make("kmerindex: kmer out of range")
	ErrBadKmerTextLen = errors.BadKmerTextLenErr{}.Make("kmerindex: kmertext length != k")
	ErrBadKmerText    = errors.IllegalKmerTextErr{}.Make("kmerindex: kmertext contains illegal character")
)

// Constraints on Kmer length.
var (
	MinKmerLen = 4 // default minimum

	// MaxKmerLen is the maximum Kmer length that can be used.
	// It is 16 on 64 bit architectures and 14 on 32 bit architectures.
	MaxKmerLen = 16 - offset
)

const offset = int(unsafe.Sizeof(int(0))%0x4) / 2

var Debug = false // Set Debug to true to prevent recovering from panics in ForEachKmer f Eval function.

// 2-bit per base packed word
type Kmer uint32 // Sensible size for word type uint64 will double the size of the index (already large for high k)

// Kmer index type
type Index struct {
	finger  []Kmer
	pos     []int
	seq     *linear.Seq
	lookUp  alphabet.Index
	k       int
	kMask   Kmer
	indexed bool
}

// Create a new Kmer Index with a word size k based on sequence
func New(k int, s *linear.Seq) (*Index, error) {
	switch {
	case k > MaxKmerLen:
		return nil, ErrKTooLarge
	case k < MinKmerLen:
		return nil, ErrKTooSmall
	case k+1 > s.Len():
		return nil, ErrShortSeq
	case s.Alpha.Len() != 4:
		return nil, ErrBadAlphabet
	}

	ki := &Index{
		finger:  make([]Kmer, util.Pow4(k)+1), // Need a Tn+1 finger position so that Tn can be recognised
		k:       k,
		kMask:   Kmer(util.Pow4(k) - 1),
		seq:     s,
		lookUp:  s.Alpha.LetterIndex(),
		indexed: false,
	}
	ki.buildKmerTable()

	return ki, nil
}

// Build the table of Kmer frequencies - called by New
func (ki *Index) buildKmerTable() {
	incrementFinger := func(index *Index, _, kmer int) {
		index.finger[kmer]++
	}
	ki.ForEachKmerOf(ki.seq, 0, ki.seq.Len(), incrementFinger)
}

// Build the Kmer position table destructively replacing Kmer frequencies
func (ki *Index) Build() {
	var sum Kmer
	for i, v := range ki.finger {
		ki.finger[i], sum = sum, sum+v
	}

	locatePositions := func(index *Index, position, kmer int) {
		index.pos[index.finger[kmer]] = position
		index.finger[kmer]++
	}
	ki.pos = make([]int, ki.seq.Len()-ki.k+1)
	ki.ForEachKmerOf(ki.seq, 0, ki.seq.Len(), locatePositions)

	ki.indexed = true
}

// Return an array of positions for the Kmer string kmertext
func (ki *Index) KmerPositionsString(kmertext string) (positions []int, err error) {
	switch {
	case len(kmertext) != ki.k:
		return nil, ErrBadKmerTextLen
	case !ki.indexed:
		return nil, errors.StateErr{}.Make("kmerindex: index not built: call Build()")
	}

	var kmer Kmer
	if kmer, err = ki.KmerOf(kmertext); err != nil {
		return nil, err
	}

	return ki.KmerPositions(kmer)
}

// Return an array of positions for the Kmer kmer
func (ki *Index) KmerPositions(kmer Kmer) (positions []int, err error) {
	if kmer > ki.kMask {
		return nil, ErrBadKmer
	}

	i := Kmer(0)
	if kmer > 0 { // special case: An has no predecessor
		i = ki.finger[kmer-1]
	}
	j := ki.finger[kmer]
	if i == j {
		return
	}

	positions = make([]int, j-i)
	for l, p := range ki.pos[i:j] {
		positions[l] = int(p)
	}

	return
}

// Return a map containing absolute Kmer frequencies and true if called before Build().
// If called after Build returns a nil map and false.
func (ki *Index) KmerFrequencies() (map[Kmer]int, bool) {
	if ki.indexed {
		return nil, false
	}

	m := map[Kmer]int{}

	for i, f := range ki.finger {
		if f > 0 {
			m[Kmer(i)] = int(f) // not always safe - perhaps check that Kmer <= MaxInt
		}
	}

	return m, true
}

// Return a map containing relative Kmer frequencies and true if called before Build().
// If called after Build returns a nil map and false.
func (ki *Index) NormalisedKmerFrequencies() (map[Kmer]float64, bool) {
	if ki.indexed {
		return nil, false
	}

	m := map[Kmer]float64{}

	l := float64(ki.seq.Len())
	for i, f := range ki.finger {
		if f > 0 {
			m[Kmer(i)] = float64(f) / l
		}
	}

	return m, true
}

// Returns a Kmer-keyed map containing slices of kmer positions and true if called after Build,
// otherwise nil and false.
func (ki *Index) KmerIndex() (map[Kmer][]int, bool) {
	if !ki.indexed {
		return nil, false
	}

	m := make(map[Kmer][]int)

	for i := range ki.finger {
		if p, _ := ki.KmerPositions(Kmer(i)); len(p) > 0 {
			m[Kmer(i)] = p
		}
	}

	return m, true
}

// Returns a string-keyed map containing slices of kmer positions and true if called after Build,
// otherwise nil and false.
func (ki *Index) StringKmerIndex() (map[string][]int, bool) {
	if !ki.indexed {
		return nil, false
	}

	m := make(map[string][]int)

	for i := range ki.finger {
		if p, _ := ki.KmerPositions(Kmer(i)); len(p) > 0 {
			m[ki.Format(Kmer(i))] = p
		}
	}

	return m, true
}

// errors should be handled through a panic which will be recovered by ForEachKmerOf
type Eval func(index *Index, j, kmer int)

// Applies the f Eval func to all kmers in s from start to end. Returns any panic raised by f as an error.
func (ki *Index) ForEachKmerOf(s *linear.Seq, start, end int, f Eval) (err error) {
	if !Debug {
		defer func() {
			if r := recover(); r != nil {
				var ok bool
				err, ok = r.(error)
				if !ok {
					err = fmt.Errorf("kmerindex: %v", r)
				}
			}
		}()
	}

	kmer := Kmer(0)
	high := 0
	var currentBase int

	// Preload the first k-1 bases of the first well defined k-mer or set high to the next position
	basePosition := start
	for ; basePosition < start+ki.k-1; basePosition++ {
		currentBase = ki.lookUp[s.Seq[basePosition]]
		if currentBase >= 0 {
			kmer = (kmer << 2) | Kmer(currentBase)
		} else {
			kmer = 0
			high = basePosition + 1
		}
	}

	// Call f(position, kmer) for each of the next well defined k-mers
	for position := basePosition - ki.k + 1; basePosition < end; position++ {
		currentBase = ki.lookUp[s.Seq[basePosition]]
		basePosition++
		if currentBase >= 0 {
			kmer = ((kmer << 2) | Kmer(currentBase)) & ki.kMask
		} else {
			kmer = 0
			high = basePosition
		}
		if position >= high {
			f(ki, position, int(kmer))
		}
	}

	return
}

// Return the Kmer length of the Index.
func (ki *Index) K() int {
	return ki.k
}

// Returns a pointer to the indexed seq.Seq.
func (ki *Index) Seq() *linear.Seq {
	return ki.seq
}

// Returns the value of the finger slice at p. This signifies the absolute kmer frequency of the Kmer(p)
// if called before Build() and points to the relevant position lookup if called after.
func (ki *Index) FingerAt(p int) int {
	return int(ki.finger[p])
}

// Returns the value of the pos slice at p. This signifies the position of the pth kmer if called after Build().
// Not valid before Build() - will panic.
func (ki *Index) PosAt(p int) int {
	return ki.pos[p]
}

// Convert a Kmer into a string of bases
func (ki *Index) Format(kmer Kmer) string {
	s, _ := Format(kmer, ki.k, ki.seq.Alpha)
	return s
}

// Convert a string of bases into a len k Kmer, returns an error if string length does not match k.
// lookUp is an index lookup table as returned by alphabet.Alphabet.LetterIndex().
func KmerOf(k int, lookUp alphabet.Index, kmertext string) (kmer Kmer, err error) {
	if len(kmertext) != k {
		return 0, ErrBadKmerTextLen
	}

	for _, v := range kmertext {
		x := lookUp[v]
		if x < 0 {
			return 0, ErrBadKmerText
		}
		kmer = (kmer << 2) | Kmer(x)
	}

	return
}

// Return the GC fraction of a Kmer
func (ki *Index) GCof(kmer Kmer) float64 {
	return GCof(ki.k, kmer)
}

// Return the GC fraction of a Kmer of len k
func GCof(k int, kmer Kmer) float64 {
	gc := 0
	for i := k - 1; i >= 0; i, kmer = i-1, kmer>>2 {
		gc += int((kmer & 1) ^ ((kmer & 2) >> 1))
	}

	return float64(gc) / float64(k)
}

// Convert a Kmer into a string of bases
func Format(kmer Kmer, k int, alpha alphabet.Alphabet) (string, error) {
	if alpha.Len() != 4 {
		return "", ErrBadAlphabet
	}
	kmertext := make([]byte, k)

	for i := k - 1; i >= 0; i, kmer = i-1, kmer>>2 {
		kmertext[i] = byte(alpha.Letter(int(kmer & 3)))
	}

	return string(kmertext), nil
}

// Reverse complement a Kmer. Complementation is performed according to letter index:
//
//  0, 1, 2, 3 = 3, 2, 1, 0
func (ki *Index) ComplementOf(kmer Kmer) (c Kmer) {
	return ComplementOf(ki.k, kmer)
}

// Reverse complement a Kmer of len k. Complementation is performed according to letter index:
//
//  0, 1, 2, 3 = 3, 2, 1, 0
func ComplementOf(k int, kmer Kmer) (c Kmer) {
	for i, j := uint(0), uint(k-1)*2; i <= j; i, j = i+2, j-2 {
		c |= (^(kmer >> (j - i)) & (3 << i)) | (^(kmer>>i)&3)<<j
	}

	return
}

// Convert a string of bases into a Kmer, returns an error if string length does not match word length
func (ki *Index) KmerOf(kmertext string) (kmer Kmer, err error) {
	if len(kmertext) != ki.k {
		return 0, ErrBadKmerTextLen
	}

	for _, v := range kmertext {
		x := ki.lookUp[v]
		if x < 0 {
			return 0, ErrBadKmerText
		}
		kmer = (kmer << 2) | Kmer(x)
	}

	return
}

// Return the Euclidean distance between two sequences measured by abolsolute kmer frequencies.
func Distance(a, b map[Kmer]float64) (dist float64) {
	c := make(map[Kmer]struct{}, len(a)+len(b))
	for k := range a {
		c[k] = struct{}{}
	}
	for k := range b {
		c[k] = struct{}{}
	}
	for k := range c {
		dist += math.Pow(a[k]-b[k], 2)
	}

	return math.Sqrt(dist)
}

// Confirm that a Build() is correct. Returns boolean indicating this and the number of kmers indexed.
func (ki *Index) Check() (ok bool, found int) {
	ok = true
	f := func(index *Index, position, kmer int) {
		hit := false
		var base Kmer
		if kmer == 0 {
			base = 0
		} else {
			base = index.finger[kmer-1]
		}
		for j := base; j < index.finger[kmer]; j++ {
			if index.pos[j] == position {
				found++
				hit = true
				break
			}
		}
		if !hit {
			ok = false
		}
	}

	if err := ki.ForEachKmerOf(ki.seq, 0, ki.seq.Len(), f); err != nil {
		ok = false
	}

	return
}

// Return a copy of the internal finger slice.
func (ki *Index) Finger() (f []Kmer) {
	f = make([]Kmer, len(ki.finger))
	copy(f, ki.finger)
	return
}

// Return a copy of the internal pos slice.
func (ki *Index) Pos() (p []int) {
	p = make([]int, len(ki.pos))
	copy(p, ki.pos)
	return
}
