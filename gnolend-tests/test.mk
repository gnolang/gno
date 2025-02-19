# Network configuration
GNOLAND_RPC_URL=http://localhost:26657
CHAINID=dev

ADDR_GNOSWAP_ADMIN := g1mgk7hmna0w9ku7kynllmkukg573gxs9tclfqx3


# Test market creation with GNS and WUGNOT
market-create-gns-wugnot:
	$(info ************ Test creating market with GNS and WUGNOT ************)
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
	$(info ************ Test supplying assets to GNS-WUGNOT market ************)
	@echo "" | gnokey maketx call \
		-pkgpath gno.land/r/gnolend \
		-func Supply \
		-args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns" \
		-args 1000000000 \
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
		-args "$(ADDR_GNOSWAP_ADMIN)" \
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
test-markets: market-create-gns-wugnot market-get-gns-wugnot supply-assets-gns-wugnot supply-shares-gns-wugnot get-position-gns-wugnot
