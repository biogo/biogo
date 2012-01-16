// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"flag"
	"fmt"
	"github.com/kortschak/BioGo/interval"
	"github.com/kortschak/BioGo/io/featio/bed"
	"math"
	"os"
)

func main() {
	var (
		region *bed.Reader
		motif  *bed.Reader
		err    error
	)

	motifName := flag.String("motif", "", "Filename for motif file.")
	regionName := flag.String("region", "", "Filename for region file.")
	verbose := flag.Bool("verbose", false, "Print details of identified motifs to stderr.")
	headerLine := flag.Bool("header", false, "Print a header line.")
	help := flag.Bool("help", false, "Print this usage message.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -motif <motif file> -region <region file>\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if *help || *regionName == "" || *motifName == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Open files
	if motif, err = bed.NewReaderName(*motifName, 3); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v.", err)
		os.Exit(0)
	} else {
		fmt.Fprintf(os.Stderr, "Reading motif features from `%s'.\n", *motifName)
	}
	defer motif.Close()

	if region, err = bed.NewReaderName(*regionName, 3); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v.", err)
		os.Exit(0)
	} else {
		fmt.Fprintf(os.Stderr, "Reading region features from `%s'.\n", *regionName)
	}
	defer region.Close()

	// Read in motif features and build interval tree to search
	intervalTree := interval.NewTree()

	for line := 1; ; line++ {
		if motifLine, err := motif.Read(); err != nil {
			break
		} else {
			if motifInterval, err := interval.New(string(motifLine.Location), motifLine.Start, motifLine.End, 0, nil); err == nil {
				intervalTree.Insert(motifInterval)
			} else {
				fmt.Fprintf(os.Stderr, "Line: %d: Feature has end < start: %v\n", line, motifLine)
			}
		}
	}

	// Read in region features and search for motifs within region
	// Calculate median motif location, sample standard deviation of locations
	// and mean distance of motif from midpoint of region for motifs contained
	// within region. Report these and n of motifs within region.
	if *headerLine {
		fmt.Println("Chromosome\tStart\tEnd\tn-hits\tMeanHitPos\tStddevHitPos\tMeanMidDistance")
	}
	for line := 1; ; line++ {
		if regionLine, err := region.Read(); err != nil {
			break
		} else {
			regionMidPoint := float64(regionLine.Start+regionLine.End) / 2
			if regionInterval, err := interval.New(string(regionLine.Location), regionLine.Start, regionLine.End, 0, regionMidPoint); err == nil {
				if *verbose {
					fmt.Fprintf(os.Stderr, "%s\t%d\t%d\n", regionLine.Location, regionLine.Start, regionLine.End)
				}
				sumOfDiffs, sumOfSquares, mean, oldmean, n := 0., 0., 0., 0., 0.
				for intersector := range intervalTree.Within(regionInterval, 0) {
					motifMidPoint := float64(intersector.Start()+intersector.End()) / 2
					if *verbose {
						fmt.Fprintf(os.Stderr, "\t%s\t%d\t%d\n", intersector.Chromosome(), intersector.Start(), intersector.End())
					}

					// The Method of Provisional Means	
					n++
					mean = oldmean + (motifMidPoint-oldmean)/n
					sumOfSquares += (motifMidPoint - oldmean) * (motifMidPoint - mean)
					oldmean = mean

					sumOfDiffs += math.Abs(motifMidPoint - regionMidPoint)
				}
				fmt.Printf("%s\t%d\t%d\t%0.f\t%0.f\t%f\t%f\n",
					regionLine.Location, regionLine.Start, regionLine.End,
					n, mean, math.Sqrt(sumOfSquares/(n-1)), sumOfDiffs/n)
			} else {
				fmt.Fprintf(os.Stderr, "Line: %d: Feature has end < start: %v\n", line, regionLine)
			}
		}
	}

}
