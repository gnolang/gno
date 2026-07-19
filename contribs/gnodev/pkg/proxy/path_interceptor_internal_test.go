package proxy

import (
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// .app/profiletx executes the tx just like .app/simulate, so it must trigger
// the same lazy package load. Without it, the first -gasprofile of a package the
// node has not loaded yet fails and still writes a profile holding only
// (ante) and (root).
func TestHandleQueryTxPathsTriggerPackageLoad(t *testing.T) {
	const pkgPath = "gno.land/r/target/foo"

	var tx std.Tx
	tx.Msgs = []std.Msg{vm.NewMsgCall(crypto.Address{}, nil, pkgPath, "Incr", nil)}
	bz, err := amino.Marshal(tx)
	require.NoError(t, err)

	for _, qpath := range []string{".app/simulate", ".app/profiletx"} {
		upaths := uniqPaths{}
		require.NoError(t, handleQuery(qpath, bz, upaths), qpath)
		assert.Equal(t, []string{pkgPath}, upaths.list(), qpath)
	}
}

func TestHandleQueryUnknownPathErrors(t *testing.T) {
	require.Error(t, handleQuery(".app/definitely-not-a-path", nil, uniqPaths{}))
}
