# The portal loop :infinity: 

## What is it ?

It's a Gnoland node that aim to run with always the latest version of gno and never loose transactions history.

For more information, see issue on github [gnolang/gno#1239](https://github.com/gnolang/gno/issues/1239)


## How to use

Start the loop with:

``` sh
$ docker compose up -d

# or using the Makefile

$ make
```

The `snapshotter` container will exec the script [switch.sh](./scripts/switch.sh) every day at 10am (defined in the docker image).

This script is doing:

- Pull the new docker image [ghcr.io/gnolang/gno]
- Backup the txs using [gnolang/tx-archive](https://github.com/gnolang/tx-archive)
- Start a new docker container with the backups files
- Changing the proxy (traefik) to redirect to the new portal loop
- Stop the previous loop

### Makefile helper

You can find a [Makefile](./Makefile) to help you interact with the portal loop

- Force switch of the portal loop with a new version

```bash
make snapshotter.switch-now
```
