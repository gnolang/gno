# bytebeat realm

The intention behind this realm is to provide a proof-of-concept for generating audio through smart contracts (for possible future use with web3 audio nfts or similar).

There are several packages added to accomplish this:

- `examples/gno.land/p/demo/audio/riff`: generates `riff` headers for `.wav` files
- `examples/gno.land/p/demo/audio/wav`: provides `Writer` interface for writing `.wav` files
- `examples/gno.land/p/demo/audio/biquad`: implements basic 2nd-order digital biquad filter for doing low-pass and high-pass filtering 
- `examples/gno.land/p/demo/audio/bytebeat`: core library for generating 16-bit byte-beat audio that is post-processed with a DC offset filter and a high pass filter to eliminate aliasing

This realm demonstrates the basic usage of these packages. In this embodiment, the `init()` function can be used to generate the audio which is stored within the chain:

```go
var data string

func init() {
	data = bytebeat.ByteBeat(1, 8000, func(t int) int {
		return (t>>10^t>>11)%5*((t>>14&3^t>>15&1)+1)*t%99 + ((3 + (t >> 14 & 3) - (t >> 16 & 1)) / 3 * t % 99 & 64)
	})
}
```

The callback for `bytebeat.ByteBeat` allows you to use any set of bytebeat operations to create audio.

## Caveats

Currently, there is a limitation to the number of CPU cycles available to a contract and the realm is currently limited to generating only 2 seconds of audio at 8 kHz.

## Usage

To use, add the packages for the audio library and then add the realm:

```
$ gnokey maketx addpkg --pkgpath "gno.land/p/demo/audio/biquad" --pkgdir "examples/gno.land/p/demo/audio/biquad" --deposit 100000000ugnot --gas-fee 1000000ugnot --gas-wanted 2000000 --broadcast --chainid dev --remote localhost:26657 <yourkey>
$ gnokey maketx addpkg --pkgpath "gno.land/p/demo/audio/riff" --pkgdir "examples/gno.land/p/demo/audio/riff" --deposit 100000000ugnot --gas-fee 1000000ugnot --gas-wanted 2000000 --broadcast --chainid dev --remote localhost:26657 <yourkey>
$ gnokey maketx addpkg --pkgpath "gno.land/p/demo/audio/wav" --pkgdir "examples/gno.land/p/demo/audio/wav" --deposit 100000000ugnot --gas-fee 1000000ugnot --gas-wanted 2000000 --broadcast --chainid dev --remote localhost:26657 <yourkey>
$ gnokey maketx addpkg --pkgpath "gno.land/p/demo/audio/bytebeat" --pkgdir "examples/gno.land/p/demo/audio/bytebeat" --deposit 100000000ugnot --gas-fee 1000000ugnot --gas-wanted 8000000 --broadcast --chainid dev --remote localhost:26657  <yourkey>
gnokey maketx addpkg --pkgpath "gno.land/r/demo/bytebeat" --pkgdir "examples/gno.land/r/demo/bytebeat" --deposit 100000000ugnot --gas-fee 1000000ugnot --gas-wanted 3000000 --broadcast --chainid dev --remote localhost:26657  <yourkey>
```

Note that adding the `gno.land/p/demo/audio/bytebeat` requires adding more to `--gas-wanted` (at least 8000000).
