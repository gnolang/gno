# Name Registry Realm

This folder contains a `gengno.go` file to generate a _Gno_ file from Handshake
protocol's names [lookup file], which contains a curated list of reserved names
extracted from custom names, TLDs and Alexa's top 100k domain names.

The generated Gno file can be used to populate the list of reserved names
stored in `gno.land/r/sys/nameregistry` realm.

Reserved names are populated though a **GovDAO** proposal.

Lookup file and other denormalized files used to generate the final JSON [lookup
file] can be found at [github.com/handshake-org/hs-names](https://github.com/handshake-org/hs-names).

For more context on Handshake's [lookup file] refer to [PR #819](https://github.com/handshake-org/hsd/pull/819).

[lookup file]: https://github.com/handshake-org/hs-names-2023/blob/3482d12e9c680030f1cec729f5e3a7aa454d0f15/build/updated/lockup.json
