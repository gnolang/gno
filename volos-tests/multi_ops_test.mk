include _info.mk
include ../gnoswap-tests/_info.mk
include ../gnoswap-tests/test.mk

# =============================================================================
# MULTI-OPERATIONS WORKFLOW FOR GRAPH DATA GENERATION
# =============================================================================
# 
# This workflow performs multiple supply, withdraw, borrow, and repay operations
# to generate rich data points for utilization rate graphs and market analysis.
#
# PREREQUISITES:
# - Markets must be created first (run full-workflow from test.mk)
# - Sufficient token balances and allowances must be set up
# - Collateral must be supplied to enable borrowing
#
# DEPENDENCIES:
# - full-workflow (from test.mk) must be completed first
# - Note: 74% of GNS is already borrowed from full-workflow, LLTV is 75%, so only 1% headroom
#
# USAGE:
# make -f test.mk full-workflow    # First create markets and basic setup
# make -f multi_ops_test.mk multi-ops-workflow  # Then run this for data generation
# =============================================================================

# Multiple operations workflow to generate more data points for graphs
# This workflow performs multiple supply, withdraw, borrow, and repay operations
multi-ops-workflow: \
	supply-multi-1 withdraw-multi-1 borrow-multi-1 repay-multi-1 \
	supply-multi-2 withdraw-multi-2 borrow-multi-2 repay-multi-2 \
	supply-multi-3 withdraw-multi-3 borrow-multi-3 repay-multi-3 \
	supply-multi-4 withdraw-multi-4 borrow-multi-4 repay-multi-4 \
	supply-multi-5 withdraw-multi-5 borrow-multi-5 repay-multi-5 \
	check-final-positions
	@echo "************ MULTI-OPS WORKFLOW FINISHED ************"

# Ensure allowances are set for multi-operations
ensure-allowances:
	$(info ************ Ensuring allowances for multi-operations ************)
	# Approve GNS for Volos contract
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/gns -func Approve -args $(ADDR_VOLOS) -args $(MAX_APPROVE) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo
	# Approve WUGNOT for Volos contract
	@echo "" | gnokey maketx call -pkgpath gno.land/r/demo/wugnot -func Approve -args $(ADDR_VOLOS) -args $(MAX_APPROVE) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Multi-operation targets for generating more data points
# Round 1 operations
supply-multi-1:
	$(info ************ Multi-op Round 1: Supply GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Supply -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 5000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

withdraw-multi-1:
	$(info ************ Multi-op Round 1: Withdraw GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Withdraw -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 2000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

borrow-multi-1:
	$(info ************ Multi-op Round 1: Borrow GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Borrow -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 200000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

repay-multi-1:
	$(info ************ Multi-op Round 1: Repay GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Repay -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 1000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Round 2 operations
supply-multi-2:
	$(info ************ Multi-op Round 2: Supply GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Supply -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 300000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

withdraw-multi-2:
	$(info ************ Multi-op Round 2: Withdraw GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Withdraw -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 1000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

borrow-multi-2:
	$(info ************ Multi-op Round 2: Borrow GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Borrow -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 50000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

repay-multi-2:
	$(info ************ Multi-op Round 2: Repay GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Repay -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 120000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Round 3 operations
supply-multi-3:
	$(info ************ Multi-op Round 3: Supply GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Supply -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 400000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

withdraw-multi-3:
	$(info ************ Multi-op Round 3: Withdraw GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Withdraw -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 150000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

borrow-multi-3:
	$(info ************ Multi-op Round 3: Borrow GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Borrow -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 3000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

repay-multi-3:
	$(info ************ Multi-op Round 3: Repay GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Repay -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 200000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Round 4 operations
supply-multi-4:
	$(info ************ Multi-op Round 4: Supply GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Supply -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 600000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

withdraw-multi-4:
	$(info ************ Multi-op Round 4: Withdraw GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Withdraw -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 2000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

borrow-multi-4:
	$(info ************ Multi-op Round 4: Borrow GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Borrow -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 3000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

repay-multi-4:
	$(info ************ Multi-op Round 4: Repay GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Repay -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 1000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Round 5 operations
supply-multi-5:
	$(info ************ Multi-op Round 5: Supply GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Supply -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 7000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

withdraw-multi-5:
	$(info ************ Multi-op Round 5: Withdraw GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Withdraw -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 2500 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

borrow-multi-5:
	$(info ************ Multi-op Round 5: Borrow GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Borrow -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 100000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

repay-multi-5:
	$(info ************ Multi-op Round 5: Repay GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Repay -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 300000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Check final positions after all multi-operations
check-final-positions:
	$(info ************ Check Final Positions After Multi-Operations ************)
	# Check GNS-WUGNOT market final state
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetTotalSupplyAssets(\"gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0\")"
	@echo
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetTotalBorrowAssets(\"gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0\")"
	@echo
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetPositionSupplyShares(\"gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0\", \"$(ADMIN)\")"
	@echo
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetPositionBorrowShares(\"gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0\", \"$(ADMIN)\")"
	@echo

# Complete workflow with prerequisites check
multi-ops-complete: ensure-allowances multi-ops-workflow
	@echo "************ MULTI-OPS COMPLETE WORKFLOW FINISHED ************"

# Quick test to verify market exists before running multi-ops
verify-market-exists:
	$(info ************ Verifying market exists before multi-operations ************)
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetMarket(\"gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0\")"
	@echo

# Check current utilization rate before multi-operations
check-current-utilization:
	$(info ************ Check Current Utilization Rate ************)
	# Get total supply assets
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetTotalSupplyAssets(\"gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0\")"
	@echo
	# Get total borrow assets
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetTotalBorrowAssets(\"gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0\")"
	@echo
	# Get LLTV (Liquidation LTV)
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetLLTV(\"gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0\")"
	@echo
