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
  -pkgdir="./r/demo/justicedao" \
  -pkgpath="gno.land/r/demo/justicedao_10" \
  teritori

# Create DAO
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/justicedao_10" \
  -func="CreateDAO" \
  -args="https://gnodao1.org" \
  -args="https://metadata.gnodao1.org" \
  -args=$GOPHER,$TERITORI \
  -args="1,1" \
  -args="40" \
  -args="30" \
  -args="10" \
  -args="10" \
  -args="1" \
  teritori

# Create Justice DAO proposal
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/justicedao_10" \
  -func="CreateJusticeProposal" \
  -args="First Justice DAO proposal" \
  -args="First Justice DAO proposal summary" \
  -args="1" \
  teritori

# Fulfill Random Words on VRF
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/vrf_08" \
  -func="FulfillRandomWords" \
  -args="7" \
  -args="f440c4980357d8b56db87ddd50f06bd551f1319b" \
  teritori

# Determine Juste DAO members
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/justicedao_10" \
  -func="DetermineJusticeDAOMembers" \
  -args="0" \
  teritori

# Propose Justice DAO Solution
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/justicedao_10" \
  -func="ProposeJusticeDAOSolution" \
  -args="0" \
  -args="50" \
  -args="Split 50:50" \
  teritori

# Vote Justice Solution Proposal
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/justicedao_10" \
  -func="VoteJusticeSolutionProposal" \
  -args="0" \
  -args="0" \
  teritori

# Tally And Execute Justice Solution
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/justicedao_10" \
  -func="TallyAndExecuteJusticeSolution" \
  -args="0" \
  teritori

# Create Normal Proposal
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/justicedao_10" \
  -func="CreateProposal" \
  -args="First proposal" \
  -args="First proposal summary" \
  -args=0 \
  -args=$GOPHER \
  -args="" \
  -args="" \
  -args="https://metadata.gnodao1.com" \
  -args="https://gnodao1.com" \
  teritori

# Query proposal
gnokey query "vm/qeval" -data="gno.land/r/demo/justicedao_10
RenderProposal(0)" -remote="51.15.236.215:26657"

# Render Juste DAO Proposal
gnokey query "vm/qeval" -data="gno.land/r/demo/justicedao_10
RenderJusticeDAOProposal(0)" -remote="51.15.236.215:26657"

# Render Justice DAO Proposals
gnokey query "vm/qeval" -data="gno.land/r/demo/justicedao_10
RenderJusticeDAOProposals(0, 1)" -remote="51.15.236.215:26657"

gnokey query "vm/qeval" -data="gno.land/r/demo/justicedao_10
GetDAOMembers()" -remote="51.15.236.215:26657"

gnokey query "vm/qeval" -data="gno.land/r/demo/justicedao_10
RenderDAOMembers(\"\",\"\")" -remote="51.15.236.215:26657"
