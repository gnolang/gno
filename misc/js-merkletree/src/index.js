const { MerkleTree } = require("merkletreejs");
const crypto = require("crypto");
const SHA256 = require("crypto-js/sha256");
const secp256k1 = require("secp256k1");
const { bech32 } = require("bech32");

// This function reproduce the output of std.DerivePkgAddr
function derive_pkg_addr(str) {
  const hash = crypto
    .createHash("sha256")
    .update("pkgPath:" + str)
    .digest();

  let hash_bytes = new Uint8Array(hash).slice(0, 20);

  const words = bech32.toWords(hash_bytes);
  return bech32.encode("g", words);
}

// Start script

for (size of [1, 3, 10, 1000]) {
  console.log(`[INFO] Start with size ${size}`);

  const addresses = Array.from({ length: size }, (_, i) =>
    derive_pkg_addr(`gno.land/r/test/${i}`),
  );

  const leaves = addresses.map((addr) =>
    SHA256(JSON.stringify({ address: addr, amount: "10000" })),
  );

  const tree = new MerkleTree(leaves, SHA256);
  const root = tree.getRoot().toString("hex");

  console.log(root);
  // const proof = tree.getProof(leaves[5]).map((p) => ({
  //   data: p.data.toString("hex"),
  //   position: p.position === "left" ? 1 : 0,
  // }));
}

// const tree = new MerkleTree(leaves, SHA256);
// const root = tree.getRoot().toString("hex");

// const proof = tree.getProof(leaves[5]).map((p) => ({
//   data: p.data.toString("hex"),
//   position: p.position === "left" ? 1 : 0,
// }));

// const req = {
//   address: derive_pkg_addr("gno.land/r/test/5"),
//   amount: "10000",
//   proof: proof,
// };

// console.log("req", JSON.stringify(req));

// console.log("root", root);
