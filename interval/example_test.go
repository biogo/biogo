package interval_test
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

import (
	"fmt"
	"github.com/kortschak/BioGo/interval"
	"strings"
)

// "example":[-22, 6), "example":[0, 4), "example":[2, 3), "example":[3, 7), "example":[5, 10), "example":[8, 12), "example":[34, 61)
func ExampleInsert() {
	tree := interval.NewTree()
	chromosome := "example"
	segments := [][]int{{0, 4}, {8, 12}, {2, 3}, {5, 10}, {3, 7}, {-22, 6}, {34, 61}}

	for _, s := range segments {
		if i, err := interval.New(chromosome, s[0], s[1], 0, nil); err == nil {
			tree.Insert(i)
		} else {
			fmt.Println(err)
		}
	}

	PrintAll(tree)
}

// "example":[-22, 6)
func ExampleIntersect() {
	tree := CreateExampleTree("example", [][]int{{0, 4}, {8, 12}, {2, 3}, {5, 10}, {3, 7}, {-22, 6}, {34, 61}})
	if i, err := interval.New("example", -15, -2, 0, nil); err == nil {
		for s := range tree.Intersect(i, 0) {
			fmt.Printf("%s\n", s)
		}
	}
}

// "example":[-22, 6) 
// "example":[3, 7)
func ExampleContain() {
	tree := CreateExampleTree("example", [][]int{{0, 4}, {8, 12}, {2, 3}, {5, 10}, {3, 7}, {-22, 6}, {34, 61}})
	if i, err := interval.New("example", 4, 6, 0, nil); err == nil {
		for s := range tree.Contain(i, 0) {
			fmt.Printf("%s\n", s)
		}
	}
}

// "example":[2, 3)
func ExampleWithin() {
	tree := CreateExampleTree("example", [][]int{{0, 4}, {8, 12}, {2, 3}, {5, 10}, {3, 7}, {-22, 6}, {34, 61}})
	if i, err := interval.New("example", 1, 5, 0, nil); err == nil {
		for s := range tree.Within(i, 0) {
			fmt.Printf("%s\n", s)
		}
	}
}

// "example":[0, 4), "example":[2, 3), "example":[3, 7), "example":[5, 10), "example":[8, 12), "example":[34, 61)
func ExampleRemove() {
	tree := CreateExampleTree("example", [][]int{{0, 4}, {8, 12}, {2, 3}, {5, 10}, {3, 7}, {-22, 6}, {34, 61}})
	if i, err := interval.New("example", -15, -2, 0, nil); err == nil {
		for s := range tree.Intersect(i, 0) {
			tree.Remove(s)
		}
	}

	PrintAll(tree)
}

// Inserted
// "example":[0, 4)
// "example":[27, 61)
// Replaced
// "example":[0, 4)
// "example":[34, 61)
// "example":[0, 4), "example":[2, 3), "example":[3, 7), "example":[5, 10), "example":[8, 12), "example":[27, 61)
func ExampleMerge() {
	tree := CreateExampleTree("example", [][]int{{0, 4}, {8, 12}, {2, 3}, {5, 10}, {3, 7}, {34, 61}})
	chromosome := "example"
	segments := [][]int{{0, 1}, {27, 42}}
	inserted := []*interval.Interval{}
	replaced := []*interval.Interval{}

	for _, s := range segments {
		if i, err := interval.New(chromosome, s[0], s[1], 0, nil); err == nil {
			n, o := tree.Merge(i, 0)
			inserted = append(inserted, n...)
			replaced = append(replaced, o...)
		} else {
			fmt.Println(err)
		}
	}

	fmt.Println("Inserted")
	for _, s := range inserted {
		fmt.Printf("%s\n", s)
	}

	fmt.Println("Replaced")
	for _, s := range replaced {
		fmt.Printf("%s\n", s)
	}

	PrintAll(tree)
}

// "example":[0, 4), "example":[2, 3), "example":[3, 7), "example":[5, 10), "example":[8, 12), "example":[27, 61)
func ExampleTraverseAll() {
	tree := CreateExampleTree("example", [][]int{{0, 4}, {8, 12}, {2, 3}, {5, 10}, {3, 7}, {27, 61}})
	segs := []string{}
	for i := range tree.TraverseAll() {
		segs = append(segs, i.String())
	}
	fmt.Println(strings.Join(segs, ", "))
}

// flattened: ["example":[0, 12) "example":[27, 61)]
// original: [["example":[0, 4) "example":[2, 3) "example":[3, 7) "example":[5, 10) "example":[8, 12)] ["example":[27, 61)]]
func ExampleFlatten() {
	tree := CreateExampleTree("example", [][]int{{0, 4}, {8, 12}, {2, 3}, {5, 10}, {3, 7}, {27, 61}})
	start, end := tree.Range("example")
	if i, err := interval.New("example", start, end, 0, nil); err == nil {
		flat, original := tree.Flatten(i, 0, 0)
		fmt.Printf("flattened: %v\noriginal: %v\n", flat, original)
	}
}

// flattened: ["example":[3, 10)]
// original: [["example":[3, 7) "example":[5, 10)]]
func ExampleFlattenContain() {
	tree := CreateExampleTree("example", [][]int{{0, 4}, {8, 12}, {2, 3}, {5, 10}, {3, 7}, {-22, 6}, {34, 61}})

	if i, err := interval.New("example", 6, 7, 0, nil); err == nil {
		flat, original := tree.FlattenContaining(i, 0, 0)
		fmt.Printf("flattened: %v\noriginal: %v\n", flat, original)
	}
}

// flattened: ["example":[2, 7)]
// original: [["example":[2, 3) "example":[3, 7)]]
func ExampleFlattenWithin() {
	tree := CreateExampleTree("example", [][]int{{0, 4}, {8, 12}, {2, 3}, {5, 10}, {3, 7}, {-22, 6}, {34, 61}})

	if i, err := interval.New("example", 2, 7, 0, nil); err == nil {
		flat, original := tree.FlattenWithin(i, 0, 0)
		fmt.Printf("flattened: %v\noriginal: %v\n", flat, original)
	}
}

// Helpers

func CreateExampleTree(chromosome string, segments [][]int) (tree interval.Tree) {
	tree = interval.NewTree()

	for _, s := range segments {
		i, _ := interval.New(chromosome, s[0], s[1], 0, nil)
		tree.Insert(i)
	}

	return
}

func PrintAll(t interval.Tree) {
	segs := []string{}
	for i := range t.TraverseAll() {
		segs = append(segs, i.String())
	}
	fmt.Println(strings.Join(segs, ", "))
}
