package bptree

// Ported from tm2/pkg/iavl/proof_ics23_test.go

import (
	"math/rand"
	"sort"
	"testing"

	ics23 "github.com/cosmos/ics23/go"
	"github.com/stretchr/testify/require"
)

type Where int

const (
	Left   Where = iota
	Middle
	Right
)

func buildTree(size int, seed int64) (*MutableTree, [][]byte, error) {
	tree := getTestTree(0)
	r := rand.New(rand.NewSource(seed))

	keys := make([][]byte, size)
	for i := 0; i < size; i++ {
		key := make([]byte, 20)
		r.Read(key)
		val := make([]byte, 20)
		r.Read(val)
		tree.Set(key, val)
		keys[i] = key
	}
	sort.Slice(keys, func(i, j int) bool {
		for k := range keys[i] {
			if k >= len(keys[j]) {
				return false
			}
			if keys[i][k] != keys[j][k] {
				return keys[i][k] < keys[j][k]
			}
		}
		return len(keys[i]) < len(keys[j])
	})
	// Deduplicate
	unique := keys[:0]
	for i, k := range keys {
		if i == 0 || string(k) != string(keys[i-1]) {
			unique = append(unique, k)
		}
	}
	tree.SaveVersion()
	return tree, unique, nil
}

func getKey(allkeys [][]byte, loc Where) []byte {
	switch loc {
	case Left:
		return allkeys[0]
	case Right:
		return allkeys[len(allkeys)-1]
	case Middle:
		return allkeys[len(allkeys)/2]
	default:
		panic("bad location")
	}
}

func getNonKey(allkeys [][]byte, loc Where) []byte {
	switch loc {
	case Left:
		// Key before the first key
		k := make([]byte, len(allkeys[0]))
		copy(k, allkeys[0])
		if k[len(k)-1] > 0 {
			k[len(k)-1]--
		} else {
			k = append(k, 0xFF)
		}
		return k
	case Right:
		// Key after the last key
		k := make([]byte, len(allkeys[len(allkeys)-1]))
		copy(k, allkeys[len(allkeys)-1])
		k = append(k, 0x01)
		return k
	case Middle:
		// Key between two middle keys
		mid := len(allkeys) / 2
		k := make([]byte, len(allkeys[mid]))
		copy(k, allkeys[mid])
		k = append(k, 0x01) // slightly after the middle key
		return k
	default:
		panic("bad location")
	}
}

func TestGetMembership(t *testing.T) {
	cases := map[string]struct {
		size int
		loc  Where
	}{
		"small left":   {size: 100, loc: Left},
		"small middle": {size: 100, loc: Middle},
		"small right":  {size: 100, loc: Right},
		"big left":     {size: 5431, loc: Left},
		"big middle":   {size: 5431, loc: Middle},
		"big right":    {size: 5431, loc: Right},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tree, allkeys, err := buildTree(tc.size, 0)
			require.NoError(t, err)

			key := getKey(allkeys, tc.loc)
			val, err := tree.Get(key)
			require.NoError(t, err)
			proof, err := tree.GetMembershipProof(key)
			require.NoError(t, err)

			root := tree.Hash()
			valid := ics23.VerifyMembership(BptreeSpec, root, proof, key, val)
			require.True(t, valid, "Membership Proof Invalid")
		})
	}
}

func TestGetNonMembership(t *testing.T) {
	cases := map[string]struct {
		size int
		loc  Where
	}{
		"small left":   {size: 100, loc: Left},
		"small middle": {size: 100, loc: Middle},
		"small right":  {size: 100, loc: Right},
		"big left":     {size: 5431, loc: Left},
		"big middle":   {size: 5431, loc: Middle},
		"big right":    {size: 5431, loc: Right},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tree, allkeys, err := buildTree(tc.size, 0)
			require.NoError(t, err)

			key := getNonKey(allkeys, tc.loc)
			// Verify key truly doesn't exist
			has, _ := tree.Has(key)
			require.False(t, has, "key should not exist")

			proof, err := tree.GetNonMembershipProof(key)
			require.NoError(t, err)

			root := tree.Hash()
			valid := ics23.VerifyNonMembership(BptreeSpec, root, proof, key)
			require.True(t, valid, "Non-Membership Proof Invalid")
		})
	}
}
