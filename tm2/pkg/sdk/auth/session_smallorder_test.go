package auth

import (
	stded25519 "crypto/ed25519"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	tu "github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// TestSmallOrderSessionKey covers delegated sessions. A session key is not
// bound to a fresh address: the session account signs for its master's
// messages. A master that installs a small-order session key (here the
// all-zero encoding, as a dApp with an uninitialized key would) must not hand
// its authority to whoever forges an S = 0 signature.
func TestSmallOrderSessionKey(t *testing.T) {
	env, anteHandler, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	var zero ed25519.PubKeyEd25519 // order 4

	// Refused at creation.
	err := MsgCreateSession{
		Creator:    masterAddr,
		SessionKey: zero,
		ExpiresAt:  ctx.BlockTime().Unix() + 3600,
		SpendLimit: sessionSpendLimit(),
	}.ValidateBasic()
	require.ErrorIs(t, err, std.InvalidPubKeyError{})

	// A session installed without that check (as one created before it
	// would be) cannot be signed for: the VerifyBytes gate rejects the forgery.
	sa := createSessionDirect(t, env, masterAddr, zero, ctx.BlockTime().Unix()+3600)

	msgs := []std.Msg{tu.NewTestMsg(masterAddr)}
	fee := tu.NewTestFee()
	// S = 0 and R one of the four points of <A>. k hashes R, so vary the memo
	// until one R matches, as an attacker would.
	for i := range 64 {
		memo := fmt.Sprintf("memo %d", i)
		signBytes, err := std.GetSignaturePayload(std.SignDoc{
			ChainID: ctx.ChainID(), AccountNumber: sa.GetAccountNumber(),
			Sequence: sa.GetSequence(), Fee: fee, Msgs: msgs, Memo: memo,
		})
		require.NoError(t, err)
		for _, r := range []string{
			"0100000000000000000000000000000000000000000000000000000000000000",
			"0000000000000000000000000000000000000000000000000000000000000000",
			"0000000000000000000000000000000000000000000000000000000000000080",
			"ecffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff7f",
		} {
			sig := make([]byte, 64)
			_, err := hex.Decode(sig, []byte(r))
			require.NoError(t, err)
			if !stded25519.Verify(zero[:], signBytes, sig) {
				continue
			}
			tx := std.NewTx(msgs, fee, []std.Signature{{PubKey: zero, SessionAddr: zero.Address(), Signature: sig}}, memo)
			checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnauthorizedError{})
			return
		}
	}
	t.Fatal("premise: crypto/ed25519 accepts no S = 0 forgery under the zero key")
}
