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

package seqio_test

import (
	"github.com/kortschak/biogo/exp/seqio"
	"github.com/kortschak/biogo/exp/seqio/fasta"
	"github.com/kortschak/biogo/exp/seqio/fastq"
	"testing"
)

func TestSeqio(t *testing.T) {
	var (
		_ seqio.Reader = &fasta.Reader{}
		_ seqio.Reader = &fastq.Reader{}
		_ seqio.Writer = &fasta.Writer{}
		_ seqio.Writer = &fastq.Writer{}
	)
}
