package main

import (
	"flag"
	"os"
	"fmt"
	"bio/seq"
	"bio/index/kmerindex"
	"bio/io/seqio/fasta"
)

func main() {
	var in1, in2 *fasta.Reader

	inName1 := flag.String("1", "", "Filename for first input.")
	inName2 := flag.String("2", "", "Filename for second input.")
	k := flag.Int("k", 6, "kmer size.")
	help := flag.Bool("help", false, "Print this usage message.")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(1)
	}

	var err os.Error

	if in1, err = fasta.NewReaderName(*inName1); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v.", err)
		os.Exit(0)
	}
	defer in1.Close()

	if in2, err = fasta.NewReaderName(*inName2); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v.", err)
		os.Exit(0)
	}
	defer in2.Close()

	var (
		seq1, seq2             *seq.Seq
		kmerFreqs1, kmerFreqs2 map[kmerindex.Kmer]float64
		ok                     bool
	)

	if seq1, err = in1.Read(); err != nil {
		os.Exit(0)
	}
	if seq2, err = in2.Read(); err != nil {
		os.Exit(0)
	}

	if index, err := kmerindex.New(*k, seq1); err != nil {
		fmt.Println(err)
		os.Exit(0)
	} else {
		if kmerFreqs1, ok = index.NormalisedKmerFrequencies(); !ok {
			fmt.Printf("Unable to determine Kmer frequences for %s\n", seq1.ID)
			os.Exit(0)
		}
	}
	if index, err := kmerindex.New(*k, seq2); err != nil {
		fmt.Println(err)
		os.Exit(0)
	} else {
		if kmerFreqs2, ok = index.NormalisedKmerFrequencies(); !ok {
			fmt.Printf("Unable to determine Kmer frequences for %s\n", seq2.ID)
			os.Exit(0)
		}
	}

	fmt.Printf("Kmer distance between %s and %s is %f\n", seq1.ID, seq2.ID, kmerindex.Distance(kmerFreqs1, kmerFreqs2))
}
