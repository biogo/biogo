package main

import (
	"flag"
	"fmt"
	"github.com/kortschak/biogo/io/seqio/fasta"
	"github.com/kortschak/biogo/seq"
	"os"
	"runtime/pprof"
)

func main() {
	var (
		in      *fasta.Reader
		out     *fasta.Writer
		e       error
		profile *os.File
	)

	inName := flag.String("in", "", "Filename for input. Defaults to stdin.")
	outName := flag.String("out", "", "Filename for output. Defaults to stdout.")
	size := flag.Int("size", 40, "Fragment size.")
	width := flag.Int("width", 60, "Fasta output width.")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to this file.")
	help := flag.Bool("help", false, "Print this usage message.")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(1)
	}

	if *cpuprofile != "" {
		if profile, e = os.Create(*cpuprofile); e != nil {
			fmt.Fprintf(os.Stderr, "Error: %v.", e)
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Writing CPU profile data to %s\n", *cpuprofile)
		pprof.StartCPUProfile(profile)
		defer pprof.StopCPUProfile()
	}

	if *inName == "" {
		in = fasta.NewReader(os.Stdin)
	} else if in, e = fasta.NewReaderName(*inName); e != nil {
		fmt.Fprintf(os.Stderr, "Error: %v.", e)
	}
	defer in.Close()

	if *outName == "" {
		out = fasta.NewWriter(os.Stdout, *width)
	} else if out, e = fasta.NewWriterName(*outName, *width); e != nil {
		fmt.Fprintf(os.Stderr, "Error: %v.", e)
	}
	defer out.Close()

	var (
		sequence *seq.Seq
		err      error
	)

	t := &seq.Seq{}

	for {
		if sequence, err = in.Read(); err != nil {
			break
		}
		length := sequence.Len()
		t.ID = sequence.ID
		switch {
		case length >= 20 && length <= 85:
			t.Seq = sequence.Seq[5:]
			out.Write(t)
		case length > 85:
			for start := 0; start+*size <= length; start += *size {
				t.Seq = sequence.Seq[start : start+*size]
				out.Write(t)
			}
		}
	}
}
