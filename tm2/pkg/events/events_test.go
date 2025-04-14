package events

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/random"
)

// TestAddListenerFireOnce sets up an EventSwitch, subscribes a single
// listener, and sends a string "ev".
func TestAddListenerFireOnce(t *testing.T) {
	evsw := NewEventSwitch()
	err := evsw.Start()
	require.NoError(t, err)
	defer evsw.Stop()

	messages := make(chan Event)
	evsw.AddListener("listener", func(ev Event) {
		// test there's no deadlock if we remove the listener inside a callback
		evsw.RemoveListener("listener")
		messages <- ev
	})
	go evsw.FireEvent(StringEvent("ev"))
	received := <-messages
	if received != StringEvent("ev") {
		t.Errorf("Message received does not match: %v", received)
	}
}

// TestAddListenerFireMany sets up an EventSwitch, subscribes a single
// listener, and sends a thousand integers.
func TestAddListenerFireMany(t *testing.T) {
	t.Parallel()

	evsw := NewEventSwitch()
	err := evsw.Start()
	require.NoError(t, err)
	defer evsw.Stop()

	doneSum := make(chan uint64)
	doneSending := make(chan uint64)
	numbers := make(chan uint64, 4)
	// subscribe one listener for one event
	evsw.AddListener("listener", func(ev Event) {
		numbers <- uint64(ev.(Uint64Event))
	})
	// collect received events
	go sumReceivedNumbers(numbers, doneSum)
	// go fire events
	go fireEvents(evsw, doneSending, uint64(1))
	checkSum := <-doneSending
	close(numbers)
	eventSum := <-doneSum
	if checkSum != eventSum {
		t.Errorf("Not all messages sent were received.\n")
	}
}

// TestAddListeners sets up an EventSwitch, subscribes three
// listeners, and sends a thousand integers for each.
func TestAddListeners(t *testing.T) {
	t.Parallel()

	evsw := NewEventSwitch()
	err := evsw.Start()
	require.NoError(t, err)
	defer evsw.Stop()

	doneSum := make(chan uint64)
	doneSending1 := make(chan uint64)
	doneSending2 := make(chan uint64)
	doneSending3 := make(chan uint64)
	numbers := make(chan uint64, 4)
	// subscribe one listener to three events
	evsw.AddListener("listener", func(ev Event) {
		numbers <- uint64(ev.(Uint64Event))
	})
	evsw.AddListener("listener", func(ev Event) {
		numbers <- uint64(ev.(Uint64Event))
	})
	evsw.AddListener("listener", func(ev Event) {
		numbers <- uint64(ev.(Uint64Event))
	})
	// collect received events
	go sumReceivedNumbers(numbers, doneSum)
	// go fire events
	go fireEvents(evsw, doneSending1, uint64(1))
	go fireEvents(evsw, doneSending2, uint64(1))
	go fireEvents(evsw, doneSending3, uint64(1))
	var checkSum uint64 = 0
	checkSum += <-doneSending1
	checkSum += <-doneSending2
	checkSum += <-doneSending3
	close(numbers)
	eventSum := <-doneSum
	if checkSum*3 != eventSum {
		t.Errorf("Not all messages sent were received.\n")
	}
}

func TestAddAndRemoveListenerConcurrency(t *testing.T) {
	t.Parallel()

	var (
		stopInputEvent = false
		roundCount     = 2000
	)

	evsw := NewEventSwitch()
	err := evsw.Start()
	require.NoError(t, err)
	defer evsw.Stop()

	done1 := make(chan struct{})
	done2 := make(chan struct{})

	// Must be executed concurrently to uncover the ev race.
	// 1. RemoveListener
	go func() {
		for range roundCount {
			evsw.RemoveListener("listener")
		}
		close(done1)
	}()

	// 2. AddListener
	go func() {
		for i := range roundCount {
			index := i
			evsw.AddListener("listener", func(ev Event) {
				t.Errorf("should not run callback for %d.\n", index)
				stopInputEvent = true
			})
		}
		close(done2)
	}()

	<-done1
	<-done2

	evsw.RemoveListener("listener") // remove the last listener

	for i := 0; i < roundCount && !stopInputEvent; i++ {
		evsw.FireEvent(Uint64Event(uint64(1001)))
	}
}

