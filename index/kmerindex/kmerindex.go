// Kmer indexing package based on Bob Edgar and Gene Meyers' approach used in PALS.
//
// Currently limited to Kmers of 15 nucleotides due to int constraints in Go.
package kmerindex
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
	"fmt"
	"github.com/kortschak/BioGo/bio"
	"github.com/kortschak/BioGo/seq"
	"github.com/kortschak/BioGo/util"
	"math"
)

var Debug bool

var MaxKmerLen = 15

// 2-bit per base packed word
type Kmer uint32 // Sensible size for word type uint64 will double the size of the index (already large for high k)

// Kmer index type
type Index struct {
	finger  []Kmer
	pos     []int
	Seq     *seq.Seq
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
func New(k int, sequence *seq.Seq) (i *Index, err error) {
	switch {
	case k > MaxKmerLen:
		return nil, bio.NewError("k greater than MaxKmerLen", 0, []int{k, MaxKmerLen})
	case k < MinKmerLen:
		return nil, bio.NewError("k less than MinKmerLen", 0, []int{k, MinKmerLen})
	case k+1 > sequence.Len():
		return nil, bio.NewError("sequence shorter than k+1-mer length", 0, []int{k + 1, sequence.Len()})
	}

	i = &Index{
		finger:  make([]Kmer, util.Pow4(k)+1), // Need a Tn+1 finger position so that Tn can be recognised
		pos:     make([]int, sequence.Len()-k+1),
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
	self.ForEachKmerOf(self.Seq, 0, self.Seq.Len(), locatePositions)

	self.indexed = true
}

// Return an array of positions for the Kmer string kmertext
func (self *Index) GetPositionsString(kmertext string) (positions []int, err error) {
	switch {
	case len(kmertext) != self.k:
		return nil, bio.NewError("Sequence length does not match Kmer length", 0, fmt.Sprintf("%d:%s", self.k, kmertext))
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
		return nil, bio.NewError("Kmer out of range", 0, []Kmer{kmer, self.kMask})
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

func (self *Index) KmerFrequencies() (map[Kmer]int, bool) {
	if self.indexed {
		return nil, false
	}

	m := map[Kmer]int{}

	for i := Kmer(0); i < self.kMask; i++ {
		if self.finger[i] > 0 {
			m[i] = int(self.finger[i]) // not always safe - perhaps check that Kmer <= MaxInt
		}
	}

	return m, true
}

func (self *Index) NormalisedKmerFrequencies() (map[Kmer]float64, bool) {
	if self.indexed {
		return nil, false
	}

	m := map[Kmer]float64{}

	for i := Kmer(0); i <= self.kMask; i++ {
		if self.finger[i] > 0 {
			m[i] = float64(self.finger[i]) / float64(self.Seq.Len())
		}
	}

	return m, true
}

func (self *Index) StringKmerIndex() (map[string][]int, bool) {
	if !self.indexed {
		return nil, false
	}

	m := make(map[string][]int)

	for i := Kmer(0); i <= self.kMask; i++ {
		if p, _ := self.GetPositionsKmer(i); len(p) > 0 {
			m[self.StringOf(i)] = p
		}
	}

	return m, true
}

func (self *Index) KmerIndex() (map[Kmer][]int, bool) {
	if !self.indexed {
		return nil, false
	}

	m := make(map[Kmer][]int)

	for i := Kmer(0); i <= self.kMask; i++ {
		if p, _ := self.GetPositionsKmer(i); len(p) > 0 {
			m[i] = p
		}
	}

	return m, true
}

// errors should be handled through a panic which will be recovered by ForEachKmerOf
type Eval func(index *Index, j, kmer int)

func (self *Index) ForEachKmerOf(s *seq.Seq, start, end int, f Eval) (err error) {
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

func (self *Index) GetK() int {
	return self.k
}

func (self *Index) GetSeq() *seq.Seq {
	return self.Seq
}

func (self *Index) FingerAt(p int) int {
	return int(self.finger[p])
}

func (self *Index) PosAt(p int) int {
	return self.pos[p]
}

// Convert a Kmer into a string of bases
func (self *Index) StringOf(kmer Kmer) string {
	kmertext := make([]byte, self.k)

	for i := self.k - 1; i >= 0; i, kmer = i-1, kmer>>2 {
		kmertext[i] = bio.N[kmer&3]
	}

	return string(kmertext)
}

// Return the GC fraction of a Kmer
func (self *Index) GCof(kmer Kmer) float32 {
	gc := 0
	for i := self.k - 1; i >= 0; i, kmer = i-1, kmer>>2 {
		gc += int((kmer & 1) ^ (kmer & 2))
	}

	return float32(gc) / float32(self.k)
}

// Convert a Kmer into a string of bases
func StringOfLen(k int, kmer Kmer) string {
	kmertext := make([]byte, k)

	for i := k - 1; i >= 0; i, kmer = i-1, kmer>>2 {
		kmertext[i] = bio.N[kmer&3]
	}

	return string(kmertext)
}

// Return the GC fraction of a Kmer
func GCof(k int, kmer Kmer) float32 {
	gc := 0
	for i := k - 1; i >= 0; i, kmer = i-1, kmer>>2 {
		gc += int((kmer & 1) ^ (kmer & 2))
	}

	return float32(gc) / float32(k)
}

// Reverse complement a Kmer
func (self *Index) ComplementOf(kmer Kmer) (c Kmer) {
	for i, j := uint(0), uint(self.k-1)*2; i < j; i, j = i+2, j-2 {
		c |= (^(kmer >> (j - i)) & (3 << i)) | (^(kmer>>i)&3)<<j
	}

	return
}

// Reverse complement a Kmer of len k
func ComplementOfLen(k int, kmer Kmer) (c Kmer) {
	for i, j := uint(0), uint(k-1)*2; i < j; i, j = i+2, j-2 {
		c |= (^(kmer >> (j - i)) & (3 << i)) | (^(kmer>>i)&3)<<j
	}

	return
}

// Convert a string of bases into a Kmer, returns an error if string length does not match word length
func (self *Index) KmerOf(kmertext string) (kmer Kmer, err error) {
	if len(kmertext) != self.k {
		return 0, bio.NewError("Sequence length does not match Kmer length", 0, fmt.Sprintf("%d:%s", self.k, kmertext))
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
	self.ForEachKmerOf(self.Seq, 0, self.Seq.Len(), f)

	return
}

func (self *Index) Finger() (f []Kmer) {
	f = make([]Kmer, len(self.finger))
	copy(f, self.finger)
	return
}

func (self *Index) Pos() (p []int) {
	p = make([]int, len(self.pos))
	copy(p, self.pos)
	return
}
