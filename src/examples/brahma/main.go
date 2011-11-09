package main

import (
	"bio/feat"
	"bio/interval"
	"bio/io/featio/gff"
	"bio/util"
	"container/heap"
	"encoding/gob"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	maxAnnotations   = 8
	annotationLength = 256
	mapLength        = 20
	maxMargin        = 0.05
)

var (
	minOverlap float64
)

type RepeatRecord struct {
	Name, Class      string
	Start, End, Left int
}

func (self *RepeatRecord) Parse(annot []byte) {
	fields := strings.Split(string(annot), " ")

	self.Name = fields[1]
	self.Class = fields[2]
	if fields[3] != "." {
		self.Start, _ = strconv.Atoi(fields[3])
	} else {
		self.Start = -1
	}
	if fields[4] != "." {
		self.End, _ = strconv.Atoi(fields[4])
	} else {
		self.End = -1
	}
	if fields[5] != "." {
		self.Left, _ = strconv.Atoi(fields[5])
	} else {
		self.Left = -1
	}
}

type Match struct {
	Interval *interval.Interval
	Overlap  int
	Strand   int8
}

type Matches []Match

func (self Matches) Len() int {
	return len(self)
}

func (self *Matches) Swap(i, j int) {
	(*self)[i], (*self)[j] = (*self)[j], (*self)[i]
}

type Overlap struct {
	*Matches
}

func (self Overlap) Less(i, j int) bool {
	return (*self.Matches)[i].Overlap < (*self.Matches)[j].Overlap
}

func (self *Overlap) Pop() interface{} {
	*(*self).Matches = (*(*self).Matches)[:len(*(*self).Matches)-1]
	return nil
}

func (self *Overlap) Push(x interface{}) {
	*(*self).Matches = append(*(*self).Matches, x.(Match))
}

type Start struct {
	*Matches
}

func (self Start) Less(i, j int) bool {
	if (*self.Matches)[i].Strand == 1 {
		return (*self.Matches)[i].Interval.Start() < (*self.Matches)[j].Interval.Start()
	}
	return (*self.Matches)[i].Interval.Start() > (*self.Matches)[j].Interval.Start()
}

