package interval
// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
// Derived from quicksect.py of bx-python ©James Taylor bitbucket.org/james_taylor/bx-python
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
	check "launchpad.net/gocheck"
	"math/rand"
	"reflect"
	"runtime"
	"testing"
)

// Helpers
func testTree(n, iLen, iLenVar, locRange int) (tree Tree) {
	tree = NewTree()
	for j := 0; j < n; j++ {
		i := randomInterval(iLen, iLenVar, locRange)
		tree.Insert(i)
	}

	return
}

func randomInterval(iLength, iLenVariance, locRange int) (i *Interval) {
	start := rand.Int() % locRange
	end := start + int(rand.NormFloat64()*2*float64(iLenVariance)) + iLength
	if end > locRange {
		end = locRange
	}
	if end < start {
		start, end = end, start
	}
	i, _ = New("", start, end, 0, nil)

	return
}

func allIntervalsOf(tree Tree, n int) (ss []*Interval) {
	ss = make([]*Interval, 0, n)
	for s := range tree.Traverse("") {
		ss = append(ss, s)
	}

	return
}

// Build a tree from a simplified Newick format returning the root node.
// Single letter node names only, no error checking and all nodes are full or leaf.
func makeTree(desc string) (tree *Interval) {
	tree = &Interval{}
	current := tree

	for _, b := range desc {
		switch b {
		case '(':
			current.left = &Interval{parent: current}
			current = current.left
		case ',':
			current.parent.right = &Interval{parent: current.parent}
			current = current.parent.right
		case ')':
			current = current.parent
		case ';':
			break
		default:
			current.line = b
		}
	}

	return
}

// Return a Newick format description of a tree defined by a node
func describeTree(n *Interval) string {
	r := make(chan byte)

	var follow func(*Interval)
	follow = func(n *Interval) {
		children := n.left != nil || n.right != nil
		if children {
			r <- '('
		}
		if n.left != nil {
			follow(n.left)
		}
		if children {
			r <- ','
		}
		if n.right != nil {
			follow(n.right)
		}
		if children {
			r <- ')'
		}
		r <- byte(n.line)
	}

	go func() {
		defer close(r)
		follow(n)
		r <- ';'
	}()

	ss := []byte{}
	for s := range r {
		ss = append(ss, s)
	}

	return string(ss)
}

// Checkers
type intLessChecker struct {
	*check.CheckerInfo
}

var lessThan check.Checker = &intLessChecker{
	&check.CheckerInfo{Name: "LessThan", Params: []string{"obtained", "expected"}},
}

func (checker *intLessChecker) Check(params []interface{}, names []string) (result bool, error string) {
	return params[0].(int) < params[1].(int), ""
}

type deepChecker struct {
	*check.CheckerInfo
}

var deepEquals check.Checker = &deepChecker{
	&check.CheckerInfo{Name: "DeepEquals", Params: []string{"obtained", "expected"}},
}

func (checker *deepChecker) Check(params []interface{}, names []string) (result bool, error string) {
	return reflect.DeepEqual(params[0], params[1]), ""
}

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

func (s *S) TestMakeAndDescribeTree(c *check.C) {
	desc := "((a,c)b,(e,g)f)d;"
	tree := makeTree(desc)

	c.Check(describeTree(tree), check.Equals, desc)
}

// ((a,c)b,(e,g)f)d -rotL-> (((a,c)b,e)d,g)f
func (s *S) TestRotateLeft(c *check.C) {
	orig := "((a,c)b,(e,g)f)d;"
	rot := "(((a,c)b,e)d,g)f;"

	tree := makeTree(orig)

	tree = tree.rotateLeft()
	c.Check(describeTree(tree), check.Equals, rot)

	rotTree := makeTree(rot)
	c.Check(tree, deepEquals, rotTree)
}

// ((a,c)b,(e,g)f)d -rotR-> (a,(c,(e,g)f)d)b
func (s *S) TestRotateRight(c *check.C) {
	orig := "((a,c)b,(e,g)f)d;"
	rot := "(a,(c,(e,g)f)d)b;"

	tree := makeTree(orig)

	tree = tree.rotateRight()
	c.Check(describeTree(tree), check.Equals, rot)

	rotTree := makeTree(rot)
	c.Check(tree, deepEquals, rotTree)
}

