#!/usr/bin/env bash
# gen-genesis.sh — mainnet genesis builder (single-file pipeline).
#
# mainnet is a FRESH chain — no hardfork, no historical replay. This
# script builds the whole genesis from the repo's examples/ tree plus a
# handful of bootstrap txs, in minutes.
#
# What the genesis contains:
#
#   1. The FILTERED_PACKAGES example set (resolved with transitive deps),
#      addpkg'd by the deterministic GenesisDeployer key.
#   2. A bootstrap MsgRun (transactions/base/bootstrap/) that seeds the
#      sole GovDAO T1 member (aeddi) and locks dao.UpdateImpl's AllowedDAOs to
#      r/gov/dao/impl/v0. Transfers are locked at genesis per §126, with
#      the independence-day exemption list applied (step 9.3). A second
#      MsgRun (transactions/base/users-preregister/) registers the initial
#      mainnet namespaces in r/sys/users via the genesis-only path.
#   3. A names.Enable MsgCall (transactions/migration/names-enable/) so
#      namespace enforcement is on from genesis. Enable is gated on the
#      admin address hardcoded in r/sys/names/verifier.gno; the tx's
#      caller field is jq-patched to that address post-sign, which the
#      chain trusts under --skip-genesis-sig-verification.
#   4. Per-validator valopers.Register MsgCalls (emitted by `gnogenesis fork
#      valoper-seed` from INITIAL_VALSET + INITIAL_VALSET_OPERATORS) so
#      the founding validators have operator-keyed valoper profiles and
#      r/sys/validators/v0 can manage the set post-genesis.
#   5. The INITIAL_VALSET as GenesisDoc.Validators (InitChainer seeds
#      valset:current from it, so v0/EndBlocker valset changes work).
#   6. Balances: the independence-day allocation sheet (~3.26M accounts,
#      downloaded by pinned URL + sha256-verified) and its §126
#      unrestricted-address list (same treatment, same pinned commit), the
#      VESTED_ACCOUNTS entries (created as vesting accounts at genesis —
#      though the pinned sheet now carries the §132 schedules itself), plus
#      exact-burn funding for every genesis-tx fee payer (measured on a
#      temp node; fee payers land at zero — or at exactly their allocation
#      if they also hold one — once the genesis txs execute).
#
# Output:
#   work/packages.gen.txt    resolved package list (audit artifact)
#   work/genesis_txs.jsonl   full genesis tx stream (audit artifact)
#   work/valoper-seed.jsonl  valoper Register txs (audit artifact)
#   genesis.json             final artifact, sha256-locked against the
#                            CHECKSUMS_DATA heredoc in this script
#
# Usage:
#   ./gen-genesis.sh                # full build
#   ./gen-genesis.sh --debug        # echo the main pipeline commands
#   ./gen-genesis.sh --no-install   # reuse previously built binaries
#
# Cross-platform: bash 3.2 minimum (macOS default), no GNU-only features.

# -u is deliberately absent: on bash 3.2 (macOS default, the floor this script
# targets) expanding an EMPTY array with "${arr[@]}" errors under -u, and both
# VESTED_ACCOUNTS and txn_dir_to_jsonl's args_array are legitimately empty.
set -eo pipefail

# =============================================================================
# Launch parameters — review before each genesis generation.
# =============================================================================

CHAIN_ID=gnoland-1 # decided 2026-09-09
# TODO(mainnet): launch time undecided — placeholder is pearl's launch time.
# Mainnet block 1 carries this timestamp forever: pin the ceremony time and
# rebuild if it slips (topaz/sapphire/pearl all launched backdated; fine on
# a testnet, ugly on mainnet).
#
# This is no longer a free parameter: since independence-day #72 the pinned
# sheet vests nearly every row continuously from an ABSOLUTE 1789084800
# (2026-09-11T00:00:00Z), which mkgenesis/vesting.go documents as *the genesis
# timestamp*. GENESIS_TIME > that start is a §126/§132 leak and
# assert_vesting_locked_at_genesis refuses to build it; GENESIS_TIME < it (the
# case today, by 15 days) ships a chain whose §132 clock starts after launch —
# valid on-chain (VestedCoins returns nothing before StartTime) but not what
# §132 says. Decide the ceremony time and re-pin a sheet generated with
# -vesting-start equal to it.
GENESIS_TIME=1787817600 # Thursday, August 27th 2026 10:00 CEST (08:00 UTC)

# Packages to include in genesis (resolved with transitive dependencies).
# Use "..." suffix to match all sub-packages.
#
# First seven lines mirror gnoland1's gen-genesis.sh FILTERED_PACKAGES. The
# last block is additions carried over from test13:
#   - p/onbloc/{uint256,int256,json}/v0: used by realms we want available
#     (uint256 is a transitive dep of int256; versioned under v0 since
#     #6159).
#   - r/sys/validators/v0: the valset realm — already matched by the
#     ./gno.land/r/sys/... pattern, kept explicit because it is load-
#     bearing: the node's EndBlocker reads valset state from this realm's
#     params; without it on chain, post-genesis valset changes can't
#     happen.
#   - r/nt/grc20reg/v0: GRC20 token registry.
#   - p/nt/grc20/v0 and p/nt/grc721/...: the token standards, listed
#     explicitly instead of being left to arrive as transitive deps — see
#     the namespace note below. Checked with `gno tool deplist -test-dep`
#     on the set above: grc20 already arrives (via r/nt/grc20reg/v0), but
#     no grc721 package arrives at all, and grc721 plus its enumerable/
#     metadata/royalty extensions is what every future NFT realm needs.
#     The quarantined grc1155/grc777 are deliberately NOT here: they are
#     not deployable until they graduate.
#
# The set is FINAL as re-curated by #6166 (v0 repaths + the token
# standards); #6162 is merged, so it resolves. Keep in mind when touching it:
# p/nt/* paths cannot be added post-genesis under namespace enforcement,
# so anything missing here is a hardfork away (r/tests/* and r/demo/*
# arrive via test deps of this set).
FILTERED_PACKAGES=(
  ./gno.land/r/sys/...
  ./gno.land/r/gov/...
  ./gno.land/r/gnoland/blog/...
  ./gno.land/r/gnoland/wugnot/...
  ./gno.land/r/gnoland/coins/...
  ./gno.land/r/gnoland/boards2/...
  ./gno.land/r/gnops/valopers/...
  ./gno.land/p/onbloc/uint256/v0
  ./gno.land/p/onbloc/int256/v0
  ./gno.land/p/onbloc/json/v0
  ./gno.land/r/sys/validators/v0
  ./gno.land/r/nt/grc20reg/v0
  ./gno.land/p/nt/grc20/v0
  ./gno.land/p/nt/grc721/...
)

# Initial mainnet validator set. Format: "name power address pub_key".
# Power 60 each — 60 divides evenly many ways (2, 3, 4, 5, 6, 10, ...),
# so later valset changes can hand out proportional fractions of the
# founders' voting power without fractional remainders.
#
# Four founding validators at equal power: one dark loses a quarter of the
# voting power — below the one-third halt boundary, so any single failure
# keeps the chain live (unlike the 3-validator testnet bootstraps).
#
# All four entries are REAL: each org's ceremony pair, received 2026-09-10/11
# (addresses cross-checked against the pubkeys by deriving them). Names
# follow the house <org>-validator-<n> form regardless of the moniker each
# org proposed. How each org protects its signing key (tmkms, horcrux,
# HSM) is its own infra, outside this genesis.
INITIAL_VALSET=(
  "gno-core-validator-1 60 g1mmgvcssjw6x4fzphupfg6mtxqt36v000c5rf2a gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqzlmd56sl6gyam3u4sht0fnpvddxl2cf6rdx76rnz7ufcsz8lyh073fcxm"
  "onbloc-validator-1 60 g1hqhetnnz0raw5hps6yxexl7q09a6f8w3anlptt gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zp95ay40jg2pc0cpn432z5fny390tfpeycf4hf8msznu9tthulswqhwtucx"
  "samourai-crew-validator-1 60 g15t7f9q6km3ldt885duwl8xu5dncs98528amk4f gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpdkjpdaqsm9gxnw90ac6a78gvnquttl4kr64zxughl87sek7xgx60gfntd"
  "berty-validator-1 60 g1l983yy3kpmapyzcfy53y5charfxupa5czjalea gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqn8u6cc4dgrzu9u8hztvlmqrvqt7mzju577xjdxjqxf8m5gjypaz798f5d"
)

# Operator address for each INITIAL_VALSET entry (same index). MUST be
# distinct from the signing address — `gnogenesis fork valoper-seed`
# rejects operator==signing_addr to keep signing-key compromise from
# collapsing into operator-slot compromise (see valoper_seed.go).
#
# The operator key is the management plane for the validator: whoever
# holds it can rotate the signing key, edit the valoper profile, and
# signal opt-out via r/gnops/valopers + r/sys/validators/v0.
#
# Valoper profiles are keyed on the operator address and `fork
# valoper-seed` rejects duplicate operators, so all slots must be
# distinct addresses.
#
# All four slots are REAL. gno-core's operator is aeddi's operational key
# — the same address as his GovDAO T1 seat (deliberate).
#
# FUNDING IS RESOLVED (independence-day#78, in the pin below): all eight
# addresses — the four signing addresses above and the four operators here —
# hold 1,000 GNOT at genesis, charged to the §122 Validator Services Treasury.
# Rotating a signing key, editing a valoper profile and signalling opt-out are
# all paid txs; there is no faucet, and §126 leaves only the exemption-listed
# funds able to send, so an address that lands at zero is stuck there forever.
# The rows carry the standard §132 schedule, which does not impede them: fees
# are collected with SendCoinsUnrestricted, which bypasses both the vesting
# lock and the transfer restriction. Step 2.8's guard asserts it on the pinned
# sheet, so a re-pin that drops the float fails the build.
#
# The operators are NOT genesis fee payers — the deployer pays the Register
# txs — so each simply keeps its 1,000 GNOT.
#
# Every address that must act post-genesis is now funded upstream: the
# validator addresses (#78), the approvals oracle (#79/#80), and the
# blog/boards owner multisig (#81) — see NAMES_ADMIN below.
INITIAL_VALSET_OPERATORS=(
  "g1aeddlftlfk27ret5rf750d7w5dume3kcsm8r8m" # gno-core-validator-1 operator (aeddi)
  "g12gtvlcexzgax49nvvkvhp2u0v6eejhunq0074p" # onbloc-validator-1 operator
  "g1n9y62agq998jt8w59az60xcqlftjknjg2grhn4" # samourai-crew-validator-1 operator
  "g1qynsu9dwj9lq0m5fkje7jh6qy3md80ztqnshhm" # berty-validator-1 operator
)

# Genesis allocation (no faucets on mainnet): the gnolang/independence-day
# balance sheet — 3,262,481 accounts totalling 1,332,999,998.328067 GNOT
# (1.333e15 ugnot, ~6900x under the int64 Coin ceiling of ~9.22e18).
# Downloaded by pinned-commit URL and verified against ALLOCATION_SHA256
# before use (the gnoland1/test13 pattern). The sheet is the "mkgenesis/
# balances.txt.gz" public-contract path of that repo.
#
# Pinned at independence-day main @ 0108ede (#81, "fund the GovDAO multisig
# with a 5,000 GNOT realm-operation float"). Relative to the 91f7f56 pin this
# build was last verified on, the TOTAL is unchanged and the account count is
# +8, across five commits:
#
#   #77  e37caa6^  a SHA256SUMS manifest asserted in that repo's CI, so the
#                  published container can no longer change silently under a
#                  non-GNU gzip. Artifacts unchanged.
#   #78  e37caa6   +6 accounts. All eight founding-validator addresses (four
#                  INITIAL_VALSET signing + four INITIAL_VALSET_OPERATORS) hold
#                  1,000 GNOT, charged OUT of the §122 Validator Services
#                  Treasury (20,000,000 -> 19,994,000). Closes the funding TODO
#                  on INITIAL_VALSET_OPERATORS above; asserted at step 2.8.
#   #79  64f0848   +1 account. The inert-package approvals oracle
#                  g1yaaa6rcp4ew5yjzdj4yms596wx2dtrj3a86704 is funded out of
#                  the §120 Core Treasury. See PKG_APPROVERS below.
#   #80  0927710   no account change — the SAME oracle row at 5,000 GNOT
#                  instead of 1,000, because a service spends per event where
#                  a founder or an operator spends occasionally. Core goes
#                  39,993,000 -> 39,989,000.
#   #81  0108ede   +1 account. NAMES_ADMIN g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx
#                  is funded at the same 5,000 GNOT tier, out of §120. NOT for
#                  names administration — r/sys/names' admin gates Enable()
#                  and nothing else, and Enable is one-shot at genesis. It is
#                  funded because the same address is the hardcoded owner of
#                  r/gnoland/blog and r/gnoland/boards2/v0, both in this
#                  genesis set, where every owner action is a paid tx and the
#                  owner cannot be reassigned without a realm upgrade. Core
#                  goes 39,989,000 -> 39,984,000. See NAMES_ADMIN below.
#
# Every float is charged out of an existing bucket rather than minted, so the
# cap does not move and ALLOCATION_EXPECTED_TOTAL is unchanged.
#
# All the new rows carry the standard §132 schedule, which does not impede
# them: fees are collected with SendCoinsUnrestricted, bypassing both the
# vesting lock and the §126 transfer restriction.
#
# TODO(mainnet): re-pin URL + sha to the FINAL independence-day commit at
# genesis cut (main moves as sale participants bind addresses; see its
# docs/history.md convention of recording which commit produced which
# chain).
ALLOCATION_GZ_URL="https://github.com/gnolang/independence-day/raw/0108ede228044557aaa3bf17db245533979f5498/mkgenesis/balances.txt.gz"
ALLOCATION_SHA256="f78673663f8d5045c402a4bb90039e4b51d308e21acc17c74e6e86e36b09ed08"
# The sha proves "this is the pinned file"; these two make its MAGNITUDE part
# of the reviewed diff. Every downstream reconciliation is internal (sheet ↔
# artifact) and would hold for any sheet — without these, a re-pin that
# changes how much money mainnet starts with is absorbed by the arithmetic
# instead of appearing as a reviewable change. Update them with every re-pin.
ALLOCATION_EXPECTED_ACCOUNTS=3262481
ALLOCATION_EXPECTED_TOTAL=1332999998328067 # ugnot ≈ 1.333e9 GNOT

# Unrestricted addresses (Constitution §126-130). Same repo, same
# fetch-and-verify treatment as the allocation sheet.
#
#   §126  "$GNOT will not be transferrable initially except for whitelisted
#          addresses. Whitelisted addresses include "Ecosystem" and "Investors"
#          funds and any additional addresses needed for the operation of the
#          chain, and funding needs or payment of investors. Whitelisted funds
#          remain subject to the vesting schedule below."
#
# 91 addresses: the Ecosystem Treasury (1), both Investors tranches (2), every
# public-sale row (79, from publicsale.txt) and every settled investor/partner
# distribution (9, from investors.txt — new in #73). The list is GENERATED in
# independence-day from the same inputs that produce the balance rows, so a
# participant cannot be funded in the genesis but left unable to move it.
#
# Pinned at the SAME commit as ALLOCATION_GZ_URL above. Keep the two pins equal
# on every re-pin; a sheet and an exemption list from different commits is
# exactly the mismatch pinning exists to prevent. The sha is unchanged across
# this re-pin, and should be: #78 and #79 fund validator operators and a service
# oracle, and neither is one of the §127 funds — they pay FEES, which
# SendCoinsUnrestricted exempts from §126 anyway, so neither needs a whitelist
# entry. A float appearing on this list would be the thing to question.
UNRESTRICTED_URL="https://github.com/gnolang/independence-day/raw/0108ede228044557aaa3bf17db245533979f5498/mkgenesis/unrestricted.txt"
UNRESTRICTED_SHA256="7b37a16822739371cfd9a3f5ae864b7ab86cef4c83155fefb5114c29abda38bf"

# Denominations subject to the §126 transfer lock. Empty = no lock, and then
# UNRESTRICTED_ADDRS is inert: bank.canSendCoins returns true before it ever
# looks at the whitelist. Both knobs are required for §126 to mean anything.
RESTRICTED_DENOMS=("ugnot")

# Vested accounts. One entry per line, in the balance-sheet vesting syntax
# (gno.land/pkg/gnoland/balance.go):
#
#   <address>=<total_coins>;vesting=<vested_coins>,<start_unix>,<end_unix>[;type=delayed]
#
# The default schedule is continuous: the vested amount unlocks linearly
# between <start_unix> and <end_unix>. `;type=delayed` makes it a cliff —
# nothing unlocks before <end_unix>, everything at once after. The vested
# coins must be <= the total, and the difference is spendable immediately.
#
# This array is now mostly REDUNDANT: as of the 0108ede pin the sheet itself
# carries the §132 schedules (independence-day #72 turned the vesting pass on
# by default), including the investors-vesting multisig
# g1x7tm26g9wj84cmg3cs74uwf3g9lqj4mjp6gax3 — 150,000,000 GNOT total, 144,000,000
# of it vesting continuously over 1789084800→1852243200 (24 months, 4% unlocked
# at start). Adding it here too would only trip the overlap guard. Prefer fixing
# a schedule upstream in the sheet over typing one here.
#
# DECIDED: nothing needs a hand-typed entry — the pinned sheet carries every
# schedule (§132 rows, the investors bucket, the treasuries hold multisig-only
# custody). The array stays as an escape hatch; an entry whose address also
# carries a sheet schedule is rejected outright by the overlap guard in step 8
# (sheet wins — fix a schedule upstream in the sheet, not here).
VESTED_ACCOUNTS=(
)