func main() {
	var (
		target    *gff.Reader
		source    *gff.Reader
		out       *gff.Writer
		indexFile *os.File
		e         error
		store     bool
	)

	targetName := flag.String("target", "", "Filename for input to be annotated. Defaults to stdin.")
	sourceName := flag.String("source", "", "Filename for source annotation.")
	indexName := flag.String("index", "", "Filename for index cache.")
	outName := flag.String("out", "", "Filename for output. Defaults to stdout.")
	flag.Float64Var(&minOverlap, "overlap", 0.05, "Overlap between features.")
	threads := flag.Int("threads", 2, "Number of threads to use.")
	bufferLen := flag.Int("buffer", 1000, "Length of ouput buffer.")
	help := flag.Bool("help", false, "Print this usage message.")

	flag.Parse()

	runtime.GOMAXPROCS(*threads)
	fmt.Fprintf(os.Stderr, "Using %d threads.\n", runtime.GOMAXPROCS(0))

	if *help || *sourceName == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *targetName == "" {
		fmt.Fprintln(os.Stderr, "Reading PALS features from stdin.")
		target = gff.NewReader(os.Stdin)
	} else if target, e = gff.NewReaderName(*targetName); e != nil {
		fmt.Fprintf(os.Stderr, "Error: %v.", e)
		os.Exit(0)
	} else {
		fmt.Fprintf(os.Stderr, "Reading target features from `%s'.\n", *targetName)
	}
	defer target.Close()

	switch {
	case *indexName == "" && *sourceName == "":
		fmt.Fprintln(os.Stderr, "No source or index provided.")
		os.Exit(0)
	case *indexName != "" && *sourceName == "":
		if indexFile, e = os.Open(*indexName); e != nil {
			fmt.Fprintf(os.Stderr, "Error: %v.\n", e)
			os.Exit(0)
		}
		defer indexFile.Close()
		store = false
	case *indexName != "" && *sourceName != "":
		if indexFile, e = os.Create(*indexName); e != nil {
			fmt.Fprintf(os.Stderr, "Error: %v.\n", e)
			os.Exit(0)
		}
		defer indexFile.Close()
		store = true
		fallthrough
	case *indexName == "" && *sourceName != "":
		if source, e = gff.NewReaderName(*sourceName); e != nil {
			fmt.Fprintf(os.Stderr, "Error: %v.\n", e)
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Reading annotation features from `%s'.\n", *sourceName)
		defer source.Close()
	}

	if *outName == "" {
		fmt.Fprintln(os.Stderr, "Writing annotation to stdout.")
		out = gff.NewWriter(os.Stdout, 2, 60, false)
	} else if out, e = gff.NewWriterName(*outName, 2, 60, true); e != nil {
		fmt.Fprintf(os.Stderr, "Error: %v.", e)
	} else {
		fmt.Fprintf(os.Stderr, "Writing annotation to `%s'.\n", *outName)
	}
	defer out.Close()

	intervalTree := interval.NewTree()

	for count := 0; ; count++ {
		if repeat, err := source.Read(); err != nil {
			break
		} else {
			fmt.Fprintf(os.Stderr, "Line: %d\r", count)
			repData := &RepeatRecord{}
			repData.Parse(repeat.Attributes)
			if repInterval, err := interval.New(string(repeat.Location), repeat.Start, repeat.End, 0, *repData); err == nil {
				intervalTree.Insert(repInterval)
			} else {
				fmt.Fprintf(os.Stderr, "Feature has end < start: %v\n", repeat)
			}
		}
	}

	process := make(chan *feat.Feature)
	buffer := make(chan *feat.Feature, *bufferLen)
	processWg := &sync.WaitGroup{}
	outputWg := &sync.WaitGroup{}

	if *threads < 2 {
		*threads = 2
	}
	for i := 0; i < *threads-1; i++ {
		processWg.Add(1)
		go processServer(intervalTree, process, buffer, processWg)
	}

	//output server
	outputWg.Add(1)
	go func() {
		defer outputWg.Done()
		for feature := range buffer {
			out.Write(feature)
		}
	}()

	for {
		if feature, err := target.Read(); err == nil {
			process <- feature
		} else {
			close(process)
			break
		}
	}

	if store {
		enc := gob.NewEncoder(indexFile)
		if e := enc.Encode(intervalTree); e != nil {
			fmt.Fprintf(os.Stderr, "Error: %v.\n", e)
			os.Exit(0)
		}
	}

	processWg.Wait()
	close(buffer)
	outputWg.Wait()
}

func processServer(index *interval.Tree, queue, output chan *feat.Feature, wg *sync.WaitGroup) {
	defer wg.Done()
	var (
		buffer      []byte   = make([]byte, 0, annotationLength)
		annotations Matches  = make(Matches, 0, maxAnnotations+1)
		o           *Overlap = &Overlap{&annotations}
		prefix      string   = ` ; Annot "`
		blank       string   = prefix + strings.Repeat("-", mapLength)
		overlap     int
	)

	for feature := range queue {
		annotations = annotations[:0]
		heap.Init(&Overlap{&annotations})
		buffer = buffer[:0]
		buffer = append(buffer, []byte(blank)...)
		if query, err := interval.New(string(feature.Location), feature.Start, feature.End, 0, nil); err != nil {
			fmt.Fprintf(os.Stderr, "Feature has end < start: %v\n", feature)
			continue
		} else {
			overlap = int(float64(feature.Len()) * minOverlap)
			if results := index.Intersect(query, overlap); results != nil {
				for hit := range results {
					o.Push(Match{
						Interval: hit,
						Overlap:  util.Min(hit.End(), query.End()) - util.Max(hit.Start(), query.Start()),
						Strand:   feature.Strand,
					})
					if len(annotations) > maxAnnotations {
						o.Pop()
					}
				}
			}
		}
		if len(annotations) > 0 {
			sort.Sort(&Start{&annotations})
			buffer = makeAnnotation(feature, annotations, len(prefix), buffer)
		}

		buffer = append(buffer, '"')
		feature.Attributes = append(feature.Attributes, buffer...)
		output <- feature
	}
}

func makeAnnotation(feature *feat.Feature, annotations Matches, prefixLen int, buffer []byte) []byte {
	var (
		annot                                []byte = buffer[prefixLen:]
		repRecord                            RepeatRecord
		repLocation                          *interval.Interval
		start, end, length                   int
		repStart, repEnd, repLeft, repLength int
		leftMargin, rightMargin              float64
		leftEnd, rightEnd                    bool
		mapStart, mapEnd                     int
		fullRepLength, repMissing            int
		i                                    int
		cLower, cUpper                       byte
		p                                    string
	)

	length = feature.Len()

	for annotIndex, repeat := range annotations {
		repRecord = repeat.Interval.Meta.(RepeatRecord)
		repLocation = repeat.Interval
		start = util.Max(repLocation.Start(), feature.Start)
		end = util.Min(repLocation.End(), feature.End)

		if repStart = repRecord.Start; repStart != -1 {
			repEnd = repRecord.End
			repLeft = repRecord.Left
			repLength = repStart + repLeft - 1
			if repLength <= 0 {
				repLength = 9999
			}

			if repLocation.Start() < feature.Start {
				repStart += feature.Start - repLocation.Start()
			}
			if repLocation.End() > feature.End {
				repEnd -= repLocation.End() - feature.End
			}

			leftMargin = float64(repStart) / float64(repLength)
			rightMargin = float64(repLeft) / float64(repLength)
			leftEnd = (leftMargin <= maxMargin)
			rightEnd = (rightMargin <= maxMargin)
		}

		mapStart = ((start - feature.Start) * mapLength) / length
		mapEnd = ((end - feature.Start) * mapLength) / length

		if feature.Strand == -1 {
			mapStart, mapEnd = mapLength-mapEnd-1, mapLength-mapStart-1
		}

		if mapStart < 0 || mapStart >= mapLength || mapEnd < 0 || mapEnd >= mapLength {
			fmt.Printf("mapStart: %d, mapEnd: %d, mapLength: %d\n", mapStart, mapEnd, mapLength)
			panic("MakeAnnot: failed to map")
		}

		cLower = 'a' + byte(annotIndex)
		cUpper = 'A' + byte(annotIndex)

		if leftEnd {
			annot[mapStart] = cUpper
		} else {
			annot[mapStart] = cLower
		}

		for i = mapStart + 1; i <= mapEnd-1; i++ {
			annot[i] = cLower
		}

		if rightEnd {
			annot[mapEnd] = cUpper
		} else {
			annot[mapEnd] = cLower
		}

		buffer = append(buffer, ' ')
		buffer = append(buffer, []byte(repRecord.Name)...)

		if repRecord.Start >= 0 {
			fullRepLength = repRecord.End + repLeft
			repMissing = repRecord.Start + repLeft
			if feature.Start > repLocation.Start() {
				repMissing += feature.Start - repLocation.Start()
			}
			if repLocation.End() > feature.End {
				repMissing += repLocation.End() - feature.End
			}
			p = fmt.Sprintf("(%.0f%%)", ((float64(fullRepLength)-float64(repMissing))*100)/float64(fullRepLength))
			buffer = append(buffer, []byte(p)...)
		}
	}

	return buffer
}
