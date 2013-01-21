// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package concurrent

import (
	"fmt"
	"runtime"
	"sync"
)

// Interface is a type that performs an operation on itself, returning any error.
type Operator interface {
	Operation() (interface{}, error)
}

// The Processor type manages a number of concurrent Processes.
type Processor struct {
	in      chan Operator
	out     chan Result
	stop    chan struct{}
	working chan bool
	wg      *sync.WaitGroup
}

// Return a new Processor to operate the function f over the number of threads specified taking
// input from queue and placing the result in buffer. Threads is limited by GOMAXPROCS, if threads is greater
// GOMAXPROCS or less than 1 then threads is set to GOMAXPROCS.
func NewProcessor(queue chan Operator, buffer int, threads int) (p *Processor) {
	if available := runtime.GOMAXPROCS(0); threads > available || threads < 1 {
		threads = available
	}

	p = &Processor{
		in:      queue,
		out:     make(chan Result, buffer),
		stop:    make(chan struct{}),
		working: make(chan bool, threads),
		wg:      &sync.WaitGroup{},
	}

	for i := 0; i < threads; i++ {
		p.wg.Add(1)
		go func() {
			p.working <- true
			defer func() {
				if err := recover(); err != nil {
					p.out <- Result{nil, fmt.Errorf("concurrent: processor panic: %v", err)}
				}
				<-p.working
				if len(p.working) == 0 {
					close(p.out)
				}
				p.wg.Done()
			}()

			for input := range p.in {
				v, e := input.Operation()
				if p.out != nil {
					p.out <- Result{v, e}
				}
				select {
				case <-p.stop:
					return
				default:
				}
			}
		}()
	}

	return
}

// Submit values for processing.
func (p *Processor) Process(value ...Operator) {
	for _, v := range value {
		p.in <- v
	}
}

// Get the next available result.
func (p *Processor) Result() (interface{}, error) {
	r := <-p.out
	return r.Value, r.Err
}

// Close the queue.
func (p *Processor) Close() {
	close(p.in)
}

// Return the number of working goroutines.
func (p *Processor) Working() int {
	return len(p.working)
}

// Terminate the goroutines.
func (p *Processor) Stop() {
	close(p.stop)
}

// Wait for all running processes to finish.
func (p *Processor) Wait() {
	p.wg.Wait()
}

type Result struct {
	Value interface{}
	Err   error
}
