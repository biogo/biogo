package main

import (
	"flag"
	"os"
	"io"
	"log"
	"time"
	"fmt"
	"runtime"
	"runtime/pprof"
	"strings"
	"strconv"
	"bio/seq"
	"bio/index/kmerindex"
	"bio/morass"
	"bio/align/pals/filter"
	"bio/align/pals/dp"
	"bio/align/pals"
)

const (
	timeFormat = "20060102150405-Mon"
)

var (
	pid     = os.Getpid()
	timer   = NewTimer()
	debug   bool
	verbose bool
)

func initLog(fileName string) {
	log.SetPrefix(fmt.Sprintf("Krishna %d: ", pid))
	if file, err := os.Create(fileName); err == nil {
		fmt.Fprintln(file, strings.Join(os.Args, " "))
		log.SetOutput(io.MultiWriter(os.Stderr, file))
	} else {
		log.Fatalf("Error: Could not open log file: %v.", err)
	}
}

func main() {
	var (
		err          error
		filterParams *filter.Params
		//		dpParams     *pals.Params
		index          *kmerindex.Index
		filterMorass   *morass.Morass
		hitFilter      *filter.Filter
		merger         *filter.Merger
		hit            filter.FilterHit
		trapezoids     []*filter.Trapezoid
		aligner        *dp.Aligner
		hitCoverage, n int
		profile        *os.File
	)

	queryName := flag.String("query", "", "Filename for query sequence.")
	targetName := flag.String("target", "", "Filename for target sequence.")
	selfCompare := flag.Bool("self", false, "Is this a self comparison?")
	sameStrand := flag.Bool("same", false, "Only compare same strand")

	outFile := flag.String("out", "", "File to send output to.")

	maxK := flag.Int("k", -1, "Maximum kmer length (negative indicates automatic detection based on architecture).")
	minHitLen := flag.Int("length", 400, "Minimum hit length.")
	minId := flag.Float64("identity", 0.94, "Minimum hit identity.")
	tubeOffset := flag.Int("tubeoffset", 32, "Tube offset.")

	tmpDir := flag.String("tmp", "", "Path for temporary files.")
	tmpChunk := flag.Int("chunk", 1<<20, "Chunk size for morass.")
	tmpConcurrent := flag.Bool("tmpcon", false, "Process morass concurrently.")

	threads := flag.Int("threads", 1, "Number of threads to use for alignment.")
	maxMem := flag.Uint64("mem", 1<<32, "Maximum nominal memory.")

	logToFile := flag.Bool("log", false, "Log to file.")

	flag.BoolVar(&debug, "debug", false, "Records line of fatal error in log.")
	flag.BoolVar(&verbose, "verbose", false, "Record more information.")

	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to this file.")

	help := flag.Bool("help", false, "Print this help message.")

	flag.Parse()

	runtime.GOMAXPROCS(*threads)

	if *cpuprofile != "" {
		if profile, err = os.Create(*cpuprofile); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v.", err)
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Writing CPU profile data to %s\n", *cpuprofile)
		pprof.StartCPUProfile(profile)
		defer pprof.StopCPUProfile()
	}

	if *help {
		flag.Usage()
		os.Exit(1)
	}

	if *logToFile {
		initLog("krishna-" + time.LocalTime().Format(timeFormat) + "-" + strconv.Itoa(pid) + ".log")
	}

	log.Println(os.Args)
	var target, query *seq.Seq
	if *targetName != "" {
		target = packSequence(*targetName)
	} else {
		if debug {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
		}
		log.Fatalln("No target provided.")
	}

	var writer *pals.Writer
	if *outFile == "" {
		writer = pals.NewWriter(os.Stdout, 2, 60, false)
	} else {
		if writer, err = pals.NewWriterName(*outFile, 2, 60, false); err != nil {
			if debug {
				log.SetFlags(log.LstdFlags | log.Lshortfile)
			}
			log.Fatalf("Could not open output file: %v", err)
		}
	}
	defer writer.Close()

	if !*selfCompare {
		if *queryName != "" {
			query = packSequence(*queryName)
		} else {
			if debug {
				log.SetFlags(log.LstdFlags | log.Lshortfile)
			}
			log.Fatalln("No query provided in non-self comparison.")
		}
	} else {
		query = target
	}

	if *maxK > 0 {
		pals.MaxKmerLen = *maxK
	}
	pals.Debug = debug
	if filterParams, /*dpParams*/ _, err = pals.OptimiseParameters(*minHitLen, *minId, target, query, *tubeOffset, *maxMem); err != nil {
		log.Fatalf("Error: %v.", err)
	}
	log.Printf("Using filter parameters:")
	log.Printf("\tWordSize = %d", filterParams.WordSize)
	log.Printf("\tMinMatch = %d", filterParams.MinMatch)
	log.Printf("\tMaxError = %d", filterParams.MaxError)
	log.Printf("\tTubeOffset = %d", filterParams.TubeOffset)
	log.Printf("Building index for %s", target.ID)
	timer.Interval()

	if index, err = kmerindex.New(filterParams.WordSize, target); err == nil {
		index.Build()
		log.Printf("Indexed in %v ms", timer.Interval()/1e6)
		hitFilter = filter.New(index, *filterParams)
	} else {
		if debug {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
		}
		log.Fatalf("Error: %v.", err)
	}

	both := !*sameStrand
	for _, comp := range [...]bool{false, true} {
		if comp {
			log.Println("Working on complementary strands")
		} else {
			log.Println("Working on self strand")
		}
		if both || !comp {
			var workingQuery *seq.Seq
			if comp {
				workingQuery, _ = query.RevComp()
			} else {
				workingQuery = query
			}

			if filterMorass, err = morass.New("krishna_"+strconv.Itoa(pid), *tmpDir, *tmpChunk, *tmpConcurrent); err == nil {
				log.Println("Filtering")
				timer.Interval()

				lifeline := make(chan struct{})
				if *threads > 1 {
					go func() {
						for {
							fmt.Printf("Recording filter hits - %d recorded\r", filterMorass.Pos())
							time.Sleep(5e8)

							select {
							case <-lifeline:
								return
							default:
							}
						}
					}()
				}

				if err = hitFilter.Filter(workingQuery, *selfCompare, comp, filterMorass); err != nil {
					if debug {
						log.SetFlags(log.LstdFlags | log.Lshortfile)
					}
					log.Fatalf("Error: Problem finalising morass: %v.", err)
				}
				close(lifeline)

				log.Printf("Identified %d filter hits in %v ms", filterMorass.Len(), timer.Interval()/1e6)
			} else {
				if debug {
					log.SetFlags(log.LstdFlags | log.Lshortfile)
				}
				log.Fatalf("Error: Could not create morass: %v.", err)
			}

			log.Println("Merging")
			timer.Interval()
			merger = filter.NewMerger(index, workingQuery, filterParams, *selfCompare)
			for {
				if err = filterMorass.Pull(&hit); err != nil {
					break
				}
				if filterMorass.Pos()%100000 == 0 {
					fmt.Fprintf(os.Stderr, "  Merging filter hit %d of %d\r", filterMorass.Pos(), filterMorass.Len())
				}
				if filterMorass.Pos()%10000 == 0 {
					switch filterMorass.Pos() / 10000 % 4 {
					case 0:
						fmt.Fprint(os.Stderr, "|\r")
					case 1:
						fmt.Fprint(os.Stderr, "/\r")
					case 2:
						fmt.Fprint(os.Stderr, "-\r")
					case 3:
						fmt.Fprint(os.Stderr, "\\\r")
					}
				}
				merger.MergeFilterHit(&hit)
			}
			if err != nil && err != io.EOF {
				if debug {
					log.SetFlags(log.LstdFlags | log.Lshortfile)
				}
				log.Fatalf("Error: Problem merging filter hits: %v.", err)
			}
			filterMorass.CleanUp()

			trapezoids = merger.FinaliseMerge()
			log.Printf("Merged %d trapezoids covering %d in %v ms", len(trapezoids), filter.SumTrapLengths(trapezoids), timer.Interval()/1e6)

			log.Println("Aligning")
			timer.Interval()

			aligner = dp.NewAligner(target, workingQuery, filterParams.WordSize, *minHitLen, *minId, *threads)
			hits := aligner.AlignTraps(trapezoids)
			if hitCoverage, err = dp.SumDPLengths(hits); err != nil {
				if debug {
					log.SetFlags(log.LstdFlags | log.Lshortfile)
				}
				log.Fatalf("Error: %v.", err)
			}
			log.Printf("Aligned %d hits covering %d in %v ms", len(hits), hitCoverage, timer.Interval()/1e6)

			log.Println("Writing results")
			timer.Interval()
			if n, err = WriteDPHits(writer, target, query, hits, comp); err != nil {
				if debug {
					log.SetFlags(log.LstdFlags | log.Lshortfile)
				}
				log.Fatalf("Error: %v.", err)
			}
			log.Printf("Wrote hits (%v bytes) in %v ms", n, timer.Interval()/1e6)
		}
	}
}
