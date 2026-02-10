## Overview

`gnogenesis` is a CLI tool for managing the Gnoland blockchain's `genesis.json` file. It provides
subcommands for setting up and manipulating the genesis file, from generating a new genesis configuration to managing
initial validators, balances, and transactions.

Refer to specific command help options (`--help`) for further customization options.

## Installation

To install gnogenesis, clone the repository and build the tool:

```shell
git clone https://github.com/gnolang/gno.git
cd gno/contribs/gnogenesis
make install
```

This will compile and install `gnogenesis` to your system path, allowing you to run commands directly.

## Features

### Generate a `genesis.json`

To create a new genesis.json, use the `generate` subcommand. You can specify parameters such as chain ID, block limits,
and more:

```shell
gnogenesis generate --chain-id gno-dev --block-max-gas 100000000 --output-path ./genesis.json
```

This command generates a genesis.json file with custom parameters, defining the chain’s identity, block limits, and
more. By default, the genesis-time is set to the current timestamp, or you can specify a future time for scheduled chain
launches.

Keep in mind the `genesis.json` is generated with an empty validator set, and you will need to manually add the initial
validators.

### Manage initial validators

The `validator` subcommands allow you to add or remove validators directly in the genesis file.

#### Add a validator

To add a validator, specify their `address`, `name`, and `pub-key`:

```shell
gnogenesis validator add --address g1rzuwh5frve732k4futyw45y78rzuty4626zy6h --name validator1 --pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zplmcmggxyxyrch0zcyg684yxmerullv3l6hmau58sk4eyxskmny9h7lsnz
```

This command will add the validator with the specified details in the genesis file.

The `address` and `pub-key` values need to be in bech32 format. They can be fetched using `gnoland secrets get`.

#### Remove a validator

If you need to remove a validator, specify their address:

```shell
gnogenesis validator remove --address g1rzuwh5frve732k4futyw45y78rzuty4626zy6h
```

This will remove the specified validator from the validator set in `genesis.json`, if it is present.

### Verify the `genesis.json`

The `verify` subcommand is helpful to confirm the integrity of a `genesis.json` file:

```shell
gnogenesis verify --genesis-path ./genesis.json
```

This validation checks for proper structure, account balance totals, and ensures validators are correctly configured,
preventing common genesis setup issues. It is advised to always run this verification step when dealing with an external
`genesis.json`.

### Manage account balances

Balances can be added or removed through the balances subcommand, either individually or using a balance sheet file.

The format for individual balance entries is `<address>=<amount>ugnot`.

#### Add Account Balances

Add a single balance directly:

```shell
gnogenesis balances add --single g1rzuwh5frve732k4futyw45y78rzuty4626zy6h=100ugnot
```

Alternatively, load multiple accounts with a balance sheet file:

```shell
gnogenesis balances add --balance-sheet ./balances.txt
```

The format of the balance sheet file is the same as with individual entries, for example:

```text
# Test accounts.
g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5=10000000000000ugnot # test1
g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj=10000000000000ugnot # test2

# Faucet accounts.
g1f4v282mwyhu29afke4vq5r2xzcm6z3ftnugcnv=1000000000000ugnot # faucet0 (jae)
g127jydsh6cms3lrtdenydxsckh23a8d6emqcvfa=1000000000000ugnot # faucet1 (moul)
g1q6jrp203fq0239pv38sdq3y3urvd6vt5azacpv=1000000000000ugnot # faucet2 (devx)
g13d7jc32adhc39erm5me38w5v7ej7lpvlnqjk73=1000000000000ugnot # faucet3 (devx)
g18l9us6trqaljw39j94wzf5ftxmd9qqkvrxghd2=1000000000000ugnot # faucet4 (adena)
```

This will update `genesis.json` with the provided accounts and balances.

#### Remove account balances

To remove an account’s balance from `genesis.json`, use:

```shell
gnogenesis balances remove --address g1rzuwh5frve732k4futyw45y78rzuty4626zy6h
```

This deletes the balance entry for the specified address, if present.

### Handle genesis transactions

The `txs` subcommand allows you to manage initial transactions.