func (s *S) TestScan(c *check.C) {
	n := int(1e4)
	tree := testTree(n, 1e3, 1e2, 1e5)

	root := tree[""]

	ss := allIntervalsOf(tree, n)

	leftMost := root.LeftMost()
	c.Check(leftMost, check.Equals, ss[0])
	c.Check(leftMost.ScanLeft(), check.IsNil)

	rightMost := root.RightMost()
	c.Check(rightMost, check.Equals, ss[len(ss)-1])
	c.Check(rightMost.ScanRight(), check.IsNil)

	var (
		last  *Interval
		count int
	)
	for s := leftMost; s != nil && count <= n+1; s, count = s.ScanRight(), count+1 {
		if last != nil {
			c.Check(s.start, check.Not(lessThan), last.start)
		}
	}
	c.Check(count, check.Equals, n)

	last, count = nil, 0
	for s := rightMost; s != nil && count <= n+1; s, count = s.ScanLeft(), count+1 {
		if last != nil {
			c.Check(last.start, check.Not(lessThan), s.start)
		}
	}
	c.Check(count, check.Equals, n)

	desc := "((a,c)b,(e,g)f)d;"
	mock := makeTree(desc)
	c.Check(mock.LeftMost().ScanLeft(), check.IsNil)
	c.Check(mock.RightMost().ScanRight(), check.IsNil)
	c.Check(string(mock.LeftMost().line), check.Equals, "a")

	for label, node := 'a', mock.LeftMost(); label <= 'g'; label, node = label+1, node.ScanRight() {
		c.Check(string(node.line), check.Equals, string(label))
	}

	for i := range ss {
		if i > 0 {
			c.Check(ss[i].ScanLeft(), deepEquals, ss[i-1])
			c.Check(ss[i].ScanLeft().ScanRight(), deepEquals, ss[i])
		}
		if i < len(ss)-1 {
			c.Check(ss[i].ScanRight(), deepEquals, ss[i+1])
			c.Check(ss[i].ScanRight().ScanLeft(), deepEquals, ss[i])
		}
	}
}

func (s *S) TestIntersect(c *check.C) {
	n := int(1e4)
	tree := testTree(n, 1e3, 1e2, 1e5)

	exhaustive := 0
	intersects := 0
	falseHits := 0
	for i := 0; i < 1e3; i++ {
		test := randomInterval(1e4, 1e2, 1e5)

		// count all intersecting intervals in the tree
		for s := range tree.Traverse("") {
			if test.start <= s.end && test.end >= s.start {
				exhaustive++
			}
		}

		for s := range tree.Intersect(test, 0) {
			intersects++
			if test.start > s.end || test.end < s.start {
				falseHits++
			}
		}
	}

	c.Check(intersects, check.Equals, exhaustive)
	c.Check(falseHits, check.Equals, 0)
}

func (s *S) TestContain(c *check.C) {
	n := int(1e4)
	tree := testTree(n, 1e3, 1e2, 1e5)

	exhaustive := 0
	contains := 0
	falseHits := 0
	for i := 0; i < 1e3; i++ {
		test := randomInterval(1e4, 1e2, 1e5)

		// count all containing intervals in the tree
		for s := range tree.Traverse("") {
			if test.start >= s.start && test.end <= s.end {
				exhaustive++
			}
		}

		for s := range tree.Contain(test, 0) {
			contains++
			if test.start < s.start || test.end > s.end {
				falseHits++
			}
		}
	}

	c.Check(contains, check.Equals, exhaustive)
	c.Check(falseHits, check.Equals, 0)
}

