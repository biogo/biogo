package main

import (
	"flag"
	"fmt"
	"github.com/kortschak/biogo/graphics/color"
	"github.com/kortschak/biogo/graphics/kmercolor"
	"github.com/kortschak/biogo/index/kmerindex"
	"github.com/kortschak/biogo/io/seqio/fasta"
	"image/png"
	"os"
)

func main() {
	var (
		in  *fasta.Reader
		out *os.File
		e   error
	)

	inName := flag.String("in", "", "Filename for input. Defaults to stdin.")
	outName := flag.String("out", "", "Filename for output. Defaults to stdout.")
	k := flag.Int("k", 6, "kmer size.")
	start := flag.Int("s", 0, "Start site - mandatory parameter > 0.")
	chunk := flag.Int("chunk", 1000, "Chunk width - < 0 indicates sequence to end.")
	desch := flag.Bool("desch", false, "Use diagonal base arrangement described by Deschavanne et al., otherwise use orthogonal arrangement.")
	help := flag.Bool("help", false, "Print this usage message.")

	flag.Parse()

	kmerindex.MinKmerLen = *k

	if *help {
		flag.Usage()
		os.Exit(1)
	}

	if *start == 0 {
		fmt.Fprintln(os.Stderr, "Must specify s > 0")
		flag.Usage()
		os.Exit(0)
	}

	if *inName == "" {
		in = fasta.NewReader(os.Stdin)
	} else if in, e = fasta.NewReaderName(*inName); e != nil {
		fmt.Fprintf(os.Stderr, "Error: %v.", e)
		os.Exit(0)
	}
	defer in.Close()

	if sequence, err := in.Read(); err != nil {
		os.Exit(0)
	} else {
		if *chunk < 0 {
			*chunk = sequence.Len() - *start - 1
		}
		fmt.Fprintf(os.Stderr, "Indexing %s\n", sequence.ID)
		if index, err := kmerindex.New(*k, sequence); err != nil {
			fmt.Println(err)
			os.Exit(0)
		} else {
			base := color.HSVA{0, 1, 1, 1}
			cgr := kmercolor.NewCGR(index, base)
			fmt.Fprintf(os.Stderr, "Painting %s\n", sequence.ID)
			cgr.Paint(kmercolor.V|kmercolor.H, *desch, *start, *chunk)
			fmt.Fprintf(os.Stderr, "Writing %s\n", sequence.ID)
			if out, e = os.Create(fmt.Sprintf("%s.png", *outName)); e != nil {
				fmt.Fprintf(os.Stderr, "Error: %v.", e)
			}
			png.Encode(out, cgr)
			out.Close()
		}
	}
}
