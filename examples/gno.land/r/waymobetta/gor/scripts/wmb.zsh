#!/bin/zsh

gnokey query \
"vm/qeval" \
-data "gno.land/r/waymobetta/gor
GetGnoGitMapping(\"g1c74t34ukg2lq39nxx5cddlkvjtfrm3zchnzvk6\")" \
-remote localhost:26657

gnokey query \
"vm/qeval" \
-data "gno.land/r/waymobetta/gor
GetGnoGitMapping(\"g1c74t34ukg2lq39nxx5cddlkvjtfrm3zchnzvk7\")" \
-remote localhost:26657

gnokey query \
"vm/qeval" \
-data "gno.land/r/waymobetta/gor
GetGitGnoMapping(\"waymobetta\")" \
-remote localhost:26657

gnokey query \
"vm/qeval" \
-data "gno.land/r/waymobetta/gor
GetGitGnoMapping(\"moul\")" \
-remote localhost:26657

# gnokey query \
# "vm/qeval" \
# -data "gno.land/r/waymobetta/gor
# GetGnoPRMapping(\"g1c74t34ukg2lq39nxx5cddlkvjtfrm3zchnzvk6\")" \
# -remote localhost:26657

# gnokey query \
# "vm/qeval" \
# -data "gno.land/r/waymobetta/gor
# GnoPRCount(\"g1c74t34ukg2lq39nxx5cddlkvjtfrm3zchnzvk6\")" \
# -remote localhost:26657

# gnokey query \
# "vm/qeval" \
# -data "gno.land/r/waymobetta/gor
# GitPRCount(\"waymobetta\")" \
# -remote localhost:26657
