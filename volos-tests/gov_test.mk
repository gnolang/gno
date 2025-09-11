VLS_PATH := gno.land/r/volos/gov/vls

# Basic governance setup and test
gov-test-flow: faucet-vls approve-vls-for-staking stake-vls transfer-vls approve-all-voters stake-all-voters create-test-proposal vote-all-on-all-proposals
	@echo "************ GOVERNANCE ENVIRONMENT SETUP COMPLETE ************"

# Basic governance setup and test
gov-test-flow-no-voting: faucet-vls approve-vls-for-staking stake-vls faucet-all-voters approve-all-voters stake-all-voters create-test-proposal
	@echo "************ GOVERNANCE ENVIRONMENT SETUP COMPLETE ************"

# Faucet VLS tokens
faucet-vls:
	$(info ************ Faucet VLS tokens ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/vls -func Faucet -args 5000000 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Approve VLS for staking
approve-vls-for-staking:
	$(info ************ Approve VLS for staking ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/vls -func Approve -args $(ADDR_STAKER) -args 10000 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Stake VLS to mint xVLS
stake-vls:
	$(info ************ Stake VLS to mint xVLS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/staker -func Stake -args 5000 -args $(ADMIN) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Transfer VLS to all voters
transfer-vls:
	$(info ************ Transfer VLS tokens to all voters ************)
	@echo "" | gnokey maketx call -pkgpath $(VLS_PATH) -func Transfer -args $(ADDR_USER_1) -args 1000000 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "transfer 1_000_000 VLS to $(ADDR_USER_1)" gnoswap_admin
	@echo "" | gnokey maketx call -pkgpath $(VLS_PATH) -func Transfer -args $(ADDR_USER_2) -args 1000000 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "transfer 1_000_000 VLS to $(ADDR_USER_2)" gnoswap_admin
	@echo "" | gnokey maketx call -pkgpath $(VLS_PATH) -func Transfer -args $(ADDR_USER_3) -args 1000000 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "transfer 1_000_000 VLS to $(ADDR_USER_3)" gnoswap_admin
	@echo

# Approve VLS for staking for all voters
approve-all-voters:
	$(info ************ Approve VLS for staking for all voters ************)
	@echo "" | gnokey maketx call -pkgpath $(VLS_PATH) -func Approve -args $(ADDR_STAKER) -args 10000 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_1)
	@echo "" | gnokey maketx call -pkgpath $(VLS_PATH) -func Approve -args $(ADDR_STAKER) -args 10000 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_2)
	@echo "" | gnokey maketx call -pkgpath $(VLS_PATH) -func Approve -args $(ADDR_STAKER) -args 10000 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_3)
	@echo

# Stake VLS to mint xVLS for all voters
stake-all-voters:
	$(info ************ Stake VLS to mint xVLS for all voters ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/staker -func Stake -args 3000 -args $(ADDR_USER_1) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_1)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/staker -func Stake -args 4000 -args $(ADDR_USER_2) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_2)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/staker -func Stake -args 2000 -args $(ADDR_USER_3) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_3)
	@echo

# Create a simple test proposal
create-test-proposal:
	$(info ************ Create test proposal ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/mocks -func CreateProposals -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo "waiting 3s for proposals to be processed..."
	@sleep 3
	@echo

# Vote yes on the proposal
vote-yes:
	$(info ************ Vote yes on proposal ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 1 -args "YES" -args "I support this proposal" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# All voters vote on proposal 1
vote-all-on-proposal1:
	$(info ************ All voters vote on proposal 1 ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 1 -args "YES" -args "Admin supports proposal 1" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 1 -args "YES" -args "ADDR_USER_1 supports proposal 1" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_1)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 1 -args "YES" -args "ADDR_USER_2 supports proposal 1" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_2)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 1 -args "YES" -args "ADDR_USER_3 supports proposal 1" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_3)
	@echo

