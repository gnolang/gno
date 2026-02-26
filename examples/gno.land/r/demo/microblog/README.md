# microblog realm

## Getting started:

(One-time) Add the microblog package:

```
gnokey maketx addpkg --pkgpath "gno.land/p/demo/microblog" --pkgdir "examples/gno.land/p/demo/microblog" \
    --deposit 100000000ugnot --gas-fee 1000000ugnot --gas-wanted 2000000 --chainid dev --remote localhost:26657 <YOURKEY>
```

(One-time) Add the microblog realm:

```
gnokey maketx addpkg --pkgpath "gno.land/r/demo/microblog" --pkgdir "examples/gno.land/r/demo/microblog" \
    --deposit 100000000ugnot --gas-fee 1000000ugnot --gas-wanted 2000000 --chainid dev --remote localhost:26657 <YOURKEY>
```

Add a microblog post:

```
gnokey maketx call --pkgpath "gno.land/r/demo/microblog" --func "NewPost" --args "hello, world" \
    --gas-fee "1000000ugnot" --gas-wanted "2000000" --chainid dev --remote localhost:26657 <YOURKEY>
```