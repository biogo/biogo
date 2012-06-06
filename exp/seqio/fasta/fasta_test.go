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

package fasta

import (
	"bytes"
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/seq/nucleic/packed"
	"code.google.com/p/biogo/exp/seq/protein"
	"fmt"
	"io"
	check "launchpad.net/gocheck"
	"testing"
)

var (
	fas = []string{testaln0, testaln1}
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

	expectS = [][]alphabet.Letter{
		[]alphabet.Letter("CPDSINAALICRGEKMSIAIMAGVLEARGH-N--VTVIDPVEKLLAVG-HYLESTVDIAESTRRIAASRIP------A-DHMVLMAGFTAGN-EKGELVVLGRNGSDYSAAVLAACLRADCCEIWTDVNGVYTCDP-------------RQVPDARLLKSMSYQEAMELSY--FGAKVLHPRTITPIAQFQIPCLIKNTGNPQAPGTL-IG--ASRDEDELP----VKGISNLN------NMAMFSVSGP-GMKGMVGMAARVFAAMS-------RARISVVLITQSSSEYSISFCVPQSDCVRAERAMLEEFY-----LELKEGLLEPLAVAERLAIISV-VGDGLRTLRGISAKF------FAALARANINIVAIA"),
		[]alphabet.Letter("-----------------VEDAVKATIDCRGEKLSIAMMKAWFEARGY-S--VHIVDPVKQLLAKG-GYLESSVEIEESTKRVDAANIA--K-DKVVLMAGF---TAGNEKGELVLLGRNGSDYSAAC-----------------LAACLGASVCEIWTDVDGVYTCDP--RLVPDARLLPTLSYREAMELSYFGAKVIHPRTIGPLLPQNIPCVIKNTGNPSAPGSI-ID--GNVKSESLQ----VKGITNLDNLAMFNVSGPGMQGM---VGMASRVFSAMSGAGISVILITQSSSEYS---ISFCVPVKSAEVAKTVLETEFA-----NELNEHQLEPIEVIKDLSIISV-VGDGMKQAKGIAARF------FSALAQANISIVAIA"),
		[]alphabet.Letter("-----------------ATESFSDFVVGHGELWSAQMLSYAIQKSGT-P--CSWMDTREVLVVNPSGANQVDPDYLESEKRLEKWFSRC-P-AETIIATGF---IASTPENIPTTLKRDGSDFSAAI-----------------IGSLVKARQVTIWTDVDGVFSADP--RKVSEAVILSTLSYQEAWEMSYFGANVLHPRTIIPVMKYNIPIVIRNIFNTSAPGTM-IC--QQPANENGDLEACVKAFATIDKLALVNVEGTGMAGV---PGTANAIFGAVKDVGANVIMISQASSEHS---VCFAVPEKEVALVSAALHARFR-----EALAAGRLSKVEVIHNCSILAT-VGLRMASTPGVSATL------FDALAKANINVRAIA"),
		[]alphabet.Letter("-----------------INDAVYAEVVGHGEVWSARLMSAVLNQQG-----LPAAWLDAREFLRAERAAQPQVDEGLSYPLLQQLLVQH-P-GKRLVVTGF---ISRNNAGETVLLGRNGSDYSATQ-----------------IGALAGVSRVTIWSDVAGVYSADP--RKVKDACLLPLLRLDEASELARLAAPVLHARTLQPVSGSEIDLQLRCSYTPDQGSTRIERVLASGTGARIVTSHDDVCLI-EFQVPASQDFKLAHKEI--DQILKRAQVRPLAVGVHNDRQLLQFCYTSEVADSALKILDEAG---------LPGELRLRQGLALVAMVGAGVTRNPLHCHRFWQQLKGQPVEFTWQSDDGISLVAVL"),
		[]alphabet.Letter("-----------------ISPREQDLLLSCGETISSVVFTSMLLDNGVKA--AALTGAQAGFLTNDQHTNAKIIEMKPER--LFSVLAN----HDAVVVAGF---QGATEKGDTTTIGRGGSDTSAAA-----------------LGAAVDAEYIDIFTDVEGVMTADP--RVVENAKPLPVVTYTEICNLAYQGAKVISPRAVEIAMQAKVPIRVRSTYS-NDKGTLVTSHHSSKVGSDVFERLITGIAH-VKDVTQFKVPAKIGQYN-----VQTEVFKAMANAGISVDFFNITPSEIVYTVAGNKTETAQR------------ILMDMGYDPMVTRNCAKVSAVGAGIMGVPGVTSKI------VSALSEKEIPILQSA"),
		[]alphabet.Letter("-----------------KRE--MDMLLSTGEQVSIALLAMSLHEKGYKA--VSLTGWQAGITTEEMHGNARIMNIDTT--RIRRCLDE----GAIVIVAGF---QGVTETGEITTLGRGGSDTTAVA-----------------LAAALKAEKCDIYTDVTGVFTTDP--RYVKTARKIKEISYDEMLELANLGAGVLHPRAVEFAKNYEVPLEVRSSME-NERGTMVK--EEVSMEQHLIVRGIAFEDQ-VTRVTVVGIEKYLQSVA--------TIFTALANRGINVDIIIQNA--------------------TNSETAS--VSFSIRTEDLPETLQVLQ-------------ALEGADVHYESGLAKVSI-VGSGMISNPGVAARV------FEVLADQGIEIKMVS"),
		[]alphabet.Letter("-----------------KRE--MDMLLATGEQVTISLLSMALQEKGYDA--VSYTGWQAGIRTEAIHGNARITDIDTS--VLADQLEK----GKIVIVAGF---QGMTEDCEITTLGRGGSDTTAVA-----------------LAAALKVDKCDIYTDVPGVFTTDP--RYVKSARKLEGISYDEMLELANLGAGVLHPRAVEFAKNYQVPLEVRSSTE-TEAGTLIE--EESSMEQNLIVRGIAFEDQ-ITRVTIYGLTSGLTTLS--------TIFTTLAKRNINVDIIIQTQ--------------------AEDKTG---ISFSVKTEDADQTVAVLEEYK---------DALEFEKIETESKLAKVSI-VGSGMVSNPGVAAEM------FAVLAQKNILIKMVS"),
		[]alphabet.Letter("-----------------ARE--MDMLLTAGERISNALVAMAIESLGAEA--QSFTGSQAGVLTTERHGNARIVDVTPG--RVREALDE----GKICIVAGF--QGVNKETRDVTTLGRGGSDTTAVA-----------------LAAALNADVCEIYSDVDGVYTADP--RIVPNAQKLEKLSFEEMLELAAVGSKILVLRSVEYARAFNVPLRVRSSYS-NDPGTLIAGSMEDIPVEEAVLTGVATDKS-EAKVTVLGISDKPGEAA--------KVFRALADAEINIDMVLQNV--------------------SSVEDGTTDITFTCPRADGRRAMEILKKLQ---------VQGNWTNVLYDDQVDKVSL-VGAGMKSHPGVTAEF------MEALRDVNVNIELIS"),
		[]alphabet.Letter("-----------------PRE--MDMLLTAGERISNALVAMAIESLGAQA--RSFTGSQAGVITTGTHGNAKIIDVTPG--RLRDALDE----GQIVLVAGF--QGVSQDSKDVTTLGRGGSDTTAVA-----------------VAAALDADVCEIYTDVDGIFTADP--RIVPNARHLDTVSFEEMLEMAACGAKVLMLRCVEYARRYNVPIHVRSSYS-DKPGTIVKGSIEDIPMEDAILTGVAHDRS-EAKVTVVGLPDVPGYAA--------KVFRAVAEADVNIDMVLQNI--------------------SKIEDGKTDITFTCARDNGPRAVEKLSALK---------SEIGFSQVLYDDHIGKVSL-IGAGMRSHPGVTATF------CEALAEAGINIDLIS"),
		[]alphabet.Letter("-----------------TSPALTDELVSHGELMSTLLFVEILRERD--V--QAQWFDVRKVMRTNDRFGRAEPDIAALAELAALQLLPR-LNEGLVITQGF---IGSENKGRTTTLGRGGSDYTAAL-----------------LAEALHASRVDIWTDVPGIYTTDP--RVVSAAKRIDEIAFAEAAEMATFGAKVLHPATLLPAVRSDIPVFVGSSKDPRAGGTLVCNKTENPPLFRALAL--RRNQT-LLTLHSLNMLHSRGFLA--------EVFGILARHNISVDLITTSEVSVALTLDTTGSTSTG----------DTLLTQSLLMELSALCRVEVEEGLALVALIG----------NDLSKACGVGKEVF"),
		[]alphabet.Letter("-----------------VSSRTVDLVMSCGEKLSCLFMTALCNDRGCKAKYVDLSHIVPSDFSASALDNSFYTFLVQALKEKLAPFVSA-KERIVPVFTGF---FGLVPTGLLNGVGRGYTDLCAAL-----------------IAVAVNADELQVWKEVDGIFTADP--RKVPEARLLDSVTPEEASELTYYGSEVIHPFTMEQVIRAKIPIRIKNVQNPLGNGTIIYPDNVAKKGESTPPHPPENLSS----SFYEKRKRGATAITTKN----DIFVINIHSNKKTLSHGFLAQIFTILDKYKLVVDLISTSEVHVSMALPIPDADS-LKSLRQAEEKLRILGSVDITKKLSIVSLVGKHMKQYIGIAG---TMFTTLAEEGINIEMIS"),
	}
)

