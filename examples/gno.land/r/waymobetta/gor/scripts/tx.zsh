#!/bin/zsh

gnokey maketx \
call \
--pkgpath "gno.land/r/waymobetta/gor" \
--func "NewPR" \
--args 1 \
--args 100 \
--args 100 \
--args 1 \
--args 50 \
--args 0 \
--args 100 \
--args 50 \
--args "waymobetta" \
--args "dao" \
--args "feature" \
--gas-fee 1000000ugnot \
--gas-wanted 2000000 \
--broadcast \
--chainid dev \
--remote localhost:26657 \
demo

gnokey maketx \
call \
--pkgpath "gno.land/r/waymobetta/gor" \
--func "NewPR" \
--args 2 \
--args 100 \
--args 100 \
--args 1 \
--args 50 \
--args 0 \
--args 100 \
--args 50 \
--args "waymobetta" \
--args "core" \
--args "bug" \
--gas-fee 1000000ugnot \
--gas-wanted 2000000 \
--broadcast \
--chainid dev \
--remote localhost:26657 \
demo

#  These commands are used to interact with the Gnokey CLI tool for executing transactions. 

# The `call` subcommand initiates this process, and we're passing several arguments to it using flags. 

# Here's what each flag does:

# - `--pkgpath`: Specifies the path of the module within its repository (in our case, it's "gno.land/r/waymobetta").

# - `--func`: Sets the name of the function we want to call on this realm (i.e., "NewPR" in both cases).

# - `--args`: Passes arguments to the selected function, one at a time using multiple instances of this flag (in our case, we're passing details of a PR for both calls).

# - `--gas-fee` and `--gas-wanted`: Set the gas fee and desired gas limit for the transaction, respectively. These values are used by the realm to calculate the actual amount of GNOT (or ugnot, a smaller denomination) required to execute this transaction on the Gno.land network. In our case, we're setting both flags to 1 million units of ugnot tokens as a way to ensure that there are enough funds available for gas fees and to avoid underestimating the required amount of ugnot during transaction execution.

# - `--broadcast`: Broadcasts the transaction hash (i.e., its unique identifier) once it's been executed on the Gno.land network, making it visible to other users who are also monitoring this address.

# - `--chainid`: Specifies which blockchain network you want to use (in our case, "dev" for Gno.land's development environment).

# - `--remote`: Sets the URL of a local Gno.land node running on your machine that will be used to execute this transaction. In our case, we're using the default value ("localhost:26657") since we've already started an instance of Gno (a Gno.land client) in another terminal window earlier on.

# - `demo`: This is a placeholder argument that you can replace with your own account (wallet) name, depending on how you want to organize and label your accounts. In our case, we're using "demo" as a simple identifier for this particular transaction.

# By executing these commands, we're able to interact directly with the realm we deployed in previous steps. The specific values passed here will depend on your use case and the desired behavior of your realm, but this should give you a good starting point for getting started with Gnokey CLI tool
