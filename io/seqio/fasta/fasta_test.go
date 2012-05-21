package fasta

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
	"github.com/kortschak/biogo/seq"
	"io"
	"io/ioutil"
	check "launchpad.net/gocheck"
	"os"
	"testing"
)

var (
	fas    = []string{"../../testdata/testaln.fasta", "../../testdata/testaln2.fasta"}
	single = "../../testdata/test.fasta"
)

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

var (
	expectN = []string{
		"AK1H_ECOLI/114-431 DESCRIPTION HERE",
		"AKH_HAEIN 114-431",
		"AKH1_MAIZE/117-440",
		"AK2H_ECOLI/112-431",
		"AK1_BACSU/66-374",
		"AK2_BACST/63-370",
		"AK2_BACSU/63-373",
		"AKAB_CORFL/63-379",
		"AKAB_MYCSM/63-379",
		"AK3_ECOLI/106-407",
		"AK_YEAST/134-472 A COMMENT FOR YEAST",
	}

	expectS = [][]byte{
		[]byte("CPDSINAALICRGEKMSIAIMAGVLEARGH-N--VTVIDPVEKLLAVG-HYLESTVDIAESTRRIAASRIP------A-DHMVLMAGFTAGN-EKGELVVLGRNGSDYSAAVLAACLRADCCEIWTDVNGVYTCDP-------------RQVPDARLLKSMSYQEAMELSY--FGAKVLHPRTITPIAQFQIPCLIKNTGNPQAPGTL-IG--ASRDEDELP----VKGISNLN------NMAMFSVSGP-GMKGMVGMAARVFAAMS-------RARISVVLITQSSSEYSISFCVPQSDCVRAERAMLEEFY-----LELKEGLLEPLAVAERLAIISV-VGDGLRTLRGISAKF------FAALARANINIVAIA"),
		[]byte("-----------------VEDAVKATIDCRGEKLSIAMMKAWFEARGY-S--VHIVDPVKQLLAKG-GYLESSVEIEESTKRVDAANIA--K-DKVVLMAGF---TAGNEKGELVLLGRNGSDYSAAC-----------------LAACLGASVCEIWTDVDGVYTCDP--RLVPDARLLPTLSYREAMELSYFGAKVIHPRTIGPLLPQNIPCVIKNTGNPSAPGSI-ID--GNVKSESLQ----VKGITNLDNLAMFNVSGPGMQGM---VGMASRVFSAMSGAGISVILITQSSSEYS---ISFCVPVKSAEVAKTVLETEFA-----NELNEHQLEPIEVIKDLSIISV-VGDGMKQAKGIAARF------FSALAQANISIVAIA"),
		[]byte("-----------------ATESFSDFVVGHGELWSAQMLSYAIQKSGT-P--CSWMDTREVLVVNPSGANQVDPDYLESEKRLEKWFSRC-P-AETIIATGF---IASTPENIPTTLKRDGSDFSAAI-----------------IGSLVKARQVTIWTDVDGVFSADP--RKVSEAVILSTLSYQEAWEMSYFGANVLHPRTIIPVMKYNIPIVIRNIFNTSAPGTM-IC--QQPANENGDLEACVKAFATIDKLALVNVEGTGMAGV---PGTANAIFGAVKDVGANVIMISQASSEHS---VCFAVPEKEVALVSAALHARFR-----EALAAGRLSKVEVIHNCSILAT-VGLRMASTPGVSATL------FDALAKANINVRAIA"),
		[]byte("-----------------INDAVYAEVVGHGEVWSARLMSAVLNQQG-----LPAAWLDAREFLRAERAAQPQVDEGLSYPLLQQLLVQH-P-GKRLVVTGF---ISRNNAGETVLLGRNGSDYSATQ-----------------IGALAGVSRVTIWSDVAGVYSADP--RKVKDACLLPLLRLDEASELARLAAPVLHARTLQPVSGSEIDLQLRCSYTPDQGSTRIERVLASGTGARIVTSHDDVCLI-EFQVPASQDFKLAHKEI--DQILKRAQVRPLAVGVHNDRQLLQFCYTSEVADSALKILDEAG---------LPGELRLRQGLALVAMVGAGVTRNPLHCHRFWQQLKGQPVEFTWQSDDGISLVAVL"),
		[]byte("-----------------ISPREQDLLLSCGETISSVVFTSMLLDNGVKA--AALTGAQAGFLTNDQHTNAKIIEMKPER--LFSVLAN----HDAVVVAGF---QGATEKGDTTTIGRGGSDTSAAA-----------------LGAAVDAEYIDIFTDVEGVMTADP--RVVENAKPLPVVTYTEICNLAYQGAKVISPRAVEIAMQAKVPIRVRSTYS-NDKGTLVTSHHSSKVGSDVFERLITGIAH-VKDVTQFKVPAKIGQYN-----VQTEVFKAMANAGISVDFFNITPSEIVYTVAGNKTETAQR------------ILMDMGYDPMVTRNCAKVSAVGAGIMGVPGVTSKI------VSALSEKEIPILQSA"),
		[]byte("-----------------KRE--MDMLLSTGEQVSIALLAMSLHEKGYKA--VSLTGWQAGITTEEMHGNARIMNIDTT--RIRRCLDE----GAIVIVAGF---QGVTETGEITTLGRGGSDTTAVA-----------------LAAALKAEKCDIYTDVTGVFTTDP--RYVKTARKIKEISYDEMLELANLGAGVLHPRAVEFAKNYEVPLEVRSSME-NERGTMVK--EEVSMEQHLIVRGIAFEDQ-VTRVTVVGIEKYLQSVA--------TIFTALANRGINVDIIIQNA--------------------TNSETAS--VSFSIRTEDLPETLQVLQ-------------ALEGADVHYESGLAKVSI-VGSGMISNPGVAARV------FEVLADQGIEIKMVS"),
		[]byte("-----------------KRE--MDMLLATGEQVTISLLSMALQEKGYDA--VSYTGWQAGIRTEAIHGNARITDIDTS--VLADQLEK----GKIVIVAGF---QGMTEDCEITTLGRGGSDTTAVA-----------------LAAALKVDKCDIYTDVPGVFTTDP--RYVKSARKLEGISYDEMLELANLGAGVLHPRAVEFAKNYQVPLEVRSSTE-TEAGTLIE--EESSMEQNLIVRGIAFEDQ-ITRVTIYGLTSGLTTLS--------TIFTTLAKRNINVDIIIQTQ--------------------AEDKTG---ISFSVKTEDADQTVAVLEEYK---------DALEFEKIETESKLAKVSI-VGSGMVSNPGVAAEM------FAVLAQKNILIKMVS"),
		[]byte("-----------------ARE--MDMLLTAGERISNALVAMAIESLGAEA--QSFTGSQAGVLTTERHGNARIVDVTPG--RVREALDE----GKICIVAGF--QGVNKETRDVTTLGRGGSDTTAVA-----------------LAAALNADVCEIYSDVDGVYTADP--RIVPNAQKLEKLSFEEMLELAAVGSKILVLRSVEYARAFNVPLRVRSSYS-NDPGTLIAGSMEDIPVEEAVLTGVATDKS-EAKVTVLGISDKPGEAA--------KVFRALADAEINIDMVLQNV--------------------SSVEDGTTDITFTCPRADGRRAMEILKKLQ---------VQGNWTNVLYDDQVDKVSL-VGAGMKSHPGVTAEF------MEALRDVNVNIELIS"),
		[]byte("-----------------PRE--MDMLLTAGERISNALVAMAIESLGAQA--RSFTGSQAGVITTGTHGNAKIIDVTPG--RLRDALDE----GQIVLVAGF--QGVSQDSKDVTTLGRGGSDTTAVA-----------------VAAALDADVCEIYTDVDGIFTADP--RIVPNARHLDTVSFEEMLEMAACGAKVLMLRCVEYARRYNVPIHVRSSYS-DKPGTIVKGSIEDIPMEDAILTGVAHDRS-EAKVTVVGLPDVPGYAA--------KVFRAVAEADVNIDMVLQNI--------------------SKIEDGKTDITFTCARDNGPRAVEKLSALK---------SEIGFSQVLYDDHIGKVSL-IGAGMRSHPGVTATF------CEALAEAGINIDLIS"),
		[]byte("-----------------TSPALTDELVSHGELMSTLLFVEILRERD--V--QAQWFDVRKVMRTNDRFGRAEPDIAALAELAALQLLPR-LNEGLVITQGF---IGSENKGRTTTLGRGGSDYTAAL-----------------LAEALHASRVDIWTDVPGIYTTDP--RVVSAAKRIDEIAFAEAAEMATFGAKVLHPATLLPAVRSDIPVFVGSSKDPRAGGTLVCNKTENPPLFRALAL--RRNQT-LLTLHSLNMLHSRGFLA--------EVFGILARHNISVDLITTSEVSVALTLDTTGSTSTG----------DTLLTQSLLMELSALCRVEVEEGLALVALIG----------NDLSKACGVGKEVF"),
		[]byte("-----------------VSSRTVDLVMSCGEKLSCLFMTALCNDRGCKAKYVDLSHIVPSDFSASALDNSFYTFLVQALKEKLAPFVSA-KERIVPVFTGF---FGLVPTGLLNGVGRGYTDLCAAL-----------------IAVAVNADELQVWKEVDGIFTADP--RKVPEARLLDSVTPEEASELTYYGSEVIHPFTMEQVIRAKIPIRIKNVQNPLGNGTIIYPDNVAKKGESTPPHPPENLSS----SFYEKRKRGATAITTKN----DIFVINIHSNKKTLSHGFLAQIFTILDKYKLVVDLISTSEVHVSMALPIPDADS-LKSLRQAEEKLRILGSVDITKKLSIVSLVGKHMKQYIGIAG---TMFTTLAEEGINIEMIS"),
	}
)

