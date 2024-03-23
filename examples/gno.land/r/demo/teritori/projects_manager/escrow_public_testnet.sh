#!/bin/sh

gnokey add gopher
- addr: g1x2xyqca98auaw9lnat2h9ycd4lx3w0jer9vjmt

gnokey add gopher2
- addr: g1c5y8jpe585uezcvlmgdjmk5jt2glfw88wxa3xq

GOPHER=g1x2xyqca98auaw9lnat2h9ycd4lx3w0jer9vjmt

# check balance
gnokey query bank/balances/$GOPHER -remote="test3.gno.land:36657"

gnokey maketx addpkg  \
  -deposit="1ugnot" \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="test3.gno.land:36657" \
  -chainid="test3" \
  -pkgdir="./r/projects_manager" \
  -pkgpath="gno.land/r/demo/projects_manager_03" \
  gopher

# Set config
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="test3.gno.land:36657" \
  -chainid="test3" \
  -pkgpath="gno.land/r/demo/projects_manager_03" \
  -func="UpdateConfig" \
  -args="g1c5y8jpe585uezcvlmgdjmk5jt2glfw88wxa3xq" \
  gopher

# Create Contract
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="test3.gno.land:36657" \
  -chainid="test3" \
  -pkgpath="gno.land/r/demo/projects_manager_03" \
  -func="CreateContract" \
  -args="g1c5y8jpe585uezcvlmgdjmk5jt2glfw88wxa3xq" \
  -args="foo20" \
  -args="100" \
  -args="60" \
  gopher

# Cancel Contract
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="test3.gno.land:36657" \
  -chainid="test3" \
  -pkgpath="gno.land/r/demo/projects_manager_03" \
  -func="CancelContract" \
  -args="0" \
  gopher

# Accept Contract
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="test3.gno.land:36657" \
  -chainid="test3" \
  -pkgpath="gno.land/r/demo/projects_manager_03" \
  -func="AcceptContract" \
  -args="0" \
  gopher

# Pause Contract
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="test3.gno.land:36657" \
  -chainid="test3" \
  -pkgpath="gno.land/r/demo/projects_manager_03" \
  -func="PauseContract" \
  -args="0" \
  gopher

# Complete Contract
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="test3.gno.land:36657" \
  -chainid="test3" \
  -pkgpath="gno.land/r/demo/projects_manager_03" \
  -func="CompleteContract" \
  -args="0" \
  gopher

# Complete Contract by DAO
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="test3.gno.land:36657" \
  -chainid="test3" \
  -pkgpath="gno.land/r/demo/projects_manager_03" \
  -func="CompleteContractByDAO" \
  -args="0" \
  -args="50" \
  gopher

# Give feedback
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="test3.gno.land:36657" \
  -chainid="test3" \
  -pkgpath="gno.land/r/demo/projects_manager_03" \
  -func="GiveFeedback" \
  -args="0" \
  -args="Amazing work" \
  gopher

# Query Contracts
gnokey query "vm/qeval" -data="gno.land/r/demo/projects_manager_03
RenderContracts(0, 10)" -remote="test3.gno.land:36657"

# Query contract
gnokey query "vm/qeval" -data="gno.land/r/demo/projects_manager_03
RenderContract(0)" -remote="test3.gno.land:36657"

# Query config
gnokey query "vm/qeval" -data="gno.land/r/demo/projects_manager_03
RenderConfig()" -remote="test3.gno.land:36657"


# Get foo20 faucet
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="test3.gno.land:36657" \
  -chainid="test3" \
  -pkgpath="gno.land/r/demo/foo20" \
  -func="Faucet" \
  gopher

# Approve tokens
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="test3.gno.land:36657" \
  -chainid="test3" \
  -pkgpath="gno.land/r/demo/foo20" \
  -func="Approve" \
  -args="g1c5y8jpe585uezcvlmgdjmk5jt2glfw88wxa3xq" \
  -args="1000" \
  gopher

# Query balance
gnokey query "vm/qeval" -data="gno.land/r/demo/foo20
BalanceOf(\"$GOPHER\")" -remote="test3.gno.land:36657" 

gnokey query "vm/qeval" -data="gno.land/r/demo/foo20
Render(\"balance/$GOPHER\")" -remote="test3.gno.land:36657"
