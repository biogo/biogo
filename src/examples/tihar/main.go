package main

import (
	"flag"
	"fmt"
	"github.com/kortschak/BioGo/bio/index/kmerindex"
	"github.com/kortschak/BioGo/bio/io/seqio/fasta"
	"github.com/kortschak/BioGo/bio/matrix"
	"github.com/kortschak/BioGo/bio/matrix/sparse"
	"github.com/kortschak/BioGo/bio/nmf"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"time"
)

func main() {
	var (
		in                *fasta.Reader
		out, csv, profile *os.File
		e                 error
	)

	inName := flag.String("in", "", "Filename for input to be factorised. Defaults to stdin.")
	outName := flag.String("out", "", "Filename for output. Defaults to stdout.")
	csvName := flag.String("csv", "", "Filename for csv output of feature details. Defaults to stdout.")
	k := flag.Int("k", 8, "kmer size to use.")
	cat := flag.Int("cat", 5, "number of categories.")
	iter := flag.Int("i", 1000, "iterations.")
	limit := flag.Int("time", 10, "time limit for NMF.")
	lo := flag.Int("lo", 1, "minimum number of kmer frequency to use in NMF.")
	hi := flag.Float64("hi", 0.9, "maximum proportion of kmer representation to use in NMF.")
	sf := flag.Float64("sf", 0.01, "factor for sparcity of estimating matrices for NMF.")
	tol := flag.Float64("tol", 0.001, "tolerance for NMF.")
	threads := flag.Int("threads", 2, "number of threads to use.")
	seed := flag.Int64("seed", -1, "seed for random number generator (-1 uses system clock).")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to this file.")
	help := flag.Bool("help", false, "print this usage message.")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(1)
	}

	runtime.GOMAXPROCS(*threads)
	sparse.MaxProcs = *threads
	fmt.Fprintf(os.Stderr, "Using %d threads.\n", runtime.GOMAXPROCS(0))
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
		fmt.Fprintln(os.Stderr, "Reading sequences from stdin.")
		in = fasta.NewReader(os.Stdin)
	} else if in, e = fasta.NewReaderName(*inName); e != nil {
		fmt.Fprintf(os.Stderr, "Error: %v.", e)
		os.Exit(0)
	} else {
		fmt.Fprintf(os.Stderr, "Reading sequence from `%s'.\n", *inName)
	}
	defer in.Close()

	if *outName == "" {
		fmt.Fprintln(os.Stderr, "Writing output to stdout.")
		out = os.Stdout
	} else if out, e = os.Create(*outName); e != nil {
		fmt.Fprintf(os.Stderr, "Error: %v.", e)
	} else {
		fmt.Fprintf(os.Stderr, "Writing output to `%s'.\n", *outName)
	}
	defer out.Close()

	if *csvName == "" {
		fmt.Fprintln(os.Stderr, "Writing csv output to stdout.")
		csv = os.Stdout
	} else if csv, e = os.Create(*csvName); e != nil {
		fmt.Fprintf(os.Stderr, "Error: %v.", e)
	} else {
		fmt.Fprintf(os.Stderr, "Writing output to `%s'.\n", *csvName)
	}
	defer csv.Close()

	kmers := make(map[kmerindex.Kmer]int)
	positions := make(map[int]int)
	motifs := make(map[kmerindex.Kmer]map[int]map[string]bool)
	maxPos := 0

	for {
		if sequence, err := in.Read(); err != nil {
			break
		} else {
			if kindex, e := kmerindex.New(*k, sequence); e != nil {
				fmt.Fprintf(os.Stderr, "Error: %v.", e)
				os.Exit(0)
			} else {
				kindex.Build()
				index, _ := kindex.KmerIndex()
				for kmer, posList := range index {
					if _, ok := motifs[kmer]; !ok {
						motifs[kmer] = make(map[int]map[string]bool)
					}
					for _, pos := range posList {
						if _, ok := motifs[kmer][pos]; !ok {
							motifs[kmer][pos] = make(map[string]bool)
						}
						motifs[kmer][pos][string(sequence.ID)] = true
						kmers[kmer]++
						positions[pos]++
						if pos > maxPos {
							maxPos = pos
						}
					}
				}
			}
		}
	}

	kmerArray := make([][]float64, 0)
	kmerTable := make([]kmerindex.Kmer, 0)
	positionsTable := make(map[int]int)
	currPos := 0

	for kmer, count := range kmers {
		if count < *lo || float64(count)/float64(maxPos) > *hi {
			continue
		}
		row := make([]float64, currPos)
		for pos, seqs := range motifs[kmer] {
			if len(seqs) < *lo {
				continue
			}
			if i, ok := positionsTable[pos]; ok {
				row[i] += float64(len(motifs[kmer][pos]))
			} else {
				positionsTable[pos] = len(row)
				row = append(row, float64(len(motifs[kmer][pos])))
				currPos++
			}
		}
		kmerArray = append(kmerArray, row)
		kmerTable = append(kmerTable, kmerindex.Kmer(kmer))
	}

	var kmerMatrix *sparse.Sparse
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "Error: %v.\n", r)
				os.Exit(0)
			}
		}()
		kmerMatrix = sparse.Matrix(kmerArray)
	}()

	f := func(i, j int, v float64) float64 {
		if kmerMatrix.At(i, j) != 0 {
			return 1
		}
		return 0
	}
	nonZero := kmerMatrix.Apply(f).Sum()

	r, c := kmerMatrix.Dims()
	density := nonZero / float64(r*c)

	if *seed == -1 {
		*seed = time.Nanoseconds()
	}
	fmt.Fprintf(os.Stderr, "Using %v as random seed.\n", *seed)
	rand.Seed(*seed)

	rows, cols := kmerMatrix.Dims()
	Wo := sparse.Random(rows, *cat, density**sf)
	Ho := sparse.Random(*cat, cols, density**sf)

	fmt.Fprintf(os.Stderr, "Dimensions of Kmer matrix = (%v, %v)\nDensity = %.3f %%\n%v\n", r, c, (density)*100, kmerMatrix)

	W, H, ok := nmf.Factors(kmerMatrix, Wo, Ho, *tol, *iter, int64(*limit)*1e9)

	fmt.Fprintf(os.Stderr, "norm(H) = %v norm(W) = %v\n\nFinished = %v\n\n", H.Norm(matrix.Fro), W.Norm(matrix.Fro), ok)

	printFeature(out, csv, kmerMatrix, W, H, motifs, kmerTable, positionsTable, maxPos, *k)
}