func TestAddAndRemoveListener(t *testing.T) {
	t.Parallel()

	evsw := NewEventSwitch()
	err := evsw.Start()
	require.NoError(t, err)
	defer evsw.Stop()

	doneSum1 := make(chan uint64)
	doneSum2 := make(chan uint64)
	doneSending1 := make(chan uint64)
	doneSending2 := make(chan uint64)
	numbers1 := make(chan uint64, 4)
	numbers2 := make(chan uint64, 4)
	// subscribe two listener to three events
	evsw.AddListener("listener", func(ev Event) {
		if uint64(ev.(Uint64Event)) <= 1000 {
			numbers1 <- uint64(ev.(Uint64Event))
		}
	})
	evsw.AddListener("listener", func(ev Event) {
		if uint64(ev.(Uint64Event)) > 1000 {
			numbers2 <- uint64(ev.(Uint64Event))
		}
	})
	// collect received events for event1
	go sumReceivedNumbers(numbers1, doneSum1)
	// collect received events for event2
	go sumReceivedNumbers(numbers2, doneSum2)
	// go fire events
	go fireEvents(evsw, doneSending1, uint64(1)) // to numbers1.
	checkSumEvent1 := <-doneSending1
	// after sending all event1, unsubscribe for all events
	evsw.RemoveListener("listener")
	go fireEvents(evsw, doneSending2, uint64(1001)) // would be to numbers2.
	checkSumEvent2 := <-doneSending2
	close(numbers1)
	close(numbers2)
	eventSum1 := <-doneSum1
	eventSum2 := <-doneSum2
	if checkSumEvent1 != eventSum1 ||
		// correct value asserted by preceding tests, suffices to be non-zero
		checkSumEvent2 == uint64(0) ||
		eventSum2 != uint64(0) {
		t.Errorf("Not all messages sent were received or unsubscription did not register.\n")
	}
}

// TestRemoveListener does basic tests on adding and removing
func TestRemoveListener(t *testing.T) {
	t.Parallel()

	evsw := NewEventSwitch()
	err := evsw.Start()
	require.NoError(t, err)
	defer evsw.Stop()

	count := 10
	sum1, sum2 := 0, 0
	// add some listeners and make sure they work
	evsw.AddListener("BEEP-listener", func(ev Event) {
		if string(ev.(StringEvent)) == "BEEP" {
			sum1++
		}
	})
	evsw.AddListener("boop-listener", func(ev Event) {
		if string(ev.(StringEvent)) == "boop" {
			sum2++
		}
	})
	for range count {
		evsw.FireEvent(StringEvent("BEEP"))
		evsw.FireEvent(StringEvent("boop"))
	}
	assert.Equal(t, count, sum1)
	assert.Equal(t, count, sum2)

	// remove one by event and make sure it is gone
	evsw.RemoveListener("boop-listener")
	for range count {
		evsw.FireEvent(StringEvent("BEEP"))
		evsw.FireEvent(StringEvent("boop"))
	}
	assert.Equal(t, count*2, sum1)
	assert.Equal(t, count, sum2)

	// remove the listener entirely and make sure both gone
	evsw.RemoveListener("BEEP-listener")
	for range count {
		evsw.FireEvent(StringEvent("BEEP"))
		evsw.FireEvent(StringEvent("boop"))
	}
	assert.Equal(t, count*2, sum1)
	assert.Equal(t, count, sum2)
}

