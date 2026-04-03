# Valset scripts

Scripts for managing valoper registrations and testing the full valset lifecycle on test12. Defaults (`GNOKEY_NAME`, `CHAIN_ID`, `REMOTE`, etc.) are inherited from `govdao-scripts/common`. Override any value inline:

```bash
GNOKEY_NAME=mykey REMOTE=http://localhost:26657 ./register-valoper.sh ...
```

## Valoper scripts

```
./register-valoper.sh PUBKEY MONIKER DESCRIPTION SERVER_TYPE
```

Register a new valoper in `r/gnops/valopers`. The address is derived from the pubkey. Sends the registration fee (`VALOPER_REGISTRATION_FEE`, default 20 GNOT) from `GNOKEY_NAME`'s balance. `SERVER_TYPE` must be `cloud`, `on-prem`, or `data-center`.

```
./update-valoper-moniker.sh ADDR NEW_MONIKER
./update-valoper-description.sh ADDR NEW_DESCRIPTION
./update-valoper-servertype.sh ADDR SERVER_TYPE
./update-valoper-keeprunning.sh ADDR true|false
```

Update fields of an existing valoper profile. The caller must be the original registrant or on the valoper's auth list. Setting `KeepRunning=false` signals intent to leave the validator set; a subsequent govDAO proposal via `add-validator-from-valopers.sh` will then remove the validator.

```
./add-auth-member.sh VALOPER_ADDR MEMBER_ADDR
./rm-auth-member.sh VALOPER_ADDR MEMBER_ADDR
```

Add or remove a member from a valoper's auth list. Only the original registrant (owner) can modify the list.

## Test suite

```bash
./test-valset.sh VAL1_PUBKEY VAL2_PUBKEY
```

Runs all test groups sequentially against the live chain. Both pubkeys must belong to validators **not** in the initial valset. Addresses are derived from the pubkeys automatically. Prompts for the `GNOKEY_NAME` password once and reuses it throughout.

### Group 1 — Direct govDAO proposals

| #   | Scenario                                              | Expected                           |
| --- | ----------------------------------------------------- | ---------------------------------- |
| 1.1 | Add val1 (not in valset, not in valopers)             | success — val1 in valset           |
| 1.2 | Add val1 again (already in valset, same pubkey+power) | success — idempotent update        |
| 1.3 | Remove val1 (in valset, not in valopers)              | success — val1 removed             |
| 1.4 | Remove val1 again (not in valset)                     | failure — `removeValidator` panics |
| 1.5 | Add val2 with power=2                                 | success — val2 in valset           |
| 1.6 | Remove val2                                           | success — val2 removed             |

### Group 2 — Valopers realm operations

| #    | Scenario                                      | Expected                                   |
| ---- | --------------------------------------------- | ------------------------------------------ |
| 2.1  | Register val1                                 | success — val1 in valopers                 |
| 2.2  | Re-register val1 (already registered)         | failure — `ErrValoperExists`               |
| 2.3  | Register with single-char moniker             | failure — moniker regex mismatch           |
| 2.4  | Register with invalid server type             | failure — `ErrInvalidServerType`           |
| 2.5  | Register with non-bech32 pubkey               | failure — address derivation fails locally |
| 2.6  | Register with 1 ugnot fee (below 20 GNOT min) | failure — insufficient fee                 |
| 2.7  | Register val2 with valid data                 | success — val2 in valopers                 |
| 2.8  | Update val1 moniker                           | success                                    |
| 2.9  | Update val1 description                       | success                                    |
| 2.10 | Update val1 server type to `data-center`      | success                                    |
| 2.11 | Update val1 server type to `invalid`          | failure — `ErrInvalidServerType`           |
| 2.12 | Set val1 `KeepRunning=false`                  | success — flag updated                     |
| 2.13 | Set val1 `KeepRunning=true`                   | success — flag restored                    |
| 2.14 | Add then remove val2 from val1's auth list    | success both steps                         |

### Group 3 — GovDAO proposals via valopers

| #   | Scenario                                                        | Expected                        |
| --- | --------------------------------------------------------------- | ------------------------------- |
| 3.1 | Add val1 from valopers (registered, not in valset)              | success — val1 in valset        |
| 3.2 | Re-add val1 from valopers (in valset, same pubkey+power)        | failure — `ErrSameValues`       |
| 3.3 | Add unregistered address via valopers proposal                  | failure — `ErrValoperMissing`   |
| 3.4 | Valopers proposal for val2 (`KeepRunning=false`, not in valset) | failure — `ErrValidatorMissing` |
| 3.5 | Add val2 from valopers (`KeepRunning=true`, not in valset)      | success — val2 in valset        |
| 3.6 | Remove val1 via valopers (`KeepRunning=false`, in valset)       | success — val1 removed          |
| 3.7 | Valopers proposal for val1 (`KeepRunning=false`, not in valset) | failure — `ErrValidatorMissing` |
| 3.8 | Add val1 via direct proposal, then re-add via valopers          | failure — `ErrSameValues`       |

### Group 4 — Edge cases

| #   | Scenario                                                                        | Expected                                               |
| --- | ------------------------------------------------------------------------------- | ------------------------------------------------------ |
| 4.1 | Direct remove of val1 (in valset and in valopers)                               | success — removed from valset, still in valopers       |
| 4.2 | Add val1 with power=2 via direct proposal, then correct to power=1 via valopers | success — valopers proposal accepted (different power) |
| 4.3 | Final cleanup — remove both test validators                                     | success — clean state restored                         |