type Weight struct {
	weight float64
	index  int
}

type WeightList []Weight

func (self WeightList) Len() int {
	return len(self)
}

func (self *WeightList) Swap(i, j int) {
	(*self)[i], (*self)[j] = (*self)[j], (*self)[i]
}

func (self WeightList) Less(i, j int) bool {
	return self[i].weight > self[j].weight
}

func printFeature(out, csv *os.File, V, W, H *sparse.Sparse, motifs map[kmerindex.Kmer]map[int]map[string]bool, kmerTable []kmerindex.Kmer, positionsTable map[int]int, maxPos, k int) {
	patternCount, posCount := H.Dims()
	kmerCount, _ := W.Dims()

	hipats := make([]WeightList, posCount)
	pats := make([]string, 0)

	for i := 0; i < patternCount; i++ {
		fmt.Fprintf(out, "Feature %v:\n", i)
		klist := make(WeightList, 0)
		for j := 0; j < kmerCount; j++ {
			klist = append(klist, Weight{weight: W.At(j, i), index: j})
		}
		sort.Sort(&klist)
		name := fmt.Sprint("[")
		for j := 0; j < len(klist); j++ {
			if klist[j].weight > 0 {
				name += fmt.Sprintf(" %s/%.3e ", kmerindex.StringOfLen(k, kmerTable[klist[j].index]), klist[j].weight)
			}
		}
		name += fmt.Sprint("]")
		pats = append(pats, name)
		fmt.Fprintln(out, name)

		plist := make(WeightList, 0)
		for j := 0; j < posCount; j++ {
			plist = append(plist, Weight{weight: H.At(i, j), index: positionsTable[j]})
			hipats[j] = append(hipats[j], Weight{weight: H.At(i, j), index: i})
		}

		sort.Sort(&plist)
		instances := ""
		for j := 0; j < len(plist); j++ {
			if plist[j].weight > 0 {
				instances += fmt.Sprintf("%d/%.3e\n", plist[j].index, plist[j].weight)
			}
		}
		fmt.Fprintln(out, instances)

		fmt.Fprintln(out)
	}

	fmt.Fprint(csv, "position\tfeature\tweight")
	for j := 0; j < patternCount; j++ {
		fmt.Fprintf(csv, "\t%d", j)
	}
	fmt.Fprintln(csv)
	for i := 0; i <= maxPos; i++ {
		if pos, ok := positionsTable[i]; ok {
			all := ""
			for _, pat := range hipats[pos] {
				all += fmt.Sprintf("\t%e", pat.weight)
			}
			sort.Sort(&hipats[pos])
			if hipats[pos][0].weight > 0 {
				fmt.Fprintf(csv, "%d\t%d\t%e%s\n", i, hipats[pos][0].index, hipats[pos][0].weight, all)
			}
		}
	}
}
