include _info.mk

# pool create
pool-create-gns-wugnot-default:
	$(info ************ create default pool (GNS:WUGNOT:0.03%) ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/gns -func Approve -args $(ADDR_POOL) -args $(MAX_UINT64) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/pool -func CreatePool -args "gno.land/r/demo/wugnot" -args "gno.land/r/gnoswap/v1/gns" -args 3000 -args 79228162514264337593543950337 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo


# mint new position
mint-gns-gnot:
	$(info ************ mint position(1) to gns:wugnot ************)
	# APPROVE FISRT
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/gns -func Approve -args $(ADDR_POOL) -args $(MAX_UINT64) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo "" | gnokey maketx call -pkgpath gno.land/r/demo/wugnot -func Approve -args $(ADDR_POOL) -args $(MAX_UINT64) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

	# APPROVE WUGNOT TO POSITION, to get refund wugnot left after wrap -> mint
	@echo "" | gnokey maketx call -pkgpath gno.land/r/demo/wugnot -func Approve -args $(ADDR_POSITION) -args $(MAX_UINT64) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

	# THEN MINT
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/position -func Mint -send "20000000ugnot" -args "gno.land/r/gnoswap/v1/gns" -args "gnot" -args 3000 -args "-49980" -args "49980" -args 20000000 -args 20000000 -args 1 -args 1 -args $(TX_EXPIRE) -args $(ADDR_GNOSWAP) -args $(ADDR_GNOSWAP) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

	# SetTokenURI
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/gnft -func SetTokenURIByImageURI -args "1" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo
	

# increase liquidity
increase-liquidity-position-01:
	$(info ************ increase position(1) liquidity gnot:gns:3000 ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/position -func IncreaseLiquidity -send "20000000ugnot" -args 1 -args 20000000 -args 20000000 -args 1 -args 1 -args $(TX_EXPIRE) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo


# decrease liquidity
decrease-liquidity-position-01:
	$(info ************ decrease position(1) liquidity gnot:gns:3000 ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/position -func DecreaseLiquidity -args 1 -args 12345678 -args 0 -args 0 -args $(TX_EXPIRE) -args "false" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo


# create external incentive
create-external-incentive:
	$(info ************ create external incentive [gns] => gnot:gns:3000 ************)
	# APPROVE REWARD (+ DepositGNS)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/gns -func Approve -args $(ADDR_STAKER) -args $(MAX_UINT64) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

	# THEN CREATE EXTERNAL INCENTIVE
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/staker -func CreateExternalIncentive -args "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000" -args "gno.land/r/gnoswap/v1/gns" -args 1000000000 -args $(TOMORROW_MIDNIGHT) -args $(INCENTIVE_END) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo


# stake token
stake-token-1:
	$(info ************ stake token 1  ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/gnft -func Approve -args $(ADDR_STAKER) -args 1 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/staker -func StakeToken -args 1 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo


# collect staking reward
collect-staking-reward-1:
	$(info ************ collect reward 1  ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/staker -func CollectReward -args 1 -args false -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo


# unstake token
unstake-token-1:
	$(info ************ unstake token 1  ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/staker -func UnStakeToken -args 1 -args true -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo


# swap exact in
swap-exact-in-gns-wugnot:
	$(info ************ swap gns -> wgnot, exact_in ************)
	# approve INPUT TOKEN to POOL
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/gns -func Approve -args $(ADDR_POOL) -args $(MAX_UINT64) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin

	# approve OUTPUT TOKEN to ROUTER ( as 0.15% fee )
	@echo "" | gnokey maketx call -pkgpath gno.land/r/demo/wugnot -func Approve -args $(ADDR_ROUTER) -args $(MAX_UINT64) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin

	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/router -func ExactInSwapRoute -args "gno.land/r/gnoswap/v1/gns" -args "gno.land/r/demo/wugnot" -args 50000 -args "gno.land/r/gnoswap/v1/gns:gno.land/r/demo/wugnot:3000" -args "100" -args "0" -args $(TX_EXPIRE) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo


# swap exact out
swap-exact-out-gns-wugnot:
	$(info ************ swap gns -> wgnot, exact_out ************)
	# approve INPUT TOKEN to POOL
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/gns -func Approve -args $(ADDR_POOL) -args $(MAX_UINT64) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin

	# approve OUTPUT TOKEN to ROUTER ( as 0.15% fee )
	@echo "" | gnokey maketx call -pkgpath gno.land/r/demo/wugnot -func Approve -args $(ADDR_ROUTER) -args $(MAX_UINT64) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin

	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/router -func ExactOutSwapRoute -args "gno.land/r/gnoswap/v1/gns" -args "gno.land/r/demo/wugnot" -args 50000 -args "gno.land/r/gnoswap/v1/gns:gno.land/r/demo/wugnot:3000" -args "100" -args "60000" -args $(TX_EXPIRE) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo


# collect swap fee
collect-swap-fee:
	$(info ************ collect swap fee ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/position -func CollectFee -args 1 -args false -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo


# delegate gns
delegate:
	$(info ************ delegate 5_000_000_000 to self ************)
	# APPROVE FIRST
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/gns -func Approve -args $(ADDR_GOV_STAKER) -args $(MAX_UINT64) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin

	# DELEGATE
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/gov/staker -func Delegate -args $(ADDR_GNOSWAP) -args 5000000000 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo

# redelegate gns
redelegate:
	$(info ************ redelegate 1_000_000_000 from self to self ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/gov/staker -func Redelegate -args $(ADDR_GNOSWAP) -args $(ADDR_GNOSWAP) -args 1000000000 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo


# undelegate gns
undelegate:
	$(info ************ undelegate 1_000_000_000 from self ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/gov/staker -func Undelegate -args $(ADDR_GNOSWAP) -args 1000000000 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo


# collect undelegated gns
collect-undelegated:
	$(info ************ collect undelegated gns ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/gov/staker -func CollectUndelegatedGns -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo


# propose text proposal
propose-text:
	$(info ************ propose text ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/gov/governance -func ProposeText -args "title_for_text" -args "desc_for_text" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo


# cancel proposal
cancel-text:
	$(info ************ cancel text ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/gov/governance -func Cancel -args 1 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo


# propose community_pool send proposal
propose-community:
	$(info ************ propose community pool spend ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/gov/governance -func ProposeCommunityPoolSpend -args "title_for_spend" -args "desc_for_spend" -args $(ADDR_GNOSWAP) -args "gno.land/r/gnoswap/v1/gns" -args 1 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo


# vote proposal
vote-community:
	$(info ************ vote community pool spend ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/gov/governance -func Vote -args 2 -args true -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo


# execute proposal
execute-community:
	$(info ************ execute community pool spend ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/gov/governance -func Execute -args 2 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo


# propose parameter change proposal
propose-param:
	$(info ************ propose param change ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/gov/governance -func ProposeParameterChange -args "title param change" -args "desc param change" -args "2" -args "gno.land/r/gnoswap/v1/gns*EXE*SetAvgBlockTimeInMs*EXE*123*GOV*gno.land/r/gnoswap/v1/community_pool*EXE*TransferToken*EXE*gno.land/r/gnoswap/v1/gns,g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d,905" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo


# create launchpad project
create-launchpad-project:
	$(info ************ create bar project ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/launchpad -func CreateProject -args "OBL PROJECT" -args "gno.land/r/gnoswap/v1/test_token/obl" -args $(ADDR_GNOSWAP) -args 1000000000000 -args "gno.land/r/gnoswap/v1/gns" -args "10" -args 50 -args 30 -args 20 -args $(TOMORROW_MIDNIGHT) -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo


# deposit to project
deposit-to-project:
	$(info ************ deposit to project ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/launchpad -func DepositGns -args "gno.land/r/gnoswap/v1/test_token/obl:4215:30" -args 1000000 -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo


# collect project token
collect-project-token:
	$(info ************ collect project token ************)
	@echo "" | gnokey maketx call -pkgpath gno.land/r/gnoswap/v1/launchpad -func CollectRewardByProjectId -args "gno.land/r/gnoswap/v1/test_token/obl:4215" -insecure-password-stdin=true -remote $(GNOLAND_RPC_URL) -broadcast=true -chainid $(CHAINID) -gas-fee 100000000ugnot -gas-wanted 1000000000 -memo "" gnoswap_admin
	@echo