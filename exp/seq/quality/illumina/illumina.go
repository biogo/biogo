//Package illumina provides support for manipulation of quality data in Illumina format.
package illumina

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
	"github.com/kortschak/biogo/exp/alphabet"
	"github.com/kortschak/biogo/exp/seq"
	"github.com/kortschak/biogo/exp/seq/quality"
)

type Illumina1_3 struct {
	*quality.Phred
}

func NewIllumina1_3(id string, q []alphabet.Qphred) *Illumina1_3 {
	return &Illumina1_3{quality.NewPhred(id, q, alphabet.Illumina1_3)}
}

func (self *Illumina1_3) Join(p *Illumina1_3, where int) (err error) {
	return self.Phred.Join(p.Phred, where)
}

func (self *Illumina1_3) Copy() seq.Quality {
	return &Illumina1_3{self.Phred.Copy().(*quality.Phred)}
}

type Illumina1_5 struct {
	*quality.Phred
}

func NewIllumina1_5(id string, q []alphabet.Qphred) *Illumina1_5 {
	return &Illumina1_5{quality.NewPhred(id, q, alphabet.Illumina1_5)}
}

func (self *Illumina1_5) Join(p *Illumina1_5, where int) (err error) {
	return self.Phred.Join(p.Phred, where)
}

func (self *Illumina1_5) Copy() seq.Quality {
	return &Illumina1_5{self.Phred.Copy().(*quality.Phred)}
}

type Illumina1_8 struct {
	*quality.Phred
}

func NewIllumina1_8(id string, q []alphabet.Qphred) *Illumina1_8 {
	return &Illumina1_8{quality.NewPhred(id, q, alphabet.Illumina1_8)}
}

func (self *Illumina1_8) Join(p *Illumina1_8, where int) (err error) {
	return self.Phred.Join(p.Phred, where)
}

func (self *Illumina1_8) Copy() seq.Quality {
	return &Illumina1_8{self.Phred.Copy().(*quality.Phred)}
}

type Illumina1_9 struct {
	*quality.Phred
}

func NewIllumina1_9(id string, q []alphabet.Qphred) *Illumina1_9 {
	return &Illumina1_9{quality.NewPhred(id, q, alphabet.Illumina1_9)}
}

func (self *Illumina1_9) Join(p *Illumina1_9, where int) (err error) {
	return self.Phred.Join(p.Phred, where)
}

func (self *Illumina1_9) Copy() seq.Quality {
	return &Illumina1_9{self.Phred.Copy().(*quality.Phred)}
}
