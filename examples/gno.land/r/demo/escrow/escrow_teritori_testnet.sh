#!/bin/sh

gnokey add gopher
- addr: g1x2xyqca98auaw9lnat2h9ycd4lx3w0jer9vjmt

gnokey add gopher2
- addr: g1c5y8jpe585uezcvlmgdjmk5jt2glfw88wxa3xq

TERITORI=g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
GOPHER=g1x2xyqca98auaw9lnat2h9ycd4lx3w0jer9vjmt

# check balance
gnokey query bank/balances/$GOPHER -remote="51.15.236.215:26657"

# Send balance to gopher2 account
gnokey maketx send  \
  -send="10000000ugnot" \
  -to="g1c5y8jpe585uezcvlmgdjmk5jt2glfw88wxa3xq" \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  teritori

gnokey maketx addpkg  \
  -deposit="1ugnot" \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgdir="./r/escrow" \
  -pkgpath="gno.land/r/demo/escrow_05" \
  teritori

# Set config
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/escrow_05" \
  -func="UpdateConfig" \
  -args="g1c5y8jpe585uezcvlmgdjmk5jt2glfw88wxa3xq" \
  teritori

# Create Contract
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/escrow_05" \
  -func="CreateContract" \
  -args="g1c5y8jpe585uezcvlmgdjmk5jt2glfw88wxa3xq" \
  -args="gopher20" \
  -args="100" \
  -args="60" \
  teritori

# Cancel Contract
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/escrow_05" \
  -func="CancelContract" \
  -args="0" \
  teritori

# Accept Contract
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/escrow_05" \
  -func="AcceptContract" \
  -args="1" \
  gopher2

# Pause Contract
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/escrow_05" \
  -func="PauseContract" \
  -args="0" \
  teritori

# Complete Contract
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/escrow_05" \
  -func="CompleteContract" \
  -args="1" \
  teritori

# Complete Contract by DAO
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/escrow_05" \
  -func="CompleteContractByDAO" \
  -args="0" \
  -args="50" \
  teritori

# Give feedback
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/escrow_05" \
  -func="GiveFeedback" \
  -args="0" \
  -args="Amazing work" \
  teritori

# Query Contracts
gnokey query "vm/qeval" -data="gno.land/r/demo/escrow_05
RenderContracts(0, 10)" -remote="51.15.236.215:26657"

# Query contract
gnokey query "vm/qeval" -data="gno.land/r/demo/escrow_05
RenderContract(0)" -remote="51.15.236.215:26657"

# Query config
gnokey query "vm/qeval" -data="gno.land/r/demo/escrow_05
RenderConfig()" -remote="51.15.236.215:26657"

# Query escrow address
gnokey query "vm/qeval" -data="gno.land/r/demo/escrow_05
CurrentRealm()" -remote="51.15.236.215:26657"


# Get gopher20 faucet
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/gopher20" \
  -func="Faucet" \
  teritori

# Approve tokens
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/gopher20" \
  -func="Approve" \
  -args="g1f7p4tuu044w2qsa9m3h64ql4lrqmmjzm2f6jws" \
  -args="1000" \
  teritori

# Query balance
gnokey query "vm/qeval" -data="gno.land/r/demo/gopher20
BalanceOf(\"$TERITORI\")" -remote="51.15.236.215:26657" 

gnokey query "vm/qeval" -data="gno.land/r/demo/gopher20
Render(\"balance/g1c5y8jpe585uezcvlmgdjmk5jt2glfw88wxa3xq\")" -remote="51.15.236.215:26657"
