# GNS Token

The `GNS` token is the governance and main utility token of the GnoSwap protocol

## Token Implementation

- Follows the `grc20` token specs ([grc spec](https://github.com/gnolang/gno/tree/master/examples/gno.land/p/demo/grc))
- Symbol: `GNS`
- Decimals: 6
- Max Supply: 1_000_000_000 (1B)

You can find the detailed tokenomics of GNS [here](https://docs.gnoswap.io/gnoswap-token/whats-gns).

## Emission Mechanism

- Block-based token emission with predefined schedule
- Token emission follows a halving model over 12 years
- Each halving period adjusts the amounts of tokens minted per block

## Halving Schedule

- 12 years emission period devided into halving periods

| Years | Description |
| --- | --- |
| 1-2 | Full emission rate |
| 3-4 | 50% of initial emission rate |
| 5-6 | 25% of initial emission rate |
| 7-8 | 12.5% of initial emission rate |
| 9-12 | 6.25% of initial emission rate |

For details about the emission and distribution of the GNS tokens, refer to our [emission contract](https://github.com/gnoswap-labs/gnoswap/tree/main/contract/r/gnoswap/emission).
