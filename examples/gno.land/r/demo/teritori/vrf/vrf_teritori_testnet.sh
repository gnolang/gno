#!/bin/sh

gnokey add gopher
- addr: g1x2xyqca98auaw9lnat2h9ycd4lx3w0jer9vjmt

gnokey add gopher2
- addr: g1c5y8jpe585uezcvlmgdjmk5jt2glfw88wxa3xq

TERITORI=g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
GOPHER=g1x2xyqca98auaw9lnat2h9ycd4lx3w0jer9vjmt

# check balance
gnokey query bank/balances/$GOPHER -remote="51.15.236.215:26657"

gnokey maketx addpkg  \
  -deposit="1ugnot" \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgdir="./r/vrf" \
  -pkgpath="gno.land/r/demo/vrf_08" \
  teritori

# Set VRF admin
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/vrf_08" \
  -func="SetVRFAdmin" \
  -args="$TERITORI" \
  teritori

# Set feeders
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/vrf_08" \
  -func="SetFeeders" \
  -args="$TERITORI" \
  teritori

# Request Random Words
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/vrf_08" \
  -func="RequestRandomWords" \
  -args="1" \
  teritori

# Fulfill Random Words
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/vrf_08" \
  -func="FulfillRandomWords" \
  -args="0" \
  -args="f440c4980357d8b56db87ddd50f06bd551f1319a" \
  teritori

# Query config
gnokey query "vm/qeval" -data="gno.land/r/demo/vrf_08
RenderConfig()" -remote="51.15.236.215:26657"

# Query Requests
gnokey query "vm/qeval" -data="gno.land/r/demo/vrf_08
RenderRequests(0, 10)" -remote="51.15.236.215:26657"

# Query request
gnokey query "vm/qeval" -data="gno.land/r/demo/vrf_08
RenderRequest(0)" -remote="51.15.236.215:26657"

# Query IsFeeder
gnokey query "vm/qeval" -data="gno.land/r/demo/vrf_08
IsFeeder(\"$TERITORI\")" -remote="51.15.236.215:26657"

# Query RandomValueFromWordsWithIndex
gnokey query "vm/qeval" -data="gno.land/r/demo/vrf_08
RandomValueFromWordsWithIndex(0, 0)" -remote="51.15.236.215:26657"

# Query RandomValueFromWordsWithIndex
gnokey query "vm/qeval" -data="gno.land/r/demo/vrf_08
RandomValueFromWords(0)" -remote="51.15.236.215:26657"
