# gnokey

`gnokey` is a tool for managing https://gno.land accounts and interact with instances.

## Install `gnokey`

    $> git clone git@github.com:gnolang/gno.git
    $> cd ./gno
    $> make install_gnokey

Also, see the [quickstart guide](../../../docs/users/interact-with-gnokey.md).

## Manual Entropy

Use `--entropy` for manual entropy instead of computer PRNG:

```bash
gnokey add mykey --entropy
```

Your input is SHA-256 hashed for deterministic key generation. Same entropy = same key.
