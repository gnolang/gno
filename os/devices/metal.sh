#!/bin/bash

# This script handles building on bare metal servers instead of in github actions

wget -O rpi/Manjaro-ARM-aarch64-latest.tar.gz -N --progress=bar:force:noscroll https://osdn.net/projects/manjaro-arm/storage/.rootfs/Manjaro-ARM-aarch64-latest.tar.gz
cp rpi/Manjaro-ARM-aarch64-latest.tar.gz c2/Manjaro-ARM-aarch64-latest.tar.gz
cd c2
docker build --tag sos/c2 --file Dockerfile ..
cd ../rpi
docker build --tag sos/rpi --file Dockerfile ..

