package scheduler

import "time"

type Clock interface {
	Now() time.Time
	Sleep(d time.Duration)
	After(d time.Duration) <-chan time.Time
}

type RealClock struct{}

func (c *RealClock) Now() time.Time {
	return time.Now().UTC()
}

func (c *RealClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

func (c *RealClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}