package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/kortschak/BioGo/bio"
	"github.com/kortschak/BioGo/io/alignio"
	"github.com/kortschak/BioGo/io/featio/gff"
	"github.com/kortschak/BioGo/io/seqio/fasta"
	"github.com/kortschak/BioGo/pwm"
	"github.com/kortschak/BioGo/seq"
	"os"
	"strconv"
	"strings"
)

func main() {
	var (
		in, min *fasta.Reader
		mf      *os.File
		matin   *bufio.Reader
		align   seq.Alignment
		out     *gff.Writer
		e       error
	)

	inName := flag.String("in", "", "Filename for input. Defaults to stdin.")
	matName := flag.String("mat", "", "Filename for matrix/alignment input.")
	num := flag.Bool("num", false, "Use numerical description rather than sequence.")
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
	}

	matrix := [][]float64{}

	if *num {
		if mf, e = os.Open(*matName); e != nil {
			fmt.Fprintf(os.Stderr, "Error: %v.\n", e)
			os.Exit(0)
		} else {
			matin = bufio.NewReader(mf)
		}
		defer mf.Close()

		for {
			line, err := matin.ReadBytes('\n')
			if err != nil {
				break
			}
			if line[len(line)-1] == '\n' {
				line = line[:len(line)-1]
			}
			fields := strings.Split(string(line), "\t")
			if len(fields) < 4 {
				break
			}
			matrix = append(matrix, make([]float64, 0, 4))
			for _, s := range fields {
				if f, err := strconv.ParseFloat(s, 64); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v.\n", err)
					os.Exit(0)
				} else {
					matrix[len(matrix)-1] = append(matrix[len(matrix)-1], f)
				}
			}
		}
	} else {
		if min, e = fasta.NewReaderName(*matName); e != nil {
			fmt.Fprintf(os.Stderr, "Error: %v.\n", e)
			os.Exit(0)
		} else {
			if align, e = alignio.NewReader(min).Read(); e != nil {
				fmt.Fprintf(os.Stderr, "Error: %v.\n", e)
				os.Exit(0)
			}
		}
		defer min.Close()

		for i := 0; i < align.Len(); i++ {
			matrix[i] = make([]float64, 4)
			if col, err := align.Column(i, 0); err != nil {
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
	}

	wm := pwm.New(matrix)

	if *outName == "" {
		out = gff.NewWriter(os.Stdout, 2, 60, true)
	} else if out, e = gff.NewWriterName(*outName, 2, 60, true); e != nil {
		fmt.Fprintf(os.Stderr, "Error: %v.\n", e)
	}
	defer out.Close()

	source := "pwmscan"
	feature := "match"

	for {
		if sequence, err := in.Read(); err != nil {
			break
		} else {
			fmt.Fprintf(os.Stderr, "Working on: %s\n", sequence.ID)
			results := wm.Search(sequence, 0, sequence.Len(), *minScore)

			for _, r := range results {
				r.Source = source
				r.Feature = feature
				out.Write(r)
			}
		}
	}
}
