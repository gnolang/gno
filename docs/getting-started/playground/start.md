---
id: start
---

# Gno Playground

## Overview

The Gno Playground is a robust web platform that enables developers to 
interactively work with the Gno language. It streamlines coding, testing,
and deploying with its diverse set of tools and features.

Essentially, the Playground supports direct browser-based Gno code execution, 
promoting quick feedback and continuous improvement. It's crucial for mastering 
and enhancing Gno skills. Additionally, users can share code, run tests, and 
deploy projects to Gno.land, enhancing collaboration and productivity.

## Prerequisites

- **Internet connection**
- 




To follow along, you will need to install a Gno.land web browser wallet, such as
[Adena](https://www.adena.app/), and create a keypair. This will allow you to
interact with the Playground.

Next, visit the [Playground](https://play.gno.land). You will be greeted with a
simple `package.gno` file.



First we should test and deploy the `whitelist` package. To do this, delete `package.gno`,
and create files like before: `whitelist.gno` & `whitelist_test.gno`. Then,
paste in the respective code, or just visit [this link](https://play.gno.land/p/t1AXy1wxafC)
with the pre-written code.

Gno Playground allows you to test, deploy, and share code in your browser.
Clicking on "Test" will open a terminal and after a few seconds you should see
the following output:



After we've verified our code works, we are ready to deploy the package code to
the test3 testnet. Clicking on the "Deploy" button will prompt a wallet connection, and then
you will see the following:



Change the deployment path as you see fit - for this we will go with
`gno.land/p/leon/whitelist`. Keep in mind that this is the path you will use
to later import the package and use it for the `WhitelistFactory` realm.

Choose `Testnet 3` for the network and click `Deploy`.

Gno Playground has a built-in faucet, which means that even if you do not have any
test3 GNOTs, the deployment should result in a success, and you will be presented
with a [Gnoscan link](https://gnoscan.io/transactions/details?txhash=pCBe5tZVD+5bvWE2vUJosxfwkSUSHJE9zbVahVs4vBA%3D)
for the deployment transaction.

After successfully deploying the package, we can continue with the realm code.

Delete the old files, and create a new one - `whitelistfactory.gno`.
Paste in the code, or simply find it on [this link](https://play.gno.land/p/M_ehuoP4jsM).


After inserting your package path, you can click deploy the realm to your chosen
path. To view the realm on chain, visit `https://test3.<your_realm_path>`.

This concludes our tutorial. Once again, congratulations on writing
your first realm in Gno. You've become a real Gno.Land hero!

If you'd like to see the full repository used for this tutorial,
it can be found [here](https://github.com/leohhhn/gno/tree/from_zero_to_gnoland_hero).
