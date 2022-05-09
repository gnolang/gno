# Basic 
## Tags

|  Tag   |   Update   |    Type    |              Description               |
|:------:|:----------:|:----------:|:---------------------------------------|
| latest | **daily**  | minimal    | Minimal Arch Linux with pacman support |
| devel  | **daily**  | base-devel | Arch Linux with base-devel installed   |

### Layer structure

The image is generated from a freshly built pacman rootfs. Pacman has configured
to delete man pages and clean the package cache after installation to keep
images small.

## Issues and improvements

If you want to contribute, get to the [issues-section of this repository](https://github.com/faddat/archlinux-docker/issues).

## Common hurdles

### Setting the timezone

Simply add the `TZ` environment-variable and define it with a valid timezone-value.

```
docker run -e TZ=America/New_York lopsided/archlinux
```

## Building it yourself

### Prerequisites

- Docker with experimental mode on (required for squash)
- sudo or root is neccessary to setup binfmt for Qemu user mode emulation

### Building

- Prepare binfmt use with Qemu user mode using `sudo ./prepare-qemu`
- Run `BUILD_ARCH=<arch> ./build` to build
  - Use `BUILD_ARCH=amd64` for x86_64
  - Use `BUILD_ARCH=arm64v8` for ARMv8 Aarch64

If you want to push the images, run `./push`. *But be aware you have no push access to the repos! Edit the scripts to push to custom Docker Hub locations!*

### Building from scratch

Since the image depends on itself, the question which arises is how this all
started. The initial containers have been created using the tarballs provided by
the Arch Linux ARM project. I used the following steps to bootstrap for each
architecture:

```
gzip -d ArchLinuxARM-armv7-latest.tar.gz
docker import ArchLinuxARM-armv7-latest.tar lopsided/archlinux-arm32v7:latest
```

## Credits

Ideas have been taken from already existing Docker files for Arch Linux.
However, this repository takes a slightly different approach to create images.

- https://github.com/agners/archlinux-docker
  - Limited architectures
  - Duplication of Dockerfiles
  - Only built weekly
  - No image with base-devel preinstalled
- https://github.com/archlinux/archlinux-docker
  - Focus on Arch Linux for x86
  - Uses docker run in priviledged mode to build images
- https://github.com/lopsided98/archlinux
  - Uses prebuilt tarballs which contain packages not required in containers
- https://github.com/faddat/archlinux
  - Automated build for github container registry
  - dropped armv7 and other 32 bit modes
  - built hourly because arch is rolling release
