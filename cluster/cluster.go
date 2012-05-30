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

// Package cluster provides interfaces and types for data clustering.
//
// At this stage only Lloyd's k-means clustering of ℝ² data is supported in subpackages.
package cluster

// Clusterer is the common interface implemented by clustering types.
type Clusterer interface {
	// Cluster the data.
	Cluster()

	// Return a slice of slices of ints representing the indices of
	// the original data grouped into clusters.
	Clusters() (c [][]int)

	// Return a slice of sum of squares distances for the clusters.
	Within() (ss []float64)
	// Return the total sum of squares distance for the original data.
	Total() (ss float64)
}

// R2 is the interogation interface implemented by ℝ² Clusterers.
type R2 interface {
	// Return a slice of centers of the clusters.
	Means() (c []Center)
	// Return the internal representation of the original data.
	Values() (v []Value)
}

// RN is the interogation interface implemented by ℝⁿ Clusterers.
type RN interface {
	// Return a slice of centers of the clusters.
	Means() (c []NCenter)
	// Return the internal representation of the original data.
	Values() (v []NValue)
}

// A type, typically a collection, that satisfies cluster.Interface can be clustered by an ℝ² Clusterer.
// The Clusterer requires that the elements of the collection be enumerated by an integer index. 
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

// A Center is a representation of a cluster center in ℝ².
type Center struct {
	val
	count int
}

// Count returns the number of members of the Center's cluster.
func (self Center) Count() int { return self.count }

// A type, typically a collection, that satisfies cluster.Interface can be clustered by an ℝⁿ Clusterer.
// The Clusterer requires that the elements of the collection be enumerated by an integer index. 
type NInterface interface {
	Len() int                   // Return the length of the data slice.
	Values(i int) (v []float64) // Return the data values for element i as []float64.
}

type nval []float64

// V returns the ith coordinate of the point.
func (self nval) V(i int) float64 { return self[i] }

// A Value is the representation of a data point within the clustering object.
type NValue struct {
	nval
	cluster int
}

// Cluster returns the cluster membership of the NValue.
func (self NValue) Cluster() int { return self.cluster }

// An NCenter is a representation of a cluster center in ℝⁿ.
type NCenter struct {
	nval
	count int
}

// Count returns the number of members of the NCenter's cluster.
func (self NCenter) Count() int { return self.count }
