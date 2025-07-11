# Gno.land home alias

Serve homepage from local .md using `--alias=/` and dynamic sync from `r/gnoland/home`,
allowing overrides.

## Prerequisite

In any environment used:

- the `gnoalias` script MUST share a `volume` with the running `gnoweb` service
- `gnoweb` should be run using the `--aliases=/=static:<path to home.md>` argument

## Docker version

See [docker-compose.yml](docker-compose.yml)

### Testing locally

- Spin `docker compose`

```sh
docker compose --profile dev up -d
```

- Then place an override into the `./home/` folder:

```sh
gnokey query vm/qrender -remote https://rpc.test6.testnets.gno.land -data "gno.land/r/leon/home:" > home/home-override.md
```

- check the updated adn overriddend home at `http://127.0.0.1/`.

## Kubernetes version

- Create a CronJob
- The cronjob runs the script but using a `kubectl` command as argument.