func (s *S) TestReadFasta(c *check.C) {
	var (
		obtainN []string
		obtainS [][]byte
	)

	for _, fa := range fas {
		if r, err := NewReaderName(fa); err != nil {
			c.Fatalf("Failed to open %q: %s", fa, err)
		} else {
			for i := 0; i < 3; i++ {
				for {
					if s, err := r.Read(); err != nil {
						if err == io.EOF {
							break
						} else {
							c.Fatalf("Failed to read %q: %s", fa, err)
						}
					} else {
						obtainN = append(obtainN, s.ID)
						obtainS = append(obtainS, s.Seq)
					}
				}
				c.Check(obtainN, check.DeepEquals, expectN)
				obtainN = nil
				c.Check(obtainS, check.DeepEquals, expectS)
				obtainS = nil
				if err = r.Rewind(); err != nil {
					c.Fatalf("Failed to rewind %s", err)
				}
			}
			r.Close()
		}
	}
}

func (s *S) TestWriteFasta(c *check.C) {
	fa := fas[0]
	o := c.MkDir()
	expectSize := 4649
	var total int
	if w, err := NewWriterName(o+"/fa", 60); err != nil {
		c.Fatalf("Failed to open %q for write: %s", o+"/fa", err)
	} else {
		s := &seq.Seq{}

		for i := range expectN {
			s.ID = expectN[i]
			s.Seq = expectS[i]
			if n, err := w.Write(s); err != nil {
				c.Fatalf("Failed to write %q: %s", o+"/fa", err)
			} else {
				total += n
			}
		}

		if err = w.Close(); err != nil {
			c.Fatalf("Failed to Close %q: %s", o+"/fa", err)
		}
		c.Check(total, check.Equals, expectSize)
		total = 0

		var (
			of, gf *os.File
			ob, gb []byte
		)
		if of, err = os.Open(fa); err != nil {
			c.Fatalf("Failed to Open %q: %s", fa, err)
		}
		if gf, err = os.Open(o + "/fa"); err != nil {
			c.Fatalf("Failed to Open %q: %s", o+"/fa", err)
		}
		if ob, err = ioutil.ReadAll(of); err != nil {
			c.Fatalf("Failed to read %q: %s", fa, err)
		}
		if gb, err = ioutil.ReadAll(gf); err != nil {
			c.Fatalf("Failed to read %q: %s", o+"/fa", err)
		}

		c.Check(gb, check.DeepEquals, ob)
	}
}

func (s *S) TestReadOneFasta(c *check.C) {
	if r, err := NewReaderName(single); err != nil {
		c.Fatalf("Failed to open %q: %s", single, err)
	} else {
		defer r.Close()
		for {
			if s, err := r.Read(); err != nil {
				if err == io.EOF {
					break
				} else {
					c.Fatalf("Failed to read %q: %s", single, err)
				}
			} else {
				c.Check(s.ID, check.Equals, "AK1H_ECOLI/114-431 DESCRIPTION HERE")
				c.Check(s.Len(), check.Equals, 378)
			}
		}
	}
}
