package cmap

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIterateKeysWithValues(t *testing.T) {
	t.Parallel()

	cmap := NewCMap()

	for i := 1; i <= 10; i++ {
		cmap.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}

	// Testing size
	assert.Equal(t, 10, cmap.Size())
	assert.Equal(t, 10, len(cmap.Keys()))
	assert.Equal(t, 10, len(cmap.Values()))

	// Iterating Keys, checking for matching Value
	for _, key := range cmap.Keys() {
		val := strings.ReplaceAll(key, "key", "value")
		assert.Equal(t, val, cmap.Get(key))
	}

	// Test if all keys are within []Keys()
	keys := cmap.Keys()
	for i := 1; i <= 10; i++ {
		assert.Contains(t, keys, fmt.Sprintf("key%d", i), "cmap.Keys() should contain key")
	}

	// Delete 1 Key
	cmap.Delete("key1")

	assert.NotEqual(t, len(keys), len(cmap.Keys()), "[]keys and []Keys() should not be equal, they are copies, one item was removed")
}

func TestContains(t *testing.T) {
	t.Parallel()

	cmap := NewCMap()

	cmap.Set("key1", "value1")

	// Test for known values
	assert.True(t, cmap.Has("key1"))
	assert.Equal(t, "value1", cmap.Get("key1"))

	// Test for unknown values
	assert.False(t, cmap.Has("key2"))
	assert.Nil(t, cmap.Get("key2"))
}

var sink any = nil

func BenchmarkCMapConcurrentInsertsDeletesHas(b *testing.B) {
	cm := NewCMap()
	keys := make([]string, 100000)
	for i := range keys {
		keys[i] = fmt.Sprintf("key%d", i)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		semaCh := make(chan bool)
		nCPU := runtime.NumCPU()
		for j := range nCPU {
			wg.Add(1)
			go func() {
				defer wg.Done()

				// Make sure that all the goroutines run at the
				// exact same time for true concurrent tests.
				<-semaCh

				for i, key := range keys {
					if (j+i)%2 == 0 {
						cm.Has(key)
					} else {
						cm.Set(key, j)
					}
					_ = cm.Size()
					if (i+1)%3 == 0 {
						cm.Delete(key)
					}

					if (i+1)%327 == 0 {
						cm.Clear()
					}
					_ = cm.Size()
					_ = cm.Keys()
				}
				_ = cm.Values()
			}()
		}
		close(semaCh)
		wg.Wait()

		sink = semaCh
	}

	if sink == nil {
		b.Fatal("Benchmark did not run!")
	}
	sink = nil
}

func BenchmarkCMapHas(b *testing.B) {
	m := NewCMap()
	for i := range 1000 {
		m.Set(fmt.Sprint(i), i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Has(fmt.Sprint(i))
	}
}