# Token-transfer policy: LOCKED at genesis, per Constitution §126 —
# "$GNOT will not be transferrable initially except for whitelisted
# addresses". That is a requirement, not a launch preference, so this is
# no longer a TODO. gnoland1 launched the same way (locked bank +
# unrestricted-accounts exemptions); only the fresh testnets launched
# open, and they have no Constitution to answer to.
#
# Launching open would also make §132 vacuous: the vesting schedule is
# anchored to "the day $GNOT becomes transferrable", which under an open
# launch is block 1 — so 100% would be liquid where 96% should be locked.
#
# RESTRICTED_DENOMS + the fetched exemption list are applied in step 9.3.
# TODO(mainnet): OnBloc/AiB still need to be told, since the launch
# checklist recorded this as open.

# ---- Inert code-submission policy (decided) ----
#
# mainnet launches under "inert": a post-genesis MsgAddPackage is STORED, not
# executed, and becomes live only when an address in PKG_APPROVERS sends
# MsgEnablePackage for the exact bytes it reviewed. Genesis replay is exempt
# (auth.IsGenesisReplay), so the 85 genesis packages still execute at block 1.
#
# Ported from the pearl builder (#6096), with three mainnet-specific
# differences, each enforced by a guard in step 2.9 rather than a comment:
#   1. the submission charge is OFF, because it is incompatible with §126;
#   2. approvers are funded from the allocation sheet, not minted here;
#   3. RUN_SUBMITTERS must cover every GovDAO T1 member, not just one seed.
CODE_SUBMISSION_POLICY=inert

# Addresses permitted to send MsgEnablePackage.
#
# Load-bearing: with no approver, every submission parks forever and the chain
# accepts deploys it can never activate (vm.enableBlockedReason ->
# ReasonNoApprovers). Nothing on chain refuses that state — Params.Validate
# checks address syntax only, and InitChain logs nothing — so step 2.9 refuses
# to build it.
#
# Keep this to the gpao oracle key and nothing else. An approver can activate
# any parked package, so it is the second most consequential key on the chain
# after governance, and it lives unattended on an internet-facing daemon.
#
# FUNDING is already resolved: independence-day#79 gives this address a genesis
# balance out of the §120 Core Treasury, raised to 5,000 GNOT by #80 — at
# contribs/gpao's default 1000000ugnot fee, ~5,000 approvals, with gpao's own
# per-run max-spend bound (100 GNOT) a fiftieth of the float. Five times the
# 1,000 tier the founders and validator operators sit at, deliberately: they
# pay for occasional management actions, this key pays per EVENT for as long as
# the chain accepts packages, and §126 leaves no way to top it up short of the
# transfer lock lifting or a GovDAO proposal.
#
# That had to land in the sheet rather than wait for this decision — with no
# faucet and §126 in force, an approver seeded at zero could never be funded
# afterwards, and would approve nothing forever. Step 2.9 re-checks it against
# the pinned sheet.
PKG_APPROVERS=(
  g1yaaa6rcp4ew5yjzdj4yms596wx2dtrj3a86704 # gpao approval oracle
)

# Addresses permitted to send MsgRun. EMPTY MEANS OPEN — the param's zero value
# is "off", so leaving this empty lets anyone execute arbitrary source on a
# chain whose whole submission policy exists to stop exactly that. "inert"
# gates package PERSISTENCE; this is the gate on EXECUTION, and only both
# together give the property the policy is named for.
#
# GovDAO proposal creation is MsgRun-only (a ProposalRequest carries a func
# value, which MsgCall cannot marshal), so every T1 member must appear here or
# governance is unreachable AND the list can never be amended.
#
# DERIVED, not typed: the members are read out of the bootstrap tx at 2.7 and
# unioned in at 2.9, for the reason the T1 funding guard gives — a second copy
# here would drift from what actually gets seeded, and the direction it drifts
# is silent, since a member dropped from this list simply cannot govern.
RUN_SUBMITTERS_INCLUDE_GOVDAO_T1=true

# Addresses that must `maketx run` WITHOUT being GovDAO members. Empty is the
# expected state; add an operator here only with a reason, since MsgRun
# executes arbitrary source.
RUN_SUBMITTERS_EXTRA=()

# OFF, and it has to be: the charge moves through the bank keeper's RESTRICTED
# SendCoins (gno.land/pkg/sdk/vm/keeper.go, "a token lock should refuse it, not
# be bypassed"), while mainnet ships restricted_denoms=["ugnot"] with the §126
# exemption list. A non-zero charge would therefore make MsgAddPackage
# unpayable for every address that is not exemption-listed — turning
# "submissions are reviewed" into "nobody may submit".
#
# The cost of leaving it off is that a parked package is permanent state paid
# for by gas alone: AddInertPackage takes no storage deposit (there are no
# realm diffs to price until enable), and nothing expires a parked blob —
# only its creator or an approver can clear it with MsgRejectPackage.
#
# Revisit only together with §126: if the transfer lock is lifted, turn the
# charge on in the same proposal.
INERT_SUBMISSION_CHARGE=
INERT_CHARGE_COLLECTOR=

# =============================================================================
# Internal — everything below is glue, you shouldn't need to change it.
# =============================================================================

# Deployer key mnemonic (deterministic — used only to sign genesis-mode txs).
# Same as gnoland1/test13/topaz so the deployer address is reproducible.
DEPLOYER_MNEMONIC="anchor hurt name seed oak spread anchor filter lesson shaft wasp home improve text behind toe segment lamp turn marriage female royal twice wealth"
DEPLOYER_KEY=GenesisDeployer
# Address derived from DEPLOYER_MNEMONIC. Used as the fee payer for the
# valoper-seed Register txs; the balance-measurement step funds it exactly.
DEPLOYER_ADDR=g1edq4dugw0sgat4zxcw9xardvuydqf6cgleuc8p

# r/sys/names admin: hardcoded in examples/gno.land/r/sys/names/verifier.gno
# — since #6131 this is the gnolang/multisigs [govdao] 4-of-7 multisig
# (aeddi, dongwon, howl, jaekwon, maxwell, milos, moul), i.e. the mainnet
# GovDAO T1 multisig. names.Enable's admin check reads
# runtime.PreviousRealm().Address(); under --skip-genesis-sig-verification,
# the chain trusts the MsgCall.Caller field as the EOA, so jq-patching
# caller to this address makes Enable's gate pass. The private key is not
# needed at build time (the multisig exists and can sign post-genesis).
#
# TODO(mainnet): (in progress — Manfred) before the package set is frozen, audit every other
# hardcoded g1... literal in the deployed set the same way (hardcoded
# addresses are unchangeable post-genesis without a realm upgrade — see
# the launch checklist, authority-and-keys section). Package CREATORS are a
# second, separate mechanism to audit: `gnogenesis txs add packages` honours
# an `[addpkg] creator` in each package's gnomod.toml over -key-name, and 35
# packages in the current set declare one across 7 contributor addresses —
# all of r/sys/* under one, all of r/gov/dao* under another. No realm in the
# set captures its deployer as owner (checked: r/gnoland/blog hardcodes the
# GovDAO multisig, r/gnops/valopers keys on the registering operator), so
# this is about who appears as mainnet's deployer of record and who holds
# the namespaces, not about hidden authority.
#
# FUNDED: 5,000 GNOT from the §120 Core Treasury (independence-day#81), the
# same tier as the approvals oracle. The reason is NOT names administration —
# the earlier TODO here had that wrong. r/sys/names' admin gates Enable() and
# nothing else, Enable is a one-shot genesis call, and the realm's own source
# calls the address "dead weight" afterwards; pause/unpause runs through a
# GovDAO T1 proposal (ProposeSetPaused), not through this key. Funding it for
# names administration would have been funding nothing.
#
# It is funded because the same address is the hardcoded OWNER of two realms
# that ship in this genesis set — r/gnoland/blog (adminAddr) and
# r/gnoland/boards2/v0 (gPerms) — where ownership is fixed at realm source and
# every owner action is a paid tx. An owner at zero under §126 with no faucet
# could not post to the chain's own blog or administer its own boards, and
# could not be topped up until the transfer lock lifts.
NAMES_ADMIN=g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx

# ---- Locked sha256 hashes.
#
# Format (matches `shasum -a 256` / `sha256sum` output exactly):
#   <sha256>  <path-relative-to-mainnet.gno.land>
#
# Two spaces between hash and path. Blank lines and `#`-prefixed lines are
# ignored. The script calls `verify_checksum <path>` after producing an
# artifact:
#
#   - listed + hash matches  → silent pass
#   - listed + hash differs  → fail, expected vs got printed
#   - not listed             → note printed with the line to append
#
# Workflow: do a fresh end-to-end run, copy the "not listed" lines printed
# below this heredoc, commit, then any future run that produces a
# different output will fail loudly.
CHECKSUMS_DATA=$(
  cat <<'EOF'
# (empty — TODO(mainnet): locked only once every launch value is final:
# run a fresh end-to-end build and paste the printed "not listed" lines.)
EOF
)

# =============================================================================
# Helper functions.
# =============================================================================

# ---- Fatal error reporter

die() {
  printf 'ERROR: %s\n' "$*" >&2
  exit 1
}

# ---- Tool dispatchers

# sha256_of <path>
# Prints lowercase hex sha256 of the file's content. Tries shasum (macOS +
# most Linux), falls back to sha256sum (some Linux distros without shasum).
sha256_of() {
  local path="$1"
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$path" | awk '{print $1}'
  elif command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$path" | awk '{print $1}'
  else
    die "neither shasum nor sha256sum is installed (need one of them)"
  fi
}

# ---- Tool preflight
# require_tools <tool>...
# Probes every named tool; if any are missing, prints the full list with
# install hints (apt + brew) and exits. "shasum|sha256sum" is an
# at-least-one group; every other name is checked independently.
require_tools() {
  local missing=""
  local tool
  for tool in "$@"; do
    case "$tool" in
    "shasum|sha256sum")
      if ! command -v shasum >/dev/null 2>&1 && ! command -v sha256sum >/dev/null 2>&1; then
        missing="$missing shasum|sha256sum"
      fi
      ;;
    *)
      if ! command -v "$tool" >/dev/null 2>&1; then
        missing="$missing $tool"
      fi
      ;;
    esac
  done

  if [ -z "$missing" ]; then
    return 0
  fi

  printf 'ERROR: missing required tools:\n' >&2
  local m
  for m in $missing; do
    printf '  - %s\n' "$m" >&2
    case "$m" in
    "shasum|sha256sum")
      printf '      install:  brew install coreutils   |   apt-get install -y coreutils\n' >&2
      ;;
    jq)
      printf '      install:  brew install jq   |   apt-get install -y jq\n' >&2
      ;;
    go)
      printf '      install:  brew install go   |   see https://go.dev/doc/install\n' >&2
      ;;
    python3)
      printf '      install:  brew install python3   |   apt-get install -y python3\n' >&2
      ;;
    curl)
      printf '      install:  brew install curl   |   apt-get install -y curl\n' >&2
      ;;
    awk | sed | grep | sort | tr | mv | cp | ls | find | wc | head | tail | cut | comm | uniq | gzip)
      printf '      install:  comes with any POSIX userland (coreutils + findutils + gzip)\n' >&2
      ;;
    *)
      printf '      install:  consult your package manager\n' >&2
      ;;
    esac
  done
  exit 1
}

# ---- Checksum verification
# verify_checksum <path> [<manifest-key>]
#
# Computes sha256 of the file at <path>, looks up <manifest-key> (default:
# <path> relative to MAINNET_DIR) in the inline CHECKSUMS_DATA heredoc,
# and one of:
#   - hash matches               → silent OK
#   - hash differs               → FAIL with expected vs got
#   - key not listed             → print computed sha256 + the line to append
CHECKSUMS_LOCKED=0
CHECKSUMS_UNLOCKED=0
CHECKSUMS_UNLOCKED_LINES=""

verify_checksum() {
  local path="$1"
  if [ -z "${MAINNET_DIR:-}" ]; then
    die "verify_checksum: MAINNET_DIR not set"
  fi
  if [ ! -f "$path" ]; then
    die "verify_checksum: $path does not exist"
  fi

  local rel="${2:-${path#"$MAINNET_DIR"/}}"
  local got
  got=$(sha256_of "$path")

  local expected
  expected=$(printf '%s\n' "$CHECKSUMS_DATA" | awk -v rel="$rel" '
    /^[[:space:]]*$/ { next }
    /^[[:space:]]*#/ { next }
    {
      if ($2 == rel) { print $1; exit }
    }
  ')

  if [ -z "$expected" ]; then
    # Counted so the closing summary can say how much of this build was
    # actually verified: the append lines below scroll past mid-run,
    # interleaved with gnogenesis output, and "no mismatch reported" reads
    # like "locked" when it means "never checked".
    CHECKSUMS_UNLOCKED=$((CHECKSUMS_UNLOCKED + 1))
    CHECKSUMS_UNLOCKED_LINES="$CHECKSUMS_UNLOCKED_LINES$got  $rel"$'\n'
    printf '  [checksum] %s\n' "$rel" >&2
    printf '             not listed in CHECKSUMS_DATA. Append to lock:\n' >&2
    printf '             %s  %s\n' "$got" "$rel" >&2
    return 0
  fi
  CHECKSUMS_LOCKED=$((CHECKSUMS_LOCKED + 1))

  if [ "$expected" = "$got" ]; then
    return 0
  fi

  printf 'ERROR: checksum mismatch for %s\n' "$rel" >&2
  printf '       expected: %s\n' "$expected" >&2
  printf '       got:      %s\n' "$got" >&2
  exit 1
}

# ---- Output helpers
# print_step_header <step> <total> <title>
#   prints e.g. `=== Step 3 of 9: Build binaries from source ===`
print_step_header() {
  local step="$1"
  local total="$2"
  local title="$3"
  printf '\n=== Step %s of %s: %s ===\n' "$step" "$total" "$title"
}

# print_substep <code> <text>
#   prints e.g. `  [3.1] Building gno...`
print_substep() {
  local code="$1"
  shift
  printf '  [%s] %s\n' "$code" "$*"
}

# ---- Formatting helpers

# format_duration <seconds>
# Prints "<H> hours <M> minutes <S> seconds" with zero parts omitted.
format_duration() {
  local s="$1"
  if [ "$s" -lt 0 ]; then s=0; fi
  local h=$((s / 3600))
  local m=$(((s % 3600) / 60))
  local sec=$((s % 60))
  local out=""
  if [ "$h" -gt 0 ]; then out="$h hours"; fi
  if [ "$m" -gt 0 ]; then
    if [ -n "$out" ]; then out="$out "; fi
    out="${out}$m minutes"
  fi
  if [ "$sec" -gt 0 ] || [ -z "$out" ]; then
    if [ -n "$out" ]; then out="$out "; fi
    out="${out}$sec seconds"
  fi
  printf '%s' "$out"
}

# format_size <bytes>
# Prints "245 MB", "4 KB", "789 B". Decimal units (1000-based).
format_size() {
  local b="$1"
  if [ "$b" -ge 1000000000 ]; then
    awk -v b="$b" 'BEGIN { printf "%.1f GB", b/1000000000 }'
  elif [ "$b" -ge 1000000 ]; then
    awk -v b="$b" 'BEGIN { printf "%.0f MB", b/1000000 }'
  elif [ "$b" -ge 1000 ]; then
    awk -v b="$b" 'BEGIN { printf "%.0f KB", b/1000 }'
  else
    printf '%s B' "$b"
  fi
}

# file_size <path>  →  bytes (uses wc -c, which is portable; stat flags differ)
file_size() {
  wc -c <"$1" | tr -d ' '
}

# sheet_total <path>  →  the ugnot the sheet grants, summed
#
# Takes the balance (the amount before any `;vesting=` suffix), which is what
# InitChain credits; the schedule only says how much of it is locked.
#
# awk sums in doubles, so the result is exact while it stays under 2^53 ugnot
# (9.007e15, ~6.8x the 1.333e9 GNOT supply). assert_exact_sum below refuses to
# let a total cross that line unnoticed.
# Pass "-" to sum stdin (POSIX awk reads stdin for the "-" file operand).
sheet_total() {
  awk -F'[=;]' '{ a = $2; sub(/ugnot$/, "", a); s += a } END { printf "%d", s + 0 }' "$1"
}

# assert_exact_sum <total> <what>  →  dies if <total> left exact-integer range
assert_exact_sum() {
  if [ "$1" -ge 9007199254740992 ]; then
    die "$2 ($1 ugnot) reached 2^53, where the awk sums in sheet_total stop being exact — switch the reconciliation to arbitrary-precision arithmetic before trusting this build"
  fi
}

# assert_vesting_locked_at_genesis <sheet-file> <label>
#
# Every `;vesting=` schedule in the sheet must still be fully locked at
# GENESIS_TIME. The schedules are ABSOLUTE unix times, not offsets from
# launch, and nothing downstream relates them to the chain's genesis time
# (VestingSchedule.Validate only checks internal consistency, and `gnogenesis
# verify` does not know GENESIS_TIME). Two ways a schedule leaks:
#   - end <= GENESIS_TIME: expired — 100% liquid at block 1, whatever the type;
#   - a NON-delayed schedule whose start precedes GENESIS_TIME: continuous is
#     the default when `;type=` is absent, and it unlocks
#     amount*(now-start)/(end-start) (tm2/pkg/std/vesting.go), so a past start
#     hands out that fraction at block 1. Only `;type=delayed` cliffs vest
#     nothing before their end. This is now the load-bearing case: at the
#     0108ede pin 3,262,417 of the sheet's 3,262,481 rows carry a CONTINUOUS
#     §132 schedule from 1789084800, and one more (the forced-lockup public-sale
#     row) has start=0 — the epoch — so its `;type=delayed` suffix is the only
#     thing between a correct lockup and ~98% of it liquid. Upstream dropping a
#     suffix, or moving GENESIS_TIME past 1789084800, must fail the build.
# §132 anchors vesting to the day $GNOT becomes transferrable — GENESIS_TIME.
assert_vesting_locked_at_genesis() {
  local sheet="$1" label="$2" leaking
  leaking=$(grep -F -- ';vesting=' "$sheet" |
    awk -F'[,;]' -v g="$GENESIS_TIME" '{
      start = $3 + 0; end = $4 + 0
      if (end <= g) {
        printf "  %s  <- schedule ENDS before genesis: 100%% liquid at block 1\n", $0
      } else if ($5 != "type=delayed" && start < g) {
        printf "  %s  <- continuous schedule STARTED before genesis: %.2f%% liquid at block 1\n", \
          $0, (g - start) * 100 / (end - start)
      }
    }' || true)
  if [ -n "$leaking" ]; then
    die "$(printf '%s\n%s\n%s' \
      "these $label vesting schedules are not fully locked at GENESIS_TIME ($GENESIS_TIME):" \
      "$leaking" \
      "fix the schedule (a pre-genesis start needs ;type=delayed), re-pin for the real launch time, or move the launch.")"
  fi
}

# =============================================================================
# Flag parsing.
# =============================================================================

DEBUG=false
NO_INSTALL=false

print_usage() {
  cat <<'EOF'
gen-genesis.sh — mainnet genesis builder (single-file pipeline).

Usage:
  ./gen-genesis.sh [flags]

Flags:
  --no-install    Reuse previously built binaries in work/bin/.
  --debug         Echo the main pipeline commands before running them.
  -h, --help      Print this help and exit.

Output:
  genesis.json    Final artifact, sha256-locked against the
                  CHECKSUMS_DATA heredoc in this script.

See misc/deployments/mainnet.gno.land/README.md for what the genesis
contains.
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
  -h | --help)
    print_usage
    exit 0
    ;;
  --debug)
    DEBUG=true
    shift
    ;;
  --no-install)
    NO_INSTALL=true
    shift
    ;;
  *)
    echo "ERROR: Unknown argument: $1" >&2
    echo "Run with --help for usage." >&2
    exit 1
    ;;
  esac