func (s *S) TestReadFasta(c *check.C) {
	var (
		obtainN []string
		obtainS [][]alphabet.Letter
	)

	for _, fa := range fas {
		r := NewReader(bytes.NewBufferString(fa), protein.NewSeq("", nil, alphabet.Protein))
		for {
			if s, err := r.Read(); err != nil {
				if err == io.EOF {
					break
				} else {
					c.Fatalf("Failed to read %q: %s", fa, err)
				}
			} else {
				t := s.(*protein.Seq)
				header := *t.Name()
				if desc := *t.Description(); len(desc) > 0 {
					header += " " + desc
				}
				obtainN = append(obtainN, header)
				obtainS = append(obtainS, *(t.Raw().(*[]alphabet.Letter)))
			}
		}
		c.Check(obtainN, check.DeepEquals, expectN)
		obtainN = nil
		for i := range obtainS {
			c.Check(len(obtainS[i]), check.Equals, len(expectS[i]))
			c.Check(obtainS[i], check.DeepEquals, expectS[i])
		}
		obtainS = nil
	}
}

func (s *S) TestReadFastaPacked(c *check.C) {
	var (
		pfan = "test"
		pfas = "cagcagcacgcaggaggctagcagcatcgatgtatagctagactatacgatc"
	)

	t, err := packed.NewSeq("", nil, alphabet.DNA)
	if err != nil {
		c.Fatalf("Failed to create new packed.Seq: %v", err)
	}
	r := NewReader(bytes.NewBufferString(fmt.Sprintf(">%s\n%s\n", pfan, pfas)), t)
	for {
		if s, err := r.Read(); err != nil {
			if err == io.EOF {
				break
			} else {
				c.Fatalf("Failed to read %q: %s", pfan, err)
			}
		} else {
			t := s.(*packed.Seq)
			header := *t.Name()
			if desc := *t.Description(); len(desc) > 0 {
				header += " " + desc
			}
			c.Check(header, check.Equals, pfan)
			c.Check(fmt.Sprintf("%s", s), check.Equals, pfas)
		}
	}
}

