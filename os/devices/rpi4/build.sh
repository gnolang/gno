#!/bin/bash
# =======================================================================
# SOS image builder
# =======================================================================


# This process uses tools and a design pattern first developed by the pikvm team for their pi-builder and os tools.
# the biggest differences between this process and theirs are:
# * we use docker buildx so we don't need to deal with qemu directly.
# * we are not offering as many choices to users and are designing around automation.
# Later we can make this work for more devices and platforms with nearly the same technique.
# Reasonable build targets include: https://archlinuxarm.org/platforms/armv8
# For example, the Odroid-N2 is the same software-wise as our Router!

# Fail on error
set -exo pipefail

# Print each command
set -o xtrace

# Get rootfs
wget -N --progress=bar:force:noscroll http://os.archlinuxarm.org/os/ArchLinuxARM-rpi-aarch64-latest.tar.gz

# BUILD IMAGE
docker buildx build --tag sos-lite --file Dockerfile --platform linux/arm64 --progress plain --load ..

# TAG AND PUSH
docker tag sos-lite ghcr.io/notional-labs/sos
docker push ghcr.io/notional-labs/sos

# PREPARE TOOLBOX
# docker buildx build --rm --tag toolbox --file toolbox/Dockerfile --load  --progress plain toolbox

# EXTRACT IMAGE
# Make a temporary directory
rm -rf .tmp | true
mkdir .tmp

# remove anything in the way of extraction
docker run --rm --tty --volume $(pwd)/./.tmp:/root/./.tmp --workdir /root/./.tmp/.. faddat/toolbox rm -rf ./.tmp/result-rootfs

# save the image to result-rootfs.tar
docker save --output ./.tmp/result-rootfs.tar sos-lite

# Extract the image using docker-extract
docker run --rm --tty --volume $(pwd)/./.tmp:/root/./.tmp --workdir /root/./.tmp/.. faddat/toolbox /tools/docker-extract --root ./.tmp/result-rootfs  ./.tmp/result-rootfs.tar

# Delete tarball to save space
rm ./.tmp/result-rootfs.tar

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
fallocate -l 4G "images/sos-lite.img"

# loop-mount the image file so it becomes a disk
export LOOP=$(sudo losetup --find --show images/sos-lite.img)

# partition the loop-mounted disk
sudo parted --script $LOOP mklabel msdos
sudo parted --script $LOOP mkpart primary fat32 0% 200M
sudo parted --script $LOOP mkpart primary ext4 200M 100%

# format the newly partitioned loop-mounted disk
sudo mkfs.vfat -F32 $(echo $LOOP)p1
sudo mkfs.ext4 -F $(echo $LOOP)p2

# Use the toolbox to copy the rootfs into the filesystem we formatted above.
# * mount the disk's /boot and / partitions
# * use rsync to copy files into the filesystem
# make a folder so we can mount the boot partition
# soon will not use toolbox

sudo mkdir -p mnt/boot mnt/rootfs
sudo mount $(echo $LOOP)p1 mnt/boot
sudo mount $(echo $LOOP)p2 mnt/rootfs
sudo rsync -a ./.tmp/result-rootfs/boot/* mnt/boot
sudo rsync -a ./.tmp/result-rootfs/* mnt/rootfs --exclude boot
sudo mkdir mnt/rootfs/boot
sudo umount mnt/boot mnt/rootfs

# Tell pi where its memory card is:  This is needed only with the mainline linux kernel provied by linux-aarch64
# sed -i 's/mmcblk0/mmcblk1/g' ./.tmp/result-rootfs/etc/fstab

# Drop the loop mount
sudo losetup -d $LOOP

# Delete .tmp and mnt
sudo rm -rf ./.tmp
sudo rm -rf mnt
