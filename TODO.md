# gnolang
  * Finish passing gnolang files tests (DONE).
  * Dry the code with select refactors.
  * Implement form of channel send/recv.
  * Complete float32/float64 implementation (as struct).
  * Check parsed AST for compile-time errors.
    - unused names,
    - XXX
  * Ensure determinism regarding 32 vs 64 bit for int/uint.
  * Ensure non-realm paths cannot mutate state.
  * Ensure native (autonative) func call types in checkType().
  * Finish implementation of allocator for native calls etc.

# /pkgs
  * Replace testify with gnolang/gno/pkgs/testify
  * `command`: make utility that parses flags using `BurntSushi/toml` or some vetted toml lib, but nothing else (besides amino json)
  * Move most of classic/sdk/ as packages in gno/pkgs/
  * Move tendermint consensus modules as packages in gno/pkgs/tendermint
  * Embedded AminoMarshaler fields should not cause the parent to become AminoMarshaler.

# other
  * Replace spf13 with gnolang/testify fork of jaekwon/testify

----------------------------------------

* Limit CPU and memory usage.
 -> memory usage: 
 -> clear cache upon beginnewblock (DONE).
 -> limit allocation per tx. (DONE)
 -> limit allocation from store. (DONE)
 -> prevent mutation of state on non-realm packages. (DONE)
 -> limit cache size on store. (DONE)

* Ensure code is proper.
 -> run through compiler for now?
 -> ...

* Realm PkgPath URL fixes.
 -> /r/boards/board.go should show the code.

* Validator set changes.
 -> re staking:
