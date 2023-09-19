#!/bin/sh

gnokey add gopher
- addr: g1x2xyqca98auaw9lnat2h9ycd4lx3w0jer9vjmt

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
  -pkgdir="./r/gnodao" \
  -pkgpath="gno.land/r/demo/gnodao_v05" \
  gopher

# Create DAO
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="test3.gno.land:36657" \
  -chainid="test3" \
  -pkgpath="gno.land/r/demo/gnodao_v05" \
  -func="CreateDAO" \
  -args="https://gnodao1.org" \
  -args="https://metadata.gnodao1.org" \
  -args=$GOPHER \
  -args="1" \
  -args="40" \
  -args="30" \
  -args="10" \
  -args="10" \
  gopher

# Create Proposal
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="test3.gno.land:36657" \
  -chainid="test3" \
  -pkgpath="gno.land/r/demo/gnodao_v05" \
  -func="CreateProposal" \
  -args=0 \
  -args="First proposal" \
  -args="First proposal summary" \
  -args=0 \
  -args=$GOPHER \
  -args="" \
  -args="" \
  -args="https://metadata.gnodao1.com" \
  -args="https://gnodao1.com" \
  gopher

# Vote Proposal
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="test3.gno.land:36657" \
  -chainid="test3" \
  -pkgpath="gno.land/r/demo/gnodao_v05" \
  -func="VoteProposal" \
  -args=0 \
  -args=0 \
  -args=0 \
  gopher

# Tally and execute
gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="test3.gno.land:36657" \
  -chainid="test3" \
  -pkgpath="gno.land/r/demo/gnodao_v05" \
  -func="TallyAndExecute" \
  -args=0 \
  -args=0 \
  gopher

# Query DAOs
gnokey query "vm/qeval" -data="gno.land/r/demo/gnodao_v05
RenderDAOs(0, 10)" -remote="test3.gno.land:36657"

# Query DAO
gnokey query "vm/qeval" -data="gno.land/r/demo/gnodao_v05
RenderDAO(0)" -remote="test3.gno.land:36657"

# Query Proposal
gnokey query "vm/qeval" -data="gno.land/r/demo/gnodao_v05
RenderProposal(0, 0)" -remote="test3.gno.land:36657"

gnokey query "vm/qeval" -data="gno.land/r/demo/gnodao_v05
RenderProposals(0, 0,10)" -remote="test3.gno.land:36657"

gnokey query "vm/qeval" -data='gno.land/r/demo/gnodao_v05
RenderDAOMembers(0, "", "zz")' -remote="test3.gno.land:36657"