func (s *S) TestReadFastaQPacked(c *check.C) {
	var (
		pfan = "test"
		pfas = "cagcagcacgcaggaggctagcagcatcgatgtatagctagactatacgatc"
	)

	t, err := packed.NewQSeq("", nil, alphabet.DNA, alphabet.Sanger)
	if err != nil {
		c.Fatalf("Failed to create new packed.QSeq: %v", err)
	}
	r := NewReader(bytes.NewBufferString(fmt.Sprintf(">%s\n%s\n", pfan, pfas)), t)
	for {
		if s, err := r.Read(); err != nil {
			if err == io.EOF {
				break
			} else {
				c.Fatalf("Failed to read %q: %s", pfan, err)
			}
		} else {
			t := s.(*packed.QSeq)
			header := *t.Name()
			if desc := *t.Description(); len(desc) > 0 {
				header += " " + desc
			}
			c.Check(header, check.Equals, pfan)
			c.Check(fmt.Sprintf("%s", s), check.Equals, pfas)
		}
	}
}

func (s *S) TestWriteFasta(c *check.C) {
	fa := fas[0]
	expectSize := 4649
	var total int
	b := &bytes.Buffer{}
	w := NewWriter(b, 60)

	seq := protein.NewSeq("", nil, alphabet.Protein)

	for i := range expectN {
		seq.ID = expectN[i]
		seq.S = expectS[i]
		if n, err := w.Write(seq); err != nil {
			c.Fatalf("Failed to write to buffer: %s", err)
		} else {
			total += n
		}
	}

	c.Check(total, check.Equals, expectSize)
	c.Check(string(b.Bytes()), check.Equals, fa)
}
