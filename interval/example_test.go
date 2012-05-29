// Copyright ©2011-2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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

package interval

import (
	"fmt"
	"strings"
)

func ExampleTree_Insert() {
	tree := NewTree()
	chromosome := "example"
	segments := [][]int{{0, 4}, {8, 12}, {2, 3}, {5, 10}, {3, 7}, {-22, 6}, {34, 61}}

	for _, s := range segments {
		if i, err := New(chromosome, s[0], s[1], 0, nil); err == nil {
			tree.Insert(i)
		} else {
			fmt.Println(err)
		}
	}

	PrintAll(tree)
	// Output:
	// "example":[-22, 6), "example":[0, 4), "example":[2, 3), "example":[3, 7), "example":[5, 10), "example":[8, 12), "example":[34, 61)
}

func ExampleTree_Intersect() {
	tree := CreateExampleTree("example", [][]int{{0, 4}, {8, 12}, {2, 3}, {5, 10}, {3, 7}, {-22, 6}, {34, 61}})
	if i, err := New("example", -15, -2, 0, nil); err == nil {
		for s := range tree.Intersect(i, 0) {
			fmt.Printf("%s\n", s)
		}
	}
	// Output:
	// "example":[-22, 6)
}

func ExampleTree_Contain() {
	tree := CreateExampleTree("example", [][]int{{0, 4}, {8, 12}, {2, 3}, {5, 10}, {3, 7}, {-22, 6}, {34, 61}})
	if i, err := New("example", 4, 6, 0, nil); err == nil {
		for s := range tree.Contain(i, 0) {
			fmt.Printf("%s\n", s)
		}
	}
	// Output:
	// "example":[-22, 6) 
	// "example":[3, 7)
}

func ExampleTree_Within() {
	tree := CreateExampleTree("example", [][]int{{0, 4}, {8, 12}, {2, 3}, {5, 10}, {3, 7}, {-22, 6}, {34, 61}})
	if i, err := New("example", 1, 5, 0, nil); err == nil {
		for s := range tree.Within(i, 0) {
			fmt.Printf("%s\n", s)
		}
	}
	// Output:
	// "example":[2, 3)
}

func ExampleTree_Remove() {
	tree := CreateExampleTree("example", [][]int{{0, 4}, {8, 12}, {2, 3}, {5, 10}, {3, 7}, {-22, 6}, {34, 61}})
	if i, err := New("example", -15, -2, 0, nil); err == nil {
		for s := range tree.Intersect(i, 0) {
			r := tree.Remove(s)
			fmt.Println(r, r.Left(), r.Right(), r.Parent())
		}
	}

	PrintAll(tree)
	// Output:
	// "example":[-22, 6) <nil> <nil> <nil>
	// "example":[0, 4), "example":[2, 3), "example":[3, 7), "example":[5, 10), "example":[8, 12), "example":[34, 61)
}

func ExampleTree_Merge() {
	tree := CreateExampleTree("example", [][]int{{0, 4}, {8, 12}, {2, 3}, {5, 10}, {3, 7}, {34, 61}})
	chromosome := "example"
	segments := [][]int{{0, 1}, {27, 42}}
	inserted := []*Interval{}
	replaced := []*Interval{}

	for _, s := range segments {
		if i, err := New(chromosome, s[0], s[1], 0, nil); err == nil {
			o := tree.Merge(i, 0)
			inserted = append(inserted, i)
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
	// Output:
	// Inserted
	// "example":[0, 4)
	// "example":[27, 61)
	// Replaced
	// "example":[0, 4)
	// "example":[34, 61)
	// "example":[0, 4), "example":[2, 3), "example":[3, 7), "example":[5, 10), "example":[8, 12), "example":[27, 61)
}

func ExampleTree_TraverseAll() {
	tree := CreateExampleTree("example", [][]int{{0, 4}, {8, 12}, {2, 3}, {5, 10}, {3, 7}, {27, 61}})
	segs := []string{}
	for i := range tree.TraverseAll() {
		segs = append(segs, i.String())
	}
	fmt.Println(strings.Join(segs, ", "))
	// Output:
	// "example":[0, 4), "example":[2, 3), "example":[3, 7), "example":[5, 10), "example":[8, 12), "example":[27, 61)
}

func ExampleTree_Flatten() {
	tree := CreateExampleTree("example", [][]int{{0, 4}, {8, 12}, {2, 3}, {5, 10}, {3, 7}, {27, 61}})
	start, end := tree.Range("example")
	if i, err := New("example", start, end, 0, nil); err == nil {
		flat, original := tree.Flatten(i, 0, 0)
		fmt.Printf("flattened: %v\noriginal: %v\n", flat, original)
	}
	// Output:
	// flattened: ["example":[0, 12) "example":[27, 61)]
	// original: [["example":[0, 4) "example":[2, 3) "example":[3, 7) "example":[5, 10) "example":[8, 12)] ["example":[27, 61)]]
}

func ExampleTree_FlattenContain() {
	tree := CreateExampleTree("example", [][]int{{0, 4}, {8, 12}, {2, 3}, {5, 10}, {3, 7}, {-22, 6}, {34, 61}})

	if i, err := New("example", 6, 7, 0, nil); err == nil {
		flat, original := tree.FlattenContaining(i, 0, 0)
		fmt.Printf("flattened: %v\noriginal: %v\n", flat, original)
	}
	// Output:
	// flattened: ["example":[3, 10)]
	// original: [["example":[3, 7) "example":[5, 10)]]
}

func ExampleTree_FlattenWithin() {
	tree := CreateExampleTree("example", [][]int{{0, 4}, {8, 12}, {2, 3}, {5, 10}, {3, 7}, {-22, 6}, {34, 61}})

	if i, err := New("example", 2, 7, 0, nil); err == nil {
		flat, original := tree.FlattenWithin(i, 0, 0)
		fmt.Printf("flattened: %v\noriginal: %v\n", flat, original)
	}
	// Output:
	// flattened: ["example":[2, 7)]
	// original: [["example":[2, 3) "example":[3, 7)]]
}

// Helpers

func CreateExampleTree(chromosome string, segments [][]int) (tree Tree) {
	tree = NewTree()

	for _, s := range segments {
		i, _ := New(chromosome, s[0], s[1], 0, nil)
		tree.Insert(i)
	}

	return
}

func PrintAll(t Tree) {
	segs := []string{}
	for i := range t.TraverseAll() {
		segs = append(segs, i.String())
	}
	fmt.Println(strings.Join(segs, ", "))
}
