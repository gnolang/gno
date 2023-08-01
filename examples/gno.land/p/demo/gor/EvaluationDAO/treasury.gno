package treasury

import "std"

// owner(treasurer) is the owner of this contract
// owner can add trustee(other members) into group
// any one can propose
// proposal approved only when quorum reached

// NOTE: admin here should be EvaluationDAO.
// validators/contributors will propose, the proposal will be approved/rejected by EvaluationDAO
// actions would be taken iff proposal approved

// xxx how to bootstrap a DAO, should be from an initial admin, deployer probaly

var admin std.Address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // TODO: helper to change admin

// how many votes at least
var threshold uint

// proposals retain approvals in a mapping, the value means:
// 0 : has not approved (default)
// 1 : has approved

type SpendProposal struct {
	payee     std.Address
	amount    uint
	approvals map[std.Address]bool
	count     uint
}

type TransferProposal struct {
	toksca    std.Address
	to        std.Address
	amount    uint
	approvals map[std.Address]bool
	count     uint
}

type flag uint8

const (
	InvalidTrustee flag = iota
	ValidTrustee
	FlaggedTrustee
)

// all trustees
var trustees *avl.Tree // address -> flag
var trusteeCount uint

func SetMembers(...std.Address) {
	// iterate add
}

func add(trustee std.Address) {}

func Flag(trustee std.Address) {}

func Replace(old, new std.Address) {}

// propose
func proposeSpend() {}

func proposeTransfer() {}

func ApproveSpend()    {}
func ApproveTransfer() {}

// fees
// here we collect fees, then distributed when proposal approval by EvaluationDAO
// xxx don't need this acatully, fees will be sent to this package without calling a function explicitly
func CollectFee() {
	std.AssertOriginCall()
	sent := std.GetOrigSend()
	println("receive fees of:", sent)
}

func render() string {
	return ""
}
