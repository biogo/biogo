package main

import (
	"flag"
	"fmt"
	"github.com/kortschak/BioGo/bio"
	"github.com/kortschak/BioGo/bio/alignment"
	"github.com/kortschak/BioGo/bio/io/alignio"
	"github.com/kortschak/BioGo/bio/io/featio/gff"
	"github.com/kortschak/BioGo/bio/io/seqio/fasta"
	"github.com/kortschak/BioGo/bio/pwm"
	"os"
)

func main() {
	var (
		in, min *fasta.Reader
		align   *alignment.Alignment
		out     *gff.Writer
		e       error
	)

	inName := flag.String("in", "", "Filename for input. Defaults to stdin.")
	matName := flag.String("mat", "", "Filename for matrix input.")
	outName := flag.String("out", "", "Filename for output. Defaults to stdout.")
	precision := flag.Int("prec", 6, "Precision for floating point output.")
	minScore := flag.Float64("score", 0.9, "Minimum score for a hit.")
	help := flag.Bool("help", false, "Print this usage message.")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(1)
	}

	bio.Precision = *precision

	if *inName == "" {
		in = fasta.NewReader(os.Stdin)
	} else if in, e = fasta.NewReaderName(*inName); e != nil {
		fmt.Fprintf(os.Stderr, "Error: %v.\n", e)
		os.Exit(0)
	}
	defer in.Close()

	if *matName == "" {
		flag.Usage()
		os.Exit(0)
	} else if min, e = fasta.NewReaderName(*matName); e != nil {
		fmt.Fprintf(os.Stderr, "Error: %v.\n", e)
		os.Exit(0)
	} else {
		if align, e = alignio.NewReader(min).Read(); e != nil {
			fmt.Fprintf(os.Stderr, "Error: %v.\n", e)
			os.Exit(0)
		}
	}
	defer min.Close()

	if *outName == "" {
		out = gff.NewWriter(os.Stdout, 2, 60, true)
	} else if out, e = gff.NewWriterName(*outName, 2, 60, true); e != nil {
		fmt.Fprintf(os.Stderr, "Error: %v.\n", e)
	}
	defer out.Close()

	matrix := make([][]float64, align.Len())
	for i := 0; i < align.Len(); i++ {
		matrix[i] = make([]float64, 4)
		if col, err := align.Column(i); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v.\n", e)
			os.Exit(0)
		} else {
			for _, v := range col {
				if base := pwm.LookUp.ValueToCode[v]; base >= 0 {
					matrix[i][base]++
				}
			}
		}
	}

	wm := pwm.New(matrix)

	source := []byte("pwmscan")
	feature := []byte("match")

	for {
		if sequence, err := in.Read(); err != nil {
			break
		} else {
			results := wm.Search(sequence, 0, sequence.Len(), *minScore)

			for _, r := range results {
				r.Source = source
				r.Feature = feature
				out.Write(r)
			}
		}
	}
}
