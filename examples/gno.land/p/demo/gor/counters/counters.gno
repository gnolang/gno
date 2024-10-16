package counters

type Counter struct {
	value uint64
}

func NewCounter() *Counter {
	return &Counter{}
}

func (counter *Counter) Current() uint64 {
	return counter.value
}

func (counter *Counter) Increment() {
	counter.value += 1
}

func (counter *Counter) Decrement() {
	value := counter.value
	if value <= 0 {
		panic("Counter: decrement overflow")
	}
	counter.value = value - 1
}

func (counter *Counter) Reset() {
	counter.value = 0
}
