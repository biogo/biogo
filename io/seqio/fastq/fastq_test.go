package fastq

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
	"github.com/kortschak/BioGo/seq"
	"io"
	"io/ioutil"
	check "launchpad.net/gocheck"
	"os"
	"testing"
)

var (
	fqs = []string{"../../testdata/testaln.fastq", "../../testdata/testaln2.fastq"}
)

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

var (
	expectN = []string{
		"FC12044_91407_8_200_406_24",
		"FC12044_91407_8_200_720_610",
		"FC12044_91407_8_200_345_133",
		"FC12044_91407_8_200_106_131",
		"FC12044_91407_8_200_916_471",
		"FC12044_91407_8_200_57_85",
		"FC12044_91407_8_200_10_437",
		"FC12044_91407_8_200_154_436",
		"FC12044_91407_8_200_336_64",
		"FC12044_91407_8_200_620_233",
		"FC12044_91407_8_200_902_349",
		"FC12044_91407_8_200_40_618",
		"FC12044_91407_8_200_83_511",
		"FC12044_91407_8_200_76_246",
		"FC12044_91407_8_200_303_427",
		"FC12044_91407_8_200_31_299",
		"FC12044_91407_8_200_553_135",
		"FC12044_91407_8_200_139_74",
		"FC12044_91407_8_200_108_33",
		"FC12044_91407_8_200_980_965",
		"FC12044_91407_8_200_981_857",
		"FC12044_91407_8_200_8_865",
		"FC12044_91407_8_200_292_484",
		"FC12044_91407_8_200_675_16",
		"FC12044_91407_8_200_285_136",
	}

	expectS = [][]byte{
		[]byte("GTTAGCTCCCACCTTAAGATGTTTA"),
		[]byte("CTCTGTGGCACCCCATCCCTCACTT"),
		[]byte("GATTTTTTAACAATAAACGTACATA"),
		[]byte("GTTGCCCAGGCTCGTCTTGAACTCC"),
		[]byte("TGATTGAAGGTAGGGTAGCATACTG"),
		[]byte("GCTCCAATAGCGCAGAGGAAACCTG"),
		[]byte("GCTGCTTGGGAGGCTGAGGCAGGAG"),
		[]byte("AGACCTTTGGATACAATGAACGACT"),
		[]byte("AGGGAATTTTAGAGGAGGGCTGCCG"),
		[]byte("TCTCCATGTTGGTCAGGCTGGTCTC"),
		[]byte("TGAACGTCGAGACGCAAGGCCCGCC"),
		[]byte("CTGTCCCCACGGCGGGGGGGCCTGG"),
		[]byte("GATGTACTCTTACACCCAGACTTTG"),
		[]byte("TCAAGGGTGGATCTTGGCTCCCAGT"),
		[]byte("TTGCGACAGAGTTTTGCTCTTGTCC"),
		[]byte("TCTGCTCCAGCTCCAAGACGCCGCC"),
		[]byte("TACGGAGCCGCGGGCGGGAAAGGCG"),
		[]byte("CCTCCCAGGTTCAAGCGATTATCCT"),
		[]byte("GTCATGGCGGCCCGCGCGGGGAGCG"),
		[]byte("ACAGTGGGTTCTTAAAGAAGAGTCG"),
		[]byte("AACGAGGGGCGCGACTTGACCTTGG"),
		[]byte("TTTCCCACCCCAGGAAGCCTTGGAC"),
		[]byte("TCAGCCTCCGTGCCCAGCCCACTCC"),
		[]byte("CTCGGGAGGCTGAGGCAGGGGGGTT"),
		[]byte("CCAAATCTTGAATTGTAGCTCCCCT"),
	}

	expectQ = [][]seq.Qsanger{
		{50, 55, 55, 51, 55, 55, 55, 55, 55, 55, 55, 55, 55, 51, 51, 50, 52, 55, 50, 50, 55, 42, 51, 44, 48},
		{46, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 51, 50, 55, 48, 51, 55, 52},
		{46, 48, 51, 46, 46, 50, 37, 46, 49, 51, 37, 37, 37, 40, 40, 46, 37, 37, 37, 37, 37, 37, 37, 37, 37},
		{55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 50, 55, 55, 55, 55, 40, 50, 51, 55, 48, 50},
		{55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 52, 55, 55, 52, 50, 55, 55, 51, 55, 54},
		{55, 37, 55, 44, 55, 50, 55, 55, 50, 55, 55, 55, 46, 50, 48, 49, 46, 46, 50, 49, 46, 37, 48, 40, 48},
		{52, 50, 55, 50, 55, 55, 55, 55, 55, 55, 52, 55, 55, 55, 50, 55, 48, 55, 55, 52, 48, 55, 55, 42, 50},
		{44, 42, 42, 44, 48, 51, 50, 49, 55, 44, 50, 48, 51, 46, 44, 49, 37, 46, 46, 40, 37, 37, 37, 37, 37},
		{50, 51, 48, 44, 46, 50, 55, 50, 55, 50, 48, 55, 48, 55, 55, 42, 55, 55, 55, 42, 37, 55, 37, 37, 42},
		{55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 50, 55, 50, 54},
		{55, 44, 55, 50, 50, 55, 44, 55, 55, 50, 55, 48, 50, 55, 51, 50, 48, 55, 37, 42, 50, 42, 51, 46, 37},
		{51, 55, 55, 55, 55, 50, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 49, 42, 37, 46, 55, 50},
		{50, 46, 55, 55, 55, 55, 55, 52, 55, 55, 55, 55, 55, 55, 48, 42, 48, 42, 42, 49, 46, 46, 48, 50, 52},
		{55, 51, 55, 51, 52, 55, 55, 55, 55, 55, 49, 55, 55, 55, 51, 55, 55, 50, 52, 55, 50, 49, 37, 55, 48},
		{55, 55, 48, 49, 46, 55, 55, 55, 55, 40, 55, 37, 48, 55, 55, 55, 46, 40, 48, 50, 50, 55, 52, 37, 37},
		{55, 49, 55, 51, 50, 55, 55, 55, 49, 55, 55, 50, 55, 48, 48, 46, 55, 48, 51, 50, 48, 50, 55, 42, 48},
		{55, 50, 48, 48, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 50, 55, 55, 44, 37, 37, 48, 55, 51, 42, 52},
		{49, 44, 55, 52, 50, 55, 51, 55, 55, 48, 55, 55, 48, 52, 55, 55, 55, 50, 48, 40, 50, 40, 50, 50, 46},
		{46, 46, 46, 50, 50, 55, 55, 50, 55, 55, 46, 44, 42, 44, 46, 37, 44, 42, 37, 46, 42, 37, 37, 37, 37},
		{51, 46, 50, 50, 49, 55, 55, 55, 50, 50, 44, 50, 55, 44, 46, 44, 55, 40, 49, 55, 46, 55, 37, 37, 50},
		{49, 55, 44, 50, 50, 55, 55, 55, 55, 50, 55, 48, 55, 48, 55, 37, 50, 55, 48, 37, 48, 42, 44, 55, 50},
		{55, 55, 55, 37, 42, 46, 49, 46, 44, 42, 46, 46, 49, 44, 40, 44, 49, 40, 40, 42, 42, 46, 49, 37, 37},
		{55, 48, 55, 46, 50, 55, 55, 55, 55, 55, 52, 55, 55, 55, 55, 40, 55, 55, 55, 55, 48, 51, 46, 55, 37},
		{46, 55, 51, 55, 55, 55, 50, 55, 55, 48, 55, 55, 46, 55, 55, 42, 44, 55, 55, 44, 55, 46, 42, 48, 37},
		{46, 50, 55, 46, 48, 55, 55, 55, 55, 55, 50, 55, 55, 52, 55, 55, 51, 55, 55, 55, 55, 51, 49, 44, 50},
	}
)

