package multisig

import (
	"chain/banker"

	msig "gno.land/p/nt/multisig/v0"
)

// SafeCount is the number of safes created.
func SafeCount() int { return len(safes) }

// SafeName returns the display name.
func SafeName(safeID int64) string { return mustSafe(safeID).Name() }

// SafeAddress returns the account that holds the safe's coins.
func SafeAddress(safeID int64) address { return mustSafe(safeID).Address() }

// SafeThreshold returns the current threshold.
func SafeThreshold(safeID int64) int { return mustSafe(safeID).Threshold() }

// SafeDelay returns the timelock for new proposals, in seconds.
func SafeDelay(safeID int64) int64 { return mustSafe(safeID).Delay() }

// SafePaused reports whether spends are paused.
func SafePaused(safeID int64) bool { return mustSafe(safeID).Paused() }

// SafeNonce is the number of proposals executed.
func SafeNonce(safeID int64) int64 { return mustSafe(safeID).Nonce() }

// SignerCount returns the size of the signer set.
func SignerCount(safeID int64) int { return mustSafe(safeID).SignerCount() }

// SignerAt returns one signer. index is zero-based.
func SignerAt(safeID int64, index int) address {
	return mustSafe(safeID).SignerAt(index)
}

// IsSigner reports membership.
func IsSigner(safeID int64, addr address) bool {
	return mustSafe(safeID).IsSigner(addr)
}

// ProposalCount includes proposals that already finished.
func ProposalCount(safeID int64) int { return mustSafe(safeID).ProposalCount() }

// ProposalStatus is pending, executed, cancelled, or expired.
func ProposalStatus(safeID, propID int64) string {
	view, err := mustSafe(safeID).View(propID)
	must(err)
	return msig.DisplayStatus(view.Status, view.ExpiresAt, nowUnix()).String()
}

// ApprovalCount counts votes that still belong to current signers.
func ApprovalCount(safeID, propID int64) int {
	n, err := mustSafe(safeID).ApprovalCount(propID)
	must(err)
	return n
}

// Balance is the safe's native-coin balance of one denom.
// A third party can send extra denoms to any address, so this reads
// the denom the caller named instead of every coin the safe holds.
func Balance(safeID int64, denom string) int64 {
	return banker.NewReadonlyBanker().GetCoin(mustSafe(safeID).Address(), denom)
}
