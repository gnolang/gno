// @ts-check

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
    tutorialSidebar: [
        'overview',
        'peace',
        {
            type: 'category',
            label: 'Getting Started',
            items: [
                'getting-started/installation',
                'getting-started/working-with-key-pairs',
                {
                    type: 'category',
                    label: 'Setting up Funds',
                    items: [
                        'getting-started/setting-up-funds/premining-balances',
                        'getting-started/setting-up-funds/running-a-faucet',
                    ]
                },
                'getting-started/setting-up-a-local-chain',
                'getting-started/browsing-gno-source-code',
            ],
        },
        {
            type: 'category',
            label: 'How-to Guides',
            items: [
                'how-to-guides/simple-contract',
                'how-to-guides/simple-library',
                'how-to-guides/testing-gno',
                'how-to-guides/deploy',
                'how-to-guides/write-simple-dapp',
                'how-to-guides/creating-grc20',
                'how-to-guides/creating-grc721',
                'how-to-guides/connect-wallet-dapp',
            ],
        },
        {
            type: 'category',
            label: 'Reference',
            items: [
                'reference/rpc-endpoints',
                'reference/standard-library',
                'reference/go-gno-compatibility',
                {
                    type: 'category',
                    label: 'tm2-js-client',
                    items: [
                        'reference/tm2-js-client/tm2-js-getting-started',
                        'reference/tm2-js-client/tm2-js-wallet',
                        {
                            type: 'category',
                            label: 'Provider',
                            items: [
                                'reference/tm2-js-client/Provider/tm2-js-provider',
                                'reference/tm2-js-client/Provider/tm2-js-json-rpc-provider',
                                'reference/tm2-js-client/Provider/tm2-js-ws-provider',
                                'reference/tm2-js-client/Provider/tm2-js-utility',
                            ]
                        },
                        {
                            type: 'category',
                            label: 'Signer',
                            items: [
                                'reference/tm2-js-client/Signer/tm2-js-signer',
                                'reference/tm2-js-client/Signer/tm2-js-key',
                                'reference/tm2-js-client/Signer/tm2-js-ledger',
                            ]
                        },
                    ]
                },
                {
                    type: 'category',
                    label: 'gno-js-client',
                    items: [
                        'reference/gno-js-client/gno-js-getting-started',
                        'reference/gno-js-client/gno-js-provider',
                        'reference/gno-js-client/gno-js-wallet',
                    ]
                },
            ],
        },
        {
            type: 'category',
            label: 'Explanation',
            items: [
                'explanation/realms',
                'explanation/tendermint2',
                'explanation/gnovm',
                'explanation/proof-of-contribution',
                'explanation/gno-language',
                'explanation/gno-modules',
                'explanation/gno-test',
                'explanation/from-go-to-gno',
                {
                    type: 'category',
                    label: 'Gno Tooling',
                    items: [
                        'explanation/gno-tooling/cli/gno-tooling-common-params',
                        'explanation/gno-tooling/cli/gno-tooling-gno',
                        'explanation/gno-tooling/cli/gno-tooling-gnokey',
                        'explanation/gno-tooling/cli/gno-tooling-gnofaucet',
                        'explanation/gno-tooling/cli/gno-tooling-gnoland',
                        'explanation/gno-tooling/cli/gno-tooling-tm2txsync',
                    ]
                },
            ],
        },
    ],
};

module.exports = sidebars;
