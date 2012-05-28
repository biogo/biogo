// Copyright ©2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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

package cluster_test

import (
	"fmt"
	"github.com/kortschak/biogo/cluster"
	"github.com/kortschak/biogo/feat"
	"strings"
)

type Features []*feat.Feature

func (f Features) Len() int                    { return len(f) }
func (f Features) Values(i int) (x, y float64) { return float64(f[i].Start), float64(f[i].End) }

var feats = []*feat.Feature{
	{ID: "0", Start: 1, End: 1700},
	{ID: "1", Start: 2, End: 1700},
	{ID: "2", Start: 3, End: 610},
	{ID: "3", Start: 2, End: 605},
	{ID: "4", Start: 1, End: 600},
	{ID: "5", Start: 2, End: 750},
	{ID: "6", Start: 650, End: 900},
	{ID: "7", Start: 700, End: 950},
	{ID: "8", Start: 1000, End: 1700},
	{ID: "9", Start: 950, End: 1712},
	{ID: "10", Start: 1000, End: 1650},
}

func ExampleInterface() {
	km := cluster.NewKmeans(Features(feats))
	km.Seed(4)
	km.Cluster()
	for ci, c := range km.Clusters() {
		fmt.Printf("Cluster %d:\n", ci)
		for _, i := range c {
			f := feats[i]
			fmt.Println(
				strings.Repeat(" ", f.Start/20),
				strings.Repeat("-", f.Len()/20),
			)
		}
		fmt.Println()
	}

	var within float64
	for _, ss := range km.Within(){
		within += ss
	}
	fmt.Printf("betweenSS / totalSS = %.6f\n", 1-(within/km.Total()))

	// Output:
	// Cluster 0:
	//  ------------------------------------------------------------------------------------
	//  ------------------------------------------------------------------------------------
	//
	// Cluster 1:
	//                                                    -----------------------------------
	//                                                 --------------------------------------
	//                                                    --------------------------------
	//
	// Cluster 2:
	//                                  ------------
	//                                     ------------
	//
	// Cluster 3:
	//  ------------------------------
	//  ------------------------------
	//  -----------------------------
	//  -------------------------------------
	//
	// betweenSS / totalSS = 0.995335
}
