# test-13 govDAO scripts

Bash helpers that wrap common chain operations. Four scripts wrap a
single-tx GovDAO proposal (`create + vote + execute` inside one MsgRun,
which works because the sole T1 vote is 100 % of supermajority). The
fifth (`register-valoper.sh`) is signed by the operator themselves and
does not go through GovDAO — it's the on-ramp for the GovDAO
add-validator flow.

The GovDAO-signed scripts default to `GNOKEY_NAME=aeddi` (the T1 after
the phase-2 rotation). Override for testing or future rotations.

## Common environment variables

| Variable | Default | Notes |
|---|---|---|
| `GNOKEY_NAME` | `aeddi` | Local key name used to sign the tx |
| `CHAIN_ID` | `test-13` | |
| `REMOTE` | `127.0.0.1:26657` | RPC endpoint (override for non-local clusters) |
| `GAS_WANTED` | `50000000` | |
| `GAS_FEE` | `1000000ugnot` | |

## Scripts

### Validator on-ramp (two-step)

Promoting a new validator on test-13 is a deliberate two-step flow:
the operator self-registers their profile first, then T1 promotes
them into the active valset. This is enforced by `r/sys/validators/v3`,
which rejects any proposal referencing an operator not yet in
`valoperCache` — and `valoperCache` is only populated by
`r/gnops/valopers.Register`, whose runtime squat guard requires
`caller == operator`.

- **`register-valoper.sh <moniker> <description> <server_type> <operator_address> <signing_pubkey>`**
  — **signed by the operator** (so `GNOKEY_NAME` should be the
  operator's local key, NOT the T1). Calls
  `r/gnops/valopers.Register` to publish the profile and seed v3's
  `valoperCache`. `server_type` ∈ `{cloud, on-prem, data-center}`.
  Prerequisite for `add-validator.sh`.
- **`add-validator.sh <operator_address> [voting_power]`** — **signed
  by T1.** Proposes promoting the given operator to the active valset
  via `r/sys/validators/v3.NewValidatorProposalRequest` with
  `Power=voting_power` (default 1). The operator must already be in
  `valoperCache` (via `register-valoper.sh`) with `KeepRunning=true`
  (the default at Register time); v3 enforces both at
  proposal-creation time. `0` is rejected by the script — use
  `rm-validator.sh` to remove.
- **`rm-validator.sh <operator_address>`** — **signed by T1.**
  Proposes removing the given operator from the active valset
  (`Power=0`). Force-remove — unlike the facade in
  `r/gnops/valopers/proposal`, this does not require the operator to
  have set `KeepRunning=false`.

End-to-end example:

```bash
# 1. operator side, on their own machine:
GNOKEY_NAME=aeddi-1 ./register-valoper.sh \
  "aeddi-1" "Aeddi node #1" cloud \
  g1s2ht24e85qq3t66gc9sgdvk5kzc38yy68aaqvr \
  gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqfr74tgql2cvzadga2uts62v3f8a5dx66dauaq6sphg3ynuhgl286cce2mn

# 2. T1 side, after the operator has registered:
./add-validator.sh g1s2ht24e85qq3t66gc9sgdvk5kzc38yy68aaqvr 10

# Later, to remove:
./rm-validator.sh g1s2ht24e85qq3t66gc9sgdvk5kzc38yy68aaqvr
```

### Other proposals
- **`register-user.sh <username> <address>`** — grants `<username>` as
  a registered name for `<address>` via
  `r/sys/users.ProposeRegisterUser`. After execution, `<address>` can
  deploy under `gno.land/r/<username>/*` and `gno.land/p/<username>/*`
  (the `r/sys/names` verifier bridges to `r/sys/users.ResolveName`).
  The proposal bypasses `r/sys/namereg/v1`'s reserved-name blacklist
  and canonical-collision check, with a voter warning auto-injected
  into the proposal description on collision. Names must match
  `^[a-z][a-z0-9]*([_-][a-z0-9]+)*$`.
- **`unrestrict-account.sh <address> [<address>…]`** — adds one or
  more addresses to the unrestricted-accounts set via
  `r/sys/params.ProposeAddUnrestrictedAcctsRequest`. Unrestricted
  addresses can transfer ugnot even while bank transfers are
  restricted (the regime test-13 boots into).
