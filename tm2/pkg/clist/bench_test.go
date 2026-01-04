package clist

import "testing"

func BenchmarkDetaching(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping testing in short mode")
	}

	lst := New()
	for i := range b.N + 1 {
		lst.PushBack(i)
	}
	start := lst.Front()
	nxt := start.Next()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start.removed = true
		start.DetachNext()
		start.DetachPrev()
		tmp := nxt
		nxt = nxt.Next()
		start = tmp
	}
}

// This is used to benchmark the time of RMutex.
func BenchmarkRemoved(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping testing in short mode")
	}

	lst := New()
	for i := range b.N + 1 {
		lst.PushBack(i)
	}
	start := lst.Front()
	nxt := start.Next()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start.Removed()
		tmp := nxt
		nxt = nxt.Next()
		start = tmp
	}
}

func BenchmarkPushBack(b *testing.B) {
	lst := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lst.PushBack(i)
	}
}