done

run() {
  if [ "$DEBUG" = true ]; then
    printf "    \033[2m\$ %s\033[0m\n" "$*" >&2
  fi
  "$@"
}

# =============================================================================
# Shared paths + cleanup trap.
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MAINNET_DIR="$SCRIPT_DIR"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
EXAMPLES_DIR="$REPO_ROOT/examples"

WORK_DIR="$SCRIPT_DIR/work"
WORK_DIR_BIN="$WORK_DIR/bin"
WORK_DIR_GNOKEY_HOME="$WORK_DIR/gnokey-home"

GNO_CMD="$REPO_ROOT/gnovm/cmd/gno"
GNOKEY_CMD="$REPO_ROOT/gno.land/cmd/gnokey"
GNOLAND_CMD="$REPO_ROOT/gno.land/cmd/gnoland"
GNOGENESIS_CMD="$REPO_ROOT/contribs/gnogenesis"
GNO_BIN="$WORK_DIR_BIN/gno"
GNOKEY_BIN="$WORK_DIR_BIN/gnokey"
GNOLAND_BIN="$WORK_DIR_BIN/gnoland"
GNOGENESIS_BIN="$WORK_DIR_BIN/gnogenesis"

FINAL_GENESIS="$SCRIPT_DIR/genesis.json"

# Clean up temp node on exit (the balance-measurement step starts one; the
# trap is a no-op when NODE_PID is unset, so it's safe at script scope).
NODE_PID=""
cleanup() { [ -n "$NODE_PID" ] && kill "$NODE_PID" 2>/dev/null && wait "$NODE_PID" 2>/dev/null || true; }
trap cleanup EXIT

# =============================================================================
# Transaction loader — converts transactions/<...>/<txdir>/{meta.json +
# optional body file} into one AnnotatedTx jsonl line appended to <outfile>.
#
# Dispatches on meta.json's "kind" field:
#   MsgRun   — signs via gnokey maketx run + sign.
#   MsgCall  — signs via gnokey maketx call + sign; optionally jq-patches
#              msg[0].caller to caller_override post-sign (for genesis-mode
#              calls that need an admin caller without holding the admin
#              key — the chain trusts the caller field under
#              --skip-genesis-sig-verification).
#
# Emitted lines are in AnnotatedTx shape {tx, metadata, reason} — strip the
# reason field (jq -c 'del(.reason)') before passing to consumers that
# don't speak AnnotatedTx (e.g. `gnogenesis txs add sheets`).
# =============================================================================

txn_dir_to_jsonl() {
  local dir="$1" outfile="$2"
  local meta="$dir/meta.json"
  [ -f "$meta" ] || die "txn_dir_to_jsonl: $meta not found"

  local kind
  kind=$(jq -r '.kind' "$meta")
  case "$kind" in
  MsgRun) _txn_msg_run "$dir" "$outfile" ;;
  MsgCall) _txn_msg_call "$dir" "$outfile" ;;
  *) die "txn_dir_to_jsonl: unknown kind '$kind' in $meta" ;;
  esac
}

_txn_msg_run() {
  local dir="$1" outfile="$2"
  local meta="$dir/meta.json"

  local reason caller_key body_file gas_wanted gas_fee acct_num seq
  reason=$(jq -r '.reason' "$meta")
  caller_key=$(jq -r '.caller_key' "$meta")
  body_file=$(jq -r '.body_file' "$meta")
  gas_wanted=$(jq -r '.gas_wanted' "$meta")
  gas_fee=$(jq -r '.gas_fee' "$meta")
  acct_num=$(jq -r '.account_number' "$meta")
  seq=$(jq -r '.sequence' "$meta")

  local body_path="$dir/$body_file"
  local tx_json="$WORK_DIR/.txn.json"

  run "$GNOKEY_BIN" maketx run \
    --gas-wanted "$gas_wanted" \
    --gas-fee "$gas_fee" \
    --chainid "$CHAIN_ID" \
    --home "$WORK_DIR_GNOKEY_HOME" \
    --broadcast=false \
    --insecure-password-stdin \
    "$caller_key" \
    "$body_path" >"$tx_json" <<<""

  echo "" | run "$GNOKEY_BIN" sign \
    --tx-path "$tx_json" \
    --chainid "$CHAIN_ID" \
    --account-number "$acct_num" \
    --account-sequence "$seq" \
    --home "$WORK_DIR_GNOKEY_HOME" \
    --insecure-password-stdin \
    "$caller_key" >/dev/null

  jq -c --arg r "$reason" '{tx: ., metadata: {block_height: "0"}, reason: $r}' \
    "$tx_json" >>"$outfile"
  rm -f "$tx_json"
}

