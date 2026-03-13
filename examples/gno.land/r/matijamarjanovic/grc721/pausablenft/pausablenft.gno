package pausablenft

import (
	"std"
	"strconv"

	"gno.land/p/demo/grc/grc721remake"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/authz"
)

var (
	NFT, ledger = grc721remake.NewNFT("PausableNFT", "PNFT")
	Teller      = NFT.CallerTeller()
	auth        = authz.New()
)

func init() {
	pausableHook := &PausableHook{paused: false}
	ledger.RegisterHook("pausable", pausableHook)
}

// Hook Related Functions

// SetPaused sets the paused state of the NFT collection
func SetPaused(paused bool) error {
	return auth.Do("set_paused", func() error {
		hook, err := ledger.GetHook("pausable")
		if err != nil {
			return err
		}

		pausableHook := hook.(*PausableHook)
		pausableHook.paused = paused

		std.Emit(
			"SetPaused",
			"paused", strconv.FormatBool(paused),
		)
		return nil
	})
}

// IsPaused returns whether the NFT collection is paused
func IsPaused() bool {
	hook, err := ledger.GetHook("pausable")
	if err != nil {
		panic(err)
	}

	pausableHook := hook.(*PausableHook)
	return pausableHook.paused
}

// Admin Functions (authz)

func MintToken(to std.Address, tid grc721remake.TokenID) error {
	return auth.Do("mint", func() error {
		return ledger.Mint(to, tid)
	})
}

func BatchMintTokens(to std.Address, tids []grc721remake.TokenID) error {
	return auth.Do("batch_mint", func() error {
		return ledger.BatchMint(to, tids)
	})
}

func BurnToken(tid grc721remake.TokenID) error {
	return auth.Do("burn", func() error {
		return ledger.Burn(tid)
	})
}

func TransferAuthority(newAuthority authz.Authority) error {
	return auth.Transfer(newAuthority)
}

func AddAdmin(addr std.Address) error {
	if memberAuth, ok := auth.Current().(*authz.MemberAuthority); ok {
		return memberAuth.AddMember(addr)
	}
	return ufmt.Errorf("current authority is not a member authority")
}

func RemoveAdmin(addr std.Address) error {
	if memberAuth, ok := auth.Current().(*authz.MemberAuthority); ok {
		return memberAuth.RemoveMember(addr)
	}
	return ufmt.Errorf("current authority is not a member authority")
}
