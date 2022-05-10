#!/bin/bash -i


echo "0. Changing working directory as home folder"
cd ~

echo "1. Updating apt-get"
sudo apt-get update -qq

echo "2. Installing basic essentials (might take some time)"
sudo apt-get install wget git build-essential -y -qq > /dev/null

echo "3. Installing golang v1.18"
wget -q -O - https://git.io/vQhTU | bash -s - --version 1.18 > /dev/null 2>&1


echo "4. Sourcing bashrc for golang to work"
source $HOME/.bashrc

export GOROOT=$HOME/.go
export PATH=$GOROOT/bin:$PATH
export GOPATH=/root/go
export PATH=$GOPATH/bin:$PATH

echo "5. Cloning gno from git"
git clone http://github.com/gnolang/gno.git > /dev/null 2>&1

echo "6. Changing working directory as gno"
cd ~/gno

echo "7. Building gno and tools"
make reset > /dev/null 2>&1
make all > /dev/null 2>&1

echo "8. Executing gnoland"
cd ~/gno; ./build/gnoland 