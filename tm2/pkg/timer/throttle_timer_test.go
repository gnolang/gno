package timer

import (
	"sync"
	"testing"
	"time"

	"github.com/jaekwon/testify/assert"
)

type thCounter struct {
	input chan struct{}
	mtx   sync.Mutex
	count int
}

func (c *thCounter) Increment() {
	c.mtx.Lock()
	c.count++
	c.mtx.Unlock()
}

func (c *thCounter) Count() int {
	c.mtx.Lock()
	val := c.count
	c.mtx.Unlock()
	return val
}

// Read should run in a go-routine and
// updates count by one every time a packet comes in
func (c *thCounter) Read() {
	for range c.input {
		c.Increment()
	}
}

func TestThrottle(t *testing.T) {
	t.Parallel()

	ms := 100
	delay := time.Duration(ms) * time.Millisecond
	longwait := time.Duration(2) * delay
	timer := NewThrottleTimer("foo", delay)

	// start at 0
	c := &thCounter{input: timer.Ch}
	assert.Equal(t, c.Count(), 0)
	go c.Read()

	// waiting does nothing
	time.Sleep(longwait)
	assert.Equal(t, c.Count(), 0)

	// send one event adds one
	timer.Set()
	time.Sleep(longwait)
	assert.Equal(t, c.Count(), 1)

	// send a burst adds one
	for i := 0; i < 5; i++ {
		timer.Set()
	}
	time.Sleep(longwait)
	assert.Equal(t, c.Count(), 2)

	// send 14, over 2 delay sections, adds 3
	short := time.Duration(ms/5) * time.Millisecond
	for i := 0; i < 14; i++ {
		timer.Set()
		time.Sleep(short)
	}
	time.Sleep(longwait)
	assert.Equal(t, c.Count(), 5)

	close(timer.Ch)
}
