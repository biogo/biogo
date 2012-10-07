// Changes copyright ©2011-2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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
//
// Derived from testing/benchmark.go Copyright 2009 The Go Authors under the BSD license.

package util

import "time"

type Timer struct {
	nanoseconds time.Duration
	start       time.Time
	interval    time.Time
}

func NewTimer() (t *Timer) {
	t = &Timer{}
	t.Start()
	return
}

// Start starts timing. This function is called automatically when a timer is created,
// but it can also used to resume timing after a call to StopTimer.
func (t *Timer) Start() { t.start = time.Now(); t.interval = t.start }

// Stop stops timing. This can be used to pause the timer while performing complex
// initialization that you don't want to measure.
func (t *Timer) Stop() time.Duration {
	if t.start.After(time.Time{}) {
		t.nanoseconds += time.Now().Sub(t.start)
	}
	t.start = time.Time{}

	return t.nanoseconds
}

// Reset stops the timer and sets the elapsed time to zero.
func (t *Timer) Reset() {
	t.start = time.Time{}
	t.nanoseconds = 0
}

// Time returns the measured time.
func (t *Timer) Time() time.Duration {
	return t.nanoseconds
}

// Start and return a time interval.
func (t *Timer) Interval() (l time.Duration) {
	if t.start.After(time.Time{}) {
		l = time.Now().Sub(t.interval)
		t.interval = time.Now()
	}

	return
}
