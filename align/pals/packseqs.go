package pals
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
	"bytes"
	"fmt"
	"github.com/kortschak/BioGo/seq"
	"github.com/kortschak/BioGo/util"
)

type Packer struct {
	Packed  *seq.Seq
	lastPad int
	length  int
}

func NewPacker(id string) (p *Packer) {
	return &Packer{
		Packed: &seq.Seq{
			ID:     []byte(id),
			Strand: 1,
			Meta:   SeqMap{},
		},
	}
}

func (self *Packer) Pack(sequence *seq.Seq) string {
	m := self.Packed.Meta.(SeqMap)

	contig := Contig{seq: sequence}

	padding := binSize - sequence.Len()%binSize
	if padding < minPadding {
		padding += binSize
	}

	//contig.from = self.Packed.Len() + 1
	contig.from = self.length + 1
	self.length += self.lastPad + sequence.Len()
	//self.Packed.Seq = []byte(string(self.Packed.Seq) + strings.Repeat("N", self.lastPad) + string(sequence.Seq))
	self.lastPad = padding

	bins := make([]int, (padding+sequence.Len())/binSize)
	for i := 0; i < len(bins); i++ {
		bins[i] = len(m.contigs)
	}

	m.binMap = append(m.binMap, bins...)
	m.contigs = append(m.contigs, contig)
	self.Packed.Meta = m

	return fmt.Sprintf("%20s\t%10d\t%7d-%-d", sequence.ID[:util.Min(20, len(sequence.ID))], sequence.Len(), len(m.binMap)-len(bins), len(m.binMap)-1)
}

func (self *Packer) FinalisePack() {
	lastPad := 0
	self.Packed.Seq = make([]byte, 0, self.length)
	for _, contig := range self.Packed.Meta.(SeqMap).contigs {
		padding := binSize - contig.seq.Len()%binSize
		if padding < minPadding {
			padding += binSize
		}
		self.Packed.Seq = append(self.Packed.Seq, bytes.Repeat([]byte("N"), lastPad)...)
		self.Packed.Seq = append(self.Packed.Seq, contig.seq.Seq...)
		lastPad = padding
	}
}