# Vote no on the proposal
vote-no:
	$(info ************ Vote no on proposal ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 2 -args "NO" -args "I do not support this proposal" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Vote abstain on the proposal
vote-abstain:
	$(info ************ Vote abstain on proposal ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 3 -args "ABSTAIN" -args "I abstain from this proposal" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Multi-voter voting scenarios
# ADDR_USER_1 votes YES on proposal 1
vote-voter1-yes:
	$(info ************ ADDR_USER_1 votes YES on proposal 1 ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 1 -args "YES" -args "ADDR_USER_1 supports proposal 1" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_1)
	@echo

# ADDR_USER_2 votes NO on proposal 2
vote-voter2-no:
	$(info ************ ADDR_USER_2 votes NO on proposal 2 ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 2 -args "NO" -args "ADDR_USER_2 opposes proposal 2" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_2)
	@echo

# ADDR_USER_3 votes ABSTAIN on proposal 3
vote-voter3-abstain:
	$(info ************ ADDR_USER_3 votes ABSTAIN on proposal 3 ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 3 -args "ABSTAIN" -args "ADDR_USER_3 abstains from proposal 3" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_3)
	@echo

# ADDR_USER_1 votes NO on proposal 2
vote-voter1-no:
	$(info ************ ADDR_USER_1 votes NO on proposal 2 ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 2 -args "NO" -args "ADDR_USER_1 opposes proposal 2" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_1)
	@echo

# ADDR_USER_2 votes YES on proposal 1
vote-voter2-yes:
	$(info ************ ADDR_USER_2 votes YES on proposal 1 ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 1 -args "YES" -args "ADDR_USER_2 supports proposal 1" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_2)
	@echo

# ADDR_USER_3 votes YES on proposal 1
vote-voter3-yes:
	$(info ************ ADDR_USER_3 votes YES on proposal 1 ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 1 -args "YES" -args "ADDR_USER_3 supports proposal 1" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_3)
	@echo

# Multi-voter test flow
multi-voter-test: vote-voter1-yes vote-voter2-no vote-voter3-abstain vote-voter1-no vote-voter2-yes vote-voter3-yes
	@echo "************ MULTI-ADDR_USER_ TEST COMPLETE ************"

# All voters vote on all proposals (1, 2, 3)
vote-all-on-all-proposals:
	$(info ************ All voters vote on all proposals ************)
	# Vote on proposal 1
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 1 -args "YES" -args "Admin supports proposal 1" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 1 -args "YES" -args "ADDR_USER_1 supports proposal 1" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_1)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 1 -args "YES" -args "ADDR_USER_2 supports proposal 1" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_2)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 1 -args "YES" -args "ADDR_USER_3 supports proposal 1" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_3)
	# Vote on proposal 2
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 2 -args "NO" -args "Admin opposes proposal 2" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 2 -args "NO" -args "ADDR_USER_1 opposes proposal 2" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_1)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 2 -args "NO" -args "ADDR_USER_2 opposes proposal 2" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_2)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 2 -args "ABSTAIN" -args "ADDR_USER_3 abstains from proposal 2" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_3)
	# Vote on proposal 3
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 3 -args "YES" -args "Admin supports proposal 3" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 3 -args "ABSTAIN" -args "ADDR_USER_1 abstains from proposal 3" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_1)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 3 -args "YES" -args "ADDR_USER_2 supports proposal 3" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_2)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Vote -args 3 -args "YES" -args "ADDR_USER_3 supports proposal 3" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" $(ADDR_USER_3)
	@echo

# Execute the proposal
execute-proposal:
	$(info ************ Execute proposal ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/gov/governance -func Execute -args 1 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Check VLS balance
check-vls-balance:
	$(info ************ Check VLS balance ************)
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/gov/vls.BalanceOf(\"$(ADMIN)\")"
	@echo

# Check xVLS balance
check-xvls-balance:
	$(info ************ Check xVLS balance ************)
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/gov/xvls.BalanceOf(\"$(ADMIN)\")"
	@echo

# Check governance membership
check-governance-membership:
	$(info ************ Check governance membership ************)
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/gov/governance.MemberSet().Has(\"$(ADMIN)\")"
	@echo

# Get proposal info
get-proposal:
	$(info ************ Get proposal info ************)
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/gov/governance.GetProposal(1)"
	@echo

# Check ADDR_USER_1 xVLS balance
check-voter1-xvls:
	$(info ************ Check ADDR_USER_1 xVLS balance ************)
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/gov/xvls.BalanceOf(\"$(ADDR_USER_1)\")"
	@echo

# Check ADDR_USER_2 xVLS balance
check-voter2-xvls:
	$(info ************ Check ADDR_USER_2 xVLS balance ************)
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/gov/xvls.BalanceOf(\"$(ADDR_USER_2)\")"
	@echo

# Check ADDR_USER_3 xVLS balance
check-voter3-xvls:
	$(info ************ Check ADDR_USER_3 xVLS balance ************)
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/gov/xvls.BalanceOf(\"$(ADDR_USER_3)\")"
	@echo

# Check all voters' xVLS balances
check-all-voters-xvls: check-voter1-xvls check-voter2-xvls check-voter3-xvls
	@echo "************ ALL ADDR_USER_S XVLS BALANCES CHECKED ************"

# Quick verification of setup
gov-verify: check-vls-balance check-xvls-balance check-governance-membership get-proposal check-all-voters-xvls
	@echo "************ GOVERNANCE ENVIRONMENT VERIFICATION COMPLETE ************"

# Comprehensive governance test with multiple proposals and all voters
gov-comprehensive-test: faucet-vls approve-vls-for-staking stake-vls faucet-all-voters approve-all-voters stake-all-voters create-multiple-proposals vote-all-on-all-proposals
	@echo "************ COMPREHENSIVE GOVERNANCE TEST COMPLETE ************" 
