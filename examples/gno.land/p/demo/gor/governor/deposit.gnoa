package governor

import std "std"
import "gno.land/p/demo/avl"

// update deposit for a proposal
// return when active
// burn when fail or not reaching quorum? NO

// deposits only burned when veto

// XXX deposit is used by proposal well as vote with veto
func (gvr *Governor) Deposit(proposalId string) error {
	// std.AssertOriginCall()
	// caller is who deposit, will got returned deposit once succeeded
	caller := std.GetOrigCaller()
	// banker do the transfer work, for native coin
	banker := std.GetBanker(std.BankerTypeOrigSend)
	pkgaddr := std.GetOrigPkgAddr()
	println("pkgaddr: ", pkgaddr)
	// get deposit coin
	var coinD std.Coin
	sentCoins := std.GetOrigSend()
	if len(sentCoins) == 1 {
		coinD = sentCoins[0]
	}
	// check if return coin
	var proposal *ProposalCore
	proposalI, found := gvr.proposals.Get(proposalId)
	if !found {
		// return deposit
		banker.SendCoins(pkgaddr, caller, sentCoins)
		return ErrProposalNotExisit
	} else {
		proposal = proposalI.(*ProposalCore)
	}

	// check deposit state
	if d, ok := gvr.GetDeposit(proposalId); ok {
		if d.IsGTE(gvr.gs.minDeposit) {
			return ErrDepositEnded
		}
	}

	// if not return, send coins to governor pkg, coins is maintained in banker
	pkgPath := std.CurrentRealmPath()
	to := std.DerivePkgAddr(pkgPath) // replace it with std.CurrentRealmPath
	banker.SendCoins(pkgaddr, to, sentCoins)

	// maintain deposits: proposalId <=> total
	total, ok := gvr.GetDeposit(proposalId)
	sum := total.Add(coinD)
	gvr.deposits.Set(proposalId, sum)

	// manipulate proposal state, via setting voteStart and voteEnd parameter
	var snapshot int64
	// check if deposit amount reached
	// check threshold
	if sum.IsGTE(gvr.gs.minDeposit) {
		// ok to vote
		snapshot = std.GetHeight()
	} else {
		snapshot = std.GetHeight() + gvr.gs.getVotingDelay()
	}

	deadline := snapshot + gvr.gs.getVotingPeriod()
	proposal.voteStart.SetDeadline(snapshot)
	proposal.voteEnd.SetDeadline(deadline)

	// maintain depositList: proposalId=>(caller=>amount)
	var deposit *avl.Tree
	idt, ok := gvr.depositList.Get(proposalId)
	if !ok {
		// first time deposit for this proposalId
		deposit = avl.NewTree()
	} else {
		deposit = idt.(*avl.Tree) // caller => amount
	}
	// caller => amount sums
	amt := std.Coin{Amount: 0, Denom: "ugnot"}
	iamt, ok := deposit.Get(caller.String())
	if ok {
		amt = iamt.(std.Coin)
	}
	sum = amt.Add(coinD)
	// update
	// deposit.Set(caller.String(), sum)
	deposit.Set(pkgaddr.String(), sum)
	gvr.depositList.Set(proposalId, deposit)

	return nil
}

// get deposit for a proposal
func (gvr *Governor) GetDeposit(proposalId string) (std.Coin, bool) {
	d, ok := gvr.deposits.Get(proposalId)
	if !ok {
		return std.Coin{Amount: 0, Denom: "ugnot"}, false
	}
	deposit := d.(std.Coin)
	return deposit, true
}

// trigger
func (gvr *Governor) Burn(proposalId string) {

}

// func (gvr *Governor) Refund(proposalId string, isBurn bool) error {
func (gvr *Governor) refund(proposalId string, isBurn bool) error {
	// banker do the transfer work, for native coin
	banker := std.GetBanker(std.BankerTypeRealmSend)
	pkgPath := std.CurrentRealmPath()
	// vault := std.DerivePkgAddr("gno.land/p/demo/governance")
	vault := std.DerivePkgAddr(pkgPath)
	println(banker.GetCoins(vault))
	// iterate depositList, get caller, and amount, update value
	idt, ok := gvr.depositList.Get(proposalId)
	if !ok {
		return ErrProposalNotExisit
	}
	deposit := idt.(*avl.Tree)

	var toAddr std.Address
	deposit.Iterate("", "", func(n *avl.Node) bool {
		c := n.Key()
		a := n.Value().(std.Coin)
		println("caller: ", c)
		println("amount: ", a)
		deposit.Set(c, std.Coins{std.Coin{Denom: "ugnot", Amount: 0}})
		if !isBurn {
			toAddr = std.Address(c)
		} else {
			toAddr = std.Address("")
		}
		// concrete send back coins
		banker.SendCoins(vault, toAddr, std.Coins{a})
		return false
	})

	// set every deposit to 0
	gvr.depositList.Set(proposalId, deposit)

	// set total amount to 0
	gvr.deposits.Set(proposalId, std.Coins{std.Coin{Denom: "ugnot", Amount: 0}})
	return nil
}

// return coins when succeed, or burn when fail
// fail in two ways: deposit not reach quorum, or 33.4% vote NOWithVeto
func (gvr *Governor) Refund(proposalId string) error {

	// if passed, iterate depositList, return value, update deposits, and depositList
	// return deposit
	if gvr.state(proposalId) == Succeeded {
		gvr.refund(proposalId, false)
	}
	// if failed, burn deposit, update deposits, and depositList
	// check NoWithVeto percentage, if exeeded 33.4% , burn, or if votes not reaching quorum, burn
	if gvr.state(proposalId) == QuorumNotReach {
		println("quorum not reach, going to burn")
		gvr.refund(proposalId, true)
	}
	if gvr.state(proposalId) == Defeated {
		println("defeated, going to burn")
		gvr.refund(proposalId, true)
	}
	return nil
}