_txn_msg_call() {
  local dir="$1" outfile="$2"
  local meta="$dir/meta.json"

  local reason caller_key caller_override pkgpath func gas_wanted gas_fee acct_num seq
  reason=$(jq -r '.reason' "$meta")
  caller_key=$(jq -r '.caller_key' "$meta")
  caller_override=$(jq -r '.caller_override // empty' "$meta")
  pkgpath=$(jq -r '.pkgpath' "$meta")
  func=$(jq -r '.func' "$meta")
  gas_wanted=$(jq -r '.gas_wanted' "$meta")
  gas_fee=$(jq -r '.gas_fee' "$meta")
  acct_num=$(jq -r '.account_number' "$meta")
  seq=$(jq -r '.sequence' "$meta")

  # Expand args: meta.json's .args is a JSON array of strings; pass each as --args.
  local args_array=() arg
  while IFS= read -r arg; do
    [ -z "$arg" ] && continue
    args_array+=(--args "$arg")
  done < <(jq -r '.args[]?' "$meta")

  # Guard the extraction: a non-array .args yields zero --args silently,
  # and a dropped empty-string arg would shift every later positional.
  local args_len
  args_len=$(jq '.args // [] | length' "$meta")
  if [ $((${#args_array[@]} / 2)) -ne "$args_len" ]; then
    die "args mismatch in $meta: extracted $((${#args_array[@]} / 2)) of $args_len (empty or non-string args are not supported)"
  fi

  local tx_json="$WORK_DIR/.txn.json"

  echo "" | run "$GNOKEY_BIN" maketx call \
    --pkgpath "$pkgpath" \
    --func "$func" \
    "${args_array[@]}" \
    --gas-wanted "$gas_wanted" \
    --gas-fee "$gas_fee" \
    --chainid "$CHAIN_ID" \
    --home "$WORK_DIR_GNOKEY_HOME" \
    --broadcast=false \
    --insecure-password-stdin \
    "$caller_key" >"$tx_json"

  echo "" | run "$GNOKEY_BIN" sign \
    --tx-path "$tx_json" \
    --chainid "$CHAIN_ID" \
    --account-number "$acct_num" \
    --account-sequence "$seq" \
    --home "$WORK_DIR_GNOKEY_HOME" \
    --insecure-password-stdin \
    "$caller_key" >/dev/null

  if [ -n "$caller_override" ]; then
    jq -c --arg c "$caller_override" --arg r "$reason" \
      '.msg[0].caller = $c | {tx: ., metadata: {block_height: "0"}, reason: $r}' \
      "$tx_json" >>"$outfile"
  else
    jq -c --arg r "$reason" \
      '{tx: ., metadata: {block_height: "0"}, reason: $r}' \
      "$tx_json" >>"$outfile"
  fi
  rm -f "$tx_json"
}

# =============================================================================
# Pipeline.
# =============================================================================

PIPELINE_START_TS=$(date +%s)
TOTAL_STEPS=9

printf '\n### mainnet genesis build ###\n'

# ---- Code-submission vm params ----
# Applied to BOTH the shipping genesis and the step-8 measurement genesis, from
# ONE function. The params-parity assertion in step 8 requires the two to be
# byte-identical, and two call sites building the same JSON is exactly how they
# drift -- the same reasoning that keeps ChainDomain read from one place.
#
# Only non-empty values are written, so an unset list stays at the null the
# generator produced. Writing [] instead would mean the same thing to Params
# but would have to be written identically on both sides to keep parity.
apply_code_submission_params() {
  local genesis="$1" patched="$1.vmparams" approvers_json run_json applied applied_n
  approvers_json=$(printf '%s\n' "${PKG_APPROVERS[@]}" | jq -R -s -c 'split("\n") | map(select(length > 0))')
  run_json=$(printf '%s\n' "${RUN_SUBMITTERS[@]}" | jq -R -s -c 'split("\n") | map(select(length > 0))')
  jq --arg policy "$CODE_SUBMISSION_POLICY" \
    --argjson approvers "$approvers_json" \
    --argjson runners "$run_json" \
    --arg charge "$INERT_SUBMISSION_CHARGE" \
    --arg collector "$INERT_CHARGE_COLLECTOR" \
    '.app_state.vm.params.code_submission_policy = $policy
     | if ($approvers | length) > 0 then .app_state.vm.params.pkg_approvers = $approvers else . end
     | if ($runners | length) > 0 then .app_state.vm.params.run_submitters = $runners else . end
     | if $charge != "" then .app_state.vm.params.inert_submission_charge = $charge else . end
     | if $collector != "" then .app_state.vm.params.inert_charge_collector = $collector else . end' \
    "$genesis" >"$patched"
  mv "$patched" "$genesis"

  # Read back. A jq path typo writes a new key and reports nothing, and the
  # chain would then boot on the generator default -- "permissionless" -- with
  # every guard above having passed.
  applied=$(jq -r '.app_state.vm.params.code_submission_policy' "$genesis")
  if [ "$applied" != "$CODE_SUBMISSION_POLICY" ]; then
    die "code_submission_policy did not apply to $genesis: wanted '$CODE_SUBMISSION_POLICY', genesis has '$applied'"
  fi
  applied_n=$(jq -r '.app_state.vm.params.pkg_approvers | length // 0' "$genesis")
  if [ "$applied_n" -ne "${#PKG_APPROVERS[@]}" ]; then
    die "pkg_approvers did not apply to $genesis: wanted ${#PKG_APPROVERS[@]}, genesis has $applied_n"
  fi
  applied_n=$(jq -r '.app_state.vm.params.run_submitters | length // 0' "$genesis")
  if [ "$applied_n" -ne "${#RUN_SUBMITTERS[@]}" ]; then
    die "run_submitters did not apply to $genesis: wanted ${#RUN_SUBMITTERS[@]}, genesis has $applied_n"
  fi
}

# ---- Step 1: Resolve script paths and tooling

print_step_header 1 "$TOTAL_STEPS" "Resolve script paths and tooling"

GENESIS_FILE="$WORK_DIR/genesis.json"
PACKAGES_GEN_FILE="$WORK_DIR/packages.gen.txt"
GENESIS_TXS_JSONL="$WORK_DIR/genesis_txs.jsonl"
DEPLOYER_BALANCES="$WORK_DIR/deployers_balances.txt"
VALOPER_CSV="$WORK_DIR/valoper_profiles.csv"
VALOPER_SEED="$WORK_DIR/valoper-seed.jsonl"
# Read at step 2 (T1 funding guard) and added to the genesis at step 5.
BOOTSTRAP_DIR="$SCRIPT_DIR/transactions/base/bootstrap"

print_substep "1.1" "MAINNET_DIR=$MAINNET_DIR"
print_substep "1.2" "REPO_ROOT=$REPO_ROOT"
print_substep "1.3" "WORK_DIR=$WORK_DIR"

# ---- Step 2: Verify required tools

print_step_header 2 "$TOTAL_STEPS" "Verify required tools"

require_tools \
  "shasum|sha256sum" \
  go jq python3 curl gzip comm uniq \
  awk sed grep sort tr mv cp ls find wc head tail cut

print_substep "2.1" "All required tools present"

# Prepare work dir; preserve bin/ when --no-install.
if [ "$NO_INSTALL" = true ]; then
  mkdir -p "$WORK_DIR"
  find "$WORK_DIR" -mindepth 1 -maxdepth 1 ! -name bin -exec rm -rf {} + 2>/dev/null || true
else
  rm -rf "$WORK_DIR"
fi
mkdir -p "$WORK_DIR_BIN"

# ---- Allocation sheet (independence-day) ----
# Fetched, sha256-verified and shape-checked here — before the expensive
# build and tx steps — so a network failure, a stale cache, or a re-pin
# mistake fails in seconds instead of ten minutes in. The .gz is cached
# next to the script (gitignored) across runs.
ALLOCATION_GZ="$SCRIPT_DIR/allocation_balances.txt.gz"
ALLOCATION_TXT="$WORK_DIR/allocation_balances.txt"
if [ -f "$ALLOCATION_GZ" ] && [ "$(sha256_of "$ALLOCATION_GZ")" = "$ALLOCATION_SHA256" ]; then
  print_substep "2.2" "Using cached allocation sheet"
else
  if [ -f "$ALLOCATION_GZ" ]; then
    print_substep "2.2" "Cached allocation sheet does not match ALLOCATION_SHA256 (stale cache or re-pin) — re-downloading..."
    rm -f "$ALLOCATION_GZ"
  else
    print_substep "2.2" "Downloading allocation sheet..."
  fi
  # Download to a temp path and move into place only after the sha check:
  # a truncated download must never wedge the cache.
  run curl -fsSL --retry 3 "$ALLOCATION_GZ_URL" -o "$ALLOCATION_GZ.part"
  got_alloc_sha=$(sha256_of "$ALLOCATION_GZ.part")
  if [ "$got_alloc_sha" != "$ALLOCATION_SHA256" ]; then
    rm -f "$ALLOCATION_GZ.part"
    die "downloaded allocation sheet sha256 mismatch: expected $ALLOCATION_SHA256, got $got_alloc_sha — check the ALLOCATION_GZ_URL pin"
  fi
  mv "$ALLOCATION_GZ.part" "$ALLOCATION_GZ"
fi
gzip -dc "$ALLOCATION_GZ" >"$ALLOCATION_TXT"
alloc_count=$(wc -l <"$ALLOCATION_TXT" | tr -d ' ')
# The sha proves "this is the pinned file"; these prove the file has the
# shape the merge arithmetic in step 8 assumes (a re-pin could change
# either): one `g1<38>=<digits>ugnot` line per account, no duplicates.
# A row may carry a declared vesting schedule -- and since independence-day #72
# (in this pin) MOST rows do: the §132 pass runs by default, so all but 63 of
# the 3,262,481 rows carry one, plus the public-sale row under a mandatory
# forced lockup. A pattern that only accepts a bare balance rejects the sheet
# outright.
if grep -qvE '^g1[0-9a-z]{38}=[1-9][0-9]*ugnot(;vesting=[0-9]+ugnot,[0-9]+,[0-9]+(;type=[a-z]+)?)?$' "$ALLOCATION_TXT"; then
  die "allocation sheet has malformed lines (expected g1<38chars>=<digits>ugnot[;vesting=...] per line)"
fi
alloc_dupes=$(cut -d= -f1 "$ALLOCATION_TXT" | sort | uniq -d)
if [ -n "$alloc_dupes" ]; then
  die "allocation sheet has duplicate addresses: $alloc_dupes"
fi
# A sheet re-pinned for an earlier target launch, or a ceremony that slips
# past a schedule, would hand out "locked" balances at block 1 — see the
# helper's comment for both leak shapes.
assert_vesting_locked_at_genesis "$ALLOCATION_TXT" "allocation-sheet"
# Same check for the hand-typed VESTED_ACCOUNTS array — the input where the
# largest single grant (§132's 150M GNOT) will eventually be typed, and the
# one input a human can get wrong in milliseconds-vs-seconds or with a
# timestamp carried over from an earlier planned launch. The step-8 preflight
# runs `gnogenesis verify` with stdout discarded, and its implausible-vesting
# check is a WARNING even when seen — this is the only hard stop.
if [ "${#VESTED_ACCOUNTS[@]}" -gt 0 ]; then
  VESTED_EARLY_SHEET="$WORK_DIR/vested_accounts_check.txt"
  printf '%s\n' "${VESTED_ACCOUNTS[@]}" >"$VESTED_EARLY_SHEET"
  assert_vesting_locked_at_genesis "$VESTED_EARLY_SHEET" "VESTED_ACCOUNTS"
fi
# Captured here because the decompressed sheet is deleted at the end of step 8;
# step 9.5 reconciles the shipped genesis against this total.
alloc_total=$(sheet_total "$ALLOCATION_TXT")
assert_exact_sum "$alloc_total" "the allocation total"
if [ "$alloc_count" -ne "$ALLOCATION_EXPECTED_ACCOUNTS" ] || [ "$alloc_total" -ne "$ALLOCATION_EXPECTED_TOTAL" ]; then
  die "$(printf '%s\n%s\n%s' \
    "the pinned sheet does not match the declared magnitude:" \
    "  accounts: $alloc_count (declared $ALLOCATION_EXPECTED_ACCOUNTS), total: $alloc_total ugnot (declared $ALLOCATION_EXPECTED_TOTAL)" \
    "a re-pin changed mainnet's money supply — review the change, then update ALLOCATION_EXPECTED_* next to the pin.")"
fi
print_substep "2.3" "Allocation sheet: $alloc_count accounts, $alloc_total ugnot (sha256 + format + magnitude verified)"

# ---- Unrestricted addresses (independence-day, Constitution §126) ----
# Same fetch-and-verify treatment as the allocation sheet, and checked here for
# the same reason: a bad pin should fail in seconds, not ten minutes in.
UNRESTRICTED_FILE="$SCRIPT_DIR/allocation_unrestricted.txt"
if [ -f "$UNRESTRICTED_FILE" ] && [ "$(sha256_of "$UNRESTRICTED_FILE")" = "$UNRESTRICTED_SHA256" ]; then
  print_substep "2.4" "Using cached unrestricted-address list"
else
  if [ -f "$UNRESTRICTED_FILE" ]; then
    print_substep "2.4" "Cached unrestricted list does not match UNRESTRICTED_SHA256 (stale cache or re-pin) — re-downloading..."
    rm -f "$UNRESTRICTED_FILE"
  else
    print_substep "2.4" "Downloading unrestricted-address list..."
  fi
  run curl -fsSL --retry 3 "$UNRESTRICTED_URL" -o "$UNRESTRICTED_FILE.part"
  got_unres_sha=$(sha256_of "$UNRESTRICTED_FILE.part")
  if [ "$got_unres_sha" != "$UNRESTRICTED_SHA256" ]; then
    rm -f "$UNRESTRICTED_FILE.part"
    die "downloaded unrestricted list sha256 mismatch: expected $UNRESTRICTED_SHA256, got $got_unres_sha — check the UNRESTRICTED_URL pin"
  fi
  mv "$UNRESTRICTED_FILE.part" "$UNRESTRICTED_FILE"
fi

# Strip comments and blanks (the non-airdrop.txt convention that file follows).
UNRESTRICTED_ADDRS_TXT="$WORK_DIR/unrestricted_addrs.txt"
sed 's/#.*//' "$UNRESTRICTED_FILE" | tr -d " \t" | grep -v '^$' >"$UNRESTRICTED_ADDRS_TXT" || true
unrestricted_count=$(wc -l <"$UNRESTRICTED_ADDRS_TXT" | tr -d ' ')
if [ "$unrestricted_count" -eq 0 ]; then
  die "unrestricted list is empty after stripping comments — refusing to lock transfers with nobody exempt"
fi
if grep -qvE '^g1[0-9a-z]{38}$' "$UNRESTRICTED_ADDRS_TXT"; then
  die "unrestricted list has malformed lines (expected one g1<38chars> per line)"
fi
unres_dupes=$(sort "$UNRESTRICTED_ADDRS_TXT" | uniq -d)
if [ -n "$unres_dupes" ]; then
  die "unrestricted list has duplicate addresses: $unres_dupes"
fi

# gnoland's InitChain PANICS on an unrestricted address that is not a genesis
# account (gno.land/pkg/gnoland/app.go: "unrestricted address must be one of the
# genesis accounts"). Catch it here rather than at chain start.
missing_unres=$(cut -d= -f1 "$ALLOCATION_TXT" | sort -u | comm -13 - <(sort -u "$UNRESTRICTED_ADDRS_TXT"))
if [ -n "$missing_unres" ]; then
  die "unrestricted addresses absent from the allocation sheet (InitChain would panic): $missing_unres"
fi

# An address that BOTH carries a vesting schedule and is unrestricted needs the
# *GnoAccount design from gno 331e17fc3 (#6095). Before it, InitChain's type
# assertion in app.go panics. chain/pearl and chain/sapphire carry the vesting
# grammar but not that fix.
vested_and_unrestricted=$(grep ';vesting=' "$ALLOCATION_TXT" | cut -d= -f1 | sort -u |
  comm -12 - <(sort -u "$UNRESTRICTED_ADDRS_TXT") || true)
if [ -n "$vested_and_unrestricted" ]; then
  if ! git -C "$REPO_ROOT" merge-base --is-ancestor 331e17fc3 HEAD 2>/dev/null; then
    die "$(printf '%s\n' \
      "these addresses are both vested and unrestricted:" \
      "$vested_and_unrestricted" \
      "that combination needs gno commit 331e17fc3 (#6095), which this tree does not contain." \
      "Building here would produce a genesis that panics at InitChain.")"
  fi
  print_substep "2.5" "Vested+unrestricted overlap present; 331e17fc3 is in tree (required)"
fi

print_substep "2.6" "Unrestricted addresses: $unrestricted_count (sha256 + format verified, all funded)"

# ---- GovDAO T1 members must hold a genesis balance ----
# The bootstrap MsgRun seeds the T1 members, who then pay gas out of their own
# pocket for the chain's first proposals. mainnet has no faucet and no
# transferable supply outside the §126 exemption list, so a member seeded
# without a balance is locked out of governance with no in-chain way to top up.
# The addresses are read back out of the bootstrap source rather than repeated
# here, so this guard cannot drift from what actually gets seeded.
#
# Any nonzero balance qualifies, and deliberately so: fees are collected with
# the bank keeper's SendCoinsUnrestricted (tm2/pkg/sdk/auth/ante.go), which
# bypasses both the §126 transfer restriction and vesting locks, so a member
# needs neither an exemption-list entry nor an unvested balance to pay. What
# does NOT qualify is exact-burn fee-payer funding: the genesis txs consume it
# (step 8), leaving the address at zero.
T1_ADDRS_FILE="$WORK_DIR/t1_members.txt"
# The sole genesis T1 member (aeddi; the other six gnolang/multisigs
# [govdao] members are added post-genesis by proposal). Asserted rather
# than derived:
# a SetMember line dropped, duplicated or reshaped out of the grep's reach
# would otherwise shrink the set this guard covers without a word, which is
# the one direction where the damage is silent.
T1_EXPECTED_COUNT=1
BOOTSTRAP_GNO="$BOOTSTRAP_DIR/$(jq -r '.body_file' "$BOOTSTRAP_DIR/meta.json")"
if [ ! -f "$BOOTSTRAP_GNO" ]; then
  die "bootstrap body '$BOOTSTRAP_GNO' does not exist — check body_file in $BOOTSTRAP_DIR/meta.json"
fi
# Commented-out SetMember lines are stripped first: they carry the same shape
# but seed nobody. `|| true` lets a zero-match run reach the count check below
# instead of dying on grep's exit 1 under pipefail, unannounced.
grep -v '^[[:space:]]*//' "$BOOTSTRAP_GNO" |
  grep -oE 'memberstore\.T1, address\("g1[0-9a-z]{38}"\)' |
  grep -oE 'g1[0-9a-z]{38}' | sort -u >"$T1_ADDRS_FILE" || true
t1_count=$(wc -l <"$T1_ADDRS_FILE" | tr -d ' ')
if [ "$t1_count" -ne "$T1_EXPECTED_COUNT" ]; then
  die "$(printf '%s\n%s' \
    "expected $T1_EXPECTED_COUNT T1 members in $BOOTSTRAP_GNO, found $t1_count." \
    "A SetMember line was added, removed, duplicated or reshaped past this guard's grep — fix the file, or update T1_EXPECTED_COUNT if membership really changed.")"
fi

# The count above is over DISTINCT addresses, so it cannot see the same member
# seeded twice. That matters: MembersByTier.SetMember returns
# ErrMemberAlreadyExists on the second call and the bootstrap's must() turns it
# into a panic, which surfaces ~90 seconds later as a node crash during the
# measurement run instead of a sentence here.
# Counted per call, not per line: `grep -c` would score two SetMember calls
# sharing a line as one, and one line seeding the same address twice is the
# exact shape this check exists to catch.
t1_seed_calls=$(grep -v '^[[:space:]]*//' "$BOOTSTRAP_GNO" |
  grep -oE 'memberstore\.T1, address\("g1[0-9a-z]{38}"\)' | wc -l | tr -d ' ') || t1_seed_calls=0
if [ "$t1_seed_calls" -ne "$t1_count" ]; then
  die "$BOOTSTRAP_GNO makes $t1_seed_calls T1 SetMember calls for $t1_count distinct addresses — the repeated call fails with ErrMemberAlreadyExists and the bootstrap MsgRun panics at InitChain"
fi

# ---- The sole member's invitation points must cover the six post-genesis adds ----
#
# Every NewT1MemberRequest burns one point of the PROPOSER's balance when the
# proposal executes, and there is no realm path that grants points after the
# fact — not a proposal, not a promotion. So the number typed into NewMember()
# at genesis is the permanent ceiling on how many members the launch member can
# ever seat, and running out is terminal for the intended founding set: the
# remaining seats could only be filled by spending the points of members he had
# just invited, which no later proposal can rebalance.
#
# T1_EXPECTED_INVITATION_POINTS = the six remaining gnolang/multisigs [govdao]
# members + the three the seven-member bootstrap gave every founder. Asserted
# rather than defaulted, because the T1 tier's own default is 3 (it governs what
# an INVITEE is seeded with, not an inviter) — so the wrong value here reads as
# the right one, and only surfaces as a stuck governance set after launch.
T1_EXPECTED_INVITATION_POINTS=9
t1_points=$(grep -v '^[[:space:]]*//' "$BOOTSTRAP_GNO" |
  grep -oE 'memberstore\.T1, address\("g1[0-9a-z]{38}"\), memberstore\.NewMember\([0-9]+\)' |
  grep -oE 'NewMember\([0-9]+\)$' | grep -oE '[0-9]+') || t1_points=""
if [ "$t1_points" != "$T1_EXPECTED_INVITATION_POINTS" ]; then
  die "$(printf '%s\n%s' \
    "expected the sole T1 member in $BOOTSTRAP_GNO to be seeded with $T1_EXPECTED_INVITATION_POINTS invitation points, found '${t1_points:-none}'." \
    "With fewer, the six remaining [govdao] members cannot all be seated by proposal — and nothing on chain can top the balance up afterwards.")"
fi

t1_unfunded=""
while IFS= read -r t1_addr; do
  t1_rc=0
  t1_line=$(grep -m1 -- "^${t1_addr}=" "$ALLOCATION_TXT") || t1_rc=$?
  if [ "$t1_rc" -gt 1 ]; then
    die "grep failed looking up T1 member $t1_addr in the allocation sheet (exit $t1_rc)"
  fi
  if [ -z "$t1_line" ]; then
    for vested in "${VESTED_ACCOUNTS[@]}"; do
      case "$vested" in "${t1_addr}="*) t1_line="$vested" ;; esac
    done
  fi
  if [ -z "$t1_line" ]; then
    t1_unfunded="$t1_unfunded  $t1_addr — no genesis balance"$'\n'
    continue
  fi
  # Allocation lines were shape-checked at 2.3, but VESTED_ACCOUNTS entries are
  # not parsed until step 8, so validate the amount before comparing it: an
  # unparseable value would make `-eq` error out and read as "funded".
  t1_amount="${t1_line#*=}"
  t1_amount="${t1_amount%%ugnot*}"
  case "$t1_amount" in
  '' | *[!0-9]*) die "T1 member $t1_addr has an unparseable balance entry: '$t1_line'" ;;
  esac
  if [ "$t1_amount" -eq 0 ]; then
    t1_unfunded="$t1_unfunded  $t1_addr — zero genesis balance ($t1_line)"$'\n'
  fi
done <"$T1_ADDRS_FILE"
if [ -n "$t1_unfunded" ]; then
  die "$(printf '%s\n%s%s' \
    "these GovDAO T1 members hold no genesis balance:" \
    "$t1_unfunded" \
    "mainnet has no faucet: fund them in the gnolang/independence-day allocation, or drop them from the bootstrap.")"
fi
print_substep "2.7" "GovDAO T1 members: $t1_count seeded with $t1_points invitation points, each holds a genesis balance"

# ---- Every founding validator address must be able to pay for a tx ----
# The same argument as 2.7, for the other set of addresses that must act on a
# chain with no faucet and §126 in force. Rotating a signing key, editing a
# valoper profile and signalling opt-out are all paid txs, so an operator that
# lands at zero is locked out of its own validator permanently — there is no
# in-chain way to top it up afterwards.
#
# independence-day#78 funds all eight (four signing + four operator) at 1,000
# GNOT from the §122 Validator Services Treasury. This asserts it on the SHEET
# WE ACTUALLY PINNED rather than trusting the pin's commit message: a re-pin
# that drops the float has to fail here, not at launch.
#
# A vesting schedule does not disqualify an address, for the reason given at
# 2.7 — fees go through SendCoinsUnrestricted, which bypasses both the lock and
# the vesting. The floor is the float rather than "nonzero" because dust is not
# the decision that was made, and a row that decayed to dust is a bug either way.
VALIDATOR_MIN_GENESIS_UGNOT=1000000000 # 1,000 GNOT
# Signing addresses come out of INITIAL_VALSET (field 3) rather than a second
# hand-written list, so this cannot drift from the set that is actually seeded.
val_addrs=()
for validator in "${INITIAL_VALSET[@]}"; do
  read -r _name _power val_signing _pub_key <<<"$validator"
  case "$val_signing" in
  g1*) ;;
  *) die "INITIAL_VALSET entry '$validator' does not carry a g1 signing address in field 3" ;;
  esac
  val_addrs+=("$val_signing")
done
val_addrs+=("${INITIAL_VALSET_OPERATORS[@]}")

val_unfunded=""
val_checked=0
for val_addr in "${val_addrs[@]}"; do
  val_checked=$((val_checked + 1))
  val_rc=0
  val_line=$(grep -m1 -- "^${val_addr}=" "$ALLOCATION_TXT") || val_rc=$?
  if [ "$val_rc" -gt 1 ]; then
    die "grep failed looking up validator address $val_addr in the allocation sheet (exit $val_rc)"
  fi
  if [ -z "$val_line" ] && [ "${#VESTED_ACCOUNTS[@]}" -gt 0 ]; then
    for vested in "${VESTED_ACCOUNTS[@]}"; do
      case "$vested" in "${val_addr}="*) val_line="$vested" ;; esac
    done
  fi
  if [ -z "$val_line" ]; then
    val_unfunded="$val_unfunded  $val_addr — no genesis balance"$'\n'
    continue
  fi
  val_amount="${val_line#*=}"
  val_amount="${val_amount%%ugnot*}"
  case "$val_amount" in
  '' | *[!0-9]*) die "validator address $val_addr has an unparseable balance entry: '$val_line'" ;;
  esac
  if [ "$val_amount" -lt "$VALIDATOR_MIN_GENESIS_UGNOT" ]; then
    val_unfunded="$val_unfunded  $val_addr — $val_amount ugnot, below the $VALIDATOR_MIN_GENESIS_UGNOT ugnot floor"$'\n'
  fi
