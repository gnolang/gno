package cache_test

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/iavl/cache"
)

// expectedResult represents the expected result of each add/remove operation.
// It can be noneRemoved or the index of the removed node in testNodes
type expectedResult int

const (
	noneRemoved expectedResult = -1
	// The rest represent the index of the removed node
)

// testNode is the node used for testing cache implementation
type testNode struct {
	key []byte
}

type cacheOp struct {
	testNodexIdx int
}

type testcase struct {
	setup               func(cache.Cache)
	cacheMax            int
	cacheOps            []cacheOp
	expectedNodeIndexes []int // contents of the cache once test case completes represent by indexes in testNodes
}

func (tn *testNode) GetKey() []byte {
	return tn.key
}

const (
	testKey = "key"
)

var _ cache.Node = (*testNode)(nil)

var testNodes = []cache.Node{
	&testNode{
		key: []byte(fmt.Sprintf("%s%d", testKey, 1)),
	},
	&testNode{
		key: []byte(fmt.Sprintf("%s%d", testKey, 2)),
	},
	&testNode{
		key: []byte(fmt.Sprintf("%s%d", testKey, 3)),
	},
}

func Test_Cache_Add(t *testing.T) {
	testcases := map[string]testcase{
		"add 1 node with 1 max - added": {
			cacheMax: 1,
			cacheOps: []cacheOp{
				{
					testNodexIdx: 0,
				},
			},
			expectedNodeIndexes: []int{0},
		},
		"add 3 nodes with 1 max - first 2 removed": {
			cacheMax: 1,
			cacheOps: []cacheOp{
				{
					testNodexIdx: 0,
				},
				{
					testNodexIdx: 1,
				},
				{
					testNodexIdx: 2,
				},
			},
			expectedNodeIndexes: []int{2},
		},
		"add 3 nodes with 2 max - first removed": {
			cacheMax: 2,
			cacheOps: []cacheOp{
				{
					testNodexIdx: 0,
				},
				{
					testNodexIdx: 1,
				},
				{
					testNodexIdx: 2,
				},
			},
			expectedNodeIndexes: []int{1, 2},
		},
		"add 3 nodes with 10 max - non removed": {
			cacheMax: 10,
			cacheOps: []cacheOp{
				{
					testNodexIdx: 0,
				},
				{
					testNodexIdx: 1,
				},
				{
					testNodexIdx: 2,
				},
			},
			expectedNodeIndexes: []int{0, 1, 2},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			cache := cache.New(tc.cacheMax)

			expectedCurSize := 0

			for _, op := range tc.cacheOps {
				evicted := cache.Add(testNodes[op.testNodexIdx])

				if !evicted {
					expectedCurSize++
				}
				require.Equal(t, expectedCurSize, cache.Len())
			}

			validateCacheContentsAfterTest(t, tc, cache)
		})
	}
}

func Test_Cache_Remove(t *testing.T) {
	testcases := map[string]testcase{
		"remove non-existent key, cache max 0 - nil returned": {
			cacheMax: 0,
			cacheOps: []cacheOp{
				{
					testNodexIdx: 0,
				},
			},
		},
		"remove non-existent key, cache max 1 - nil returned": {
			setup: func(c cache.Cache) {
				c.Add(testNodes[1])
				require.Equal(t, 1, c.Len())
			},
			cacheMax: 1,
			cacheOps: []cacheOp{
				{
					testNodexIdx: 0,
				},
			},
			expectedNodeIndexes: []int{1},
		},
		"remove existent key, cache max 1 - removed": {
			setup: func(c cache.Cache) {
				c.Add(testNodes[0])
				require.Equal(t, 1, c.Len())
			},
			cacheMax: 1,
			cacheOps: []cacheOp{
				{
					testNodexIdx: 0,
				},
			},
		},
		"remove twice, cache max 1 - removed first time, then nil": {
			setup: func(c cache.Cache) {
				c.Add(testNodes[0])
				require.Equal(t, 1, c.Len())
			},
			cacheMax: 1,
			cacheOps: []cacheOp{
				{
					testNodexIdx: 0,
				},
				{
					testNodexIdx: 0,
				},
			},
		},
		"remove all, cache max 3": {
			setup: func(c cache.Cache) {
				c.Add(testNodes[0])
				c.Add(testNodes[1])
				c.Add(testNodes[2])
				require.Equal(t, 3, c.Len())
			},
			cacheMax: 3,
			cacheOps: []cacheOp{
				{
					testNodexIdx: 2,
				},
				{
					testNodexIdx: 0,
				},
				{
					testNodexIdx: 1,
				},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			cache := cache.New(tc.cacheMax)

			if tc.setup != nil {
				tc.setup(cache)
			}

			expectedCurSize := cache.Len()

			for _, op := range tc.cacheOps {
				present := cache.Remove(testNodes[op.testNodexIdx].GetKey())

				if present {
					expectedCurSize--
				}
				require.Equal(t, expectedCurSize, cache.Len())
			}

			validateCacheContentsAfterTest(t, tc, cache)
		})
	}
}

func validateCacheContentsAfterTest(t *testing.T, tc testcase, cache cache.Cache) { //nolint: thelper
	require.Equal(t, len(tc.expectedNodeIndexes), cache.Len())
	for _, idx := range tc.expectedNodeIndexes {
		expectedNode := testNodes[idx]
		require.True(t, cache.Has(expectedNode.GetKey()))
		require.Equal(t, expectedNode, cache.Get(expectedNode.GetKey()))
	}
}

func randBytes(length int) []byte {
	key := make([]byte, length)
	// math.rand.Read always returns err=nil
	// we do not need cryptographic randomness for this test:

	rand.Read(key) //nolint:errcheck
	return key
}
