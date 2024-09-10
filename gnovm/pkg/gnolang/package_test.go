package gnolang

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
)

// Tries to reproduce the bug #1036 on all registered types
func TestAminoJSONRegisteredTypes(t *testing.T) {
	t.Parallel()

	for _, typ := range Package.Types {
		// Instantiate registered type
		x := reflect.New(typ.Type).Interface()

		// Call MarshalAmino directly on 'x'
		_, err := amino.MarshalJSON(x)
		require.NoError(t, err, "marshal type %s", typ.Type.Name())

		// Call MarshalAmino on a struct that embeds 'x' in a field of type any
		xx := struct {
			X any
		}{X: x}
		_, err = amino.MarshalJSON(xx)
		require.NoError(t, err, "marshal type %s from struct", typ.Type.Name())
	}

	// Check unmarshaling (can't reuse package.Types because some internal values
	// must be filled properly
	bi := BigintValue{V: big.NewInt(1)}
	bz := amino.MustMarshalJSON(bi)
	amino.MustUnmarshalJSON(bz, &bi)
	// Check unmarshaling with an embedding struct
	x := struct{ X any }{X: bi}
	bz = amino.MustMarshalJSON(x)
	amino.MustUnmarshalJSON(bz, &x)
}
