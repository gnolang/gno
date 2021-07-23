# Gno-Lang
  * Finish passing gnolang files tests (public invited).
  * Dry the code with select refactors.
  * Implement form of channel send/recv.
  * Complete float32/float64 implementation (as struct).

# pkgs
  * replace testify with gnolang/gno/pkgs/testify
  * `command`: make utility that parses flags using `BurntSushi/toml` or some vetted toml lib, but nothing else (besides amino json)
  * move most of classic/sdk/ as packages in gno/pkgs/
  * move tendermint consensus modules as packages in gno/pkgs/tendermint
