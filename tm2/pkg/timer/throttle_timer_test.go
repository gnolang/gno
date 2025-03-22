package timer

import (
	"sync"
	"testing"
	"time"

	asrt "github.com/stretchr/testify/assert"
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

func TestThrottle(test *testing.T) {
	test.Parallel()

	assert := asrt.New(test)

	ms := 100
	delay := time.Duration(ms) * time.Millisecond
	longwait := time.Duration(2) * delay
	t := NewThrottleTimer("foo", delay)

	// start at 0
	c := &thCounter{input: t.Ch}
	assert.Equal(0, c.Count())
	go c.Read()

	// waiting does nothing
	time.Sleep(longwait)
	assert.Equal(0, c.Count())

	// send one event adds one
	t.Set()
	time.Sleep(longwait)
	assert.Equal(1, c.Count())

	// send a burst adds one
	for range 5 {
		t.Set()
	}
	time.Sleep(longwait)
	assert.Equal(2, c.Count())

	// send 14, over 2 delay sections, adds 3
	short := time.Duration(ms/5) * time.Millisecond
	for range 14 {
		t.Set()
		time.Sleep(short)
	}
	time.Sleep(longwait)
	assert.Equal(5, c.Count())

	close(t.Ch)
}
