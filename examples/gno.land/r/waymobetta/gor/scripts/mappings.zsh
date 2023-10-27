#!/bin/zsh

gnokey query \
"vm/qeval" \
-data "gno.land/r/waymobetta/gor
GitGnoMapping" \
-remote localhost:26657

gnokey query \
"vm/qeval" \
-data "gno.land/r/waymobetta/gor
GnoGitMapping" \
-remote localhost:26657

gnokey query \
"vm/qeval" \
-data "gno.land/r/waymobetta/gor
GnoPRMapping" \
-remote localhost:26657
