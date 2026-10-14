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

## Latest blog posts

A line `<!-- latest-blogposts -->` in `home-override.md` is replaced by the three
most recent posts of `r/gnoland/blog`, refreshed on every round of `gnoalias.sh`
(hourly). The script only rewrites `home.md`: gnoweb re-reads a static alias
file when its modification time changes, so nothing restarts. If the query
fails, the current page stays as it is. The posts come from mainnet; on a testnet, set
`BLOG_LINK_BASE=https://gno.land` (see [docker-compose.yml](docker-compose.yml))
so their links lead there.

## Prerequisites

In any environment used:

- the `gnoalias` script MUST share a `volume` with the running `gnoweb` service: the folder, not the `home.md` file alone, since the script renames a new file over it
- `gnoweb` should be run using the `--aliases=/=static:<path to home.md>` argument
- the script's first argument is the folder of the aliased home file (default `./home`; `/gnoroot/home` in the compose setup)

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
- The Job runs the script. A reload command (such as a `kubectl rollout restart`)
  is only needed with a gnoweb that does not re-read static files yet.
