# Gnoland Validator Scenario Harness

This repo generates local Gnoland validator networks in Docker and runs scripted failure / recovery scenarios against them.

It is inspired by `../gno-val-test`, but the setup here is reusable and scenario-driven:

- each validator or sentry runs in its own container
- validators can optionally run with a controllable remote-signer sidecar
- the network is generated from a small Bash DSL
- scenarios can stop, restart, and reset nodes
- scenarios can deploy realms and submit transactions with `gnokey`
- sentry-based topologies are supported, including sentry container recreation to force a new container IP while validators keep dialing the same DNS name

## Prerequisites

- `docker`
- `docker compose`
- `jq`
- `curl`
- `bash` (4+)

## Build The Local Images

The scripts expect three local Docker images:

- `gno-val-scenario-core:local`: built from the root `Dockerfile` `all` target; contains `gnoland` and `gnokey`
- `gnogenesis:local`: built from the root `Dockerfile` `gnocontribs` target
- `valsignerd:local`: built from `misc/val-scenarios/Dockerfile`; contains only the scenario signer sidecar

```bash
make build-images
```

Override image tags with `IMAGE=...`, `GNOKEY_IMAGE=...`, `GNOGENESIS_IMAGE=...`, and `VALSIGNER_IMAGE=...` if needed. By default, `GNOKEY_IMAGE` is the same image as `IMAGE`.

To build images from a GitHub fork, set `GH_USER`. `GH_REPO` defaults to `gno` and `GH_BRANCH` defaults to `master`. Image tags are derived automatically as `<base>:<GH_USER>-<GH_BRANCH>` (slashes in the branch name become dashes), so multiple versions can coexist without overwriting each other.

```bash
make build-images GH_USER=gnolang
# -> gno-val-scenario-core:gnolang-master, gnogenesis:gnolang-master, valsignerd:gnolang-master

make build-images GH_USER=gnolang GH_REPO=gno GH_BRANCH=feat/my-branch
# -> gno-val-scenario-core:gnolang-feat-my-branch, gnogenesis:gnolang-feat-my-branch, valsignerd:gnolang-feat-my-branch
```

When `GH_USER` is set, all images build from the fetched remote branch by default. You can override each source checkout independently:

- `CORE_GNO_ROOT`: source for the core image (`gnoland` and `gnokey`)
- `GNOGENESIS_GNO_ROOT`: source for the `gnogenesis` image
- `VALSIGNER_GNO_ROOT`: source for the `valsignerd` image

This is useful when testing a branch that does not contain every scenario tool. For example, build the chain binaries from a remote branch but use the local `valsignerd`:

```bash
GH_USER=moul GH_BRANCH=feat/valset-params-v3 VALSIGNER_GNO_ROOT=$PWD make build-images
```

The repository is cloned once to `/tmp/gno-remote-build` and reused across subsequent builds. To force a fresh clone, run `make fetch-remote` with the same variables.

To run a scenario against previously built fork images, pass the matching tag variables through the Makefile:

```bash
make scenario-12 GH_USER=gnolang GH_BRANCH=feat/my-branch
```

## Run A Scenario

```bash
make test          # run scenarios marked SCENARIO_CI=true
make test-local    # run scenarios marked SCENARIO_CI=false
make test-all      # run all scenarios
make scenario-01
make scenario-04
```

`make test-basics` / `make basics` are aliases for `make test-ci`.
`make test-advanced` / `make advanced` are aliases for `make test-local`.

Each run writes generated node data, keys, genesis, and compose output under:

```bash
/tmp/gno-val-tests/<scenario-name>/
```

By default the scenario tears containers down on exit but keeps the generated data. To keep the network running after the script exits:

```bash
KEEP_UP=1 ./scenarios/05_sentry_ip_rotation.sh
```

## Scenario Selection

All scenario scripts live in `scenarios/`. Each script declares whether it should run in CI:

- `SCENARIO_CI=true`: included in `.github/workflows/ci-val-scenarios.yml` and `make test`
- `SCENARIO_CI=false`: local-only, usually because the scenario needs `valsignerd`

### CI Scenarios

- `scenarios/01_four_validators_reset_three.sh`: start 4 validators, run 60s, stop/reset 3, restart them, run 60s again
- `scenarios/02_three_validators_restart_staggered.sh`: start 3 validators, stop all after 60s, restart one by one
- `scenarios/03_three_validators_restart_parallel.sh`: start 3 validators, stop all after 60s, restart all together
- `scenarios/04_counter_realm_churn.sh`: deploy a sample counter realm, submit transactions, reset one validator, continue submitting txs
- `scenarios/05_sentry_ip_rotation.sh`: run validators behind a sentry, recreate the sentry to force a new container IP, and verify the network keeps progressing
- `scenarios/06_gas_nondeterminism_check.sh`: restart a subset of validators, estimate addpkg gas on a warm node, and fail if the chain halts after the trigger tx
- `scenarios/07_four_validators_reset_one.sh`: start 4 validators, stop/reset/restart 1; 3/4 remain above the 2/3 threshold so the chain must keep advancing throughout
- `scenarios/08_five_validators_reset_two_below_consensus.sh`: start 5 validators, stop/reset 2; 3/5 drops below the 2/3 threshold so the chain must halt, then verify it resumes after both validators are restarted
- `scenarios/09_four_validators_safe_reset_one.sh`: same as 07 but uses a safe reset (db + wal only, `priv_validator_state` preserved) to avoid double signing
- `scenarios/10_four_validators_safe_reset_two_below_consensus.sh`: start 4 validators, safe-reset 2, verify the chain halts below consensus and resumes after restart
- `scenarios/11_weighted_voting_power_majority.sh`: 4 validators with voting power 10/1/1/1; val1 alone holds more than 2/3 of total power, so stopping val2-4 must not halt the chain
- `scenarios/12_duplicate_addr_in_val_proposal.sh`: single proposal with two entries for the same validator address; EndBlocker deduplicates, val1 ends up with VotingPower=5 and the chain keeps advancing
- `scenarios/13_duplicate_addr_across_proposals.sh`: two separate proposals in the same block targeting the same validator address; EndBlocker deduplicates, val1 ends up with VotingPower=5 and the chain keeps advancing
- `scenarios/17_govdao_add_remove_validator.sh`: add and remove a validator through GovDAO proposals using `r/sys/validators/v2`
- `scenarios/18_govdao_v3_add_remove_validator.sh`: add and remove a synced val4 node through `r/sys/validators/v3` GovDAO proposals (registers val4 in `r/gnops/valopers` first)