done
# Zero addresses checked would pass the loop silently; that can only mean the
# two arrays were emptied or renamed, which is exactly when this guard matters.
val_expected=$((${#INITIAL_VALSET[@]} * 2))
if [ "$val_checked" -ne "$val_expected" ]; then
  die "validator funding guard checked $val_checked addresses, expected $val_expected (${#INITIAL_VALSET[@]} slots x signing+operator)"
fi
if [ -n "$val_unfunded" ]; then
  die "$(printf '%s\n%s%s' \
    "these founding validator addresses cannot pay for a transaction at genesis:" \
    "$val_unfunded" \
    "mainnet has no faucet and §126 locks transfers: fund them in the gnolang/independence-day allocation (see its genesisValidators list) and re-pin.")"
fi
print_substep "2.8" "Founding validators: $val_checked addresses, each holds at least $VALIDATOR_MIN_GENESIS_UGNOT ugnot"

# ---- NAMES_ADMIN must match the admin compiled into r/sys/names ----
# names.Enable is gated on an address hardcoded in the realm source, and
# NAMES_ADMIN is a copy of it. That copy has gone stale once already (#6131
# moved the admin to the multisigs [govdao] key), and a stale copy surfaces
# ~90 seconds in as a "caller is not admin" panic during the measurement run.
# Read the authority out of the tree instead. Step 6 checks the other half:
# that the tx's caller_override matches NAMES_ADMIN too.
NAMES_VERIFIER_GNO="$REPO_ROOT/examples/gno.land/r/sys/names/verifier.gno"
if [ ! -f "$NAMES_VERIFIER_GNO" ]; then
  die "cannot find $NAMES_VERIFIER_GNO — r/sys/names moved; re-point this check and re-confirm NAMES_ADMIN"
fi
names_admin_in_tree=$(grep -oE 'admin[[:space:]]*=[[:space:]]*address\("g1[0-9a-z]{38}"\)' "$NAMES_VERIFIER_GNO" |
  grep -oE 'g1[0-9a-z]{38}') || names_admin_in_tree=""
if [ -z "$names_admin_in_tree" ]; then
  die "no admin address found in $NAMES_VERIFIER_GNO — the realm's admin declaration changed shape; re-confirm NAMES_ADMIN by hand"
fi
if [ "$names_admin_in_tree" != "$NAMES_ADMIN" ]; then
  die "$(printf '%s\n%s\n%s' \
    "NAMES_ADMIN ($NAMES_ADMIN) is not the admin compiled into r/sys/names ($names_admin_in_tree)." \
    "names.Enable would be rejected at genesis: the caller_override patch only works for the address the realm actually trusts." \
    "Update NAMES_ADMIN and transactions/migration/names-enable/meta.json after confirming who that address belongs to.")"
fi
print_substep "2.9" "names admin matches r/sys/names in-tree: $names_admin_in_tree"

# ---- Inert policy: the four values have to agree with each other ----
# RUN_SUBMITTERS is assembled here, after 2.7 has read the seeded T1 members
# out of the bootstrap source, so the list cannot disagree with the membership
# the same build is about to seed.
RUN_SUBMITTERS=()
if [ "$RUN_SUBMITTERS_INCLUDE_GOVDAO_T1" = true ]; then
  while IFS= read -r t1_addr; do
    [ -n "$t1_addr" ] && RUN_SUBMITTERS+=("$t1_addr")
  done <"$T1_ADDRS_FILE"
fi
for extra_addr in "${RUN_SUBMITTERS_EXTRA[@]}"; do
  case " ${RUN_SUBMITTERS[*]} " in
  *" $extra_addr "*) die "RUN_SUBMITTERS_EXTRA lists $extra_addr, which is already a seeded GovDAO T1 member" ;;
  *) RUN_SUBMITTERS+=("$extra_addr") ;;
  esac
done

# Every one of these is a state the chain accepts and cannot repair on its
# own: Params.Validate checks address syntax and nothing else, and InitChain
# logs a note for an empty run_submitters but says nothing about an empty
# pkg_approvers. A genesis is forever, so they are build-time errors here.
case "$CODE_SUBMISSION_POLICY" in
permissionless | permissioned | inert) ;;
*) die "CODE_SUBMISSION_POLICY must be permissionless, permissioned or inert (got '$CODE_SUBMISSION_POLICY')" ;;
esac

# An approver list that is set under the wrong policy is as broken as an empty
# one under "inert" -- nothing would ever park, so the approvers would have
# nothing to enable and the key would be armed for no reason.
if [ "$CODE_SUBMISSION_POLICY" = inert ] && [ "${#PKG_APPROVERS[@]}" -eq 0 ]; then
  die "$(printf '%s\n%s' \
    "CODE_SUBMISSION_POLICY=inert with an empty PKG_APPROVERS: every post-genesis submission would park with nobody able to enable it, forever." \
    "Paste the gpao oracle address into PKG_APPROVERS.")"
fi
if [ "$CODE_SUBMISSION_POLICY" != inert ] && [ "${#PKG_APPROVERS[@]}" -gt 0 ]; then
  die "PKG_APPROVERS is set but CODE_SUBMISSION_POLICY is '$CODE_SUBMISSION_POLICY': nothing would ever park, so the approvers would have nothing to enable"
fi

for inert_addr in "${PKG_APPROVERS[@]}" "${RUN_SUBMITTERS[@]}"; do
  case "$inert_addr" in
  g1*) [ "${#inert_addr}" -eq 40 ] || die "malformed address in PKG_APPROVERS/RUN_SUBMITTERS: $inert_addr" ;;
  *) die "malformed address in PKG_APPROVERS/RUN_SUBMITTERS: $inert_addr" ;;
  esac
done

# An approver pays the gas for every MsgEnablePackage it sends, and mainnet has
# no faucet and a transfer lock, so an unfunded oracle can never approve
# anything and cannot be topped up by anyone outside the exemption list. Same
# check and same reasoning as the T1 guard at 2.7 -- and deliberately against
# the allocation sheet rather than minting a balance here, so the shipped
# supply stays tied to the sha256-verified sheet and step 9.5 still reconciles.
approver_unfunded=""
for inert_addr in "${PKG_APPROVERS[@]}"; do
  inert_rc=0
  inert_line=$(grep -m1 -- "^${inert_addr}=" "$ALLOCATION_TXT") || inert_rc=$?
  if [ "$inert_rc" -gt 1 ]; then
    die "grep failed looking up approver $inert_addr in the allocation sheet (exit $inert_rc)"
  fi
  if [ -z "$inert_line" ]; then
    approver_unfunded="$approver_unfunded  $inert_addr — no genesis balance"$'\n'
    continue
  fi
  inert_amount="${inert_line#*=}"
  inert_amount="${inert_amount%%ugnot*}"
  case "$inert_amount" in
  '' | *[!0-9]*) die "approver $inert_addr has an unparseable balance entry: '$inert_line'" ;;
  esac
  if [ "$inert_amount" -eq 0 ]; then
    approver_unfunded="$approver_unfunded  $inert_addr — zero genesis balance ($inert_line)"$'\n'
  fi
done
if [ -n "$approver_unfunded" ]; then
  die "$(printf '%s\n%s%s' \
    "these package approvers hold no genesis balance:" \
    "$approver_unfunded" \
    "mainnet has no faucet and §126 locks transfers: fund them in the gnolang/independence-day allocation, or the oracle can never enable anything.")"
fi

# RUN_SUBMITTERS armed must cover every seeded GovDAO T1 member.
#
# Proposal creation is MsgRun-only, so a T1 member missing from this list
# cannot propose -- and since amending the list is itself a proposal, a list
# that omits ALL of them is unamendable. Checked against the addresses read out
# of the bootstrap tx at 2.7, not a second copy typed here, so the two cannot
# drift.
if [ "${#RUN_SUBMITTERS[@]}" -gt 0 ]; then
  run_missing=""
  while IFS= read -r t1_addr; do
    case " ${RUN_SUBMITTERS[*]} " in
    *" $t1_addr "*) ;;
    *) run_missing="$run_missing  $t1_addr"$'\n' ;;
    esac
  done <"$T1_ADDRS_FILE"
  if [ -n "$run_missing" ]; then
    die "$(printf '%s\n%s%s' \
      "RUN_SUBMITTERS is armed but omits these seeded GovDAO T1 members:" \
      "$run_missing" \
      "proposal creation is MsgRun-only, so they could not govern and the list could never be amended.")"
  fi
fi

# The submission charge is incompatible with the §126 transfer lock: it is paid
# through the RESTRICTED SendCoins path, so under restricted_denoms=["ugnot"]
# only an exemption-listed address could afford to submit at all.
if [ -n "$INERT_SUBMISSION_CHARGE" ] && [ "${#RESTRICTED_DENOMS[@]}" -gt 0 ]; then
  die "$(printf '%s\n%s' \
    "INERT_SUBMISSION_CHARGE=$INERT_SUBMISSION_CHARGE with restricted_denoms=[${RESTRICTED_DENOMS[*]}]: the charge goes through the restricted bank path, so no address outside the §126 exemption list could pay it and MsgAddPackage would be refused for everyone else." \
    "Leave the charge empty while the transfer lock is on, or lift the lock in the same proposal.")"
fi
if [ -z "$INERT_SUBMISSION_CHARGE" ] && [ -n "$INERT_CHARGE_COLLECTOR" ]; then
  die "INERT_CHARGE_COLLECTOR is set but INERT_SUBMISSION_CHARGE is empty: nothing would ever be collected"
fi
print_substep "2.10" "$(printf 'Code submission: policy=%s, approvers=%d, run_submitters=%d (gate %s), charge=%s' \
  "$CODE_SUBMISSION_POLICY" "${#PKG_APPROVERS[@]}" "${#RUN_SUBMITTERS[@]}" \
  "$([ "${#RUN_SUBMITTERS[@]}" -gt 0 ] && echo ARMED || echo OPEN)" "${INERT_SUBMISSION_CHARGE:-off}")"

# ---- Step 3: Build binaries from source

print_step_header 3 "$TOTAL_STEPS" "Build binaries from source"

if [ "$NO_INSTALL" = true ]; then
  print_substep "3.1" "--no-install — reusing $WORK_DIR_BIN"
  for bin in "$GNO_BIN" "$GNOKEY_BIN" "$GNOLAND_BIN" "$GNOGENESIS_BIN"; do
    if [ ! -x "$bin" ]; then
      die "--no-install but $bin not found. Run without --no-install first."
    fi
  done
else
  print_substep "3.1" "Building gno..."
  run go build -C "$GNO_CMD" -o "$GNO_BIN" .
  print_substep "3.2" "Building gnokey..."
  run go build -C "$GNOKEY_CMD" -o "$GNOKEY_BIN" .
  print_substep "3.3" "Building gnoland..."
  run go build -C "$GNOLAND_CMD" -o "$GNOLAND_BIN" .
  print_substep "3.4" "Building gnogenesis..."
  run go build -C "$GNOGENESIS_CMD" -o "$GNOGENESIS_BIN" .
fi

# ---- Step 4: Generate filtered examples genesis txs

print_step_header 4 "$TOTAL_STEPS" "Generate filtered examples genesis txs"

print_substep "4.1" "Resolving dependencies..."
# -test-dep also resolves test-only imports (uassert, urequire, ...).
# Packages ship on-chain with their _test.gno files (MPUserAll), but
# addpkg type-checks production files only (tests are stored and
# syntax-parsed), so a missing test dep does not fail the deploy — the
# deps are included so the shipped test files keep their imports
# resolvable on-chain.
pkg_dirs=$(cd "$EXAMPLES_DIR" && "$GNO_BIN" tool deplist -test-dep "${FILTERED_PACKAGES[@]}")
pkg_count=$(echo "$pkg_dirs" | wc -l | tr -d ' ')
print_substep "4.2" "Resolved $pkg_count packages in topological order"

# Save resolved package list (used for audit + tracked by CHECKSUMS).
{
  echo "# Generated by gen-genesis.sh — do not edit."
  # shellcheck disable=SC2001 # path contains slashes; `|` as sed delimiter is cleaner than ${//} escaping
  echo "$pkg_dirs" | sed "s|$EXAMPLES_DIR/||g"
} >"$PACKAGES_GEN_FILE"
verify_checksum "$PACKAGES_GEN_FILE"

print_substep "4.3" "Copying packages to staging..."
# Rebuild staging from scratch — leftovers from a previous partial run
# would be addpkg'd into the genesis alongside the current set.
WORK_DIR_EXAMPLES="$WORK_DIR/examples"
rm -rf "$WORK_DIR_EXAMPLES"
mkdir -p "$WORK_DIR_EXAMPLES"
while IFS= read -r dir; do
  [ -z "$dir" ] && continue
  rel="${dir#"$EXAMPLES_DIR"/}"
  dest="$WORK_DIR_EXAMPLES/$rel"
  mkdir -p "$dest"
  find "$dir" -maxdepth 1 -type f -exec cp {} "$dest/" \;
  if [ -d "$dir/filetests" ]; then
    cp -r "$dir/filetests" "$dest/filetests"
  fi
done <<<"$pkg_dirs"

print_substep "4.4" "Creating deployer key..."
printf '%s\n\n' "$DEPLOYER_MNEMONIC" | run "$GNOKEY_BIN" add --recover "$DEPLOYER_KEY" --home "$WORK_DIR_GNOKEY_HOME" --insecure-password-stdin 2>&1 | sed 's/^/    /'

# DEPLOYER_ADDR is load-bearing three ways (valoper fee payer, readiness
# probe, measured funding) — assert it matches the recovered key.
deployer_listed=$("$GNOKEY_BIN" list --home "$WORK_DIR_GNOKEY_HOME" 2>/dev/null || true)
case "$deployer_listed" in
*"$DEPLOYER_ADDR"*) ;;
*) die "DEPLOYER_ADDR $DEPLOYER_ADDR is not the address derived from DEPLOYER_MNEMONIC" ;;
esac

print_substep "4.5" "Generating empty genesis..."
run "$GNOGENESIS_BIN" generate -chain-id "$CHAIN_ID" -genesis-time "$GENESIS_TIME" --output-path "$GENESIS_FILE" 2>&1 | sed 's/^/    /'

print_substep "4.6" "Adding $pkg_count packages to genesis..."
echo "" | run "$GNOGENESIS_BIN" txs add packages "$WORK_DIR_EXAMPLES" -gno-home "$WORK_DIR_GNOKEY_HOME" -key-name "$DEPLOYER_KEY" --genesis-path "$GENESIS_FILE" --insecure-password-stdin 2>&1 | sed 's/^/    /'

print_substep "4.7" "Exporting txs..."
run "$GNOGENESIS_BIN" txs export "$GENESIS_TXS_JSONL" --genesis-path "$GENESIS_FILE" 2>&1 | sed 's/^/    /'

# $pkg_count is what deplist resolved; this is what actually became a deploy.
# Three mechanisms drop a package between the two without a word: the staging
# copy at 4.3 uses `find -exec cp {} \;`, where a failed cp still leaves find
# exiting 0 (verified); ReadPkgListFromDir skips any directory whose
# gnomod.toml is missing (gnovm/pkg/packages/readpkglist.go); and
# GetNonIgnoredPkgs drops ignore-marked packages AND everything that depends
# on them (gnovm/pkg/packages/pkglist.go). A p/nt/* path that goes missing
# here cannot be added post-genesis — it is a relaunch.
addpkg_count=$(jq -r '[.tx.msg[] | select(.["@type"] == "/vm.m_addpkg")] | length' "$GENESIS_TXS_JSONL" |
  awk '{ s += $1 } END { print s + 0 }')
if [ "$addpkg_count" -ne "$pkg_count" ]; then
  die "$pkg_count packages resolved but $addpkg_count addpkg txs landed in the genesis — a package was dropped between staging and deploy (staging cp failure, missing gnomod.toml, or an ignore-marked dependency)"
fi
print_substep "4.8" "Reconciled: $addpkg_count addpkg txs for $pkg_count resolved packages"

# ---- Step 5: Add the bootstrap MsgRun (transactions/base/bootstrap/)
# Seeds the sole GovDAO T1 member and locks AllowedDAOs. The
# §126 transfer lock is applied at step 9.3 as genesis params rather than
# via r/sys/params proposals, so there is nothing to propose here.

print_step_header 5 "$TOTAL_STEPS" "Add bootstrap MsgRuns (GovDAO seed + namespace preregistration)"

BOOTSTRAP_JSONL="$WORK_DIR/bootstrap_tx.jsonl"

