// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
