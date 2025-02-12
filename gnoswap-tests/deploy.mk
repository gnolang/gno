include _info.mk

## INIT
.PHONY: init
init: send-ugnot-must deploy-test-tokens deploy-libraries deploy-consts deploy-common deploy-base-tokens deploy-gnoswap-realms # pool-create-gns-wugnot-default

.PHONY: deploy-test-tokens
deploy-test-tokens: deploy-usdc deploy-foo deploy-bar deploy-baz deploy-qux deploy-obl 

.PHONY: deploy-libraries
deploy-libraries: deploy-uint256 deploy-int256 deploy-math

.PHONY: deploy-base-tokens
deploy-base-tokens: deploy-gns deploy-gnft

.PHONY: deploy-gnoswap-realms
deploy-gnoswap-realms: deploy-gov-xgns deploy-emission deploy-protocol-fee deploy-pool deploy-position deploy-router deploy-staker deploy-community-pool deploy-gov-staker deploy-gov-governance deploy-launchpad deploy-referral


# send ugnot to necessary accounts
send-ugnot-must:
	$(info ************ send ugnot to necessary accounts ************)
	@echo "" | gnokey maketx send -send 10000000000ugnot -to $(ADDR_GNOSWAP) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" test1
	@echo

# deploy packages
deploy-uint256:
	$(info ************ deploy uint256 ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/p/gnoswap/uint256 -pkgpath gno.land/p/gnoswap/uint256 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo

deploy-int256:
	$(info ************ deploy int256 ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/p/gnoswap/int256 -pkgpath gno.land/p/gnoswap/int256 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo

deploy-math:
	$(info ************ deploy package gnsmath ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/p/gnoswap/gnsmath -pkgpath gno.land/p/gnoswap/gnsmath -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo

deploy-consts:
	$(info ************ deploy consts ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/p/gnoswap/consts -pkgpath gno.land/p/gnoswap/consts -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo

deploy-common:
	$(info ************ deploy common ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/r/gnoswap/common -pkgpath gno.land/r/gnoswap/v1/common -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo

deploy-gns:
	$(info ************ deploy gns ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/r/gnoswap/gns -pkgpath gno.land/r/gnoswap/v1/gns -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo

deploy-gnft:
	$(info ************ deploy gnft ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/r/gnoswap/gnft -pkgpath gno.land/r/gnoswap/v1/gnft -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo

deploy-gov-xgns:
	$(info ************ deploy gov/xgns ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/r/gnoswap/gov/xgns -pkgpath gno.land/r/gnoswap/v1/gov/xgns -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo
	
deploy-emission:
	$(info ************ deploy emission ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/r/gnoswap/emission -pkgpath gno.land/r/gnoswap/v1/emission -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo

deploy-protocol-fee:
	$(info ************ deploy protocol fee ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/r/gnoswap/protocol_fee -pkgpath gno.land/r/gnoswap/v1/protocol_fee -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo

deploy-pool:
	$(info ************ deploy pool ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/r/gnoswap/pool -pkgpath gno.land/r/gnoswap/v1/pool -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo

deploy-position:
	$(info ************ deploy position ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/r/gnoswap/position -pkgpath gno.land/r/gnoswap/v1/position -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo

deploy-router:
	$(info ************ deploy router ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/r/gnoswap/router -pkgpath gno.land/r/gnoswap/v1/router -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo

deploy-staker:
	$(info ************ deploy staker ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/r/gnoswap/staker -pkgpath gno.land/r/gnoswap/v1/staker -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo

deploy-community-pool:
	$(info ************ deploy community pool ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/r/gnoswap/community_pool -pkgpath gno.land/r/gnoswap/v1/community_pool -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo

deploy-gov-staker:
	$(info ************ deploy gov/staker ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/r/gnoswap/gov/staker -pkgpath gno.land/r/gnoswap/v1/gov/staker -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo

deploy-gov-governance:
	$(info ************ deploy gov/governance ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/r/gnoswap/gov/governance -pkgpath gno.land/r/gnoswap/v1/gov/governance -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo

deploy-launchpad:
	$(info ************ deploy launchpad ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/r/gnoswap/launchpad -pkgpath gno.land/r/gnoswap/v1/launchpad -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo

deploy-referral:
	$(info ************ deploy referral ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/r/gnoswap/referral -pkgpath gno.land/r/gnoswap/v1/referral -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo

deploy-usdc:
	$(info ************ deploy usdc ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/r/gnoswap/test_token/usdc -pkgpath gno.land/r/gnoswap/v1/test_token/usdc -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo

deploy-foo:
	$(info ************ deploy foo ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/r/gnoswap/test_token/foo -pkgpath gno.land/r/gnoswap/v1/test_token/foo -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo

deploy-bar:
	$(info ************ deploy bar ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/r/gnoswap/test_token/bar -pkgpath gno.land/r/gnoswap/v1/test_token/bar -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo

deploy-baz:
	$(info ************ deploy baz ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/r/gnoswap/test_token/baz -pkgpath gno.land/r/gnoswap/v1/test_token/baz -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo

deploy-qux:
	$(info ************ deploy qux ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/r/gnoswap/test_token/qux -pkgpath gno.land/r/gnoswap/v1/test_token/qux -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo

deploy-obl:
	$(info ************ deploy obl ************)
	@echo "" | gnokey maketx addpkg -pkgdir $(ROOT_DIR)/contract/r/gnoswap/test_token/obl -pkgpath gno.land/r/gnoswap/v1/test_token/obl -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 100000000 -memo "" gnoswap_admin
	@echo