print_substep "5.1" "Building AnnotatedTx from $BOOTSTRAP_DIR/..."
: >"$BOOTSTRAP_JSONL"
txn_dir_to_jsonl "$BOOTSTRAP_DIR" "$BOOTSTRAP_JSONL"

# `txs add sheets` consumes plain TxWithMetadata (no reason field); strip it.
BOOTSTRAP_TX_FILE="$WORK_DIR/bootstrap_tx_stripped.jsonl"
jq -c 'del(.reason)' "$BOOTSTRAP_JSONL" >"$BOOTSTRAP_TX_FILE"

print_substep "5.2" "Adding bootstrap tx to genesis..."
run "$GNOGENESIS_BIN" txs add sheets "$BOOTSTRAP_TX_FILE" --genesis-path "$GENESIS_FILE" 2>&1 | sed 's/^/    /'
cat "$BOOTSTRAP_TX_FILE" >>"$GENESIS_TXS_JSONL"

# ---- Namespace preregistration (transactions/base/users-preregister/) ----
# Registers the initial mainnet names in r/sys/users through the genesis-only
# r/sys/users/init.RegisterUser wrapper (the controller gate is skipped at
# height 0, so no authority survives genesis — see the .gno body's header).
# Ordered AFTER the addpkg stream like every genesis tx, so r/sys/users and
# r/sys/users/init exist when it runs.
USERS_PREREGISTER_DIR="$SCRIPT_DIR/transactions/base/users-preregister"
USERS_PREREGISTER_JSONL="$WORK_DIR/users_preregister_tx.jsonl"

print_substep "5.3" "Building AnnotatedTx from $USERS_PREREGISTER_DIR/..."
: >"$USERS_PREREGISTER_JSONL"
txn_dir_to_jsonl "$USERS_PREREGISTER_DIR" "$USERS_PREREGISTER_JSONL"

USERS_PREREGISTER_TX_FILE="$WORK_DIR/users_preregister_tx_stripped.jsonl"
jq -c 'del(.reason)' "$USERS_PREREGISTER_JSONL" >"$USERS_PREREGISTER_TX_FILE"

print_substep "5.4" "Adding namespace-preregistration tx to genesis..."
run "$GNOGENESIS_BIN" txs add sheets "$USERS_PREREGISTER_TX_FILE" --genesis-path "$GENESIS_FILE" 2>&1 | sed 's/^/    /'
cat "$USERS_PREREGISTER_TX_FILE" >>"$GENESIS_TXS_JSONL"

# ---- Step 6: Add the names.Enable MsgCall (transactions/migration/names-enable/)
# Namespace enforcement on from genesis. Enable is gated on the admin
# hardcoded in r/sys/names/verifier.gno; caller_override makes the tx
# appear as that admin (trusted under --skip-genesis-sig-verification).
# Ordered AFTER every addpkg, so enforcement never gates the genesis
# deploys themselves.

print_step_header 6 "$TOTAL_STEPS" "Add names.Enable MsgCall (namespace enforcement)"

NAMES_ENABLE_DIR="$SCRIPT_DIR/transactions/migration/names-enable"
NAMES_ENABLE_JSONL="$WORK_DIR/names_enable_tx.jsonl"

# The admin-gated call is the single security-relevant genesis tx; the
# caller the chain will trust comes from meta.json's caller_override, not
# from NAMES_ADMIN — assert they agree instead of logging an unverified
# value.
names_caller=$(jq -r '.caller_override // empty' "$NAMES_ENABLE_DIR/meta.json")
if [ "$names_caller" != "$NAMES_ADMIN" ]; then
  die "names-enable caller_override ($names_caller) != NAMES_ADMIN ($NAMES_ADMIN)"
fi

print_substep "6.1" "Building AnnotatedTx from $NAMES_ENABLE_DIR/..."
: >"$NAMES_ENABLE_JSONL"
txn_dir_to_jsonl "$NAMES_ENABLE_DIR" "$NAMES_ENABLE_JSONL"

NAMES_ENABLE_TX_FILE="$WORK_DIR/names_enable_tx_stripped.jsonl"
jq -c 'del(.reason)' "$NAMES_ENABLE_JSONL" >"$NAMES_ENABLE_TX_FILE"

print_substep "6.2" "Adding names.Enable tx to genesis (caller=$NAMES_ADMIN)..."
run "$GNOGENESIS_BIN" txs add sheets "$NAMES_ENABLE_TX_FILE" --genesis-path "$GENESIS_FILE" 2>&1 | sed 's/^/    /'
cat "$NAMES_ENABLE_TX_FILE" >>"$GENESIS_TXS_JSONL"

# ---- Step 7: Add the valoper-seed Register MsgCalls
# Builds a CSV from (INITIAL_VALSET, INITIAL_VALSET_OPERATORS) and runs
# `gnogenesis fork valoper-seed` to produce a deterministic .jsonl of
# genesis-mode valopers.Register MsgCalls — one valoper profile per
# founding validator, keyed on its operator address. Without these the
# chain still boots (the valoper coverage assertion only fires in
# hardfork mode), but the founding validators would have no operator-
# keyed management plane in r/sys/validators/v0.

print_step_header 7 "$TOTAL_STEPS" "Add valoper-seed Register MsgCalls"

if [ "${#INITIAL_VALSET_OPERATORS[@]}" -ne "${#INITIAL_VALSET[@]}" ]; then
  die "INITIAL_VALSET_OPERATORS length (${#INITIAL_VALSET_OPERATORS[@]}) must match INITIAL_VALSET length (${#INITIAL_VALSET[@]})"
fi

print_substep "7.1" "Building CSV from INITIAL_VALSET + INITIAL_VALSET_OPERATORS..."
{
  echo "operator_addr,signing_pubkey,moniker,description,server_type"
  for i in "${!INITIAL_VALSET[@]}"; do
    read -r name _power _address pub_key <<<"${INITIAL_VALSET[$i]}"
    op_addr="${INITIAL_VALSET_OPERATORS[$i]}"
    # description and server_type are templates — edit if a specific
    # founder needs different metadata. Description must be non-empty
    # and <=2048 chars; server_type ∈ {cloud, on-prem, data-center}.
    printf '%s,%s,%s,mainnet founding validator (%s),cloud\n' \
      "$op_addr" "$pub_key" "$name" "$name"
  done
} >"$VALOPER_CSV"

print_substep "7.2" "Running gnogenesis fork valoper-seed..."
# --caller is the fee payer for each Register MsgCall (1 ugnot fee — see
# the Coin amino zero-collapse rationale in valoper_seed.go). The deployer
# pays; the balance-measurement step (step 8) funds it exactly. The
# operator from the CSV row is passed in MsgCall.Args[3], so each operator
# gets registered correctly; the squat guard (caller==operator) is
# bypassed at genesis-mode (ChainHeight()==0).
run "$GNOGENESIS_BIN" fork valoper-seed \
  --csv "$VALOPER_CSV" \
  --output "$VALOPER_SEED" \
  --caller "$DEPLOYER_ADDR" 2>&1 | sed 's/^/    /'
verify_checksum "$VALOPER_SEED"

VALOPER_TX_FILE="$WORK_DIR/valoper_seed_stripped.jsonl"
jq -c 'del(.reason)' "$VALOPER_SEED" >"$VALOPER_TX_FILE"

# One Register per CSV row, or a founding validator silently launches with no
# operator-keyed valoper profile — no way to rotate its signing key or signal
# opt-out through r/sys/validators/v0, and the chain boots anyway.
valoper_tx_count=$(wc -l <"$VALOPER_SEED" | tr -d ' ')
if [ "$valoper_tx_count" -ne "${#INITIAL_VALSET[@]}" ]; then
  die "valoper-seed emitted $valoper_tx_count Register txs for ${#INITIAL_VALSET[@]} validators — a CSV row was rejected or dropped"
fi

print_substep "7.3" "Adding $valoper_tx_count valoper Register txs to genesis..."
run "$GNOGENESIS_BIN" txs add sheets "$VALOPER_TX_FILE" --genesis-path "$GENESIS_FILE" 2>&1 | sed 's/^/    /'
cat "$VALOPER_TX_FILE" >>"$GENESIS_TXS_JSONL"
verify_checksum "$GENESIS_TXS_JSONL"

tx_count=$(wc -l <"$GENESIS_TXS_JSONL" | tr -d ' ')
print_substep "7.4" "Total genesis txs: $tx_count"

# `txs add sheets` dedupes against the genesis while the jsonl is a plain
# concat — a duplicated tx would be measured (and funded) in step 8 but
# not shipped. Reconcile the two before measuring.
genesis_tx_count=$(jq '.app_state.txs | length' "$GENESIS_FILE")
if [ "$genesis_tx_count" -ne "$tx_count" ]; then
  die "genesis holds $genesis_tx_count txs but the exported stream has $tx_count — duplicated or dropped txs"
fi

# ---- Step 8: Calculate genesis fee-payer balances
# Same approach as gnoland1's gen-genesis.sh: spin up a temp node, pre-fund
# every creator/caller address with $INITIAL_BALANCE, let the genesis txs
# burn through fees, then query remaining balances. The amount actually
# spent is what we credit each fee payer in the real genesis so their
# balance lands at zero post-genesis — the final state then holds ONLY the
# allocation balances. The fee payers are the deployer (addpkgs, bootstrap,
# valoper Register fees), the names admin (names.Enable fee), and every
# gnomod.toml [addpkg] creator address in the package set (used as that
# package's addpkg creator in place of the deployer).
#
# Run twice for safety:
#   run 1: measure actual consumption with over-provisioned balances
#   run 2: verify the measured balances land everyone at zero
# If run 2 disagrees, something is non-deterministic and we abort.

# The code-submission vm params go into the shipping genesis BEFORE the
# measurement: step 8's params-parity assertion compares vm.params between
# the shipping and the temp-node genesis, and the temp side is patched with
# the same function. (Genesis replay is policy-exempt via IsGenesisReplay,
# so this changes no measured amount.) Step 9.5 reads the values back from
# the final artifact.
apply_code_submission_params "$GENESIS_FILE"

print_step_header 8 "$TOTAL_STEPS" "Calculate genesis fee-payer balances"

BALANCES_TMP_DIR="$WORK_DIR/balances-work"
BALANCES_TMP_FILE="$BALANCES_TMP_DIR/balances.txt"
BALANCES_TMP_GNOLAND_DATA="$BALANCES_TMP_DIR/gnoland-data"
BALANCES_TMP_GNOLAND_LOG="$BALANCES_TMP_DIR/node.log"
BALANCES_TMP_GENESIS="$BALANCES_TMP_DIR/genesis.json"
BALANCES_TMP_CREATOR_ADDRESSES="$BALANCES_TMP_DIR/gen-creators.txt"
INITIAL_BALANCE=1000000000000000
NODE_TIMEOUT=120

pick_free_port() {
  python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1",0)); print(s.getsockname()[1]); s.close()'
}

rm -rf "$BALANCES_TMP_DIR"
mkdir -p "$BALANCES_TMP_DIR"

print_substep "8.1" "Extracting creator/caller addresses..."
# The extraction below assumes these msg types (their signer field is
# creator/caller). Any other type names its fee payer differently and would
# end up unfunded — and an unfunded genesis signer is not an error: InitChain
# creates the account and MINTS it 10,000 GNOT (genesisSignerFunding,
# gno.land/pkg/gnoland/app.go), which lands in the supply counter. The chain
# would boot fine, with an unaccounted balance nobody chose and a supply the
# step 9.5 reconciliation cannot see, since it reads the genesis file rather
# than post-InitChain state.
unexpected_types=$(jq -r '.tx.msg[]["@type"]' "$GENESIS_TXS_JSONL" | sort -u |
  grep -vE '^/vm\.(m_addpkg|m_run|m_call)$' || true)
if [ -n "$unexpected_types" ]; then
  die "unexpected msg types in genesis txs: $unexpected_types"
fi
# Read the signer fields by path, not by pattern: package file bodies ride
# in these same lines, so a fixture containing the literal "caller":"g1..."
# would otherwise be funded as a phantom fee payer (it pays no fee, so the
# range guard in step 8.4 would abort the build on it).
jq -r '.tx.msg[] | (.creator // empty), (.caller // empty)' "$GENESIS_TXS_JSONL" |
  awk 'NF' |
  sort -u >"$BALANCES_TMP_CREATOR_ADDRESSES"
if grep -qvE '^g1[0-9a-z]{38}$' "$BALANCES_TMP_CREATOR_ADDRESSES"; then
  die "extracted a fee payer that is not a bech32 address: $(grep -vE '^g1[0-9a-z]{38}$' "$BALANCES_TMP_CREATOR_ADDRESSES" | tr '\n' ' ')"
fi
addr_count=$(wc -l <"$BALANCES_TMP_CREATOR_ADDRESSES" | tr -d ' ')
# 91 txs cannot have zero signers: an empty extraction means the signer field
# moved, and every downstream count check would compare zero against zero.
if [ "$addr_count" -eq 0 ]; then
  die "no creator/caller extracted from $GENESIS_TXS_JSONL — the signer field shape changed"
fi
print_substep "8.2" "Found $addr_count unique creator/caller addresses"

# Overlap rules for mainnet (the final balance sheet keeps one entry per
# address, last write wins, so every overlap must be resolved explicitly;
# the allocation sheet itself was fetched and verified at step 2):
#   - a fee payer MAY also hold an allocation: its final entry becomes
#     allocation + measured burn (summed during the measure run below),
#     so it lands at exactly its allocation once the genesis txs execute;
#   - a VESTED_ACCOUNTS entry may NOT hold an allocation: the sheet is the
#     single source of schedules (decided), so an address in both inputs
#     dies loudly — fix the schedule upstream in the sheet instead;
#   - vested entries must be well-formed and mutually unique, and may not
#     be fee payers (unchanged from the testnet builders).
VESTED_ADDRS_FILE="$BALANCES_TMP_DIR/vested-addrs.txt"
{
  for vested in "${VESTED_ACCOUNTS[@]}"; do
    # Balance.Parse trims whitespace, so a padded entry would parse fine
    # downstream while its extracted address silently misses the guard —
    # reject malformed entries here instead of failing open.
    case "$vested" in
    *[[:space:]]*) die "malformed VESTED_ACCOUNTS entry (contains whitespace): '$vested'" ;;
    *=*) ;;
    *) die "malformed VESTED_ACCOUNTS entry (missing '='): '$vested'" ;;
    esac
    echo "${vested%%=*}"
  done
} >"$VESTED_ADDRS_FILE"
vested_dupes=$(sort "$VESTED_ADDRS_FILE" | uniq -d)
if [ -n "$vested_dupes" ]; then
  die "address(es) appearing more than once in VESTED_ACCOUNTS: $vested_dupes"
fi
while IFS= read -r vested_addr; do
  if grep -qxF -- "$vested_addr" "$BALANCES_TMP_CREATOR_ADDRESSES"; then
    die "vested account $vested_addr is also a genesis-tx fee payer — its entry would be silently overwritten"
  fi
  if grep -q -- "^${vested_addr}=" "$ALLOCATION_TXT"; then
    die "vested account $vested_addr also holds an independence-day allocation — the sheet is the single source of schedules; fix it upstream instead of overriding here"
  fi
done <"$VESTED_ADDRS_FILE"

# Pre-flight the vested entries: `balances add` catches syntax errors
# (Balance.Parse) and `verify` catches schedule semantics (Balance.Verify
# — verify is its only caller), so a malformed entry dies here in seconds
# instead of after two temp-node runs.
if [ "${#VESTED_ACCOUNTS[@]}" -gt 0 ]; then
  VESTED_PREFLIGHT_GENESIS="$BALANCES_TMP_DIR/vested-preflight.json"
  VESTED_PREFLIGHT_SHEET="$BALANCES_TMP_DIR/vested-preflight-sheet.txt"
  for vested in "${VESTED_ACCOUNTS[@]}"; do
    echo "$vested"
  done >"$VESTED_PREFLIGHT_SHEET"
  run "$GNOGENESIS_BIN" generate -chain-id "$CHAIN_ID" -genesis-time "$GENESIS_TIME" -output-path "$VESTED_PREFLIGHT_GENESIS" >/dev/null 2>&1
  run "$GNOGENESIS_BIN" balances add -balance-sheet "$VESTED_PREFLIGHT_SHEET" -genesis-path "$VESTED_PREFLIGHT_GENESIS" >/dev/null
  # verify refuses an empty validator set; borrow the first founding
  # validator to make the throwaway genesis verifiable.
  read -r pf_name pf_power pf_address pf_pub_key <<<"${INITIAL_VALSET[0]}"
  run "$GNOGENESIS_BIN" validator add -name "$pf_name" -power "$pf_power" -address "$pf_address" -pub-key "$pf_pub_key" -genesis-path "$VESTED_PREFLIGHT_GENESIS" >/dev/null
  run "$GNOGENESIS_BIN" verify -genesis-path "$VESTED_PREFLIGHT_GENESIS" >/dev/null
  rm -f "$VESTED_PREFLIGHT_GENESIS" "$VESTED_PREFLIGHT_SHEET"
fi

