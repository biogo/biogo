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
				if e := recover(); e != nil {
					p.out <- Result{nil, bio.NewError("concurrent.Processor panic", 1, e)}
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
func (self *Processor) Process(value ...Operator) {
	for _, v := range value {
		self.in <- v
	}
}

// Get the next available result.
func (self *Processor) Result() (interface{}, error) {
	r := <-self.out
	return r.Value, r.Err
}

// Close the queue.
func (self *Processor) Close() {
	close(self.in)
}

// Return the number of working goroutines.
func (self *Processor) Working() int {
	return len(self.working)
}

// Terminate the goroutines.
func (self *Processor) Stop() {
	close(self.stop)
}

// Wait for all running processes to finish.
func (self *Processor) Wait() {
	self.wg.Wait()
}

type Result struct {
	Value interface{}
	Err   error
}
