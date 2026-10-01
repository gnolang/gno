# The portal loop :infinity:

## What is it?

It's a Gnoland node that aim to run with always the latest version of gno and never loose transactions history.

For more information, see issue on github [gnolang/gno#1239](https://github.com/gnolang/gno/issues/1239)

## How to use

Start the loop with:

```sh
docker compose up -d

# or using the Makefile
make docker.start
```

The `portalloopd` binary is starting inside of the docker container `portalloopd`

This script is doing:

- Setup the current portal-loop in read only mode
- Pull the latest version of [ghcr.io/gnolang/gno](https://ghcr.io/gnolang/gno)
- Backup the txs using [contribs/tx-archive](https://github.com/gnolang/gno/tree/master/contribs/tx-archive)
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

### Running in production

- Create an `.env` file adding all the entries from `.env.example`
- Setup the DNS names present in the `docker-compose.yml` file
- run using `make all`

### Pulling in Portal Loop state `from tx-exports`

To pull Portal Loop state from tx-exports, run the following make directive:

```bash
make pull-exports
```

This will run the following steps:

- stop any running portal loop containers -> Portal Loop will be down
- clone the `gnolang/tx-exports` repository and prepare the backup txs sheets located there as the genesis transactions
  for Portal Loop
- start the portal loop containers -> Portal Loop will start back up again