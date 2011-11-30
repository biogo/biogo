package main

import "time"

type Timer struct {
	nanoseconds int64
	start       int64
	interval    int64
}

func NewTimer() (t *Timer) {
	t = &Timer{}
	t.Start()
	return
}

// Start starts timing.  This function is called automatically
// when a timer is created, but it can also used to resume timing after
// a call to StopTimer.
func (self *Timer) Start() { self.start = time.Nanoseconds() }

// Stop stops timing.  This can be used to pause the timer
// while performing complex initialization that you don't
// want to measure.
func (self *Timer) Stop() int64 {
	if self.start > 0 {
		self.nanoseconds += time.Nanoseconds() - self.start
	}
	self.start = 0

	return self.nanoseconds
}

// Reset stops the timer and sets the elapsed time to zero.
func (self *Timer) Reset() {
	self.start = 0
	self.nanoseconds = 0
}

// Time returns the measured time.
func (self *Timer) Time() int64 {
	return self.nanoseconds
}

// Start and return a time interval.
func (self *Timer) Interval() (l int64) {
	if self.start > 0 {
		l = time.Nanoseconds() - self.interval
		self.interval = time.Nanoseconds()
	}

	return
}
