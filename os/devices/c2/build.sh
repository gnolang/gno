#!/bin/bash

# Fail on error
set -exo pipefail

# Print each command
set -o xtrace

# Get the 64 bit rpi rootfs for Odroid C2
wget -N --progress=bar:force:noscroll http://os.archlinuxarm.org/os/ArchLinuxARM-odroid-c2-latest.tar.gz

# Build the base image
docker buildx build --file Dockerfile --tag ghcr.io/faddat/sos/c2 --platform linux/arm64 --load --progress plain ..

# TAG AND PUSH
# docker push ghcr.io/faddat/sos/c2

# New rootfs extraction
# https://chromium.googlesource.com/external/github.com/docker/containerd/+/refs/tags/v0.2.0/docs/bundle.md
# create the container with a temp name so that we can export it
docker create --name tempc2 ghcr.io/faddat/sos-c2 /bin/bash

# export it into the rootfs directory
sudo rm -rf .tmp/
mkdir -p .tmp/result-rootfs
docker export tempc2 | tar -C .tmp/result-rootfs -xf -

# remove the container now that we have exported
docker rm tempc2

# EXTRACT IMAGE
# Make a temporary directory
# rm -rf .tmp | true
# mkdir .tmp

# save the image to result-rootfs.tar
# docker save --output ./.tmp/result-rootfs.tar faddat/sos-c2

# Extract the image using docker-extract
# docker run --rm --tty --volume $(pwd)/./.tmp:/root/./.tmp --workdir /root/./.tmp/.. faddat/toolbox /tools/docker-extract --root ./.tmp/result-rootfs  ./.tmp/result-rootfs.tar

# Set hostname while the image is just in the filesystem.
sudo bash -c "echo sos > ./.tmp/result-rootfs/etc/hostname"


# ===================================================================================
# IMAGE: Make a .img file and compress it.
# Uses Techniques from Disconnected Systems:
# https://disconnected.systems/blog/raspberry-pi-archlinuxarm-setup/
# ===================================================================================


# Unmount anything on the loop device
sudo umount /dev/loop0p2 || true
sudo umount /dev/loop0p1 || true


# Detach from the loop device
sudo losetup -d /dev/loop0 || true

# Unmount anything on the loop device
sudo umount /dev/loop0p2 || true
sudo umount /dev/loop0p1 || true


# Create a folder for images
rm -rf images || true
mkdir -p images

# Make the image file
fallocate -l 3G "images/c2.img"

# loop-mount the image file so it becomes a disk
export LOOP=$(sudo losetup --find --show images/c2.img)

# Zero the first 8 1MB blocks
sudo dd if=/dev/zero of=$LOOP bs=1M count=8

# partition the loop-mounted disk
sudo parted --script $LOOP mklabel msdos
sudo parted --script $LOOP mkpart primary ext4 8M 100%

# Format the disk
sudo mkfs.ext4 -O ^metadata_csum,^64bit $(echo $LOOP)p1

# might neeed sd_fusing for u-boot
cd ./.tmp/result-rootfs/boot
sudo ./sd_fusing.sh $(echo $LOOP)
cd ../../..

# tar -xpf ArchLinuxARM-odroid-c2-latest.tar.gz -C root
# cd root/boot
#sudo ./sd_fusing.sh $LOOP

# Use the toolbox to copy the rootfs into the filesystem we formatted above.
# * mount the disk's /boot and / partitions
# * use rsync to copy files into the filesystem
# make a folder so we can mount the boot partition
# soon will not use toolbox

sudo mkdir -p mnt/rootfs
sudo mount $(echo $LOOP)p1 mnt/rootfs
sudo rsync -a ./.tmp/result-rootfs/* mnt/rootfs
sudo umount mnt/rootfs



# Drop the loop mount
sudo losetup -d $LOOP

# Delete .tmp and mnt
sudo rm -rf ./.tmp
sudo rm -rf mnt