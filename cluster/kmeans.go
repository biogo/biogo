// Package cluster provides limited data clustering support.
//
// The cluster package is intended to become a more extensive clustering package.
// At this stage only Lloyd's k-means clustering for ℝ² data is supported.
package cluster

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

import (
	"math"
	"math/rand"
)

// A type, typically a collection, that satisfies cluster.Interface can be clustered by a Kmeans.
// The Kmeans Cluster method requires that the elements of the collection be enumerated by an integer index. 
type Interface interface {
	Len() int                    // Return the length of the data slice.
	Values(i int) (x, y float64) // Return the data values for element i as float64.
}

type val struct {
	x, y float64
}

// X returns the x-coordinate of the point.
func (self val) X() float64 { return self.x }

// Y returns the y-coordinate of the point.
func (self val) Y() float64 { return self.y }

// A Value is the representation of a data point within the clustering object.
type Value struct {
	val
	cluster int
}

// Cluster returns the cluster membership of the Value.
func (self Value) Cluster() int { return self.cluster }

type Center struct {
	val
	count int
}

// Count returns the number of members of the Center's cluster.
func (self Center) Count() int { return self.count }

// A Kmeans clusters ℝ² data according to the Lloyd k-means algorithm.
type Kmeans struct {
	values []Value
	means  []Center
}

// NewKmeans creates a new Kmeans object populated with data from an Interface value, data.
func NewKmeans(data Interface) *Kmeans {
	return &Kmeans{
		values: convert(data),
	}
}

// Convert the data to the internal float64 representation.
func convert(data Interface) []Value {
	va := make([]Value, data.Len())
	for i := 0; i < data.Len(); i++ {
		x, y := data.Values(i)
		va[i] = Value{val: val{x: x, y: y}}
	}

	return va
}

// Seed generates the initial means for the k-means algorithm.
func (self *Kmeans) Seed(k int) {
	self.means = make([]Center, k)

	self.means[0].val = self.values[rand.Intn(len(self.values))].val
	d := make([]float64, len(self.values))
	for i := 1; i < k; i++ {
		sum := 0.
		for j, v := range self.values {
			_, min := self.nearest(v.val)
			d[j] = min * min
			sum += d[j]
		}
		target := rand.Float64() * sum
		j := 0
		for sum = d[0]; sum < target; sum += d[j] {
			j++
		}
		self.means[i].val = self.values[j].val
	}
}

// Find the nearest center to the point v. Returns c, the index of the nearest center
// and min, the distance from v to that center.
func (self *Kmeans) nearest(v val) (c int, min float64) {
	min = math.Hypot(v.x-self.means[0].x, v.y-self.means[0].y)

	for i := 1; i < len(self.means); i++ {
		d := math.Hypot(v.x-self.means[i].x, v.y-self.means[i].y)
		if d < min {
			min = d
			c = i
		}
	}

	return
}

// Cluster the data using the standard k-means algorithm.
func (self *Kmeans) Cluster() {
	for i, v := range self.values {
		n, _ := self.nearest(v.val)
		self.values[i].cluster = n
	}

	for {
		for i := range self.means {
			self.means[i] = Center{}
		}
		for _, v := range self.values {
			self.means[v.cluster].x += v.x
			self.means[v.cluster].y += v.y
			self.means[v.cluster].count++
		}
		for i := range self.means {
			inv := 1 / float64(self.means[i].count)
			self.means[i].x *= inv
			self.means[i].y *= inv
		}

		deltas := 0
		for i, v := range self.values {
			if n, _ := self.nearest(v.val); n != v.cluster {
				deltas++
				self.values[i].cluster = n
			}
		}
		if deltas == 0 {
			break
		}
	}
}

// Within calculates the total sum of squares for the data relative to the data mean.
func (self *Kmeans) Total() (ss float64) {
	var x, y float64

	for _, v := range self.values {
		x += v.x
		y += v.y
	}
	inv := 1 / float64(len(self.values))
	x *= inv
	y *= inv

	for _, v := range self.values {
		dx, dy := x-v.x, y-v.y
		ss += dx*dx + dy*dy
	}

	return
}

// Within calculates the sum of squares within each cluster.
// Returns nil if Cluster has not been called.
func (self *Kmeans) Within() (ss []float64) {
	if self.means == nil {
		return
	}
	ss = make([]float64, len(self.means))

	for _, v := range self.values {
		dx, dy := self.means[v.cluster].x-v.x, self.means[v.cluster].y-v.y
		ss[v.cluster] += dx*dx + dy*dy
	}

	return
}

// Means returns the k-means.
func (self *Kmeans) Means() (c []Center) { return self.means }

// Features returns a slice of the values in the Kmeans.
func (self *Kmeans) Values() (v []Value) { return self.values }

// Clusters returns the k clusters.
// Returns nil if Cluster has not been called.
func (self *Kmeans) Clusters() (c [][]int) {
	if self.means == nil {
		return
	}
	c = make([][]int, len(self.means))

	for i := range c {
		c[i] = make([]int, 0, self.means[i].count)
	}
	for i, v := range self.values {
		c[v.cluster] = append(c[v.cluster], i)
	}

	return
}
