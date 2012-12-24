// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Kmer indexing package based on Bob Edgar and Gene Meyers' approach used in PALS.
//
// Currently limited to Kmers of 15 nucleotides due to int constraints in Go.
package kmerindex

import (
	"code.google.com/p/biogo/bio"
	"code.google.com/p/biogo/exp/seq/linear"
	"code.google.com/p/biogo/util"
	"fmt"
	"math"
)

var Debug = false // Set Debug to true to prevent recovering from panics in ForEachKmer f Eval function.

var MaxKmerLen = 15

// 2-bit per base packed word
type Kmer uint32 // Sensible size for word type uint64 will double the size of the index (already large for high k)

// Kmer index type
type Index struct {
	finger  []Kmer
	pos     []int
	Seq     *linear.Seq
	k       int
	kMask   Kmer
	indexed bool
}

var (
	lookUp     util.CTL
	MinKmerLen = 4 // default minimum
)

func init() {
	m := make(map[int]int)

	for i, v := range bio.N {
		m[int(v)] = i % 4
	}

	lookUp = *util.NewCTL(m)
}

// Create a new Kmer Index with a word size k based on sequence
func New(k int, sequence *linear.Seq) (i *Index, err error) {
	switch {
	case k > MaxKmerLen:
		return nil, bio.NewError("k greater than MaxKmerLen", 0, k, MaxKmerLen)
	case k < MinKmerLen:
		return nil, bio.NewError("k less than MinKmerLen", 0, k, MinKmerLen)
	case k+1 > sequence.Len():
		return nil, bio.NewError("sequence shorter than k+1-mer length", 0, k+1, sequence.Len())
	}

	i = &Index{
		finger:  make([]Kmer, util.Pow4(k)+1), // Need a Tn+1 finger position so that Tn can be recognised
		k:       k,
		kMask:   Kmer(util.Pow4(k) - 1),
		Seq:     sequence,
		indexed: false,
	}

	i.buildKmerTable()

	return
}

// Build the table of Kmer frequencies - called by New
func (self *Index) buildKmerTable() {
	incrementFinger := func(index *Index, _, kmer int) {
		index.finger[kmer]++
	}
	self.ForEachKmerOf(self.Seq, 0, self.Seq.Len(), incrementFinger)
}

// Build the Kmer position table destructively replacing Kmer frequencies
func (self *Index) Build() {
	var sum Kmer
	for i, v := range self.finger {
		self.finger[i], sum = sum, sum+v
	}

	locatePositions := func(index *Index, position, kmer int) {
		index.pos[index.finger[kmer]] = position
		index.finger[kmer]++
	}
	self.pos = make([]int, self.Seq.Len()-self.k+1)
	self.ForEachKmerOf(self.Seq, 0, self.Seq.Len(), locatePositions)

	self.indexed = true
}

// Return an array of positions for the Kmer string kmertext
func (self *Index) GetPositionsString(kmertext string) (positions []int, err error) {
	switch {
	case len(kmertext) != self.k:
		return nil, bio.NewError("Sequence length does not match Kmer length", 0, self.k, kmertext)
	case !self.indexed:
		return nil, bio.NewError("Index not built: call Build()", 0, self)
	}

	var kmer Kmer
	if kmer, err = self.KmerOf(kmertext); err != nil {
		return nil, err
	}

	return self.GetPositionsKmer(kmer)
}

// Return an array of positions for the Kmer kmer
func (self *Index) GetPositionsKmer(kmer Kmer) (positions []int, err error) {
	if kmer > self.kMask {
		return nil, bio.NewError("Kmer out of range", 0, kmer, self.kMask)
	}

	i := Kmer(0)
	if kmer > 0 { // special case: An has no predecessor
		i = self.finger[kmer-1]
	}
	j := self.finger[kmer]
	if i == j {
		return
	}

	positions = make([]int, j-i)
	for l, p := range self.pos[i:j] {
		positions[l] = int(p)
	}

	return
}

// Return a map containing absolute Kmer frequencies and true if called before Build().
// If called after Build returns a nil map and false.
func (self *Index) KmerFrequencies() (map[Kmer]int, bool) {
	if self.indexed {
		return nil, false
	}

	m := map[Kmer]int{}

	for i, f := range self.finger {
		if f > 0 {
			m[Kmer(i)] = int(f) // not always safe - perhaps check that Kmer <= MaxInt
		}
	}

	return m, true
}

// Return a map containing relative Kmer frequencies and true if called before Build().
// If called after Build returns a nil map and false.
func (self *Index) NormalisedKmerFrequencies() (map[Kmer]float64, bool) {
	if self.indexed {
		return nil, false
	}

	m := map[Kmer]float64{}

	l := float64(self.Seq.Len())
	for i, f := range self.finger {
		if f > 0 {
			m[Kmer(i)] = float64(f) / l
		}
	}

	return m, true
}

// Returns a Kmer-keyed map containing slices of kmer positions and true if called after Build,
// otherwise nil and false.
func (self *Index) KmerIndex() (map[Kmer][]int, bool) {
	if !self.indexed {
		return nil, false
	}

	m := make(map[Kmer][]int)

	for i := range self.finger {
		if p, _ := self.GetPositionsKmer(Kmer(i)); len(p) > 0 {
			m[Kmer(i)] = p
		}
	}

	return m, true
}

