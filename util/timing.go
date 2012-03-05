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

// Start starts timing.  This function is called automatically
// when a timer is created, but it can also used to resume timing after
// a call to StopTimer.
func (self *Timer) Start() { self.start = time.Now() }

// Stop stops timing.  This can be used to pause the timer
// while performing complex initialization that you don't
// want to measure.
func (self *Timer) Stop() time.Duration {
	if self.start.After(time.Time{}) {
		self.nanoseconds += time.Now().Sub(self.start)
	}
	self.start = time.Time{}

	return self.nanoseconds
}

// Reset stops the timer and sets the elapsed time to zero.
func (self *Timer) Reset() {
	self.start = time.Time{}
	self.nanoseconds = 0
}

// Time returns the measured time.
func (self *Timer) Time() time.Duration {
	return self.nanoseconds
}

// Start and return a time interval.
func (self *Timer) Interval() (l time.Duration) {
	if self.start.After(time.Time{}) {
		l = time.Now().Sub(self.interval)
		self.interval = time.Now()
	}

	return
}
