// @ts-check

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
    tutorialSidebar: [
        'overview',
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
                'how-to-guides/testing-gno',
                'how-to-guides/deploy',
                'how-to-guides/creating-grc20',
                'how-to-guides/creating-grc721',
                'how-to-guides/connect-wallet-dapp',
                'how-to-guides/write-simple-dapp',
                'how-to-guides/sync-gno-nodes',
            ],
        },
        {
            type: 'category',
            label: 'Reference',
            items: [
                'reference/rpc-endpoints',
                'reference/node-configuration',
                'reference/standard-library',
                /*
                TODO: add tm2-js-client
                TODO: add gno-js-client
                 */
            ],
        },
        {
            type: 'category',
            label: 'Explanation',
            items: [
                'explanation/realms',
                'explanation/explain-standard-library',
                'explanation/tendermint2',
                'explanation/gnovm',
                'explanation/ibc',
                'explanation/gno-language',
                'explanation/gno-tooling',
                'explanation/gno-modules',
                'explanation/gno-test',
                'explanation/gno-doc',
            ],
        },
    ],
};

module.exports = sidebars;