# Vested entries also ride the temp-node sheets so both measurement runs
# exercise vesting-account creation in InitChain. They are not fee payers
# (the guard above), so they cannot perturb the burn measurement, and the
# zero-verify loops iterate fee payers only.
append_vested_to_sheet() {
  local vested
  for vested in "${VESTED_ACCOUNTS[@]}"; do
    echo "$vested" >>"$BALANCES_TMP_FILE"
  done
}

print_substep "8.3" "Generating over-provisioned balances..."
while IFS= read -r addr; do
  echo "${addr}=${INITIAL_BALANCE}ugnot" >>"$BALANCES_TMP_FILE"
done <"$BALANCES_TMP_CREATOR_ADDRESSES"
append_vested_to_sheet

# Helper: spin up a temp node with the current genesis + balance sheet.
# Sets NODE_PID; aborts if the node doesn't come up in NODE_TIMEOUT seconds.
start_temp_node() {
  local run_label="$1"
  rm -rf "$BALANCES_TMP_GNOLAND_DATA" "$BALANCES_TMP_GENESIS"
  NODE_RPC_PORT=$(pick_free_port)
  NODE_P2P_PORT=$((NODE_RPC_PORT + 1))
  NODE_RPC_ADDR="127.0.0.1:$NODE_RPC_PORT"

  run "$GNOGENESIS_BIN" generate -chain-id "$CHAIN_ID" -genesis-time "$(date +%s)" -output-path "$BALANCES_TMP_GENESIS"
  run "$GNOGENESIS_BIN" txs add sheets "$GENESIS_TXS_JSONL" -genesis-path "$BALANCES_TMP_GENESIS"
  run "$GNOGENESIS_BIN" balances add -balance-sheet "$BALANCES_TMP_FILE" -genesis-path "$BALANCES_TMP_GENESIS"

  # The shipped chain replays the genesis txs UNDER the §126 lock — InitGenesis
  # applies restricted_denoms before tx replay — so the temp node must too, or
  # a future genesis tx that moves ugnot through the restricted path (a -send,
  # or realm code calling banker.SendCoins) would pass both measurement runs
  # here and then panic InitChain on every mainnet validator. Exemptions stay
  # empty: strictly more restrictive, and no genesis tx needs one today —
  # one that does should fail loudly here first.
  temp_rd_json=$(printf '%s\n' "${RESTRICTED_DENOMS[@]}" | jq -R -s -c 'split("\n") | map(select(length > 0))')
  jq --argjson rd "$temp_rd_json" '.app_state.bank.params.restricted_denoms = $rd' \
    "$BALANCES_TMP_GENESIS" >"$BALANCES_TMP_GENESIS.locked"
  mv "$BALANCES_TMP_GENESIS.locked" "$BALANCES_TMP_GENESIS"
  if [ "$(jq -c '.app_state.bank.params.restricted_denoms' "$BALANCES_TMP_GENESIS")" != "$temp_rd_json" ]; then
    die "temp-node genesis did not take restricted_denoms=$temp_rd_json — measurement would run without the shipped transfer lock"
  fi
  # Same code-submission params as the shipping genesis. The measurement
  # replays the genesis txs, which are exempt from the policy via
  # IsGenesisReplay, so this changes no measured amount -- it is here because
  # the parity assertion below compares vm.params and would otherwise fail the
  # moment the shipping genesis gets them.
  apply_code_submission_params "$BALANCES_TMP_GENESIS"

  # The measured burns are dominated by storage deposits priced by the
  # chain's vm/auth params. This genesis is regenerated rather than copied
  # (a future GENESIS_TIME would stall the temp node), so assert its
  # fee-governing params equal the shipping genesis AS IT STANDS NOW —
  # otherwise every measured amount is wrong by the parameter ratio and run 2
  # would agree with it. Step 9.3 then adds auth.params.unrestricted_addrs and
  # the same restricted_denoms to the shipping genesis, deliberately after
  # this point: neither prices a fee (both fee collection and storage deposits
  # go through the bank keeper's unrestricted path), and comparing them here
  # would fail on the exemption list, which cannot affect the measurement.
  # vm/auth are plain `gnogenesis generate` defaults today, so this only
  # starts biting when the inert-package params land — patch vm.params on both
  # genesis files and this assertion holds you to it.
  params_parity_lhs=$(jq -cS '[.app_state.vm.params, .app_state.auth.params]' "$BALANCES_TMP_GENESIS")
  if [ "$params_parity_lhs" = "[null,null]" ]; then
    die "temp-node genesis carries no vm/auth params (app_state moved?) — the parity check below would compare nothing"
  fi
  if [ "$params_parity_lhs" != \
    "$(jq -cS '[.app_state.vm.params, .app_state.auth.params]' "$GENESIS_FILE")" ]; then
    die "temp-node genesis fee params differ from the shipping genesis — measurement would be invalid"
  fi
  run "$GNOLAND_BIN" config init -config-path "$BALANCES_TMP_GNOLAND_DATA/config/config.toml"
  run "$GNOLAND_BIN" config set rpc.laddr "tcp://$NODE_RPC_ADDR" -config-path "$BALANCES_TMP_GNOLAND_DATA/config/config.toml"
  run "$GNOLAND_BIN" config set p2p.laddr "tcp://127.0.0.1:$NODE_P2P_PORT" -config-path "$BALANCES_TMP_GNOLAND_DATA/config/config.toml"
  run "$GNOLAND_BIN" secrets init -data-dir "$BALANCES_TMP_GNOLAND_DATA/secrets"
  run "$GNOGENESIS_BIN" validator add \
    --address "$("$GNOLAND_BIN" secrets get validator_key.address --raw -data-dir "$BALANCES_TMP_GNOLAND_DATA/secrets")" \
    --pub-key "$("$GNOLAND_BIN" secrets get validator_key.pub_key --raw -data-dir "$BALANCES_TMP_GNOLAND_DATA/secrets")" \
    --name balance_generator \
    --power 1 \
    -genesis-path "$BALANCES_TMP_GENESIS"

  printf "  Starting node (%s)...\n" "$run_label"
  "$GNOLAND_BIN" start --skip-genesis-sig-verification -data-dir "$BALANCES_TMP_GNOLAND_DATA" -genesis "$BALANCES_TMP_GENESIS" >"$BALANCES_TMP_GNOLAND_LOG" 2>&1 &
  NODE_PID=$!

  # Readiness = the deployer account visible in committed state, not just
  # an answering RPC (see account_in_state). Balance reads are only
  # trustworthy from that point on.
  local elapsed=0
  while [ "$elapsed" -lt "$NODE_TIMEOUT" ]; do
    if ! kill -0 "$NODE_PID" 2>/dev/null; then
      echo "ERROR: Node stopped unexpectedly. Last log lines:" >&2
      tail -20 "$BALANCES_TMP_GNOLAND_LOG" >&2
      exit 1
    fi
    if account_in_state "$DEPLOYER_ADDR"; then
      printf "  Node ready (%ss)\n" "$elapsed"
      return
    fi
    sleep 1
    elapsed=$((elapsed + 1))
  done
  kill "$NODE_PID" 2>/dev/null || true
  echo "ERROR: Node did not start within ${NODE_TIMEOUT}s." >&2
  printf 'last probe response: %s\n' \
    "$("$GNOKEY_BIN" query -remote "$NODE_RPC_ADDR" "auth/accounts/$DEPLOYER_ADDR" 2>&1 || true)" >&2
  echo "Last log lines:" >&2
  tail -20 "$BALANCES_TMP_GNOLAND_LOG" >&2
  exit 1
}

stop_temp_node() {
  kill "$NODE_PID" 2>/dev/null || true
  wait "$NODE_PID" 2>/dev/null || true
  NODE_PID=""
}

# account_in_state ADDR → success if ADDR's account exists in committed
# state. Queries read the last committed block: while InitChain is still
# executing (or block 1 is not yet committed) the store reads empty and
# auth/accounts answers `data: null` with exit 0, so a bare exit-status
# probe cannot tell "node up" from "state committed" — the account JSON
# can. Genesis state (balance credits included) commits atomically with
# block 1, so any funded address proves the whole genesis is readable.
# Matched with a bash pattern, not a pipe to grep: grep -q closing the
# pipe early can SIGPIPE the producer, which pipefail would report as a
# failed probe despite a match.
account_in_state() {
  local addr="$1" out
  out=$("$GNOKEY_BIN" query -remote "$NODE_RPC_ADDR" "auth/accounts/$addr" 2>/dev/null || true)
  case "$out" in
  *"\"address\": \"$addr\""*) return 0 ;;
  *) return 1 ;;
  esac
}

# query_balance ADDR → echoes ugnot amount.
# An empty `data:` response is how a zero balance is encoded (empty
# coins) — but it is also what an unreadable store returns. A zero is
# only echoed after account_in_state proves the address exists in
# committed state; otherwise the build dies.
query_balance() {
  local addr="$1"
  local retry=0
  while [ "$retry" -lt "$NODE_TIMEOUT" ]; do
    if ! kill -0 "$NODE_PID" 2>/dev/null; then
      echo "ERROR: Node stopped unexpectedly during balance query. Last log lines:" >&2
      tail -20 "$BALANCES_TMP_GNOLAND_LOG" >&2
      exit 1
    fi
    local out
    out=$("$GNOKEY_BIN" query -remote "$NODE_RPC_ADDR" "bank/balances/$addr" 2>&1 || true)
    # case instead of `echo | grep -q`, and no `| head -1` after the sed:
    # early-exit consumers are the SIGPIPE-under-pipefail shape that
    # account_in_state was rewritten to avoid. gnokey prints exactly one
    # `data:` line; if that ever changes, the multi-line payload fails the
    # exact-match parses below loudly instead of being truncated silently.
    case "$out" in
    'data:'* | *$'\n''data:'*)
      local payload
      payload=$(echo "$out" | sed -n 's/^data: //p')
      if [ "$payload" = '""' ]; then
        # Empty coins encode a zero balance — but an unreadable store
        # answers identically. Only trust the zero if the account
        # provably exists in committed state.
        account_in_state "$addr" ||
          die "balance for $addr reads zero but its account is not in committed state"
        echo 0
        return
      fi
      # Exactly one ugnot coin and nothing else: a second denomination on
      # a fee payer must abort the measurement, not parse as 0.
      local r
      r=$(echo "$payload" | sed -n 's/^"\([0-9][0-9]*\)ugnot"$/\1/p')
      [ -n "$r" ] || die "unparseable balance for $addr: $payload (expected a single ugnot coin or empty)"
      echo "$r"
      return
      ;;
    esac
    sleep 1
    retry=$((retry + 1))
  done
  echo "ERROR: Could not query balance for $addr after ${NODE_TIMEOUT}s." >&2
  exit 1
}

start_temp_node "run 1: measure gas costs"
print_substep "8.4" "Querying remaining balances..."
rm -f "$BALANCES_TMP_FILE"
EXPECTED_REMAINDERS="$BALANCES_TMP_DIR/expected-remainders.txt"
OVERLAP_ADDRS="$BALANCES_TMP_DIR/allocation-overlap-addrs.txt"
: >"$EXPECTED_REMAINDERS"
: >"$OVERLAP_ADDRS"
merged_alloc_total=0
while IFS= read -r addr; do
  remaining=$(query_balance "$addr")
  # A fee payer funded with the float and charged at least one fee must
  # land strictly between 0 and INITIAL_BALANCE.
  if [ "$remaining" -eq 0 ] || [ "$remaining" -ge "$INITIAL_BALANCE" ]; then
    die "fee payer $addr reads $remaining ugnot remaining in the measure run (float: $INITIAL_BALANCE) — node state not readable or fees not charged"
  fi
  final=$((INITIAL_BALANCE - remaining))
  # A fee payer that also holds an allocation gets ONE sheet entry of
  # allocation + burn: post-genesis it lands at exactly its allocation
  # instead of zero (the collision gnoland1 left as an open TODO). The
  # merge happens HERE, before run 2, so the verify run replays the exact
  # entries that ship and checks each expected remainder.
  rc=0
  alloc_line=$(grep -m1 -- "^${addr}=" "$ALLOCATION_TXT") || rc=$?
  if [ "$rc" -gt 1 ]; then
    die "grep failed reading the allocation sheet for $addr (exit $rc)"
  fi
  if [ "$rc" -eq 0 ]; then
    # An allocation row is `<addr>=<amount>ugnot[;<suffix>]`, and since
    # independence-day #72 turned the §132 pass on by default the suffix is
    # normally a vesting schedule. At the 0108ede pin all three
    # allocation-holding genesis fee payers carry one:
    #
    #   g125em6arxsnj49vx35f0n0z34putv5ty3376fg5      10,000 GNOT
    #   g1kfd9f5zlvcvy6aammcmqswa7cyjpu2nyt9qfen      10,000 GNOT
    #   g1manfred47kzduec920z88wfr64ylksmdcedlf5     111,000 GNOT
    #
    # all three 96%-vesting, 1789084800->1852243200. They were plain balances
    # at the 9d1cfde pin, which is why the merge below only had to handle bare
    # amounts until now.
    #
    # The merge adds the burn to the LIQUID part and carries the schedule
    # through VERBATIM: `<addr>=<alloc+burn>ugnot;vesting=<unchanged>,...`.
    # The locked amount does not move, so the extra coins are spendable
    # immediately (tm2 treats total-minus-vesting as liquid) and the account
    # lands post-genesis at exactly its allocation under exactly its lockup.
    #
    # Leaving a vested fee payer unmerged is NOT the same as the plain-balance
    # case: fees are collected with SendCoinsUnrestricted, which bypasses the
    # schedule, so it would pay its genesis-tx fees out of LOCKED coins and
    # land BELOW its allocation.
    alloc_rhs="${alloc_line#*=}"
    case "$alloc_rhs" in
    *';'*)
      fp_alloc="${alloc_rhs%%;*}"
      alloc_suffix=";${alloc_rhs#*;}"
      ;;
    *)
      fp_alloc="$alloc_rhs"
      alloc_suffix=""
      ;;
    esac
    fp_alloc="${fp_alloc%ugnot}"
    # The parse is load-bearing for the money: a shape this does not
    # understand must stop the build, not be summed as if it were an amount.
    case "$fp_alloc" in
    '' | *[!0-9]*) die "fee payer $addr: cannot read an amount out of the allocation row '$alloc_line' — the sheet's row shape changed" ;;
    esac
    case "$alloc_suffix" in
    '' | ';vesting='*) ;;
    *) die "fee payer $addr: allocation row '$alloc_line' carries an unrecognised suffix '$alloc_suffix' — only ;vesting= is understood by the merge" ;;
    esac
    # Tracked for the step 9.5 reconciliation: these allocations are counted
    # in the fee-payer sheet, and their rows leave the allocation sheet.
    merged_alloc_total=$((merged_alloc_total + fp_alloc))
    if [ -n "$alloc_suffix" ]; then
      printf "    %s = %s ugnot (+ %s allocation, schedule carried)\n" "$addr" "$final" "$fp_alloc"
    else
      printf "    %s = %s ugnot (+ %s allocation)\n" "$addr" "$final" "$fp_alloc"
    fi
    merged_row="${addr}=$((final + fp_alloc))ugnot${alloc_suffix}"
    # The schedule must survive the merge byte for byte. A dropped or rewritten
    # suffix is the failure that unlocks a §132 balance at block 1, and it would
    # otherwise be invisible: the totals still reconcile either way.
    if [ -n "$alloc_suffix" ] && [ "${merged_row#*ugnot}" != "$alloc_suffix" ]; then
      die "fee payer $addr: the vesting schedule did not survive the merge (row '$merged_row', expected suffix '$alloc_suffix')"
    fi
    echo "$merged_row" >>"$BALANCES_TMP_FILE"
    echo "${addr} ${fp_alloc}" >>"$EXPECTED_REMAINDERS"
    echo "$addr" >>"$OVERLAP_ADDRS"
  else
    printf "    %s = %s ugnot\n" "$addr" "$final"
    echo "${addr}=${final}ugnot" >>"$BALANCES_TMP_FILE"
    echo "${addr} 0" >>"$EXPECTED_REMAINDERS"
  fi
done <"$BALANCES_TMP_CREATOR_ADDRESSES"
append_vested_to_sheet
stop_temp_node

start_temp_node "run 2: verify expected remainders"
print_substep "8.5" "Verifying fee payers land at their expected remainders..."
all_expected=true
verified_count=0
while IFS=' ' read -r addr expected; do
  verified_count=$((verified_count + 1))
  remaining=$(query_balance "$addr")
  if [ "$remaining" -ne "$expected" ]; then
    printf "    FAIL: %s has %sugnot remaining (expected %s)\n" "$addr" "$remaining" "$expected"
    all_expected=false
  else
    printf "    ok: %s (%s)\n" "$addr" "$expected"
  fi
done <"$EXPECTED_REMAINDERS"
stop_temp_node

if [ "$all_expected" != true ]; then
  die "Some fee payers did not land at their expected remainder. Check $BALANCES_TMP_FILE."
fi
# An empty expected-remainders file would sail through the loop above and
# report verified costs having checked nothing.
if [ "$verified_count" -ne "$addr_count" ]; then
  die "verified $verified_count fee payers but $addr_count were funded — $EXPECTED_REMAINDERS is short"
