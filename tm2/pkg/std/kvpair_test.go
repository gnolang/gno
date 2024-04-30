package std

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKVPairs(t *testing.T) {
	t.Parallel()

	kvs := KVPairs{
		{Key: []byte("k2"), Value: []byte("")},
		{Key: []byte("k1"), Value: []byte("2")},
		{Key: []byte("k1"), Value: []byte("1")},
		{Key: []byte("k1"), Value: []byte("2")},
	}

	// Sort() essentially tests Less() and Swap() as well
	assert.Equal(t, 4, kvs.Len())
	kvs.Sort()
	assert.Equal(t, 4, kvs.Len())

	kvs2 := KVPairs{
		{Key: []byte("k1"), Value: []byte("1")},
		{Key: []byte("k1"), Value: []byte("2")},
		{Key: []byte("k1"), Value: []byte("2")},
		{Key: []byte("k2"), Value: []byte("")},
	}
	for i := 0; i < kvs.Len(); i++ {
		assert.Equal(t, kvs[i].Key, kvs2[i].Key)
		assert.Equal(t, kvs[i].Value, kvs2[i].Value)
	}
}

func TestKI64Pairs(t *testing.T) {
	t.Parallel()

	kvs := KI64Pairs{
		{Key: []byte("k2"), Value: 0},
		{Key: []byte("k1"), Value: 2},
		{Key: []byte("k1"), Value: 1},
		{Key: []byte("k1"), Value: 2},
	}

	// Sort() essentially tests Less() and Swap() as well
	assert.Equal(t, 4, kvs.Len())
	kvs.Sort()
	assert.Equal(t, 4, kvs.Len())

	kvs2 := KI64Pairs{
		{Key: []byte("k1"), Value: 1},
		{Key: []byte("k1"), Value: 2},
		{Key: []byte("k1"), Value: 2},
		{Key: []byte("k2"), Value: 0},
	}
	for i := 0; i < kvs.Len(); i++ {
		assert.Equal(t, kvs[i].Key, kvs2[i].Key)
		assert.Equal(t, kvs[i].Value, kvs2[i].Value)
	}
}
