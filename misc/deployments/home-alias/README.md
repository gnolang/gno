# Gno.land home alias overrides

Alias system allows ⁠`/r/gnoland/home` to be replaced with the content from a static local markdown file using the `--aliases` flag of gnoweb.

## Which homepage file

Package paths differ per network, so the homepage does too. Copy the matching
file as `home-override.md` in the aliased folder:

| File | Network | Differs by |
|------|---------|-----------|
| [home.mainnet.md](home.mainnet.md) | `gnoland-1` (gno.land) | `boards2/v0`, faucet label |
| [home.testnet.md](home.testnet.md) | testnets (pearl) | `boards2/v1`, testnet notice |

`boards2/v1` is what the running pearl has. #6172 renumbered `examples/` to `v0`,
so a testnet regenerated from master needs that line back at `v0`.

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
gnokey query vm/qrender -remote https://rpc.pearl.testnets.gno.land -data "gno.land/r/leon/home:" > home/home-override.md
```

- check the updated and overridden home at `http://127.0.0.1/`.

## Kubernetes version

- Create a Kubernetes Job resource
- The Job runs the script but using a `kubectl` command as argument rather than `docker`.
