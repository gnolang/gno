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
	ensure-allowances \
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

# =============================================================================
# MULTI-OPS DEMO WORKFLOW - ASSUMES MARKETS AND TOKENS ALREADY EXIST
# =============================================================================
# 
# Simple demo workflow that just does supply/withdraw/borrow/repay operations.
# Assumes markets are already created and tokens are available.
#
# USAGE:
# make -f multi_ops_test.mk multi-ops-demo-workflow
# =============================================================================

# Simple demo workflow - just the operations
multi-ops-demo-workflow: \
	ensure-allowances \
	demo-supply-collateral-1 demo-supply-1 demo-withdraw-1 demo-borrow-1 demo-repay-1 \
	demo-supply-2 demo-withdraw-2 demo-borrow-2 demo-repay-2 \
	demo-supply-3 demo-withdraw-3 demo-borrow-3 demo-repay-3 \
	demo-supply-4 demo-withdraw-4 demo-borrow-4 demo-repay-4 \
	demo-supply-5 demo-withdraw-5 demo-borrow-5 demo-repay-5 \
	demo-supply-6 demo-withdraw-6 demo-borrow-6 demo-repay-6 \
	demo-supply-7 demo-withdraw-7 demo-borrow-7 demo-repay-7 \
	demo-supply-8 demo-withdraw-8
	@echo "************ MULTI-OPS DEMO WORKFLOW FINISHED ************"

# Initial collateral supply to enable borrowing
demo-supply-collateral-1:
	$(info ************ Demo: Supply WUGNOT collateral ************)
	# Wrap UGNOT to WUGNOT first
	@echo "" | gnokey maketx call -pkgpath gno.land/r/demo/wugnot -func Deposit -send "1000000000ugnot" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo
	# Supply WUGNOT as collateral
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func SupplyCollateral -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 1000000 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Demo Round 1 operations
demo-supply-1:
	$(info ************ Demo Round 1: Supply GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Supply -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 5000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

demo-withdraw-1:
	$(info ************ Demo Round 1: Withdraw GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Withdraw -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 2000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

demo-borrow-1:
	$(info ************ Demo Round 1: Borrow GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Borrow -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 1000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

demo-repay-1:
	$(info ************ Demo Round 1: Repay GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Repay -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 1000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Demo Round 2 operations
demo-supply-2:
	$(info ************ Demo Round 2: Supply GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Supply -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 300000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

demo-withdraw-2:
	$(info ************ Demo Round 2: Withdraw GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Withdraw -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 1000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

demo-borrow-2:
	$(info ************ Demo Round 2: Borrow GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Borrow -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 200000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

demo-repay-2:
	$(info ************ Demo Round 2: Repay GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Repay -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 120000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Demo Round 3 operations
demo-supply-3:
	$(info ************ Demo Round 3: Supply GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Supply -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 400000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

demo-withdraw-3:
	$(info ************ Demo Round 3: Withdraw GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Withdraw -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 150000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

demo-borrow-3:
	$(info ************ Demo Round 3: Borrow GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Borrow -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 100000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

demo-repay-3:
	$(info ************ Demo Round 3: Repay GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Repay -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 200000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Demo Round 4 operations
demo-supply-4:
	$(info ************ Demo Round 4: Supply GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Supply -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 600000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

demo-withdraw-4:
	$(info ************ Demo Round 4: Withdraw GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Withdraw -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 2000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

demo-borrow-4:
	$(info ************ Demo Round 4: Borrow GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Borrow -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 200000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

demo-repay-4:
	$(info ************ Demo Round 4: Repay GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Repay -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 1000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Demo Round 5 operations
demo-supply-5:
	$(info ************ Demo Round 5: Supply GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Supply -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 7000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

demo-withdraw-5:
	$(info ************ Demo Round 5: Withdraw GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Withdraw -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 2500 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

demo-borrow-5:
	$(info ************ Demo Round 5: Borrow GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Borrow -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 5000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

demo-repay-5:
	$(info ************ Demo Round 5: Repay GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Repay -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 300000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Demo Round 6 operations
demo-supply-6:
	$(info ************ Demo Round 6: Supply GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Supply -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 800000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

demo-withdraw-6:
	$(info ************ Demo Round 6: Withdraw GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Withdraw -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 5000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

demo-borrow-6:
	$(info ************ Demo Round 6: Borrow GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Borrow -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 100000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

demo-repay-6:
	$(info ************ Demo Round 6: Repay GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Repay -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 150000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Demo Round 7 operations
demo-supply-7:
	$(info ************ Demo Round 7: Supply GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Supply -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 1200000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

demo-withdraw-7:
	$(info ************ Demo Round 7: Withdraw GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Withdraw -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 8000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

demo-borrow-7:
	$(info ************ Demo Round 7: Borrow GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Borrow -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 200000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

demo-repay-7:
	$(info ************ Demo Round 7: Repay GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Repay -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 200000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Demo Round 8 operations
demo-supply-8:
	$(info ************ Demo Round 8: Supply GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Supply -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 1500000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

demo-withdraw-8:
	$(info ************ Demo Round 8: Withdraw GNS ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Withdraw -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 10000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo