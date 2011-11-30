package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/kortschak/BioGo/bio/matrix"
	"github.com/kortschak/BioGo/bio/matrix/sparse"
	"github.com/kortschak/BioGo/bio/nmf"
	"io"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"time"
)

func main() {
	var (
		in      *bufio.Reader
		out     *bufio.Writer
		profile *os.File
		e       error
	)

	inName := flag.String("in", "", "Filename for input to be factorised. Defaults to stdin.")
	outName := flag.String("out", "", "Filename for output. Defaults to stdout.")
	transpose := flag.Bool("t", false, "Transpose columns and rows.")
	sep := flag.String("sep", "\t", "Column delimiter.")
	cat := flag.Int("cat", 5, "number of categories.")
	iter := flag.Int("i", 1000, "iterations.")
	rep := flag.Int("rep", 1, "Resample replicates.")
	limit := flag.Int("time", 10, "time limit for NMF.")
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
		fmt.Fprintln(os.Stderr, "Reading table from stdin.")
		in = bufio.NewReader(os.Stdin)
	} else if f, err := os.Open(*inName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v.", err)
		os.Exit(0)
	} else {
		defer f.Close()
		in = bufio.NewReader(f)
		fmt.Fprintf(os.Stderr, "Reading table from `%s'.\n", *inName)
	}

	if *outName == "" {
		fmt.Fprintln(os.Stderr, "Writing output to stdout.")
		out = bufio.NewWriter(os.Stdout)
	} else if f, err := os.Create(*outName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v.", err)
		os.Exit(0)
	} else {
		defer f.Close()
		out = bufio.NewWriter(f)
		fmt.Fprintf(os.Stderr, "Writing output to `%s'.\n", *outName)
	}
	defer out.Flush()

	var colNames, rowNames []string
	array := make([][]float64, 0)

	if line, err := in.ReadString('\n'); err != nil {
		fmt.Fprintln(os.Stderr, "No table to read\n")
		os.Exit(0)
	} else {
		line = strings.TrimSpace(line)
		colNames = strings.Split(line, "\t")
		colNames = colNames[1:]
	}

	for count := 1; ; count++ {
		if line, err := in.ReadString('\n'); err != nil {
			break
		} else {
			line = strings.TrimSpace(line)
			if row := strings.Split(line, *sep); len(row) != len(colNames)+1 {
				fmt.Fprintf(os.Stderr, "Table row mismatch at line %d.\n", count)
				os.Exit(0)
			} else {
				rowData := make([]float64, len(row)-1)
				for i, val := range row[1:] {
					if rowData[i], e = strconv.Atof64(val); e != nil {
						fmt.Fprintf(os.Stderr, "Float conversion error %v at line %d element %d.\n", e, count, i)
						os.Exit(0)
					}
				}
				rowNames = append(rowNames, row[0])
				array = append(array, rowData)
			}
		}
	}

	var dataMatrix *sparse.Sparse
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "Error: %v.", r)
				os.Exit(0)
			}
		}()
		dataMatrix = sparse.Matrix(array)
	}()

	f := func(i, j int, v float64) float64 {
		if dataMatrix.At(i, j) != 0 {
			return 1
		}
		return 0
	}
	nonZero := dataMatrix.Apply(f).Sum()

	if *transpose {
		colNames, rowNames = rowNames, colNames
		dataMatrix = dataMatrix.T()
	}
	r, c := dataMatrix.Dims()

	density := nonZero / float64(r*c)

	if *seed == -1 {
		*seed = time.Nanoseconds()
	}
	fmt.Fprintf(os.Stderr, "Using %v as random seed.\n", *seed)
	rand.Seed(*seed)

	rows, cols := dataMatrix.Dims()

	fmt.Fprintf(os.Stderr, "Dimensions of matrix = (%v, %v)\nDensity = %.3f %%\n%v\n", r, c, (density)*100, dataMatrix)

	for run := 0; run < *rep; run++ {
		if *rep > 1 {
			fmt.Fprintf(os.Stderr, "Replicate #%d\n", run+1)
		}
		Wo := sparse.Random(rows, *cat, density**sf)
		Ho := sparse.Random(*cat, cols, density**sf)

		W, H, ok := nmf.Factors(dataMatrix, Wo, Ho, *tol, *iter, int64(*limit)*1e9)

		fmt.Fprintf(os.Stderr, "norm(H) = %v norm(W) = %v\n\nFinished = %v\n\n", H.Norm(matrix.Fro), W.Norm(matrix.Fro), ok)

		printFeature(out, run, dataMatrix, W, H, rowNames, colNames)
	}
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

func printFeature(out io.Writer, run int, V, W, H *sparse.Sparse, rowNames, colNames []string) {
	patternCount, colCount := H.Dims()
	rowCount, _ := W.Dims()

	hipats := make([]WeightList, colCount)
	pats := make([]string, 0)

	for i := 0; i < patternCount; i++ {
		rlist := make(WeightList, 0)
		for j := 0; j < rowCount; j++ {
			rlist = append(rlist, Weight{weight: W.At(j, i), index: j})
		}
		sort.Sort(&rlist)
		name := []string{}
		for j := 0; j < len(rlist); j++ {
			if rlist[j].weight > 0 {
				name = append(name, fmt.Sprintf("%s/%.3e", rowNames[rlist[j].index], rlist[j].weight))
			}
		}
		nameString := strings.Join(name, ",")
		pats = append(pats, nameString)

		clist := make(WeightList, 0)
		for j := 0; j < colCount; j++ {
			clist = append(clist, Weight{weight: H.At(i, j), index: j})
			hipats[j] = append(hipats[j], Weight{weight: H.At(i, j), index: i})
		}

		sort.Sort(&clist)
		instances := []string{}
		for j := 0; j < len(clist); j++ {
			if clist[j].weight > 0 {
				instances = append(instances, fmt.Sprintf("%s/%.3e", colNames[clist[j].index], clist[j].weight))
			}
		}
		instanceString := strings.Join(instances, ",")
		fmt.Fprintf(out, "%d\t[%s]\t(%s)\n", run, nameString, instanceString)
	}

	for j, col := range hipats {
		sort.Sort(&col)
		fmt.Fprintf(os.Stderr, "%s/%e: %d\n", colNames[j], col[0].weight, col[0].index)
	}
}
