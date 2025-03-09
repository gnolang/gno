package valopers_proposal

import (
  "std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
  "gno.land/r/gnoland/valopers"
	_ "gno.land/r/gov/dao/init" // so that the govdao initializer is executed
  "gno.land/r/gov/dao/bridge"
	govdao "gno.land/r/gov/dao/v2"
)

const (
	minFee     int64    = 20 * 1_000_000 // minimum gnot must be paid to register.

  moniker    string   = "moniker"
  description string   = "description"
	pubKey   = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p"
)

var (
  test1Address = testutils.TestAddress("test1")
	validP2PAddress   = ufmt.Sprintf("%s@127.0.0.1:26656", test1Address)
)

func TestValopers_MakeProposal(t *testing.T) {
  // Send coins to be able to register a valoper
  std.TestSetOriginSend(std.Coins{std.NewCoin("ugnot", minFee)}, nil)

  uassert.NotPanics(t, func() {
    valopers.Register(moniker, description, test1Address, pubKey, validP2PAddress)
  })

  var valoper valopers.Valoper

  uassert.NotPanics(t, func() {
    valoper = valopers.GetByAddr(test1Address)
  })

  dao := bridge.GovDAO()
  propStore := dao.GetPropStore()

  uassert.Equal(t, 0, propStore.Size())

  uassert.NotPanics(t, func() {
    MakeProposal(test1Address)
  })

  uassert.Equal(t, 1, propStore.Size())

  proposal, err := propStore.ProposalByID(0)
  uassert.NoError(t, err, "proposal not found")

  // Check that the proposal is correct
  uassert.Equal(t, valoper.Render(), proposal.Description())
}