func (s *S) TestWithin(c *check.C) {
	n := int(1e4)
	tree := testTree(n, 1e3, 1e2, 1e5)

	exhaustive := 0
	withins := 0
	falseHits := 0
	for i := 0; i < 1e3; i++ {
		test := randomInterval(1e4, 1e2, 1e5)

		// count all contained intervals in the tree
		for s := range tree.Traverse("") {
			if test.start <= s.start && test.end >= s.end {
				exhaustive++
			}
		}

		for s := range tree.Within(test, 0) {
			withins++
			if test.start > s.start || test.end < s.end {
				falseHits++
			}
		}
	}

	c.Check(withins, check.Equals, exhaustive)
	c.Check(falseHits, check.Equals, 0)
}

func (s *S) TestLinearRootDelete(c *check.C) {
	bug := check.Bug("Bug in *Interval.remove(). Fixed in 8aa60.")
	tree := NewTree()
	a, err := New("", 0, 1, 0, nil)
	c.Check(err, check.Equals, nil)
	b, err := New("", 2, 3, 0, nil)
	c.Check(err, check.Equals, nil)
	tree.Insert(a)
	tree.Insert(b)
	r := tree.Remove(tree[""]) // remove the root
	switch r {
	case a:
		c.Check(tree[""], check.Equals, b, bug)
	case b:
		c.Check(tree[""], check.Equals, a, bug)
	case nil:
		c.Errorf("Remove returned <nil>. %s", bug)
	default:
		c.Errorf("Unexpected return value: %q. %s", r, bug)
	}
}

func (s *S) TestRemove(c *check.C) {
	n := int(1e5)
	tree := testTree(n, 1e3, 1e2, 1e6)

	root := tree[""]

	count := 0
	for _ = range tree.Intersect(root, 0) {
		count++
	}
	// Remove the root
	tree.Remove(root)
	// Check one less interval here
	for _ = range tree.Intersect(root, 0) {
		count--
	}
	c.Check(count, check.Equals, 1)

	// Remove all intersectors of the root
	ss := []*Interval{}
	for s := range tree.Intersect(root, 0) {
		ss = append(ss, s)
	}
	for _, s := range ss {
		tree.FastRemove(s)
	}
	tree.AdjustRange("")
	// Check no interval left here
	found := []*Interval{}
	for s := range tree.Intersect(root, 0) {
		found = append(found, s)
	}

	c.Check(len(found), check.Equals, 0)
}

func (s *S) TestInvariants(c *check.C) {
	n := int(1e4)
	tree := testTree(n, 1e3, 1e2, 1e5)
	for s := range tree.Traverse("") {
		if s.parent != nil {
			c.Check(s.parent.priority, check.Not(lessThan), s.priority)
		}
	}

	// test removal of root
	tree.Remove(tree[""])
	var last *Interval
	for s := range tree.Traverse("") {
		if last != nil {
			c.Check(s.start, check.Not(lessThan), last.start)
		}
		if s.parent != nil {
			c.Check(s.parent.priority, check.Not(lessThan), s.priority)
		}
		last = s
	}

	// get a set of all the intervals in the tree
	ss := allIntervalsOf(tree, n)
	rnd := make(map[int]struct{}, len(ss))
	for i := range ss {
		rnd[i] = struct{}{}
	}

	// test removal of left-most
	tree.Remove(ss[0])
	last = nil
	for s := range tree.Traverse("") {
		if last != nil {
			c.Check(s.start, check.Not(lessThan), last.start)
		}
		if s.parent != nil {
			c.Check(s.parent.priority, check.Not(lessThan), s.priority)
		}
		last = s
	}

	// test removal of parent of right-most
	tree.Remove(ss[len(ss)-1].parent)
	for s := range tree.Traverse("") {
		if s.parent != nil {
			c.Check(s.parent.priority, check.Not(lessThan), s.priority)
		}
	}

	// test random (hash order) removal of 100 or n (whichever is less) nodes
	l := 0
	for i := range rnd {
		tree.Remove(ss[i])
		l++
		if l >= 100 || l >= n {
			break
		}
	}
	last = nil
	for s := range tree.Traverse("") {
		if last != nil {
			c.Check(s.start, check.Not(lessThan), last.start)
		}
		if s.parent != nil {
			c.Check(s.parent.priority, check.Not(lessThan), s.priority)
		}
		last = s
	}
}

