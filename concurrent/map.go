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

package concurrent

import (
	"code.google.com/p/biogo/bio"
	"code.google.com/p/biogo/util"
	"math"
)

// A Mapper is an Operator that can subdivide itself.
type Mapper interface {
	Operator
	Slice(i, j int) Mapper
	Len() int
}

// Map routines to iterate a function over an array, potentially splitting the array slice into
// chunks so that each chunk is processed concurrently. When using concurrent processing the
// Chunk size is either the nearest even division of the total array over the chosen concurrent
// processing goroutines or a specified maximum chunk size, whichever is smaller. Reducing
// chunk size can reduce the impact of divergence in time for processing chunks, but may add
// to overhead.
func Map(set Mapper, threads, maxChunkSize int) (results []interface{}, err error) {
	queue := make(chan Operator, 1)
	p := NewProcessor(queue, 0, threads)
	defer p.Stop()

	chunkSize := util.Min(int(math.Ceil(float64(set.Len())/float64(threads))), maxChunkSize)

	quit := make(chan struct{})

	go func() {
		for s := 0; s*chunkSize < set.Len(); s++ {
			select {
			case <-quit:
				break
			default:
				endChunk := util.Min(chunkSize*(s+1), set.Len())
				queue <- set.Slice(chunkSize*s, endChunk)
			}
		}
	}()

	for r := 0; r*chunkSize < set.Len(); r++ {
		result := <-p.out
		if result.Err != nil {
			err = bio.NewError("Map failed", 0, err)
			close(quit)
			break
		}
		results = append(results, result.Value)
	}

	return
}

// A future Map function - synchronisation is via a Promise.
func PromiseMap(set Mapper, threads, maxChunkSize int) *Promise {
	promise := NewPromise(false, false, false)

	go func() {
		result, err := Map(set, threads, maxChunkSize)
		if err == nil {
			promise.Fulfill(result)
		} else {
			promise.Fail(result, err)
		}
	}()

	return promise
}
