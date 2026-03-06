package merkle

// NOTE(tb): Adapted from cosmos-sdk/store/internal/proofs/convert_test.go

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/random"
)

func TestLeafOp(t *testing.T) {
	_, proofs, keys := SimpleProofsFromMap(buildMap(20))
	// pick a random key in the middle
	key := keys[random.RandInt()%(len(keys)-2)+1]
	proof := proofs[key]

	converted, err := ConvertExistenceProof(proof, []byte(key), toValue(key))
	if err != nil {
		t.Fatal(err)
	}

	leaf := converted.GetLeaf()
	if leaf == nil {
		t.Fatalf("Missing leaf node")
	}

	hash, err := leaf.Apply(converted.Key, converted.Value)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(hash, proof.LeafHash) {
		t.Errorf("Calculated: %X\nExpected:   %X", hash, proof.LeafHash)
	}
}

func TestBuildPath(t *testing.T) {
	cases := map[string]struct {
		idx      int
		total    int
		expected []bool
	}{
		"pair left": {
			idx:      0,
			total:    2,
			expected: []bool{true},
		},
		"pair right": {
			idx:      1,
			total:    2,
			expected: []bool{false},
		},
		"power of 2": {
			idx:      3,
			total:    8,
			expected: []bool{false, false, true},
		},
		"size of 7 right most": {
			idx:      6,
			total:    7,
			expected: []bool{false, false},
		},
		"size of 6 right-left (from top)": {
			idx:      4,
			total:    6,
			expected: []bool{true, false},
		},
		"size of 6 left-right-left (from top)": {
			idx:      2,
			total:    7,
			expected: []bool{true, false, true},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			path := buildPath(tc.idx, tc.total)
			if len(path) != len(tc.expected) {
				t.Fatalf("Got %v\nExpected %v", path, tc.expected)
			}
			for i := range path {
				if path[i] != tc.expected[i] {
					t.Fatalf("Differ at %d\nGot %v\nExpected %v", i, path, tc.expected)
				}
			}
		})
	}
}

func TestConvertProof(t *testing.T) {
	for i := 0; i < 100; i++ {
		t.Run(fmt.Sprintf("Run %d", i), func(t *testing.T) {
			root, proofs, keys := SimpleProofsFromMap(buildMap(157))
			// take first key
			key := keys[0]
			proof := proofs[key]

			converted, err := ConvertExistenceProof(proof, []byte(key), toValue(key))
			if err != nil {
				t.Fatal(err)
			}

			calc, err := converted.Calculate()
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(calc, root) {
				t.Errorf("Calculated: %X\nExpected:   %X", calc, root)
			}
		})
	}
}

// buildMap creates random key/values and stores in a map,
// returns a list of all keys in sorted order
func buildMap(size int) map[string][]byte {
	data := make(map[string][]byte)
	// insert lots of info and store the bytes
	for i := 0; i < size; i++ {
		key := random.RandStr(20)
		data[key] = toValue(key)
	}
	return data
}

func toValue(key string) []byte {
	return []byte("value_for_" + key)
}
