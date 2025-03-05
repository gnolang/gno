# Network configuration
GNOLAND_RPC_URL=http://localhost:26657
CHAINID=dev

ADMIN := g1mgk7hmna0w9ku7kynllmkukg573gxs9tclfqx3
ADDR_GNOLEND := g1vppywurq38q4x2rk2hyulv8tptcfq06lzapwhr
MAX_UINT64 := 18446744073709551615

# Test market creation with GNS and WUGNOT
market-create-gns-wugnot:
	$(info ************ Test creating market with GNS (supply/borrow) and WUGNOT (collateral) ************)
	@echo "" | gnokey maketx call \
		-pkgpath gno.land/r/gnolend \
		-func CreateMarket \
		-args "gno.land/r/gnoswap/v1/gns" \
		-args "gno.land/r/demo/wugnot" \
		-insecure-password-stdin=true \
		-remote $(GNOLAND_RPC_URL) \
		-broadcast=true \
		-chainid $(CHAINID) \
		-gas-fee 100000000ugnot \
		-gas-wanted 1000000000 \
		-memo "" \
		gnoswap_admin
	@echo

# Test getting market info for GNS-WUGNOT pair
market-get-gns-wugnot:
	$(info ************ Test getting market info for GNS-WUGNOT pair ************)
	@echo "" | gnokey maketx call \
		-pkgpath gno.land/r/gnolend \
		-func GetMarket \
		-args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns" \
		-insecure-password-stdin=true \
		-remote $(GNOLAND_RPC_URL) \
		-broadcast=true \
		-chainid $(CHAINID) \
		-gas-fee 100000000ugnot \
		-gas-wanted 1000000000 \
		-memo "" \
		gnoswap_admin
	@echo

# Test supplying assets to GNS-WUGNOT market
supply-assets-gns-wugnot:
	$(info ************ Test supplying GNS assets to GNS-WUGNOT market ************)
	# APPROVE FIRST
	@echo "" | gnokey maketx call \
		-pkgpath gno.land/r/gnoswap/v1/gns \
		-func Approve \
		-args $(ADDR_GNOLEND) \
		-args $(MAX_UINT64) \
		-insecure-password-stdin=true \
		-remote $(GNOLAND_RPC_URL) \
		-broadcast=true \
		-chainid $(CHAINID) \
		-gas-fee 100000000ugnot \
		-gas-wanted 1000000000 \
		-memo "" \
		gnoswap_admin
	@echo

	# THEN SUPPLY
	@echo "" | gnokey maketx call \
		-pkgpath gno.land/r/gnolend \
		-func Supply \
		-args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns" \
		-args 1000000 \
		-args 0 \
		-insecure-password-stdin=true \
		-remote $(GNOLAND_RPC_URL) \
		-broadcast=true \
		-chainid $(CHAINID) \
		-gas-fee 100000000ugnot \
		-gas-wanted 1000000000 \
		-memo "" \
		gnoswap_admin
	@echo

# Test supplying shares to GNS-WUGNOT market
supply-shares-gns-wugnot:
	$(info ************ Test supplying shares to GNS-WUGNOT market ************)
	# APPROVE FIRST
	@echo "" | gnokey maketx call \
		-pkgpath gno.land/r/gnoswap/v1/gns \
		-func Approve \
		-args $(ADDR_GNOLEND) \
		-args $(MAX_UINT64) \
		-insecure-password-stdin=true \
		-remote $(GNOLAND_RPC_URL) \
		-broadcast=true \
		-chainid $(CHAINID) \
		-gas-fee 100000000ugnot \
		-gas-wanted 1000000000 \
		-memo "" \
		gnoswap_admin
	@echo

	# THEN SUPPLY
	@echo "" | gnokey maketx call \
		-pkgpath gno.land/r/gnolend \
		-func Supply \
		-args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns" \
		-args 0 \
		-args 1000000 \
		-insecure-password-stdin=true \
		-remote $(GNOLAND_RPC_URL) \
		-broadcast=true \
		-chainid $(CHAINID) \
		-gas-fee 100000000ugnot \
		-gas-wanted 1000000000 \
		-memo "" \
		gnoswap_admin
	@echo

# Test getting user position in GNS-WUGNOT market
get-position-gns-wugnot:
	$(info ************ Test getting user position in GNS-WUGNOT market ************)
	@echo "" | gnokey maketx call \
		-pkgpath gno.land/r/gnolend \
		-func GetPosition \
		-args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns" \
		-args "$(ADMIN)" \
		-insecure-password-stdin=true \
		-remote $(GNOLAND_RPC_URL) \
		-broadcast=true \
		-chainid $(CHAINID) \
		-gas-fee 100000000ugnot \
		-gas-wanted 1000000000 \
		-memo "" \
		gnoswap_admin
	@echo

# Check GNS balance
check-gns-balance:
	$(info ************ Check GNS balance ************)
	@echo "" | gnokey maketx call \
		-pkgpath gno.land/r/gnoswap/v1/gns \
		-func BalanceOf \
		-args $(ADMIN) \
		-insecure-password-stdin=true \
		-remote $(GNOLAND_RPC_URL) \
		-broadcast=true \
		-chainid $(CHAINID) \
		-gas-fee 100000000ugnot \
		-gas-wanted 1000000000 \
		-memo "" \
		gnoswap_admin
	@echo

# Test borrowing GNS tokens
borrow-gns:
	$(info ************ Test borrowing GNS tokens ************)
	@echo "" | gnokey maketx call \
		-pkgpath gno.land/r/gnolend \
		-func Borrow \
		-args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns" \
		-args 500000 \
		-args 0 \
		-insecure-password-stdin=true \
		-remote $(GNOLAND_RPC_URL) \
		-broadcast=true \
		-chainid $(CHAINID) \
		-gas-fee 100000000ugnot \
		-gas-wanted 1000000000 \
		-memo "" \
		gnoswap_admin
	@echo

# Run all market tests
test-markets: market-create-gns-wugnot market-get-gns-wugnot supply-assets-gns-wugnot supply-shares-gns-wugnot get-position-gns-wugnot check-gns-balance borrow-gns
