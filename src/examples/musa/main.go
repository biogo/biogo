//give some idea about patterns reported by ganesh - not rigorous
package main

import (
	"flag"
	"os"
	"bufio"
	"strings"
	"strconv"
	"fmt"
	"math"
)

type Weight struct {
	Name   string
	Weight float64
}

func main() {
	var in *bufio.Reader

	inName := flag.String("in", "", "Filename for input to be analysed. Defaults to stdin.")
	filterString := flag.String("f", "", "Comma separated list of categories to filter on.")
	closest := flag.Bool("relax", false, "Find the longest near match.")
	help := flag.Bool("help", false, "Print this usage message.")

	flag.Parse()

	if *help || *filterString == "" {
		flag.Usage()
		os.Exit(1)
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

	root := []*TrieNode{NewTrieNode(), NewTrieNode()}

	count := 0
	for ; ; count++ {
		if line, err := in.ReadString('\n'); err != nil {
			break
		} else {
			line = strings.TrimSpace(line)
			if row := strings.Split(line, "\t"); len(row) != 3 {
				fmt.Fprintln(os.Stderr, "Table corrupted at line %d.\n", count)
				os.Exit(0)
			} else {
				for i, part := range row[1:] {
					list := []*Weight{}
					part = strings.Trim(part, "[]()")
					for _, p := range strings.Split(part, ",") {
						if nw := strings.Split(p, "/"); len(nw) == 2 {
							if w, err := strconv.Atof64(nw[1]); err != nil {
								fmt.Fprintf(os.Stderr, "Float conversion error %v at line %d.\n", err, count)
							} else {
								list = append(list, &Weight{Name: nw[0], Weight: w})
							}
						} else {
							break
						}
					}
					if len(list) > 0 {
						root[i].Insert(list)
					}
				}
			}
		}
	}

	filter := strings.Split(*filterString, ",")
	result := root[0].Retrieve(filter)
	if !*closest && len(result) != len(filter) {
		os.Exit(1)
	}

	fmt.Printf("Probability: %e |", float64(len(result[len(result)-1]))/float64(count))
	for i, weights := range result {
		avg, stdev, n := Stat(weights)
		fmt.Printf(" %s/%e±%eSD n=%f", filter[i], avg, stdev, n)
	}
	fmt.Println()
}

func Stat(a []float64) (avg, stdev, n float64) {
	n = float64(len(a))
	for _, v := range a {
		avg += v
	}
	avg /= n

	for _, v := range a {
		stdev += math.Pow(v-avg, 2)
	}
	stdev /= n - 1

	return
}

type TrieNode struct {
	Weights  []float64
	Children map[string]*TrieNode
}

func NewTrieNode() *TrieNode {
	return &TrieNode{
		Weights:  make([]float64, 0),
		Children: map[string]*TrieNode{},
	}
}

func (self *TrieNode) Insert(item []*Weight) {
	name := item[0].Name
	if _, ok := self.Children[name]; !ok {
		self.Children[name] = NewTrieNode()
	}
	self.Children[name].Weights = append(self.Children[name].Weights, item[0].Weight)
	if len(item) > 1 {
		self.Children[name].Insert(item[1:])
	}
}

func (self *TrieNode) Retrieve(item []string) (w [][]float64) {
	if node, ok := self.Children[item[0]]; ok {
		w = append(w, node.Weights)
		if len(item) > 1 {
			cw := node.Retrieve(item[1:])
			w = append(w, cw...)
		}
	}

	return
}
