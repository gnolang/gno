# Gno

At first, there was Bitcoin, out of entropy soup of the greater All.
Then, there was Ethereum, which was created in the likeness of Bitcoin,
but made Turing complete.

Among these were Tendermint and Cosmos to engineer robust PoS and IBC.
Then came Gno upon Cosmos and there spring forth Gnoland,
simulated by the Gnomes of the Greater Resistance.

<b>This README is a placeholder, check back again for updates</b>
<b>This is NOT the same project or token as the excellent Gnosis.io project.</b>

## Language Features

 * Like interpreted Go, but more ambitious.
 * Completely deterministic, for complete accountability.
 * Transactional persistence across data realms.
 * Designed for concurrent blockchain smart contracts systems.
 
## Status

_Pinned/Sticky Update_

The best way to test the functionality of gnolang is to run the tests,
or to run the steps laid out in /examples/gno.land/r/boards/README.md.
To run a smart contract locally, copy one of the test files in
/tests/files2/zrealm_\*.go into one of your own.

You can run these tests by running:
`go test tests/*.go -v -run "TestFiles2/zrealm"`.

----------------------------------------
_Update Mar 23rd, 2022: CPU and memory allocation limitations_

Completed a simple implementation of CPU and memory allocation limitations.
The deduction of allocation units from garbage-collected allocated
and structures does not occur until the end of the transaction.

NOTE: This means memory usage will more inefficient for functions that rely on
the garbage collection for memory reclaimation.  In the future, a finer scoped
ability to account for garbage collection between functions will be
implemented.

NOTE: CPU counting and some memory allocation values are placeholders, and need
to be determined through empirical measurements.

## Resources

 * [GnoKey Client Tool](/cmd/gnokey) universal Gno client
 * [Amino](/pkgs/amino) complete with .proto generation
 * [BFT Consensus](/pkgs/bft) minimal port of Tendermint
 * [SDK](/pkgs/sdk) minimal port of Cosmos-SDK
 * [Logos Browser](/logos) future terminal browser
 * [Plan](/PLAN.md) project plan
 * [Roadmap](/ROADMAP.md) development roadmap
 * [Philosophy](/PHILOSOPHY.md) project philosophy

## Contact

If you can read this, the project is evolving (fast) every day.  Check
"github.com/gnolang/gno" and @jaekwon frequently.

The best way to reach out right now is to create an issue on github, but this
will change soon.

 * Telegram: t.me/gnoland (info on gnoland)
 * Telegram: t.me/gnolang (devs only -- invite only)
