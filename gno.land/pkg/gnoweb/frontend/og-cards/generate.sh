#!/bin/sh
# generate.sh redraws the community share cards into ../static/imgs; run it
# with `make og-cards` from gno.land/pkg/gnoweb after changing a card's
# wording here. It needs macOS (swift, CoreText) and Go.
set -eu
cd "$(dirname "$0")"
imgs=../static/imgs
font=../static/fonts/intervar/Intervar.woff2
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

card() { # name headline subtitle footer
	HEAD=$2 SUB=$3 FOOT=$4 swift card.swift "$imgs/og-gnoland.png" "$font" "$tmp/$1.png"
	go run quantize.go "$tmp/$1.png" "$imgs/og-community-$1.png"
}

card realm "Community realm" "Open source, running on-chain. Read its code, call its functions." "Deployed by its author · check before you sign"
card package "Community package" "Open source, stored on-chain. Read its code, import it." "Deployed by its author · read before you import"
card user "Community profile" "Realms and packages published on gno.land." "Written by its owner · check before you sign"
