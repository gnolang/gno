# Gno.land home alias overrides

Alias system allows ⁠`/r/gnoland/home` to be replaced with the content from a static local markdown file using the `--aliases` flag of gnoweb.

## Which homepage file

Mainnet and the testnets do not list the same links. Copy the matching file as
`home-override.md` in the aliased folder:

| File | Network | Differs by |
|------|---------|-----------|
| [home.mainnet.md](home.mainnet.md) | `gnoland-1` (gno.land) | "Live on mainnet" projects, faucet labeled testnet-only |
| [home.testnet.md](home.testnet.md) | testnets (onyx) | A page for builders: onboarding by case, live activity, a deploy command |

Staging is not concerned: it serves `r/gnoland/home` from the chain, with no alias.

The purpose of this solution is either:

- allowing overriding the current home alias, using a file called `home-override.md` placed in the same folder of currently aliased home file
- adding extra blocks to the current home, by placing them in a file called `extra-blocks.md`. The latter file should be placed in the same folder of currently aliased home file too.

## What's new block

[whatsnew.sh](whatsnew.sh) fills a block of a home file in place, between its
`<!-- whats-new:start -->` and `<!-- whats-new:end -->` lines. It is a single
file that needs `bash`, `curl` and `jq`. Without it, the two lines render as
nothing.

```sh
./whatsnew.sh --dry-run home.md   # print the block, leave the page alone
./whatsnew.sh home.md             # update the page
./whatsnew.sh --force home.md     # same, ignoring the refresh intervals
```

The block has the three columns of the rest of the page: the latest X posts, the
latest blog posts, videos or releases under 60 days old with the last merged
pull requests, and the most used realms with the most imported packages. With
`--builders`, for a testnet, it holds the week's totals, then the latest
deployments, the most used realms and the last merged pull requests:

```sh
INDEXER_URL=https://indexer.example/graphql/query USED_NAMESPACES= \
  ./whatsnew.sh --builders home.md
```

Text from X comes from X itself (FxTwitter only lists the post IDs), all of it
is escaped, and the only links are ones the script builds. Realms are ranked by
distinct addresses, and only those of the `USED_NAMESPACES` namespaces can appear.

| Variable | Default | Purpose |
|----------|---------|---------|
| `RPC_URL` | `https://rpc.gno.land:443` | Package list of "Top packages" |
| `BLOG_RPC_URL` | `https://rpc.gno.land:443` | Where the blog is read |
| `LINK_BASE` | empty | Prefix of blog links, for a page served elsewhere than mainnet |
| `INDEXER_URL` | empty | tx-indexer GraphQL endpoint; without it, what needs an indexer stays out |
| `X_ACCOUNT` | `_gnoland` | The account shown |
| `YOUTUBE_CHANNEL` | `UCHlHaup7HEDQFQrb1goxAkg` | The channel shown |
| `VIDEO_SKIP` | `Jae Kwon` | Regular expression of video titles to leave out; empty keeps all |
| `DEPLOYED_SKIP` | `/smoke/` | Regular expression of paths `--builders` leaves out of its deployments and their total; empty keeps all |
| `USED_NAMESPACES` | `gnoland gnoswap akkadia` | Namespaces the ranking may show; empty shows all, for a testnet |
| `NETWORK` | `onyx` | The testnet `--builders` names and links to |
| `ON_CHANGE` | empty | Command run after the page changed |
| `STATE_DIR` | `.whatsnew` beside the page | Last good output of each source |
| `FEED_TTL` | `30` | Minutes between reads of X, the blog, YouTube and GitHub |
| `USED_TTL` | `1440` | Minutes between indexer scans, about 10 queries each, of "Top realms" and, with `--builders`, of the deployments |
| `IMPORTED_TTL` | `10080` | Minutes between "Top packages" scans, one `vm/qdoc` per package |

Run it often: a source is only read once its interval has passed, whether or not
its last read worked. One that fails keeps its previous content, except the
rankings of the week, which go after seven days. The page is replaced atomically
and keeps its mode, only when its content changes, and never by an empty block.
Point the script at the file itself, not at a symlink. It exits 0 once it has
run, and 2 if the page does not hold each marker line once, in order; a failed
`ON_CHANGE` is logged and tried again on the next run.

The block only shows if gnoweb re-reads a static page when it changes (#6327).
Without that, `ON_CHANGE` can restart gnoweb on a single host. Do not do that on
a pool behind a load balancer: every host would drop requests for the same post.

On a pool, ship the script and its units in the image, keep `STATE_DIR` out of
any folder a sync replaces, and enable the timer so that it starts once the page is in place. With
`gnoalias.sh`, run it on `home-override.md` instead, with `ON_CHANGE` empty.

```ini
# whatsnew.service
[Unit]
Wants=network-online.target
After=network-online.target

[Service]
Type=oneshot
Environment=INDEXER_URL=https://indexer.example/graphql/query
Environment=STATE_DIR=/var/lib/whatsnew
ExecStart=/usr/local/bin/whatsnew.sh /path/to/home.md

# whatsnew.timer
[Timer]
OnBootSec=5min
OnUnitActiveSec=10min
RandomizedDelaySec=2min

[Install]
WantedBy=timers.target
```

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
