# Running Benchmarks

These instructions are mainly for running the benchmarks on an cloud instance that is intended to be thrown away, not on a dev machine.  Be careful with the install scripts locally.

This has only been tested on Ubuntu 16.04 and 18.04.  It *should* work on Ubuntu 14.04 as well.  It *may* work on Debian, but has never been tested.


## Setting up the machine

Put the files on the machine and login (all code assumes you are in this directory locally)

```
scp -r setup user@host:
ssh user@host
```

Run the install script (once per machine)

```
cd setup
chmod +x *
sudo ./INSTALL_ROOT.sh
```

## Running the tests

Run the benchmarks in a screen:

```
screen
./RUN_BENCHMARKS.sh
```

Copy them back from your local machine:

```
scp user@host:go/src/github.com/tendermint/classic/iavl/results.txt results.txt
git add results
```