// More precisely it randomly subscribes new listeners.
// At the same time it starts randomly unsubscribing these additional listeners.
// NOTE: it is important to run this test with race conditions tracking on,
// `go test -race`, to examine for possible race conditions.
func TestRemoveListenersAsync(t *testing.T) {
	t.Parallel()

	evsw := NewEventSwitch()
	err := evsw.Start()
	require.NoError(t, err)
	defer evsw.Stop()

	doneSum1 := make(chan uint64)
	doneSum2 := make(chan uint64)
	doneSending1 := make(chan uint64)
	doneSending2 := make(chan uint64)
	doneSending3 := make(chan uint64)
	numbers1 := make(chan uint64, 4)
	numbers2 := make(chan uint64, 4)
	// subscribe two listener to three events
	evsw.AddListener("listener1", func(ev Event) {
		evi := uint64(ev.(Uint64Event))
		if 1 <= evi && evi <= 1000 {
			numbers1 <- uint64(ev.(Uint64Event))
		}
	})
	evsw.AddListener("listener1", func(ev Event) {
		evi := uint64(ev.(Uint64Event))
		if 1001 <= evi && evi <= 2000 {
			numbers1 <- uint64(ev.(Uint64Event))
		}
	})
	evsw.AddListener("listener1", func(ev Event) {
		evi := uint64(ev.(Uint64Event))
		if 2001 <= evi && evi <= 3000 {
			numbers1 <- uint64(ev.(Uint64Event))
		}
	})
	evsw.AddListener("listener2", func(ev Event) {
		evi := uint64(ev.(Uint64Event))
		if 1 <= evi && evi <= 1000 {
			numbers2 <- uint64(ev.(Uint64Event))
		}
	})
	evsw.AddListener("listener2", func(ev Event) {
		evi := uint64(ev.(Uint64Event))
		if 1001 <= evi && evi <= 2000 {
			numbers2 <- uint64(ev.(Uint64Event))
		}
	})
	evsw.AddListener("listener2", func(ev Event) {
		evi := uint64(ev.(Uint64Event))
		if 2001 <= evi && evi <= 3000 {
			numbers2 <- uint64(ev.(Uint64Event))
		}
	})
	// collect received events for event1
	go sumReceivedNumbers(numbers1, doneSum1)
	// collect received events for event2
	go sumReceivedNumbers(numbers2, doneSum2)
	addListenersStress := func() {
		r1 := random.NewRand()
		r1.Seed(time.Now().UnixNano())
		for k := uint16(0); k < 400; k++ {
			listenerNumber := r1.Intn(100) + 3
			go evsw.AddListener(fmt.Sprintf("listener%v", listenerNumber),
				func(_ Event) {})
		}
	}
	removeListenersStress := func() {
		r2 := random.NewRand()
		r2.Seed(time.Now().UnixNano())
		for k := uint16(0); k < 80; k++ {
			listenerNumber := r2.Intn(100) + 3
			go evsw.RemoveListener(fmt.Sprintf("listener%v", listenerNumber))
		}
	}
	addListenersStress()
	// go fire events
	go fireEvents(evsw, doneSending1, uint64(1))
	removeListenersStress()
	go fireEvents(evsw, doneSending2, uint64(1001))
	go fireEvents(evsw, doneSending3, uint64(2001))
	checkSumEvent1 := <-doneSending1
	checkSumEvent2 := <-doneSending2
	checkSumEvent3 := <-doneSending3
	checkSum := checkSumEvent1 + checkSumEvent2 + checkSumEvent3
	close(numbers1)
	close(numbers2)
	eventSum1 := <-doneSum1
	eventSum2 := <-doneSum2
	if checkSum != eventSum1 ||
		checkSum != eventSum2 {
		t.Errorf("Not all messages sent were received.\n")
	}
}

// ------------------------------------------------------------------------------
// Helper functions

// sumReceivedNumbers takes two channels and adds all numbers received
// until the receiving channel `numbers` is closed; it then sends the sum
// on `doneSum` and closes that channel.  Expected to be run in a go-routine.
func sumReceivedNumbers(numbers, doneSum chan uint64) {
	var sum uint64
	for {
		j, more := <-numbers
		sum += j
		if !more {
			doneSum <- sum
			close(doneSum)
			return
		}
	}
}

// fireEvents takes an EventSwitch and fires a thousand integers with the
// integers mootonically increasing from `offset` to `offset` + 999.  It
// additionally returns the addition of all integers sent on `doneChan` for
// assertion that all events have been sent, and enabling the test to assert
// all events have also been received.
func fireEvents(evsw EventSwitch, doneChan chan uint64,
	offset uint64,
) {
	var sentSum uint64
	for i := offset; i <= offset+uint64(999); i++ {
		sentSum += i
		evsw.FireEvent(Uint64Event(i))
	}
	doneChan <- sentSum
	close(doneChan)
}

type Uint64Event uint64

func (Uint64Event) AssertEvent() {}

type StringEvent string

func (StringEvent) AssertEvent() {}
