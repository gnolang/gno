package merkle

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
)

// proofNodeChain returns n nodes, each the Parent of the next.
func proofNodeChain(n int) *SimpleProofNode {
	var node *SimpleProofNode
	for i := range n {
		node = &SimpleProofNode{Hash: []byte{byte(i)}, Parent: node}
	}
	return node
}

// amino's maxAnyDepth: each SimpleProofNode level counts against it.
const nestCeiling = 64

func TestSimpleProofNodeDecodeDepth(t *testing.T) {
	t.Parallel()

	reflectCdc := amino.NewCodec()
	reflectCdc.RegisterPackage(Package)

	paths := []struct {
		name      string
		marshal   func(any) ([]byte, error)
		unmarshal func([]byte, any) error
	}{
		{"generated", amino.Marshal, amino.Unmarshal},
		{"reflect", amino.Marshal, reflectCdc.UnmarshalReflect},
		{"json", amino.MarshalJSON, amino.UnmarshalJSON},
	}
	for _, p := range paths {
		t.Run(p.name, func(t *testing.T) {
			t.Parallel()

			bz, err := p.marshal(proofNodeChain(nestCeiling))
			require.NoError(t, err)
			var got SimpleProofNode
			require.NoError(t, p.unmarshal(bz, &got))
			assert.Equal(t, *proofNodeChain(nestCeiling), got)

			bz, err = p.marshal(proofNodeChain(nestCeiling + 1))
			require.NoError(t, err)
			require.ErrorContains(t, p.unmarshal(bz, &got), "exceeded max nesting depth 64")
		})
	}
}

// Only a type that can contain itself is counted, so decoding the other
// merkle types is unchanged.
func TestMerkleRecursiveTypes(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		v    any
		want bool
	}{
		{SimpleProofNode{}, true},
		{SimpleProof{}, false},
		{Proof{}, false},
		{ProofOp{}, false},
	} {
		info, err := amino.GetTypeInfo(reflect.TypeOf(tc.v))
		require.NoError(t, err)
		assert.Equal(t, tc.want, info.Recursive, "%T", tc.v)
	}
}
