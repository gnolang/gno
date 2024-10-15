# The portal loop :infinity: 

## What is it?

It's a Gnoland node that aim to run with always the latest version of gno and never loose transactions history.

For more information, see issue on github [gnolang/gno#1239](https://github.com/gnolang/gno/issues/1239)


## How to use

Start the loop with:

```sh
$ docker compose up -d

# or using the Makefile

$ make
```

The [`portalloopd`](./cmd/portalloopd) binary is starting inside of the docker container `portalloopd`

This script is doing:

- Setup the current portal-loop in read only mode
- Pull the latest version of [ghcr.io/gnolang/gno]()
- Backup the txs using [gnolang/tx-archive](https://github.com/gnolang/tx-archive)
- Start a new docker container with the backups files
- Changing the proxy (traefik) to redirect to the new portal loop
- Unlock read only mode
- Stop the previous loop

### Makefile helper

You can find a [Makefile](./Makefile) to help you interact with the portal loop

- Force switch of the portal loop with a new version

```bash
make portalloopd.switch
```
