# Run a demo

> make

## sign an addpkg (add avl package) transaction.

./build/gnokey maketx addpkg test1 --pkgpath "gno.land/p/avl" --pkgdir "examples/gno.land/p/avl" --deposit 100gnot --gas-fee 1gnot --gas-wanted 200000 > addpkg.avl.unsigned.txt
./build/gnokey query "auth/accounts/g17d0j59qka0d2vtp4l78w0c59uy34cwsdpfvk5l"
./build/gnokey sign test1 --txpath addpkg.avl.unsigned.txt --chainid "testchain" --number 1 --sequence 0 > addpkg.avl.signed.txt
./build/gnokey broadcast addpkg.avl.signed.txt

## sign an addpkg (add dom package) transaction.

./build/gnokey maketx addpkg test1 --pkgpath "gno.land/p/dom" --pkgdir "examples/gno.land/p/dom" --deposit 100gnot --gas-fee 1gnot --gas-wanted 200000 > addpkg.dom.unsigned.txt
./build/gnokey query "auth/accounts/g17d0j59qka0d2vtp4l78w0c59uy34cwsdpfvk5l"
./build/gnokey sign test1 --txpath addpkg.dom.unsigned.txt --chainid "testchain" --number 1 --sequence 1 > addpkg.dom.signed.txt
./build/gnokey broadcast addpkg.dom.signed.txt

## sign an addpkg (add example realm) transaction.

./build/gnokey maketx addpkg test1 --pkgpath "gno.land/r/example" --pkgdir "examples/gno.land/r/example" --deposit 100gnot --gas-fee 1gnot --gas-wanted 200000 > addrealm.unsigned.txt
./build/gnokey query "auth/accounts/g17d0j59qka0d2vtp4l78w0c59uy34cwsdpfvk5l"
./build/gnokey sign test1 --txpath addrealm.unsigned.txt --chainid "testchain" --number 1 --sequence 2 > addrealm.signed.txt
./build/gnokey broadcast addrealm.signed.txt

## sign an eval (expression evaluation) transaction.

./build/gnokey maketx eval test1 --pkgpath "gno.land/r/example" --expr "AddPost(\"SUCCESS\", \"A POST BODY\")" --gas-fee 1gnot --gas-wanted 200000 > addpage.unsigned.txt
./build/gnokey query "auth/accounts/g17d0j59qka0d2vtp4l78w0c59uy34cwsdpfvk5l"
./build/gnokey sign test1 --txpath addpage.unsigned.txt --chainid "testchain" --number 1 --sequence 3 > addpage.signed.txt
./build/gnokey broadcast addpage.signed.txt

## render page with ABCI query (evalquery).

./build/gnokey query "vm/qeval" --data "gno.land/r/example
RenderPage(\"woot\")"