fi
print_substep "8.6" "All fee payers land at their expected remainders — costs verified"
# The temp sheet also carries the vested entries (so the measurement runs
# exercise their creation); keep only the measured fee-payer lines here —
# step 9 appends the vested entries itself.
#
# Selected BY ADDRESS, not by the absence of `;vesting=`. A fee payer whose
# allocation carries a §132 schedule is merged above into a row that keeps the
# schedule, so a `grep -v ';vesting='` here would drop exactly those rows —
# and with their allocation rows already stripped from the allocation sheet,
# the addresses would vanish from genesis entirely.
sed 's/^/^/; s/$/=/' "$BALANCES_TMP_CREATOR_ADDRESSES" >"$BALANCES_TMP_DIR/feepayers.pat"
rc=0
grep -f "$BALANCES_TMP_DIR/feepayers.pat" "$BALANCES_TMP_FILE" >"$DEPLOYER_BALANCES" || rc=$?
if [ "$rc" -gt 1 ]; then
  die "grep failed selecting the fee-payer rows out of $BALANCES_TMP_FILE (exit $rc)"
fi
deployer_rows=$(wc -l <"$DEPLOYER_BALANCES" | tr -d ' ')
if [ "$deployer_rows" -ne "$addr_count" ]; then
  die "the measured fee-payer sheet has $deployer_rows rows for $addr_count fee payers — a row was dropped or duplicated between the measurement and the shipped sheet"
fi
# This sheet can now carry §132 schedules (a fee payer whose allocation is
# vested keeps its schedule through the merge), so it gets the same
# fully-locked-at-genesis check the allocation sheet got in step 2. The merge
# copies the schedule verbatim, so this can only fire if GENESIS_TIME moved
# past a cliff — which is exactly when it should.
assert_vesting_locked_at_genesis "$DEPLOYER_BALANCES" "merged-fee-payers"
# This small sheet is where the MEASURED numbers live; locking it localises a
# future genesis.json mismatch to "the burn changed" instead of "196MB of
# bytes changed somewhere".
verify_checksum "$DEPLOYER_BALANCES"

# Strip the merged addresses from the allocation sheet in a single pass,
# then assert the fee-payer and allocation sheets are disjoint — one
# entry per address in the final genesis, and last-write-wins would
# silently drop one side of any leftover overlap.
ALLOCATION_STRIPPED="$WORK_DIR/allocation_stripped.txt"
merged_count=$(wc -l <"$OVERLAP_ADDRS" | tr -d ' ')
if [ -s "$OVERLAP_ADDRS" ]; then
  sed 's/^/^/; s/$/=/' "$OVERLAP_ADDRS" >"$OVERLAP_ADDRS.pat"
  rc=0
  grep -v -f "$OVERLAP_ADDRS.pat" "$ALLOCATION_TXT" >"$ALLOCATION_STRIPPED" || rc=$?
  if [ "$rc" -gt 1 ]; then
    die "grep failed stripping merged addresses from the allocation sheet (exit $rc)"
  fi
else
  cp "$ALLOCATION_TXT" "$ALLOCATION_STRIPPED"
fi
sheet_overlap=$(comm -12 <(cut -d= -f1 "$ALLOCATION_STRIPPED" | sort) <(cut -d= -f1 "$DEPLOYER_BALANCES" | sort))
if [ -n "$sheet_overlap" ]; then
  die "fee-payer and allocation sheets are not disjoint after the merge: $sheet_overlap"
fi
rm -f "$ALLOCATION_TXT"

# ---- Step 9: Add validators + balances, verify, move into place

print_step_header 9 "$TOTAL_STEPS" "Add validators + balances, verify genesis"

print_substep "9.1" "Adding the initial validator set..."
for validator in "${INITIAL_VALSET[@]}"; do
  read -r name power address pub_key <<<"$validator"
  printf "    %s (power=%s, %s)\n" "$name" "$power" "$address"
  run "$GNOGENESIS_BIN" validator add -name "$name" -power "$power" -address "$address" -pub-key "$pub_key" --genesis-path "$GENESIS_FILE"
done

# Fee payers (exact burn, or allocation + burn for allocation holders) +
# the vested accounts, one sheet. One entry per address (last write
# wins), so every overlap was resolved or rejected by step 8. Vested
# entries already carry the full balance-sheet syntax (amount + vesting
# schedule) and are appended verbatim.
FULL_BALANCES_FILE="$WORK_DIR/balances.txt"
cp "$DEPLOYER_BALANCES" "$FULL_BALANCES_FILE"
for vested in "${VESTED_ACCOUNTS[@]}"; do
  echo "$vested" >>"$FULL_BALANCES_FILE"
done
balance_count=$(wc -l <"$FULL_BALANCES_FILE" | tr -d ' ')
alloc_stripped_count=$(wc -l <"$ALLOCATION_STRIPPED" | tr -d ' ')
print_substep "9.2" "Adding $balance_count balances (fee payers + ${#VESTED_ACCOUNTS[@]} vested; $merged_count hold allocations) + $alloc_stripped_count allocation accounts..."
run "$GNOGENESIS_BIN" balances add -balance-sheet "$FULL_BALANCES_FILE" --genesis-path "$GENESIS_FILE" >/dev/null
# The allocation sheet goes in last, as its own `balances add` call: it
# bloats the genesis and makes subsequent gnogenesis calls slow (the
# gnoland1 pattern). Fee payers holding allocations were already stripped
# from it at step 8.6, so the two sheets are disjoint.
run "$GNOGENESIS_BIN" balances add -balance-sheet "$ALLOCATION_STRIPPED" --genesis-path "$GENESIS_FILE" >/dev/null

# ---- §126 transfer lock + exemption list
# Two independent knobs, and BOTH are required:
#
#   bank.params.restricted_denoms   turns the global lock on. Empty means
#                                   canSendCoins() returns true before it ever
#                                   looks at the whitelist.
#   auth.params.unrestricted_addrs  the addresses exempt from it.
#
# Setting only the second is the quiet failure mode: the genesis looks like it
# honours §126, and every address can transfer.
#
# This is applied AFTER the balance sheets are in, because InitChain requires
# every unrestricted address to already be a genesis account (checked at step
# 2.4, re-checked here against what actually shipped).
print_substep "9.3" "Applying the §126 transfer lock and exemption list..."
unrestricted_json=$(jq -R -s -c 'split("\n") | map(select(length > 0))' "$UNRESTRICTED_ADDRS_TXT")
restricted_json=$(printf '%s\n' "${RESTRICTED_DENOMS[@]}" | jq -R -s -c 'split("\n") | map(select(length > 0))')
jq --argjson unres "$unrestricted_json" --argjson rd "$restricted_json" \
  '.app_state.auth.params.unrestricted_addrs = $unres
   | .app_state.bank.params.restricted_denoms = $rd' \
  "$GENESIS_FILE" >"$GENESIS_FILE.locked"
mv "$GENESIS_FILE.locked" "$GENESIS_FILE"

# Prove it landed, and prove every exempt address is really in the sheet that
# shipped -- not the one we downloaded. A mismatch here is an InitChain panic.
applied_unres=$(jq -r '.app_state.auth.params.unrestricted_addrs | length' "$GENESIS_FILE")
applied_rd=$(jq -r '.app_state.bank.params.restricted_denoms | join(",")' "$GENESIS_FILE")
if [ "$applied_unres" != "$unrestricted_count" ]; then
  die "unrestricted_addrs did not apply: expected $unrestricted_count, genesis has $applied_unres"
fi
if [ -z "$applied_rd" ]; then
  die "restricted_denoms is empty — the exemption list would be inert and §126 unmet"
fi
# Non-emptiness is not the property that matters: canSendCoins only refuses a
# transfer whose coins contain a restricted denom (tm2/pkg/sdk/bank/keeper.go),
# so a single typo — "ugnots", or the "gnot" everyone writes in prose — leaves
# every ugnot transfer open while this step still prints a lock. Every genesis
# balance is denominated in ugnot (enforced by the sheet format check at 2.3),
# so that is the denom the lock has to name.
case ",$applied_rd," in
*,ugnot,*) ;;
*) die "restricted_denoms is [$applied_rd], which does not include ugnot — every ugnot transfer would be permitted and §126 unmet" ;;
esac
# Every exempt address must be a genesis account or InitChain panics
# (gno.land/pkg/gnoland/app.go: "unrestricted address must be one of the
# genesis accounts"). Step 2 checked the downloaded sheet; this checks the
# balances that actually shipped — app_state carries no account list, the
# balances ARE the accounts.
genesis_addrs=$(jq -r '.app_state.balances[]' "$GENESIS_FILE" | cut -d= -f1 | sort -u)
if [ -z "$genesis_addrs" ]; then
  die "the shipped genesis has no balances — nothing was added, or app_state.balances moved"
fi
not_funded=$(printf '%s\n' "$genesis_addrs" | comm -13 - <(sort -u "$UNRESTRICTED_ADDRS_TXT"))
if [ -n "$not_funded" ]; then
  die "unrestricted addresses missing from the shipped genesis balances (InitChain would panic): $not_funded"
fi
# ---- Code-submission policy (inert) ----
# Applied BEFORE step 8 (the params-parity assertion there requires the
# shipping and measurement genesis to carry identical vm.params); read back
# here from the artifact that ships.
applied_policy=$(jq -r '.app_state.vm.params.code_submission_policy' "$GENESIS_FILE")
applied_approvers=$(jq -r '.app_state.vm.params.pkg_approvers | length // 0' "$GENESIS_FILE")
applied_runners=$(jq -r '.app_state.vm.params.run_submitters | length // 0' "$GENESIS_FILE")
print_substep "9.4" "Transfer lock: restricted_denoms=[$applied_rd], $applied_unres addresses exempt"
print_substep "9.5" "$(printf 'Code submission: policy=%s, %d approver(s), MsgRun gate %s (%d)' \
  "$applied_policy" "$applied_approvers" \
  "$([ "$applied_runners" -gt 0 ] && echo ARMED || echo OPEN)" "$applied_runners")"

# ---- Reconcile account count and total supply against the pinned sheet ----
# Everything above counts lines in files the script itself wrote. This reads
# the account count and the supply back out of the genesis that ships and ties
# both to the two independent sources they must come from: the sha256-verified
# independence-day sheet, and the burn measured on the temp node. A truncated
# `balances add`, a sheet added twice, a mis-stripped merge row or a jq patch
# that dropped part of app_state all break one of these identities.
#
# Supply is not the allocation total: the fee payers are funded with the exact
# gas + storage cost of the genesis txs, which is minted on top of it and burnt
# again as the txs execute.
#
# The two launch parameters nothing downstream would notice: `gnogenesis
# verify` has no opinion on either, and the temp nodes deliberately run with
# their own generated genesis (their own chain-id-free time), so a `generate`
# call that silently dropped one would ship a chain nobody can join.
genesis_chain_id=$(jq -r '.chain_id' "$GENESIS_FILE")
if [ "$genesis_chain_id" != "$CHAIN_ID" ]; then
  die "shipped genesis carries chain_id '$genesis_chain_id', expected '$CHAIN_ID'"
fi
genesis_time_shipped=$(jq -r '.genesis_time' "$GENESIS_FILE")
genesis_time_want=$(date -u -r "$GENESIS_TIME" +%Y-%m-%dT%H:%M:%SZ 2>/dev/null ||
  date -u -d "@$GENESIS_TIME" +%Y-%m-%dT%H:%M:%SZ)
if [ "$genesis_time_shipped" != "$genesis_time_want" ]; then
  die "shipped genesis carries genesis_time '$genesis_time_shipped', expected '$genesis_time_want' (GENESIS_TIME=$GENESIS_TIME)"
fi

genesis_accounts=$(jq -r '.app_state.balances | length' "$GENESIS_FILE")
genesis_supply=$(jq -r '.app_state.balances[]' "$GENESIS_FILE" | sheet_total -)
assert_exact_sum "$genesis_supply" "the genesis supply"

stripped_total=$(sheet_total "$ALLOCATION_STRIPPED")
fee_payer_total=$(sheet_total "$DEPLOYER_BALANCES")
vested_total=0
for vested in "${VESTED_ACCOUNTS[@]}"; do
  vested_amount="${vested#*=}"
  vested_total=$((vested_total + ${vested_amount%%ugnot*}))
done
burn_total=$((fee_payer_total - merged_alloc_total))

expected_accounts=$((alloc_stripped_count + balance_count))
if [ "$genesis_accounts" -ne "$expected_accounts" ]; then
  die "genesis holds $genesis_accounts balance entries, expected $expected_accounts ($alloc_stripped_count allocation + $balance_count fee-payer/vested)"
fi
if [ "$alloc_stripped_count" -ne $((alloc_count - merged_count)) ]; then
  die "the stripped allocation sheet has $alloc_stripped_count rows, expected $((alloc_count - merged_count)) ($alloc_count in the pinned sheet less $merged_count merged into fee-payer entries)"
fi
if [ "$stripped_total" -ne $((alloc_total - merged_alloc_total)) ]; then
  die "the stripped allocation sheet is worth $stripped_total ugnot, expected $((alloc_total - merged_alloc_total)) ($alloc_total in the pinned sheet less the $merged_alloc_total merged into fee-payer entries) — the sheet lost the wrong rows or an amount changed after step 2 verified it"
fi
expected_supply=$((alloc_total + burn_total + vested_total))
if [ "$genesis_supply" -ne "$expected_supply" ]; then
  die "$(printf '%s\n%s\n%s\n%s' \
    "genesis supply is $genesis_supply ugnot, expected $expected_supply:" \
    "  allocation (pinned sheet)  $alloc_total" \
    "  fee-payer burn (measured)  $burn_total" \
    "  vested entries             $vested_total")"
fi
print_substep "9.6" "$(printf 'Reconciled: %s accounts, %s ugnot = %s allocation + %s burn + %s vested' \
  "$genesis_accounts" "$genesis_supply" "$alloc_total" "$burn_total" "$vested_total")"

print_substep "9.7" "Running gnogenesis verify..."
# -skip-signature-check: the names.Enable tx carries a post-sign caller
# patch and the valoper Register txs carry placeholder signatures, so
# per-tx signature verification cannot pass by design (nodes accept both
# under --skip-genesis-sig-verification). The remaining scope is shallow —
# amino decode, GenesisDoc/params sanity, stateless per-tx and per-balance
# format checks. Execution and fee-payer funding are proven by step 8's
# temp-node runs, not by this step.
run "$GNOGENESIS_BIN" verify -genesis-path "$GENESIS_FILE" -skip-signature-check

# Verify before moving: a mismatch must not clobber the previously-good
# (gitignored, so invisible to git status) genesis.json at the root.
verify_checksum "$GENESIS_FILE" genesis.json
print_substep "9.8" "Moving $GENESIS_FILE -> $FINAL_GENESIS"
mv "$GENESIS_FILE" "$FINAL_GENESIS"

# ---- Summary

PIPELINE_END_TS=$(date +%s)
PIPELINE_DURATION=$((PIPELINE_END_TS - PIPELINE_START_TS))
FINAL_SHA=$(sha256_of "$FINAL_GENESIS")
FINAL_BYTES=$(file_size "$FINAL_GENESIS")

printf '\n### mainnet build complete: genesis.json (%s, sha256=%s) ###\n' \
  "$(format_size "$FINAL_BYTES")" "$FINAL_SHA"
printf '    total pipeline time: %s\n' "$(format_duration "$PIPELINE_DURATION")"

# What produced these bytes. packages.gen.txt records package PATHS, not
# content, so the source tree is the rest of the input: two runs from
# different commits — or from a dirty tree — legitimately differ, and without
# this a future checksum mismatch is unexplainable.
# Both inputs count: examples/ is what gets deployed, and this folder is what
# assembles it. A commit hash printed next to a dirty tree would be a lie.
SOURCE_COMMIT=$(git -C "$REPO_ROOT" rev-parse HEAD 2>/dev/null || echo "unknown")
SOURCE_DIRTY=$(git -C "$REPO_ROOT" status --porcelain -- examples/ "$MAINNET_DIR" 2>/dev/null | wc -l | tr -d ' ')
printf '    source commit:       %s%s\n' "$SOURCE_COMMIT" \
  "$([ "$SOURCE_DIRTY" -gt 0 ] && printf ' + %s UNCOMMITTED path(s) in examples/ or this folder' "$SOURCE_DIRTY")"

# The checksum manifest is the only thing standing between "the bytes I
# reviewed" and "the bytes I am handing to four validators", so say plainly
# how much of this build it covered.
printf '    artifacts locked:    %s of %s\n' "$CHECKSUMS_LOCKED" "$((CHECKSUMS_LOCKED + CHECKSUMS_UNLOCKED))"
if [ "$CHECKSUMS_UNLOCKED" -gt 0 ]; then
  printf '\n    %s artifact(s) NOT locked — this build is not reproducible-verified.\n' "$CHECKSUMS_UNLOCKED"
  printf '    Paste into CHECKSUMS_DATA once every launch value is final:\n\n'
  printf '%s' "$CHECKSUMS_UNLOCKED_LINES" | sed 's/^/      /'
  printf '\n'
fi
