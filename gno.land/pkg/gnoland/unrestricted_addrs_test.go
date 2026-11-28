package gnoland

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	storebptree "github.com/gnolang/gno/tm2/pkg/store/bptree"
)

// Governance can hand out and take back the token-lock whitelist by writing
// unrestricted_addrs. bank.canSendCoins reads the resulting bit to decide who
// may move a restricted denom, so the revoke half is the one with teeth.
//
// Tested here rather than in auth, where the change is implemented: it skips
// any account that is not an AccountUnrestricter, and std.BaseAccount -- what
// auth's own tests are built on -- is not one. A test written there would pass
// over every address and prove nothing. *GnoAccount is the type that carries
// the bit in production.
func TestUnrestrictedAddrsParamGrantsAndRevokes(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	mainKey := store.NewStoreKey("mainKey")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(mainKey, storebptree.FastStoreConstructor, db)
	require.NoError(t, ms.LoadLatestVersion())
	ctx := sdk.NewContext(sdk.RunTxModeDeliver, ms.MultiCacheWrap(), &bft.Header{ChainID: "test"}, log.NewNoopLogger())

	prmk := params.NewParamsKeeper(mainKey)
	acck := auth.NewAccountKeeper(mainKey, prmk.ForModule(auth.ModuleName), ProtoGnoAccount, std.ProtoBaseSessionAccount)
	bankk := bank.NewBankKeeper(acck, prmk.ForModule(bank.ModuleName), mainKey, []string{ugnot.Denom})
	prmk.Register(auth.ModuleName, acck)
	prmk.Register(bank.ModuleName, bankk)

	revoked := crypto.AddressFromPreimage([]byte("was-unrestricted"))
	granted := crypto.AddressFromPreimage([]byte("newly-unrestricted"))
	for _, addr := range []crypto.Address{revoked, granted} {
		acck.SetAccount(ctx, acck.NewAccountWithAddress(ctx, addr))
	}

	whitelisted := func(t *testing.T, addr crypto.Address) bool {
		t.Helper()
		acc := acck.GetAccount(ctx, addr)
		require.NotNil(t, acc)
		u, ok := acc.(std.AccountUnrestricter)
		require.True(t, ok, "the account must carry the token-lock attributes")
		return u.IsTokenLockWhitelisted()
	}

	// Start with one address already holding the privilege, which is the state
	// the params say is current.
	acc := acck.GetAccount(ctx, revoked)
	acc.(std.AccountUnrestricter).SetTokenLockWhitelisted(true)
	acck.SetAccount(ctx, acc)
	require.True(t, whitelisted(t, revoked))
	require.False(t, whitelisted(t, granted))

	p := auth.DefaultParams()
	p.UnrestrictedAddrs = []crypto.Address{revoked}
	// Both are needed: WillSetParam re-validates the struct it reads from the
	// store, and the change itself takes the previous set from the context.
	require.NoError(t, acck.SetParams(ctx, p))
	ctx = ctx.WithValue(auth.AuthParamsContextKey{}, p)

	// The write governance makes: replace the set with a different address.
	acck.WillSetParam(ctx, "p:unrestricted_addrs", []string{granted.String()})

	assert.False(t, whitelisted(t, revoked),
		"an address dropped from unrestricted_addrs must lose the whitelist bit")
	assert.True(t, whitelisted(t, granted),
		"an address added to unrestricted_addrs must gain the whitelist bit")
}