func (s *S) TestReadFastq(c *check.C) {
	var (
		obtainN []string
		obtainS [][]byte
		obtainQ [][]seq.Qsanger
	)

	for _, fq := range fqs {
		if r, err := NewReaderName(fq); err != nil {
			c.Fatalf("Failed to open %q: %s", fq, err)
		} else {
			for i := 0; i < 3; i++ {
				for {
					if s, err := r.Read(); err != nil {
						if err == io.EOF {
							break
						} else {
							c.Fatalf("Failed to read %q: %s", fq, err)
						}
					} else {
						obtainN = append(obtainN, s.ID)
						obtainS = append(obtainS, s.Seq)
						obtainQ = append(obtainQ, s.Quality.Qual)
					}
				}
				c.Check(obtainN, check.Equals, expectN)
				obtainN = nil
				c.Check(obtainS, check.Equals, expectS)
				obtainS = nil
				c.Check(obtainQ, check.Equals, expectQ)
				obtainQ = nil
				if err = r.Rewind(); err != nil {
					c.Fatalf("Failed to rewind %s", err)
				}
			}
			r.Close()
		}
	}
}

func (s *S) TestWriteFastq(c *check.C) {
	fq := fqs[0]
	o := c.MkDir()
	names := 0
	for _, n := range expectN {
		names += len(n)
	}
	expectSize := []int{2722, 2722 - names}
	var total int
	for j := 0; j < 2; j++ {
		if w, err := NewWriterName(o + "/fq"); err != nil {
			c.Fatalf("Failed to open %q for write: %s", o, err)
		} else {
			w.QID = j == 0
			s := &seq.Seq{Quality: &seq.Quality{}}

			for i := range expectN {
				s.ID = expectN[i]
				s.Seq = expectS[i]
				s.Quality.Qual = expectQ[i]
				if n, err := w.Write(s); err != nil {
					c.Fatalf("Failed to write %q: %s", o, err)
				} else {
					total += n
				}
			}

			if err = w.Close(); err != nil {
				c.Fatalf("Failed to Close %q: %s", o+"/fq", err)
			}
			c.Check(total, check.Equals, expectSize[j])
			total = 0

			if w.QID {
				var (
					of, gf *os.File
					ob, gb []byte
				)
				if of, err = os.Open(fq); err != nil {
					c.Fatalf("Failed to Open %q: %s", fq, err)
				}
				if gf, err = os.Open(o + "/fq"); err != nil {
					c.Fatalf("Failed to Open %q: %s", o+"/fq", err)
				}
				if ob, err = ioutil.ReadAll(of); err != nil {
					c.Fatalf("Failed to read %q: %s", fq, err)
				}
				if gb, err = ioutil.ReadAll(gf); err != nil {
					c.Fatalf("Failed to read %q: %s", o+"/fq", err)
				}

				c.Check(gb, check.Equals, ob)
			}
		}
	}
}
