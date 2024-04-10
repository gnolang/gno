// @ts-check

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
    tutorialSidebar: [
        'overview',
        {
            type: 'category',
            label: 'Getting Started',
            link: {type: 'doc', id: 'getting-started/getting-started'},
            items: [
                'getting-started/playground-start',
                {
                 type: 'category',
                 label: 'Local Setup',
                    items: [
                        'getting-started/local-setup/local-setup',
                        'getting-started/local-setup/working-with-key-pairs',
                        'getting-started/local-setup/premining-balances',
                        'getting-started/local-setup/setting-up-a-local-chain',
                        'getting-started/local-setup/browsing-gno-source-code',
                    ]
                },
            ],
        },
        {
            type: 'category',
            label: 'How-to Guides',
            link: {type: 'doc', id: 'how-to-guides/how-to-guides'},
            items: [
                'how-to-guides/simple-contract',
                'how-to-guides/simple-library',
                'how-to-guides/testing-gno',
                'how-to-guides/deploy',
                'how-to-guides/write-simple-dapp',
                'how-to-guides/creating-grc20',
                'how-to-guides/connect-from-go',
                'how-to-guides/connect-wallet-dapp',
            ],
        },
        {
            type: 'category',
            label: 'Concepts',
            link: {type: 'doc', id: 'concepts/concepts'},
            items: [
                'concepts/realms',
                'concepts/packages',
                {
                    type: 'category',
                    label: 'Standard Libraries',
                    link: {type: 'doc', id: 'concepts/stdlibs/stdlibs'},
                    items: [
                        'concepts/stdlibs/banker',
                        'concepts/stdlibs/coin',
                        'concepts/stdlibs/gnopher-hole-stdlib',
                    ]
                },
                'concepts/gnovm',
                'concepts/gno-language',
                'concepts/testnets',
                'concepts/effective-gno',
                'concepts/proof-of-contribution',
                'concepts/tendermint2',
                'concepts/portal-loop',
                'concepts/gno-modules',
                'concepts/gno-test',
                'concepts/from-go-to-gno',
            ],
        },
        {
            type: 'category',
            label: 'Gno Tooling',
            link: {type: 'doc', id: 'gno-tooling/gno-tooling'},
            items: [
                'gno-tooling/cli/gno-tooling-gno',
                'gno-tooling/cli/gno-tooling-gnokey',
                'gno-tooling/cli/gno-tooling-gnodev',
                'gno-tooling/cli/gno-tooling-gnoland',
                {
                    type: 'category',
                    label: 'gnofaucet',
                    link: {type: 'doc', id: 'gno-tooling/cli/faucet/gno-tooling-gnofaucet'},
                    items: [
                        'gno-tooling/cli/faucet/running-a-faucet',
                    ]
                },
            ]
        },
        {
            type: 'category',
            label: 'Reference',
            link: {type: 'doc', id: 'reference/reference'},
            items: [
                'reference/rpc-endpoints',
                'reference/network-config',
                {
                    type: 'category',
                    label: 'Standard Libraries',
                    link: {type: 'doc', id: 'reference/stdlibs/stdlibs'},
                    items: [
                        {
                            type: 'category',
                            label: 'std',
                            items: [
                                'reference/stdlibs/std/address',
                                'reference/stdlibs/std/banker',
                                'reference/stdlibs/std/coin',
                                'reference/stdlibs/std/coins',
                                'reference/stdlibs/std/chain',
                                'reference/stdlibs/std/testing',
                            ]
                        }
                    ]
                },
                'reference/go-gno-compatibility',
                {
                    type: 'category',
                    label: 'tm2-js-client',
                    link: {type: 'doc', id: 'reference/tm2-js-client/tm2-js-client'},
                    items: [
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
                    link: {type: 'doc', id: 'reference/gno-js-client/gno-js-client'},
                    items: [
                        'reference/gno-js-client/gno-js-provider',
                        'reference/gno-js-client/gno-js-wallet',
                    ]
                },
                {
                    type: 'category',
                    label: 'gnoclient',
                    link: {type: 'doc', id: 'reference/gnoclient/gnoclient'},
                    items: [
                        'reference/gnoclient/signer',
                        'reference/gnoclient/client'
                    ]
                },
            ],
        },
    ],
};

module.exports = sidebars;
