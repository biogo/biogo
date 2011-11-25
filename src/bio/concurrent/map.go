// Map routines to iterate a function over an array, potentially splitting the array slice into
// chunks so that each chunk is processed concurrently. When using concurrent processing the
// Chunk size is either the nearest even division of the total array over the chosen concurrent
// processing goroutines or a specified maximum chunk size, whichever is smaller. Reducing
// chunk size can reduce the impact of divergence in time for processing chunks, but may add
// to overhead.
//
package concurrent
// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
//   This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU General Public License for more details.
//
//   You should have received a copy of the GNU General Public License
//   along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
import (
	"math"
	"bio"
	"bio/future"
	"bio/util"
)

// Apply a function to an array slice using a Processor
func Map(f Eval, slice []interface{}, threads, maxChunkSize int) (error error) {
	queue := make(chan interface{}, 1)
	buffer := make(chan Result)
	p := NewProcessor(f, threads, queue, buffer)
	defer p.Stop()

	chunkSize := util.Min(int(math.Ceil(float64(len(slice))/float64(threads))), maxChunkSize)

	quit := make(chan struct{})

	go func() {
		for s := 0; s*chunkSize < len(slice); s++ {
			select {
			case <-quit:
				break
			default:
				endChunk := util.Min(chunkSize*(s+1)-1, len(slice)-1)
				p.in <- slice[chunkSize*s : endChunk]
			}
		}
	}()

	for r := 0; r*chunkSize < len(slice); r++ {
		result := <-p.out
		if result.Error != nil {
			error = bio.NewError("Map failed", 0, error)
			close(quit)
			break
		}
	}

	return
}

// A future Map function - synchronisation is via a bio/future.Promise
func SpawnMap(f Eval, slice []interface{}, threads, maxChunkSize int) *future.Promise {
	promise := future.NewPromise(false, false, false)

	go func() {
		e := Map(f, slice, threads, maxChunkSize)
		if e == nil {
			promise.Fulfill(slice)
		} else {
			promise.Fail(nil, e)
		}
	}()

	return promise
}