func (s *S) TestRemoveInsert(c *check.C) {
	n := int(1e4)
	tree := testTree(n, 1e3, 1e2, 1e5)
	ss1 := make([]string, 0, n)
	ss2 := make([]string, 0, n)
	for s := range tree.Traverse("") {
		ss1 = append(ss1, s.String())
	}
	tree.Insert(tree.Remove(tree[""]))
	for s := range tree.Traverse("") {
		ss2 = append(ss2, s.String())
	}
	c.Check(ss1, check.Equals, ss2)
}

// Benchmarks
func repeatInsertion(tree Tree, n, iLen, iLenVar, locRange int, b *testing.B) {
	for j := 0; j < n; j++ {
		b.StopTimer()
		i := randomInterval(iLen, iLenVar, locRange)
		b.StartTimer()
		tree.Insert(i)
	}
}

func repeatInsert(b *testing.B, n int) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < 1e6/n; j++ {
			tree := NewTree()
			repeatInsertion(tree, n, 1e3, 1e2, 1e5, b)
		}
	}
}

func BenchmarkTreeMillionInsert1e2(b *testing.B) {
	repeatInsert(b, 1e2)
}

func BenchmarkTreeMillionInsert1e3(b *testing.B) {
	repeatInsert(b, 1e3)
}

func BenchmarkTreeMillionInsert1e5(b *testing.B) {
	repeatInsert(b, 1e5)
}

func repeatIntersect(b *testing.B, n int) {
	b.StopTimer()
	tree := testTree(n, 3e2, 1e1, 1e3*n)
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		s := randomInterval(1e3, 1e2, 1e5)
		b.StartTimer()
		for _ = range tree.Intersect(s, 0) {
		}
	}
}

func BenchmarkTreeIntersect1e2(b *testing.B) {
	repeatIntersect(b, 1e2)
}

func BenchmarkTreeIntersect1e4(b *testing.B) {
	repeatIntersect(b, 1e4)
}

func BenchmarkTreeIntersect1e6(b *testing.B) {
	repeatIntersect(b, 1e6)
}

func repeatTraverse(b *testing.B, n int) {
	b.StopTimer()
	tree := testTree(n, 1e3, 1e2, 1e5)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for _ = range tree.Traverse("") {
		}
	}
}

func BenchmarkTreeTraverse1e2(b *testing.B) {
	repeatTraverse(b, 1e2)
}

func BenchmarkTreeTraverse1e4(b *testing.B) {
	repeatTraverse(b, 1e4)
}

func BenchmarkTreeTraverse1e6(b *testing.B) {
	repeatTraverse(b, 1e6)
}

func repeatFlatten(b *testing.B, n int) {
	b.StopTimer()
	tree := testTree(n, 1e3, 1e2, 1e5)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 1e6/n; j++ {
			start, end := tree.Range("")
			s, _ := New("", start, end, 0, nil)
			tree.Flatten(s, 0, 0)
		}
	}
}

func BenchmarkTreeFlatten1e2(b *testing.B) {
	repeatFlatten(b, 1e2)
}

func BenchmarkTreeFlatten1e4(b *testing.B) {
	repeatFlatten(b, 1e4)
}

func BenchmarkTreeFlatten1e6(b *testing.B) {
	repeatFlatten(b, 1e6)
}

func repeatRemove(b *testing.B, n int) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < 1e3/n; j++ {
			b.StopTimer()
			runtime.GC()
			tree := testTree(n, 1e3, 1e2, 1e5)
			ss := make([]*Interval, 0, n)
			for s := range tree.Traverse("") {
				ss = append(ss, s)
			}
			b.StartTimer()
			for j := range ss {
				tree.Remove(ss[j])
				ss[j] = nil
			}
		}
	}
}

func BenchmarkTreeRemove1e2(b *testing.B) {
	repeatRemove(b, 1e2)
}

func BenchmarkTreeRemove1e3(b *testing.B) {
	repeatRemove(b, 1e3)
}
