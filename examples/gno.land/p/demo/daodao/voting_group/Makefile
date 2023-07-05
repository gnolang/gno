pkgpath=p/demo/daodao/voting_group
pkguri=gno.land/$(pkgpath)
pkgsrc=.

gnohome=$(HOME)/Code/gno

chainid=dev
gnoland=localhost:26657
gnoweb=http://localhost:8888
wallet=onblock-test

#chainid=test3
#gnoland=test3.gno.land:36657
#gnoweb=https://test3.gno.land
#wallet=dev

.PHONY: deploy.package
deploy.pkg:
	gnokey maketx addpkg  \
		-deposit="1ugnot" \
		-gas-fee="1ugnot" \
		-gas-wanted="5000000" \
		-broadcast="true" \
		-remote="$(gnoland)" \
		-chainid="$(chainid)" \
		-pkgdir="./$(pkgsrc)" \
		-pkgpath="$(pkguri)" \
		$(wallet)

.PHONY: open.pkg
open.pkg:
	open "$(gnoweb)/$(pkgpath)"

.PHONY: test
test:
	gno test -verbose=true -root-dir="$(gnohome)" ./$(pkgsrc)