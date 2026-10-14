# Gno.land home alias overrides

Alias system allows ⁠`/r/gnoland/home` to be replaced with the content from a static local markdown file using the `--aliases` flag of gnoweb.

## Which homepage file

Mainnet and the testnets do not list the same links. Copy the matching file as
`home-override.md` in the aliased folder:

| File | Network | Differs by |
|------|---------|-----------|
| [home.mainnet.md](home.mainnet.md) | `gnoland-1` (gno.land) | "Live on mainnet" projects, faucet labeled testnet-only |
| [home.testnet.md](home.testnet.md) | testnets (onyx) | Faucet as the first "use" link, "On this testnet" realms, testnet notice |

Staging is not concerned: it serves `r/gnoland/home` from the chain, with no alias.

The purpose of this solution is either:

- allowing overriding the current home alias, using a file called `home-override.md` placed in the same folder of currently aliased home file
- adding extra blocks to the current home, by placing them in a file called `extra-blocks.md`. The latter file should be placed in the same folder of currently aliased home file too.

## Prerequisites

In any environment used:

- the `gnoalias` script MUST share a `volume` with the running `gnoweb` service
- `gnoweb` should be run using the `--aliases=/=static:<path to home.md>` argument
- (opt.) `ALIAS_HOME_FOLDER` env variable specifies the folder of currently aliased home file. If not present, fallbacks to the `./home` folder, `/gnoroot/home` in a gnoweb container

## Docker version

See [docker-compose.yml](docker-compose.yml)

### Testing locally

- Spin `docker compose`

```sh
docker compose --profile dev up -d
```

- Then place an override into the `./home/` folder:

```sh
gnokey query vm/qrender -remote https://rpc.onyx.testnets.gno.land -data "gno.land/r/gnoland/blog:" > home/home-override.md
```

- check the updated and overridden home at `http://127.0.0.1/`.

## Kubernetes version

- Create a Kubernetes Job resource
- The Job runs the script but using a `kubectl` command as argument rather than `docker`.