// Returns a string-keyed map containing slices of kmer positions and true if called after Build,
// otherwise nil and false.
func (self *Index) StringKmerIndex() (map[string][]int, bool) {
	if !self.indexed {
		return nil, false
	}

	m := make(map[string][]int)

	for i := range self.finger {
		if p, _ := self.GetPositionsKmer(Kmer(i)); len(p) > 0 {
			m[self.Stringify(Kmer(i))] = p
		}
	}

	return m, true
}

// errors should be handled through a panic which will be recovered by ForEachKmerOf
type Eval func(index *Index, j, kmer int)

// Applies the f Eval func to all kmers in s from start to end. Returns any panic raised by f as an error.
func (self *Index) ForEachKmerOf(s *linear.Seq, start, end int, f Eval) (err error) {
	defer func() {
		if !Debug {
			if r := recover(); r != nil {
				var ok bool
				err, ok = r.(error)
				if !ok {
					err = bio.NewError(fmt.Sprintf("pkg: %v", r), 1, r)
				}
			}
		}
	}()

	kmer := Kmer(0)
	high := 0
	var currentBase int

	// Preload the first k-1 bases of the first well defined k-mer or set high to the next position
	basePosition := start
	for ; basePosition < start+self.k-1; basePosition++ {
		currentBase = lookUp.ValueToCode[s.Seq[basePosition]]
		if currentBase >= 0 {
			kmer = (kmer << 2) | Kmer(currentBase)
		} else {
			kmer = 0
			high = basePosition + 1
		}
	}

	// Call f(position, kmer) for each of the next well defined k-mers
	for position := basePosition - self.k + 1; basePosition < end; position++ {
		currentBase = lookUp.ValueToCode[s.Seq[basePosition]]
		basePosition++
		if currentBase >= 0 {
			kmer = ((kmer << 2) | Kmer(currentBase)) & self.kMask
		} else {
			kmer = 0
			high = basePosition
		}
		if position >= high {
			f(self, position, int(kmer))
		}
	}

	return
}

// Return the Kmer length of the Index.
func (self *Index) GetK() int {
	return self.k
}

// Returns a pointer to the indexed seq.Seq.
func (self *Index) GetSeq() *linear.Seq {
	return self.Seq
}

// Returns the value of the finger slice at p. This signifies the absolute kmer frequency of the Kmer(p)
// if called before Build() and points to the relevant position lookup if called after.
func (self *Index) FingerAt(p int) int {
	return int(self.finger[p])
}

// Returns the value of the pos slice at p. This signifies the position of the pth kmer if called after Build().
// Not valid before Build() - will panic.
func (self *Index) PosAt(p int) int {
	return self.pos[p]
}

// Convert a Kmer into a string of bases
func (self *Index) Stringify(kmer Kmer) string {
	return Stringify(self.k, kmer)
}

// Convert a string of bases into a len k Kmer, returns an error if string length does not match k
func KmerOf(k int, kmertext string) (kmer Kmer, err error) {
	if len(kmertext) != k {
		return 0, bio.NewError("Sequence length does not match Kmer length", 0, k, kmertext)
	}

	for _, v := range kmertext {
		x := lookUp.ValueToCode[v]
		if x < 0 {
			return 0, bio.NewError("Kmer contains illegal character", 0, kmertext)
		}
		kmer = (kmer << 2) | Kmer(x)
	}

	return
}

// Return the GC fraction of a Kmer
func (self *Index) GCof(kmer Kmer) float64 {
	return GCof(self.k, kmer)
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
func Stringify(k int, kmer Kmer) string {
	kmertext := make([]byte, k)

	for i := k - 1; i >= 0; i, kmer = i-1, kmer>>2 {
		kmertext[i] = bio.N[kmer&3]
	}

	return string(kmertext)
}

// Reverse complement a Kmer
func (self *Index) ComplementOf(kmer Kmer) (c Kmer) {
	return ComplementOf(self.k, kmer)
}

// Reverse complement a Kmer of len k
func ComplementOf(k int, kmer Kmer) (c Kmer) {
	for i, j := uint(0), uint(k-1)*2; i <= j; i, j = i+2, j-2 {
		c |= (^(kmer >> (j - i)) & (3 << i)) | (^(kmer>>i)&3)<<j
	}

	return
}

// Convert a string of bases into a Kmer, returns an error if string length does not match word length
func (self *Index) KmerOf(kmertext string) (kmer Kmer, err error) {
	if len(kmertext) != self.k {
		return 0, bio.NewError("Sequence length does not match Kmer length", 0, self.k, kmertext)
	}

	for _, v := range kmertext {
		x := lookUp.ValueToCode[v]
		if x < 0 {
			return 0, bio.NewError("Kmer contains illegal character", 0, kmertext)
		}
		kmer = (kmer << 2) | Kmer(x)
	}

	return
}

// Return the Euclidian distance between two sequences measured by abolsolute kmer frequencies.
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
func (self *Index) Check() (ok bool, found int) {
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

	if err := self.ForEachKmerOf(self.Seq, 0, self.Seq.Len(), f); err != nil {
		ok = false
	}

	return
}

// Return a copy of the internal finger slice.
func (self *Index) Finger() (f []Kmer) {
	f = make([]Kmer, len(self.finger))
	copy(f, self.finger)
	return
}

// Return a copy of the internal pos slice.
func (self *Index) Pos() (p []int) {
	p = make([]int, len(self.pos))
	copy(p, self.pos)
	return
}