It is a bit more robust than the `balances` command suite, in the sense that it supports:

- adding transactions from transaction sheets
- generating and adding deploy transactions from a directory (ex. like `examples`)

The format for transactions in the transaction sheet is the following:

- Transaction (`std.Tx`) is encoded in Amino JSON
- Transactions are saved single-line, 1 line 1 tx
- File format of the transaction sheet file is `jsonl`

#### Add genesis transactions

To add genesis transactions from a file:

```shell
gnogenesis txs add sheets ./txs.json
```

This outputs the initial transaction count.

An example transaction sheet:

```json lines
{"msg":[{"@type":"/vm.m_call","caller":"g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq","send":"","pkg_path":"gno.land/r/demo/users","func":"Invite","args":["g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj:10\ng1589c8cekvmjfmy0qrd4f3z52r7fn7rgk02667s:1\ng13sm84nuqed3fuank8huh7x9mupgw22uft3lcl8:1\ng1m6732pkrngu9vrt0g7056lvr9kcqc4mv83xl5q:1\ng1wg88rhzlwxjd2z4j5de5v5xq30dcf6rjq3dhsj:1\ng18pmaskasz7mxj6rmgrl3al58xu45a7w0l5nmc0:1\ng19wwhkmqlns70604ksp6rkuuu42qhtvyh05lffz:1\ng187982000zsc493znqt828s90cmp6hcp2erhu6m:1\ng1ndpsnrspdnauckytvkfv8s823t3gmpqmtky8pl:1\ng16ja66d65emkr0zxd2tu7xjvm7utthyhpej0037:1\ng1ds24jj9kqjcskd0gzu24r9e4n62ggye230zuv5:1\ng1trkzq75ntamsnw9xnrav2v7gy2lt5g6p29yhdr:1\ng1rrf8s5mrmu00sx04fzfsvc399fklpeg2x0a7mz:1\ng19p5ntfvpt4lwq4jqsmnxsnelhf3tff9scy3w8w:1\ng1tue8l73d6rq4vhqdsp2sr3zhuzpure3k2rnwpz:1\ng14hhsss4ngx5kq77je5g0tl4vftg8qp45ceadk3:1\ng1768hvkh7anhd40ch4h7jdh6j3mpcs7hrat4gl0:1\ng15fa8kyjhu88t9dr8zzua8fwdvkngv5n8yqsm0n:1\ng1xhccdjcscuhgmt3quww6qdy3j3czqt3urc2eac:1\ng1z629z04f85k4t5gnkk5egpxw9tqxeec435esap:1\ng1pfldkplz9puq0v82lu9vqcve9nwrxuq9qe5ttv:1\ng152pn0g5qfgxr7yx8zlwjq48hytkafd8x7egsfv:1\ng1cf2ye686ke38vjyqakreprljum4xu6rwf5jskq:1\ng1c5shztyaj4gjrc5zlwmh9xhex5w7l4asffs2w6:1\ng1lhpx2ktk0ha3qw42raxq4m24a4c4xqxyrgv54q:1\ng1026p54q0j902059sm2zsv37krf0ghcl7gmhyv7:1\ng1n4yvwnv77frq2ccuw27dmtjkd7u4p4jg0pgm7k:1\ng13m7f2e6r3lh3ykxupacdt9sem2tlvmaamwjhll:1\ng19uxluuecjlsqvwmwu8sp6pxaaqfhk972q975xd:1\ng1j80fpcsumfkxypvydvtwtz3j4sdwr8c2u0lr64:1\ng1tjdpptuk9eysq6z38nscqyycr998xjyx3w8jvw:1\ng19t3n89slfemgd3mwuat4lajwcp0yxrkadgeg7a:1\ng1yqndt8xx92l9h494jfruz2w79swzjes3n4wqjc:1\ng13278z0a5ufeg80ffqxpda9dlp599t7ekregcy6:1\ng1ht236wjd83x96uqwh9rh3fq6pylyn78mtwq9v6:1\ng1fj9jccm3zjnqspq7lp2g7lj4czyfq0s35600g9:1\ng1wwppuzdns5u6c6jqpkzua24zh6ppsus6399cea:1\ng1k8pjnguyu36pkc8hy0ufzgpzfmj2jl78la7ek3:1\ng1e8umkzumtxgs8399lw0us4rclea3xl5gxy9spp:1\ng14qekdkj2nmmwea4ufg9n002a3pud23y8k7ugs5:1\ng19w2488ntfgpduzqq3sk4j5x387zynwknqdvjqf:1\ng1495y3z7zrej4rendysnw5kaeu4g3d7x7w0734g:1\ng1hygx8ga9qakhkczyrzs9drm8j8tu4qds9y5e3r:1\ng1f977l6wxdh3qu60kzl75vx2wmzswu68l03r8su:1\ng1644qje5rx6jsdqfkzmgnfcegx4dxkjh6rwqd69:1\ng1mzjajymvmtksdwh3wkrndwj6zls2awl9q83dh6:1\ng14da4n9hcynyzz83q607uu8keuh9hwlv42ra6fa:10\ng14vhcdsyf83ngsrrqc92kmw8q9xakqjm0v8448t:5\n"]}],"fee":{"gas_wanted":"2000000","gas_fee":"1000000ugnot"},"signatures":[{"pub_key":{"@type":"/tm.PubKeySecp256k1","value":"AmG6kzznyo1uNqWPAYU6wDpsmzQKDaEOrVRaZ08vOyX0"},"signature":"S8iMMzlOMK8dmox78R9Z8+pSsS8YaTCXrIcaHDpiOgkOy7gqoQJ0oftM0zf8zAz4xpezK8Lzg8Q0fCdXJxV76w=="}],"memo":""}
{"msg":[{"@type":"/vm.m_call","caller":"g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq","send":"","pkg_path":"gno.land/r/demo/users","func":"Invite","args":["g1thlf3yct7n7ex70k0p62user0kn6mj6d3s0cg3\ng1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5\n"]}],"fee":{"gas_wanted":"2000000","gas_fee":"1000000ugnot"},"signatures":[{"pub_key":{"@type":"/tm.PubKeySecp256k1","value":"AmG6kzznyo1uNqWPAYU6wDpsmzQKDaEOrVRaZ08vOyX0"},"signature":"njczE6xYdp01+CaUU/8/v0YC/NuZD06+qLind+ZZEEMNaRe/4Ln+4z7dG6HYlaWUMsyI1KCoB6NIehoE0PZ44Q=="}],"memo":""}
{"msg":[{"@type":"/vm.m_call","caller":"g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq","send":"","pkg_path":"gno.land/r/demo/users","func":"Invite","args":["g1589c8cekvmjfmy0qrd4f3z52r7fn7rgk02667s\ng13sm84nuqed3fuank8huh7x9mupgw22uft3lcl8\ng1m6732pkrngu9vrt0g7056lvr9kcqc4mv83xl5q\ng1wg88rhzlwxjd2z4j5de5v5xq30dcf6rjq3dhsj\ng18pmaskasz7mxj6rmgrl3al58xu45a7w0l5nmc0\ng19wwhkmqlns70604ksp6rkuuu42qhtvyh05lffz\ng187982000zsc493znqt828s90cmp6hcp2erhu6m\ng1ndpsnrspdnauckytvkfv8s823t3gmpqmtky8pl\ng16ja66d65emkr0zxd2tu7xjvm7utthyhpej0037\ng1ds24jj9kqjcskd0gzu24r9e4n62ggye230zuv5\ng1trkzq75ntamsnw9xnrav2v7gy2lt5g6p29yhdr\ng1rrf8s5mrmu00sx04fzfsvc399fklpeg2x0a7mz\ng19p5ntfvpt4lwq4jqsmnxsnelhf3tff9scy3w8w\ng1tue8l73d6rq4vhqdsp2sr3zhuzpure3k2rnwpz\ng14hhsss4ngx5kq77je5g0tl4vftg8qp45ceadk3\ng1768hvkh7anhd40ch4h7jdh6j3mpcs7hrat4gl0\ng15fa8kyjhu88t9dr8zzua8fwdvkngv5n8yqsm0n\ng1xhccdjcscuhgmt3quww6qdy3j3czqt3urc2eac\ng1z629z04f85k4t5gnkk5egpxw9tqxeec435esap\ng1pfldkplz9puq0v82lu9vqcve9nwrxuq9qe5ttv\ng152pn0g5qfgxr7yx8zlwjq48hytkafd8x7egsfv\ng1cf2ye686ke38vjyqakreprljum4xu6rwf5jskq\ng1c5shztyaj4gjrc5zlwmh9xhex5w7l4asffs2w6\ng1lhpx2ktk0ha3qw42raxq4m24a4c4xqxyrgv54q\ng1026p54q0j902059sm2zsv37krf0ghcl7gmhyv7\ng1n4yvwnv77frq2ccuw27dmtjkd7u4p4jg0pgm7k\ng13m7f2e6r3lh3ykxupacdt9sem2tlvmaamwjhll\ng19uxluuecjlsqvwmwu8sp6pxaaqfhk972q975xd\ng1j80fpcsumfkxypvydvtwtz3j4sdwr8c2u0lr64\ng1tjdpptuk9eysq6z38nscqyycr998xjyx3w8jvw\ng19t3n89slfemgd3mwuat4lajwcp0yxrkadgeg7a\ng1yqndt8xx92l9h494jfruz2w79swzjes3n4wqjc\ng13278z0a5ufeg80ffqxpda9dlp599t7ekregcy6\ng1ht236wjd83x96uqwh9rh3fq6pylyn78mtwq9v6\ng1fj9jccm3zjnqspq7lp2g7lj4czyfq0s35600g9\ng1wwppuzdns5u6c6jqpkzua24zh6ppsus6399cea\ng1k8pjnguyu36pkc8hy0ufzgpzfmj2jl78la7ek3\ng1e8umkzumtxgs8399lw0us4rclea3xl5gxy9spp\ng14qekdkj2nmmwea4ufg9n002a3pud23y8k7ugs5\ng19w2488ntfgpduzqq3sk4j5x387zynwknqdvjqf\ng1495y3z7zrej4rendysnw5kaeu4g3d7x7w0734g\ng1hygx8ga9qakhkczyrzs9drm8j8tu4qds9y5e3r\ng1f977l6wxdh3qu60kzl75vx2wmzswu68l03r8su\ng1644qje5rx6jsdqfkzmgnfcegx4dxkjh6rwqd69\ng1mzjajymvmtksdwh3wkrndwj6zls2awl9q83dh6\ng1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq\ng14da4n9hcynyzz83q607uu8keuh9hwlv42ra6fa\ng14vhcdsyf83ngsrrqc92kmw8q9xakqjm0v8448t\n"]}],"fee":{"gas_wanted":"4000000","gas_fee":"1000000ugnot"},"signatures":[{"pub_key":{"@type":"/tm.PubKeySecp256k1","value":"AmG6kzznyo1uNqWPAYU6wDpsmzQKDaEOrVRaZ08vOyX0"},"signature":"7AmlhZhsVkxCUl0bbpvpPMnIKihwtG7A5IFR6Tg4xStWLgaUr05XmWRKlO2xjstTtwbVKQT5mFL4h5wyX4SQzw=="}],"memo":""}
```

To add genesis (deploy) transactions from a directory:

```shell
gnogenesis txs add packages ./examples
```

This will generate `MsgAddPkg` transactions, and add them to the given `genesis.json`.

#### Remove genesis transactions

To clear specific transactions, use the transaction hash:

```shell
gnogenesis txs remove "5HuU9LN8WUa2NsjiNxp8Xii9n0zlSGXc9UqzLHB+DPs="
```
To specify a deployer address (package creator) on add packages command
```shell
gnogenesis txs add packages ./examples --deployer-address=SOME_ADDRESS
```

The transaction hash is the base64 encoding of the Amino-Binary encoded `std.Tx` transaction hash.

The steps to get this sort of hash are:

- get the `std.Tx`
- marshal it using `amino.Marshal`
- cast the result to `types.Tx` (`bft`)
- call `Hash` on the `types.Tx`
- encode the result into base64
