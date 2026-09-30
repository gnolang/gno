package main

import (
	"fmt"
	"strings"
)

// The live network this page points at, so a visitor who mistook this node for
// gno.land can find the real one.
const (
	mainnetURL     = "https://gno.land"
	mainnetChainID = "gnoland-1"
)

// addpkgGasWanted is the -gas-wanted the page suggests for a deployment.
//
// Measured, not copied: a five-line realm costs 4_282_275 gas, so the
// 2_000_000 used by the `maketx call` examples in docs/ runs out. This is the
// figure the addpkg walkthroughs use (quickstart.md, getting-started.md) and it
// clears that cost with room for a larger package.
const addpkgGasWanted = 20_000_000

// stagingHomeMarkdown renders the landing page gnodev serves at "/" in staging
// mode, in place of the aliased /r/gnoland/home.
//
// Without it a preview chain renders the real gno.land homepage, which is the
// one page guaranteed to make it look like the live network. This page says
// what the node is, how it differs from gno.land in ways that change whether
// your code works, and how to use it.
//
// rpcRemote is the operator-supplied public RPC (-web-help-remote). When it is
// unset the RPC line is left out rather than printing the node's own listen
// address, which is meaningless to anyone but the operator, and the deploy
// command carries a visible <this-chain-rpc> placeholder instead: dropping the
// flag would make gnokey fall back to its own localhost default.
func stagingHomeMarkdown(chainID, rpcRemote string) string {
	var b strings.Builder

	b.WriteString("# A gnodev preview chain\n\n")
	b.WriteString("This is not a gno.land network. It is a single node someone started with " +
		"`gnodev`, running the same GnoVM, for trying code out before it reaches a real chain.\n\n")

	b.WriteString(fmt.Sprintf("- **Chain ID**: `%s`\n", chainID))
	if rpcRemote != "" {
		b.WriteString(fmt.Sprintf("- **RPC**: `%s`\n", rpcRemote))
	}
	b.WriteString("\n")

	b.WriteString("## What that means\n\n")
	b.WriteString("One validator produces every block. The operator can reset or restart this node " +
		"at any time, so treat anything you put here as disposable. The GNOT is premined for " +
		"testing: it has no value and buys nothing.\n\n")

	b.WriteString("**It is not a faithful rehearsal of mainnet.** gnodev leaves the chain more " +
		"permissive than gno.land, in ways that decide whether a deployment works at all:\n\n")
	b.WriteString("- namespace enforcement is off here, so `gno.land/r/<anything>/hello` deploys. " +
		"On mainnet `r/sys/names` is enabled from block 1 and the namespace has to be yours.\n")
	b.WriteString("- code submission is permissionless here. Mainnet runs " +
		"`code_submission_policy=inert`, so the same package lands inert and needs approval.\n")
	b.WriteString("- ugnot moves freely here. On mainnet transfers are restricted to an " +
		"exemption list.\n\n")

	b.WriteString("## The real network\n\n")
	b.WriteString(fmt.Sprintf("gno.land is at [%s](%s), chain ID `%s`. Nothing you do here reaches it.\n\n",
		strings.TrimPrefix(mainnetURL, "https://"), mainnetURL, mainnetChainID))
	b.WriteString("Your keys, however, do work on both: the same mnemonic produces the same `g1` " +
		"address on every chain. **Use a test key here, not one holding mainnet funds.**\n\n")

	b.WriteString("## What you can do here\n\n")
	b.WriteString("- **Browse** what is deployed, starting from [/r/](/r/).\n")
	b.WriteString("- **Read the source** of any package by adding `$source` to its path.\n")
	b.WriteString("- **Call a realm** from the Actions tab on its page.\n")
	b.WriteString("- **Deploy your own**, the same command you would use against mainnet.\n\n")

	b.WriteString("There is no faucet. Only the addresses the operator premined hold GNOT, " +
		"so ask them if you need funding.\n\n")

	b.WriteString("```\ngnokey maketx addpkg \\\n" +
		"  -pkgpath \"gno.land/r/<your-address>/hello\" -pkgdir . \\\n" +
		fmt.Sprintf("  -gas-fee 1000000ugnot -gas-wanted %d \\\n", addpkgGasWanted) +
		fmt.Sprintf("  -broadcast -chainid %s", chainID))
	// Always emit -remote. Omitting it is not neutral: gnokey falls back to
	// 127.0.0.1:26657, so a copied command either fails with connection
	// refused or, worse, silently lands on the reader's own local chain, since
	// that one defaults to chain ID "dev" as well.
	remote := rpcRemote
	if remote == "" {
		remote = "<this-chain-rpc>"
	}
	b.WriteString(fmt.Sprintf(" -remote %s", remote))
	b.WriteString(" \\\n  <your-key>\n```\n\n")
	b.WriteString("The package name must match the last element of `-pkgpath`.\n\n")

	b.WriteString("## Run your own\n\n")
	b.WriteString("This chain is one command. There is no compose file, no reverse-proxy dance and " +
		"no second container:\n\n")
	b.WriteString("```\ngnodev staging -chain-id my-preview -state-dir ./data\n```\n\n")
	b.WriteString("Point it at your own contracts by passing their directory, and every transaction " +
		"anyone sends is appended to `./data` and replayed the next time it starts. Use `gnodev` on " +
		"its own for local development, and `gnoland` only if you are running a validator on a real " +
		"network.\n\n")

	b.WriteString("## Learn more\n\n")
	b.WriteString("- [docs.gno.land](https://docs.gno.land)\n")
	b.WriteString("- [gnodev, which runs this chain](https://docs.gno.land/resources/gnodev)\n")

	return b.String()
}
