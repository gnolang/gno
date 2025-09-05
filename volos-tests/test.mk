include _info.mk
include ../gnoswap-tests/_info.mk
include ../gnoswap-tests/test.mk

# Complete flow that includes both GNS-WUGNOT and BAR-WUGNOT operations
full-workflow: pool-create-gns-wugnot-default mint-gns-gnot enable-irm enable-lltv market-create-gns-wugnot supply-assets-gns-wugnot supply-collateral-gns-wugnot borrow-gns \
	pool-create-bar-wugnot-default mint-bar-wugnot market-create-bar-wugnot supply-assets-bar-wugnot supply-collateral-bar-wugnot borrow-bar \
	check-position-gns-wugnot check-position-bar-wugnot
	@echo "************ WORKFLOW FINISHED ************"

# Enable the linear IRM
enable-irm:
	$(info ************ Enable linear IRM ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func EnableIRM -args "linear" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Enable LLTV (75% = 75 as int64)
enable-lltv:
	$(info ************ Enable LLTV 75% ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func EnableLLTV -args 75 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Test market creation with GNS and WUGNOT
market-create-gns-wugnot:
	$(info ************ Test creating market with GNS (supply/borrow) and WUGNOT (collateral) ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func CreateMarket -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000" -args false -args "linear" -args 75 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Test market creation with GNS and WUGNOT
market-create-bar-wugnot:
	$(info ************ Test creating market with BAR (supply/borrow) and WUGNOT (collateral) ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func CreateMarket -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/test_token/bar:3000" -args false -args "linear" -args 75 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Test getting pool price for GNS-WUGNOT market
market-get-price-gns-wugnot:
	$(info ************ Test getting pool price for GNS-WUGNOT market ************)
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetMarketPrice(\"gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0\")"
	@echo

# Test getting market info for GNS-WUGNOT pair
market-get-gns-wugnot:
	$(info ************ Test getting market info for GNS-WUGNOT pair ************)
	# GET TOTAL SUPPLY ASSETS
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetTotalSupplyAssets(\"gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0\")"
	@echo

	# GET TOTAL SUPPLY SHARES
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetTotalSupplyShares(\"gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0\")"
	@echo

	# GET TOTAL BORROW ASSETS
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetTotalBorrowAssets(\"gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0\")"
	@echo

	# GET TOTAL BORROW SHARES
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetTotalBorrowShares(\"gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0\")"
	@echo

	# GET LIQUIDATION LTV
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetLLTV(\"gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0\")"
	@echo

	# GET MARKET FEE
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetFee(\"gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0\")"
	@echo

# Test supplying assets to GNS-WUGNOT market
supply-assets-gns-wugnot:
	$(info ************ Test supplying GNS assets to GNS-WUGNOT market ************)
	# APPROVE FIRST
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/gns -func Approve -args $(ADDR_VOLOS) -args $(MAX_APPROVE) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

	# THEN SUPPLY
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Supply -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 1480000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Test supplying shares to GNS-WUGNOT market
supply-shares-gns-wugnot:
	$(info ************ Test supplying shares to GNS-WUGNOT market ************)
	# APPROVE FIRST
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/gns -func Approve -args $(ADDR_VOLOS) -args $(MAX_APPROVE) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

	# THEN SUPPLY
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Supply -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 0 -args 1000000 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Test withdrawing assets from GNS-WUGNOT market
withdraw-assets-gns-wugnot:
	$(info ************ Test withdrawing GNS assets from GNS-WUGNOT market ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Withdraw -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 994940 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Check user position in GNS-WUGNOT market
check-position-gns-wugnot:
	$(info ************ Check user position in GNS-WUGNOT market ************)
	# Check supply shares
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetPositionSupplyShares(\"gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0\", \"$(ADMIN)\")"
	@echo

	# Check borrow shares
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetPositionBorrowShares(\"gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0\", \"$(ADMIN)\")"
	@echo

	# Check collateral
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetPositionCollateral(\"gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0\", \"$(ADMIN)\")"
	@echo

# Check GNS balance
check-gns-balance:
	$(info ************ Check GNS balance ************)
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/gnoswap/v1/gns.BalanceOf(\"$(ADMIN)\")"
	@echo

# Check Volos contract GNS balance
check-volos-gns-balance:
	$(info ************ Check Volos contract GNS balance ************)
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/gnoswap/v1/gns.BalanceOf(\"$(ADDR_VOLOS)\")"
	@echo

# Check Volos contract WUGNOT balance
check-volos-wugnot-balance:
	$(info ************ Check Volos contract WUGNOT balance ************)
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/demo/wugnot.BalanceOf(\"$(ADDR_VOLOS)\")"
	@echo

# Test accruing interest on GNS-WUGNOT market
accrue-interest-gns-wugnot:
	$(info ************ Test accruing interest on GNS-WUGNOT market ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func AccrueInterest -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Check WUGNOT balance
check-wugnot-balance:
	$(info ************ Check WUGNOT balance ************)
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/demo/wugnot.BalanceOf(\"$(ADMIN)\")"
	@echo

# Wrap UGNOT to WUGNOT
wrap-ugnot:
	$(info ************ Wrap UGNOT to WUGNOT ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/demo/wugnot -func Deposit -send "1000000000ugnot" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Test borrowing GNS tokens
borrow-gns:
	$(info ************ Test borrowing GNS tokens ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Borrow -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 74000000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Test repaying GNS tokens
repay-gns:
	$(info ************ Test repaying GNS tokens ************)
	# APPROVE FIRST
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/gns -func Approve -args $(ADDR_VOLOS) -args $(MAX_APPROVE) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

	# THEN REPAY
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Repay -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 50 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Test liquidating GNS position
liquidate-gns:
	$(info ************ Test liquidating GNS position ************)
	# APPROVE FIRST
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/gns -func Approve -args $(ADDR_VOLOS) -args $(MAX_APPROVE) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

	# THEN LIQUIDATE
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Liquidate -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args $(ADMIN) -args 0 -args 25 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Test supplying collateral to GNS-WUGNOT market
supply-collateral-gns-wugnot:
	$(info ************ Test supplying collateral to GNS-WUGNOT market ************)
	# APPROVE FIRST
	@echo "" | gnokey maketx call -pkgpath gno.land/r/demo/wugnot -func Approve -args $(ADDR_VOLOS) -args $(MAX_APPROVE) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

	# Wrap UGNOT to WUGNOT
	@echo "" | gnokey maketx call -pkgpath gno.land/r/demo/wugnot -func Deposit -send "1000000000ugnot" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

	# THEN SUPPLY COLLATERAL
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func SupplyCollateral -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 1000000 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Test withdrawing collateral from GNS-WUGNOT market
withdraw-collateral-gns-wugnot:
	$(info ************ Test withdrawing collateral from GNS-WUGNOT market ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func WithdrawCollateral -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0" -args 500 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Test supplying assets to BAR-WUGNOT market
supply-assets-bar-wugnot:
	$(info ************ Test supplying BAR assets to BAR-WUGNOT market ************)
	# APPROVE FIRST
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/test_token/bar -func Approve -args $(ADDR_VOLOS) -args $(MAX_APPROVE) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

	# Wrap UGNOT to WUGNOT if needed
	@echo "" | gnokey maketx call -pkgpath gno.land/r/demo/wugnot -func Deposit -send "1000000000ugnot" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

	# THEN SUPPLY
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Supply -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/test_token/bar:3000:0" -args 1000000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Test supplying collateral to BAR-WUGNOT market
supply-collateral-bar-wugnot:
	$(info ************ Test supplying collateral to BAR-WUGNOT market ************)
	# APPROVE FIRST
	@echo "" | gnokey maketx call -pkgpath gno.land/r/demo/wugnot -func Approve -args $(ADDR_VOLOS) -args $(MAX_APPROVE) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

	# THEN SUPPLY COLLATERAL
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func SupplyCollateral -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/test_token/bar:3000:0" -args 1000000 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Test borrowing BAR tokens
borrow-bar:
	$(info ************ Test borrowing BAR tokens ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/volos/core -func Borrow -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/test_token/bar:3000:0" -args 500000 -args 0 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# Check user position in BAR-WUGNOT market
check-position-bar-wugnot:
	$(info ************ Check user position in BAR-WUGNOT market ************)
	# Check supply shares
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetPositionSupplyShares(\"gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/test_token/bar:3000:0\", \"$(ADMIN)\")"
	@echo

	# Check borrow shares
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetPositionBorrowShares(\"gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/test_token/bar:3000:0\", \"$(ADMIN)\")"
	@echo

	# Check collateral
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetPositionCollateral(\"gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/test_token/bar:3000:0\", \"$(ADMIN)\")"
	@echo

	# Check health factor
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetHealthFactor(\"gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/test_token/bar:3000:0\", \"$(ADMIN)\")"
	@echo

# Check volos owner
check-volos-owner:
	$(info ************ Check volos owner ************)
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetOwner()"
	@echo

# Test flash loan
test-flashloan:
	$(info ************ Testing Flash Loan ************)
	@echo "" | gnokey maketx run gnoswap_admin flashloan.gno -gas-wanted 200000000 -gas-fee 1000000ugnot -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID)
	@echo

# Get Volos contract address derived from path
get-volos-address:
	$(info ************ Getting Volos Contract Address ************)
	gnokey query vm/qeval -remote $(GNOLAND_RPC_URL) -data "gno.land/r/volos/core.GetAddress()"
	@echo