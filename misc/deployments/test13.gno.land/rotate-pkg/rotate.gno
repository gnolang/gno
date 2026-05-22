// Package rotate is a single-use realm shipped at test-13 genesis whose
// sole purpose is to swap the gnoland1 sole T1 (`oldT1`) for the test-13
// sole T1 (`newT1`) after historical replay completes.
//
// The proposal-flow rotation pattern (manfred proposes self-withdraw →
// new-T1 votes + executes) is unworkable at genesis-mode replay because
// gov/dao exposes no API to discover the latest proposal ID across MsgRun
// txs — and the post-replay proposal counter depends on how many
// proposals manfred opened on gnoland1, which we can't determine
// statically without re-executing the replay. This realm sidesteps the
// problem entirely: a single genesis-mode MsgCall to Rotate does the
// transfer through direct memberstore writes, no proposals involved.
//
// Phase-1 bootstrap (`govdao_prop1_test13.gno`) wires this package into
// `dao.UpdateImpl(...AllowedDAOs)` at lock time, so memberstore.Get()
// works from inside Rotate. Once Rotate runs it self-ejects from
// AllowedDAOs, leaving the membership-mutation surface back at
// `gno.land/r/gov/dao/v3/impl` alone.
//
// Single-use is enforced by the `runtime.ChainHeight() != 0` gate.
// Migration txs run with BlockHeight=0 (forced by
// `gnogenesis fork generate`'s readMigrationTxs), so during genesis
// replay ChainHeight()==0 and Rotate proceeds; post-genesis any caller
// gets a panic.
package rotate

import (
	"chain/runtime"

	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/memberstore"
)

const (
	newT1 = address("g1aeddlftlfk27ret5rf750d7w5dume3kcsm8r8m")
	oldT1 = address("g1manfred47kzduec920z88wfr64ylksmdcedlf5")
)

func Rotate(cur realm) {
	if runtime.ChainHeight() != 0 {
		panic("rotate is genesis-only")
	}

	ms := memberstore.Get(0, cur)
	tier, ok := memberstore.GetTier(memberstore.T1)
	if !ok {
		panic("T1 tier not found")
	}
	if err := ms.SetMember(memberstore.T1, newT1, memberstore.NewMember(tier.InvitationPoints)); err != nil {
		panic(err.Error())
	}
	ms.RemoveMember(oldT1)

	// Pass nil DAO: bootstrap already installed impl.NewGovDAO() into dao.dao;
	// rotate only narrows AllowedDAOs (removing itself) — DAO ref stays intact.
	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(nil, []string{"gno.land/r/gov/dao/v3/impl"}))
}
