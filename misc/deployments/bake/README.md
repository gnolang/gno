# Docker Bake

Building Docker images by _baking_ them.

* Bake Gno basic images from `${project_root_folder}`

```sh
docker buildx bake --allow fs=\* --file misc/deployments/bake/docker-bake.hcl --set \*.context=. --set \*.dockerfile=Dockerfile
```

* Bake the Gno family images from this folder

```sh
docker buildx bake --allow fs=\* --file docker-bake.hcl --set \*.dockerfile=Dockerfile _gno
```

* Bake a single target for a single platform and output into a local folder `build/`, from `${project_root_folder}`

```sh
docker buildx bake --allow fs=\* --file misc/deployments/bake/docker-bake.hcl --set \*.context=. --set \*.dockerfile=Dockerfile --set common.platform=linux/arm64 --set common.output.type=local --set common.output.dest=build/ gnofaucet

* Bake a single target and push to registry, from `${project_root_folder}`

```sh
docker buildx bake --allow fs=\* --file misc/deployments/bake/docker-bake.hcl --set \*.context=. --set \*.dockerfile=Dockerfile --push gnoland
```

## See Also

* [docker buildx bake](https://docs.docker.com/reference/cli/docker/buildx/bake/)
* [Bake file reference](https://docs.docker.com/build/bake/reference/#use-environment-variable-as-default)
* [docker/bake-action: GitHub Action to use Docker Buildx Bake as a high-level build command](https://github.com/docker/bake-action)
