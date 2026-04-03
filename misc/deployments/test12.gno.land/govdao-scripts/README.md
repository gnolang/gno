# govDAO Scripts

Scripts for govDAO members to manage the test12 chain. If you're a validator operator (valoper), you can ignore this directory.

Defaults (`GNOKEY_NAME`, `CHAIN_ID`, `REMOTE`, `GAS_WANTED`, `GAS_FEE`) are defined in `common`. Override any value inline:

```bash
GNOKEY_NAME=mykey ./add-validator.sh ...
```

```bash
./add-validator-from-valopers.sh ADDR       # add a validator registered at r/gnops/valopers
./add-validator.sh PUBKEY [POWER]           # add a validator (address derived from pubkey)
./rm-validator.sh ADDR                      # remove a validator
./extend-govdao-t1.sh                       # add 6 T1 members to govDAO (one-time bootstrap)
./unrestrict-account.sh ADDR [ADDR...]      # allow address(es) to transfer ugnot
./enable-govdao-namespaces.sh               # enable PA + custom namespace support (one-time)
./add-namespace.sh NAMESPACE ADDR           # register a custom namespace in r/sys/names/v2
./rm-namespace.sh NAMESPACE                 # remove a custom namespace from r/sys/names/v2
./set-valoper-minfee.sh AMOUNT_UGNOT        # update the valoper registration minimum fee
```