### Local-Only Valsignerd Scenarios

- `scenarios/14_four_validators_drop_proposals_with_signers.sh`: 4 validators with controllable signer sidecars; drop proposal signatures on all validators and assert consensus resumes after clearing the rules
- `scenarios/15_four_validators_drop_prevotes_thresholds.sh`: 4 validators with controllable signer sidecars; drop prevotes below and above quorum thresholds
- `scenarios/16_four_validators_precommit_delays_thresholds.sh`: 4 validators with controllable signer sidecars; delay precommits below and above `timeout_commit`

## Reusable Scenario API

Scenarios source `lib/scenario.sh` and use a small set of helpers:

- `scenario_init <name>`
- `gen_validator <name> [--rpc-port <port>] [--sentry <sentry-name>] [--controllable-signer] [--not-in-genesis]`
- `gen_sentry <name> [--rpc-port <port>]`
- `prepare_network`
- `start_all_nodes`
- `start_validator <name>`
- `stop_validator <name>`
- `reset_validator <name>`
- `wait_for_seconds <n>`
- `wait_for_blocks <node> <delta> <timeout>`
- `add_pkg <target-node> <pkgdir> <pkgpath>`
- `call_realm <target-node> <pkgpath> <func> [args...]`
- `do_transaction addpkg|call|run|send ...`
- `signer_state <validator>`
- `signer_drop <validator> proposal|prevote|precommit [height] [round]`
- `signer_delay <validator> proposal|prevote|precommit <duration> [height] [round]`
- `signer_clear <validator> [phase]`
- `rotate_sentry_ip <sentry-name>`
- `print_cluster_status`

`wait_for_seconds` is used instead of `wait` to avoid colliding with Bash's built-in `wait`.

## Controllable Signers

Pass `--controllable-signer` to `gen_validator` to launch a `valsignerd` sidecar for that validator. The validator itself still runs stock `gnoland`; only the signing path is redirected through the sidecar via the existing remote-signer configuration.

Each controllable validator gets:

- a sidecar service named `<validator>-signer`
- an HTTP control API on host port `<validator-rpc-port + 1>`
- a remote signer endpoint inside the compose network at `tcp://<validator>-signer:26659`

`prepare_network` writes an inventory file at:

```bash
/tmp/gno-val-tests/<scenario-name>/inventory.json
```

That file lists validator RPC URLs and signer control URLs for use by an external cockpit.

The sidecar currently supports live rules for:

- drop proposal signatures
- drop prevote signatures
- drop precommit signatures
- delay proposal / prevote / precommit signatures
- optional height / round scoping

This approach does not modify vote contents or proposal contents. It controls whether a validator signs, and when.

Example live control commands against a running scenario:

```bash
# Inspect current signer state.
curl -fsS http://127.0.0.1:26658/state | jq

# Drop all precommits from val1.
curl -fsS -X PUT http://127.0.0.1:26658/rules/precommit \
  -H 'Content-Type: application/json' \
  -d '{"action":"drop"}'

# Delay only round 0 prevotes at height 25 by 8 seconds.
curl -fsS -X PUT http://127.0.0.1:26658/rules/prevote \
  -H 'Content-Type: application/json' \
  -d '{"action":"delay","delay":"8s","height":25,"round":0}'

# Clear just the precommit rule.
curl -fsS -X DELETE http://127.0.0.1:26658/rules/precommit

# Clear all rules on that signer.
curl -fsS -X POST http://127.0.0.1:26658/reset
```

## Adding A New Scenario

Scenario files must follow the naming pattern `NN_<name>.sh` (e.g. `18_my_new_scenario.sh`), where `NN` is a zero-padded two-digit number. Place the file under `scenarios/`, declare `SCENARIO_CI=true` or `SCENARIO_CI=false`, and call `scenario_init "scenario-NN"` with the matching number so that the docker-compose project name and work directory align with the generated targets.

The Makefile auto-discovers files matching `scenarios/*.sh`, and generates `scenario-NN`, `logs-NN`, and `clean-NN` targets from the numeric prefix.

The intended flow is:

1. place the file in `scenarios/`
2. set `SCENARIO_CI=true` or `SCENARIO_CI=false`
3. `source` the shared library
4. call `scenario_init "scenario-NN"`
5. declare validators / sentries with `gen_validator` and `gen_sentry`
6. call `prepare_network`
7. compose the scenario out of lifecycle and transaction helpers

Use `gen_validator <name> --not-in-genesis` for a validator node that should exist in the generated docker-compose topology but enter the validator set later through a transaction or proposal.

See any file under `scenarios/` for examples.
