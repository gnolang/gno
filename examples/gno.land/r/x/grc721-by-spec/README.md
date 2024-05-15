# grc721-by-spec

This folder contains 3 main parts:
1. The Go implementation of the ERC721 standard, called GRC721, which was used 
for the initial implementation & testing
2. The Gno package implementation ported from the Go implementation mentioned above
3. An example NFT collection realm utilizing the Gno package

To test this out, install `gnodev` and run the following command in this folder:
```shell
gnodev ./grc721-gno ./exampleNFT
```

Then, visit [`localhost:8888/r/example/nft`](http://localhost:8888/r/example/nft).
