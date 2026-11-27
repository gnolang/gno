// PKGPATH: gno.land/r/nt/multisig/v0/filetests/z_signers

package z_signers

import (
	"testing"
	"time"

	"gno.land/p/nt/testutils/v0"

	"gno.land/r/nt/multisig/v0"
)

const t0 int64 = 1_700_000_000

func main(cur realm) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	carol := testutils.TestAddress("carol")
	dest := testutils.TestAddress("dest")
	signers := alice.String() + "," + bob.String() + "," + carol.String()

	as(t0, alice)
	id := multisig.Create(cross(cur), "crew", signers, 2, 0)
	spend := multisig.ProposeCoin(cross(cur), id, dest, "ugnot", 1, t0+1000, "")
	remove := multisig.ProposeRemoveSigner(cross(cur), id, carol, t0+1000, "")
	multisig.Approve(cross(cur), id, spend)
	multisig.Approve(cross(cur), id, remove)

	as(t0, carol)
	multisig.Approve(cross(cur), id, spend)

	as(t0, bob)
	multisig.Approve(cross(cur), id, remove)
	multisig.Execute(cross(cur), id, remove)

	if multisig.IsSigner(id, carol) {
		panic("carol should have been removed")
	}
	if multisig.ApprovalCount(id, spend) != 1 {
		panic("carol's approval should not count after removal")
	}
	if multisig.ProposalStatus(id, spend) == "executed" {
		panic("spend must not execute on a removed signer's vote")
	}
	println("ok")
}

func as(unix int64, user address) {
	ctx := testing.GetContext()
	ctx.Time = time.Unix(unix, 0)
	ctx.CurrentRealm = testing.NewUserRealm(user)
	ctx.OriginCaller = user
	testing.SetContext(ctx)
}

// Output:
// ok
