# bytebeat realm

The intention behind this realm is to provide a proof-of-concept for generating audio through smart contracts (for possible future use with web3 audio nfts or similar).

There are several packages added to accomplish this:

- `examples/gno.land/p/demo/audio/riff`: generates `riff` headers for `.wav` files
- `examples/gno.land/p/demo/audio/wav`: provides `Writer` interface for writing `.wav` files
- `examples/gno.land/p/demo/audio/biquad`: implements basic 2nd-order digital biquad filter for doing low-pass and high-pass filtering 
- `examples/gno.land/p/demo/audio/bytebeat`: core library for generating 16-bit byte-beat audio that is post-processed with a DC offset filter and a high pass filter to eliminate aliasing


## Caveats

Currently, there is a limitation to the number of CPU cycles available to a contract and the realm is currently limited to generating only 2 seconds of audio at 8 kHz.

## Usage

To use, add the packages for the audio library and then add the realm:

```
cat ~/password | gnokey maketx addpkg --pkgpath "gno.land/p/demo/math_eval/int32" --pkgdir "examples/gno.land/p/demo/math_eval/int32" --deposit 100000000ugnot --gas-fee 1000000ugnot --gas-wanted 2000000 --broadcast --chainid dev --remote localhost:26657 --insecure-password-stdin=true yourkey
cat ~/password | gnokey maketx addpkg --pkgpath "gno.land/p/demo/audio/biquad" --pkgdir "examples/gno.land/p/demo/audio/biquad" --deposit 100000000ugnot --gas-fee 1000000ugnot --gas-wanted 2000000 --broadcast --chainid dev --remote localhost:26657 --insecure-password-stdin=true yourkey
cat ~/password | gnokey maketx addpkg --pkgpath "gno.land/p/demo/audio/riff" --pkgdir "examples/gno.land/p/demo/audio/riff" --deposit 100000000ugnot --gas-fee 1000000ugnot --gas-wanted 2000000 --broadcast --chainid dev --remote localhost:26657 --insecure-password-stdin=true yourkey
cat ~/password | gnokey maketx addpkg --pkgpath "gno.land/p/demo/audio/wav" --pkgdir "examples/gno.land/p/demo/audio/wav" --deposit 100000000ugnot --gas-fee 1000000ugnot --gas-wanted 2000000 --broadcast --chainid dev --remote localhost:26657 --insecure-password-stdin=true yourkey
cat ~/password | gnokey maketx addpkg --pkgpath "gno.land/p/demo/audio/bytebeat" --pkgdir "examples/gno.land/p/demo/audio/bytebeat" --deposit 100000000ugnot --gas-fee 1000000ugnot --gas-wanted 8000000 --broadcast --chainid dev --remote localhost:26657 --insecure-password-stdin=true  yourkey
cat ~/password | gnokey maketx addpkg --pkgpath "gno.land/r/demo/bytebeat" --pkgdir "examples/gno.land/r/demo/bytebeat" --deposit 100000000ugnot --gas-fee 1000000ugnot --gas-wanted 4000000 --broadcast --chainid dev --remote localhost:26657 --insecure-password-stdin=true  yourkey

```

Note that adding the `gno.land/p/demo/audio/bytebeat` package requires adding more to `--gas-wanted` (at least 8000000).
