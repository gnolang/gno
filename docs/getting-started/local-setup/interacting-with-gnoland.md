---
id: interacting-with-gnoland
---

# Interacting with Gno.land code

## Overview
In this tutorial, you will learn how to interact with Gno.land code.
You will understand how to use your keypair to send transactions to realms
and packages, send native coins, and more.

## Prerequisites
- **`gnokey` installed.** Reference the
[Local Setup](installation.md#2-installing-the-required-tools-) guide for steps
- **A keypair in `gnokey`.** Reference the 
[Working with Key Pairs](working-with-key-pairs.md#adding-a-private-key-using-a-mnemonic) guide for steps

## 1. Get testnet GNOTs
For interacting with any Gno.land chain, you will need a certain amount of GNOTs
to pay gas fees with. 

For this example, we will use the [Portal Loop](../../concepts/testnets.md#portal-loop) 
testnet. We can access the Portal Loop faucet through the
[Gno Faucet Hub](https://faucet.gno.land). There, you will find a card for each
available faucet.

After choosing the "Gno Portal Loop" card, you will get a prompt to input your address, 
select the amount of testnet GNOT you want to receive, and solve a captcha. In 
this case, 1 GNOT is enough.

After inputting your Gno address and solving the captcha, you can check if you
have received funds with the following `gnokey` command:

```bash
gnokey query bank/balances/<your_gno_address> --remote "https://rpc.gno.land:443"    
```

If the faucet request was successful, you should see something similar to the 
following:

```
❯ gnokey query bank/balances/<your_gno_address> --remote "https://rpc.gno.land:443"
height: 0
data: "10000000ugnot"
```

## 2. Visit a realm

For this example, we will use the [Userbook realm](https://gno.land/r/demo/userbook).
The Userbook realm is a simple app that allows users to sign up, and keeps track
of when they signed up. It also displays the currently signed-up users and the block
height at which they have signed up. The realm can be found on its path, 
[`gno.land/r/demo/userbook`](https://gno.land/r/demo/userbook).

To see what functions are available to call on the Userbook realm, click
the [`[help]`](https://gno.land/r/demo/userbook?help) button in the top right
corner. There, you will be able to see all callable functions in the realm,
and an interface that will generate `gnokey` commands based on your inputs.

In this case, we want to call the `SignUp()` function. First, you can input your
address (or keypair name) at the top. After doing this, `gnoweb` will give you a
command ready to paste into your terminal. For example, the following command will 
call the `SignUp` function with the keypair `MyKeypair`: 

```
gnokey maketx call \
-pkgpath "gno.land/r/demo/userbook" \
-func "SignUp" \
-gas-fee 1000000ugnot \
-gas-wanted 2000000 \
-send "" \
-broadcast \
-chainid "portal-loop" \
-remote "https://rpc.gno.land:443" \
MyKeypair
```

To see what each option and flag in this command does, read the `gnokey` 
[reference page](../../gno-tooling/cli/gnokey.md). 

## Conclusion

That's it! Congratulations on executing your first transaction on a Gno network! 🎉

If the previous transaction was successful, you should be able
to see your address on the main page of the Userbook realm. 

This concludes the "Local Setup" tutorial. For next steps, see the 
[How-to guides section](../../how-to-guides/how-to-guides.md), where you will 
learn how to write your first realm, package, and much more.