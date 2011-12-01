package main

import (
	"flag"
	"fmt"
	"github.com/kortschak/BioGo/graphics/color"
	"github.com/kortschak/BioGo/graphics/kmercolor"
	"github.com/kortschak/BioGo/index/kmerindex"
	"github.com/kortschak/BioGo/io/seqio/fasta"
	"image"
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
	chunk := flag.Int("chunk", 1000, "Chunk width.")
	height := flag.Int("h", 100, "Rainbow height.")
	help := flag.Bool("help", false, "Print this usage message.")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(1)
	}

	if *inName == "" {
		in = fasta.NewReader(os.Stdin)
	} else if in, e = fasta.NewReaderName(*inName); e != nil {
		fmt.Fprintf(os.Stderr, "Error: %v.", e)
		os.Exit(0)
	}
	defer in.Close()

	count := 0
	for {
		count++
		if sequence, err := in.Read(); err != nil {
			os.Exit(0)
		} else {
			if index, err := kmerindex.New(*k, sequence); err != nil {
				fmt.Println(err)
				os.Exit(0)
			} else {
				base := &color.HSVAColor{0, 1, 0, 1}
				rainbow := kmercolor.NewKmerRainbow(image.Rect(0, 0, sequence.Len() / *chunk, *height), index, base)
				for i := 0; (i+1)**chunk < sequence.Len(); i++ {
					rainbow.Paint(kmercolor.V, i, *chunk, i, i+1)
				}
				if out, e = os.Create(fmt.Sprintf("%s-%d.png", *outName, count)); e != nil {
					fmt.Fprintf(os.Stderr, "Error: %v.", e)
				}
				png.Encode(out, rainbow)
				out.Close()
			}
		}
	}
}
